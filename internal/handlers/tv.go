package handlers

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"video-server/internal/middleware"
	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/response"
)

func (a *API) TVHome(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	payload, err := a.appSvc.TVHome(c.Request.Context(), userID, c.Query("q"), page, pageSize)
	if err != nil {
		response.Error(c, 2101, err.Error())
		return
	}
	ok(c, payload)
}

func (a *API) TVSearch(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	_ = userID
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	payload, err := a.appSvc.TVSearch(c.Request.Context(), c.Query("q"), page, pageSize)
	if err != nil {
		response.Error(c, 2103, err.Error())
		return
	}
	ok(c, payload)
}

func (a *API) TVCatalogWall(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	_ = userID
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 24)
	payload, err := a.appSvc.TVCatalogWall(c.Request.Context(), c.Query("kind"), page, pageSize)
	if err != nil {
		response.Error(c, 2104, err.Error())
		return
	}
	ok(c, payload)
}

func (a *API) TVSeriesDetail(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	seriesID, okSeries := parseInt64(c.Param("id"))
	if !okSeries {
		bad(c, "invalid series id")
		return
	}
	payload, err := a.appSvc.TVSeriesDetail(c.Request.Context(), userID, seriesID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || repository.IsNotFound(err) {
			response.Error(c, 404, "series not found")
			return
		}
		response.Error(c, 2102, err.Error())
		return
	}
	ok(c, payload)
}

func (a *API) AdminTVSeries(c *gin.Context) {
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	var active *bool
	switch strings.TrimSpace(c.Query("active")) {
	case "1", "true":
		v := true
		active = &v
	case "0", "false":
		v := false
		active = &v
	}
	var hasPlayable *bool
	switch strings.TrimSpace(c.Query("has_playable")) {
	case "1", "true":
		v := true
		hasPlayable = &v
	case "0", "false":
		v := false
		hasPlayable = &v
	}
	items, total, err := a.repo.ListAdminTVSeries(c.Request.Context(), repository.AdminTVSeriesFilter{
		Page:        page,
		PageSize:    pageSize,
		Query:       c.Query("q"),
		Active:      active,
		HasPlayable: hasPlayable,
	})
	if err != nil {
		response.Error(c, 2110, err.Error())
		return
	}
	ok(c, gin.H{
		"items":       items,
		"total_count": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

func (a *API) AdminTVSeriesDetail(c *gin.Context) {
	seriesID, okSeries := parseInt64(c.Param("id"))
	if !okSeries {
		bad(c, "invalid series id")
		return
	}
	payload, err := a.repo.GetAdminTVSeriesDetail(c.Request.Context(), seriesID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || repository.IsNotFound(err) {
			response.Error(c, 404, "series not found")
			return
		}
		response.Error(c, 2111, err.Error())
		return
	}
	ok(c, payload)
}

func (a *API) AdminCreateTVSeries(c *gin.Context) {
	var req models.AdminTvSeriesInput
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		bad(c, "title is required")
		return
	}
	payload, err := a.repo.CreateAdminTVSeries(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 2112, err.Error())
		return
	}
	ok(c, payload)
}

func (a *API) AdminUpdateTVSeries(c *gin.Context) {
	seriesID, okSeries := parseInt64(c.Param("id"))
	if !okSeries {
		bad(c, "invalid series id")
		return
	}
	var req models.AdminTvSeriesInput
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		bad(c, "title is required")
		return
	}
	payload, err := a.repo.UpdateAdminTVSeries(c.Request.Context(), seriesID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || repository.IsNotFound(err) {
			response.Error(c, 404, "series not found")
			return
		}
		response.Error(c, 2113, err.Error())
		return
	}
	ok(c, payload)
}

func (a *API) AdminDeleteTVSeries(c *gin.Context) {
	seriesID, okSeries := parseInt64(c.Param("id"))
	if !okSeries {
		bad(c, "invalid series id")
		return
	}
	if err := a.repo.DeleteAdminTVSeries(c.Request.Context(), seriesID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) || repository.IsNotFound(err) {
			response.Error(c, 404, "series not found")
			return
		}
		response.Error(c, 2114, err.Error())
		return
	}
	ok(c, gin.H{"deleted": true})
}

func (a *API) AdminCreateTVSeason(c *gin.Context) {
	seriesID, okSeries := parseInt64(c.Param("id"))
	if !okSeries {
		bad(c, "invalid series id")
		return
	}
	var req models.AdminTvSeasonInput
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	if req.SeasonNumber <= 0 {
		bad(c, "season_number must be positive")
		return
	}
	payload, err := a.repo.CreateAdminTVSeason(c.Request.Context(), seriesID, req)
	if err != nil {
		response.Error(c, 2115, err.Error())
		return
	}
	ok(c, payload)
}

func (a *API) AdminUpdateTVSeason(c *gin.Context) {
	seasonID, okSeason := parseInt64(c.Param("id"))
	if !okSeason {
		bad(c, "invalid season id")
		return
	}
	var req models.AdminTvSeasonInput
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	if req.SeasonNumber <= 0 {
		bad(c, "season_number must be positive")
		return
	}
	payload, err := a.repo.UpdateAdminTVSeason(c.Request.Context(), seasonID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || repository.IsNotFound(err) {
			response.Error(c, 404, "season not found")
			return
		}
		response.Error(c, 2116, err.Error())
		return
	}
	ok(c, payload)
}

func (a *API) AdminDeleteTVSeason(c *gin.Context) {
	seasonID, okSeason := parseInt64(c.Param("id"))
	if !okSeason {
		bad(c, "invalid season id")
		return
	}
	if err := a.repo.DeleteAdminTVSeason(c.Request.Context(), seasonID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) || repository.IsNotFound(err) {
			response.Error(c, 404, "season not found")
			return
		}
		response.Error(c, 2117, err.Error())
		return
	}
	ok(c, gin.H{"deleted": true})
}

func (a *API) AdminCreateTVEpisode(c *gin.Context) {
	seasonID, okSeason := parseInt64(c.Param("id"))
	if !okSeason {
		bad(c, "invalid season id")
		return
	}
	var req models.AdminTvEpisodeInput
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	if req.EpisodeNumber <= 0 {
		bad(c, "episode_number must be positive")
		return
	}
	payload, err := a.repo.CreateAdminTVEpisode(c.Request.Context(), seasonID, req)
	if err != nil {
		response.Error(c, 2118, err.Error())
		return
	}
	ok(c, payload)
}

func (a *API) AdminUpdateTVEpisode(c *gin.Context) {
	episodeID, okEpisode := parseInt64(c.Param("id"))
	if !okEpisode {
		bad(c, "invalid episode id")
		return
	}
	var req models.AdminTvEpisodeInput
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	if req.EpisodeNumber <= 0 {
		bad(c, "episode_number must be positive")
		return
	}
	payload, err := a.repo.UpdateAdminTVEpisode(c.Request.Context(), episodeID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || repository.IsNotFound(err) {
			response.Error(c, 404, "episode not found")
			return
		}
		response.Error(c, 2119, err.Error())
		return
	}
	ok(c, payload)
}

func (a *API) AdminDeleteTVEpisode(c *gin.Context) {
	episodeID, okEpisode := parseInt64(c.Param("id"))
	if !okEpisode {
		bad(c, "invalid episode id")
		return
	}
	if err := a.repo.DeleteAdminTVEpisode(c.Request.Context(), episodeID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) || repository.IsNotFound(err) {
			response.Error(c, 404, "episode not found")
			return
		}
		response.Error(c, 2120, err.Error())
		return
	}
	ok(c, gin.H{"deleted": true})
}
