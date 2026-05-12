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
	if plan.CRF != "23" {
		t.Fatalf("expected crf=23, got=%s", plan.CRF)
	}
	if plan.TargetBitrateKbps != 0 {
		t.Fatalf("expected target bitrate 0 in fallback mode, got=%d", plan.TargetBitrateKbps)
	}
	if plan.TranscodeProfile != transcodeProfileHEVCLongform {
		t.Fatalf("expected hevc longform profile, got=%s", plan.TranscodeProfile)
	}
}

func TestBuildTranscodePlanLongformUsesCRFEvenWithSourceBitrate(t *testing.T) {
	tests := []struct {
		name      string
		videoType string
	}{
		{name: "movie", videoType: "movie"},
		{name: "episode", videoType: "episode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := buildTranscodePlan(ffmpeg.VideoProbe{
				Width:         3840,
				Height:        2160,
				BitrateKbps:   12000,
				AudioChannels: 6,
			}, nil, tt.videoType)

			if plan.Mode != transcodeModeCRFFallback {
				t.Fatalf("expected crf mode, got=%s", plan.Mode)
			}
			if plan.CRF != "23" {
				t.Fatalf("expected crf=23, got=%s", plan.CRF)
			}
			if plan.TargetBitrateKbps != 0 {
				t.Fatalf("expected no bitrate target for longform, got=%d", plan.TargetBitrateKbps)
			}
			if plan.SourceBitrateKbps != 12000 {
				t.Fatalf("expected source bitrate preserved, got=%d", plan.SourceBitrateKbps)
			}
			if plan.TranscodeProfile != transcodeProfileHEVCLongform {
				t.Fatalf("expected hevc longform profile, got=%s", plan.TranscodeProfile)
			}
		})
	}
}

