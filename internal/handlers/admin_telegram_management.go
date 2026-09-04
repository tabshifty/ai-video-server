package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/middleware"
	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/response"
	"video-server/internal/services"
	"video-server/internal/telegram"
)

type telegramManagementService interface {
	Status(ctx context.Context, actorID uuid.UUID, requestID string) (telegram.ControlStatus, error)
	IssueConfirmation(ctx context.Context, actorID uuid.UUID, action, password string, confirmed bool) (services.TelegramConfirmationView, error)
	StartPhone(ctx context.Context, actorID uuid.UUID, requestID, confirmationTicket string) (telegram.AuthorizationView, error)
	StartQR(ctx context.Context, actorID uuid.UUID, requestID, confirmationTicket string) (telegram.AuthorizationView, error)
	GetAuthorization(ctx context.Context, actorID uuid.UUID, authorizationID, requestID string) (telegram.AuthorizationView, error)
	SubmitCode(ctx context.Context, actorID uuid.UUID, authorizationID, requestID, code string) (telegram.AuthorizationView, error)
	SubmitPassword(ctx context.Context, actorID uuid.UUID, authorizationID, requestID, password string) (telegram.AuthorizationView, error)
	CancelAuthorization(ctx context.Context, actorID uuid.UUID, authorizationID, requestID string) error
	PreviewSource(ctx context.Context, actorID uuid.UUID, requestID, chatRef, confirmationTicket string) (telegram.ChatPreview, error)
	ConfirmSource(ctx context.Context, actorID uuid.UUID, requestID, previewID, chatRef, confirmationTicket string) (models.TelegramSource, error)
	ListSources(ctx context.Context) ([]models.TelegramSource, error)
	PauseSource(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error)
	ResumeSource(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error)
	RecoverSource(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error)
	StartBackfill(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error)
	GetProgress(ctx context.Context, sourceID uuid.UUID) (services.TelegramSourceProgress, error)
	ListAudits(ctx context.Context, filter models.TelegramAuditFilter) ([]models.TelegramAuditLog, int, error)
}

type adminTelegramConfirmationRequest struct {
	Action          string `json:"action"`
	CurrentPassword string `json:"current_password"`
	Confirmed       bool   `json:"confirmed"`
}

type adminTelegramAuthorizationStartRequest struct {
	ConfirmationTicket string `json:"confirmation_ticket"`
}

type adminTelegramCodeRequest struct {
	Code string `json:"code"`
}

type adminTelegramPasswordRequest struct {
	Password string `json:"password"`
}

type adminTelegramSourcePreviewRequest struct {
	ChatRef            string `json:"chat_ref"`
	ConfirmationTicket string `json:"confirmation_ticket"`
}

type adminTelegramSourceConfirmRequest struct {
	PreviewID          string `json:"preview_id"`
	ChatRef            string `json:"chat_ref"`
	ConfirmationTicket string `json:"confirmation_ticket"`
}

// AdminTelegramStatus returns collector state through the private control
// channel. The route itself remains protected by the regular admin middleware.
func (a *API) AdminTelegramStatus(c *gin.Context) {
	service, actorID, available := a.requireTelegramManagementService(c)
	if !available {
		return
	}
	status, err := service.Status(c.Request.Context(), actorID, telegramAdminRequestID())
	if err != nil {
		writeAdminTelegramManagementError(c, err)
		return
	}
	ok(c, status)
}

// AdminTelegramIssueConfirmation verifies the current administrator password
// before a high-risk authorization or private-source action.
func (a *API) AdminTelegramIssueConfirmation(c *gin.Context) {
	service, actorID, available := a.requireTelegramManagementService(c)
	if !available {
		return
	}
	var request adminTelegramConfirmationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		bad(c, "请求 JSON 格式错误")
		return
	}
	view, err := service.IssueConfirmation(c.Request.Context(), actorID, request.Action, request.CurrentPassword, request.Confirmed)
	if err != nil {
		writeAdminTelegramManagementError(c, err)
		return
	}
	ok(c, view)
}

// AdminTelegramStartPhoneAuthorization starts a phone-code authorization after
// a one-time high-risk confirmation has been consumed.
func (a *API) AdminTelegramStartPhoneAuthorization(c *gin.Context) {
	a.startTelegramAuthorization(c, func(service telegramManagementService, actorID uuid.UUID, requestID, ticket string) (telegram.AuthorizationView, error) {
		return service.StartPhone(c.Request.Context(), actorID, requestID, ticket)
	})
}

// AdminTelegramStartQRAuthorization starts a QR reauthorization after a
// one-time high-risk confirmation has been consumed.
func (a *API) AdminTelegramStartQRAuthorization(c *gin.Context) {
	a.startTelegramAuthorization(c, func(service telegramManagementService, actorID uuid.UUID, requestID, ticket string) (telegram.AuthorizationView, error) {
		return service.StartQR(c.Request.Context(), actorID, requestID, ticket)
	})
}

