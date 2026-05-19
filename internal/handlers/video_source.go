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
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
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
		if path != "" {
			return path, nil
		}
		return strings.TrimSpace(video.TranscodedPath), nil
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
	thumbPath = chooseVideoThumbnailVariantPath(video.Type, video.Metadata, c.Query("variant"), thumbPath)
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

func chooseVideoThumbnailVariantPath(videoType string, rawMetadata []byte, requestedVariant, fallback string) string {
	videoType = strings.ToLower(strings.TrimSpace(videoType))
	requestedVariant = strings.ToLower(strings.TrimSpace(requestedVariant))
	if videoType == "av" {
		return chooseAVThumbnailVariantPath(rawMetadata, requestedVariant, fallback)
	}
	if videoType == "movie" && requestedVariant == "backdrop" {
		return chooseMovieBackdropVariantPath(rawMetadata, fallback)
	}
	return strings.TrimSpace(fallback)
}

func chooseMovieBackdropVariantPath(rawMetadata []byte, fallback string) string {
	fallback = strings.TrimSpace(fallback)
	if len(rawMetadata) == 0 {
		return fallback
	}
	var metadata map[string]any
	if err := json.Unmarshal(rawMetadata, &metadata); err != nil {
		return fallback
	}
	for _, key := range []string{"backdrop_url", "backdrop_path"} {
		candidate := strings.TrimSpace(stringFromAny(metadata[key]))
		if isLocalMovieBackdropFilePath(candidate) {
			return candidate
		}
	}
	if tmdb, ok := metadata["tmdb"].(map[string]any); ok {
		for _, key := range []string{"backdrop_url", "backdrop_path"} {
			candidate := strings.TrimSpace(stringFromAny(tmdb[key]))
			if isLocalMovieBackdropFilePath(candidate) {
				return candidate
			}
		}
	}
	return fallback
}

func isLocalMovieBackdropFilePath(candidate string) bool {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return false
	}
	if strings.HasPrefix(candidate, "http://") || strings.HasPrefix(candidate, "https://") {
		return false
	}
	if strings.HasPrefix(candidate, "/api/") {
		return false
	}
	normalized := strings.ReplaceAll(candidate, "\\", "/")
	return strings.Contains(normalized, "/videos/") && strings.Contains(strings.ToLower(normalized), "backdrop")
}

func chooseAVThumbnailVariantPath(rawMetadata []byte, requestedVariant, fallback string) string {
	fallback = strings.TrimSpace(fallback)
	if len(rawMetadata) == 0 {
		return fallback
	}
	var metadata map[string]any
	if err := json.Unmarshal(rawMetadata, &metadata); err != nil {
		return fallback
	}
	requestedVariant = strings.ToLower(strings.TrimSpace(requestedVariant))
	if requestedVariant == "" {
		requestedVariant = strings.ToLower(strings.TrimSpace(stringFromAny(metadata["poster_variant"])))
	}
	originalFilePath := strings.TrimSpace(stringFromAny(metadata["poster_original_file_path"]))
	thumbFilePath := strings.TrimSpace(stringFromAny(metadata["poster_thumb_file_path"]))
	croppedFilePath := strings.TrimSpace(stringFromAny(metadata["poster_cropped_file_path"]))
	switch requestedVariant {
	case "thumb":
		if thumbFilePath != "" {
			return thumbFilePath
		}
		if croppedFilePath != "" {
			return croppedFilePath
		}
		if originalFilePath != "" {
			return originalFilePath
		}
	case "cropped":
		if croppedFilePath != "" {
			return croppedFilePath
		}
		if thumbFilePath != "" {
			return thumbFilePath
		}
		if originalFilePath != "" {
			return originalFilePath
		}
	case "original":
		if originalFilePath != "" {
			return originalFilePath
		}
		if thumbFilePath != "" {
			return thumbFilePath
		}
		if croppedFilePath != "" {
			return croppedFilePath
		}
	default:
		if thumbFilePath != "" {
			return thumbFilePath
		}
		if croppedFilePath != "" {
			return croppedFilePath
		}
		if originalFilePath != "" {
			return originalFilePath
		}
	}
	return fallback
}
