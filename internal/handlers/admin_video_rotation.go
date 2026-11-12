package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"video-server/internal/middleware"
	"video-server/internal/queue"
	"video-server/internal/response"
	"video-server/internal/services"
)

// AdminPreviewVideoRotation resolves the method before destructive confirmation.
func (a *API) AdminPreviewVideoRotation(c *gin.Context) {
	id, valid := parseUUID(c.Param("id"))
	if !valid {
		bad(c, "视频标识无效")
		return
	}
	var req struct {
		Mode string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "请求格式无效")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	plan, err := services.NewVideoRotationService(a.repo, a.storageRoot).Preview(ctx, id, req.Mode)
	if err != nil {
		response.Error(c, 1090, err.Error())
		return
	}
	ok(c, plan)
}

// AdminSubmitVideoRotation persists work before sending a queue wakeup.
func (a *API) AdminSubmitVideoRotation(c *gin.Context) {
	id, valid := parseUUID(c.Param("id"))
	if !valid {
		bad(c, "视频标识无效")
		return
	}
	user, valid := middleware.UserIDFromContext(c)
	if !valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var req struct {
		RequestID string `json:"request_id"`
		Angle     int    `json:"angle"`
		Mode      string `json:"mode"`
		Version   string `json:"version"`
		Confirmed bool   `json:"reencode_confirmed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "请求格式无效")
		return
	}
	requestID, err := uuid.Parse(req.RequestID)
	if err != nil || requestID == uuid.Nil {
		bad(c, "请求标识无效")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()
	job, err := services.NewVideoRotationService(a.repo, a.storageRoot).Submit(ctx, id, user, requestID, req.Angle, req.Mode, req.Version, req.Confirmed)
	if err != nil {
		response.Error(c, 1091, err.Error())
		return
	}
	if job.Status == "queued" || job.Status == "running" || job.Status == "cleanup" {
		if err = a.enqueuer.EnqueueTranscode(queue.TranscodePayload{RotationJobID: job.ID.String()}); err != nil {
			a.logger.Warn("rotation wakeup deferred to reconciler", "job_id", job.ID, "error", err)
		}
	}
	ok(c, job)
}

// AdminVideoRotation returns the newest persisted job, including cleanup failures.
func (a *API) AdminVideoRotation(c *gin.Context) {
	id, valid := parseUUID(c.Param("id"))
	if !valid {
		bad(c, "视频标识无效")
		return
	}
	job, err := a.repo.LatestRotation(c.Request.Context(), id)
	if err != nil {
		response.Error(c, 1092, "读取方向修正任务失败")
		return
	}
	ok(c, gin.H{"job": job})
}
