package services

import (
	"errors"
	"testing"

	"video-server/pkg/ffmpeg"
)

func TestBuildPlaybackCompatibilityMetadataOK(t *testing.T) {
	t.Parallel()

	metadata := BuildPlaybackCompatibilityMetadata(
		ffmpeg.PlaybackCompatibilityProbe{
			VideoStreamFound:   true,
			Codec:              "hevc",
			CodecTag:           "dvh1",
			DolbyVision:        true,
			DolbyVisionProfile: 8,
			HDR:                true,
		},
		nil,
		ffmpeg.PlaybackCompatibilityProbe{
			VideoStreamFound: true,
			Codec:            "hevc",
			CodecTag:         "hvc1",
			DolbyVision:      false,
		},
		nil,
	)

	if metadata["version"] != playbackCompatibilityVersion {
		t.Fatalf("expected version %d, got %v", playbackCompatibilityVersion, metadata["version"])
	}
	if metadata["status"] != playbackCompatibilityStatusOK {
		t.Fatalf("expected status ok, got %v", metadata["status"])
	}
	if metadata["dolby_vision_risk"] != true {
		t.Fatalf("expected dolby_vision_risk true, got %v", metadata["dolby_vision_risk"])
	}
	source := metadata["source"].(map[string]any)
	if source["dolby_vision"] != true || source["dolby_vision_profile"] != 8 {
		t.Fatalf("unexpected source metadata: %#v", source)
	}
	output := metadata["output"].(map[string]any)
	if output["dolby_vision"] != false || output["codec_tag"] != "hvc1" {
		t.Fatalf("unexpected output metadata: %#v", output)
	}
}

func TestBuildPlaybackCompatibilityMetadataProbeFailed(t *testing.T) {
	t.Parallel()

	metadata := BuildPlaybackCompatibilityMetadata(
		ffmpeg.PlaybackCompatibilityProbe{},
		errors.New("source ffprobe failed"),
		ffmpeg.PlaybackCompatibilityProbe{VideoStreamFound: false},
		nil,
	)

	if metadata["status"] != playbackCompatibilityStatusProbeFailed {
		t.Fatalf("expected probe_failed status, got %v", metadata["status"])
	}
	if metadata["source_probe_ok"] != false || metadata["output_probe_ok"] != false {
		t.Fatalf("expected probe ok flags false, got source=%v output=%v", metadata["source_probe_ok"], metadata["output_probe_ok"])
	}
	if metadata["source_probe_error"] != "source ffprobe failed" {
		t.Fatalf("expected source probe error, got %v", metadata["source_probe_error"])
	}
	if metadata["output_probe_error"] != "output video stream not found" {
		t.Fatalf("expected output video missing error, got %v", metadata["output_probe_error"])
	}
}
