package utils

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func VideoPlayURL(videoID uuid.UUID) string {
	return fmt.Sprintf("/api/v1/videos/%s/source", videoID.String())
}

func VideoPlayURLWithProfile(videoID uuid.UUID, profile string) string {
	base := VideoPlayURL(videoID)
	profile = strings.TrimSpace(profile)
	if profile == "" {
		return base
	}
	params := url.Values{}
	params.Set("profile", profile)
	return base + "?" + params.Encode()
}

func VideoThumbnailURL(videoID uuid.UUID) string {
	return fmt.Sprintf("/api/v1/videos/%s/thumbnail", videoID.String())
}

func TVSeriesPosterURL(seriesID int64) string {
	return fmt.Sprintf("/api/v1/tv/series/%d/poster", seriesID)
}

func TVSeriesBackdropURL(seriesID int64) string {
	return fmt.Sprintf("/api/v1/tv/series/%d/backdrop", seriesID)
}

func VideoSubtitleURL(videoID, subtitleID uuid.UUID) string {
	return fmt.Sprintf("/api/v1/videos/%s/subtitles/%s/file", videoID.String(), subtitleID.String())
}

func AdminImageViewURL(imageID uuid.UUID, width, height int, fit string, quality int) string {
	return imageViewURL("/api/v1/admin/images", imageID, width, height, fit, quality)
}

func AppImageViewURL(imageID uuid.UUID, width, height int, fit string, quality int) string {
	return imageViewURL("/api/v1/images", imageID, width, height, fit, quality)
}

func imageViewURL(prefix string, imageID uuid.UUID, width, height int, fit string, quality int) string {
	base := fmt.Sprintf("%s/%s/view", prefix, imageID.String())
	params := url.Values{}
	if fit != "" {
		params.Set("fit", fit)
	}
	if height > 0 {
		params.Set("h", strconv.Itoa(height))
	}
	if quality > 0 {
		params.Set("q", strconv.Itoa(quality))
	}
	if width > 0 {
		params.Set("w", strconv.Itoa(width))
	}
	if len(params) == 0 {
		return base
	}
	return base + "?" + params.Encode()
}
