package handlers

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
	"video-server/internal/queue"
	"video-server/internal/response"
)

type adminEd2kDownloadCreateRequest struct {
	Entries []adminEd2kDownloadCreateEntry `json:"entries"`
	Links   []string                       `json:"links"`
}

type adminEd2kDownloadCreateEntry struct {
	LineNumber int    `json:"line_number"`
	SourceLink string `json:"source_link"`
}

func timePtr(t time.Time) *time.Time {
	return &t
}

type ed2kDownloadRetryRepository interface {
	GetEd2kDownloadTask(ctx context.Context, id uuid.UUID) (models.AdminEd2kDownloadTask, error)
	RequeueFilesCleanedEd2kDownloadTask(ctx context.Context, id uuid.UUID, history models.AdminEd2kDownloadTaskHistoryItem, startedAt *time.Time) (models.AdminEd2kDownloadTask, error)
	RestoreFilesCleanedEd2kDownloadTask(ctx context.Context, id uuid.UUID, history models.AdminEd2kDownloadTaskHistoryItem, startedAt, cleanedAt *time.Time) (models.AdminEd2kDownloadTask, error)
}

type ed2kDownloadRetryEnqueuer interface {
	EnqueueEd2kDownload(queue.Ed2kDownloadPayload) error
	HasEd2kDownloadTask(taskID string) (bool, error)
}

type ed2kDownloadCreateRepository interface {
	GetEd2kDownloadTaskByHash(ctx context.Context, resourceHash string) (models.AdminEd2kDownloadTask, error)
	AppendEd2kDownloadTaskHistory(ctx context.Context, id uuid.UUID, item models.AdminEd2kDownloadTaskHistoryItem) error
	CreateEd2kDownloadTask(ctx context.Context, input models.AdminEd2kDownloadTask) (models.AdminEd2kDownloadTask, error)
	UpdateEd2kDownloadTaskStatus(ctx context.Context, id uuid.UUID, status, progressText, errorMessage string, startedAt, finishedAt, deletedAt *time.Time, outputDir, downloadedPath string, files []models.AdminEd2kDownloadTaskFile, retryDelta int, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error)
}

type ed2kCreateLine struct {
	lineNumber int
	sourceLink string
}

func parseEd2kDownloadLink(raw string) (sourceLink, title, filename, resourceHash string, declaredSize int64, ok bool) {
	sourceLink = strings.TrimSpace(raw)
	if sourceLink == "" || !strings.HasPrefix(strings.ToLower(sourceLink), "ed2k://") {
		return "", "", "", "", 0, false
	}
	parts := strings.Split(sourceLink, "|")
	if len(parts) < 6 || strings.ToLower(strings.TrimSpace(parts[1])) != "file" {
		return "", "", "", "", 0, false
	}
	filename = strings.TrimSpace(parts[2])
	if decoded, err := url.QueryUnescape(filename); err == nil {
		filename = decoded
	}
	parsedSize, err := strconv.ParseInt(strings.TrimSpace(parts[3]), 10, 64)
	if err != nil || parsedSize < 0 {
		parsedSize = 0
	}
	resourceHash = strings.ToUpper(strings.TrimSpace(parts[4]))
	if resourceHash == "" {
		return "", "", "", "", 0, false
	}
	return sourceLink, filename, filename, resourceHash, parsedSize, true
}

func buildEd2kDownloadTaskTitle(linkFilename string) string {
	if title := strings.Join(strings.Fields(strings.TrimSpace(linkFilename)), " "); title != "" {
		return title
	}
	return "ED2K 下载任务"
}

