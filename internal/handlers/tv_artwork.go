package handlers

import (
	"net/http"
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

func (a *API) TVEpisodeStill(c *gin.Context) {
	seriesID, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || seriesID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid tv series id"})
		return
	}
	seasonNumber, err := strconv.Atoi(strings.TrimSpace(c.Param("season")))
	if err != nil || seasonNumber <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid tv season number"})
		return
	}
	episodeNumber, err := strconv.Atoi(strings.TrimSpace(c.Param("episode")))
	if err != nil || episodeNumber <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid tv episode number"})
		return
	}

	rawPath, err := a.repo.GetTVEpisodeStillPath(c.Request.Context(), seriesID, seasonNumber, episodeNumber)
	if err != nil {
		if repository.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"msg": "tv episode not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "query tv episode still failed"})
		return
	}
	if strings.TrimSpace(rawPath) == "" {
		c.JSON(http.StatusNotFound, gin.H{"msg": "tv episode still not found"})
		return
	}

	localPath := tvEpisodeStillLocalPath(a.storageRoot, seriesID, seasonNumber, episodeNumber)
	a.serveTVArtworkPath(c, rawPath, localPath, "tv episode still not found")
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
	a.serveTVArtworkPath(c, rawPath, localPath, "tv artwork not found")
}

func (a *API) serveTVArtworkPath(c *gin.Context, rawPath, localPath, notFoundMessage string) {
	rawPath = strings.TrimSpace(rawPath)
	localPath = strings.TrimSpace(localPath)
	if localPath != "" {
		served, err := tryServeLocalArtworkPath(c, localPath)
		if served {
			return
		}
		if err != nil && !isLocalImageNotFound(err) && !canRedirectTVArtwork(rawPath) && rawPath == localPath {
			writeLocalImageOpenError(c, err, notFoundMessage, "tv artwork temporarily unavailable")
			return
		}
	}
	if strings.HasPrefix(rawPath, "http://") || strings.HasPrefix(rawPath, "https://") {
		c.Redirect(http.StatusTemporaryRedirect, rawPath)
		return
	}
	if strings.HasPrefix(rawPath, "/") && !looksLikeLocalFilesystemPath(rawPath) {
		c.Redirect(http.StatusTemporaryRedirect, "https://image.tmdb.org/t/p/original"+rawPath)
		return
	}
	if rawPath != "" && rawPath != localPath {
		served, err := tryServeLocalArtworkPath(c, rawPath)
		if served {
			return
		}
		if err != nil && !isLocalImageNotFound(err) {
			writeLocalImageOpenError(c, err, notFoundMessage, "tv artwork temporarily unavailable")
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"msg": notFoundMessage})
}

func tryServeLocalArtworkPath(c *gin.Context, path string) (bool, error) {
	file, info, err := openLocalImageFile(c.Request.Context(), path)
	if err != nil {
		return false, err
	}
	defer file.Close()
	serveOpenedLocalImage(c, path, file, info)
	return true, nil
}

func canRedirectTVArtwork(path string) bool {
	path = strings.TrimSpace(path)
	return strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") ||
		(strings.HasPrefix(path, "/") && !looksLikeLocalFilesystemPath(path))
}

func looksLikeLocalFilesystemPath(path string) bool {
	normalized := strings.ReplaceAll(strings.TrimSpace(path), "\\", "/")
	return strings.HasPrefix(normalized, "/Volumes/") ||
		strings.HasPrefix(normalized, "/Users/") ||
		strings.HasPrefix(normalized, "/private/") ||
		strings.HasPrefix(normalized, "/tmp/") ||
		strings.HasPrefix(normalized, "/var/")
}

func tvEpisodeStillLocalPath(storageRoot string, seriesID int64, seasonNumber, episodeNumber int) string {
	return filepath.Join(
		storageRoot,
		"tv",
		"series",
		strconv.FormatInt(seriesID, 10),
		"episodes",
		"s"+twoDigit(seasonNumber)+"e"+twoDigit(episodeNumber)+".jpg",
	)
}

func twoDigit(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}
