package handlers

import (
	"github.com/gin-gonic/gin"

	"video-server/internal/middleware"
	"video-server/internal/response"
)

func (a *API) ToggleLike(c *gin.Context) {
	a.toggleAction(c, "like")
}

func (a *API) ToggleFavorite(c *gin.Context) {
	a.toggleAction(c, "favorite")
}

func (a *API) ToggleDislike(c *gin.Context) {
	a.toggleAction(c, "dislike")
}

func (a *API) toggleAction(c *gin.Context, action string) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	videoID, okVideo := parseUUID(c.Param("id"))
	if !okVideo {
		bad(c, "invalid video id")
		return
	}
	enabled, err := a.appSvc.ToggleInteraction(c.Request.Context(), userID, videoID, action)
	if err != nil {
		response.Error(c, 24, err.Error())
		return
	}
	ok(c, gin.H{"action": action, "enabled": enabled})
}