func (a *API) AdminEd2kDownloadTasks(c *gin.Context) {
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	items, total, err := a.repo.ListEd2kDownloadTasks(c.Request.Context(), c.Query("status"), page, pageSize)
	if err != nil {
		response.Error(c, 1090, err.Error())
		return
	}
	ok(c, gin.H{
		"items":       items,
		"total_count": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

func (a *API) AdminCreateEd2kDownloadTasks(c *gin.Context) {
	var req adminEd2kDownloadCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	entries := normalizeEd2kCreateEntries(req)
	if len(entries) == 0 {
		bad(c, "links 不能为空")
		return
	}
	results, err := buildEd2kCreateResults(c.Request.Context(), a.repo, a.enqueuer, entries)
	if err != nil {
		response.Error(c, 1091, err.Error())
		return
	}
	ok(c, gin.H{
		"results": results,
	})
}

func buildEd2kCreateResults(ctx context.Context, repo ed2kDownloadCreateRepository, enqueuer interface {
	EnqueueEd2kDownload(queue.Ed2kDownloadPayload) error
}, entries []ed2kCreateLine) ([]models.AdminEd2kDownloadCreateResult, error) {
	results := make([]models.AdminEd2kDownloadCreateResult, 0, len(entries))
	seenInRequest := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		sourceLink, filename, _, resourceHash, declaredSize, ok := parseEd2kDownloadLink(entry.sourceLink)
		trimmedRaw := strings.TrimSpace(entry.sourceLink)
		if _, duplicated := seenInRequest[trimmedRaw]; duplicated {
			results = append(results, models.AdminEd2kDownloadCreateResult{
				LineNumber: entry.lineNumber,
				SourceLink: trimmedRaw,
				Status:     "duplicate",
				Message:    "同次重复，未创建新任务",
			})
			continue
		}
		seenInRequest[trimmedRaw] = struct{}{}
		if !ok {
			results = append(results, models.AdminEd2kDownloadCreateResult{
				LineNumber: entry.lineNumber,
				SourceLink: trimmedRaw,
				Status:     "invalid",
				Message:    "无效 ED2K 文件链接",
			})
			continue
		}
		existing, err := repo.GetEd2kDownloadTaskByHash(ctx, resourceHash)
		if err == nil {
			if err := repo.AppendEd2kDownloadTaskHistory(ctx, existing.ID, models.AdminEd2kDownloadTaskHistoryItem{
				Kind:    "history",
				Label:   "历史命中",
				Message: "该资源已经存在，未重新创建任务",
				At:      time.Now(),
			}); err != nil {
				return nil, err
			}
			results = append(results, models.AdminEd2kDownloadCreateResult{
				LineNumber:   entry.lineNumber,
				SourceLink:   sourceLink,
				ResourceHash: resourceHash,
				Status:       "reused",
				Message:      "命中历史任务，未重新创建",
				Task:         &existing,
			})
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

		task := models.AdminEd2kDownloadTask{
			ID:           uuid.New(),
			SourceLink:   sourceLink,
			ResourceHash: resourceHash,
			Title:        buildEd2kDownloadTaskTitle(filename),
			Filename:     filename,
			DeclaredSize: declaredSize,
			Status:       "queued",
			ProgressText: "等待执行器接管",
			Files:        []models.AdminEd2kDownloadTaskFile{},
			History: []models.AdminEd2kDownloadTaskHistoryItem{
				{
					Kind:    "created",
					Label:   "已创建",
					Message: "任务已加入下载工作台",
					At:      time.Now(),
				},
			},
		}
		item, createErr := repo.CreateEd2kDownloadTask(ctx, task)
		if createErr != nil {
			results = append(results, models.AdminEd2kDownloadCreateResult{
				LineNumber:   entry.lineNumber,
				SourceLink:   sourceLink,
				ResourceHash: resourceHash,
				Status:       "create_failed",
				Message:      createErr.Error(),
			})
			continue
		}
		if enqueuer == nil {
			item = markEd2kCreateEnqueueFailed(ctx, repo, item, "queue not configured")
			results = append(results, models.AdminEd2kDownloadCreateResult{
				LineNumber:   entry.lineNumber,
				SourceLink:   sourceLink,
				ResourceHash: resourceHash,
				Status:       "enqueue_failed",
				Message:      "queue not configured",
				Task:         &item,
			})
			continue
		}
		if err := enqueuer.EnqueueEd2kDownload(queue.Ed2kDownloadPayload{TaskID: item.ID.String()}); err != nil {
			item = markEd2kCreateEnqueueFailed(ctx, repo, item, err.Error())
			results = append(results, models.AdminEd2kDownloadCreateResult{
				LineNumber:   entry.lineNumber,
				SourceLink:   sourceLink,
				ResourceHash: resourceHash,
				Status:       "enqueue_failed",
				Message:      err.Error(),
				Task:         &item,
			})
			continue
		}
		results = append(results, models.AdminEd2kDownloadCreateResult{
			LineNumber:   entry.lineNumber,
			SourceLink:   sourceLink,
			ResourceHash: resourceHash,
			Status:       "created",
			Message:      "任务已加入下载工作台",
			Task:         &item,
		})
	}
	return results, nil
}

func markEd2kCreateEnqueueFailed(ctx context.Context, repo ed2kDownloadCreateRepository, item models.AdminEd2kDownloadTask, message string) models.AdminEd2kDownloadTask {
	failedTask, err := repo.UpdateEd2kDownloadTaskStatus(
		ctx,
		item.ID,
		"failed",
		"提交下载引擎失败",
		message,
		nil,
		timePtr(time.Now()),
		nil,
		item.OutputDir,
		item.DownloadedPath,
		item.Files,
		0,
		models.AdminEd2kDownloadTaskHistoryItem{
			Kind:    "failed",
			Label:   "失败",
			Message: message,
			At:      time.Now(),
		},
	)
	if err == nil {
		return failedTask
	}
	return item
}

func (a *API) AdminEd2kDownloadTaskDetail(c *gin.Context) {
	taskID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid task id")
		return
	}
	task, err := a.repo.GetEd2kDownloadTask(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "task not found")
			return
		}
		response.Error(c, 1093, err.Error())
		return
	}
	ok(c, task)
}

