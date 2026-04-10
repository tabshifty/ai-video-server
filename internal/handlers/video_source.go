package handlers

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/middleware"
	"video-server/internal/repository"
	"video-server/internal/utils"
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

	sourcePath, ok := a.resolvePlayableSource(c, videoID)
	if !ok {
		return
	}
	c.File(sourcePath)
}

func (a *API) VideoSourceSigned(c *gin.Context) {
	videoID, okID := parseUUID(c.Param("id"))
	if !okID {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid video id"})
		return
	}
	expRaw := strings.TrimSpace(c.Query("exp"))
	sig := strings.TrimSpace(c.Query("sig"))
	if expRaw == "" || sig == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "missing exp or sig"})
		return
	}
	exp, err := strconv.ParseInt(expRaw, 10, 64)
	if err != nil || exp <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid exp"})
		return
	}
	if time.Now().Unix() > exp {
		c.JSON(http.StatusGone, gin.H{"msg": "signed url expired"})
		return
	}
	if !utils.VerifyVideoSourceSign(a.playSignSecret, videoID, exp, sig) {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid signature"})
		return
	}

	sourcePath, ok := a.resolvePlayableSource(c, videoID)
	if !ok {
		return
	}
	c.File(sourcePath)
}

func (a *API) resolvePlayableSource(c *gin.Context, videoID uuid.UUID) (string, bool) {
	video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
	if err != nil {
		if repository.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"msg": "video not found"})
			return "", false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "query video failed"})
		return "", false
	}
	if video.Status != "ready" {
		c.JSON(http.StatusConflict, gin.H{"msg": "video not ready"})
		return "", false
	}
	sourcePath := strings.TrimSpace(video.TranscodedPath)
	if sourcePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"msg": "video source not found"})
		return "", false
	}
	if _, statErr := os.Stat(sourcePath); statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			c.JSON(http.StatusNotFound, gin.H{"msg": "video source not found"})
			return "", false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "read video source failed"})
		return "", false
	}
	return sourcePath, true
}
