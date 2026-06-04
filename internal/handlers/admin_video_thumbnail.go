package handlers

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/repository"
	"video-server/internal/response"
	"video-server/pkg/ffmpeg"
)

func (a *API) AdminCaptureVideoThumbnail(c *gin.Context) {
	videoID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid video id")
		return
	}

	var req struct {
		TimeSeconds float64 `json:"time_seconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	if req.TimeSeconds < 0 || math.IsNaN(req.TimeSeconds) || math.IsInf(req.TimeSeconds, 0) {
		bad(c, "time_seconds must be a non-negative number")
		return
	}

	video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
	if err != nil {
		if repository.IsNotFound(err) {
			response.Error(c, 404, "video not found")
			return
		}
		response.Error(c, 1060, err.Error())
		return
	}
	if strings.ToLower(strings.TrimSpace(video.Status)) != "ready" {
		response.Error(c, 1061, "video not ready")
		return
	}

	sourcePath := strings.TrimSpace(video.TranscodedPath)
	if sourcePath == "" {
		sourcePath = strings.TrimSpace(video.OriginalPath)
	}
	if sourcePath == "" {
		response.Error(c, 1062, "video source not found")
		return
	}
	file, _, err := openLocalImageFile(c.Request.Context(), sourcePath)
	if err != nil {
		response.Error(c, 1063, fmt.Sprintf("video source missing: %v", err))
		return
	}
	_ = file.Close()

	captureAt := req.TimeSeconds
	if video.DurationSeconds > 0 {
		maxSeconds := float64(video.DurationSeconds)
		if captureAt >= maxSeconds {
			if maxSeconds > 0.2 {
				captureAt = maxSeconds - 0.2
			} else {
				captureAt = 0
			}
		}
	}
	if captureAt < 0 {
		captureAt = 0
	}

	targetPath := resolveVideoThumbnailPath(a.storageRoot, videoID, video.ThumbnailPath)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		response.Error(c, 1064, err.Error())
		return
	}

	tempPath := buildCaptureTempPath(targetPath)
	defer func() {
		_ = os.Remove(tempPath)
	}()
	thumbCtx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()
	if err := ffmpeg.ThumbnailAt(thumbCtx, sourcePath, tempPath, captureAt); err != nil {
		response.Error(c, 1065, err.Error())
		return
	}
	if err := replaceFileAtomic(tempPath, targetPath); err != nil {
		response.Error(c, 1066, err.Error())
		return
	}
	if err := a.repo.AdminUpdateVideoThumbnail(c.Request.Context(), videoID, targetPath); err != nil {
		response.Error(c, 1067, err.Error())
		return
	}

	ok(c, gin.H{
		"video_id":            videoID,
		"thumbnail_path":      targetPath,
		"captured_at_seconds": captureAt,
	})
}

func resolveVideoThumbnailPath(storageRoot string, videoID uuid.UUID, currentPath string) string {
	path := strings.TrimSpace(currentPath)
	lower := strings.ToLower(path)
	if path != "" && !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return path
	}
	return filepath.Join(storageRoot, "videos", videoID.String(), "thumb.jpg")
}

func buildCaptureTempPath(targetPath string) string {
	dir := filepath.Dir(targetPath)
	ext := filepath.Ext(targetPath)
	if strings.TrimSpace(ext) == "" {
		ext = ".jpg"
	}
	base := strings.TrimSuffix(filepath.Base(targetPath), filepath.Ext(targetPath))
	if strings.TrimSpace(base) == "" || base == "." {
		base = "thumb"
	}
	return filepath.Join(dir, "."+base+".tmp-"+uuid.New().String()+ext)
}

func replaceFileAtomic(tempPath, targetPath string) error {
	if err := os.Rename(tempPath, targetPath); err == nil {
		return nil
	}
	if err := os.Remove(targetPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove old thumbnail: %w", err)
	}
	if err := os.Rename(tempPath, targetPath); err != nil {
		return fmt.Errorf("replace thumbnail file: %w", err)
	}
	return nil
}
