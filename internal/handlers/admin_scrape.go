package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"video-server/internal/models"
	"video-server/internal/queue"
	"video-server/internal/response"
	"video-server/internal/services"
	"video-server/internal/utils"
)

// @Summary Admin preview scrape candidates
// @Tags admin-scrape
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body AdminScrapePreviewRequest true "preview payload"
// @Success 200 {object} APIResponse
// @Failure 200 {object} APIResponse
// @Router /admin/scrape/preview [post]
func (a *API) AdminScrapePreview(c *gin.Context) {
	var req struct {
		VideoID       string `json:"video_id"`
		Title         string `json:"title"`
		Year          int    `json:"year"`
		Type          string `json:"type"` // movie | tv | av
		SeasonNumber  int    `json:"season_number"`
		EpisodeNumber int    `json:"episode_number"`
		BypassCache   bool   `json:"bypass_cache"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	if req.Type != "" && req.Type != "movie" && req.Type != "tv" && req.Type != "av" {
		bad(c, "type must be movie, tv or av")
		return
	}

	if strings.TrimSpace(req.VideoID) != "" {
		videoID, okUUID := parseUUID(req.VideoID)
		if !okUUID {
			bad(c, "invalid video_id")
			return
		}
		video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
		if err != nil {
			response.Error(c, 2, err.Error())
			return
		}
		derivedType := "movie"
		if video.Type == "episode" {
			derivedType = "tv"
		} else if video.Type == "av" {
			derivedType = "av"
		}
		if req.Type == "" || req.Type != derivedType {
			req.Type = derivedType
		}
		if strings.TrimSpace(req.Title) == "" {
			req.Title = strings.TrimSpace(video.Title)
		}
		if strings.TrimSpace(req.Title) == "" && strings.TrimSpace(video.OriginalPath) != "" {
			name := filepath.Base(video.OriginalPath)
			if title, year, season, episode, parsed := utils.ParseFilename(name); parsed {
				req.Title = title
				if req.Year == 0 {
					req.Year = year
				}
				if req.SeasonNumber <= 0 {
					req.SeasonNumber = season
				}
				if req.EpisodeNumber <= 0 {
					req.EpisodeNumber = episode
				}
			}
		}

		req.Title, req.SeasonNumber, req.EpisodeNumber = normalizeTVPreviewRequest(req.Type, req.Title, req.SeasonNumber, req.EpisodeNumber)

		shouldUseExistingMetadata := req.Type != "av" && !(req.Type == "movie" && req.BypassCache)
		if shouldUseExistingMetadata && len(video.Metadata) > 0 {
			var raw map[string]any
			if err := json.Unmarshal(video.Metadata, &raw); err == nil {
				if video.TMDBID == nil {
					// 非 AV 类型确认刮削需要 tmdb_id，缓存中缺失时继续走在线预览
					goto previewSearch
				}
				ok(c, gin.H{
					"from_cache": true,
					"video_id":   video.ID,
					"candidates": []map[string]any{withTVPreviewContext(map[string]any{
						"tmdb_id":      *video.TMDBID,
						"title":        video.Title,
						"overview":     video.Description,
						"poster_path":  video.ThumbnailPath,
						"metadata":     raw,
						"release_date": raw["release_date"],
					}, req.Type, req.Title, req.SeasonNumber, req.EpisodeNumber)},
				})
				return
			}
		}
	}

previewSearch:
	req.Title, req.SeasonNumber, req.EpisodeNumber = normalizeTVPreviewRequest(req.Type, req.Title, req.SeasonNumber, req.EpisodeNumber)
	if strings.TrimSpace(req.Title) == "" {
		bad(c, "title is required when preview cache is unavailable")
		return
	}
	if req.Type == "" {
		req.Type = "movie"
	}

	var (
		candidates []map[string]any
		err        error
	)
	if req.Type == "tv" {
		candidates, err = a.scrapeSvc.PreviewTV(c.Request.Context(), req.Title, req.Year)
	} else if req.Type == "av" {
		candidates, err = a.scrapeSvc.PreviewAV(c.Request.Context(), req.Title)
	} else {
		candidates, err = a.scrapeSvc.PreviewMovieWithOptions(
			c.Request.Context(),
			req.Title,
			req.Year,
			services.MoviePreviewOptions{BypassCache: req.BypassCache},
		)
	}
	if err != nil {
		response.Error(c, 3, err.Error())
		return
	}
	for _, candidate := range candidates {
		withTVPreviewContext(candidate, req.Type, req.Title, req.SeasonNumber, req.EpisodeNumber)
	}
	ok(c, gin.H{"candidates": candidates})
}

func (a *API) AdminAVScrapeConfig(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet:
		cfg, err := a.repo.GetAVScraperConfig(c.Request.Context())
		if err != nil {
			response.Error(c, 2, err.Error())
			return
		}
		ok(c, cfg)
	case http.MethodPut:
		var req models.AVScraperSiteConfig
		if err := c.ShouldBindJSON(&req); err != nil {
			bad(c, "invalid payload")
			return
		}
		if err := a.repo.UpsertAVScraperConfig(c.Request.Context(), req); err != nil {
			response.Error(c, 3, err.Error())
			return
		}
		ok(c, gin.H{"saved": true})
	default:
		c.Status(http.StatusMethodNotAllowed)
	}
}

func (a *API) AdminAVScrapePreview(c *gin.Context) {
	var req struct {
		VideoID      string `json:"video_id"`
		Title        string `json:"title"`
		SiteCategory string `json:"site_category"`
		SiteSource   string `json:"site_source"`
		FilePath     string `json:"file_path"`
		DetailURL    string `json:"detail_url"`
		BypassCache  *bool  `json:"bypass_cache"`
	}
	osHash := ""
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}

	if strings.TrimSpace(req.VideoID) != "" {
		videoID, okUUID := parseUUID(req.VideoID)
		if !okUUID {
			bad(c, "invalid video_id")
			return
		}
		video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
		if err != nil {
			response.Error(c, 2, err.Error())
			return
		}
		osHash = video.OSHash
		if strings.TrimSpace(req.SiteCategory) == "" {
			req.SiteCategory = avSiteCategoryFromAdminVideo(video)
		}
		if strings.TrimSpace(req.Title) == "" {
			req.Title = strings.TrimSpace(video.Title)
		}
		if strings.TrimSpace(req.Title) == "" && strings.TrimSpace(video.OriginalPath) != "" {
			req.Title = strings.TrimSpace(strings.TrimSuffix(filepath.Base(video.OriginalPath), filepath.Ext(video.OriginalPath)))
		}
	}
	if strings.TrimSpace(req.Title) == "" {
		bad(c, "title is required")
		return
	}

	bypassCache := true
	if req.BypassCache != nil {
		bypassCache = *req.BypassCache
	}
	result, err := a.scrapeSvc.PreviewAVSearch(c.Request.Context(), req.Title, services.AVPreviewOptions{
		BypassCache:  bypassCache,
		SiteCategory: req.SiteCategory,
		SiteSource:   req.SiteSource,
		FilePath:     req.FilePath,
		DetailURL:    req.DetailURL,
		OSHash:       osHash,
	})
	if err != nil {
		response.Error(c, 3, err.Error())
		return
	}
	ok(c, gin.H{
		"from_cache":         false,
		"recommended_source": result.RecommendedSource,
		"used_source":        result.UsedSource,
		"site_category":      result.SiteCategory,
		"enabled_sources":    result.EnabledSources,
		"candidates":         result.Candidates,
	})
}

func (a *API) AdminAVScrapeConfirm(c *gin.Context) {
	var req struct {
		VideoID      string         `json:"video_id"`
		ExternalID   string         `json:"external_id"`
		Title        string         `json:"title"`
		Overview     string         `json:"overview"`
		PosterURL    string         `json:"poster_url"`
		ReleaseDate  string         `json:"release_date"`
		SiteCategory string         `json:"site_category"`
		SiteSource   string         `json:"site_source"`
		Metadata     map[string]any `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	videoID, okID := parseUUID(req.VideoID)
	if !okID {
		bad(c, "invalid video_id")
		return
	}
	video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
	if err != nil {
		response.Error(c, 2, err.Error())
		return
	}
	if video.Type != "av" {
		bad(c, "video type must be av")
		return
	}
	if strings.TrimSpace(req.ExternalID) == "" {
		bad(c, "external_id is required for av")
		return
	}
	if req.Metadata == nil {
		req.Metadata = map[string]any{}
	}
	if strings.TrimSpace(req.SiteSource) != "" {
		req.Metadata["scrape_source"] = req.SiteSource
	}
	if strings.TrimSpace(req.SiteCategory) != "" {
		req.Metadata["site_category"] = req.SiteCategory
	}
	if err := a.scrapeSvc.ConfirmAV(c.Request.Context(), services.ConfirmScrapeInput{
		VideoID:     videoID,
		Status:      video.Status,
		ExternalID:  req.ExternalID,
		Title:       req.Title,
		Overview:    req.Overview,
		PosterURL:   req.PosterURL,
		ReleaseDate: req.ReleaseDate,
		Metadata:    req.Metadata,
	}); err != nil {
		response.Error(c, 3, err.Error())
		return
	}
	// shouldEnqueueAdminScrapeConfirmTranscode intentionally reads the pre-confirm
	// status; ConfirmAV turns av_scrape_pending into processing before this point.
	if shouldEnqueueAdminScrapeConfirmTranscode(video) {
		if a.enqueuer == nil {
			response.Error(c, 5, "queue not configured")
			return
		}
		payload := queue.TranscodePayload{
			VideoID:      videoID.String(),
			InputPath:    video.OriginalPath,
			OutputDir:    filepath.Join(a.storageRoot, "videos"),
			TargetFormat: "mp4",
			Force:        true,
		}
		if err := a.enqueuer.EnqueueTranscode(payload); err != nil {
			response.Error(c, 5, err.Error())
			return
		}
	}
	ok(c, gin.H{
		"saved":    true,
		"video_id": videoID,
	})
}

// @Summary Admin confirm scrape metadata
// @Tags admin-scrape
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body AdminScrapeConfirmRequest true "confirm payload"
// @Success 200 {object} APIResponse
// @Failure 200 {object} APIResponse
// @Router /admin/scrape/confirm [put]
func (a *API) AdminScrapeConfirm(c *gin.Context) {
	var req struct {
		VideoID       string         `json:"video_id"`
		TMDBID        int            `json:"tmdb_id"`
		ExternalID    string         `json:"external_id"`
		Title         string         `json:"title"`
		Overview      string         `json:"overview"`
		PosterURL     string         `json:"poster_url"`
		BackdropURL   string         `json:"backdrop_url"`
		ReleaseDate   string         `json:"release_date"`
		Metadata      map[string]any `json:"metadata"`
		SeasonNumber  int            `json:"season_number"`
		EpisodeNumber int            `json:"episode_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	videoID, okID := parseUUID(req.VideoID)
	if !okID {
		bad(c, "invalid video_id")
		return
	}

	video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
	if err != nil {
		response.Error(c, 2, err.Error())
		return
	}
	if video.Type != "av" && req.TMDBID <= 0 {
		bad(c, "tmdb_id is required")
		return
	}
	if video.Type == "av" && strings.TrimSpace(req.ExternalID) == "" {
		bad(c, "external_id is required for av")
		return
	}

	input := services.ConfirmScrapeInput{
		VideoID:       videoID,
		TMDBID:        req.TMDBID,
		ExternalID:    req.ExternalID,
		Title:         req.Title,
		Overview:      req.Overview,
		PosterURL:     req.PosterURL,
		BackdropURL:   req.BackdropURL,
		ReleaseDate:   req.ReleaseDate,
		Metadata:      req.Metadata,
		SeasonNumber:  req.SeasonNumber,
		EpisodeNumber: req.EpisodeNumber,
	}

	switch video.Type {
	case "movie":
		if err := a.scrapeSvc.ConfirmMovie(c.Request.Context(), input); err != nil {
			response.Error(c, 3, err.Error())
			return
		}
	case "episode":
		if err := a.scrapeSvc.ConfirmEpisode(c.Request.Context(), input); err != nil {
			response.Error(c, 3, err.Error())
			return
		}
	case "av":
		input.Status = video.Status
		if err := a.scrapeSvc.ConfirmAV(c.Request.Context(), input); err != nil {
			response.Error(c, 3, err.Error())
			return
		}
	default:
		response.Error(c, 4, "only movie, episode or av supports scrape confirm")
		return
	}
	// shouldEnqueueAdminScrapeConfirmTranscode intentionally reads the pre-confirm
	// status; ConfirmAV turns av_scrape_pending into processing before this point.
	if shouldEnqueueAdminScrapeConfirmTranscode(video) {
		if a.enqueuer == nil {
			response.Error(c, 5, "queue not configured")
			return
		}
		payload := queue.TranscodePayload{
			VideoID:      videoID.String(),
			InputPath:    video.OriginalPath,
			OutputDir:    filepath.Join(a.storageRoot, "videos"),
			TargetFormat: "mp4",
			Force:        video.Type == "av" && video.Status == "av_scrape_pending",
		}
		if err := a.enqueuer.EnqueueTranscode(payload); err != nil {
			response.Error(c, 5, err.Error())
			return
		}
	}

	ok(c, gin.H{
		"saved":    true,
		"video_id": videoID,
	})
}

func shouldEnqueueAdminScrapeConfirmTranscode(video models.Video) bool {
	if video.Type == "av" {
		return video.Status == "av_scrape_pending" && strings.TrimSpace(video.OriginalPath) != ""
	}
	return strings.TrimSpace(video.OriginalPath) != "" && video.Status != "ready" && video.Status != "processing"
}

func (a *API) AdminScrapeSkip(c *gin.Context) {
	var req struct {
		VideoID string `json:"video_id"`
		Reason  string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	videoID, okID := parseUUID(req.VideoID)
	if !okID {
		bad(c, "invalid video_id")
		return
	}
	video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
	if err != nil {
		response.Error(c, 2, err.Error())
		return
	}
	if video.Type != "av" {
		bad(c, "video type must be av")
		return
	}
	if video.Status != "av_scrape_pending" {
		bad(c, "video status must be av_scrape_pending")
		return
	}
	if strings.TrimSpace(video.OriginalPath) == "" {
		response.Error(c, 3, "original path is empty")
		return
	}
	if a.enqueuer == nil {
		response.Error(c, 4, "queue not configured")
		return
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "管理员弃刮"
	}
	if err := a.repo.MergeVideoMetadata(c.Request.Context(), videoID, map[string]any{
		"scrape_skipped":     true,
		"scrape_skip_reason": reason,
		"scrape_skipped_at":  time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		response.Error(c, 5, err.Error())
		return
	}
	if err := a.repo.UpdateVideoStatus(c.Request.Context(), videoID, "processing"); err != nil {
		response.Error(c, 5, err.Error())
		return
	}
	payload := queue.TranscodePayload{
		VideoID:      videoID.String(),
		InputPath:    video.OriginalPath,
		OutputDir:    filepath.Join(a.storageRoot, "videos"),
		TargetFormat: "mp4",
		Force:        true,
	}
	if err := a.enqueuer.EnqueueTranscode(payload); err != nil {
		response.Error(c, 6, err.Error())
		return
	}
	ok(c, gin.H{
		"skipped":  true,
		"video_id": videoID,
	})
}

func avSiteCategoryFromAdminVideo(video models.Video) string {
	var metadata map[string]any
	if len(video.Metadata) > 0 {
		_ = json.Unmarshal(video.Metadata, &metadata)
	}
	raw, _ := metadata["site_category"].(string)
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "western":
		return "western"
	case "fc2":
		return "fc2"
	case "japanese", "default", "jp":
		return "japanese"
	default:
		if strings.TrimSpace(video.OSHash) != "" {
			return "western"
		}
		return ""
	}
}

func normalizeTVPreviewRequest(typ, title string, seasonNumber, episodeNumber int) (string, int, int) {
	if typ != "tv" {
		return strings.TrimSpace(title), seasonNumber, episodeNumber
	}

	parsedTitle, parsedSeason, parsedEpisode, ok := utils.ParseSeriesEpisode(title)
	if !ok {
		return strings.TrimSpace(title), seasonNumber, episodeNumber
	}
	if seasonNumber <= 0 {
		seasonNumber = parsedSeason
	}
	if episodeNumber <= 0 {
		episodeNumber = parsedEpisode
	}
	return strings.TrimSpace(parsedTitle), seasonNumber, episodeNumber
}

func withTVPreviewContext(candidate map[string]any, typ, parsedTitle string, seasonNumber, episodeNumber int) map[string]any {
	if typ != "tv" || candidate == nil {
		return candidate
	}
	if strings.TrimSpace(parsedTitle) != "" {
		candidate["parsed_title"] = strings.TrimSpace(parsedTitle)
	}
	if seasonNumber > 0 {
		candidate["parsed_season_number"] = seasonNumber
	}
	if episodeNumber > 0 {
		candidate["parsed_episode_number"] = episodeNumber
	}
	return candidate
}

func stringFromAny(v any) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}
