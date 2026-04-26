package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"video-server/internal/repository"
)

func (a *API) TVSeriesPoster(c *gin.Context) {
	a.serveTVSeriesArtwork(c, "poster")
}

func (a *API) TVSeriesBackdrop(c *gin.Context) {
	a.serveTVSeriesArtwork(c, "backdrop")
}

func (a *API) serveTVSeriesArtwork(c *gin.Context, kind string) {
	seriesID, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || seriesID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid tv series id"})
		return
	}

	posterPath, backdropPath, err := a.repo.GetTVSeriesArtworkPaths(c.Request.Context(), seriesID)
	if err != nil {
		if repository.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"msg": "tv series not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "query tv series artwork failed"})
		return
	}

	rawPath := strings.TrimSpace(posterPath)
	if kind == "backdrop" {
		rawPath = strings.TrimSpace(backdropPath)
	}
	if rawPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"msg": "tv artwork not found"})
		return
	}

	localPath := filepath.Join(a.storageRoot, "tv", "series", strconv.FormatInt(seriesID, 10), kind+".jpg")
	if stat, statErr := os.Stat(localPath); statErr == nil && !stat.IsDir() {
		c.File(localPath)
		return
	}

	if stat, statErr := os.Stat(rawPath); statErr == nil && !stat.IsDir() {
		c.File(rawPath)
		return
	}

	if strings.HasPrefix(rawPath, "http://") || strings.HasPrefix(rawPath, "https://") {
		c.Redirect(http.StatusTemporaryRedirect, rawPath)
		return
	}
	if strings.HasPrefix(rawPath, "/") {
		c.Redirect(http.StatusTemporaryRedirect, "https://image.tmdb.org/t/p/original"+rawPath)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"msg": "tv artwork not found"})
}
