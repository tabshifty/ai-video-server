package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/response"
	"video-server/internal/services"
)

const telegramAdminErrorCode = 2601

type telegramSourceService interface {
	List(ctx context.Context) ([]models.TelegramSource, error)
	Add(ctx context.Context, chatRef string) (models.TelegramSource, error)
	Pause(ctx context.Context, sourceID uuid.UUID) (models.TelegramSource, error)
	Resume(ctx context.Context, sourceID uuid.UUID) (models.TelegramSource, error)
	StartBackfill(ctx context.Context, sourceID uuid.UUID) (models.TelegramSource, error)
	GetProgress(ctx context.Context, sourceID uuid.UUID) (services.TelegramSourceProgress, error)
}

type adminTelegramAddSourceRequest struct {
	ChatRef string `json:"chat_ref"`
}

// AdminTelegramSources lists all configured Telegram sources.
func (a *API) AdminTelegramSources(c *gin.Context) {
	if !a.requireTelegramSourceService(c) {
		return
	}
	items, err := a.telegramSourceSvc.List(c.Request.Context())
	if err != nil {
		writeAdminTelegramError(c, err)
		return
	}
	ok(c, gin.H{"items": items})
}

// AdminTelegramAddSource creates a Telegram source for asynchronous resolution.
func (a *API) AdminTelegramAddSource(c *gin.Context) {
	if !a.requireTelegramSourceService(c) {
		return
	}
	var req adminTelegramAddSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "请求 JSON 格式错误")
		return
	}
	source, err := a.telegramSourceSvc.Add(c.Request.Context(), req.ChatRef)
	if err != nil {
		writeAdminTelegramError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"code": 0,
		"msg":  "",
		"data": source,
	})
}

// AdminTelegramPauseSource pauses one Telegram source.
func (a *API) AdminTelegramPauseSource(c *gin.Context) {
	if !a.requireTelegramSourceService(c) {
		return
	}
	a.updateAdminTelegramSource(c, a.telegramSourceSvc.Pause)
}

// AdminTelegramResumeSource resumes one Telegram source at its current cursor.
func (a *API) AdminTelegramResumeSource(c *gin.Context) {
	if !a.requireTelegramSourceService(c) {
		return
	}
	a.updateAdminTelegramSource(c, a.telegramSourceSvc.Resume)
}

// AdminTelegramStartBackfill resets and restarts source history scanning.
func (a *API) AdminTelegramStartBackfill(c *gin.Context) {
	if !a.requireTelegramSourceService(c) {
		return
	}
	a.updateAdminTelegramSource(c, a.telegramSourceSvc.StartBackfill)
}

// AdminTelegramProgress returns processing counts for one Telegram source.
func (a *API) AdminTelegramProgress(c *gin.Context) {
	if !a.requireTelegramSourceService(c) {
		return
	}
	sourceID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "Telegram 来源 ID 格式错误")
		return
	}
	progress, err := a.telegramSourceSvc.GetProgress(c.Request.Context(), sourceID)
	if err != nil {
		writeAdminTelegramError(c, err)
		return
	}
	ok(c, progress)
}

func (a *API) updateAdminTelegramSource(c *gin.Context, update func(context.Context, uuid.UUID) (models.TelegramSource, error)) {
	sourceID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "Telegram 来源 ID 格式错误")
		return
	}
	source, err := update(c.Request.Context(), sourceID)
	if err != nil {
		writeAdminTelegramError(c, err)
		return
	}
	ok(c, source)
}

func (a *API) requireTelegramSourceService(c *gin.Context) bool {
	if a != nil && a.telegramSourceSvc != nil {
		return true
	}
	response.Error(c, telegramAdminErrorCode, "Telegram 来源服务不可用")
	return false
}

func writeAdminTelegramError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrTelegramSourceNotFound):
		response.Error(c, 404, "Telegram 来源不存在")
	case errors.Is(err, services.ErrTelegramChatRefRequired), errors.Is(err, services.ErrTelegramSourceAlreadyExists):
		bad(c, err.Error())
	default:
		response.Error(c, telegramAdminErrorCode, strings.TrimSpace(err.Error()))
	}
}
