package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/cloud"
	"video-server/internal/models"
)

const cloudFileDeletionStaleAfter = 5 * time.Minute

// cloudFileDeletionTaskRepository is the persistence surface needed to validate
// and lazily resolve a cloud-download task.
type cloudFileDeletionTaskRepository interface {
	GetByID(ctx context.Context, id int64) (models.CloudDownloadTask, error)
	SetCloudFileIDs(ctx context.Context, id int64, cloudFileID, cloudParentID, cloudUserID string, cloudFileCategory int) (bool, error)
}

// cloudFileDeletionImportRepository tells the deletion service whether the local
// source has already been consumed by a terminal media import. After import the
// local file may be gone while the media is safe, so that is not data loss.
type cloudFileDeletionImportRepository interface {
	GetByTaskID(ctx context.Context, taskID int64) (models.CloudDownloadImport, error)
}

// cloudFileDeletionRepository owns the one-per-task cleanup record.
type cloudFileDeletionRepository interface {
	GetByTaskID(ctx context.Context, taskID int64) (models.CloudDownloadFileDeletion, error)
	Start(ctx context.Context, taskID int64, fileID, parentID string, requestedBy *uuid.UUID) (models.CloudDownloadFileDeletion, bool, error)
	SetFileIDs(ctx context.Context, taskID int64, attempt int, fileID, parentID string) (bool, error)
	MarkDeleted(ctx context.Context, taskID int64, attempt int, note string) (bool, error)
	MarkFailed(ctx context.Context, taskID int64, attempt int, errorMessage string) (bool, error)
	ListSummariesByTaskIDs(ctx context.Context, taskIDs []int64) (map[int64]models.CloudDownloadFileDeletionSummary, error)
}

// cloudFileDeletionSession is one attached 115 browser session. It is separate
// from cloudSession so tests can exercise the cleanup flow without downloads.
type cloudFileDeletionSession interface {
	Close()
	UserID() string
	FindUniqueTaskDeep(ctx context.Context, infoHash string, maxPages int) (*cloud.Task, error)
	DeleteFile(ctx context.Context, parentID, fileID string) (cloud.DeleteFileResult, error)
}

type cloudFileDeletionSessionProvider func(ctx context.Context) (cloudFileDeletionSession, error)

// CloudFileDeletionService moves a completed task's 115 cloud file to the 115
// recycle bin. It is intentionally synchronous and separate from the download
// worker.
type CloudFileDeletionService struct {
	tasks         cloudFileDeletionTaskRepository
	imports       cloudFileDeletionImportRepository
	deletions     cloudFileDeletionRepository
	ensureSession cloudFileDeletionSessionProvider
	downloadDir   string
	logger        *slog.Logger
}

// NewCloudFileDeletionService builds the service.
func NewCloudFileDeletionService(
	tasks cloudFileDeletionTaskRepository,
	imports cloudFileDeletionImportRepository,
	deletions cloudFileDeletionRepository,
	client *cloud.Client,
	downloadDir string,
	logger *slog.Logger,
) *CloudFileDeletionService {
	svc := &CloudFileDeletionService{
		tasks:       tasks,
		imports:     imports,
		deletions:   deletions,
		downloadDir: downloadDir,
		logger:      logger,
	}
	if client != nil {
		svc.ensureSession = func(ctx context.Context) (cloudFileDeletionSession, error) {
			session, state, err := client.EnsureSession(ctx)
			if err != nil {
				return nil, err
			}
			return liveCloudFileDeletionSession{session: session, userID: state.UserID}, nil
		}
	}
	return svc
}

