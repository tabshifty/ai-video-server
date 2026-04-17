package handlers

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

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
		VideoID string `json:"video_id"`
		Title   string `json:"title"`
		Year    int    `json:"year"`
		Type    string `json:"type"` // movie | tv | av
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
			if title, year, _, _, parsed := utils.ParseFilename(name); parsed {
				req.Title = title
				if req.Year == 0 {
					req.Year = year
				}
			}
		}

		if len(video.Metadata) > 0 {
			var raw map[string]any
			if err := json.Unmarshal(video.Metadata, &raw); err == nil {
				if req.Type == "av" {
					externalID := stringFromAny(raw["external_id"])
					if externalID == "" {
						if javdb, ok := raw["javdb"].(map[string]any); ok {
							externalID = stringFromAny(javdb["external_id"])
						}
					}
					ok(c, gin.H{
						"from_cache": true,
						"video_id":   video.ID,
						"candidates": []map[string]any{{
							"external_id":     externalID,
							"title":           video.Title,
							"overview":        video.Description,
							"poster_url":      video.ThumbnailPath,
							"release_date":    raw["release_date"],
							"metadata":        raw,
							"media_type_hint": "av",
							"scrape_source":   "javdb",
						}},
					})
					return
				}
				if video.TMDBID == nil {
					// 非 AV 类型确认刮削需要 tmdb_id，缓存中缺失时继续走在线预览
					goto previewSearch
				}
				ok(c, gin.H{
					"from_cache": true,
					"video_id":   video.ID,
					"candidates": []map[string]any{{
						"tmdb_id":      *video.TMDBID,
						"title":        video.Title,
						"overview":     video.Description,
						"poster_path":  video.ThumbnailPath,
						"metadata":     raw,
						"release_date": raw["release_date"],
					}},
				})
				return
			}
		}
	}

previewSearch:
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
		candidates, err = a.scrapeSvc.PreviewMovie(c.Request.Context(), req.Title, req.Year)
	}
	if err != nil {
		response.Error(c, 3, err.Error())
		return
	}
	ok(c, gin.H{"candidates": candidates})
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
		if err := a.scrapeSvc.ConfirmAV(c.Request.Context(), input); err != nil {
			response.Error(c, 3, err.Error())
			return
		}
	default:
		response.Error(c, 4, "only movie, episode or av supports scrape confirm")
		return
	}

	ok(c, gin.H{
		"saved":    true,
		"video_id": videoID,
	})
}

func stringFromAny(v any) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}
