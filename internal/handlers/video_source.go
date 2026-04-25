package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/middleware"
	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/utils"
)

var (
	errInvalidPlaybackProfile = errors.New("invalid playback profile")
	errCompatSourceNotFound   = errors.New("video compat source not found")
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

func (a *API) VideoThumbnail(c *gin.Context) {
	videoID, okID := parseUUID(c.Param("id"))
	if !okID {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid video id"})
		return
	}

	thumbPath, ok := a.resolveThumbnailSource(c, videoID)
	if !ok {
		return
	}
	c.File(thumbPath)
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
	sourcePath, err := resolveProfiledPlayableSource(video, c.Query("profile"))
	if err != nil {
		if errors.Is(err, errCompatSourceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"msg": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		}
		return "", false
	}
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

func resolveProfiledPlayableSource(video models.Video, requestedProfile string) (string, error) {
	switch normalizePlaybackProfile(requestedProfile) {
	case "", "primary":
		return strings.TrimSpace(video.TranscodedPath), nil
	case "compat":
		path := compatTranscodedPathFromMetadata(video.Metadata)
		if path == "" {
			return "", errCompatSourceNotFound
		}
		return path, nil
	default:
		return "", errInvalidPlaybackProfile
	}
}

func normalizePlaybackProfile(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

func compatTranscodedPathFromMetadata(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	if direct, ok := payload["compat_transcoded_path"].(string); ok {
		return strings.TrimSpace(direct)
	}
	profiles, ok := payload["playback_profiles"].(map[string]any)
	if !ok {
		return ""
	}
	compat, ok := profiles["compat"].(map[string]any)
	if !ok {
		return ""
	}
	path, _ := compat["path"].(string)
	return strings.TrimSpace(path)
}

func (a *API) resolveThumbnailSource(c *gin.Context, videoID uuid.UUID) (string, bool) {
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
	thumbPath := strings.TrimSpace(video.ThumbnailPath)
	if thumbPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"msg": "thumbnail not found"})
		return "", false
	}
	if _, statErr := os.Stat(thumbPath); statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			c.JSON(http.StatusNotFound, gin.H{"msg": "thumbnail not found"})
			return "", false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "read thumbnail failed"})
		return "", false
	}
	return thumbPath, true
}