// DeleteCloudFile validates the task, claims one deletion attempt, resolves the
// actual delete_file_id when necessary, calls the authenticated 115 page, and
// records the outcome. A missing 115 file is an idempotent success.
func (s *CloudFileDeletionService) DeleteCloudFile(ctx context.Context, taskID int64, requestedBy *uuid.UUID) (models.CloudDownloadTask, error) {
	if s.tasks == nil || s.deletions == nil {
		return models.CloudDownloadTask{}, models.ErrCloudFileDeleteUnavailable
	}
	opCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	task, err := s.tasks.GetByID(opCtx, taskID)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.CloudDownloadTask{}, models.ErrCloudFileDeleteNotFound
	}
	if err != nil {
		return models.CloudDownloadTask{}, err
	}

	var existing models.CloudDownloadFileDeletion
	existingErr := error(nil)
	if s.deletions != nil {
		existing, existingErr = s.deletions.GetByTaskID(opCtx, taskID)
	}
	if existingErr == nil && existing.Status == models.CloudFileDeleted {
		return attachCloudFileDeletion(task, existing), nil
	}
	if existingErr == nil && existing.Status == models.CloudFileDeleting && time.Since(existing.UpdatedAt) < cloudFileDeletionStaleAfter {
		return models.CloudDownloadTask{}, models.ErrCloudFileDeleteInProgress
	}
	if existingErr != nil && !errors.Is(existingErr, models.ErrCloudFileDeletionNotFound) {
		return models.CloudDownloadTask{}, existingErr
	}

	if task.Status != models.CloudDownloadCompleted {
		return models.CloudDownloadTask{}, models.ErrCloudFileDeleteNotCompleted
	}
	if !s.localSourceOrImported(opCtx, task) {
		return models.CloudDownloadTask{}, models.ErrCloudFileDeleteLocalMissing
	}
	if s.ensureSession == nil {
		return models.CloudDownloadTask{}, models.ErrCloudFileDeleteBrowserDown
	}

	// Claim before touching the browser. Dependency failures still get a failed
	// record so the UI can retry instead of leaving a permanent placeholder.
	deletion, claimed, err := s.deletions.Start(opCtx, taskID, task.CloudFileID, task.CloudParentID, requestedBy)
	if err != nil {
		return models.CloudDownloadTask{}, err
	}
	if !claimed {
		if deletion.Status == models.CloudFileDeleted {
			return attachCloudFileDeletion(task, deletion), nil
		}
		return models.CloudDownloadTask{}, models.ErrCloudFileDeleteInProgress
	}

	session, err := s.ensureSession(opCtx)
	if err != nil {
		mapped := mapCloudFileDeletionSessionError(err)
		s.finishFailed(ctx, taskID, deletion.Attempt, cloudFileDeletionDependencyMessage(mapped))
		return models.CloudDownloadTask{}, mapped
	}
	defer session.Close()

	if task.CloudUserID != "" && task.CloudUserID != session.UserID() {
		message := "115 当前登录账号与任务创建账号不一致，已拒绝删除"
		s.finishFailed(ctx, taskID, deletion.Attempt, message)
		return models.CloudDownloadTask{}, models.ErrCloudFileDeleteUnavailable
	}

	fileID := strings.TrimSpace(task.CloudFileID)
	parentID := strings.TrimSpace(task.CloudParentID)
	category := task.CloudFileCategory
	if fileID == "" || parentID == "" || category == 0 {
		resolved, findErr := session.FindUniqueTaskDeep(opCtx, task.InfoHash, 5)
		if errors.Is(findErr, cloud.ErrAmbiguousTask) {
			s.finishFailed(ctx, taskID, deletion.Attempt, "115 中存在多个同 hash 任务，无法唯一定位源文件")
			return models.CloudDownloadTask{}, models.ErrCloudFileDeleteUnavailable
		}
		if findErr != nil {
			mapped := mapCloudFileDeletionSessionError(findErr)
			s.finishFailed(ctx, taskID, deletion.Attempt, "查询 115 任务失败: "+cloudFileDeletionDependencyMessage(mapped))
			if errors.Is(mapped, models.ErrCloudFileDeleteBrowserDown) ||
				errors.Is(mapped, models.ErrCloudFileDeleteLoginRequired) ||
				errors.Is(mapped, models.ErrCloudFileDeleteTimeout) {
				return models.CloudDownloadTask{}, mapped
			}
			return models.CloudDownloadTask{}, findErr
		}
		if resolved == nil || strings.TrimSpace(resolved.DeleteFileID) == "" || strings.TrimSpace(resolved.WPPathID) == "" {
			s.finishFailed(ctx, taskID, deletion.Attempt, models.ErrCloudFileDeleteUnavailable.Error())
			return models.CloudDownloadTask{}, models.ErrCloudFileDeleteUnavailable
		}
		if resolved.FileCategory > 0 && resolved.FileCategory != 1 {
			s.finishFailed(ctx, taskID, deletion.Attempt, "115 云端对象不是普通文件，已拒绝删除")
			return models.CloudDownloadTask{}, models.ErrCloudFileDeleteUnavailable
		}
		if task.DeclaredSize == nil || *task.DeclaredSize <= 0 || resolved.Size <= 0 || resolved.Size != *task.DeclaredSize {
			s.finishFailed(ctx, taskID, deletion.Attempt, "115 文件大小与任务记录不一致，已拒绝删除")
			return models.CloudDownloadTask{}, models.ErrCloudFileDeleteUnavailable
		}
		if !sameCloudFileName(task.DeclaredName, resolved.Name) {
			s.finishFailed(ctx, taskID, deletion.Attempt, "115 文件名与任务记录不一致，已拒绝删除")
			return models.CloudDownloadTask{}, models.ErrCloudFileDeleteUnavailable
		}
		fileID = strings.TrimSpace(resolved.DeleteFileID)
		parentID = strings.TrimSpace(resolved.WPPathID)
		category = resolved.FileCategory
		ok, setErr := s.tasks.SetCloudFileIDs(opCtx, taskID, fileID, parentID, session.UserID(), category)
		if setErr != nil {
			s.finishFailed(ctx, taskID, deletion.Attempt, setErr.Error())
			return models.CloudDownloadTask{}, setErr
		}
		if !ok {
			s.finishFailed(ctx, taskID, deletion.Attempt, "115 文件标识已被其他请求修改，已拒绝删除")
			return models.CloudDownloadTask{}, models.ErrCloudFileDeleteUnavailable
		}
	}
	if !isNumeric115ID(fileID) || !isNumeric115ID(parentID) {
		s.finishFailed(ctx, taskID, deletion.Attempt, "115 文件标识格式无效，已拒绝删除")
		return models.CloudDownloadTask{}, models.ErrCloudFileDeleteUnavailable
	}
	if category != 1 {
		message := "115 云端对象不是普通文件，仅支持单文件删除"
		s.finishFailed(ctx, taskID, deletion.Attempt, message)
		return models.CloudDownloadTask{}, models.ErrCloudFileDeleteUnavailable
	}
	fenced, err := s.deletions.SetFileIDs(opCtx, taskID, deletion.Attempt, fileID, parentID)
	if err != nil {
		return models.CloudDownloadTask{}, err
	}
	if !fenced {
		authoritative, getErr := s.deletions.GetByTaskID(opCtx, taskID)
		if getErr != nil {
			return models.CloudDownloadTask{}, getErr
		}
		if authoritative.Status == models.CloudFileDeleted {
			return attachCloudFileDeletion(task, authoritative), nil
		}
		return models.CloudDownloadTask{}, models.ErrCloudFileDeleteInProgress
	}
	task.CloudFileID = fileID
	task.CloudParentID = parentID
	task.CloudFileCategory = category

	result, err := session.DeleteFile(opCtx, parentID, fileID)
	if err != nil {
		mapped := mapCloudFileDeletionSessionError(err)
		s.finishFailed(ctx, taskID, deletion.Attempt, cloudFileDeletionDependencyMessage(mapped))
		if errors.Is(mapped, models.ErrCloudFileDeleteBrowserDown) ||
			errors.Is(mapped, models.ErrCloudFileDeleteLoginRequired) ||
			errors.Is(mapped, models.ErrCloudFileDeleteTimeout) {
			return models.CloudDownloadTask{}, mapped
		}
		return models.CloudDownloadTask{}, fmt.Errorf("115 云盘删除请求失败: %w", err)
	}
	if result.JSONParsed && result.State && result.HTTPStatus == 200 && strings.TrimSpace(result.Error) == "" && strings.TrimSpace(result.Errno) == "" && strings.TrimSpace(result.ErrType) == "" {
		item, finishErr := s.finishDeleted(ctx, taskID, deletion.Attempt, "")
		if finishErr != nil {
			return models.CloudDownloadTask{}, finishErr
		}
		s.logInfo("cloud download file moved to 115 recycle bin", "task_id", taskID, "file_id", fileID)
		return attachCloudFileDeletion(task, item), nil
	}
	if result.JSONParsed && result.HTTPStatus == 200 && !result.State && isCloudFileAlreadyMissing(result) {
		note := "115 侧文件已不存在"
		item, finishErr := s.finishDeleted(ctx, taskID, deletion.Attempt, note)
		if finishErr != nil {
			return models.CloudDownloadTask{}, finishErr
		}
		return attachCloudFileDeletion(task, item), nil
	}
	message := cloudFileDeletionFailureMessage(result)
	s.finishFailed(ctx, taskID, deletion.Attempt, message)
	return models.CloudDownloadTask{}, fmt.Errorf("%w: %s", models.ErrCloudFileDeleteFailed, message)
}

