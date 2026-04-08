package handlers

import (
	"github.com/gin-gonic/gin"

	"video-server/internal/middleware"
	"video-server/internal/response"
)

func (a *API) RecordHistory(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	var req struct {
		VideoID      string `json:"video_id"`
		WatchSeconds int    `json:"watch_seconds"`
		Completed    bool   `json:"completed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	videoID, okVideo := parseUUID(req.VideoID)
	if !okVideo {
		bad(c, "invalid video_id")
		return
	}
	if err := a.appSvc.RecordHistory(c.Request.Context(), userID, videoID, req.WatchSeconds, req.Completed); err != nil {
		response.Error(c, 21, err.Error())
		return
	}
	ok(c, gin.H{"saved": true})
}

func (a *API) ContinueHistory(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("limit"), 20)
	result, err := a.appSvc.ContinueWatching(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, 22, err.Error())
		return
	}
	ok(c, result)
}

func (a *API) DeleteHistory(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	videoID, okVideo := parseUUID(c.Param("video_id"))
	if !okVideo {
		bad(c, "invalid video id")
		return
	}
	if err := a.appSvc.DeleteHistory(c.Request.Context(), userID, videoID); err != nil {
		response.Error(c, 23, err.Error())
		return
	}
	ok(c, gin.H{"deleted": true})
}
