package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

func scanEd2kDownloadTask(row rowScanner) (models.AdminEd2kDownloadTask, error) {
	var out models.AdminEd2kDownloadTask
	var filesRaw, historyRaw []byte
	if err := row.Scan(
		&out.ID,
		&out.SourceLink,
		&out.ResourceHash,
		&out.Title,
		&out.Filename,
		&out.DeclaredSize,
		&out.Status,
		&out.ProgressText,
		&out.ErrorMessage,
		&out.OutputDir,
		&out.DownloadedPath,
		&out.RetryCount,
		&filesRaw,
		&historyRaw,
		&out.CreatedAt,
		&out.UpdatedAt,
		&out.StartedAt,
		&out.FinishedAt,
		&out.CleanedAt,
		&out.DeletedAt,
	); err != nil {
		return models.AdminEd2kDownloadTask{}, err
	}
	if len(filesRaw) > 0 {
		_ = json.Unmarshal(filesRaw, &out.Files)
	}
	if len(historyRaw) > 0 {
		_ = json.Unmarshal(historyRaw, &out.History)
	}
	if out.Files == nil {
		out.Files = []models.AdminEd2kDownloadTaskFile{}
	}
	if out.History == nil {
		out.History = []models.AdminEd2kDownloadTaskHistoryItem{}
	}
	return out, nil
}

func ed2kDownloadTaskHistory(kind, label, message string) []models.AdminEd2kDownloadTaskHistoryItem {
	return []models.AdminEd2kDownloadTaskHistoryItem{
		{
			Kind:    kind,
			Label:   label,
			Message: message,
			At:      time.Now(),
		},
	}
}

func (r *VideoRepository) ListEd2kDownloadTasks(ctx context.Context, status string, page, pageSize int) ([]models.AdminEd2kDownloadTask, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 4)
	next := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}
	if s := strings.TrimSpace(strings.ToLower(status)); s != "" && s != "all" {
		where = append(where, "status = "+next(s))
	}
	baseWhere := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM ed2k_download_tasks WHERE "+baseWhere, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count ed2k download tasks: %w", err)
	}
	args = append(args, pageSize, (page-1)*pageSize)
	sql := `
SELECT id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
FROM ed2k_download_tasks
WHERE ` + baseWhere + `
	ORDER BY CASE WHEN status IN ('queued','running','canceling') THEN 0 ELSE 1 END,
	         CASE WHEN status IN ('completed','files_cleaned','cancelled') THEN COALESCE(finished_at, updated_at, created_at)
	              ELSE updated_at END DESC,
	         created_at DESC
LIMIT $` + fmt.Sprintf("%d", len(args)-1) + ` OFFSET $` + fmt.Sprintf("%d", len(args))
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list ed2k download tasks: %w", err)
	}
	defer rows.Close()

	items := make([]models.AdminEd2kDownloadTask, 0, pageSize)
	for rows.Next() {
		item, scanErr := scanEd2kDownloadTask(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("scan ed2k download task: %w", scanErr)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *VideoRepository) GetEd2kDownloadTask(ctx context.Context, id uuid.UUID) (models.AdminEd2kDownloadTask, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
FROM ed2k_download_tasks
WHERE id = $1
`, id)
	item, err := scanEd2kDownloadTask(row)
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("get ed2k download task: %w", err)
	}
	return item, nil
}

func (r *VideoRepository) GetEd2kDownloadTaskByHash(ctx context.Context, resourceHash string) (models.AdminEd2kDownloadTask, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
FROM ed2k_download_tasks
WHERE resource_hash = $1 AND status NOT IN ('deleted', 'cancelled')
`, strings.TrimSpace(resourceHash))
	item, err := scanEd2kDownloadTask(row)
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("get ed2k download task by hash: %w", err)
	}
	return item, nil
}

func (r *VideoRepository) CreateEd2kDownloadTask(ctx context.Context, input models.AdminEd2kDownloadTask) (models.AdminEd2kDownloadTask, error) {
	now := time.Now()
	if input.CreatedAt.IsZero() {
		input.CreatedAt = now
	}
	input.UpdatedAt = input.CreatedAt
	if len(input.History) == 0 {
		input.History = ed2kDownloadTaskHistory("created", "已创建", "任务已加入工作台")
	}
	filesRaw, err := json.Marshal(input.Files)
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k download task files: %w", err)
	}
	historyRaw, err := json.Marshal(input.History)
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k download task history: %w", err)
	}
	row := r.pool.QueryRow(ctx, `
INSERT INTO ed2k_download_tasks (
  id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20
)
RETURNING id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
`, input.ID, input.SourceLink, input.ResourceHash, input.Title, input.Filename, input.DeclaredSize, input.Status, input.ProgressText, input.ErrorMessage, input.OutputDir, input.DownloadedPath, input.RetryCount, filesRaw, historyRaw, input.CreatedAt, input.UpdatedAt, input.StartedAt, input.FinishedAt, input.CleanedAt, input.DeletedAt)
	item, scanErr := scanEd2kDownloadTask(row)
	if scanErr != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("create ed2k download task: %w", scanErr)
	}
	return item, nil
}