func (s *CloudFileDeletionService) localSourceOrImported(ctx context.Context, task models.CloudDownloadTask) bool {
	if _, _, err := resolveCloudImportSource(s.downloadDir, task.LocalPath); err == nil {
		return true
	}
	if s.imports == nil {
		return false
	}
	item, err := s.imports.GetByTaskID(ctx, task.ID)
	if err != nil {
		return false
	}
	return item.Status == models.CloudImportImported || item.Status == models.CloudImportSkippedExisting
}

func (s *CloudFileDeletionService) finishDeleted(ctx context.Context, taskID int64, attempt int, note string) (models.CloudDownloadFileDeletion, error) {
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	changed, err := s.deletions.MarkDeleted(writeCtx, taskID, attempt, note)
	if err != nil {
		return models.CloudDownloadFileDeletion{}, err
	}
	if !changed {
		return s.authoritativeDeletion(writeCtx, taskID, models.CloudFileDeleted)
	}
	return s.deletions.GetByTaskID(writeCtx, taskID)
}

func (s *CloudFileDeletionService) finishFailed(ctx context.Context, taskID int64, attempt int, message string) {
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	changed, err := s.deletions.MarkFailed(writeCtx, taskID, attempt, truncateMessage(message, 500))
	if err != nil {
		s.logWarn("mark cloud download file deletion failed", "task_id", taskID, "error", err)
		return
	}
	if !changed {
		item, getErr := s.deletions.GetByTaskID(writeCtx, taskID)
		if getErr == nil && item.Status == models.CloudFileDeleted {
			s.logInfo("cloud download file deletion already completed by another attempt", "task_id", taskID)
		}
	}
}

