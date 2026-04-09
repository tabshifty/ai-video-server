package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"video-server/internal/models"
)

func (r *VideoRepository) CountVideosByType(ctx context.Context) (shorts, movies, episodes int64, err error) {
	err = r.pool.QueryRow(ctx, `
SELECT
  COUNT(*) FILTER (WHERE type='short') AS shorts,
  COUNT(*) FILTER (WHERE type='movie') AS movies,
  COUNT(*) FILTER (WHERE type='episode') AS episodes
FROM videos
`).Scan(&shorts, &movies, &episodes)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("count videos by type: %w", err)
	}
	return shorts, movies, episodes, nil
}

func (r *VideoRepository) CountUsers(ctx context.Context) (int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return total, nil
}

func (r *VideoRepository) CountTodayUploads(ctx context.Context) (int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM videos
WHERE created_at >= date_trunc('day', NOW())
`).Scan(&total); err != nil {
		return 0, fmt.Errorf("count today uploads: %w", err)
	}
	return total, nil
}

func (r *VideoRepository) WeeklyUploadTrend(ctx context.Context) ([]models.DailyUploads, error) {
	rows, err := r.pool.Query(ctx, `
SELECT to_char(day, 'YYYY-MM-DD') AS day, COALESCE(v.cnt, 0) AS cnt
FROM generate_series(current_date - interval '6 day', current_date, interval '1 day') day
LEFT JOIN (
  SELECT date_trunc('day', created_at)::date AS d, COUNT(*) AS cnt
  FROM videos
  GROUP BY d
) v ON v.d = day::date
ORDER BY day ASC
`)
	if err != nil {
		return nil, fmt.Errorf("weekly upload trend: %w", err)
	}
	defer rows.Close()

	out := make([]models.DailyUploads, 0, 7)
	for rows.Next() {
		var item models.DailyUploads
		if err := rows.Scan(&item.Day, &item.Count); err != nil {
			return nil, fmt.Errorf("scan weekly trend: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *VideoRepository) AdminListVideos(ctx context.Context, f models.AdminVideoFilter) ([]models.AdminVideoListItem, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 10)
	next := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if t := strings.TrimSpace(strings.ToLower(f.Type)); t != "" {
		where = append(where, "v.type = "+next(t))
	}
	if s := strings.TrimSpace(strings.ToLower(f.Status)); s != "" {
		where = append(where, "v.status = "+next(s))
	}
	if q := strings.TrimSpace(strings.ToLower(f.Keyword)); q != "" {
		like := "%" + q + "%"
		p := next(like)
		where = append(where, "(LOWER(v.title) LIKE "+p+" OR LOWER(COALESCE(v.description,'')) LIKE "+p+" OR EXISTS (SELECT 1 FROM video_tags vt WHERE vt.video_id=v.id AND LOWER(vt.tag) LIKE "+p+"))")
	}
	if u := strings.TrimSpace(strings.ToLower(f.User)); u != "" {
		p := next("%" + u + "%")
		where = append(where, "(LOWER(COALESCE(us.username,'')) LIKE "+p+" OR LOWER(COALESCE(us.email,'')) LIKE "+p+")")
	}
	if tag := strings.TrimSpace(strings.ToLower(f.Tag)); tag != "" {
		p := next("%" + tag + "%")
		where = append(where, "EXISTS (SELECT 1 FROM video_tags vt WHERE vt.video_id=v.id AND LOWER(vt.tag) LIKE "+p+")")
	}
	if f.StartTime != nil {
		where = append(where, "v.created_at >= "+next(*f.StartTime))
	}
	if f.EndTime != nil {
		where = append(where, "v.created_at <= "+next(*f.EndTime))
	}

	baseWhere := strings.Join(where, " AND ")
	var total int
	countSQL := "SELECT COUNT(*) FROM videos v LEFT JOIN users us ON us.id = v.user_id WHERE " + baseWhere
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("admin count videos: %w", err)
	}

	limit := f.PageSize
	offset := (f.Page - 1) * f.PageSize
	args = append(args, limit, offset)
	listSQL := "SELECT v.id, v.title, v.type, v.status, COALESCE(v.thumbnail_path,''), v.user_id, COALESCE(us.username,''), v.created_at, v.updated_at FROM videos v LEFT JOIN users us ON us.id = v.user_id WHERE " + baseWhere + " ORDER BY v.created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)-1) + " OFFSET $" + fmt.Sprintf("%d", len(args))
	rows, err := r.pool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("admin list videos: %w", err)
	}
	defer rows.Close()

	items := make([]models.AdminVideoListItem, 0, limit)
	for rows.Next() {
		var item models.AdminVideoListItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Type, &item.Status, &item.Thumbnail, &item.UploadUserID, &item.UploadUser, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan admin video: %w", err)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *VideoRepository) AdminVideoDetail(ctx context.Context, videoID uuid.UUID) (models.AdminVideoDetail, error) {
	var out models.AdminVideoDetail
	var metadata []byte
	err := r.pool.QueryRow(ctx, `
SELECT
  id,
  user_id,
  COALESCE(title, ''),
  COALESCE(description, ''),
  type,
  status,
  COALESCE(duration_seconds, 0),
  COALESCE(width, 0),
  COALESCE(height, 0),
  COALESCE(original_path, ''),
  COALESCE(transcoded_path, ''),
  COALESCE(thumbnail_path, ''),
  COALESCE(metadata, '{}'::jsonb),
  created_at,
  updated_at
FROM videos
WHERE id=$1
`, videoID).Scan(
		&out.ID, &out.UserID, &out.Title, &out.Description, &out.Type, &out.Status, &out.DurationSeconds,
		&out.Width, &out.Height, &out.OriginalPath, &out.TranscodedPath, &out.ThumbnailPath, &metadata, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return models.AdminVideoDetail{}, fmt.Errorf("admin video detail: %w", err)
	}
	if len(metadata) == 0 {
		metadata = []byte(`{}`)
	}
	out.Metadata = metadata

	rows, err := r.pool.Query(ctx, `SELECT tag FROM video_tags WHERE video_id=$1 ORDER BY tag`, videoID)
	if err != nil {
		return models.AdminVideoDetail{}, fmt.Errorf("admin video tags: %w", err)
	}
	defer rows.Close()
	tags := make([]string, 0, 8)
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return models.AdminVideoDetail{}, fmt.Errorf("scan admin tag: %w", err)
		}
		tags = append(tags, tag)
	}
	out.Tags = tags
	return out, rows.Err()
}

func (r *VideoRepository) AdminUpdateVideo(ctx context.Context, videoID uuid.UUID, title, description, thumbnail string, tags []string, metadata map[string]any) error {
	raw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal admin metadata: %w", err)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
UPDATE videos
SET title=$2, description=$3, thumbnail_path=$4, metadata=$5, updated_at=NOW()
WHERE id=$1
`, videoID, title, description, thumbnail, raw); err != nil {
		return fmt.Errorf("admin update video: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM video_tags WHERE video_id=$1`, videoID); err != nil {
		return fmt.Errorf("clear tags: %w", err)
	}
	for _, t := range tags {
		tag := strings.TrimSpace(strings.ToLower(t))
		if tag == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO video_tags(video_id, tag, weight) VALUES ($1,$2,1.0)`, videoID, tag); err != nil {
			return fmt.Errorf("insert admin tag: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit admin update video: %w", err)
	}
	return nil
}

func (r *VideoRepository) AdminUpdateVideoStatus(ctx context.Context, videoID uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE videos SET status=$2, updated_at=NOW() WHERE id=$1`, videoID, status)
	if err != nil {
		return fmt.Errorf("admin update video status: %w", err)
	}
	return nil
}

