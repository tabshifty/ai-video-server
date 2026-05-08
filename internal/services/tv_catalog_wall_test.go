package services

import (
	"testing"

	"github.com/google/uuid"

	"video-server/internal/models"
)

func TestBuildTVCatalogWallSeriesItems(t *testing.T) {
	t.Parallel()

	items := buildTVCatalogWallSeriesItems([]models.TvSeriesSummaryDto{
		{
			ID:               7,
			Title:            "雾城档案",
			Overview:         "调查悬案。",
			PosterURL:        "/poster.jpg",
			BackdropURL:      "/backdrop.jpg",
			PlayableEpisodes: 8,
		},
	})

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Type != "tv" {
		t.Fatalf("expected tv type, got %s", items[0].Type)
	}
	if items[0].Title != "雾城档案" || items[0].PosterURL != "/poster.jpg" || items[0].BackdropURL != "/backdrop.jpg" {
		t.Fatalf("unexpected mapped item: %#v", items[0])
	}
}

func mustParseUUID(t *testing.T, raw string) uuid.UUID {
	t.Helper()
	value, err := uuid.Parse(raw)
	if err != nil {
		t.Fatalf("parse uuid: %v", err)
	}
	return value
}

func TestBuildTVCatalogWallVideoItems(t *testing.T) {
	t.Parallel()

	items := buildTVCatalogWallVideoItems([]models.VideoListItem{
		{
			ID:            mustParseUUID(t, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			Title:         "午夜列车",
			Type:          "movie",
			ThumbnailPath: "/movie.jpg",
		},
	}, "movie")

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Type != "movie" {
		t.Fatalf("expected movie type, got %s", items[0].Type)
	}
	if items[0].Title != "午夜列车" || items[0].PosterURL != "/movie.jpg" {
		t.Fatalf("unexpected mapped item: %#v", items[0])
	}
}

func TestBuildTVCatalogWallPayloadPreservesPaging(t *testing.T) {
	t.Parallel()

	payload := buildTVCatalogWallPayload(2, 24, 37, []models.TvCatalogWallItemDto{
		{ID: "item-1", Type: "tv", Title: "雾城档案"},
	})

	if payload.Page != 2 || payload.PageSize != 24 || payload.TotalCount != 37 {
		t.Fatalf("unexpected paging: %#v", payload)
	}
	if len(payload.Items) != 1 || payload.Items[0].ID != "item-1" {
		t.Fatalf("unexpected payload items: %#v", payload.Items)
	}
}
