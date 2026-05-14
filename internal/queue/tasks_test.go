package queue

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"video-server/internal/models"
	"video-server/internal/services"
)

func TestBuildTranscodeTaskOptionsIncludesExplicitTimeout(t *testing.T) {
	t.Parallel()

	opts := buildTranscodeTaskOptions("transcode", 6*time.Hour)

	assertOption(t, opts, asynq.QueueOpt, "transcode")
	assertOption(t, opts, asynq.ProcessInOpt, 2*time.Second)
	assertOption(t, opts, asynq.TimeoutOpt, 6*time.Hour)
	assertOption(t, opts, asynq.MaxRetryOpt, 3)
}

func assertOption(t *testing.T, opts []asynq.Option, typ asynq.OptionType, want any) {
	t.Helper()
	for _, opt := range opts {
		if opt.Type() != typ {
			continue
		}
		if got := opt.Value(); got != want {
			t.Fatalf("option %v value=%v want=%v", typ, got, want)
		}
		return
	}
	t.Fatalf("expected option type %v in %v", typ, opts)
}

func TestResolveTranscodePersistencePreservesAVScrapedPoster(t *testing.T) {
	t.Parallel()

	existingMetadata := map[string]any{
		"poster_original_file_path": "/storage/posters/av-original.jpg",
		"poster_cropped_file_path":  "/storage/posters/av-cropped.jpg",
		"poster_variant":            "cropped",
		"poster_url":                "https://img.example/av.jpg",
		"thumb_url":                 "https://img.example/av-thumb.jpg",
		"codec":                     "old-codec",
	}
	video := models.Video{
		ID:            uuid.New(),
		Type:          "av",
		ThumbnailPath: "/storage/posters/av-cropped.jpg",
		Metadata:      mustMarshalMetadata(t, existingMetadata),
	}
	result := services.TranscodeResult{
		ThumbnailPath: "/storage/videos/thumb.jpg",
		Metadata: map[string]any{
			"codec":           "h264",
			"playback_codec":  "avc",
			"poster_variant":  "transcode-should-not-replace",
			"transcode_mode":  "bitrate",
			"playback_source": "generated",
		},
	}

	thumbPath, metadata := resolveTranscodePersistence(video, result)

	if thumbPath != video.ThumbnailPath {
		t.Fatalf("expected AV thumbnail to keep scraped poster, got=%s want=%s", thumbPath, video.ThumbnailPath)
	}
	if metadata["poster_original_file_path"] != existingMetadata["poster_original_file_path"] {
		t.Fatalf("expected poster_original_file_path to be preserved, got=%v", metadata["poster_original_file_path"])
	}
	if metadata["poster_cropped_file_path"] != existingMetadata["poster_cropped_file_path"] {
		t.Fatalf("expected poster_cropped_file_path to be preserved, got=%v", metadata["poster_cropped_file_path"])
	}
	if metadata["poster_variant"] != existingMetadata["poster_variant"] {
		t.Fatalf("expected poster_variant to be preserved, got=%v", metadata["poster_variant"])
	}
	if metadata["thumb_url"] != existingMetadata["thumb_url"] {
		t.Fatalf("expected thumb_url to be preserved, got=%v", metadata["thumb_url"])
	}
	if metadata["codec"] != result.Metadata["codec"] {
		t.Fatalf("expected playback codec metadata to be refreshed, got=%v", metadata["codec"])
	}
	if metadata["transcode_mode"] != result.Metadata["transcode_mode"] {
		t.Fatalf("expected transcode metadata to be added, got=%v", metadata["transcode_mode"])
	}
}

func TestResolveTranscodePersistenceKeepsAVDiagnosticsWhenNoScrapedPoster(t *testing.T) {
	t.Parallel()

	existingMetadata := map[string]any{
		"scrape_error":  "未找到封面",
		"scrape_source": "javdb",
	}
	video := models.Video{
		ID:            uuid.New(),
		Type:          "av",
		ThumbnailPath: "/storage/posters/old.jpg",
		Metadata:      mustMarshalMetadata(t, existingMetadata),
	}
	result := services.TranscodeResult{
		ThumbnailPath: "/storage/videos/thumb.jpg",
		Metadata: map[string]any{
			"codec":          "h264",
			"playback_codec": "avc",
		},
	}

	thumbPath, metadata := resolveTranscodePersistence(video, result)

	if thumbPath != result.ThumbnailPath {
		t.Fatalf("expected AV without poster metadata to use transcode thumbnail, got=%s want=%s", thumbPath, result.ThumbnailPath)
	}
	if metadata["scrape_error"] != existingMetadata["scrape_error"] {
		t.Fatalf("expected scrape_error to be preserved, got=%v", metadata["scrape_error"])
	}
	if metadata["codec"] != result.Metadata["codec"] {
		t.Fatalf("expected transcode metadata, got=%v", metadata["codec"])
	}
}

func TestResolveTranscodePersistenceRestoresAVPosterThumbnailWhenCurrentPathEmpty(t *testing.T) {
	t.Parallel()

	video := models.Video{
		ID:   uuid.New(),
		Type: "av",
		Metadata: mustMarshalMetadata(t, map[string]any{
			"poster_variant":            "cropped",
			"poster_original_file_path": "/storage/posters/av-original.jpg",
			"poster_cropped_file_path":  "/storage/posters/av-cropped.jpg",
		}),
	}
	result := services.TranscodeResult{
		ThumbnailPath: "/storage/videos/thumb.jpg",
		Metadata: map[string]any{
			"codec": "h264",
		},
	}

	thumbPath, _ := resolveTranscodePersistence(video, result)

	if thumbPath != "/storage/posters/av-cropped.jpg" {
		t.Fatalf("expected AV thumbnail to restore cropped poster path, got=%s", thumbPath)
	}
}

func TestResolveTranscodePersistenceDoesNotMergeNonAVMetadata(t *testing.T) {
	t.Parallel()

	for _, typ := range []string{"movie", "episode", "short"} {
		typ := typ
		t.Run(typ, func(t *testing.T) {
			t.Parallel()

			video := models.Video{
				ID:            uuid.New(),
				Type:          typ,
				ThumbnailPath: "/storage/posters/old.jpg",
				Metadata: mustMarshalMetadata(t, map[string]any{
					"poster_path":  "/tmdb/poster.jpg",
					"scrape_error": "旧诊断",
				}),
			}
			result := services.TranscodeResult{
				ThumbnailPath: "/storage/videos/thumb.jpg",
				Metadata: map[string]any{
					"codec": "h264",
				},
			}

			thumbPath, metadata := resolveTranscodePersistence(video, result)

			if thumbPath != result.ThumbnailPath {
				t.Fatalf("expected non-AV thumbnail to use transcode thumbnail, got=%s want=%s", thumbPath, result.ThumbnailPath)
			}
			if _, ok := metadata["poster_path"]; ok {
				t.Fatalf("expected non-AV metadata not to merge old poster_path, got=%v", metadata)
			}
			if _, ok := metadata["scrape_error"]; ok {
				t.Fatalf("expected non-AV metadata not to merge old scrape_error, got=%v", metadata)
			}
			if metadata["codec"] != result.Metadata["codec"] {
				t.Fatalf("expected transcode metadata, got=%v", metadata["codec"])
			}
		})
	}
}

func mustMarshalMetadata(t *testing.T, metadata map[string]any) []byte {
	t.Helper()
	raw, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}
	return raw
}