func (r *VideoRepository) AdminListUsers(ctx context.Context, page, pageSize int) ([]models.AdminUserListItem, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("admin count users: %w", err)
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, username, email, role, created_at
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("admin list users: %w", err)
	}
	defer rows.Close()

	out := make([]models.AdminUserListItem, 0, pageSize)
	for rows.Next() {
		var item models.AdminUserListItem
		if err := rows.Scan(&item.ID, &item.Username, &item.Email, &item.Role, &item.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan admin user: %w", err)
		}
		out = append(out, item)
	}
	return out, total, rows.Err()
}

func (r *VideoRepository) AdminUpdateUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET role=$2, updated_at=NOW() WHERE id=$1`, userID, role)
	if err != nil {
		return fmt.Errorf("admin update user role: %w", err)
	}
	return nil
}

func (r *VideoRepository) AdminListTranscodingTasks(ctx context.Context, page, pageSize int) ([]models.AdminTaskListItem, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM transcoding_jobs`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, video_id, user_id, status, retry_count, COALESCE(error_message,''), started_at, finished_at
FROM transcoding_jobs
ORDER BY id DESC
LIMIT $1 OFFSET $2
`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	out := make([]models.AdminTaskListItem, 0, pageSize)
	for rows.Next() {
		var item models.AdminTaskListItem
		if err := rows.Scan(&item.ID, &item.VideoID, &item.UserID, &item.Status, &item.RetryCount, &item.Error, &item.StartedAt, &item.FinishedAt); err != nil {
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		out = append(out, item)
	}
	return out, total, rows.Err()
}

func (r *VideoRepository) AdminRetryFailedTask(ctx context.Context, jobID int64) (*uuid.UUID, error) {
	var videoID *uuid.UUID
	err := r.pool.QueryRow(ctx, `
UPDATE transcoding_jobs
SET status='queued', retry_count=retry_count+1, error_message='', started_at=NOW(), finished_at=NULL
WHERE id=$1 AND status='failed'
RETURNING video_id
`, jobID).Scan(&videoID)
	if err != nil {
		return nil, fmt.Errorf("retry failed task: %w", err)
	}
	return videoID, nil
}
