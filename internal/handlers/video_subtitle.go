package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"video-server/internal/repository"
)

func (a *API) VideoSubtitleFile(c *gin.Context) {
	videoID, okVideo := parseUUID(c.Param("id"))
	if !okVideo {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid video id"})
		return
	}
	subtitleID, okSubtitle := parseUUID(c.Param("subtitle_id"))
	if !okSubtitle {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid subtitle id"})
		return
	}
	video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
	if err != nil {
		if repository.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"msg": "video not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "query video failed"})
		return
	}
	if video.Status != "ready" {
		c.JSON(http.StatusConflict, gin.H{"msg": "video not ready"})
		return
	}
	subtitle, err := a.repo.GetVideoSubtitle(c.Request.Context(), subtitleID)
	if err != nil {
		if repository.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"msg": "subtitle not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "query subtitle failed"})
		return
	}
	if subtitle.VideoID != videoID || strings.TrimSpace(subtitle.StoredPath) == "" {
		c.JSON(http.StatusNotFound, gin.H{"msg": "subtitle not found"})
		return
	}
	if subtitle.MIMEType != "" {
		c.Header("Content-Type", subtitle.MIMEType)
	}
	if !tryServeLocalImagePath(c, subtitle.StoredPath, "subtitle file not found", "read subtitle file failed") {
		return
	}
}
