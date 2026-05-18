package services

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
)

func TestBuildTVHomePayloadBrowseMode(t *testing.T) {
	t.Parallel()

	continueWatching := &models.TvContinueWatchingDto{
		SeriesID:        7,
		SeriesTitle:     "雾城档案",
		SeasonNumber:    2,
		EpisodeNumber:   4,
		EpisodeTitle:    "暗线浮现",
		ProgressPercent: 64,
	}
	series := []models.TvSeriesSummaryDto{
		{
			ID:                   7,
			Title:                "雾城档案",
			Overview:             "调查悬案。",
			FirstAirDate:         "2025-04-01",
			TotalSeasons:         2,
			TotalEpisodes:        16,
			PlayableEpisodes:     12,
			LatestEpisodeAirDate: "2025-04-23",
		},
	}

	payload := buildTVHomePayload("", continueWatching, series, 1, 20)

	if payload.ContinueWatching == nil {
		t.Fatal("expected continue watching in browse mode")
	}
	if len(payload.Sections) == 0 {
		t.Fatal("expected sections in browse mode")
	}
	if len(payload.SearchResults) != 0 {
		t.Fatalf("expected empty search results in browse mode, got %d", len(payload.SearchResults))
	}
}

func TestBuildTVHomePayloadFixedSectionsIncludeAV(t *testing.T) {
	t.Parallel()

	payload := models.TvHomePayload{
		TvSeries: []models.TvHomeVideoDto{{ID: "7", Type: "tv", Title: "雾城档案"}},
		Movies:   []models.TvHomeVideoDto{{ID: "m-1", Type: "movie", Title: "午夜列车"}},
		AV:       []models.TvHomeVideoDto{{ID: "av-1", Type: "av", Title: "SNIS-001"}},
	}

	if len(payload.TvSeries) != 1 || payload.TvSeries[0].Type != "tv" {
		t.Fatalf("expected tv_series fixed section, got=%#v", payload.TvSeries)
	}
	if len(payload.Movies) != 1 || payload.Movies[0].Type != "movie" {
		t.Fatalf("expected movies fixed section, got=%#v", payload.Movies)
	}
	if len(payload.AV) != 1 || payload.AV[0].Type != "av" {
		t.Fatalf("expected av fixed section, got=%#v", payload.AV)
	}
}

func TestBuildTypedTVHomePayloadForSeries(t *testing.T) {
	t.Parallel()

	continueWatching := &models.TvContinueWatchingDto{
		Type:            "tv",
		SeriesID:        7,
		SeriesTitle:     "雾城档案",
		SeasonNumber:    2,
		EpisodeNumber:   4,
		EpisodeTitle:    "暗线浮现",
		VideoID:         "episode-video-7",
		ProgressPercent: 64,
	}
	updates := []models.TvHomeVideoDto{
		{
			ID:            "7",
			Type:          "tv",
			Title:         "雾城档案",
			SeasonNumber:  2,
			EpisodeNumber: 4,
			VideoID:       "episode-video-7",
		},
	}

	payload := buildTypedTVHomePayload("tv", continueWatching, updates, 1, 20)

	if payload.Kind != "tv" {
		t.Fatalf("expected kind tv, got=%s", payload.Kind)
	}
	if payload.Featured == nil || payload.Featured.Title != "雾城档案" || payload.Featured.Type != "tv" {
		t.Fatalf("expected continue watching as featured tv item, got=%#v", payload.Featured)
	}
	if len(payload.RecentWatching) != 1 || payload.RecentWatching[0].SeasonNumber != 2 || payload.RecentWatching[0].EpisodeNumber != 4 {
		t.Fatalf("expected series/season/episode recent watching, got=%#v", payload.RecentWatching)
	}
	if len(payload.RecentUpdates) != 1 || payload.RecentUpdates[0].Title != "雾城档案" {
		t.Fatalf("expected recent updates, got=%#v", payload.RecentUpdates)
	}
}

