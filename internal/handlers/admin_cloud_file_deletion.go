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

// cloudFileDeletionService is the admin-facing contract for moving a 115
// cloud source file to the recycle bin.
type cloudFileDeletionService interface {
	DeleteCloudFile(ctx context.Context, taskID int64, requestedBy *uuid.UUID) (models.CloudDownloadTask, error)
}

// SetCloudFileDeletionService wires the manual cleanup service.
func (a *API) SetCloudFileDeletionService(svc cloudFileDeletionService) {
	a.cloudFileDeletion = svc
}

func (a *API) requireCloudFileDeletion(c *gin.Context) bool {
	if a.cloudFileDeletion == nil {
		response.Error(c, 503, "115 云盘清理功能未启用")
		return false
	}
	return true
}

// AdminCloudDownloadDeleteCloudFile moves one completed task's actual 115 file
// to the 115 recycle bin. The request must explicitly carry confirm=true.
func (a *API) AdminCloudDownloadDeleteCloudFile(c *gin.Context) {
	if !a.requireCloudDownload(c) || !a.requireCloudFileDeletion(c) {
		return
	}
	taskID, okID := parseInt64(c.Param("id"))
	if !okID {
		response.Error(c, 400, "任务 id 无效")
		return
	}
	var req struct {
		Confirm bool `json:"confirm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || !req.Confirm {
		response.Error(c, 400, "请确认移入 115 回收站")
		return
	}
	var requestedBy *uuid.UUID
	if userID, okUser := middleware.UserIDFromContext(c); okUser {
		requestedBy = &userID
	}
	task, err := a.cloudFileDeletion.DeleteCloudFile(c.Request.Context(), taskID, requestedBy)
	if err != nil {
		respondCloudFileDeletionError(c, err)
		return
	}
	ok(c, gin.H{"task": task})
}

func respondCloudFileDeletionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, models.ErrCloudFileDeleteNotFound):
		response.Error(c, 404, "云下载任务不存在")
	case errors.Is(err, models.ErrCloudFileDeleteNotCompleted):
		response.Error(c, 409, "云下载任务尚未完成")
	case errors.Is(err, models.ErrCloudFileDeleteInProgress):
		response.Error(c, 409, "该任务的 115 云盘文件正在移入回收站，请稍后刷新")
	case errors.Is(err, models.ErrCloudFileDeleteUnavailable):
		response.Error(c, 409, "无法定位或校验 115 云盘文件，可能缺少文件标识、账号不一致或不是普通文件")
	case errors.Is(err, models.ErrCloudFileDeleteLocalMissing):
		response.Error(c, 410, "本地文件不存在且没有已完成的导入记录，不能删除唯一的 115 源文件")
	case errors.Is(err, models.ErrCloudFileDeleteBrowserDown):
		response.Error(c, 503, "115 浏览器未运行，请先打开并登录 115 浏览器")
	case errors.Is(err, models.ErrCloudFileDeleteLoginRequired):
		response.Error(c, 503, "115 登录已过期，请先在管理端扫码登录后重试")
	case errors.Is(err, models.ErrCloudFileDeleteTimeout):
		response.Error(c, 504, "115 云盘删除请求超时，请先确认回收站状态后再重试")
	case errors.Is(err, models.ErrCloudFileDeleteFailed):
		response.Error(c, 502, "115 拒绝移入回收站，请稍后重试")
	default:
		response.Error(c, 500, "移入 115 回收站失败，请稍后重试")
	}
}
