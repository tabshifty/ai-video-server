package ffmpeg

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPreferredHardwareHevcEncoder(t *testing.T) {
	if got := preferredHardwareHevcEncoder(); got != "hevc_videotoolbox" {
		t.Fatalf("preferredHardwareHevcEncoder()=%s want=hevc_videotoolbox", got)
	}
}

func TestPreferredHardwareAvcEncoder(t *testing.T) {
	if got := preferredHardwareAvcEncoder(); got != "h264_videotoolbox" {
		t.Fatalf("preferredHardwareAvcEncoder()=%s want=h264_videotoolbox", got)
	}
}

func TestBuildTranscodeVideoArgsForHevcPrimary(t *testing.T) {
	args := buildTranscodeVideoArgs("/tmp/in.mov", "/tmp/out.mp4", TranscodeProfileHEVCPrimary, TranscodeOptions{CRF: "23", SpatialAQ: true})

	assertArgPair(t, args, "-c:v", "hevc_videotoolbox")
	assertArgPair(t, args, "-pix_fmt", "yuv420p")
	assertArgPair(t, args, "-tag:v", "hvc1")
	assertArgPair(t, args, "-allow_sw", "0")
	assertArgPair(t, args, "-spatial_aq", "1")
	assertArgPair(t, args, "-crf", "23")
	assertArgPair(t, args, "-map", "0:v:0")
	assertArgPair(t, args, "-map", "0:a?")
	assertArgPair(t, args, "-c:a", "aac")
	assertArgPair(t, args, "-b:a", "128k")
	assertArgAbsent(t, args, "-ac")
	assertArgAbsent(t, args, "-preset")
	assertArgAbsent(t, args, "-x265-params")
}

func TestBuildTranscodeVideoArgsForAvcCompat(t *testing.T) {
	args := buildTranscodeVideoArgs("/tmp/in.mov", "/tmp/out.mp4", TranscodeProfileAVCCompat, TranscodeOptions{VideoBitrateKbps: 4200})

	assertArgPair(t, args, "-c:v", "h264_videotoolbox")
	assertArgPair(t, args, "-pix_fmt", "yuv420p")
	assertArgPair(t, args, "-allow_sw", "0")
	assertArgPair(t, args, "-b:v", "4200k")
	assertArgPair(t, args, "-maxrate", "8400k")
	assertArgPair(t, args, "-bufsize", "16800k")
	assertArgPair(t, args, "-map", "0:v:0")
	assertArgPair(t, args, "-map", "0:a?")
	assertArgPair(t, args, "-c:a", "aac")
	assertArgPair(t, args, "-b:a", "128k")
	assertArgAbsent(t, args, "-ac")
	assertArgAbsent(t, args, "-preset")
	assertArgAbsent(t, args, "-x265-params")
	assertArgAbsent(t, args, "-spatial_aq")
}

func TestBuildConvertSubtitleToWebVTTArgs(t *testing.T) {
	args := buildConvertSubtitleToWebVTTArgs("/tmp/in.ass", "/tmp/out.vtt")

	assertArgPair(t, args, "-i", "/tmp/in.ass")
	assertArgPair(t, args, "-c:s", "webvtt")
	assertArgPair(t, args, "-f", "webvtt")
	if args[len(args)-1] != "/tmp/out.vtt" {
		t.Fatalf("expected output path last, got args=%v", args)
	}
	assertArgAbsent(t, args, "-map")
}