// AdminTelegramAuthorization returns sanitized progress for one authorization.
func (a *API) AdminTelegramAuthorization(c *gin.Context) {
	service, actorID, available := a.requireTelegramManagementService(c)
	if !available {
		return
	}
	view, err := service.GetAuthorization(c.Request.Context(), actorID, c.Param("id"), telegramAdminRequestID())
	if err != nil {
		writeAdminTelegramManagementError(c, err)
		return
	}
	ok(c, view)
}

// AdminTelegramSubmitAuthorizationCode forwards one transient phone code.
func (a *API) AdminTelegramSubmitAuthorizationCode(c *gin.Context) {
	service, actorID, available := a.requireTelegramManagementService(c)
	if !available {
		return
	}
	var request adminTelegramCodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		bad(c, "请求 JSON 格式错误")
		return
	}
	view, err := service.SubmitCode(c.Request.Context(), actorID, c.Param("id"), telegramAdminRequestID(), request.Code)
	if err != nil {
		writeAdminTelegramManagementError(c, err)
		return
	}
	ok(c, view)
}

// AdminTelegramSubmitAuthorizationPassword forwards one transient Telegram 2FA
// password without persisting it in the API process.
func (a *API) AdminTelegramSubmitAuthorizationPassword(c *gin.Context) {
	service, actorID, available := a.requireTelegramManagementService(c)
	if !available {
		return
	}
	var request adminTelegramPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		bad(c, "请求 JSON 格式错误")
		return
	}
	view, err := service.SubmitPassword(c.Request.Context(), actorID, c.Param("id"), telegramAdminRequestID(), request.Password)
	if err != nil {
		writeAdminTelegramManagementError(c, err)
		return
	}
	ok(c, view)
}

// AdminTelegramCancelAuthorization stops an authorization owned by the caller.
func (a *API) AdminTelegramCancelAuthorization(c *gin.Context) {
	service, actorID, available := a.requireTelegramManagementService(c)
	if !available {
		return
	}
	if err := service.CancelAuthorization(c.Request.Context(), actorID, c.Param("id"), telegramAdminRequestID()); err != nil {
		writeAdminTelegramManagementError(c, err)
		return
	}
	ok(c, gin.H{"cancelled": true})
}

// AdminTelegramPreviewSource resolves a candidate source. A private invitation
// must carry a confirmation ticket; public references do not require one.
func (a *API) AdminTelegramPreviewSource(c *gin.Context) {
	service, actorID, available := a.requireTelegramManagementService(c)
	if !available {
		return
	}
	var request adminTelegramSourcePreviewRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		bad(c, "请求 JSON 格式错误")
		return
	}
	preview, err := service.PreviewSource(c.Request.Context(), actorID, telegramAdminRequestID(), request.ChatRef, request.ConfirmationTicket)
	if err != nil {
		writeAdminTelegramManagementError(c, err)
		return
	}
	ok(c, preview)
}

// AdminTelegramConfirmSource rechecks a preview and persists only its canonical
// chat identity and display metadata.
func (a *API) AdminTelegramConfirmSource(c *gin.Context) {
	service, actorID, available := a.requireTelegramManagementService(c)
	if !available {
		return
	}
	var request adminTelegramSourceConfirmRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		bad(c, "请求 JSON 格式错误")
		return
	}
	source, err := service.ConfirmSource(c.Request.Context(), actorID, telegramAdminRequestID(), request.PreviewID, request.ChatRef, request.ConfirmationTicket)
	if err != nil {
		writeAdminTelegramManagementError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "msg": "", "data": source})
}

// AdminTelegramManagementSources lists persisted sources without accepting a
// direct raw-reference create operation.
func (a *API) AdminTelegramManagementSources(c *gin.Context) {
	service, _, available := a.requireTelegramManagementService(c)
	if !available {
		return
	}
	items, err := service.ListSources(c.Request.Context())
	if err != nil {
		writeAdminTelegramManagementError(c, err)
		return
	}
	ok(c, gin.H{"items": items})
}

// AdminTelegramManagementPauseSource pauses a source without clearing its cursor.
func (a *API) AdminTelegramManagementPauseSource(c *gin.Context) {
	a.updateTelegramManagementSource(c, func(service telegramManagementService, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
		return service.PauseSource(c.Request.Context(), actorID, sourceID)
	})
}

// AdminTelegramManagementResumeSource resumes a source at its current cursor.
func (a *API) AdminTelegramManagementResumeSource(c *gin.Context) {
	a.updateTelegramManagementSource(c, func(service telegramManagementService, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
		return service.ResumeSource(c.Request.Context(), actorID, sourceID)
	})
}

// AdminTelegramManagementRecoverSource restarts a source that ended in an
// error state without resetting its history cursor.
func (a *API) AdminTelegramManagementRecoverSource(c *gin.Context) {
	a.updateTelegramManagementSource(c, func(service telegramManagementService, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
		return service.RecoverSource(c.Request.Context(), actorID, sourceID)
	})
}

// AdminTelegramManagementStartBackfill clears a source cursor and starts a new history scan.
func (a *API) AdminTelegramManagementStartBackfill(c *gin.Context) {
	a.updateTelegramManagementSource(c, func(service telegramManagementService, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
		return service.StartBackfill(c.Request.Context(), actorID, sourceID)
	})
}

