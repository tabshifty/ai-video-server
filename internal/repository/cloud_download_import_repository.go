package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"video-server/internal/models"
)

// CloudDownloadImportRepository persists explicit cloud download imports.
type CloudDownloadImportRepository struct {
	pool *pgxpool.Pool
}

// NewCloudDownloadImportRepository builds the repository.
func NewCloudDownloadImportRepository(pool *pgxpool.Pool) *CloudDownloadImportRepository {
	return &CloudDownloadImportRepository{pool: pool}
}

const cloudDownloadImportSelect = `
SELECT
    id,
    task_id,
    kind,
    status,
    COALESCE(source_path, ''),
    COALESCE(source_size, 0),
    COALESCE(sha256, ''),
    video_id,
    archive_batch_id,
    attempt,
    COALESCE(error_message, ''),
    created_by,
    created_at,
    started_at,
    imported_at
FROM cloud_download_imports`

func scanCloudDownloadImport(row interface{ Scan(...any) error }) (models.CloudDownloadImport, error) {
	var item models.CloudDownloadImport
	err := row.Scan(
		&item.ID,
		&item.TaskID,
		&item.Kind,
		&item.Status,
		&item.SourcePath,
		&item.SourceSize,
		&item.SHA256,
		&item.VideoID,
		&item.ArchiveBatchID,
		&item.Attempt,
		&item.ErrorMessage,
		&item.CreatedBy,
		&item.CreatedAt,
		&item.StartedAt,
		&item.ImportedAt,
	)
	return item, err
}

// CreateOrGet inserts one import per task. The bool reports whether this call
// created the row.
func (r *CloudDownloadImportRepository) CreateOrGet(ctx context.Context, item models.CloudDownloadImport) (models.CloudDownloadImport, bool, error) {
	tag, err := r.pool.Exec(ctx, `
INSERT INTO cloud_download_imports (task_id, kind, status, source_path, source_size, created_by)
VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6)
ON CONFLICT (task_id) DO NOTHING`,
		item.TaskID, item.Kind, item.Status, item.SourcePath, item.SourceSize, item.CreatedBy)
	if err != nil {
		return models.CloudDownloadImport{}, false, fmt.Errorf("insert cloud download import: %w", err)
	}
	existing, err := r.GetByTaskID(ctx, item.TaskID)
	if err != nil {
		return models.CloudDownloadImport{}, false, err
	}
	return existing, tag.RowsAffected() > 0, nil
}

// GetByID returns one import record.
func (r *CloudDownloadImportRepository) GetByID(ctx context.Context, id int64) (models.CloudDownloadImport, error) {
	row := r.pool.QueryRow(ctx, cloudDownloadImportSelect+` WHERE id = $1`, id)
	item, err := scanCloudDownloadImport(row)
	if err != nil {
		return models.CloudDownloadImport{}, fmt.Errorf("get cloud download import: %w", err)
	}
	return item, nil
}

// GetByTaskID returns the import attached to a cloud task.
func (r *CloudDownloadImportRepository) GetByTaskID(ctx context.Context, taskID int64) (models.CloudDownloadImport, error) {
	row := r.pool.QueryRow(ctx, cloudDownloadImportSelect+` WHERE task_id = $1`, taskID)
	item, err := scanCloudDownloadImport(row)
	if err != nil {
		return models.CloudDownloadImport{}, fmt.Errorf("get cloud download import by task: %w", err)
	}
	return item, nil
}

// UpdateProgress records a non-terminal status transition and its error.
func (r *CloudDownloadImportRepository) UpdateProgress(ctx context.Context, id int64, status, errorMessage string) error {
	_, err := r.pool.Exec(ctx, `
UPDATE cloud_download_imports SET
    status = $2,
    error_message = NULLIF($3, ''),
    started_at = CASE
        WHEN $2 IN ('hashing', 'copying', 'creating') AND started_at IS NULL THEN now()
        ELSE started_at
    END
WHERE id = $1`, id, status, errorMessage)
	if err != nil {
		return fmt.Errorf("update cloud download import progress: %w", err)
	}
	return nil
}

