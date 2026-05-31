package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestVideoPlayURL(t *testing.T) {
	videoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	got := VideoPlayURL(videoID)
	want := "/api/v1/videos/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee/source"
	if got != want {
		t.Fatalf("unexpected play url: got=%s want=%s", got, want)
	}
}

func TestVideoPlayURLWithProfile(t *testing.T) {
	videoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	got := VideoPlayURLWithProfile(videoID, "compat")
	want := "/api/v1/videos/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee/source?profile=compat"
	if got != want {
		t.Fatalf("unexpected profiled play url: got=%s want=%s", got, want)
	}
}

func TestVideoThumbnailURL(t *testing.T) {
	videoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	got := VideoThumbnailURL(videoID)
	want := "/api/v1/videos/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee/thumbnail"
	if got != want {
		t.Fatalf("unexpected thumbnail url: got=%s want=%s", got, want)
	}
}

func TestAdminImageViewURL(t *testing.T) {
	imageID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	got := AdminImageViewURL(imageID, 320, 240, "cover", 82)
	want := "/api/v1/admin/images/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee/view?fit=cover&h=240&q=82&w=320"
	if got != want {
		t.Fatalf("unexpected admin image view url: got=%s want=%s", got, want)
	}
}

func TestAppImageViewURL(t *testing.T) {
	imageID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	got := AppImageViewURL(imageID, 320, 240, "cover", 82)
	want := "/api/v1/images/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee/view?fit=cover&h=240&q=82&w=320"
	if got != want {
		t.Fatalf("unexpected app image view url: got=%s want=%s", got, want)
	}
}

func TestTVSeriesPosterURL(t *testing.T) {
	got := TVSeriesPosterURL(42)
	want := "/api/v1/tv/series/42/poster"
	if got != want {
		t.Fatalf("unexpected tv poster url: got=%s want=%s", got, want)
	}
}

func TestTVSeriesBackdropURL(t *testing.T) {
	got := TVSeriesBackdropURL(42)
	want := "/api/v1/tv/series/42/backdrop"
	if got != want {
		t.Fatalf("unexpected tv backdrop url: got=%s want=%s", got, want)
	}
}

func TestTVEpisodeStillURL(t *testing.T) {
	got := TVEpisodeStillURL(42, 1, 2)
	want := "/api/v1/tv/series/42/seasons/1/episodes/2/still"
	if got != want {
		t.Fatalf("unexpected tv episode still url: got=%s want=%s", got, want)
	}
}
