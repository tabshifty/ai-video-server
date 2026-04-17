package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/middleware"
	"video-server/internal/queue"
	"video-server/internal/repository"
	"video-server/internal/response"
	"video-server/internal/services"
)

// Upload handles media upload and transcode task enqueue.
// @Summary Upload video
// @Tags upload
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param file formData file true "video file"
// @Param type formData string true "short|movie|episode|av"
// @Param title formData string false "title"
// @Param description formData string false "description"
// @Param tags formData string false "tags json array or csv"
// @Param actor_ids formData string false "actor ids json array or csv"
// @Param actor_names formData string false "actor names json array or csv"
// @Param hash formData string true "sha256 hash"
// @Success 202 {object} APIResponse
// @Failure 200 {object} APIResponse
// @Router /upload [post]
func (a *API) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		bad(c, "file is required")
		return
	}
	typ := c.PostForm("type")
	if typ != "short" && typ != "movie" && typ != "episode" && typ != "av" {
		bad(c, "type must be one of: short, movie, episode, av")
		return
	}
	hash := strings.TrimSpace(c.PostForm("hash"))
	if !isSHA256Hex(hash) {
		bad(c, "invalid hash")
		return
	}
	if !isAllowedVideoExt(file.Filename) {
		bad(c, "unsupported file extension")
		return
	}
	if ct := strings.ToLower(strings.TrimSpace(file.Header.Get("Content-Type"))); ct != "" && !strings.HasPrefix(ct, "video/") {
		bad(c, "unsupported content type")
		return
	}
	if a.maxVideoSize > 0 && file.Size > a.maxVideoSize {
		bad(c, "file size exceeds limit")
		return
	}

	title := c.PostForm("title")
	desc := c.PostForm("description")
	tags := parseUploadTags(c.PostForm("tags"))
	actorIDs, err := parseUploadActorIDs(c.PostForm("actor_ids"))
	if err != nil {
		bad(c, "演员ID格式错误")
		return
	}
	actorNames := parseUploadStringList(c.PostForm("actor_names"))
	collectionIDs, err := parseUploadCollectionIDs(c.PostForm("collection_ids"))
	if err != nil {
		bad(c, "合集ID格式错误")
		return
	}
	if err := validateCollectionsForType(typ, collectionIDs); err != nil {
		bad(c, "仅短视频支持合集")
		return
	}
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	role, okRole := middleware.RoleFromContext(c)
	if !okRole {
		response.Error(c, 401, "unauthorized")
		return
	}
	if role != "admin" && typ != "short" {
		response.Error(c, 403, "insufficient role for video type")
		return
	}

	result, err := a.uploadSvc.SaveUpload(c.Request.Context(), userID, file, title, desc, typ, tags, actorIDs, actorNames, collectionIDs, hash, a.maxVideoSize)
	if err != nil {
		a.logger.Error("upload failed", "error", err)
		switch {
		case errors.Is(err, services.ErrHashMismatch):
			bad(c, "hash mismatch")
		case errors.Is(err, services.ErrFileTooLarge):
			bad(c, "file size exceeds limit")
		case errors.Is(err, services.ErrInvalidType):
			bad(c, "invalid type")
		case errors.Is(err, services.ErrHashTypeConflict):
			response.Error(c, 409, "hash exists with another video type")
		case errors.Is(err, repository.ErrCollectionsOnlyForShort):
			bad(c, "仅短视频支持合集")
		default:
			response.Error(c, 2, err.Error())
		}
		return
	}

	if result.AlreadyExists {
		ok(c, gin.H{
			"exists":   true,
			"video_id": result.VideoID,
		})
		return
	}

	if typ == "short" && result.Enqueue {
		payload := queue.TranscodePayload{
			VideoID:      result.VideoID.String(),
			InputPath:    result.InputPath,
			OutputDir:    result.OutputDir,
			TargetFormat: result.TargetFormat,
		}
		if err := a.enqueuer.EnqueueTranscode(payload); err != nil {
			a.logger.Error("enqueue transcode failed", "video_id", result.VideoID, "error", err)
			response.Error(c, 3, err.Error())
			return
		}
	}
	if typ == "movie" || typ == "episode" || typ == "av" {
		payload := queue.ScrapePayload{
			VideoID:  result.VideoID.String(),
			FilePath: result.InputPath,
			Filename: file.Filename,
			Type:     typ,
		}
		if typ == "movie" {
			if err := a.enqueuer.EnqueueScrapeMovie(payload); err != nil {
				a.logger.Error("enqueue scrape movie failed", "video_id", result.VideoID, "error", err)
				response.Error(c, 3, err.Error())
				return
			}
		} else if typ == "episode" {
			if err := a.enqueuer.EnqueueScrapeTV(payload); err != nil {
				a.logger.Error("enqueue scrape tv failed", "video_id", result.VideoID, "error", err)
				response.Error(c, 3, err.Error())
				return
			}
		} else {
			if err := a.enqueuer.EnqueueScrapeAV(payload); err != nil {
				a.logger.Error("enqueue scrape av failed", "video_id", result.VideoID, "error", err)
				response.Error(c, 3, err.Error())
				return
			}
		}
	}

	respStatus := result.Status
	if typ == "movie" || typ == "episode" || typ == "av" {
		respStatus = "scraping"
	}

	c.JSON(http.StatusAccepted, gin.H{
		"code": 0,
		"msg":  "upload accepted",
		"data": gin.H{
			"video_id": result.VideoID,
			"status":   respStatus,
		},
	})
}

