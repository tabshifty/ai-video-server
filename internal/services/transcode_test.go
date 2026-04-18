package services

import (
	"errors"
	"path/filepath"
	"testing"

	"video-server/pkg/ffmpeg"
)

func TestDecideVideoBitrate(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		height     int
		source     int
		wantTier   string
		want       int
		wantCapped bool
	}{
		{
			name:       "1080 high bitrate capped to 4000k",
			width:      1920,
			height:     1080,
			source:     6200,
			wantTier:   resolutionTier1080,
			want:       4000,
			wantCapped: true,
		},
		{
			name:       "1080 low bitrate keeps source",
			width:      1920,
			height:     1080,
			source:     3200,
			wantTier:   resolutionTier1080,
			want:       3200,
			wantCapped: false,
		},
		{
			name:       "4k high bitrate capped to 8000k",
			width:      3840,
			height:     2160,
			source:     12000,
			wantTier:   resolutionTier4K,
			want:       8000,
			wantCapped: true,
		},
		{
			name:       "4k lower bitrate keeps source",
			width:      3840,
			height:     2160,
			source:     7600,
			wantTier:   resolutionTier4K,
			want:       7600,
			wantCapped: false,
		},
		{
			name:       "non 1080 and non 4k keeps source",
			width:      1280,
			height:     720,
			source:     2500,
			wantTier:   resolutionTierOther,
			want:       2500,
			wantCapped: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, tier, capped := decideVideoBitrate(tt.width, tt.height, tt.source)
			if target != tt.want {
				t.Fatalf("target bitrate mismatch: got=%d want=%d", target, tt.want)
			}
			if tier != tt.wantTier {
				t.Fatalf("tier mismatch: got=%s want=%s", tier, tt.wantTier)
			}
			if capped != tt.wantCapped {
				t.Fatalf("capped mismatch: got=%v want=%v", capped, tt.wantCapped)
			}
		})
	}
}

func TestBuildTranscodePlanWithoutSourceBitrateFallsBackToCRF(t *testing.T) {
	plan := buildTranscodePlan(ffmpeg.VideoProbe{
		Width:       1920,
		Height:      1080,
		BitrateKbps: 0,
	}, nil, "movie")

	if plan.Mode != transcodeModeCRFFallback {
		t.Fatalf("expected fallback mode, got=%s", plan.Mode)
	}
	if plan.CRF != "21" {
		t.Fatalf("expected crf=21, got=%s", plan.CRF)
	}
	if plan.TargetBitrateKbps != 0 {
		t.Fatalf("expected target bitrate 0 in fallback mode, got=%d", plan.TargetBitrateKbps)
	}
}

func TestResolveProbeFieldsWithProbeErrorUsesDefaults(t *testing.T) {
	plan := transcodePlan{
		Mode:              transcodeModeCRFFallback,
		CRF:               "23",
		ResolutionTier:    resolutionTierOther,
		SourceBitrateKbps: 0,
		TargetBitrateKbps: 0,
		BitrateCapped:     false,
	}
	duration, width, height, metadata := resolveProbeFields(ffmpeg.VideoProbe{}, errors.New("ffprobe failed"), plan)

	if duration != 0 || width != 0 || height != 0 {
		t.Fatalf("expected zero defaults, got duration=%d width=%d height=%d", duration, width, height)
	}
	if metadata["codec"] != "unknown" {
		t.Fatalf("expected codec unknown, got %v", metadata["codec"])
	}
	if metadata["transcode_mode"] != transcodeModeCRFFallback {
		t.Fatalf("expected transcode_mode fallback, got=%v", metadata["transcode_mode"])
	}
	if metadata["crf"] != "23" {
		t.Fatalf("expected crf=23, got %v", metadata["crf"])
	}
	if _, ok := metadata["probe_error"]; !ok {
		t.Fatalf("expected probe_error metadata")
	}
}

func TestResolveProbeFieldsWithValidProbeParsesValues(t *testing.T) {
	plan := transcodePlan{
		Mode:              transcodeModeBitrate,
		ResolutionTier:    resolutionTier1080,
		SourceBitrateKbps: 5200,
		TargetBitrateKbps: 4000,
		BitrateCapped:     true,
	}
	duration, width, height, metadata := resolveProbeFields(ffmpeg.VideoProbe{
		Duration: 12.6,
		Width:    1920,
		Height:   1080,
		Codec:    "h265",
	}, nil, plan)

	if duration != 13 || width != 1920 || height != 1080 {
		t.Fatalf("unexpected parsed values duration=%d width=%d height=%d", duration, width, height)
	}
	if metadata["codec"] != "h265" {
		t.Fatalf("expected codec h265, got %v", metadata["codec"])
	}
	if metadata["transcode_mode"] != transcodeModeBitrate {
		t.Fatalf("expected transcode_mode bitrate, got %v", metadata["transcode_mode"])
	}
	if metadata["resolution_tier"] != resolutionTier1080 {
		t.Fatalf("expected resolution_tier 1080, got %v", metadata["resolution_tier"])
	}
	if metadata["source_bitrate_kbps"] != 5200 {
		t.Fatalf("expected source_bitrate_kbps 5200, got %v", metadata["source_bitrate_kbps"])
	}
	if metadata["target_bitrate_kbps"] != 4000 {
		t.Fatalf("expected target_bitrate_kbps 4000, got %v", metadata["target_bitrate_kbps"])
	}
	if metadata["bitrate_capped"] != true {
		t.Fatalf("expected bitrate_capped true, got %v", metadata["bitrate_capped"])
	}
	if _, ok := metadata["crf"]; ok {
		t.Fatalf("did not expect crf metadata in bitrate mode")
	}
	if _, ok := metadata["probe_error"]; ok {
		t.Fatalf("did not expect probe_error metadata")
	}
}

func TestIsSameFilePath(t *testing.T) {
	base := filepath.Join(t.TempDir(), "videos", "abc", "video.mp4")
	same := filepath.Join(filepath.Dir(base), ".", "video.mp4")
	other := filepath.Join(filepath.Dir(base), "video-2.mp4")

	if !isSameFilePath(base, same) {
		t.Fatalf("expected same file path")
	}
	if isSameFilePath(base, other) {
		t.Fatalf("expected different file path")
	}
	if isSameFilePath("", base) {
		t.Fatalf("expected empty path as different")
	}
}

func TestBuildTranscodeOutputTempPath(t *testing.T) {
	target := filepath.Join("/tmp", "videos", "abc", "video.mp4")
	got := buildTranscodeOutputTempPath(target)
	if got == target {
		t.Fatalf("temp path should differ from target")
	}
	if filepath.Ext(got) != ".mp4" {
		t.Fatalf("expected .mp4 extension, got %s", filepath.Ext(got))
	}
	if filepath.Dir(got) != filepath.Dir(target) {
		t.Fatalf("expected same dir, got %s", filepath.Dir(got))
	}
}