func normalizeEd2kCreateEntries(req adminEd2kDownloadCreateRequest) []ed2kCreateLine {
	if len(req.Entries) > 0 {
		out := make([]ed2kCreateLine, 0, len(req.Entries))
		for index, entry := range req.Entries {
			lineNumber := entry.LineNumber
			if lineNumber <= 0 {
				lineNumber = index + 1
			}
			out = append(out, ed2kCreateLine{
				lineNumber: lineNumber,
				sourceLink: entry.SourceLink,
			})
		}
		return out
	}
	out := make([]ed2kCreateLine, 0, len(req.Links))
	for index, link := range req.Links {
		out = append(out, ed2kCreateLine{
			lineNumber: index + 1,
			sourceLink: link,
		})
	}
	return out
}

func (a *API) AdminRetryEd2kDownloadTask(c *gin.Context) {
	taskID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid task id")
		return
	}
	item, err := retryEd2kDownloadTask(c.Request.Context(), taskID, a.repo, a.enqueuer, time.Now())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "task not found")
			return
		}
		if errors.Is(err, errEd2kDownloadRetryInvalidStatus) {
			response.Error(c, 1095, err.Error())
			return
		}
		if errors.Is(err, queue.ErrEd2kDownloadTaskInFlight) {
			response.Error(c, 1096, "任务已在执行中")
			return
		}
		response.Error(c, 1096, err.Error())
		return
	}
	ok(c, item)
}

type ed2kCleanupRetryRepository interface {
	GetEd2kDownloadTask(ctx context.Context, id uuid.UUID) (models.AdminEd2kDownloadTask, error)
	MarkEd2kDownloadTaskCancelledCleanupResolved(ctx context.Context, id uuid.UUID, cleanedAt *time.Time, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error)
}

type ed2kDownloadCancelRepository interface {
	MarkEd2kDownloadTaskCanceling(ctx context.Context, id uuid.UUID, progressText string, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error)
	MarkEd2kDownloadTaskCancellationFailed(ctx context.Context, id uuid.UUID, progressText, errorMessage string, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error)
	MarkEd2kDownloadTaskCancelled(ctx context.Context, id uuid.UUID, progressText, errorMessage string, finishedAt, cleanedAt *time.Time, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error)
}