// AdminTelegramManagementProgress returns persisted media status counts.
func (a *API) AdminTelegramManagementProgress(c *gin.Context) {
	service, _, available := a.requireTelegramManagementService(c)
	if !available {
		return
	}
	sourceID, valid := parseUUID(c.Param("id"))
	if !valid || sourceID == uuid.Nil {
		bad(c, "Telegram 来源 ID 格式错误")
		return
	}
	progress, err := service.GetProgress(c.Request.Context(), sourceID)
	if err != nil {
		writeAdminTelegramManagementError(c, err)
		return
	}
	ok(c, progress)
}

// AdminTelegramAudits lists sanitized administrative history.
func (a *API) AdminTelegramAudits(c *gin.Context) {
	service, _, available := a.requireTelegramManagementService(c)
	if !available {
		return
	}
	items, total, err := service.ListAudits(c.Request.Context(), models.TelegramAuditFilter{
		Page:     parsePage(c.Query("page"), 1),
		PageSize: parsePageSize(c.Query("page_size"), 20),
		Action:   strings.TrimSpace(c.Query("action")),
		Result:   strings.TrimSpace(c.Query("result")),
	})
	if err != nil {
		writeAdminTelegramManagementError(c, err)
		return
	}
	ok(c, gin.H{"items": items, "total": total})
}

func (a *API) startTelegramAuthorization(c *gin.Context, start func(telegramManagementService, uuid.UUID, string, string) (telegram.AuthorizationView, error)) {
	service, actorID, available := a.requireTelegramManagementService(c)
	if !available {
		return
	}
	var request adminTelegramAuthorizationStartRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		bad(c, "请求 JSON 格式错误")
		return
	}
	view, err := start(service, actorID, telegramAdminRequestID(), request.ConfirmationTicket)
	if err != nil {
		writeAdminTelegramManagementError(c, err)
		return
	}
	ok(c, view)
}

func (a *API) updateTelegramManagementSource(c *gin.Context, update func(telegramManagementService, uuid.UUID, uuid.UUID) (models.TelegramSource, error)) {
	service, actorID, available := a.requireTelegramManagementService(c)
	if !available {
		return
	}
	sourceID, valid := parseUUID(c.Param("id"))
	if !valid || sourceID == uuid.Nil {
		bad(c, "Telegram 来源 ID 格式错误")
		return
	}
	source, err := update(service, actorID, sourceID)
	if err != nil {
		writeAdminTelegramManagementError(c, err)
		return
	}
	ok(c, source)
}

func (a *API) requireTelegramManagementService(c *gin.Context) (telegramManagementService, uuid.UUID, bool) {
	if a == nil || a.telegramManagementSvc == nil {
		response.Error(c, telegramAdminErrorCode, "Telegram 管理服务不可用")
		return nil, uuid.Nil, false
	}
	actorID, ok := middleware.UserIDFromContext(c)
	if !ok || actorID == uuid.Nil {
		response.Error(c, http.StatusUnauthorized, "管理员身份无效")
		return nil, uuid.Nil, false
	}
	return a.telegramManagementSvc, actorID, true
}

func telegramAdminRequestID() string {
	return uuid.NewString()
}

func writeAdminTelegramManagementError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrTelegramConfirmationRequired),
		errors.Is(err, services.ErrTelegramConfirmationInvalidPassword),
		errors.Is(err, services.ErrTelegramConfirmationNotFound),
		errors.Is(err, services.ErrTelegramConfirmationNotOwner),
		errors.Is(err, services.ErrTelegramConfirmationActionMismatch),
		errors.Is(err, services.ErrTelegramSourceAlreadyExists),
		errors.Is(err, services.ErrTelegramSourceUnresolved),
		errors.Is(err, services.ErrTelegramSourceTypeInvalid),
		errors.Is(err, services.ErrTelegramSourceNotRecoverable):
		bad(c, err.Error())
	case errors.Is(err, repository.ErrTelegramSourceNotFound):
		response.Error(c, http.StatusNotFound, "Telegram 来源不存在")
	case errors.Is(err, telegram.ErrAuthorizationInProgress):
		response.Error(c, http.StatusConflict, "Telegram 授权正在进行中")
	case errors.Is(err, telegram.ErrAuthorizationNotOwner):
		response.Error(c, http.StatusForbidden, "无权提交该 Telegram 授权")
	case errors.Is(err, telegram.ErrAuthorizationNotFound):
		response.Error(c, http.StatusNotFound, "Telegram 授权不存在")
	case errors.Is(err, telegram.ErrControlUnavailable):
		response.Error(c, http.StatusServiceUnavailable, "Telegram 采集器控制通道不可用")
	default:
		var controlErr *telegram.ControlRequestError
		if telegram.AsControlRequestError(err, &controlErr) {
			status := controlErr.StatusCode
			if status < http.StatusBadRequest || status >= http.StatusInternalServerError {
				status = telegramAdminErrorCode
			}
			response.Error(c, status, strings.TrimSpace(controlErr.Message))
			return
		}
		response.Error(c, telegramAdminErrorCode, strings.TrimSpace(err.Error()))
	}
}