func (r *VideoRepository) AppendEd2kDownloadTaskHistory(ctx context.Context, id uuid.UUID, item models.AdminEd2kDownloadTaskHistoryItem) error {
	raw, err := json.Marshal([]models.AdminEd2kDownloadTaskHistoryItem{item})
	if err != nil {
		return fmt.Errorf("marshal ed2k download task history item: %w", err)
	}
	tag, err := r.pool.Exec(ctx, `
UPDATE ed2k_download_tasks
SET history = COALESCE(history, '[]'::jsonb) || $2::jsonb,
    updated_at = NOW()
WHERE id = $1
`, id, raw)
	if err != nil {
		return fmt.Errorf("append ed2k download task history: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *VideoRepository) UpdateEd2kDownloadTaskStatus(ctx context.Context, id uuid.UUID, status, progressText, errorMessage string, startedAt, finishedAt, deletedAt *time.Time, outputDir, downloadedPath string, files []models.AdminEd2kDownloadTaskFile, retryDelta int, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	filesRaw, err := json.Marshal(files)
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k download task files: %w", err)
	}
	historyRaw, err := json.Marshal([]models.AdminEd2kDownloadTaskHistoryItem{history})
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k download task history: %w", err)
	}
	row := r.pool.QueryRow(ctx, `
UPDATE ed2k_download_tasks
SET status = $2,
    progress_text = $3,
    error_message = $4,
    started_at = COALESCE($5, started_at),
    finished_at = $6,
    deleted_at = $7,
    output_dir = COALESCE($8, output_dir),
    downloaded_path = COALESCE($9, downloaded_path),
    files = COALESCE($10::jsonb, files),
    retry_count = retry_count + $11,
    history = COALESCE(history, '[]'::jsonb) || $12::jsonb,
    updated_at = NOW()
WHERE id = $1
RETURNING id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
`, id, status, progressText, errorMessage, startedAt, finishedAt, deletedAt, outputDir, downloadedPath, filesRaw, retryDelta, historyRaw)
	item, scanErr := scanEd2kDownloadTask(row)
	if scanErr != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("update ed2k download task status: %w", scanErr)
	}
	return item, nil
}

func (r *VideoRepository) MarkEd2kDownloadTaskRunning(ctx context.Context, id uuid.UUID, progressText string, history models.AdminEd2kDownloadTaskHistoryItem) error {
	historyRaw, err := json.Marshal([]models.AdminEd2kDownloadTaskHistoryItem{history})
	if err != nil {
		return fmt.Errorf("marshal ed2k download task history: %w", err)
	}
	row := r.pool.QueryRow(ctx, `
UPDATE ed2k_download_tasks
SET status = 'running',
    progress_text = $2,
    error_message = '',
    started_at = COALESCE($3, started_at),
    cleaned_at = NULL,
    updated_at = NOW(),
    history = COALESCE(history, '[]'::jsonb) || $4::jsonb
WHERE id = $1 AND status = 'queued'
RETURNING id
`, id, progressText, timePtr(time.Now()), historyRaw)
	var ignored uuid.UUID
	if scanErr := row.Scan(&ignored); scanErr != nil {
		return fmt.Errorf("mark ed2k download task running: %w", scanErr)
	}
	return nil
}

func (r *VideoRepository) MarkEd2kDownloadTaskCanceling(ctx context.Context, id uuid.UUID, progressText string, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	historyRaw, err := json.Marshal([]models.AdminEd2kDownloadTaskHistoryItem{history})
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k download task history: %w", err)
	}
	row := r.pool.QueryRow(ctx, `
UPDATE ed2k_download_tasks
SET status = 'canceling',
    progress_text = $2,
    error_message = '',
    updated_at = NOW(),
    history = COALESCE(history, '[]'::jsonb) || $3::jsonb
WHERE id = $1 AND status = 'running'
RETURNING id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
`, id, progressText, historyRaw)
	item, scanErr := scanEd2kDownloadTask(row)
	if scanErr != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("mark ed2k download task canceling: %w", scanErr)
	}
	return item, nil
}

