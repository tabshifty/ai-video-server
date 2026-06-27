package handlers

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
	"video-server/internal/response"
)

type adminEd2kDownloadCreateRequest struct {
	Links []string `json:"links"`
	Title string   `json:"title"`
}

func normalizeEd2kDownloadTitle(title string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(title)), " ")
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

func buildEd2kDownloadTaskTitle(linkFilename, requestedTitle string) string {
	if title := normalizeEd2kDownloadTitle(requestedTitle); title != "" {
		return title
	}
	if title := normalizeEd2kDownloadTitle(linkFilename); title != "" {
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
	req.Title = normalizeEd2kDownloadTitle(req.Title)
	if len(req.Links) == 0 {
		bad(c, "links 不能为空")
		return
	}

	created := make([]models.AdminEd2kDownloadTask, 0, len(req.Links))
	reused := make([]models.AdminEd2kDownloadTask, 0, len(req.Links))
	rejected := make([]string, 0, len(req.Links))
	for _, raw := range req.Links {
		sourceLink, filename, _, resourceHash, declaredSize, ok := parseEd2kDownloadLink(raw)
		if !ok {
			rejected = append(rejected, strings.TrimSpace(raw))
			continue
		}
		existing, err := a.repo.GetEd2kDownloadTaskByHash(c.Request.Context(), resourceHash)
		if err == nil {
			if err := a.repo.AppendEd2kDownloadTaskHistory(c.Request.Context(), existing.ID, models.AdminEd2kDownloadTaskHistoryItem{
				Kind:    "history",
				Label:   "历史命中",
				Message: "该资源已经存在，未重新创建任务",
				At:      time.Now(),
			}); err != nil {
				response.Error(c, 1091, err.Error())
				return
			}
			reused = append(reused, existing)
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 1091, err.Error())
			return
		}

		task := models.AdminEd2kDownloadTask{
			ID:           uuid.New(),
			SourceLink:   sourceLink,
			ResourceHash: resourceHash,
			Title:        buildEd2kDownloadTaskTitle(filename, req.Title),
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
		item, createErr := a.repo.CreateEd2kDownloadTask(c.Request.Context(), task)
		if createErr != nil {
			response.Error(c, 1092, createErr.Error())
			return
		}
		created = append(created, item)
	}
	ok(c, gin.H{
		"created":  created,
		"reused":   reused,
		"rejected": rejected,
	})
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

func (a *API) AdminRetryEd2kDownloadTask(c *gin.Context) {
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
		response.Error(c, 1094, err.Error())
		return
	}
	if task.Status != "failed" && task.Status != "deleted" {
		response.Error(c, 1095, "only failed or deleted task can retry")
		return
	}
	item, err := a.repo.UpdateEd2kDownloadTaskStatus(
		c.Request.Context(),
		taskID,
		"queued",
		"等待执行器接管",
		"",
		nil,
		nil,
		nil,
		task.OutputDir,
		task.DownloadedPath,
		task.Files,
		1,
		models.AdminEd2kDownloadTaskHistoryItem{
			Kind:    "retry",
			Label:   "重新排队",
			Message: "管理员手动重试下载任务",
			At:      time.Now(),
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "task not found")
			return
		}
		response.Error(c, 1096, err.Error())
		return
	}
	ok(c, item)
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
	if task.Status == "running" {
		item, err := a.repo.UpdateEd2kDownloadTaskStatus(
			c.Request.Context(),
			taskID,
			"deleted",
			"任务已取消",
			"",
			nil,
			nil,
			&now,
			task.OutputDir,
			task.DownloadedPath,
			task.Files,
			0,
			models.AdminEd2kDownloadTaskHistoryItem{
				Kind:    "deleted",
				Label:   "已取消",
				Message: "任务已从工作台移除",
				At:      now,
			},
		)
		if err != nil {
			response.Error(c, 1098, err.Error())
			return
		}
		ok(c, item)
		return
	}
	response.Error(c, 1099, "only queued or running task can delete")
}
