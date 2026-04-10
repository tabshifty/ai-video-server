package handlers

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"video-server/internal/middleware"
	"video-server/internal/repository"
)

func (a *API) VideoSource(c *gin.Context) {
	if _, ok := middleware.UserIDFromContext(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "unauthorized"})
		return
	}

	videoID, okID := parseUUID(c.Param("id"))
	if !okID {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid video id"})
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

	sourcePath := strings.TrimSpace(video.TranscodedPath)
	if sourcePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"msg": "video source not found"})
		return
	}
	if _, statErr := os.Stat(sourcePath); statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			c.JSON(http.StatusNotFound, gin.H{"msg": "video source not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "read video source failed"})
		return
	}

	c.File(sourcePath)
}
