package services

import (
	"context"
	"errors"
	"testing"
)

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

func TestSaveImportedFileRejectsInvalidType(t *testing.T) {
	t.Parallel()

	service := &UploadService{}
	_, err := service.SaveImportedFile(context.Background(), LocalUploadInput{Type: "telegram"}, 0)
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("SaveImportedFile() error=%v, want ErrInvalidType", err)
	}
}

func TestShouldUpdateExistingVideoOriginalPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		refresh bool
		status  string
		want    bool
	}{
		{name: "manual uploaded", refresh: true, status: "uploaded", want: true},
		{name: "import uploaded", refresh: false, status: "uploaded", want: false},
		{name: "manual ready", refresh: true, status: "ready", want: false},
		{name: "import failed", refresh: false, status: "failed", want: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldUpdateExistingVideoOriginalPath(tt.refresh, tt.status); got != tt.want {
				t.Fatalf("shouldUpdateExistingVideoOriginalPath(%v, %q)=%v, want %v", tt.refresh, tt.status, got, tt.want)
			}
		})
	}
}
