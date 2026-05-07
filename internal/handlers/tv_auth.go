package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"video-server/internal/middleware"
	"video-server/internal/repository"
	"video-server/internal/response"
)

func (a *API) CreateTVAuthSession(c *gin.Context) {
	var req struct {
		DeviceID   string `json:"device_id"`
		DeviceName string `json:"device_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	serverBaseURL := inferServerBaseURL(c)
	payload, err := a.appSvc.CreateTVAuthSession(
		c.Request.Context(),
		strings.TrimSpace(req.DeviceID),
		strings.TrimSpace(req.DeviceName),
		serverBaseURL,
	)
	if err != nil {
		response.Error(c, 2201, err.Error())
		return
	}
	ok(c, payload)
}

func (a *API) GetTVAuthSession(c *gin.Context) {
	sessionID, okSession := parseUUID(c.Param("session_id"))
	if !okSession {
		bad(c, "invalid session id")
		return
	}
	payload, err := a.appSvc.PollTVAuthSession(c.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || repository.IsNotFound(err) {
			response.Error(c, 404, "session not found")
			return
		}
		response.Error(c, 2202, err.Error())
		return
	}
	payload.ServerBaseURL = inferServerBaseURL(c)
	ok(c, payload)
}

func (a *API) ApproveTVAuthSession(c *gin.Context) {
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
	user, err := a.repo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, 2203, err.Error())
		return
	}
	tokens, err := a.issueTokens(c, user.ID, user.Role)
	if err != nil {
		response.Error(c, 2204, err.Error())
		return
	}
	if err := a.appSvc.ApproveTVAuthSession(c.Request.Context(), sessionID, user, tokens); err != nil {
		if repository.IsNotFound(err) {
			response.Error(c, 404, "session not found or not pending")
			return
		}
		response.Error(c, 2205, err.Error())
		return
	}
	ok(c, gin.H{"approved": true})
}

func (a *API) DenyTVAuthSession(c *gin.Context) {
	_, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	sessionID, okSession := parseUUID(c.Param("session_id"))
	if !okSession {
		bad(c, "invalid session id")
		return
	}
	if err := a.appSvc.DenyTVAuthSession(c.Request.Context(), sessionID); err != nil {
		if repository.IsNotFound(err) {
			response.Error(c, 404, "session not found or not pending")
			return
		}
		response.Error(c, 2206, err.Error())
		return
	}
	ok(c, gin.H{"denied": true})
}

func inferServerBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, strings.TrimSpace(c.Request.Host))
}
