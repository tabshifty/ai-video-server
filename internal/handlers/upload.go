package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"video-server/internal/middleware"
	"video-server/internal/response"
)

// Upload handles media upload and transcode task enqueue.
func (a *API) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		bad(c, "file is required")
		return
	}
	typ := c.DefaultPostForm("type", "short")
	title := c.PostForm("title")
	desc := c.PostForm("description")
	tags := parseTags(c.PostForm("tags"))
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}

	video, err := a.uploadSvc.SaveUpload(c.Request.Context(), userID, file, title, desc, typ, tags)
	if err != nil {
		a.logger.Error("upload failed", "error", err)
		response.Error(c, 2, err.Error())
		return
	}

	if err := a.enqueuer.EnqueueTranscode(video.ID); err != nil {
		a.logger.Error("enqueue transcode failed", "video_id", video.ID, "error", err)
		response.Error(c, 3, err.Error())
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"code": 0,
		"msg":  "",
		"data": gin.H{
			"video_id": video.ID,
			"status":   "uploaded",
		},
	})
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
