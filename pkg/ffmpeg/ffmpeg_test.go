package ffmpeg

import "testing"

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
