package handlers

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/cloud"
	"video-server/internal/middleware"
	"video-server/internal/models"
	"video-server/internal/response"
)

// cloudDownloadService is the admin-facing contract for cloud downloads.
type cloudDownloadService interface {
	Submit(ctx context.Context, postID *int64, rawLink string, createdBy *uuid.UUID) (models.CloudDownloadTask, bool, error)
	Retry(ctx context.Context, id int64) (models.CloudDownloadTask, error)
	List(ctx context.Context, status string, page, pageSize int) ([]models.CloudDownloadTask, int, error)
	StatusByHashes(ctx context.Context, hashes []string) (map[string]models.CloudDownloadStatus, error)
	BrowserState(ctx context.Context) cloud.State
	LoginQR(ctx context.Context) (string, error)
}

// SetCloudDownloadService wires the service. Routes register only when set,
// which keeps CLOUD_DOWNLOAD_ENABLED=false a real off switch.
func (a *API) SetCloudDownloadService(svc cloudDownloadService) {
	a.cloudDownload = svc
}

func (a *API) requireCloudDownload(c *gin.Context) bool {
	if a.cloudDownload == nil {
		response.Error(c, 503, "云下载功能未启用")
		return false
	}
	return true
}

// AdminCloudDownloadList lists cloud download tasks.
func (a *API) AdminCloudDownloadList(c *gin.Context) {
	if !a.requireCloudDownload(c) {
		return
	}
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	items, total, err := a.cloudDownload.List(c.Request.Context(), strings.TrimSpace(c.Query("status")), page, pageSize)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	ok(c, gin.H{"items": items, "total_count": total, "page": page, "page_size": pageSize})
}

// AdminCloudDownloadCreate parses one link and schedules it.
func (a *API) AdminCloudDownloadCreate(c *gin.Context) {
	if !a.requireCloudDownload(c) {
		return
	}
	var req struct {
		PostID *int64 `json:"post_id"`
		Link   string `json:"link"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if strings.TrimSpace(req.Link) == "" {
		response.Error(c, 400, "link 不能为空")
		return
	}
	var createdBy *uuid.UUID
	if userID, okUser := middleware.UserIDFromContext(c); okUser {
		createdBy = &userID
	}
	task, created, err := a.cloudDownload.Submit(c.Request.Context(), req.PostID, req.Link, createdBy)
	if err != nil {
		if errors.Is(err, cloud.ErrEmptyLink) ||
			errors.Is(err, cloud.ErrUnsupportedLink) ||
			errors.Is(err, cloud.ErrMalformedLink) {
			response.Error(c, 400, err.Error())
			return
		}
		response.Error(c, 500, err.Error())
		return
	}
	ok(c, gin.H{"task": task, "created": created})
}

// AdminCloudDownloadStatuses returns the compact status map for a hash batch.
func (a *API) AdminCloudDownloadStatuses(c *gin.Context) {
	if !a.requireCloudDownload(c) {
		return
	}
	var req struct {
		InfoHashes []string `json:"info_hashes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if len(req.InfoHashes) > 200 {
		req.InfoHashes = req.InfoHashes[:200]
	}
	hashes := make([]string, 0, len(req.InfoHashes))
	for _, hash := range req.InfoHashes {
		hash = strings.ToLower(strings.TrimSpace(hash))
		if hash != "" {
			hashes = append(hashes, hash)
		}
	}
	statuses, err := a.cloudDownload.StatusByHashes(c.Request.Context(), hashes)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	ok(c, gin.H{"statuses": statuses})
}

// AdminCloudDownloadRetry requeues a failed or blocked task.
func (a *API) AdminCloudDownloadRetry(c *gin.Context) {
	if !a.requireCloudDownload(c) {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, 400, "任务 id 无效")
		return
	}
	task, err := a.cloudDownload.Retry(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrCloudDownloadNotFound):
			response.Error(c, 404, err.Error())
		case errors.Is(err, models.ErrCloudDownloadNotRetryable):
			response.Error(c, 409, err.Error())
		default:
			response.Error(c, 500, err.Error())
		}
		return
	}
	ok(c, gin.H{"task": task})
}

// AdminCloudDownloadBrowser reports browser/session/quota state.
func (a *API) AdminCloudDownloadBrowser(c *gin.Context) {
	if !a.requireCloudDownload(c) {
		return
	}
	state := a.cloudDownload.BrowserState(c.Request.Context())
	ok(c, gin.H{
		"up":        state.Up,
		"logged_in": state.LoggedIn,
		"page_url":  state.PageURL,
		"quota":     state.Quota,
	})
}

// AdminCloudDownloadLoginQR returns a short-lived 115 login QR image.
func (a *API) AdminCloudDownloadLoginQR(c *gin.Context) {
	if !a.requireCloudDownload(c) {
		return
	}
	image, err := a.cloudDownload.LoginQR(c.Request.Context())
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
	ok(c, gin.H{"image": image, "expires_in_seconds": 120})
}
