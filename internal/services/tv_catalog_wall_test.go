package services

import (
	"strings"
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

func TestBuildTVCatalogWallVideoItemsPreservesAVType(t *testing.T) {
	t.Parallel()

	items := buildTVCatalogWallVideoItems([]models.VideoListItem{
		{
			ID:            mustParseUUID(t, "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			Title:         "SNIS-001",
			Type:          "av",
			ThumbnailPath: "/av.jpg",
		},
	}, "av")

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Type != "av" {
		t.Fatalf("expected av type, got %s", items[0].Type)
	}
	if items[0].Title != "SNIS-001" || items[0].PosterURL != "/av.jpg" {
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

func TestNormalizeTVCatalogWallSort(t *testing.T) {
	t.Parallel()

	sort := normalizeTVCatalogWallSort("release", "asc")
	if sort.By != "release" || sort.Order != "asc" {
		t.Fatalf("expected release asc, got %#v", sort)
	}

	sort = normalizeTVCatalogWallSort("", "")
	if sort.By != "added" || sort.Order != "desc" {
		t.Fatalf("expected default added desc, got %#v", sort)
	}

	sort = normalizeTVCatalogWallSort("unknown", "sideways")
	if sort.By != "added" || sort.Order != "desc" {
		t.Fatalf("expected invalid sort fallback, got %#v", sort)
	}
}

func TestTVCatalogWallSortOrderClause(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		sort tvCatalogWallSort
		want string
	}{
		{
			name: "series added desc",
			sort: tvCatalogWallSort{By: "added", Order: "desc"},
			want: "MAX(v.created_at) DESC NULLS LAST, s.title ASC",
		},
		{
			name: "series release asc",
			sort: tvCatalogWallSort{By: "release", Order: "asc"},
			want: "s.first_air_date ASC NULLS LAST, s.title ASC",
		},
		{
			name: "video added asc",
			sort: tvCatalogWallSort{By: "added", Order: "asc"},
			want: "v.created_at ASC",
		},
		{
			name: "video release desc",
			sort: tvCatalogWallSort{By: "release", Order: "desc"},
			want: "NULLIF(v.metadata->>'release_date', '')::date DESC NULLS LAST, v.created_at DESC",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tvCatalogWallSeriesOrderClause(tc.sort); strings.HasPrefix(tc.name, "series") && got != tc.want {
				t.Fatalf("unexpected series clause: %s", got)
			}
			if got := tvCatalogWallVideoOrderClause(tc.sort); strings.HasPrefix(tc.name, "video") && got != tc.want {
				t.Fatalf("unexpected video clause: %s", got)
			}
		})
	}
}
