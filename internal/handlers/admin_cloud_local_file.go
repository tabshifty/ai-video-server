package handlers

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/middleware"
	"video-server/internal/models"
	"video-server/internal/response"
)

// cloudLocalFileDeletionService is the admin-facing contract for deleting the
// DOWNLOAD_DIR file of a completed cloud download.
type cloudLocalFileDeletionService interface {
	Preview(ctx context.Context, taskID int64) (models.CloudLocalFilePreview, error)
	Delete(ctx context.Context, taskID int64, requestedBy *uuid.UUID) (models.CloudDownloadTask, error)
}

// SetCloudLocalFileDeletionService wires the explicit local staging cleanup.
func (a *API) SetCloudLocalFileDeletionService(svc cloudLocalFileDeletionService) {
	a.cloudLocalFileDeletion = svc
}

func (a *API) requireCloudLocalFileDeletion(c *gin.Context) bool {
	if a.cloudLocalFileDeletion == nil {
		response.Error(c, 503, "云下载本地文件清理功能未启用")
		return false
	}
	return true
}

// AdminCloudDownloadLocalFilePreview reports path, size and existence before
// the admin confirms the irreversible delete.
func (a *API) AdminCloudDownloadLocalFilePreview(c *gin.Context) {
	if !a.requireCloudDownload(c) || !a.requireCloudLocalFileDeletion(c) {
		return
	}
	taskID, okID := parseInt64(c.Param("id"))
	if !okID || taskID <= 0 {
		response.Error(c, 400, "任务 id 无效")
		return
	}
	preview, err := a.cloudLocalFileDeletion.Preview(c.Request.Context(), taskID)
	if err != nil {
		a.respondCloudLocalFileError(c, err)
		return
	}
	ok(c, preview)
}

// AdminCloudDownloadDeleteLocalFile deletes only the downloaded file; the task
// record and the 115 cloud copy are untouched.
func (a *API) AdminCloudDownloadDeleteLocalFile(c *gin.Context) {
	if !a.requireCloudDownload(c) || !a.requireCloudLocalFileDeletion(c) {
		return
	}
	taskID, okID := parseInt64(c.Param("id"))
	if !okID || taskID <= 0 {
		response.Error(c, 400, "任务 id 无效")
		return
	}
	var req struct {
		Confirm bool `json:"confirm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || !req.Confirm {
		response.Error(c, 400, "需要显式确认删除本地文件")
		return
	}
	var requestedBy *uuid.UUID
	if userID, okUser := middleware.UserIDFromContext(c); okUser {
		requestedBy = &userID
	}
	task, err := a.cloudLocalFileDeletion.Delete(c.Request.Context(), taskID, requestedBy)
	if err != nil {
		a.respondCloudLocalFileError(c, err)
		return
	}
	ok(c, gin.H{"task": task})
}

func (a *API) respondCloudLocalFileError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, models.ErrCloudLocalFileNotFound), errors.Is(err, models.ErrCloudDownloadNotFound):
		response.Error(c, 404, "云下载任务不存在")
	case errors.Is(err, models.ErrCloudLocalFileNotCompleted):
		response.Error(c, 409, "只有已完成的任务才能删除本地文件")
	case errors.Is(err, models.ErrCloudLocalFileNotRecorded):
		response.Error(c, 409, "任务没有记录本地文件路径，无需删除")
	case errors.Is(err, models.ErrCloudLocalFileImportActive):
		response.Error(c, 409, "该任务正在导入，暂不能删除本地文件")
	case errors.Is(err, models.ErrCloudLocalFileUnavailable):
		response.Error(c, 409, "本地文件路径不在下载目录内，无法删除")
	default:
		if a.logger != nil {
			a.logger.Error("delete cloud local file failed", "error", err)
		}
		response.Error(c, 500, "删除本地文件失败，请稍后重试")
	}
}
