package handlers

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"video-server/internal/middleware"
	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/response"
	"video-server/internal/services"
)

func (a *API) ListTVDevices(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	payload, err := a.appSvc.ListTVDevices(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, 2301, err.Error())
		return
	}
	ok(c, payload)
}

func (a *API) CreateTVRemoteSession(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	var req struct {
		DeviceID      string                        `json:"device_id"`
		Items         []models.TvRemoteSessionItem  `json:"items"`
		CurrentIndex  int                           `json:"current_index"`
		SearchContext *models.TvRemoteSearchContext `json:"search_context"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	payload, err := a.appSvc.StartTVRemoteSession(c.Request.Context(), userID, strings.TrimSpace(req.DeviceID), req.Items, req.CurrentIndex, req.SearchContext)
	if err != nil {
		switch {
		case repository.IsNotFound(err):
			response.Error(c, 404, "tv device not found")
		case errors.Is(err, services.ErrTVRemoteDeviceOffline):
			response.Error(c, 2302, err.Error())
		case errors.Is(err, services.ErrTVRemoteItemsRequired), errors.Is(err, services.ErrTVRemoteIndexOutOfRange):
			response.Error(c, 2303, err.Error())
		default:
			response.Error(c, 2304, err.Error())
		}
		return
	}
	ok(c, payload)
}

func (a *API) GetTVRemoteSession(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	sessionID, okSession := parseUUID(c.Param("session_id"))
	if !okSession {
		bad(c, "invalid session id")
		return
	}
	payload, err := a.appSvc.GetTVRemoteSession(c.Request.Context(), userID, sessionID)
	if err != nil {
		if repository.IsNotFound(err) {
			response.Error(c, 404, "session not found")
			return
		}
		response.Error(c, 2305, err.Error())
		return
	}
	ok(c, payload)
}

func (a *API) TVRemotePrevious(c *gin.Context) {
	a.stepTVRemoteSession(c, -1)
}

func (a *API) TVRemoteNext(c *gin.Context) {
	a.stepTVRemoteSession(c, 1)
}

func (a *API) TVRemoteAutoNext(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	sessionID, okSession := parseUUID(c.Param("session_id"))
	if !okSession {
		bad(c, "invalid session id")
		return
	}
	payload, err := a.appSvc.AutoNextTVRemoteSession(c.Request.Context(), userID, sessionID)
	if err != nil {
		switch {
		case repository.IsNotFound(err):
			response.Error(c, 404, "session not found")
		case errors.Is(err, services.ErrTVRemoteSessionInactive):
			response.Error(c, 2314, err.Error())
		default:
			response.Error(c, 2315, err.Error())
		}
		return
	}
	ok(c, payload)
}

func (a *API) stepTVRemoteSession(c *gin.Context, delta int) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	sessionID, okSession := parseUUID(c.Param("session_id"))
	if !okSession {
		bad(c, "invalid session id")
		return
	}
	payload, err := a.appSvc.StepTVRemoteSession(c.Request.Context(), userID, sessionID, delta)
	if err != nil {
		switch {
		case repository.IsNotFound(err):
			response.Error(c, 404, "session not found")
		case errors.Is(err, services.ErrTVRemoteSessionInactive):
			response.Error(c, 2306, err.Error())
		default:
			response.Error(c, 2307, err.Error())
		}
		return
	}
	ok(c, payload)
}

func (a *API) GetCurrentTVRemoteSessionForDevice(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	payload, err := a.appSvc.GetCurrentTVRemoteSessionForDevice(
		c.Request.Context(),
		userID,
		c.Query("device_id"),
		c.Query("legacy_device_id"),
	)
	if err != nil {
		if repository.IsNotFound(err) {
			response.Error(c, 404, "tv device not found")
			return
		}
		response.Error(c, 2308, err.Error())
		return
	}
	ok(c, models.TvRemoteDeviceSessionPayload{Session: payload})
}

func (a *API) UpdateTVRemoteSessionCurrentIndex(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	sessionID, okSession := parseUUID(c.Param("session_id"))
	if !okSession {
		bad(c, "invalid session id")
		return
	}
	var req struct {
		CurrentIndex int `json:"current_index"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	payload, err := a.appSvc.UpdateTVRemoteSessionCurrentIndex(c.Request.Context(), userID, sessionID, req.CurrentIndex)
	if err != nil {
		switch {
		case repository.IsNotFound(err):
			response.Error(c, 404, "session not found")
		case errors.Is(err, services.ErrTVRemoteIndexOutOfRange), errors.Is(err, services.ErrTVRemoteSessionInactive):
			response.Error(c, 2309, err.Error())
		default:
			response.Error(c, 2310, err.Error())
		}
		return
	}
	ok(c, payload)
}

func (a *API) UpdateTVRemoteSessionAutoplayNext(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	sessionID, okSession := parseUUID(c.Param("session_id"))
	if !okSession {
		bad(c, "invalid session id")
		return
	}
	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Enabled == nil {
		bad(c, "invalid payload")
		return
	}
	payload, err := a.appSvc.UpdateTVRemoteSessionAutoplayNext(c.Request.Context(), userID, sessionID, *req.Enabled)
	if err != nil {
		switch {
		case repository.IsNotFound(err):
			response.Error(c, 404, "session not found")
		case errors.Is(err, services.ErrTVRemoteSessionInactive):
			response.Error(c, 2312, err.Error())
		default:
			response.Error(c, 2313, err.Error())
		}
		return
	}
	ok(c, payload)
}

func (a *API) EndTVRemoteSession(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	sessionID, okSession := parseUUID(c.Param("session_id"))
	if !okSession {
		bad(c, "invalid session id")
		return
	}
	var req struct {
		EndedReason string `json:"ended_reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := a.appSvc.EndTVRemoteSession(c.Request.Context(), userID, sessionID, req.EndedReason); err != nil {
		if repository.IsNotFound(err) {
			response.Error(c, 404, "session not found")
			return
		}
		response.Error(c, 2311, err.Error())
		return
	}
	ok(c, gin.H{"ended": true})
}