func TestBuildTranscodePlanNonLongformKeepsBitrateStrategy(t *testing.T) {
	plan := buildTranscodePlan(ffmpeg.VideoProbe{
		Width:       1920,
		Height:      1080,
		BitrateKbps: 6200,
	}, nil, "av")

	if plan.Mode != transcodeModeBitrate {
		t.Fatalf("expected bitrate mode, got=%s", plan.Mode)
	}
	if plan.TargetBitrateKbps != 4000 {
		t.Fatalf("expected capped bitrate 4000, got=%d", plan.TargetBitrateKbps)
	}
	if plan.TranscodeProfile != transcodeProfileAVCCompat {
		t.Fatalf("expected avc compat profile, got=%s", plan.TranscodeProfile)
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
	if metadata["transcode_profile"] != transcodeProfileAVCCompat {
		t.Fatalf("expected transcode_profile avc compat, got %v", metadata["transcode_profile"])
	}
	if metadata["crf"] != "23" {
		t.Fatalf("expected crf=23, got %v", metadata["crf"])
	}
	if metadata["audio_codec"] != "aac" {
		t.Fatalf("expected audio_codec aac, got %v", metadata["audio_codec"])
	}
	if metadata["audio_channels"] != 2 {
		t.Fatalf("expected audio_channels 2, got %v", metadata["audio_channels"])
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
	if metadata["transcode_profile"] != transcodeProfileAVCCompat {
		t.Fatalf("expected transcode_profile avc compat, got %v", metadata["transcode_profile"])
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
	if metadata["audio_codec"] != "aac" {
		t.Fatalf("expected audio_codec aac, got %v", metadata["audio_codec"])
	}
	if metadata["audio_channels"] != 2 {
		t.Fatalf("expected audio_channels 2, got %v", metadata["audio_channels"])
	}
	if metadata["audio_downmixed"] != false {
		t.Fatalf("expected audio_downmixed false, got %v", metadata["audio_downmixed"])
	}
	if _, ok := metadata["crf"]; ok {
		t.Fatalf("did not expect crf metadata in bitrate mode")
	}
	if _, ok := metadata["probe_error"]; ok {
		t.Fatalf("did not expect probe_error metadata")
	}
}

func TestResolveProbeFieldsMarksAudioDownmixedWhenSourceHasMoreChannels(t *testing.T) {
	plan := transcodePlan{
		Mode:                transcodeModeBitrate,
		ResolutionTier:      resolutionTier1080,
		SourceBitrateKbps:   5200,
		TargetBitrateKbps:   4000,
		BitrateCapped:       true,
		SourceAudioChannels: 6,
	}
	_, _, _, metadata := resolveProbeFields(ffmpeg.VideoProbe{
		Duration:      12.6,
		Width:         1920,
		Height:        1080,
		Codec:         "h265",
		AudioCodec:    "aac",
		AudioChannels: 2,
	}, nil, plan)

	if metadata["audio_codec"] != "aac" {
		t.Fatalf("expected audio_codec aac, got %v", metadata["audio_codec"])
	}
	if metadata["audio_channels"] != 2 {
		t.Fatalf("expected audio_channels 2, got %v", metadata["audio_channels"])
	}
	if metadata["audio_downmixed"] != true {
		t.Fatalf("expected audio_downmixed true, got %v", metadata["audio_downmixed"])
	}
}

func TestBuildPlaybackMetadataUsesSingleAVCOutput(t *testing.T) {
	metadata := buildPlaybackMetadata(
		transcodeOutputProfile{
			Path:             "/tmp/video-avc.mp4",
			PlaybackCodec:    "h264",
			TranscodeProfile: transcodeProfileAVCCompat,
		},
	)

	if metadata["playback_codec"] != "h264" {
		t.Fatalf("expected playback_codec h264, got %v", metadata["playback_codec"])
	}
	if metadata["playback_path"] != "/tmp/video-avc.mp4" {
		t.Fatalf("expected playback_path /tmp/video-avc.mp4, got %v", metadata["playback_path"])
	}
	if _, ok := metadata["compat_transcoded_path"]; ok {
		t.Fatalf("did not expect compat_transcoded_path, got %v", metadata["compat_transcoded_path"])
	}
	if _, ok := metadata["playback_profiles"]; ok {
		t.Fatalf("did not expect playback_profiles, got %v", metadata["playback_profiles"])
	}
}

func TestChooseTranscodeOutputProfileUsesHEVCForLongform(t *testing.T) {
	outputDir := filepath.Join("/tmp", "videos", "abc")
	tests := []struct {
		name      string
		videoType string
	}{
		{name: "movie", videoType: "movie"},
		{name: "episode", videoType: "episode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := chooseTranscodeOutputProfile(outputDir, tt.videoType)

			if profile.Path != filepath.Join(outputDir, "video-hevc.mp4") {
				t.Fatalf("expected hevc output path, got=%s", profile.Path)
			}
			if profile.PlaybackCodec != "hevc" {
				t.Fatalf("expected playback codec hevc, got=%s", profile.PlaybackCodec)
			}
			if profile.FFmpegProfile != ffmpeg.TranscodeProfileHEVCPrimary {
				t.Fatalf("expected ffmpeg hevc profile, got=%s", profile.FFmpegProfile)
			}
			if !profile.SpatialAQ {
				t.Fatalf("expected spatial aq enabled for longform")
			}
			if profile.TranscodeProfile != transcodeProfileHEVCLongform {
				t.Fatalf("expected hevc longform metadata profile, got=%s", profile.TranscodeProfile)
			}
		})
	}
}

func TestChooseTranscodeOutputProfileKeepsAVCForOtherTypes(t *testing.T) {
	outputDir := filepath.Join("/tmp", "videos", "abc")
	tests := []struct {
		name      string
		videoType string
	}{
		{name: "short", videoType: "short"},
		{name: "av", videoType: "av"},
		{name: "unknown", videoType: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := chooseTranscodeOutputProfile(outputDir, tt.videoType)

			if profile.Path != filepath.Join(outputDir, "video-avc.mp4") {
				t.Fatalf("expected avc output path, got=%s", profile.Path)
			}
			if profile.PlaybackCodec != "h264" {
				t.Fatalf("expected playback codec h264, got=%s", profile.PlaybackCodec)
			}
			if profile.FFmpegProfile != ffmpeg.TranscodeProfileAVCCompat {
				t.Fatalf("expected ffmpeg avc profile, got=%s", profile.FFmpegProfile)
			}
			if profile.SpatialAQ {
				t.Fatalf("did not expect spatial aq for non-longform")
			}
			if profile.TranscodeProfile != transcodeProfileAVCCompat {
				t.Fatalf("expected avc compat metadata profile, got=%s", profile.TranscodeProfile)
			}
		})
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
