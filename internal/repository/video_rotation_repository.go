package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"video-server/internal/models"
)

// ErrVideoMutationBusy indicates another process owns this video's files.
var ErrVideoMutationBusy = errors.New("视频文件正在处理，请稍后重试")

// LockVideoFiles holds a PostgreSQL session lock through all file operations.
// The returned release function must always be called.
func (r *VideoRepository) LockVideoFiles(ctx context.Context, id uuid.UUID) (func(), error) {
	// A dedicated session avoids exhausting the query pool with long encodes.
	c, err := pgx.ConnectConfig(ctx, r.pool.Config().ConnConfig.Copy())
	if err != nil {
		return nil, err
	}
	var locked bool
	if err = c.QueryRow(ctx, "SELECT pg_try_advisory_lock(hashtextextended($1, 49349))", id.String()).Scan(&locked); err != nil || !locked {
		_ = c.Close(context.WithoutCancel(ctx))
		if err != nil {
			return nil, err
		}
		return nil, ErrVideoMutationBusy
	}
	return func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var unlocked bool
		if err := c.QueryRow(releaseCtx, "SELECT pg_advisory_unlock(hashtextextended($1, 49349))", id.String()).Scan(&unlocked); err != nil || !unlocked {
			_ = c.Close(releaseCtx)
		}
		_ = c.Close(releaseCtx)
	}, nil
}

const rotationColumns = `id,video_id,requested_by,angle,mode,status,input_path,input_size,input_mtime,original_path,old_thumbnail,output_path,thumbnail_path,progress,error_message,created_at,updated_at`

func scanRotation(row pgx.Row) (models.VideoRotationJob, error) {
	var j models.VideoRotationJob
	err := row.Scan(&j.ID, &j.VideoID, &j.RequestedBy, &j.Angle, &j.Mode, &j.Status, &j.InputPath, &j.InputSize, &j.InputMtime, &j.OriginalPath, &j.OldThumbnail, &j.OutputPath, &j.ThumbnailPath, &j.Progress, &j.ErrorMessage, &j.CreatedAt, &j.UpdatedAt)
	return j, err
}

// CreateRotation serializes admission with all updates to the video row.
func (r *VideoRepository) CreateRotation(ctx context.Context, j models.VideoRotationJob) (models.VideoRotationJob, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return j, err
	}
	defer tx.Rollback(ctx)
	var path, status string
	if err = tx.QueryRow(ctx, `SELECT COALESCE(transcoded_path,''),status FROM videos WHERE id=$1 FOR UPDATE`, j.VideoID).Scan(&path, &status); err != nil {
		return j, err
	}
	previous, e := scanRotation(tx.QueryRow(ctx, `SELECT `+rotationColumns+` FROM video_rotation_jobs WHERE id=$1`, j.ID))
	if e == nil {
		if previous.VideoID != j.VideoID || previous.Angle != j.Angle || previous.Mode != j.Mode || previous.RequestedBy != j.RequestedBy {
			return j, fmt.Errorf("请求标识已被其它操作使用")
		}
		return previous, tx.Commit(ctx)
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return j, e
	}
	if status != "ready" || path != j.InputPath {
		return j, fmt.Errorf("视频已变化，请刷新后重新确认")
	}
	if _, err = tx.Exec(ctx, `INSERT INTO video_rotation_jobs(id,video_id,requested_by,angle,mode,input_path,input_size,input_mtime,original_path,old_thumbnail,output_path,thumbnail_path) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, j.ID, j.VideoID, j.RequestedBy, j.Angle, j.Mode, j.InputPath, j.InputSize, j.InputMtime, j.OriginalPath, j.OldThumbnail, j.OutputPath, j.ThumbnailPath); err != nil {
		return j, fmt.Errorf("该视频已有修正或清理任务: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return j, err
	}
	return r.RotationByID(ctx, j.ID)
}

// RotationByID reads a persisted job.
func (r *VideoRepository) RotationByID(ctx context.Context, id uuid.UUID) (models.VideoRotationJob, error) {
	return scanRotation(r.pool.QueryRow(ctx, `SELECT `+rotationColumns+` FROM video_rotation_jobs WHERE id=$1`, id))
}

// LatestRotation reads the newest job for a video's detail panel.
func (r *VideoRepository) LatestRotation(ctx context.Context, id uuid.UUID) (*models.VideoRotationJob, error) {
	j, err := scanRotation(r.pool.QueryRow(ctx, `SELECT `+rotationColumns+` FROM video_rotation_jobs WHERE video_id=$1 ORDER BY created_at DESC LIMIT 1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &j, err
}

// ActiveRotations lists durable work for queue reconciliation.
func (r *VideoRepository) ActiveRotations(ctx context.Context) ([]models.VideoRotationJob, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+rotationColumns+` FROM video_rotation_jobs WHERE status IN ('queued','running','cleanup') ORDER BY updated_at LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.VideoRotationJob{}
	for rows.Next() {
		j, e := scanRotation(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// SetRotationState updates progress without changing video availability.
func (r *VideoRepository) SetRotationState(ctx context.Context, id uuid.UUID, state string, progress float64, message string) error {
	_, err := r.pool.Exec(ctx, `UPDATE video_rotation_jobs SET status=$2,progress=$3,error_message=$4,updated_at=NOW() WHERE id=$1`, id, state, progress, message)
	return err
}

// PublishRotation atomically switches pointers and records the cleanup phase.
func (r *VideoRepository) PublishRotation(ctx context.Context, j models.VideoRotationJob, size int64, width, height int, thumb string, patch map[string]any) error {
	raw, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT set_config('app.video_rotation_job',$1,true)`, j.ID.String()); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `UPDATE videos SET transcoded_path=$3,original_path='',thumbnail_path=$4,transcoded_file_size=$5,width=$6,height=$7,metadata=COALESCE(metadata,'{}'::jsonb)||$8::jsonb,updated_at=NOW() WHERE id=$1 AND status='ready' AND transcoded_path=$2 AND COALESCE(thumbnail_path,'')=$9 AND COALESCE(original_path,'')=$10`, j.VideoID, j.InputPath, j.OutputPath, thumb, size, width, height, raw, j.OldThumbnail, j.OriginalPath)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("视频文件或封面已变化，停止发布")
	}
	if _, err = tx.Exec(ctx, `UPDATE video_rotation_jobs SET status='cleanup',progress=100,error_message='',updated_at=NOW() WHERE id=$1 AND status='running'`, j.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// RotationFileReferenced prevents cleanup of paths used by any current resource.
func (r *VideoRepository) RotationFileReferenced(ctx context.Context, path string) (bool, error) {
	var used bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM videos WHERE original_path=$1 OR transcoded_path=$1 OR thumbnail_path=$1 OR metadata::text LIKE '%' || $1 || '%') OR EXISTS(SELECT 1 FROM images WHERE original_path=$1 OR stored_path=$1) OR EXISTS(SELECT 1 FROM video_subtitles WHERE stored_path=$1)`, path).Scan(&used)
	return used, err
}
