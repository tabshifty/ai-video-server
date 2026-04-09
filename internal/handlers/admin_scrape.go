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

func (a *API) AdminScrapePreview(c *gin.Context) {
	var req struct {
		VideoID string `json:"video_id"`
		Title   string `json:"title"`
		Year    int    `json:"year"`
		Type    string `json:"type"` // movie | tv
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	if req.Type != "" && req.Type != "movie" && req.Type != "tv" {
		bad(c, "type must be movie or tv")
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
		if req.Type == "" {
			if video.Type == "episode" {
				req.Type = "tv"
			} else {
				req.Type = "movie"
			}
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

		if video.TMDBID != nil && len(video.Metadata) > 0 {
			var raw map[string]any
			if err := json.Unmarshal(video.Metadata, &raw); err == nil {
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
	} else {
		candidates, err = a.scrapeSvc.PreviewMovie(c.Request.Context(), req.Title, req.Year)
	}
	if err != nil {
		response.Error(c, 3, err.Error())
		return
	}
	ok(c, gin.H{"candidates": candidates})
}

func (a *API) AdminScrapeConfirm(c *gin.Context) {
	var req struct {
		VideoID       string         `json:"video_id"`
		TMDBID        int            `json:"tmdb_id"`
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
	if req.TMDBID <= 0 {
		bad(c, "tmdb_id is required")
		return
	}

	video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
	if err != nil {
		response.Error(c, 2, err.Error())
		return
	}

	input := services.ConfirmScrapeInput{
		VideoID:       videoID,
		TMDBID:        req.TMDBID,
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
	default:
		response.Error(c, 4, "only movie or episode supports scrape confirm")
		return
	}

	ok(c, gin.H{
		"saved":    true,
		"video_id": videoID,
	})
}