var hashPattern = regexp.MustCompile("^[a-fA-F0-9]{64}$")
var errInvalidActorID = errors.New("invalid actor id")
var errInvalidCollectionID = errors.New("invalid collection id")
var errCollectionsOnlyForShort = errors.New("collections only support short videos")

func isSHA256Hex(raw string) bool {
	return hashPattern.MatchString(strings.TrimSpace(raw))
}

func parseUploadTags(raw string) []string {
	list := parseUploadStringList(raw)
	if len(list) == 0 {
		return nil
	}
	out := make([]string, 0, len(list))
	seen := map[string]struct{}{}
	for _, item := range list {
		tag := strings.ToLower(strings.TrimSpace(item))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

func parseUploadStringList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var rawList []string
	if strings.HasPrefix(raw, "[") {
		if err := json.Unmarshal([]byte(raw), &rawList); err == nil {
			out := make([]string, 0, len(rawList))
			seen := map[string]struct{}{}
			for _, item := range rawList {
				v := strings.Join(strings.Fields(strings.TrimSpace(item)), " ")
				if v == "" {
					continue
				}
				key := strings.ToLower(v)
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				out = append(out, v)
			}
			return out
		}
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		v := strings.Join(strings.Fields(strings.TrimSpace(part)), " ")
		if v == "" {
			continue
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	return out
}

func parseUploadActorIDs(raw string) ([]uuid.UUID, error) {
	list := parseUploadStringList(raw)
	if len(list) == 0 {
		return nil, nil
	}
	out := make([]uuid.UUID, 0, len(list))
	seen := map[uuid.UUID]struct{}{}
	for _, item := range list {
		id, err := uuid.Parse(item)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errInvalidActorID, item)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func parseUploadCollectionIDs(raw string) ([]uuid.UUID, error) {
	list := parseUploadStringList(raw)
	if len(list) == 0 {
		return nil, nil
	}
	out := make([]uuid.UUID, 0, len(list))
	seen := map[uuid.UUID]struct{}{}
	for _, item := range list {
		id, err := uuid.Parse(item)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errInvalidCollectionID, item)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func validateCollectionsForType(typ string, collectionIDs []uuid.UUID) error {
	if len(collectionIDs) == 0 {
		return nil
	}
	if strings.ToLower(strings.TrimSpace(typ)) != "short" {
		return errCollectionsOnlyForShort
	}
	return nil
}

func isAllowedVideoExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(name)))
	switch ext {
	case ".mp4", ".mov", ".mkv", ".avi", ".wmv", ".flv", ".webm", ".m4v":
		return true
	default:
		return false
	}
}

// Scrape handles movie/series metadata search and sync.
func (a *API) Scrape(c *gin.Context) {
	var req struct {
		VideoID       string `json:"video_id"`
		FilePath      string `json:"file_path"`
		TMDBID        int    `json:"tmdb_id"`
		MediaType     string `json:"media_type"` // movie | tv
		SeasonNumber  int    `json:"season_number"`
		EpisodeNumber int    `json:"episode_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}

	if req.TMDBID == 0 {
		if req.FilePath == "" {
			bad(c, "file_path is required when tmdb_id is empty")
			return
		}
		results, err := a.scrapeSvc.SearchMovieCandidates(c.Request.Context(), req.FilePath)
		if err != nil {
			response.Error(c, 4, err.Error())
			return
		}
		ok(c, gin.H{"candidates": results})
		return
	}

	var videoIDStr string
	if req.VideoID != "" {
		videoIDStr = req.VideoID
	} else if req.FilePath != "" {
		video, err := a.repo.GetVideoByOriginalPath(c.Request.Context(), req.FilePath)
		if err != nil {
			response.Error(c, 5, err.Error())
			return
		}
		videoIDStr = video.ID.String()
	} else {
		bad(c, "video_id or file_path is required")
		return
	}

	videoID, okID := parseUUID(videoIDStr)
	if !okID {
		bad(c, "invalid video_id")
		return
	}

	mediaType := req.MediaType
	if mediaType == "" {
		mediaType = "movie"
	}

	if mediaType == "tv" {
		if req.SeasonNumber <= 0 || req.EpisodeNumber <= 0 {
			bad(c, "season_number and episode_number are required for tv")
			return
		}
		if err := a.scrapeSvc.SyncTVEpisode(c.Request.Context(), videoID, req.TMDBID, req.SeasonNumber, req.EpisodeNumber); err != nil {
			response.Error(c, 6, err.Error())
			return
		}
	} else {
		if err := a.scrapeSvc.SyncMovieMetadata(c.Request.Context(), videoID, req.TMDBID); err != nil {
			response.Error(c, 6, err.Error())
			return
		}
	}

	ok(c, gin.H{"video_id": videoID, "tmdb_id": req.TMDBID, "media_type": mediaType, "synced": true})
}