func (s *CloudFileDeletionService) authoritativeDeletion(ctx context.Context, taskID int64, want string) (models.CloudDownloadFileDeletion, error) {
	item, err := s.deletions.GetByTaskID(ctx, taskID)
	if err != nil {
		return models.CloudDownloadFileDeletion{}, err
	}
	if item.Status != want {
		return models.CloudDownloadFileDeletion{}, models.ErrCloudFileDeleteInProgress
	}
	return item, nil
}

func (s *CloudFileDeletionService) logInfo(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Info(msg, args...)
	}
}

func (s *CloudFileDeletionService) logWarn(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Warn(msg, args...)
	}
}

func attachCloudFileDeletion(task models.CloudDownloadTask, deletion models.CloudDownloadFileDeletion) models.CloudDownloadTask {
	task.CloudFileID = deletion.FileID
	task.CloudParentID = deletion.ParentID
	task.FileDeletion = &models.CloudDownloadFileDeletionSummary{
		Status:       deletion.Status,
		ErrorMessage: deletion.ErrorMessage,
		Note:         deletion.Note,
		FinishedAt:   deletion.FinishedAt,
		UpdatedAt:    deletion.UpdatedAt,
	}
	return task
}

func cloudFileDeletionDependencyMessage(err error) string {
	switch {
	case errors.Is(err, models.ErrCloudFileDeleteBrowserDown):
		return "115 浏览器未运行"
	case errors.Is(err, models.ErrCloudFileDeleteLoginRequired):
		return "115 登录已过期"
	case errors.Is(err, models.ErrCloudFileDeleteTimeout):
		return "115 删除请求超时"
	default:
		return err.Error()
	}
}

