package handlers

import (
	"errors"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"video-server/internal/repository"
	"video-server/internal/response"
)

func (a *API) AdminVideoSubtitles(c *gin.Context) {
	videoID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid video id")
		return
	}
	items, err := a.repo.AdminListVideoSubtitles(c.Request.Context(), videoID)
	if err != nil {
		response.Error(c, 1040, err.Error())
		return
	}
	ok(c, gin.H{"items": items})
}

func (a *API) AdminUploadVideoSubtitle(c *gin.Context) {
	videoID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid video id")
		return
	}
	video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
	if err != nil {
		response.Error(c, 1041, err.Error())
		return
	}
	if !subtitleSupportedVideoType(video.Type) {
		bad(c, "当前视频类型不支持字幕管理")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		bad(c, "subtitle file is required")
		return
	}
	subtitle, err := a.subtitleSvc.SaveUploadedSubtitle(
		c.Request.Context(),
		videoID,
		file,
		c.PostForm("language_code"),
		c.PostForm("label"),
		c.PostForm("is_default") == "1" || strings.EqualFold(c.PostForm("is_default"), "true"),
	)
	if err != nil {
		response.Error(c, 1042, err.Error())
		return
	}
	ok(c, subtitle)
}

func (a *API) AdminRescanVideoSubtitles(c *gin.Context) {
	videoID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid video id")
		return
	}
	video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
	if err != nil {
		response.Error(c, 1043, err.Error())
		return
	}
	if !subtitleSupportedVideoType(video.Type) {
		bad(c, "当前视频类型不支持字幕管理")
		return
	}
	inputPath := strings.TrimSpace(video.OriginalPath)
	if inputPath == "" {
		inputPath = strings.TrimSpace(video.TranscodedPath)
	}
	items, err := a.subtitleSvc.SyncEmbeddedSubtitles(c.Request.Context(), videoID, inputPath)
	if err != nil {
		response.Error(c, 1044, err.Error())
		return
	}
	ok(c, gin.H{"items": items})
}

func (a *API) AdminUpdateVideoSubtitle(c *gin.Context) {
	videoID, okVideo := parseUUID(c.Param("id"))
	if !okVideo {
		bad(c, "invalid video id")
		return
	}
	subtitleID, okSubtitle := parseUUID(c.Param("subtitle_id"))
	if !okSubtitle {
		bad(c, "invalid subtitle id")
		return
	}
	var req struct {
		LanguageCode string `json:"language_code"`
		Label        string `json:"label"`
		SortOrder    int    `json:"sort_order"`
		IsDefault    *bool  `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	subtitle, err := a.repo.GetVideoSubtitle(c.Request.Context(), subtitleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || repository.IsNotFound(err) {
			response.Error(c, 404, "subtitle not found")
			return
		}
		response.Error(c, 1045, err.Error())
		return
	}
	if subtitle.VideoID != videoID {
		response.Error(c, 404, "subtitle not found")
		return
	}
	if req.IsDefault != nil && subtitle.SourceType != "uploaded" && *req.IsDefault {
		bad(c, "仅外挂字幕可设置为默认轨")
		return
	}
	if err := a.repo.UpdateVideoSubtitleMetadata(c.Request.Context(), subtitleID, req.LanguageCode, req.Label, req.SortOrder); err != nil {
		response.Error(c, 1046, err.Error())
		return
	}
	if req.IsDefault != nil {
		if err := a.repo.SetVideoSubtitleDefault(c.Request.Context(), subtitleID, *req.IsDefault); err != nil {
			response.Error(c, 1047, err.Error())
			return
		}
	}
	updated, err := a.repo.GetVideoSubtitle(c.Request.Context(), subtitleID)
	if err != nil {
		response.Error(c, 1048, err.Error())
		return
	}
	ok(c, updated)
}

func (a *API) AdminDeleteVideoSubtitle(c *gin.Context) {
	videoID, okVideo := parseUUID(c.Param("id"))
	if !okVideo {
		bad(c, "invalid video id")
		return
	}
	subtitleID, okSubtitle := parseUUID(c.Param("subtitle_id"))
	if !okSubtitle {
		bad(c, "invalid subtitle id")
		return
	}
	subtitle, err := a.repo.GetVideoSubtitle(c.Request.Context(), subtitleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || repository.IsNotFound(err) {
			response.Error(c, 404, "subtitle not found")
			return
		}
		response.Error(c, 1049, err.Error())
		return
	}
	if subtitle.VideoID != videoID {
		response.Error(c, 404, "subtitle not found")
		return
	}
	if subtitle.SourceType != "uploaded" {
		bad(c, "仅允许删除外挂字幕")
		return
	}
	if err := a.repo.DeleteVideoSubtitle(c.Request.Context(), subtitleID); err != nil {
		response.Error(c, 1050, err.Error())
		return
	}
	if strings.TrimSpace(subtitle.StoredPath) != "" {
		_ = os.Remove(subtitle.StoredPath)
	}
	ok(c, gin.H{"deleted": true, "subtitle_id": subtitleID})
}

func subtitleSupportedVideoType(videoType string) bool {
	switch strings.TrimSpace(strings.ToLower(videoType)) {
	case "movie", "episode", "av":
		return true
	default:
		return false
	}
}