func retryEd2kDownloadTask(ctx context.Context, taskID uuid.UUID, repo ed2kDownloadRetryRepository, enqueuer ed2kDownloadRetryEnqueuer, now time.Time) (models.AdminEd2kDownloadTask, error) {
	task, err := repo.GetEd2kDownloadTask(ctx, taskID)
	if err != nil {
		return models.AdminEd2kDownloadTask{}, err
	}
	if task.Status != "files_cleaned" {
		return models.AdminEd2kDownloadTask{}, errEd2kDownloadRetryInvalidStatus
	}
	item, err := repo.RequeueFilesCleanedEd2kDownloadTask(ctx, taskID, models.AdminEd2kDownloadTaskHistoryItem{
		Kind:    "retry",
		Label:   "重新下载",
		Message: "管理员重新下载已清理的历史任务",
		At:      now,
	}, timePtr(now))
	if err != nil {
		if isEd2kConditionalUpdateMiss(err) {
			current, reloadErr := repo.GetEd2kDownloadTask(ctx, taskID)
			if reloadErr == nil {
				if current.Status == "running" {
					return current, nil
				}
				if current.Status == "queued" && enqueuer != nil {
					hasTask, inspectErr := enqueuer.HasEd2kDownloadTask(taskID.String())
					if inspectErr != nil {
						return models.AdminEd2kDownloadTask{}, fmt.Errorf("%w; inspect queued task after stale retry update: %v", err, inspectErr)
					}
					if hasTask {
						return current, nil
					}
				}
			}
			if reloadErr != nil {
				return models.AdminEd2kDownloadTask{}, fmt.Errorf("%w; reload task after stale retry update: %v", err, reloadErr)
			}
		}
		return models.AdminEd2kDownloadTask{}, err
	}
	if enqueuer == nil {
		_, rollbackErr := repo.RestoreFilesCleanedEd2kDownloadTask(ctx, taskID, models.AdminEd2kDownloadTaskHistoryItem{
			Kind:    "retry_rollback",
			Label:   "重试已撤销",
			Message: "下载队列未配置，已恢复为文件已清理历史任务",
			At:      now,
		}, task.StartedAt, task.CleanedAt)
		if rollbackErr != nil {
			return models.AdminEd2kDownloadTask{}, fmt.Errorf("queue not configured; rollback requeued task: %v", rollbackErr)
		}
		return models.AdminEd2kDownloadTask{}, errors.New("queue not configured")
	}
	if err := enqueuer.EnqueueEd2kDownload(queue.Ed2kDownloadPayload{TaskID: taskID.String()}); err != nil {
		hasTask, inspectErr := enqueuer.HasEd2kDownloadTask(taskID.String())
		if inspectErr != nil {
			return models.AdminEd2kDownloadTask{}, fmt.Errorf("%w; inspect enqueued task after enqueue error: %v", err, inspectErr)
		}
		if hasTask {
			current, reloadErr := repo.GetEd2kDownloadTask(ctx, taskID)
			if reloadErr != nil {
				return models.AdminEd2kDownloadTask{}, fmt.Errorf("%w; reload task after enqueue error with existing job: %v", err, reloadErr)
			}
			if current.Status == "queued" || current.Status == "running" {
				return current, nil
			}
		}
		if _, rollbackErr := repo.RestoreFilesCleanedEd2kDownloadTask(ctx, taskID, models.AdminEd2kDownloadTaskHistoryItem{
			Kind:    "retry_rollback",
			Label:   "重试已撤销",
			Message: "重新加入下载队列失败，已恢复为文件已清理历史任务",
			At:      now,
		}, task.StartedAt, task.CleanedAt); rollbackErr != nil {
			return models.AdminEd2kDownloadTask{}, fmt.Errorf("%w; rollback requeued task: %v", err, rollbackErr)
		}
		return models.AdminEd2kDownloadTask{}, err
	}
	return item, nil
}

var errEd2kDownloadRetryInvalidStatus = errors.New("only files_cleaned task can retry")
var errEd2kDownloadCleanupRetryInvalidStatus = errors.New("only cancelled task with cleanup failure can retry cleanup")