// SetHash persists the SHA-256 calculated from the imported source file.
func (r *CloudDownloadImportRepository) SetHash(ctx context.Context, id int64, sha256 string) error {
	if _, err := r.pool.Exec(ctx, `
UPDATE cloud_download_imports SET sha256 = NULLIF($2, '') WHERE id = $1`, id, sha256); err != nil {
		return fmt.Errorf("set cloud download import sha256: %w", err)
	}
	return nil
}

// SetVideo links the imported video target.
func (r *CloudDownloadImportRepository) SetVideo(ctx context.Context, id int64, videoID uuid.UUID) error {
	if _, err := r.pool.Exec(ctx, `
UPDATE cloud_download_imports SET video_id = $2 WHERE id = $1`, id, videoID); err != nil {
		return fmt.Errorf("set cloud download import video: %w", err)
	}
	return nil
}

// SetArchiveBatch links the created archive import batch.
func (r *CloudDownloadImportRepository) SetArchiveBatch(ctx context.Context, id int64, batchID uuid.UUID) error {
	if _, err := r.pool.Exec(ctx, `
UPDATE cloud_download_imports SET archive_batch_id = $2 WHERE id = $1`, id, batchID); err != nil {
		return fmt.Errorf("set cloud download import archive batch: %w", err)
	}
	return nil
}

// Finish writes a terminal state.
func (r *CloudDownloadImportRepository) Finish(ctx context.Context, id int64, status string, videoID, archiveBatchID *uuid.UUID, importedAt time.Time) error {
	if _, err := r.pool.Exec(ctx, `
UPDATE cloud_download_imports SET
    status = $2,
    video_id = $3,
    archive_batch_id = $4,
    imported_at = $5,
    error_message = NULL,
    started_at = COALESCE(started_at, now())
WHERE id = $1`, id, status, videoID, archiveBatchID, importedAt); err != nil {
		return fmt.Errorf("finish cloud download import: %w", err)
	}
	return nil
}

// ResetForRetry bumps the attempt and returns the row to queued. The bool is
// false when a concurrent request already moved the record out of failed.
func (r *CloudDownloadImportRepository) ResetForRetry(ctx context.Context, id int64) (models.CloudDownloadImport, bool, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE cloud_download_imports SET
    attempt = attempt + 1,
    status = 'queued',
    error_message = NULL,
    started_at = NULL,
    imported_at = NULL
WHERE id = $1 AND status = 'failed'`, id)
	if err != nil {
		return models.CloudDownloadImport{}, false, fmt.Errorf("reset cloud download import for retry: %w", err)
	}
	item, err := r.GetByID(ctx, id)
	if err != nil {
		return models.CloudDownloadImport{}, false, err
	}
	return item, tag.RowsAffected() > 0, nil
}

// ListSummariesByTaskIDs returns the compact import view for task list rows.
func (r *CloudDownloadImportRepository) ListSummariesByTaskIDs(ctx context.Context, taskIDs []int64) (map[int64]models.CloudDownloadImportSummary, error) {
	out := make(map[int64]models.CloudDownloadImportSummary, len(taskIDs))
	if len(taskIDs) == 0 {
		return out, nil
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, task_id, kind, status, video_id, archive_batch_id, COALESCE(error_message, ''), created_at
FROM cloud_download_imports
WHERE task_id = ANY($1)`, taskIDs)
	if err != nil {
		return nil, fmt.Errorf("list cloud download import summaries: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var taskID int64
		var item models.CloudDownloadImportSummary
		if err := rows.Scan(&item.ID, &taskID, &item.Kind, &item.Status, &item.VideoID, &item.ArchiveBatchID, &item.ErrorMessage, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan cloud download import summary: %w", err)
		}
		out[taskID] = item
	}
	return out, rows.Err()
}
