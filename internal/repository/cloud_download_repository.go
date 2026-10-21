package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"video-server/internal/models"
)

// CloudDownloadRepository persists cloud download tasks.
type CloudDownloadRepository struct {
	pool *pgxpool.Pool
}

// NewCloudDownloadRepository builds the repository.
func NewCloudDownloadRepository(pool *pgxpool.Pool) *CloudDownloadRepository {
	return &CloudDownloadRepository{pool: pool}
}

const cloudDownloadSelect = `
SELECT
    t.id,
    t.post_id,
    COALESCE(p.title, ''),
    t.source,
    t.protocol,
    t.info_hash,
    t.source_url,
    COALESCE(t.declared_name, ''),
    t.declared_size,
    t.status,
    COALESCE(t.blocked_reason, ''),
    COALESCE(t.error_message, ''),
    t.retry_count,
    COALESCE(t.pick_code, ''),
    COALESCE(t.cloud_file_id, ''),
    COALESCE(t.cloud_parent_id, ''),
    COALESCE(t.cloud_user_id, ''),
    COALESCE(t.cloud_file_category, 0),
    COALESCE(t.cloud_status, ''),
    t.progress_percent,
    t.progress_updated_at,
    COALESCE(t.local_path, ''),
    t.download_triggered_at,
    t.created_by,
    t.created_at,
    t.updated_at,
    t.started_at,
    t.finished_at
FROM cloud_download_tasks t
LEFT JOIN collected_forum_posts p ON p.id = t.post_id`

func scanCloudDownload(row interface{ Scan(...any) error }) (models.CloudDownloadTask, error) {
	var task models.CloudDownloadTask
	err := row.Scan(
		&task.ID,
		&task.PostID,
		&task.PostTitle,
		&task.Source,
		&task.Protocol,
		&task.InfoHash,
		&task.SourceURL,
		&task.DeclaredName,
		&task.DeclaredSize,
		&task.Status,
		&task.BlockedReason,
		&task.ErrorMessage,
		&task.RetryCount,
		&task.PickCode,
		&task.CloudFileID,
		&task.CloudParentID,
		&task.CloudUserID,
		&task.CloudFileCategory,
		&task.CloudStatus,
		&task.ProgressPercent,
		&task.ProgressUpdatedAt,
		&task.LocalPath,
		&task.DownloadTriggeredAt,
		&task.CreatedBy,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.StartedAt,
		&task.FinishedAt,
	)
	return task, err
}

// CreateOrGet inserts a task unless the info hash is already tracked. The
// bool reports whether this call created the row.
func (r *CloudDownloadRepository) CreateOrGet(ctx context.Context, task models.CloudDownloadTask) (models.CloudDownloadTask, bool, error) {
	source := strings.TrimSpace(task.Source)
	if source == "" {
		// Match the column default; the service normally resolves the source
		// before reaching the repository.
		source = models.CloudDownloadSourceForumPost
	}
	tag, err := r.pool.Exec(ctx, `
INSERT INTO cloud_download_tasks (post_id, source, protocol, info_hash, source_url, declared_name, declared_size, status, created_by)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9)
ON CONFLICT (info_hash) DO NOTHING`,
		task.PostID, source, task.Protocol, task.InfoHash, task.SourceURL, task.DeclaredName, task.DeclaredSize, task.Status, task.CreatedBy)
	if err != nil {
		return models.CloudDownloadTask{}, false, fmt.Errorf("insert cloud download task: %w", err)
	}
	existing, err := r.GetByInfoHash(ctx, task.InfoHash)
	if err != nil {
		return models.CloudDownloadTask{}, false, err
	}
	return existing, tag.RowsAffected() > 0, nil
}

// GetByInfoHash returns a task by its normalized info hash.
func (r *CloudDownloadRepository) GetByInfoHash(ctx context.Context, infoHash string) (models.CloudDownloadTask, error) {
	row := r.pool.QueryRow(ctx, cloudDownloadSelect+` WHERE t.info_hash = $1`, infoHash)
	task, err := scanCloudDownload(row)
	if err != nil {
		return models.CloudDownloadTask{}, fmt.Errorf("get cloud download task by hash: %w", err)
	}
	return task, nil
}

