package services

import (
	"context"
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

func (s *AppService) TVSeriesDetail(ctx context.Context, userID uuid.UUID, seriesID int64) (models.TvSeriesDetailDto, error) {
	return s.repo.GetTVSeriesDetail(ctx, userID, seriesID)
}