func (r *VideoRepository) MarkEd2kDownloadTaskCancellationFailed(ctx context.Context, id uuid.UUID, progressText, errorMessage string, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	historyRaw, err := json.Marshal([]models.AdminEd2kDownloadTaskHistoryItem{history})
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k download task history: %w", err)
	}
	row := r.pool.QueryRow(ctx, `
UPDATE ed2k_download_tasks
SET status = 'canceling',
    progress_text = $2,
    error_message = $3,
    updated_at = NOW(),
    history = COALESCE(history, '[]'::jsonb) || $4::jsonb
WHERE id = $1 AND status = 'canceling'
RETURNING id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
`, id, progressText, errorMessage, historyRaw)
	item, scanErr := scanEd2kDownloadTask(row)
	if scanErr != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("mark ed2k download task cancellation failed: %w", scanErr)
	}
	return item, nil
}

func (r *VideoRepository) MarkEd2kDownloadTaskFailed(ctx context.Context, id uuid.UUID, errorMessage string, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	historyRaw, err := json.Marshal([]models.AdminEd2kDownloadTaskHistoryItem{history})
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k download task history: %w", err)
	}
	row := r.pool.QueryRow(ctx, `
UPDATE ed2k_download_tasks
SET status = 'failed',
    progress_text = '下载失败',
    error_message = $2,
    finished_at = $3,
    updated_at = NOW(),
    history = COALESCE(history, '[]'::jsonb) || $4::jsonb
WHERE id = $1 AND status = 'running'
RETURNING id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
`, id, errorMessage, timePtr(time.Now()), historyRaw)
	item, scanErr := scanEd2kDownloadTask(row)
	if scanErr != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("mark ed2k download task failed: %w", scanErr)
	}
	return item, nil
}

func (r *VideoRepository) MarkEd2kDownloadTaskCancelled(ctx context.Context, id uuid.UUID, progressText, errorMessage string, finishedAt, cleanedAt *time.Time, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	historyRaw, err := json.Marshal([]models.AdminEd2kDownloadTaskHistoryItem{history})
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k download task history: %w", err)
	}
	row := r.pool.QueryRow(ctx, `
UPDATE ed2k_download_tasks
SET status = 'cancelled',
    progress_text = $2,
    error_message = $3,
    finished_at = COALESCE($4, finished_at),
    cleaned_at = $5,
    updated_at = NOW(),
    history = COALESCE(history, '[]'::jsonb) || $6::jsonb
WHERE id = $1 AND status = 'canceling'
RETURNING id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
`, id, progressText, errorMessage, finishedAt, cleanedAt, historyRaw)
	item, scanErr := scanEd2kDownloadTask(row)
	if scanErr != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("mark ed2k download task cancelled: %w", scanErr)
	}
	return item, nil
}

func (r *VideoRepository) MarkEd2kDownloadTaskCompleted(ctx context.Context, id uuid.UUID, outputDir, downloadedPath string, files []models.AdminEd2kDownloadTaskFile, progressText string, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	if strings.TrimSpace(progressText) == "" {
		progressText = "下载已完成"
	}
	filesRaw, err := json.Marshal(files)
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k download task files: %w", err)
	}
	historyRaw, err := json.Marshal([]models.AdminEd2kDownloadTaskHistoryItem{history})
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k download task history: %w", err)
	}
	row := r.pool.QueryRow(ctx, `
UPDATE ed2k_download_tasks
SET status = 'completed',
    progress_text = $2,
    error_message = '',
    finished_at = $3,
    output_dir = COALESCE($4, output_dir),
    downloaded_path = COALESCE($5, downloaded_path),
    files = $6::jsonb,
    updated_at = NOW(),
    history = COALESCE(history, '[]'::jsonb) || $7::jsonb
WHERE id = $1 AND status = 'running'
RETURNING id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
`, id, progressText, timePtr(time.Now()), outputDir, downloadedPath, filesRaw, historyRaw)
	item, scanErr := scanEd2kDownloadTask(row)
	if scanErr != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("mark ed2k download task completed: %w", scanErr)
	}
	return item, nil
}

func (r *VideoRepository) MarkEd2kDownloadTaskCancelledCleanupResolved(ctx context.Context, id uuid.UUID, cleanedAt *time.Time, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	historyRaw, err := json.Marshal([]models.AdminEd2kDownloadTaskHistoryItem{history})
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k download task history: %w", err)
	}
	row := r.pool.QueryRow(ctx, `
UPDATE ed2k_download_tasks
SET progress_text = '任务已取消',
    error_message = '',
    cleaned_at = COALESCE($2, cleaned_at),
    updated_at = NOW(),
    history = COALESCE(history, '[]'::jsonb) || $3::jsonb
WHERE id = $1 AND status = 'cancelled'
RETURNING id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
`, id, cleanedAt, historyRaw)
	item, scanErr := scanEd2kDownloadTask(row)
	if scanErr != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("mark ed2k download task cancelled cleanup resolved: %w", scanErr)
	}
	return item, nil
}

