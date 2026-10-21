package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"video-server/internal/models"
)

// CloudDownloadFileDeletionRepository persists the one-per-task 115 recycle
// bin cleanup record.
type CloudDownloadFileDeletionRepository struct {
	pool *pgxpool.Pool
}

// NewCloudDownloadFileDeletionRepository builds the repository.
func NewCloudDownloadFileDeletionRepository(pool *pgxpool.Pool) *CloudDownloadFileDeletionRepository {
	return &CloudDownloadFileDeletionRepository{pool: pool}
}

const cloudDownloadFileDeletionSelect = `
SELECT
    task_id,
    COALESCE(file_id, ''),
    COALESCE(parent_id, ''),
    status,
    COALESCE(error_message, ''),
    COALESCE(note, ''),
    attempt,
    requested_by,
    requested_at,
    finished_at,
    updated_at
FROM cloud_download_file_deletions`

func scanCloudDownloadFileDeletion(row interface{ Scan(...any) error }) (models.CloudDownloadFileDeletion, error) {
	var item models.CloudDownloadFileDeletion
	err := row.Scan(
		&item.TaskID,
		&item.FileID,
		&item.ParentID,
		&item.Status,
		&item.ErrorMessage,
		&item.Note,
		&item.Attempt,
		&item.RequestedBy,
		&item.RequestedAt,
		&item.FinishedAt,
		&item.UpdatedAt,
	)
	return item, err
}

// GetByTaskID returns the cleanup record, or ErrCloudFileDeletionNotFound.
func (r *CloudDownloadFileDeletionRepository) GetByTaskID(ctx context.Context, taskID int64) (models.CloudDownloadFileDeletion, error) {
	row := r.pool.QueryRow(ctx, cloudDownloadFileDeletionSelect+` WHERE task_id = $1`, taskID)
	item, err := scanCloudDownloadFileDeletion(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.CloudDownloadFileDeletion{}, models.ErrCloudFileDeletionNotFound
	}
	if err != nil {
		return models.CloudDownloadFileDeletion{}, fmt.Errorf("get cloud download file deletion: %w", err)
	}
	return item, nil
}

// Start claims a new attempt. A recent deleting row and a deleted row are not
// claimable; failed rows and stale deleting rows (older than five minutes) are
// reset with a new attempt. The bool reports whether this caller owns the
// attempt.
func (r *CloudDownloadFileDeletionRepository) Start(ctx context.Context, taskID int64, fileID, parentID string, requestedBy *uuid.UUID) (models.CloudDownloadFileDeletion, bool, error) {
	tag, err := r.pool.Exec(ctx, `
INSERT INTO cloud_download_file_deletions (
    task_id, file_id, parent_id, status, attempt, requested_by, requested_at, updated_at
)
VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), 'deleting', 1, $4, now(), now())
ON CONFLICT (task_id) DO UPDATE SET
    file_id = COALESCE(EXCLUDED.file_id, cloud_download_file_deletions.file_id),
    parent_id = COALESCE(EXCLUDED.parent_id, cloud_download_file_deletions.parent_id),
    status = 'deleting',
    error_message = NULL,
    note = NULL,
    attempt = cloud_download_file_deletions.attempt + 1,
    requested_by = EXCLUDED.requested_by,
    requested_at = now(),
    finished_at = NULL,
    updated_at = now()
WHERE cloud_download_file_deletions.status = 'failed'
   OR (
        cloud_download_file_deletions.status = 'deleting'
        AND cloud_download_file_deletions.updated_at < now() - interval '5 minutes'
   )`, taskID, fileID, parentID, requestedBy)
	if err != nil {
		return models.CloudDownloadFileDeletion{}, false, fmt.Errorf("start cloud download file deletion: %w", err)
	}
	item, err := r.GetByTaskID(ctx, taskID)
	if err != nil {
		return models.CloudDownloadFileDeletion{}, false, err
	}
	return item, tag.RowsAffected() > 0, nil
}

// SetFileIDs fills in the resolved ids for the attempt that owns the claim.
func (r *CloudDownloadFileDeletionRepository) SetFileIDs(ctx context.Context, taskID int64, attempt int, fileID, parentID string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE cloud_download_file_deletions SET
    file_id = COALESCE(NULLIF($3, ''), file_id),
    parent_id = COALESCE(NULLIF($4, ''), parent_id),
    updated_at = now()
WHERE task_id = $1 AND attempt = $2 AND status = 'deleting'`, taskID, attempt, fileID, parentID)
	if err != nil {
		return false, fmt.Errorf("set cloud download file deletion ids: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// MarkDeleted finishes the attempt that owns the current attempt number.
func (r *CloudDownloadFileDeletionRepository) MarkDeleted(ctx context.Context, taskID int64, attempt int, note string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE cloud_download_file_deletions SET
    status = 'deleted',
    error_message = NULL,
    note = NULLIF($3, ''),
    finished_at = now(),
    updated_at = now()
WHERE task_id = $1 AND attempt = $2 AND status = 'deleting'`, taskID, attempt, note)
	if err != nil {
		return false, fmt.Errorf("mark cloud download file deleted: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// MarkFailed finishes the attempt that owns the current attempt number.
func (r *CloudDownloadFileDeletionRepository) MarkFailed(ctx context.Context, taskID int64, attempt int, errorMessage string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE cloud_download_file_deletions SET
    status = 'failed',
    error_message = NULLIF($3, ''),
    note = NULL,
    finished_at = now(),
    updated_at = now()
WHERE task_id = $1 AND attempt = $2 AND status = 'deleting'`, taskID, attempt, errorMessage)
	if err != nil {
		return false, fmt.Errorf("mark cloud download file deletion failed: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// ListSummariesByTaskIDs returns the compact cleanup view for task list rows.
func (r *CloudDownloadFileDeletionRepository) ListSummariesByTaskIDs(ctx context.Context, taskIDs []int64) (map[int64]models.CloudDownloadFileDeletionSummary, error) {
	out := make(map[int64]models.CloudDownloadFileDeletionSummary, len(taskIDs))
	if len(taskIDs) == 0 {
		return out, nil
	}
	rows, err := r.pool.Query(ctx, `
SELECT task_id, status, COALESCE(error_message, ''), COALESCE(note, ''), finished_at, updated_at
FROM cloud_download_file_deletions
WHERE task_id = ANY($1)`, taskIDs)
	if err != nil {
		return nil, fmt.Errorf("list cloud download file deletion summaries: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var taskID int64
		var item models.CloudDownloadFileDeletionSummary
		if err := rows.Scan(&taskID, &item.Status, &item.ErrorMessage, &item.Note, &item.FinishedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan cloud download file deletion summary: %w", err)
		}
		out[taskID] = item
	}
	return out, rows.Err()
}
