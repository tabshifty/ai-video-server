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

