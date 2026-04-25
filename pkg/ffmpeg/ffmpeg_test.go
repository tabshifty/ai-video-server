package ffmpeg

import "testing"

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
	args := buildTranscodeVideoArgs("/tmp/in.mov", "/tmp/out.mp4", TranscodeProfileHEVCPrimary, TranscodeOptions{CRF: "21"})

	assertArgPair(t, args, "-c:v", "hevc_videotoolbox")
	assertArgPair(t, args, "-pix_fmt", "yuv420p")
	assertArgPair(t, args, "-tag:v", "hvc1")
	assertArgPair(t, args, "-crf", "21")
	assertArgPair(t, args, "-ac", "2")
	assertArgAbsent(t, args, "-preset")
}

func TestBuildTranscodeVideoArgsForAvcCompat(t *testing.T) {
	args := buildTranscodeVideoArgs("/tmp/in.mov", "/tmp/out.mp4", TranscodeProfileAVCCompat, TranscodeOptions{VideoBitrateKbps: 4200})

	assertArgPair(t, args, "-c:v", "h264_videotoolbox")
	assertArgPair(t, args, "-pix_fmt", "yuv420p")
	assertArgPair(t, args, "-b:v", "4200k")
	assertArgPair(t, args, "-maxrate", "4200k")
	assertArgPair(t, args, "-bufsize", "8400k")
	assertArgPair(t, args, "-ac", "2")
	assertArgAbsent(t, args, "-preset")
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
	if probe.Duration != 120.5 {
		t.Fatalf("expected duration 120.5, got %v", probe.Duration)
	}
}