// GetByID returns a task by primary key.
func (r *CloudDownloadRepository) GetByID(ctx context.Context, id int64) (models.CloudDownloadTask, error) {
	row := r.pool.QueryRow(ctx, cloudDownloadSelect+` WHERE t.id = $1`, id)
	task, err := scanCloudDownload(row)
	if err != nil {
		return models.CloudDownloadTask{}, fmt.Errorf("get cloud download task: %w", err)
	}
	return task, nil
}

// List returns a filtered page plus the total count.
func (r *CloudDownloadRepository) List(ctx context.Context, status string, page, pageSize int) ([]models.CloudDownloadTask, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	where := "WHERE 1=1"
	args := []any{}
	if status != "" {
		args = append(args, status)
		where += fmt.Sprintf(" AND t.status = $%d", len(args))
	}

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM cloud_download_tasks t "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count cloud download tasks: %w", err)
	}

	args = append(args, pageSize, (page-1)*pageSize)
	query := fmt.Sprintf("%s %s ORDER BY t.id DESC LIMIT $%d OFFSET $%d", cloudDownloadSelect, where, len(args)-1, len(args))
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list cloud download tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]models.CloudDownloadTask, 0, pageSize)
	for rows.Next() {
		task, err := scanCloudDownload(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan cloud download task: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

// SetStatus updates the lifecycle status and the human-facing reason fields.
func (r *CloudDownloadRepository) SetStatus(ctx context.Context, id int64, status, blockedReason, errorMessage string) error {
	_, err := r.pool.Exec(ctx, `
UPDATE cloud_download_tasks SET
    status = $2,
    blocked_reason = NULLIF($3, ''),
    error_message = NULLIF($4, ''),
    started_at = CASE WHEN $2 IN ('submitting', 'queued', 'downloading') AND started_at IS NULL THEN now() ELSE started_at END,
    finished_at = CASE WHEN $2 IN ('completed', 'failed') THEN now() ELSE finished_at END,
    updated_at = now()
WHERE id = $1`, id, status, blockedReason, errorMessage)
	if err != nil {
		return fmt.Errorf("set cloud download status: %w", err)
	}
	return nil
}

// SetProgress records the 115-side percentage.
func (r *CloudDownloadRepository) SetProgress(ctx context.Context, id int64, percent float64, cloudStatus string) error {
	_, err := r.pool.Exec(ctx, `
UPDATE cloud_download_tasks SET
    progress_percent = $2,
    cloud_status = NULLIF($3, ''),
    progress_updated_at = now(),
    updated_at = now()
WHERE id = $1`, id, percent, cloudStatus)
	if err != nil {
		return fmt.Errorf("set cloud download progress: %w", err)
	}
	return nil
}

// SetReady records the pick code and the actual 115 cloud file identity once
// 115 reports the file is available.
func (r *CloudDownloadRepository) SetReady(ctx context.Context, id int64, pickCode, cloudFileID, cloudParentID, cloudUserID string, cloudFileCategory int) error {
	_, err := r.pool.Exec(ctx, `
UPDATE cloud_download_tasks SET
    pick_code = COALESCE(NULLIF($2, ''), pick_code),
    cloud_file_id = COALESCE(NULLIF($3, ''), cloud_file_id),
    cloud_parent_id = COALESCE(NULLIF($4, ''), cloud_parent_id),
    cloud_user_id = COALESCE(NULLIF($5, ''), cloud_user_id),
    cloud_file_category = CASE WHEN $6 > 0 THEN $6 ELSE cloud_file_category END,
    status = $7,
    progress_percent = 100,
    progress_updated_at = now(),
    updated_at = now()
WHERE id = $1`, id, pickCode, cloudFileID, cloudParentID, cloudUserID, cloudFileCategory, models.CloudDownloadReady)
	if err != nil {
		return fmt.Errorf("set cloud download ready: %w", err)
	}
	return nil
}

// SetCloudFileIDs stores the lazily resolved 115 delete_file_id/wp_path_id for
// historical tasks that predate these columns.
func (r *CloudDownloadRepository) SetCloudFileIDs(ctx context.Context, id int64, cloudFileID, cloudParentID, cloudUserID string, cloudFileCategory int) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE cloud_download_tasks SET
    cloud_file_id = COALESCE(NULLIF($2, ''), cloud_file_id),
    cloud_parent_id = COALESCE(NULLIF($3, ''), cloud_parent_id),
    cloud_user_id = COALESCE(NULLIF($4, ''), cloud_user_id),
    cloud_file_category = CASE WHEN $5 > 0 THEN $5 ELSE cloud_file_category END,
    updated_at = now()
WHERE id = $1
  AND (cloud_file_id IS NULL OR cloud_file_id = '' OR cloud_file_id = $2)`, id, cloudFileID, cloudParentID, cloudUserID, cloudFileCategory)
	if err != nil {
		return false, fmt.Errorf("set cloud download file ids: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// SetLocalPath records the downloaded file path and completes the task.
func (r *CloudDownloadRepository) SetLocalPath(ctx context.Context, id int64, localPath string) error {
	_, err := r.pool.Exec(ctx, `
UPDATE cloud_download_tasks SET
    local_path = NULLIF($2, ''),
    status = $3,
    finished_at = now(),
    updated_at = now()
WHERE id = $1`, id, localPath, models.CloudDownloadCompleted)
	if err != nil {
		return fmt.Errorf("set cloud download local path: %w", err)
	}
	return nil
}

// ClaimDownloadTrigger atomically claims the one-shot native download trigger.
// A redelivered or concurrent job that loses the claim must not trigger again.
func (r *CloudDownloadRepository) ClaimDownloadTrigger(ctx context.Context, id int64) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE cloud_download_tasks SET
    download_triggered_at = now(),
    updated_at = now()
WHERE id = $1 AND download_triggered_at IS NULL`, id)
	if err != nil {
		return false, fmt.Errorf("claim cloud download trigger: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// ClearDownloadTrigger removes the claim after the trigger call itself failed,
// so a retry can attempt the trigger again.
func (r *CloudDownloadRepository) ClearDownloadTrigger(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `
UPDATE cloud_download_tasks SET
    download_triggered_at = NULL,
    updated_at = now()
WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("clear cloud download trigger: %w", err)
	}
	return nil
}

// ListUnfinished returns tasks that may still need a worker job. Used by the
// startup/periodic reconciliation; terminal and manually blocked rows are
// intentionally excluded. The retry_count is needed to build the asynq task
// ID: a manual retry bumps it and therefore gets a fresh unique ID.
func (r *CloudDownloadRepository) ListUnfinished(ctx context.Context, limit int) ([]models.CloudDownloadTask, error) {
	if limit <= 0 || limit > 500 {
		limit = 500
	}
	query := cloudDownloadSelect + `
 WHERE t.status IN ('pending', 'submitting', 'queued', 'ready', 'downloading')
 ORDER BY t.updated_at ASC
 LIMIT $1`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("list unfinished cloud download tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]models.CloudDownloadTask, 0, limit)
	for rows.Next() {
		task, err := scanCloudDownload(rows)
		if err != nil {
			return nil, fmt.Errorf("scan unfinished cloud download task: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

// BumpRetry increments the retry counter and requeues a failed/blocked task.
// The bool reports whether this call won the conditional update; a concurrent
// retry that already moved the row must not enqueue a second job.
func (r *CloudDownloadRepository) BumpRetry(ctx context.Context, id int64) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE cloud_download_tasks SET
    retry_count = retry_count + 1,
    status = 'pending',
    blocked_reason = NULL,
    error_message = NULL,
    finished_at = NULL,
    download_triggered_at = NULL,
    updated_at = now()
WHERE id = $1 AND status IN ('failed', 'blocked')`, id)
	if err != nil {
		return false, fmt.Errorf("retry cloud download: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// StatusByHashes returns the compact status map used by the forum list.
func (r *CloudDownloadRepository) StatusByHashes(ctx context.Context, infoHashes []string) (map[string]models.CloudDownloadStatus, error) {
	out := make(map[string]models.CloudDownloadStatus, len(infoHashes))
	if len(infoHashes) == 0 {
		return out, nil
	}
	rows, err := r.pool.Query(ctx, `
SELECT info_hash, id, status, COALESCE(blocked_reason, '')
FROM cloud_download_tasks
WHERE info_hash = ANY($1)`, infoHashes)
	if err != nil {
		return nil, fmt.Errorf("cloud download statuses: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var hash string
		var status models.CloudDownloadStatus
		if err := rows.Scan(&hash, &status.TaskID, &status.Status, &status.BlockedReason); err != nil {
			return nil, fmt.Errorf("scan cloud download status: %w", err)
		}
		out[hash] = status
	}
	return out, rows.Err()
}
