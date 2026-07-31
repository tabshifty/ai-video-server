package queue

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"video-server/internal/models"
	"video-server/internal/services"
	ffmpegpkg "video-server/pkg/ffmpeg"
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

func TestBuildDolbyVisionDirectCopyResultUsesSourceAsPlayableOutput(t *testing.T) {
	t.Parallel()

	sourcePath := "/storage/videos/video-1/source-dv.mkv"
	sourceProbe := ffmpegpkg.PlaybackCompatibilityProbe{
		VideoStreamFound:   true,
		Codec:              "hevc",
		DolbyVision:        true,
		DolbyVisionProfile: 8,
	}
	var thumbnailInput string
	var thumbnailOutput string

	result, err := buildDolbyVisionDirectCopyResult(
		context.Background(),
		sourcePath,
		sourceProbe,
		func(ctx context.Context, path string) (ffmpegpkg.VideoProbe, error) {
			if path != sourcePath {
				t.Fatalf("probe path=%s want=%s", path, sourcePath)
			}
			return ffmpegpkg.VideoProbe{
				Duration: 95.4,
				Width:    3840,
				Height:   2160,
				Codec:    "hevc",
			}, nil
		},
		func(ctx context.Context, inputPath, outputPath string) error {
			thumbnailInput = inputPath
			thumbnailOutput = outputPath
			return nil
		},
	)
	if err != nil {
		t.Fatalf("buildDolbyVisionDirectCopyResult() error = %v", err)
	}

	if result.TranscodedPath != sourcePath {
		t.Fatalf("expected direct copy source as transcoded path, got=%s want=%s", result.TranscodedPath, sourcePath)
	}
	if result.ThumbnailPath != "/storage/videos/video-1/thumb.jpg" {
		t.Fatalf("unexpected thumbnail path: %s", result.ThumbnailPath)
	}
	if thumbnailInput != sourcePath || thumbnailOutput != result.ThumbnailPath {
		t.Fatalf("thumbnail input/output = %s/%s, want %s/%s", thumbnailInput, thumbnailOutput, sourcePath, result.ThumbnailPath)
	}
	if result.Duration != 95 || result.Width != 3840 || result.Height != 2160 {
		t.Fatalf("unexpected media dimensions: duration=%d width=%d height=%d", result.Duration, result.Width, result.Height)
	}
	if result.Metadata["transcode_mode"] != "direct_copy" {
		t.Fatalf("expected direct_copy mode, got=%v", result.Metadata["transcode_mode"])
	}
	if result.Metadata["transcode_profile"] != "dv_source_copy" {
		t.Fatalf("expected dv_source_copy profile, got=%v", result.Metadata["transcode_profile"])
	}
	if result.Metadata["playback_path"] != sourcePath {
		t.Fatalf("expected playback_path=%s, got=%v", sourcePath, result.Metadata["playback_path"])
	}
	if result.Metadata["playback_codec"] != "hevc" {
		t.Fatalf("expected playback_codec=hevc, got=%v", result.Metadata["playback_codec"])
	}

	playbackCompat, ok := result.Metadata[services.PlaybackCompatibilityMetadataKey].(map[string]any)
	if !ok {
		t.Fatalf("expected playback_compat map, got %#v", result.Metadata[services.PlaybackCompatibilityMetadataKey])
	}
	if playbackCompat["source_playback_path"] != sourcePath {
		t.Fatalf("expected source_playback_path=%s, got=%v", sourcePath, playbackCompat["source_playback_path"])
	}
	if playbackCompat["dolby_vision_risk"] != true {
		t.Fatalf("expected dolby_vision_risk=true, got=%v", playbackCompat["dolby_vision_risk"])
	}
	sourceBlock, ok := playbackCompat["source"].(map[string]any)
	if !ok {
		t.Fatalf("expected source playback block, got %#v", playbackCompat["source"])
	}
	outputBlock, ok := playbackCompat["output"].(map[string]any)
	if !ok {
		t.Fatalf("expected output playback block, got %#v", playbackCompat["output"])
	}
	if sourceBlock["dolby_vision"] != true || outputBlock["dolby_vision"] != true {
		t.Fatalf("expected source and output to both be Dolby Vision, source=%v output=%v", sourceBlock["dolby_vision"], outputBlock["dolby_vision"])
	}
}

func TestShouldPreserveEpisodeDolbyVisionSource(t *testing.T) {
	t.Parallel()

	dolbyVisionProbe := ffmpegpkg.PlaybackCompatibilityProbe{
		VideoStreamFound: true,
		DolbyVision:      true,
	}
	probeErr := errors.New("probe failed")
	tests := []struct {
		name      string
		videoType string
		probe     ffmpegpkg.PlaybackCompatibilityProbe
		probeErr  error
		want      bool
	}{
		{
			name:      "episode dolby vision",
			videoType: " Episode ",
			probe:     dolbyVisionProbe,
			want:      true,
		},
		{
			name:      "movie dolby vision",
			videoType: "movie",
			probe:     dolbyVisionProbe,
			want:      false,
		},
		{
			name:      "probe error",
			videoType: "episode",
			probe:     dolbyVisionProbe,
			probeErr:  probeErr,
			want:      false,
		},
		{
			name:      "no video stream",
			videoType: "episode",
			probe: ffmpegpkg.PlaybackCompatibilityProbe{
				DolbyVision: true,
			},
			want: false,
		},
		{
			name:      "not dolby vision",
			videoType: "episode",
			probe: ffmpegpkg.PlaybackCompatibilityProbe{
				VideoStreamFound: true,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := shouldPreserveEpisodeDolbyVisionSource(tt.videoType, tt.probe, tt.probeErr)
			if got != tt.want {
				t.Fatalf("shouldPreserveEpisodeDolbyVisionSource()=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestPreserveEpisodeDolbyVisionSourceReusesExistingTargetWhenInputWasMoved(t *testing.T) {
	t.Parallel()

	storageRoot := t.TempDir()
	videoID := uuid.New()
	video := models.Video{
		ID:   videoID,
		Type: "episode",
	}
	inputPath := filepath.Join(t.TempDir(), "upload.mkv")
	targetPath := filepath.Join(storageRoot, "videos", videoID.String(), "source-dv.mkv")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("dv-source"), 0o644); err != nil {
		t.Fatalf("write target source: %v", err)
	}
	processor := &Processor{storageRoot: storageRoot}
	probe := ffmpegpkg.PlaybackCompatibilityProbe{
		VideoStreamFound: true,
		DolbyVision:      true,
	}

	got, err := processor.preserveEpisodeDolbyVisionSource(video, inputPath, probe, nil)
	if err != nil {
		t.Fatalf("preserveEpisodeDolbyVisionSource() error = %v", err)
	}
	if got != targetPath {
		t.Fatalf("preserved path=%s want=%s", got, targetPath)
	}
	raw, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read target source: %v", err)
	}
	if string(raw) != "dv-source" {
		t.Fatalf("target source content changed: %q", raw)
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
