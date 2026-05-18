package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
)

func buildTVSections(series []models.TvSeriesSummaryDto, _ time.Time) []models.TvSectionDto {
	if len(series) == 0 {
		return nil
	}
	clone := append([]models.TvSeriesSummaryDto(nil), series...)

	latest := append([]models.TvSeriesSummaryDto(nil), clone...)
	sort.SliceStable(latest, func(i, j int) bool {
		if latest[i].LatestEpisodeAirDate == latest[j].LatestEpisodeAirDate {
			return latest[i].Title < latest[j].Title
		}
		return latest[i].LatestEpisodeAirDate > latest[j].LatestEpisodeAirDate
	})

	binge := repository.NormalizeTVSeriesSearchResults(clone)

	classic := append([]models.TvSeriesSummaryDto(nil), clone...)
	sort.SliceStable(classic, func(i, j int) bool {
		if classic[i].FirstAirDate == classic[j].FirstAirDate {
			return classic[i].Title < classic[j].Title
		}
		if classic[i].FirstAirDate == "" {
			return false
		}
		if classic[j].FirstAirDate == "" {
			return true
		}
		return classic[i].FirstAirDate < classic[j].FirstAirDate
	})

	sections := []models.TvSectionDto{
		{
			Title:    "最近更新",
			Subtitle: "按最近播出时间排序",
			Items:    takeSeries(latest, 12),
		},
		{
			Title:    "高能连播",
			Subtitle: "优先展示可直接播放的剧集",
			Items:    takeSeries(binge, 12),
		},
		{
			Title:    "经典补档",
			Subtitle: "从较早首播的系列开始补看",
			Items:    takeSeries(classic, 12),
		},
	}

	out := make([]models.TvSectionDto, 0, len(sections))
	for _, section := range sections {
		if len(section.Items) == 0 {
			continue
		}
		out = append(out, section)
	}
	return out
}

func takeSeries(items []models.TvSeriesSummaryDto, limit int) []models.TvSeriesSummaryDto {
	if len(items) == 0 || limit <= 0 {
		return nil
	}
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func buildTVHomePayload(
	query string,
	continueWatching *models.TvContinueWatchingDto,
	series []models.TvSeriesSummaryDto,
	page int,
	pageSize int,
) models.TvHomePayload {
	payload := models.TvHomePayload{
		Page:          page,
		PageSize:      pageSize,
		Sections:      []models.TvSectionDto{},
		SearchResults: []models.TvSeriesSummaryDto{},
	}
	if strings.TrimSpace(query) == "" {
		payload.ContinueWatching = continueWatching
		payload.Sections = buildTVSections(series, time.Now())
		return payload
	}
	payload.SearchResults = repository.NormalizeTVSeriesSearchResults(series)
	return payload
}

func buildTypedTVHomePayload(
	kind string,
	continueWatching *models.TvContinueWatchingDto,
	recentUpdates []models.TvHomeVideoDto,
	page int,
	pageSize int,
	historyItems ...models.HistoryItem,
) models.TvHomePayload {
	normalizedKind := normalizeTVHomeKind(kind)
	recentWatching := buildTVRecentWatchingItems(normalizedKind, continueWatching, historyItems)
	updates := filterTVHomeVideosByKind(normalizedKind, recentUpdates)
	payload := models.TvHomePayload{
		Kind:           normalizedKind,
		Featured:       firstTVHomeVideo(recentWatching, updates),
		RecentWatching: recentWatching,
		RecentUpdates:  updates,
		Page:           page,
		PageSize:       pageSize,
	}
	if normalizedKind == "tv" {
		payload.ContinueWatching = continueWatching
		payload.TvSeries = updates
		payload.Sections = []models.TvSectionDto{
			{
				Title:    "最近更新",
				Subtitle: "当前类型最新可播放内容",
				Items:    tvHomeVideosToSeriesSummaries(updates),
			},
		}
		return payload
	}
	if normalizedKind == "movie" {
		payload.Movies = updates
	} else if normalizedKind == "av" {
		payload.AV = updates
	}
	return payload
}

func normalizeTVHomeKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "movie":
		return "movie"
	case "av":
		return "av"
	default:
		return "tv"
	}
}