func TestConvertToWebPFallsBackWhenFFmpegMissing(t *testing.T) {
	root := t.TempDir()
	inputPath := filepath.Join(root, "input.jpg")
	outputPath := filepath.Join(root, "output.webp")
	if err := os.WriteFile(inputPath, []byte("jpeg"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("create bin dir: %v", err)
	}
	cwebpPath := filepath.Join(binDir, "cwebp")
	if err := os.WriteFile(cwebpPath, []byte(`#!/bin/sh
set -eu
input=""
output=""
expect_output=0
skip_next=0
for arg in "$@"; do
	if [ "$skip_next" -eq 1 ]; then
		skip_next=0
		continue
	fi
	case "$arg" in
		-o)
			expect_output=1
			;;
		-q)
			skip_next=1
			;;
		-*)
			;;
		*)
			if [ "$expect_output" -eq 1 ]; then
				output="$arg"
				expect_output=0
			else
				input="$arg"
			fi
			;;
	esac
done
: "${input:?missing input}"
: "${output:?missing output}"
: > "$output"
`), 0o755); err != nil {
		t.Fatalf("write cwebp shim: %v", err)
	}

	t.Setenv("PATH", binDir)

	if err := ConvertToWebP(context.Background(), inputPath, outputPath, 82); err != nil {
		t.Fatalf("ConvertToWebP() error = %v", err)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output file missing: %v", err)
	}
}

func TestBuildExtractSubtitleToAssArgs(t *testing.T) {
	args := buildExtractSubtitleToAssArgs("/tmp/in.mkv", 3, "/tmp/out.ass")

	assertArgPair(t, args, "-i", "/tmp/in.mkv")
	assertArgPair(t, args, "-map", "0:3")
	assertArgPair(t, args, "-c:s", "copy")
	if args[len(args)-1] != "/tmp/out.ass" {
		t.Fatalf("expected output path last, got args=%v", args)
	}
	assertArgAbsent(t, args, "-f")
	assertArgAbsent(t, args, "webvtt")
}

func TestParseBitrateKbps(t *testing.T) {
	tests := []struct {
		name          string
		streamBitrate string
		formatBitrate string
		want          int
	}{
		{
			name:          "prefer stream bitrate",
			streamBitrate: "5400000",
			formatBitrate: "3200000",
			want:          5400,
		},
		{
			name:          "fallback to format bitrate",
			streamBitrate: "",
			formatBitrate: "4500000",
			want:          4500,
		},
		{
			name:          "invalid bitrate returns zero",
			streamBitrate: "invalid",
			formatBitrate: "bad",
			want:          0,
		},
		{
			name:          "non positive bitrate returns zero",
			streamBitrate: "-1",
			formatBitrate: "0",
			want:          0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBitrateKbps(tt.streamBitrate, tt.formatBitrate)
			if got != tt.want {
				t.Fatalf("parseBitrateKbps() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParsePlaybackCompatibilityProbeOutputDetectsDolbyVisionSideData(t *testing.T) {
	raw := []byte(`{
		"streams": [
			{"codec_type": "audio", "codec_name": "aac"},
			{
				"codec_type": "video",
				"codec_name": "hevc",
				"codec_tag_string": "dvh1",
				"profile": "Main 10",
				"pix_fmt": "yuv420p10le",
				"color_transfer": "smpte2084",
				"color_primaries": "bt2020",
				"color_space": "bt2020nc",
				"side_data_list": [
					{
						"side_data_type": "DOVI configuration record",
						"dv_profile": "8",
						"dv_level": 6,
						"dv_bl_signal_compatibility_id": "1"
					}
				]
			}
		]
	}`)

	probe, err := parsePlaybackCompatibilityProbeOutput(raw)
	if err != nil {
		t.Fatalf("parsePlaybackCompatibilityProbeOutput() error = %v", err)
	}
	if !probe.VideoStreamFound {
		t.Fatalf("expected video stream")
	}
	if !probe.DolbyVision {
		t.Fatalf("expected Dolby Vision detection")
	}
	if probe.DolbyVisionProfile != 8 || probe.DolbyVisionLevel != 6 || probe.DolbyVisionCompatID != 1 {
		t.Fatalf("unexpected Dolby Vision fields: profile=%d level=%d compat=%d", probe.DolbyVisionProfile, probe.DolbyVisionLevel, probe.DolbyVisionCompatID)
	}
	if !probe.HDR {
		t.Fatalf("expected HDR detection")
	}
}

func TestParsePlaybackCompatibilityProbeOutputDoesNotTreatPlainHDRAsDolbyVision(t *testing.T) {
	raw := []byte(`{
		"streams": [
			{
				"codec_type": "video",
				"codec_name": "hevc",
				"codec_tag_string": "hvc1",
				"profile": "Main 10",
				"pix_fmt": "p010le",
				"color_transfer": "arib-std-b67",
				"color_primaries": "bt2020",
				"color_space": "bt2020nc"
			}
		]
	}`)

	probe, err := parsePlaybackCompatibilityProbeOutput(raw)
	if err != nil {
		t.Fatalf("parsePlaybackCompatibilityProbeOutput() error = %v", err)
	}
	if !probe.VideoStreamFound {
		t.Fatalf("expected video stream")
	}
	if probe.DolbyVision {
		t.Fatalf("did not expect Dolby Vision detection")
	}
	if !probe.HDR {
		t.Fatalf("expected HDR detection")
	}
}

func TestParsePlaybackCompatibilityProbeOutputWithoutVideoStream(t *testing.T) {
	raw := []byte(`{"streams":[{"codec_type":"audio","codec_name":"aac"}]}`)

	probe, err := parsePlaybackCompatibilityProbeOutput(raw)
	if err != nil {
		t.Fatalf("parsePlaybackCompatibilityProbeOutput() error = %v", err)
	}
	if probe.VideoStreamFound {
		t.Fatalf("did not expect video stream")
	}
	if probe.DolbyVision || probe.HDR {
		t.Fatalf("did not expect playback risk flags, got=%#v", probe)
	}
}

func assertArgPair(t *testing.T, args []string, key, value string) {
	t.Helper()
	for idx := 0; idx < len(args)-1; idx++ {
		if args[idx] == key && args[idx+1] == value {
			return
		}
	}
	t.Fatalf("expected args to contain %s %s, got=%v", key, value, args)
}

func assertArgAbsent(t *testing.T, args []string, key string) {
	t.Helper()
	for _, arg := range args {
		if arg == key {
			t.Fatalf("expected args not to contain %s, got=%v", key, args)
		}
	}
}

func TestParseProgressValueToSeconds(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		want  int
		ok    bool
	}{
		{
			name:  "out time ms parse",
			key:   "out_time_ms",
			value: "2500000",
			want:  3,
			ok:    true,
		},
		{
			name:  "out time us parse",
			key:   "out_time_us",
			value: "1800000",
			want:  2,
			ok:    true,
		},
		{
			name:  "out time text parse",
			key:   "out_time",
			value: "00:00:03.500000",
			want:  4,
			ok:    true,
		},
		{
			name:  "invalid format",
			key:   "out_time",
			value: "bad",
			want:  0,
			ok:    false,
		},
		{
			name:  "unknown key",
			key:   "frame",
			value: "12",
			want:  0,
			ok:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseProgressValueToSeconds(tt.key, tt.value)
			if ok != tt.ok {
				t.Fatalf("parseProgressValueToSeconds() ok=%v want=%v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("parseProgressValueToSeconds()=%d want=%d", got, tt.want)
			}
		})
	}
}

func TestBuildTranscodeProgress(t *testing.T) {
	tests := []struct {
		name      string
		processed int
		total     int
		want      TranscodeProgress
	}{
		{
			name:      "normal progress",
			processed: 24,
			total:     60,
			want: TranscodeProgress{
				SourceDurationSeconds: 60,
				ProcessedSeconds:      24,
				RemainingSeconds:      36,
				ProgressPercent:       40,
			},
		},
		{
			name:      "cap over total",
			processed: 99,
			total:     80,
			want: TranscodeProgress{
				SourceDurationSeconds: 80,
				ProcessedSeconds:      80,
				RemainingSeconds:      0,
				ProgressPercent:       100,
			},
		},
		{
			name:      "no total duration",
			processed: 20,
			total:     0,
			want: TranscodeProgress{
				SourceDurationSeconds: 0,
				ProcessedSeconds:      20,
				RemainingSeconds:      0,
				ProgressPercent:       0,
			},
		},
		{
			name:      "negative processed fixed",
			processed: -1,
			total:     30,
			want: TranscodeProgress{
				SourceDurationSeconds: 30,
				ProcessedSeconds:      0,
				RemainingSeconds:      30,
				ProgressPercent:       0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTranscodeProgress(tt.processed, tt.total)
			if got != tt.want {
				t.Fatalf("buildTranscodeProgress()=%+v want=%+v", got, tt.want)
			}
		})
	}
}

func TestIsEncoderUnavailableOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		encoder string
		want    bool
	}{
		{
			name:    "unknown encoder single quote",
			output:  "[vost#0:0] Unknown encoder 'libwebp'",
			encoder: "libwebp",
			want:    true,
		},
		{
			name:    "unknown encoder double quote",
			output:  `[vost#0:0] Unknown encoder "libwebp"`,
			encoder: "libwebp",
			want:    true,
		},
		{
			name:    "encoder not found with libwebp",
			output:  "Error opening output files: Encoder not found (libwebp)",
			encoder: "libwebp",
			want:    true,
		},
		{
			name:    "unknown webp encoder",
			output:  "Unknown encoder 'webp'",
			encoder: "webp",
			want:    true,
		},
		{
			name:    "other ffmpeg error",
			output:  "Invalid argument",
			encoder: "libwebp",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEncoderUnavailableOutput(tt.output, tt.encoder)
			if got != tt.want {
				t.Fatalf("isEncoderUnavailableOutput()=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestParseSubtitleProbeOutput(t *testing.T) {
	raw := []byte(`{
  "streams": [
    {
      "index": 2,
      "codec_name": "subrip",
      "codec_type": "subtitle",
      "tags": {
        "language": "zh",
        "title": "简体中文"
      },
      "disposition": {
        "default": 1
      }
    },
    {
      "index": 3,
      "codec_name": "ass",
      "codec_type": "subtitle",
      "tags": {
        "language": "en",
        "title": "English Signs"
      },
      "disposition": {
        "default": 0
      }
    },
    {
      "index": 4,
      "codec_name": "h264",
      "codec_type": "video"
    }
  ]
}`)

	tracks, err := parseSubtitleProbeOutput(raw)
	if err != nil {
		t.Fatalf("parseSubtitleProbeOutput() error = %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("subtitle track count = %d, want 2", len(tracks))
	}

	if tracks[0].Index != 2 || tracks[0].Codec != "subrip" || tracks[0].Language != "zh" || tracks[0].Title != "简体中文" || !tracks[0].IsDefault {
		t.Fatalf("tracks[0] = %+v", tracks[0])
	}
	if tracks[1].Index != 3 || tracks[1].Codec != "ass" || tracks[1].Language != "en" || tracks[1].Title != "English Signs" || tracks[1].IsDefault {
		t.Fatalf("tracks[1] = %+v", tracks[1])
	}
}

func TestParseSubtitleProbeOutputRejectsInvalidJSON(t *testing.T) {
	if _, err := parseSubtitleProbeOutput([]byte(`{`)); err == nil {
		t.Fatal("expected error for invalid json")
	}
}

func TestParseProbeOutputIncludesAudioStreamMetadata(t *testing.T) {
	raw := []byte(`{
  "streams": [
    {
      "codec_type": "video",
      "width": 1920,
      "height": 1080,
      "codec_name": "hevc",
      "bit_rate": "8200000"
    },
    {
      "codec_type": "audio",
      "codec_name": "aac",
      "channels": 6
    },
    {
      "codec_type": "audio",
      "codec_name": "ac3",
      "channels": 2
    }
  ],
  "format": {
    "duration": "120.5",
    "bit_rate": "9000000"
  }
}`)

	probe, err := parseProbeOutput(raw)
	if err != nil {
		t.Fatalf("parseProbeOutput() error = %v", err)
	}
	if probe.Width != 1920 || probe.Height != 1080 {
		t.Fatalf("unexpected video size: %+v", probe)
	}
	if probe.Codec != "hevc" {
		t.Fatalf("expected video codec hevc, got %s", probe.Codec)
	}
	if probe.BitrateKbps != 8200 {
		t.Fatalf("expected bitrate 8200, got %d", probe.BitrateKbps)
	}
	if probe.AudioCodec != "aac" {
		t.Fatalf("expected audio codec aac, got %s", probe.AudioCodec)
	}
	if probe.AudioChannels != 6 {
		t.Fatalf("expected audio channels 6, got %d", probe.AudioChannels)
	}
	if probe.AudioTrackCount != 2 {
		t.Fatalf("expected audio track count 2, got %d", probe.AudioTrackCount)
	}
	if probe.Duration != 120.5 {
		t.Fatalf("expected duration 120.5, got %v", probe.Duration)
	}
}