func TestBuildTypedTVHomePayloadForMovieAndAV(t *testing.T) {
	t.Parallel()

	movieHistory := tvHistoryItem("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "午夜列车", "movie")
	movieUpdate := models.TvHomeVideoDto{ID: movieHistory.VideoID.String(), Type: "movie", Title: "午夜列车"}
	moviePayload := buildTypedTVHomePayload("movie", nil, []models.TvHomeVideoDto{movieUpdate}, 1, 20, movieHistory)

	if moviePayload.Kind != "movie" {
		t.Fatalf("expected movie kind, got=%s", moviePayload.Kind)
	}
	if moviePayload.Featured == nil || moviePayload.Featured.Type != "movie" || moviePayload.Featured.VideoID != movieHistory.VideoID.String() {
		t.Fatalf("expected movie history as featured, got=%#v", moviePayload.Featured)
	}
	if len(moviePayload.RecentWatching) != 1 || moviePayload.RecentWatching[0].Type != "movie" {
		t.Fatalf("expected movie recent watching, got=%#v", moviePayload.RecentWatching)
	}
	if len(moviePayload.RecentUpdates) != 1 || moviePayload.RecentUpdates[0].Title != "午夜列车" {
		t.Fatalf("expected movie recent updates, got=%#v", moviePayload.RecentUpdates)
	}

	avHistory := tvHistoryItem("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "SNIS-001", "av")
	avUpdate := models.TvHomeVideoDto{ID: avHistory.VideoID.String(), Type: "av", Title: "SNIS-001"}
	avPayload := buildTypedTVHomePayload("av", nil, []models.TvHomeVideoDto{avUpdate}, 1, 20, avHistory)

	if avPayload.Kind != "av" {
		t.Fatalf("expected av kind, got=%s", avPayload.Kind)
	}
	if avPayload.Featured == nil || avPayload.Featured.Type != "av" {
		t.Fatalf("expected av featured item, got=%#v", avPayload.Featured)
	}
	if len(avPayload.RecentWatching) != 1 || avPayload.RecentWatching[0].Type != "av" {
		t.Fatalf("expected av recent watching, got=%#v", avPayload.RecentWatching)
	}
	if len(avPayload.RecentUpdates) != 1 || avPayload.RecentUpdates[0].Type != "av" {
		t.Fatalf("expected av recent updates, got=%#v", avPayload.RecentUpdates)
	}
}

func TestBuildTVHomePayloadSearchMode(t *testing.T) {
	t.Parallel()

	payload := buildTVHomePayload("雾城", &models.TvContinueWatchingDto{SeriesID: 1}, []models.TvSeriesSummaryDto{
		{
			ID:               3,
			Title:            "雾城档案",
			PlayableEpisodes: 8,
		},
	}, 2, 10)

	if payload.ContinueWatching != nil {
		t.Fatal("expected search mode to hide continue watching")
	}
	if len(payload.Sections) != 0 {
		t.Fatalf("expected search mode to hide sections, got %d", len(payload.Sections))
	}
	if len(payload.SearchResults) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(payload.SearchResults))
	}
	if payload.Page != 2 || payload.PageSize != 10 {
		t.Fatalf("unexpected paging: page=%d pageSize=%d", payload.Page, payload.PageSize)
	}
}

func tvHistoryItem(id string, title string, typ string) models.HistoryItem {
	return models.HistoryItem{
		VideoID:       uuid.MustParse(id),
		Title:         title,
		Type:          typ,
		ThumbnailPath: "/" + typ + ".jpg",
		Duration:      1000,
		WatchSeconds:  250,
		Progress:      0.25,
		LastWatchedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	}
}

func TestBuildTVSectionsGroupsSeriesByRule(t *testing.T) {
	t.Parallel()

	series := []models.TvSeriesSummaryDto{
		{
			ID:                   1,
			Title:                "旧城往事",
			FirstAirDate:         "2020-02-01",
			PlayableEpisodes:     3,
			LatestEpisodeAirDate: "2020-03-01",
		},
		{
			ID:                   2,
			Title:                "今日上线",
			FirstAirDate:         "2025-03-01",
			PlayableEpisodes:     6,
			LatestEpisodeAirDate: "2025-04-22",
		},
		{
			ID:                   3,
			Title:                "本周热播",
			FirstAirDate:         "2025-01-10",
			PlayableEpisodes:     10,
			LatestEpisodeAirDate: "2025-04-23",
		},
	}

	sections := buildTVSections(series, time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC))

	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}
	if sections[0].Title != "最近更新" {
		t.Fatalf("unexpected first section title: %s", sections[0].Title)
	}
	if len(sections[0].Items) == 0 || sections[0].Items[0].Title != "本周热播" {
		t.Fatalf("expected newest series first in recent updates, got %#v", sections[0].Items)
	}
	if sections[1].Title != "高能连播" {
		t.Fatalf("unexpected second section title: %s", sections[1].Title)
	}
	if len(sections[1].Items) == 0 || sections[1].Items[0].Title != "本周热播" {
		t.Fatalf("expected most playable series first in binge section, got %#v", sections[1].Items)
	}
	if sections[2].Title != "经典补档" {
		t.Fatalf("unexpected third section title: %s", sections[2].Title)
	}
	if len(sections[2].Items) == 0 || sections[2].Items[0].Title != "旧城往事" {
		t.Fatalf("expected oldest series first in archive section, got %#v", sections[2].Items)
	}
}