func cancelEd2kDownloadTask(ctx context.Context, task models.AdminEd2kDownloadTask, repo ed2kDownloadCancelRepository, paths ed2kDeletePaths, now time.Time) (models.AdminEd2kDownloadTask, error) {
	cancelingItem := task
	if task.Status == "running" {
		var err error
		cancelingItem, err = repo.MarkEd2kDownloadTaskCanceling(ctx, task.ID, "正在取消下载任务", models.AdminEd2kDownloadTaskHistoryItem{
			Kind:    "canceling",
			Label:   "取消中",
			Message: "正在停止下载并清理残留",
			At:      now,
		})
		if err != nil {
			return models.AdminEd2kDownloadTask{}, err
		}
	}
	cancelErr := cancelFailedEd2kDownload(cancelingItem, ed2kDeletePaths{
		downloadRoot:   paths.downloadRoot,
		downloadSubdir: paths.downloadSubdir,
		amulecmdBin:    paths.amulecmdBin,
		remoteHost:     paths.remoteHost,
		remotePort:     paths.remotePort,
		remotePassword: paths.remotePassword,
	})
	if cancelErr != nil {
		return repo.MarkEd2kDownloadTaskCancellationFailed(ctx, task.ID, "取消失败，请重试", cancelErr.Error(), models.AdminEd2kDownloadTaskHistoryItem{
			Kind:    "cancel_failed",
			Label:   "取消失败",
			Message: cancelErr.Error(),
			At:      time.Now(),
		})
	}
	cleanupErr := deleteTaskWorkspaceArtifacts(cancelingItem, ed2kDeletePaths{
		downloadRoot:   paths.downloadRoot,
		downloadSubdir: paths.downloadSubdir,
	})
	finishedAt := timePtr(time.Now())
	if cleanupErr != nil {
		return repo.MarkEd2kDownloadTaskCancelled(ctx, task.ID, "任务已取消，仍有残留待清理", cleanupErr.Error(), finishedAt, nil, models.AdminEd2kDownloadTaskHistoryItem{
			Kind:    "cancelled",
			Label:   "已取消",
			Message: "下载已停止，但残留清理失败",
			At:      time.Now(),
		})
	}
	return repo.MarkEd2kDownloadTaskCancelled(ctx, task.ID, "任务已取消", "", finishedAt, finishedAt, models.AdminEd2kDownloadTaskHistoryItem{
		Kind:    "cancelled",
		Label:   "已取消",
		Message: "下载已停止，残留已清理",
		At:      time.Now(),
	})
}

func (a *API) AdminDeleteEd2kDownloadTask(c *gin.Context) {
	taskID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid task id")
		return
	}
	task, err := a.repo.GetEd2kDownloadTask(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "task not found")
			return
		}
		response.Error(c, 1097, err.Error())
		return
	}
	now := time.Now()
	if task.Status == "queued" {
		if err := a.repo.DeleteEd2kDownloadTask(c.Request.Context(), taskID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				response.Error(c, 404, "task not found")
				return
			}
			response.Error(c, 1098, err.Error())
			return
		}
		ok(c, gin.H{
			"id":      taskID,
			"deleted": true,
			"status":  "deleted",
			"at":      now,
		})
		return
	}
	if task.Status == "failed" {
		if err := deleteFailedEd2kDownloadArtifacts(task, ed2kDeletePaths{
			downloadRoot:   a.ed2kDownloadRoot,
			downloadSubdir: a.ed2kDownloadSubdir,
			amulecmdBin:    a.amulecmdBin,
			remoteHost:     a.amuleRemoteHost,
			remotePort:     a.amuleRemotePort,
			remotePassword: a.amuleRemotePassword,
		}); err != nil {
			response.Error(c, 1098, err.Error())
			return
		}
		if err := a.repo.DeleteEd2kDownloadTask(c.Request.Context(), taskID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				response.Error(c, 404, "task not found")
				return
			}
			response.Error(c, 1098, err.Error())
			return
		}
		ok(c, gin.H{
			"id":      taskID,
			"deleted": true,
			"status":  "deleted",
			"at":      now,
		})
		return
	}
	if task.Status == "running" || task.Status == "canceling" {
		item, err := cancelEd2kDownloadTask(c.Request.Context(), task, a.repo, ed2kDeletePaths{
			downloadRoot:   a.ed2kDownloadRoot,
			downloadSubdir: a.ed2kDownloadSubdir,
			amulecmdBin:    a.amulecmdBin,
			remoteHost:     a.amuleRemoteHost,
			remotePort:     a.amuleRemotePort,
			remotePassword: a.amuleRemotePassword,
		}, now)
		if err != nil {
			if isEd2kConditionalUpdateMiss(err) {
				response.Error(c, 1099, "任务状态已变化，请刷新后重试")
				return
			}
			response.Error(c, 1098, err.Error())
			return
		}
		ok(c, item)
		return
	}
	response.Error(c, 1099, "only queued, failed, running or canceling task can delete")
}

