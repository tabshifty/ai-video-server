package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

// cloudLocalFileTaskRepository is the persistence surface for explicit local
// staging deletion. *repository.CloudDownloadRepository implements it.
type cloudLocalFileTaskRepository interface {
	GetByID(ctx context.Context, id int64) (models.CloudDownloadTask, error)
	MarkLocalFileDeleted(ctx context.Context, id int64, deletedBy *uuid.UUID) error
}

// cloudLocalFileImportRepository tells the service whether an import still owns
// the downloaded file.
type cloudLocalFileImportRepository interface {
	GetByTaskID(ctx context.Context, taskID int64) (models.CloudDownloadImport, error)
}

// CloudLocalFileDeletionService deletes the DOWNLOAD_DIR file of a completed
// cloud download. It never touches the 115 cloud file or the task record:
// deleting local staging is orthogonal to both.
type CloudLocalFileDeletionService struct {
	tasks       cloudLocalFileTaskRepository
	imports     cloudLocalFileImportRepository
	downloadDir string
	logger      *slog.Logger
}

// NewCloudLocalFileDeletionService builds the service.
func NewCloudLocalFileDeletionService(
	tasks cloudLocalFileTaskRepository,
	imports cloudLocalFileImportRepository,
	downloadDir string,
	logger *slog.Logger,
) *CloudLocalFileDeletionService {
	return &CloudLocalFileDeletionService{
		tasks:       tasks,
		imports:     imports,
		downloadDir: downloadDir,
		logger:      logger,
	}
}

// Preview reports the recorded path and the actual file size without deleting
// anything. A path outside DOWNLOAD_DIR is reported as not existing; Delete
// refuses it with a typed error.
func (s *CloudLocalFileDeletionService) Preview(ctx context.Context, taskID int64) (models.CloudLocalFilePreview, error) {
	task, err := s.loadTask(ctx, taskID)
	if err != nil {
		return models.CloudLocalFilePreview{}, err
	}
	preview := models.CloudLocalFilePreview{
		Path:    strings.TrimSpace(task.LocalPath),
		Deleted: task.LocalFileDeletedAt != nil,
	}
	if preview.Path == "" {
		return preview, nil
	}
	path, err := resolveManagedLocalPath(s.downloadDir, preview.Path)
	if err != nil {
		return preview, nil
	}
	preview.Managed = true
	if info, statErr := os.Stat(path); statErr == nil && info.Mode().IsRegular() {
		preview.Exists = true
		preview.Size = info.Size()
	}
	return preview, nil
}

// Delete removes the downloaded file once and records who discarded it. An
// already missing file is an idempotent success so stale rows can be cleaned.
func (s *CloudLocalFileDeletionService) Delete(ctx context.Context, taskID int64, requestedBy *uuid.UUID) (models.CloudDownloadTask, error) {
	task, err := s.loadTask(ctx, taskID)
	if err != nil {
		return models.CloudDownloadTask{}, err
	}
	if task.Status != models.CloudDownloadCompleted {
		return task, models.ErrCloudLocalFileNotCompleted
	}
	if task.LocalFileDeletedAt != nil {
		return task, nil
	}
	if strings.TrimSpace(task.LocalPath) == "" {
		return task, models.ErrCloudLocalFileNotRecorded
	}
	if err := s.ensureNoActiveImport(ctx, task.ID); err != nil {
		return task, err
	}
	size, err := s.removeManagedFile(task.LocalPath)
	if err != nil {
		return task, err
	}
	// The file is already gone; finish the bookkeeping even when the client
	// disconnects, mirroring the import and 115 cleanup services.
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := s.tasks.MarkLocalFileDeleted(writeCtx, task.ID, requestedBy); err != nil {
		return task, err
	}
	if s.logger != nil {
		s.logger.Info("cloud download local file deleted", "task_id", task.ID, "path", task.LocalPath, "size", size)
	}
	updated, err := s.tasks.GetByID(writeCtx, task.ID)
	if err != nil {
		// The file and the marker are already written; a failed reload must not
		// turn a successful deletion into a user-visible failure.
		if s.logger != nil {
			s.logger.Warn("reload cloud download task after local deletion failed", "task_id", task.ID, "error", err)
		}
		now := time.Now().UTC()
		task.LocalFileDeletedAt = &now
		task.LocalFileDeletedBy = requestedBy
		return task, nil
	}
	return updated, nil
}

func (s *CloudLocalFileDeletionService) loadTask(ctx context.Context, taskID int64) (models.CloudDownloadTask, error) {
	task, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.CloudDownloadTask{}, models.ErrCloudLocalFileNotFound
		}
		return models.CloudDownloadTask{}, fmt.Errorf("读取云下载任务: %w", err)
	}
	return task, nil
}

func (s *CloudLocalFileDeletionService) ensureNoActiveImport(ctx context.Context, taskID int64) error {
	if s.imports == nil {
		return nil
	}
	item, err := s.imports.GetByTaskID(ctx, taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("读取云导入状态: %w", err)
	}
	if models.CloudImportActive(item.Status) {
		return models.ErrCloudLocalFileImportActive
	}
	return nil
}

// removeManagedFile deletes the file when it still exists and returns its size.
// A missing file is not an error; a path outside DOWNLOAD_DIR or a non-regular
// target is refused.
func (s *CloudLocalFileDeletionService) removeManagedFile(localPath string) (int64, error) {
	if strings.TrimSpace(localPath) == "" {
		return 0, nil
	}
	path, err := resolveManagedLocalPath(s.downloadDir, localPath)
	if err != nil {
		return 0, err
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("读取本地文件: %w", err)
	}
	if !info.Mode().IsRegular() {
		return 0, models.ErrCloudLocalFileUnavailable
	}
	if err := os.Remove(path); err != nil {
		return 0, fmt.Errorf("删除本地文件: %w", err)
	}
	return info.Size(), nil
}

// resolveManagedLocalPath validates that localPath stays inside downloadDir.
// The target may not exist anymore; its parent chain is resolved instead so a
// symlinked download root still compares correctly.
func resolveManagedLocalPath(downloadDir, localPath string) (string, error) {
	root := strings.TrimSpace(downloadDir)
	if root == "" {
		return "", models.ErrCloudLocalFileUnavailable
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", models.ErrCloudLocalFileUnavailable
	}
	if resolved, resolveErr := filepath.EvalSymlinks(rootAbs); resolveErr == nil {
		rootAbs = resolved
	}
	targetAbs, err := filepath.Abs(strings.TrimSpace(localPath))
	if err != nil {
		return "", models.ErrCloudLocalFileUnavailable
	}
	if resolved, resolveErr := filepath.EvalSymlinks(targetAbs); resolveErr == nil {
		targetAbs = resolved
	} else if parent, parentErr := filepath.EvalSymlinks(filepath.Dir(targetAbs)); parentErr == nil {
		// A missing file cannot be resolved directly; resolving its parent
		// keeps containment correct when the download root is a symlink (for
		// example ./downloads -> /Volumes/large/...), so "already deleted"
		// stays an idempotent success instead of a false path error.
		targetAbs = filepath.Join(parent, filepath.Base(targetAbs))
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil || rel == "." || rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", models.ErrCloudLocalFileUnavailable
	}
	return targetAbs, nil
}
