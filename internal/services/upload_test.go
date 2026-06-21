package services

import "testing"

func TestPredictedUploadStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		videoType string
		want      string
	}{
		{videoType: "short", want: "uploaded"},
		{videoType: "movie", want: "scraping"},
		{videoType: "episode", want: "scraping"},
		{videoType: "av", want: "scraping"},
	}

	for _, tt := range tests {
		t.Run(tt.videoType, func(t *testing.T) {
			if got := predictedUploadStatus(tt.videoType); got != tt.want {
				t.Fatalf("predictedUploadStatus(%q) = %q, want %q", tt.videoType, got, tt.want)
			}
		})
	}
}

func TestShouldRefreshExistingVideoOriginalPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status string
		want   bool
	}{
		{status: "uploaded", want: true},
		{status: "scraping", want: true},
		{status: "tv_pending", want: true},
		{status: "av_scrape_pending", want: true},
		{status: "failed", want: true},
		{status: "ready", want: false},
		{status: "processing", want: false},
		{status: "existing", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := shouldRefreshExistingVideoOriginalPath(tt.status); got != tt.want {
				t.Fatalf("shouldRefreshExistingVideoOriginalPath(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