func (a *API) AdminRetryCancelledEd2kDownloadCleanup(c *gin.Context) {
	taskID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid task id")
		return
	}
	item, err := retryCancelledEd2kDownloadCleanup(c.Request.Context(), taskID, a.repo, ed2kDeletePaths{
		downloadRoot:   a.ed2kDownloadRoot,
		downloadSubdir: a.ed2kDownloadSubdir,
	}, time.Now())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "task not found")
			return
		}
		if errors.Is(err, errEd2kDownloadCleanupRetryInvalidStatus) {
			response.Error(c, 1099, err.Error())
			return
		}
		response.Error(c, 1098, err.Error())
		return
	}
	ok(c, item)
}

func retryCancelledEd2kDownloadCleanup(ctx context.Context, taskID uuid.UUID, repo ed2kCleanupRetryRepository, paths ed2kDeletePaths, now time.Time) (models.AdminEd2kDownloadTask, error) {
	task, err := repo.GetEd2kDownloadTask(ctx, taskID)
	if err != nil {
		return models.AdminEd2kDownloadTask{}, err
	}
	if task.Status != "cancelled" || strings.TrimSpace(task.ErrorMessage) == "" || task.CleanedAt != nil {
		return models.AdminEd2kDownloadTask{}, errEd2kDownloadCleanupRetryInvalidStatus
	}
	if err := deleteTaskWorkspaceArtifacts(task, paths); err != nil {
		return models.AdminEd2kDownloadTask{}, err
	}
	item, err := repo.MarkEd2kDownloadTaskCancelledCleanupResolved(ctx, taskID, timePtr(now), models.AdminEd2kDownloadTaskHistoryItem{
		Kind:    "cancelled_cleanup",
		Label:   "残留已清理",
		Message: "管理员补做残留清理成功",
		At:      now,
	})
	if err != nil {
		return models.AdminEd2kDownloadTask{}, err
	}
	return item, nil
}

func isEd2kConditionalUpdateMiss(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), pgx.ErrNoRows.Error())
}

func (a *API) AdminCleanEd2kDownloadTaskFiles(c *gin.Context) {
	taskID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid task id")
		return
	}
	task, err := a.repo.GetEd2kDownloadTask(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "task not found")
			return
		}
		response.Error(c, 1098, err.Error())
		return
	}
	if task.Status != "completed" {
		response.Error(c, 1099, "only completed task can clean files")
		return
	}
	if err := cleanCompletedEd2kDownloadArtifacts(task, ed2kDeletePaths{
		downloadRoot:   a.ed2kDownloadRoot,
		downloadSubdir: a.ed2kDownloadSubdir,
	}); err != nil {
		response.Error(c, 1098, err.Error())
		return
	}
	item, err := a.repo.MarkEd2kDownloadTaskFilesCleaned(c.Request.Context(), taskID, "已下载，暂存文件已清理", models.AdminEd2kDownloadTaskHistoryItem{
		Kind:    "files_cleaned",
		Label:   "文件已清理",
		Message: "已删除暂存文件并保留历史下载记录",
		At:      time.Now(),
	})
	if err != nil {
		response.Error(c, 1098, err.Error())
		return
	}
	ok(c, item)
}

type ed2kDeletePaths struct {
	downloadRoot   string
	downloadSubdir string
	amulecmdBin    string
	remoteHost     string
	remotePort     string
	remotePassword string
}

func deleteFailedEd2kDownloadArtifacts(task models.AdminEd2kDownloadTask, paths ed2kDeletePaths) error {
	if err := cancelFailedEd2kDownload(task, paths); err != nil {
		return err
	}
	return deleteTaskWorkspaceArtifacts(task, paths)
}

