package handlers

import (
	"github.com/gin-gonic/gin"

	"video-server/internal/middleware"
	"video-server/internal/response"
)

// RandomShort returns random short video feed.
func (a *API) RandomShort(c *gin.Context) {
	pageSize := parsePageSize(c.Query("page_size"), 20)
	videos, err := a.recSvc.RandomShortFeed(c.Request.Context(), pageSize)
	if err != nil {
		response.Error(c, 7, err.Error())
		return
	}
	ok(c, gin.H{"items": videos})
}

// Recommend returns personalized recommendation feed.
func (a *API) Recommend(c *gin.Context) {
	userID, okID := middleware.UserIDFromContext(c)
	if !okID {
		response.Error(c, 401, "unauthorized")
		return
	}
	pageSize := parsePageSize(c.Query("page_size"), 20)

	videos, err := a.recSvc.Recommend(c.Request.Context(), userID, pageSize)
	if err != nil {
		response.Error(c, 8, err.Error())
		return
	}
	ok(c, gin.H{"items": videos})
}

// RecordAction stores user feedback events for recommendation.
func (a *API) RecordAction(c *gin.Context) {
	var req struct {
		VideoID      string `json:"video_id"`
		ActionType   string `json:"action_type"`
		WatchSeconds int    `json:"watch_seconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	userID, okUser := middleware.UserIDFromContext(c)
	videoID, okVideo := parseUUID(req.VideoID)
	if !okUser || !okVideo {
		bad(c, "invalid auth user or video_id")
		return
	}
	if req.ActionType == "" {
		bad(c, "action_type is required")
		return
	}

	if err := a.repo.UpsertAction(c.Request.Context(), userID, videoID, req.ActionType, req.WatchSeconds); err != nil {
		response.Error(c, 9, err.Error())
		return
	}
	ok(c, gin.H{"saved": true})
}
