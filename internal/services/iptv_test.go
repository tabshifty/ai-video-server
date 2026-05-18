package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"video-server/internal/models"
)

func TestParseM3UPlaylistKeepsOrderAndSkipsInvalidEntries(t *testing.T) {
	t.Parallel()

	raw := `#EXTM3U
#EXTINF:-1 tvg-id="cctv1" tvg-logo="https://img.example/c1.png" group-title="央视频道",CCTV-1 综合
https://live.example/cctv1.m3u8
#EXTINF:-1 group-title="央视频道",
https://live.example/no-name.m3u8
#EXTINF:-1 group-title="本地",本地频道
rtmp://live.example/local
#EXTINF:-1,新闻频道
http://live.example/news.m3u8
`

	channels, skipped := ParseM3UPlaylist(strings.NewReader(raw))

	if skipped != 2 {
		t.Fatalf("expected 2 skipped entries, got %d", skipped)
	}
	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}
	first := channels[0]
	if first.ID == "" {
		t.Fatal("expected generated channel id")
	}
	if first.Name != "CCTV-1 综合" || first.URL != "https://live.example/cctv1.m3u8" || first.Group != "央视频道" {
		t.Fatalf("unexpected first channel: %#v", first)
	}
	if first.LogoURL != "https://img.example/c1.png" || first.TVGID != "cctv1" || first.SortOrder != 0 {
		t.Fatalf("unexpected first channel attrs: %#v", first)
	}
	if channels[1].Name != "新闻频道" || channels[1].SortOrder != 1 {
		t.Fatalf("expected order to be preserved, got %#v", channels[1])
	}
}

func TestParseM3UPlaylistSkipsAudioOnlyEntries(t *testing.T) {
	t.Parallel()

	raw := `#EXTM3U
#EXTINF:-1 group-title="Audio",CCTV-1 音频
https://piccpndali.v.myalicdn.com/audio/cctv1_2.m3u8
#EXTINF:-1 group-title="央视频道",CCTV-1 综合
https://live.example/cctv1.m3u8
`

	channels, skipped := ParseM3UPlaylist(strings.NewReader(raw))

	if skipped != 1 {
		t.Fatalf("expected audio-only entry to be skipped, got skipped=%d channels=%#v", skipped, channels)
	}
	if len(channels) != 1 || channels[0].Name != "CCTV-1 综合" {
		t.Fatalf("expected only video channel to remain, got %#v", channels)
	}
}

func TestBuildIPTVPlaylistStatusGroupsChannels(t *testing.T) {
	t.Parallel()

	updatedAt := time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC)
	status := BuildIPTVPlaylistStatus(models.IPTVPlaylistMeta{
		SourceURL:    "https://example.com/list.m3u",
		UpdatedAt:    &updatedAt,
		SkippedCount: 1,
	}, []models.IPTVChannel{
		{Name: "B", Group: "体育", SortOrder: 1},
		{Name: "A", Group: "新闻", SortOrder: 0},
		{Name: "C", Group: "体育", SortOrder: 2},
		{Name: "D", Group: "", SortOrder: 3},
	})

	if status.SourceURL != "https://example.com/list.m3u" || status.UpdatedAt == nil {
		t.Fatalf("unexpected playlist meta: %#v", status)
	}
	if status.ChannelCount != 4 || status.SkippedCount != 1 {
		t.Fatalf("unexpected counts: %#v", status)
	}
	wantGroups := []string{"体育", "新闻", "未分组"}
	if len(status.Groups) != len(wantGroups) {
		t.Fatalf("expected groups %#v, got %#v", wantGroups, status.Groups)
	}
	for i := range wantGroups {
		if status.Groups[i] != wantGroups[i] {
			t.Fatalf("expected groups %#v, got %#v", wantGroups, status.Groups)
		}
	}
}

func TestIPTVServiceRefreshFailsWithoutSourceURL(t *testing.T) {
	t.Parallel()

	svc := NewIPTVService(&fakeIPTVRepository{}, nil)
	_, err := svc.Refresh(context.Background())

	if !errors.Is(err, ErrIPTVSourceURLRequired) {
		t.Fatalf("expected ErrIPTVSourceURLRequired, got %v", err)
	}
}

type fakeIPTVRepository struct {
	meta     models.IPTVPlaylistMeta
	channels []models.IPTVChannel
}

func (r *fakeIPTVRepository) GetIPTVPlaylist(context.Context) (models.IPTVPlaylistMeta, []models.IPTVChannel, error) {
	return r.meta, r.channels, nil
}

func (r *fakeIPTVRepository) SaveIPTVSourceURL(_ context.Context, sourceURL string) (models.IPTVPlaylistMeta, []models.IPTVChannel, error) {
	r.meta.SourceURL = sourceURL
	return r.meta, r.channels, nil
}

func (r *fakeIPTVRepository) ReplaceIPTVPlaylist(_ context.Context, meta models.IPTVPlaylistMeta, channels []models.IPTVChannel) (models.IPTVPlaylistMeta, []models.IPTVChannel, error) {
	r.meta = meta
	r.channels = channels
	return r.meta, r.channels, nil
}