func deleteTaskWorkspaceArtifacts(task models.AdminEd2kDownloadTask, paths ed2kDeletePaths) error {
	candidates := collectFailedEd2kDeleteCandidates(task, paths)
	for _, candidate := range candidates {
		if err := removeFailedEd2kPath(candidate); err != nil {
			return err
		}
	}
	return nil
}

func cleanCompletedEd2kDownloadArtifacts(task models.AdminEd2kDownloadTask, paths ed2kDeletePaths) error {
	taskRoot := strings.TrimSpace(task.OutputDir)
	taskRootResolved := taskRoot
	if resolved, err := filepath.EvalSymlinks(taskRoot); err == nil {
		taskRootResolved = resolved
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("删除暂存文件失败: 校验任务目录失败: %w", err)
	}
	managedRoot := managedEd2kBaseDir(paths)
	managedRootResolved := managedRoot
	if resolved, err := filepath.EvalSymlinks(managedRoot); err == nil {
		managedRootResolved = resolved
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("删除暂存文件失败: 校验受管目录失败: %w", err)
	}
	if !pathWithinRoot(taskRootResolved, managedRootResolved) {
		return fmt.Errorf("删除暂存文件失败: 任务目录超出受管目录: %s", taskRoot)
	}
	for _, file := range task.Files {
		target := strings.TrimSpace(file.Path)
		if target == "" {
			continue
		}
		if filepath.IsAbs(target) {
			return fmt.Errorf("删除暂存文件失败: 结果路径必须是相对路径: %s", target)
		}
		linkPath := filepath.Join(task.OutputDir, target)
		if !pathWithinRoot(linkPath, taskRoot) {
			return fmt.Errorf("删除暂存文件失败: 结果路径越出任务目录: %s", target)
		}
		if realParent, err := filepath.EvalSymlinks(filepath.Dir(linkPath)); err == nil && !pathWithinRoot(realParent, taskRootResolved) {
			return fmt.Errorf("删除暂存文件失败: 结果路径父目录越出任务目录: %s", target)
		}
		if err := removeManagedEd2kLinkAndTarget(linkPath, paths); err != nil {
			return err
		}
	}
	return removeManagedOutputDir(task.OutputDir, paths)
}

func removeManagedEd2kLinkAndTarget(linkPath string, paths ed2kDeletePaths) error {
	info, err := os.Lstat(linkPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("删除暂存文件失败: %s: %w", linkPath, err)
	}
	baseDownloadDir := managedEd2kBaseDir(paths)
	if !pathWithinRoot(linkPath, baseDownloadDir) {
		return fmt.Errorf("删除暂存文件失败: 超出受管目录: %s", linkPath)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		targetPath, err := os.Readlink(linkPath)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("读取暂存链接失败: %s: %w", linkPath, err)
		}
		if err := os.Remove(linkPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("删除暂存链接失败: %s: %w", linkPath, err)
		}
		if targetPath == "" {
			return nil
		}
		if !filepath.IsAbs(targetPath) {
			targetPath = filepath.Join(filepath.Dir(linkPath), targetPath)
		}
		targetPath = filepath.Clean(targetPath)
		if !pathWithinRoot(targetPath, baseDownloadDir) {
			return fmt.Errorf("删除暂存真文件失败: 超出受管目录: %s", targetPath)
		}
		if err := removeFailedEd2kPath(targetPath); err != nil {
			return err
		}
		return nil
	}
	if info.IsDir() {
		return removeFailedEd2kPath(linkPath)
	}
	if err := os.Remove(linkPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("删除暂存文件失败: %s: %w", linkPath, err)
	}
	return nil
}