func buildTVRecentWatchingItems(
	kind string,
	continueWatching *models.TvContinueWatchingDto,
	historyItems []models.HistoryItem,
) []models.TvHomeVideoDto {
	if kind == "tv" {
		if continueWatching == nil {
			return nil
		}
		return []models.TvHomeVideoDto{{
			ID:              fmt.Sprintf("%d", continueWatching.SeriesID),
			Type:            "tv",
			Title:           continueWatching.SeriesTitle,
			Overview:        continueWatching.EpisodeTitle,
			PosterURL:       continueWatching.PosterURL,
			BackdropURL:     continueWatching.BackdropURL,
			VideoID:         continueWatching.VideoID,
			SeasonNumber:    continueWatching.SeasonNumber,
			EpisodeNumber:   continueWatching.EpisodeNumber,
			ProgressPercent: continueWatching.ProgressPercent,
		}}
	}
	out := make([]models.TvHomeVideoDto, 0, len(historyItems))
	for _, item := range historyItems {
		if normalizeTVSearchType(item.Type) != kind {
			continue
		}
		out = append(out, buildTVHomeVideoFromHistoryItem(item))
	}
	return out
}

func buildTVHomeVideoFromHistoryItem(item models.HistoryItem) models.TvHomeVideoDto {
	progressPercent := int(item.Progress * 100)
	if progressPercent == 0 && item.Duration > 0 {
		progressPercent = (item.WatchSeconds * 100) / item.Duration
	}
	if progressPercent > 100 {
		progressPercent = 100
	}
	typ := normalizeTVSearchType(item.Type)
	return models.TvHomeVideoDto{
		ID:              item.VideoID.String(),
		Type:            typ,
		Title:           item.Title,
		PosterURL:       item.ThumbnailPath,
		BackdropURL:     item.ThumbnailPath,
		VideoID:         item.VideoID.String(),
		ProgressPercent: progressPercent,
	}
}

func filterTVHomeVideosByKind(kind string, items []models.TvHomeVideoDto) []models.TvHomeVideoDto {
	out := make([]models.TvHomeVideoDto, 0, len(items))
	for _, item := range items {
		item.Type = normalizeTVSearchType(item.Type)
		if item.Type == "" {
			item.Type = kind
		}
		if item.Type != kind {
			continue
		}
		out = append(out, item)
	}
	return out
}

func firstTVHomeVideo(groups ...[]models.TvHomeVideoDto) *models.TvHomeVideoDto {
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}
		item := group[0]
		return &item
	}
	return nil
}

func tvHomeVideosToSeriesSummaries(items []models.TvHomeVideoDto) []models.TvSeriesSummaryDto {
	out := make([]models.TvSeriesSummaryDto, 0, len(items))
	for _, item := range items {
		id, ok := parseTVSeriesID(item.ID)
		if !ok {
			continue
		}
		out = append(out, models.TvSeriesSummaryDto{
			ID:          id,
			Title:       item.Title,
			Overview:    item.Overview,
			PosterURL:   item.PosterURL,
			BackdropURL: item.BackdropURL,
		})
	}
	return out
}

func parseTVSeriesID(raw string) (int64, bool) {
	var id int64
	if _, err := fmt.Sscanf(strings.TrimSpace(raw), "%d", &id); err != nil {
		return 0, false
	}
	return id, id > 0
}

func (s *AppService) TVSeriesDetail(ctx context.Context, userID uuid.UUID, seriesID int64) (models.TvSeriesDetailDto, error) {
	return s.repo.GetTVSeriesDetail(ctx, userID, seriesID)
}