func mapCloudFileDeletionSessionError(err error) error {
	switch {
	case errors.Is(err, cloud.ErrBrowserDown):
		return models.ErrCloudFileDeleteBrowserDown
	case errors.Is(err, cloud.ErrLoginRequired):
		return models.ErrCloudFileDeleteLoginRequired
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return models.ErrCloudFileDeleteTimeout
	default:
		return err
	}
}

func isCloudFileAlreadyMissing(result cloud.DeleteFileResult) bool {
	if !result.JSONParsed {
		return false
	}
	text := strings.ToLower(strings.Join([]string{result.Error, result.Errno, result.ErrType}, " "))
	normalized := strings.NewReplacer("_", " ", "-", " ", ".", " ").Replace(text)
	if strings.Contains(normalized, "目录") || strings.Contains(normalized, "parent") || strings.Contains(normalized, "folder") || strings.Contains(normalized, "directory") {
		return false
	}
	for _, marker := range []string{
		"文件不存在", "文件已删除", "file not found", "file does not exist", "file not exist", "file deleted",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func sameCloudFileName(declared, resolved string) bool {
	clean := func(value string) string {
		value = strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
		return strings.TrimSpace(path.Base(value))
	}
	left, right := clean(declared), clean(resolved)
	return left != "" && right != "" && left == right
}

var numeric115ID = regexp.MustCompile(`^[0-9]{1,20}$`)

func isNumeric115ID(value string) bool {
	return numeric115ID.MatchString(strings.TrimSpace(value))
}

func cloudFileDeletionFailureMessage(result cloud.DeleteFileResult) string {
	parts := make([]string, 0, 3)
	if strings.TrimSpace(result.Error) != "" {
		parts = append(parts, strings.TrimSpace(result.Error))
	}
	if strings.TrimSpace(result.Errno) != "" {
		parts = append(parts, "errno="+strings.TrimSpace(result.Errno))
	}
	if strings.TrimSpace(result.ErrType) != "" {
		parts = append(parts, "errtype="+strings.TrimSpace(result.ErrType))
	}
	if len(parts) == 0 {
		return "115 未返回删除结果"
	}
	return truncateMessage(strings.Join(parts, " "), 500)
}

// liveCloudFileDeletionSession adapts the CDP session to the cleanup flow.
type liveCloudFileDeletionSession struct {
	session *cloud.Session
	userID  string
}

func (l liveCloudFileDeletionSession) Close() { l.session.Close() }

func (l liveCloudFileDeletionSession) UserID() string { return l.userID }

func (l liveCloudFileDeletionSession) FindUniqueTaskDeep(ctx context.Context, infoHash string, maxPages int) (*cloud.Task, error) {
	return cloud.FindUniqueTaskDeep(ctx, l.session, infoHash, maxPages)
}

func (l liveCloudFileDeletionSession) DeleteFile(ctx context.Context, parentID, fileID string) (cloud.DeleteFileResult, error) {
	return cloud.DeleteFile(ctx, l.session, parentID, fileID)
}
