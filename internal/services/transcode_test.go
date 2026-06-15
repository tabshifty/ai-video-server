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
		videoType  string
		width      int
		height     int
		source     int
		wantTier   string
		want       int
		wantCapped bool
	}{
		{
			name:       "1080 high bitrate capped to 5000k",
			videoType:  "movie",
			width:      1920,
			height:     1080,
			source:     6200,
			wantTier:   resolutionTier1080,
			want:       5000,
			wantCapped: true,
		},
		{
			name:       "1080 low bitrate keeps source",
			videoType:  "movie",
			width:      1920,
			height:     1080,
			source:     3200,
			wantTier:   resolutionTier1080,
			want:       3200,
			wantCapped: false,
		},
		{
			name:       "4k high bitrate capped to 10000k",
			videoType:  "episode",
			width:      3840,
			height:     2160,
			source:     15000,
			wantTier:   resolutionTier4K,
			want:       10000,
			wantCapped: true,
		},
		{
			name:       "4k lower bitrate keeps source",
			videoType:  "movie",
			width:      3840,
			height:     2160,
			source:     9000,
			wantTier:   resolutionTier4K,
			want:       9000,
			wantCapped: false,
		},
		{
			name:       "non 1080 and non 4k keeps source",
			videoType:  "movie",
			width:      1280,
			height:     720,
			source:     2500,
			wantTier:   resolutionTierOther,
			want:       2500,
			wantCapped: false,
		},
		{
			name:       "av 1080 keeps previous 4000k cap",
			videoType:  "av",
			width:      1920,
			height:     1080,
			source:     6200,
			wantTier:   resolutionTier1080,
			want:       4000,
			wantCapped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, tier, capped := decideVideoBitrate(tt.videoType, tt.width, tt.height, tt.source)
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

func TestBuildTranscodePlanLongformUsesBitrateStrategyWhenSourceBitrateKnown(t *testing.T) {
	tests := []struct {
		name       string
		videoType  string
		width      int
		height     int
		sourceKbps int
		wantKbps   int
		wantCapped bool
		wantTier   string
	}{
		{name: "movie 4k capped", videoType: "movie", width: 3840, height: 2160, sourceKbps: 15000, wantKbps: 10000, wantCapped: true, wantTier: resolutionTier4K},
		{name: "episode 1080 capped", videoType: "episode", width: 1920, height: 1080, sourceKbps: 6200, wantKbps: 5000, wantCapped: true, wantTier: resolutionTier1080},
		{name: "episode 1080 keeps lower source", videoType: "episode", width: 1920, height: 1080, sourceKbps: 3200, wantKbps: 3200, wantCapped: false, wantTier: resolutionTier1080},
		{name: "movie 4k keeps lower source", videoType: "movie", width: 3840, height: 2160, sourceKbps: 9000, wantKbps: 9000, wantCapped: false, wantTier: resolutionTier4K},
		{name: "movie 720 keeps source", videoType: "movie", width: 1280, height: 720, sourceKbps: 2500, wantKbps: 2500, wantCapped: false, wantTier: resolutionTierOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := buildTranscodePlan(ffmpeg.VideoProbe{
				Width:         tt.width,
				Height:        tt.height,
				BitrateKbps:   tt.sourceKbps,
				AudioChannels: 6,
			}, nil, tt.videoType)

			if plan.Mode != transcodeModeBitrate {
				t.Fatalf("expected bitrate mode, got=%s", plan.Mode)
			}
			if plan.CRF != "23" {
				t.Fatalf("expected crf=23, got=%s", plan.CRF)
			}
			if plan.TargetBitrateKbps != tt.wantKbps {
				t.Fatalf("expected target bitrate %d, got=%d", tt.wantKbps, plan.TargetBitrateKbps)
			}
			if plan.SourceBitrateKbps != tt.sourceKbps {
				t.Fatalf("expected source bitrate preserved, got=%d", plan.SourceBitrateKbps)
			}
			if plan.BitrateCapped != tt.wantCapped {
				t.Fatalf("expected bitrate capped=%v, got=%v", tt.wantCapped, plan.BitrateCapped)
			}
			if plan.ResolutionTier != tt.wantTier {
				t.Fatalf("expected resolution tier %s, got=%s", tt.wantTier, plan.ResolutionTier)
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
	if metadata["audio_channels"] != 0 {
		t.Fatalf("expected audio_channels 0, got %v", metadata["audio_channels"])
	}
	if metadata["audio_track_count"] != 0 {
		t.Fatalf("expected audio_track_count 0, got %v", metadata["audio_track_count"])
	}
	if _, ok := metadata["probe_error"]; !ok {
		t.Fatalf("expected probe_error metadata")
	}
}

func TestSubtitleUploadPlanForFilenameKeepsAssAndSsaSource(t *testing.T) {
	tests := []struct {
		filename       string
		uploadedFormat string
		storedFormat   string
		storedMIME     string
		needsWebVTT    bool
	}{
		{filename: "movie.zh.srt", uploadedFormat: "srt", storedFormat: "srt", storedMIME: "application/x-subrip", needsWebVTT: false},
		{filename: "movie.zh.vtt", uploadedFormat: "vtt", storedFormat: "vtt", storedMIME: "text/vtt", needsWebVTT: false},
		{filename: "movie.zh.ass", uploadedFormat: "ass", storedFormat: "ass", storedMIME: "text/x-ssa", needsWebVTT: false},
		{filename: "movie.zh.ssa", uploadedFormat: "ssa", storedFormat: "ass", storedMIME: "text/x-ssa", needsWebVTT: false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			plan, err := subtitleUploadPlanForFilename(tt.filename)
			if err != nil {
				t.Fatalf("subtitleUploadPlanForFilename() error = %v", err)
			}
			if plan.UploadedFormat != tt.uploadedFormat {
				t.Fatalf("UploadedFormat = %s, want %s", plan.UploadedFormat, tt.uploadedFormat)
			}
			if plan.StoredFormat != tt.storedFormat {
				t.Fatalf("StoredFormat = %s, want %s", plan.StoredFormat, tt.storedFormat)
			}
			if plan.StoredMIMEType != tt.storedMIME {
				t.Fatalf("StoredMIMEType = %s, want %s", plan.StoredMIMEType, tt.storedMIME)
			}
			if plan.NeedsWebVTT != tt.needsWebVTT {
				t.Fatalf("NeedsWebVTT = %v, want %v", plan.NeedsWebVTT, tt.needsWebVTT)
			}
		})
	}
}

func TestSubtitleUploadPlanForFilenameRejectsUnsupportedFormats(t *testing.T) {
	if _, err := subtitleUploadPlanForFilename("subtitle.txt"); err == nil {
		t.Fatalf("expected unsupported subtitle format error")
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
		Duration:        12.6,
		Width:           1920,
		Height:          1080,
		Codec:           "h265",
		AudioTrackCount: 2,
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
	if metadata["audio_channels"] != 0 {
		t.Fatalf("expected audio_channels 0, got %v", metadata["audio_channels"])
	}
	if metadata["audio_downmixed"] != false {
		t.Fatalf("expected audio_downmixed false, got %v", metadata["audio_downmixed"])
	}
	if metadata["audio_track_count"] != 2 {
		t.Fatalf("expected audio_track_count 2, got %v", metadata["audio_track_count"])
	}
	if _, ok := metadata["crf"]; ok {
		t.Fatalf("did not expect crf metadata in bitrate mode")
	}
	if _, ok := metadata["probe_error"]; ok {
		t.Fatalf("did not expect probe_error metadata")
	}
}

func TestResolveProbeFieldsKeepsMultichannelAudioMetadataWithoutDownmix(t *testing.T) {
	plan := transcodePlan{
		Mode:                transcodeModeBitrate,
		ResolutionTier:      resolutionTier1080,
		SourceBitrateKbps:   5200,
		TargetBitrateKbps:   4000,
		BitrateCapped:       true,
		SourceAudioChannels: 6,
	}
	_, _, _, metadata := resolveProbeFields(ffmpeg.VideoProbe{
		Duration:        12.6,
		Width:           1920,
		Height:          1080,
		Codec:           "h265",
		AudioCodec:      "aac",
		AudioChannels:   6,
		AudioTrackCount: 3,
	}, nil, plan)

	if metadata["audio_codec"] != "aac" {
		t.Fatalf("expected audio_codec aac, got %v", metadata["audio_codec"])
	}
	if metadata["audio_channels"] != 6 {
		t.Fatalf("expected audio_channels 6, got %v", metadata["audio_channels"])
	}
	if metadata["audio_track_count"] != 3 {
		t.Fatalf("expected audio_track_count 3, got %v", metadata["audio_track_count"])
	}
	if metadata["audio_downmixed"] != false {
		t.Fatalf("expected audio_downmixed false, got %v", metadata["audio_downmixed"])
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

	if !IsSameFilePath(base, same) {
		t.Fatalf("expected same file path")
	}
	if IsSameFilePath(base, other) {
		t.Fatalf("expected different file path")
	}
	if IsSameFilePath("", base) {
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