func (r *VideoRepository) MarkEd2kDownloadTaskFilesCleaned(ctx context.Context, id uuid.UUID, progressText string, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	if strings.TrimSpace(progressText) == "" {
		progressText = "暂存文件已清理"
	}
	filesRaw, err := json.Marshal([]models.AdminEd2kDownloadTaskFile{})
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k cleaned files: %w", err)
	}
	historyRaw, err := json.Marshal([]models.AdminEd2kDownloadTaskHistoryItem{history})
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k download task history: %w", err)
	}
	row := r.pool.QueryRow(ctx, `
UPDATE ed2k_download_tasks
SET status = 'files_cleaned',
    progress_text = $2,
    error_message = '',
    downloaded_path = '',
    files = $3::jsonb,
    cleaned_at = NOW(),
    updated_at = NOW(),
    history = COALESCE(history, '[]'::jsonb) || $4::jsonb
WHERE id = $1
RETURNING id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
`, id, progressText, filesRaw, historyRaw)
	item, scanErr := scanEd2kDownloadTask(row)
	if scanErr != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("mark ed2k download task files cleaned: %w", scanErr)
	}
	return item, nil
}

func (r *VideoRepository) RequeueFilesCleanedEd2kDownloadTask(ctx context.Context, id uuid.UUID, history models.AdminEd2kDownloadTaskHistoryItem, startedAt *time.Time) (models.AdminEd2kDownloadTask, error) {
	filesRaw, err := json.Marshal([]models.AdminEd2kDownloadTaskFile{})
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k retry files: %w", err)
	}
	historyRaw, err := json.Marshal([]models.AdminEd2kDownloadTaskHistoryItem{history})
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k download task history: %w", err)
	}
	row := r.pool.QueryRow(ctx, `
UPDATE ed2k_download_tasks
SET status = 'queued',
    progress_text = '等待执行器接管',
    error_message = '',
    started_at = COALESCE($4, started_at),
    downloaded_path = '',
    files = $2::jsonb,
    cleaned_at = NULL,
    retry_count = retry_count + 1,
    updated_at = NOW(),
    history = COALESCE(history, '[]'::jsonb) || $3::jsonb
WHERE id = $1 AND status = 'files_cleaned'
RETURNING id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
`, id, filesRaw, historyRaw, startedAt)
	item, scanErr := scanEd2kDownloadTask(row)
	if scanErr != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("requeue files_cleaned ed2k download task: %w", scanErr)
	}
	return item, nil
}

func (r *VideoRepository) RestoreFilesCleanedEd2kDownloadTask(ctx context.Context, id uuid.UUID, history models.AdminEd2kDownloadTaskHistoryItem, startedAt, cleanedAt *time.Time) (models.AdminEd2kDownloadTask, error) {
	historyRaw, err := json.Marshal([]models.AdminEd2kDownloadTaskHistoryItem{history})
	if err != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("marshal ed2k download task history: %w", err)
	}
	row := r.pool.QueryRow(ctx, `
UPDATE ed2k_download_tasks
SET status = 'files_cleaned',
    progress_text = '已下载，暂存文件已清理',
    error_message = '',
    started_at = $2,
    cleaned_at = $3,
    retry_count = GREATEST(retry_count - 1, 0),
    updated_at = NOW(),
    history = COALESCE(history, '[]'::jsonb) || $4::jsonb
WHERE id = $1 AND status = 'queued'
RETURNING id, source_link, resource_hash, title, filename, declared_size, status, progress_text, error_message, output_dir, downloaded_path, retry_count, files, history, created_at, updated_at, started_at, finished_at, cleaned_at, deleted_at
`, id, startedAt, cleanedAt, historyRaw)
	item, scanErr := scanEd2kDownloadTask(row)
	if scanErr != nil {
		return models.AdminEd2kDownloadTask{}, fmt.Errorf("restore files_cleaned ed2k download task: %w", scanErr)
	}
	return item, nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func (r *VideoRepository) DeleteEd2kDownloadTask(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM ed2k_download_tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete ed2k download task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// HasActiveEd2kDownloadTasks 返回是否存在 queued/running/canceling 状态的 ED2K 下载任务。
// 供 server.met 定时刷新任务判定在途下载：重启 amuled 会打断在途下载，故任一在途即跳过本轮刷新。
func (r *VideoRepository) HasActiveEd2kDownloadTasks(ctx context.Context) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM ed2k_download_tasks WHERE status IN ('queued','running','canceling')`,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("count active ed2k download tasks: %w", err)
	}
	return count > 0, nil
}