func removeManagedOutputDir(path string, paths ed2kDeletePaths) error {
	cleanPath := strings.TrimSpace(path)
	if cleanPath == "" {
		return nil
	}
	baseDownloadDir := managedEd2kBaseDir(paths)
	baseResolved := baseDownloadDir
	if resolved, err := filepath.EvalSymlinks(baseDownloadDir); err == nil {
		baseResolved = resolved
	}
	pathResolved := cleanPath
	if resolved, err := filepath.EvalSymlinks(cleanPath); err == nil {
		pathResolved = resolved
	}
	if !pathWithinRoot(pathResolved, baseResolved) {
		return fmt.Errorf("删除暂存目录失败: 超出受管目录: %s", cleanPath)
	}
	if err := removeFailedEd2kPath(cleanPath); err != nil {
		return err
	}
	return nil
}

func managedEd2kBaseDir(paths ed2kDeletePaths) string {
	root := strings.TrimSpace(paths.downloadRoot)
	if root == "" {
		return ""
	}
	if subdir := strings.TrimSpace(paths.downloadSubdir); subdir != "" {
		return filepath.Join(root, subdir)
	}
	return root
}

func collectFailedEd2kDeleteCandidates(task models.AdminEd2kDownloadTask, paths ed2kDeletePaths) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, 6)
	add := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		clean := filepath.Clean(path)
		if clean == "." {
			return
		}
		if _, ok := seen[clean]; ok {
			return
		}
		seen[clean] = struct{}{}
		out = append(out, clean)
	}

	baseDownloadDir := ""
	if root := strings.TrimSpace(paths.downloadRoot); root != "" {
		baseDownloadDir = root
		if subdir := strings.TrimSpace(paths.downloadSubdir); subdir != "" {
			baseDownloadDir = filepath.Join(root, subdir)
		}
	}
	outputDirWithinManagedRoot := false
	if outputDir := strings.TrimSpace(task.OutputDir); outputDir != "" {
		if pathWithinRoot(outputDir, baseDownloadDir) {
			outputDirWithinManagedRoot = true
			add(outputDir)
		}
	}
	if downloadedPath := strings.TrimSpace(task.DownloadedPath); downloadedPath != "" {
		if (outputDirWithinManagedRoot && pathWithinRoot(downloadedPath, task.OutputDir)) || pathWithinRoot(downloadedPath, baseDownloadDir) {
			add(downloadedPath)
		}
	}
	if baseDownloadDir != "" {
		add(filepath.Join(baseDownloadDir, task.ResourceHash))
		if task.ID != uuid.Nil {
			add(filepath.Join(baseDownloadDir, task.ID.String()))
		}
		if name := strings.TrimSpace(task.Filename); name != "" {
			add(filepath.Join(baseDownloadDir, name))
		}
	}
	return out
}

func cancelFailedEd2kDownload(task models.AdminEd2kDownloadTask, paths ed2kDeletePaths) error {
	hash := strings.TrimSpace(task.ResourceHash)
	command := strings.TrimSpace(paths.amulecmdBin)
	if hash == "" || command == "" || strings.TrimSpace(paths.remotePassword) == "" {
		return nil
	}
	cmd := exec.Command(
		command,
		"-h", strings.TrimSpace(paths.remoteHost),
		"-p", strings.TrimSpace(paths.remotePort),
		"-P", strings.TrimSpace(paths.remotePassword),
		"-c", "cancel "+hash,
	)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	text := strings.TrimSpace(string(output))
	if text == "" {
		return fmt.Errorf("取消 aMule 下载失败: %w", err)
	}
	lower := strings.ToLower(text)
	if strings.Contains(lower, "filehash not found") || strings.Contains(text, "文件校验码不存在") {
		return nil
	}
	return fmt.Errorf("取消 aMule 下载失败: %s", text)
}

func pathWithinRoot(path, root string) bool {
	path = strings.TrimSpace(path)
	root = strings.TrimSpace(root)
	if path == "" || root == "" {
		return false
	}
	cleanPath := filepath.Clean(path)
	cleanRoot := filepath.Clean(root)
	rel, err := filepath.Rel(cleanRoot, cleanPath)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)))
}

func removeFailedEd2kPath(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("删除下载残留失败: %s: %w", path, err)
	}
	if info.IsDir() {
		if err := os.RemoveAll(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("删除下载残留目录失败: %s: %w", path, err)
		}
		return nil
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("删除下载残留文件失败: %s: %w", path, err)
	}
	return nil
}
