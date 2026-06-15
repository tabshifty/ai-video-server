package handlers

import (
	"testing"

	"video-server/internal/models"
)

func TestResolveProfiledPlayableSource_DefaultsToPrimary(t *testing.T) {
	video := models.Video{
		TranscodedPath: "/tmp/video-avc.mp4",
	}

	path, err := resolveProfiledPlayableSource(video, "")
	if err != nil {
		t.Fatalf("resolveProfiledPlayableSource() error = %v", err)
	}
	if path != "/tmp/video-avc.mp4" {
		t.Fatalf("expected primary path, got %s", path)
	}
}

func TestResolveProfiledPlayableSource_UsesCompatPathWhenRequested(t *testing.T) {
	video := models.Video{
		TranscodedPath: "/tmp/video-avc.mp4",
		Metadata:       []byte(`{"compat_transcoded_path":"/tmp/video-avc.mp4"}`),
	}

	path, err := resolveProfiledPlayableSource(video, "compat")
	if err != nil {
		t.Fatalf("resolveProfiledPlayableSource() error = %v", err)
	}
	if path != "/tmp/video-avc.mp4" {
		t.Fatalf("expected compat path, got %s", path)
	}
}

func TestResolveProfiledPlayableSource_UsesPrimaryPathWhenCompatMissing(t *testing.T) {
	video := models.Video{
		TranscodedPath: "/tmp/video-avc.mp4",
		Metadata:       []byte(`{}`),
	}

	path, err := resolveProfiledPlayableSource(video, "compat")
	if err != nil {
		t.Fatalf("resolveProfiledPlayableSource() error = %v", err)
	}
	if path != "/tmp/video-avc.mp4" {
		t.Fatalf("expected primary path fallback, got %s", path)
	}
}

func TestResolveProfiledPlayableSource_UsesDolbyVisionSourcePathWhenRequested(t *testing.T) {
	video := models.Video{
		TranscodedPath: "/tmp/video-avc.mp4",
		Metadata:       []byte(`{"playback_compat":{"version":1,"status":"ok","source_playback_path":"/tmp/source-dv.mkv"}}`),
	}

	path, err := resolveProfiledPlayableSource(video, "dv_source")
	if err != nil {
		t.Fatalf("resolveProfiledPlayableSource() error = %v", err)
	}
	if path != "/tmp/source-dv.mkv" {
		t.Fatalf("expected dv source path, got %s", path)
	}
}

func TestResolveProfiledPlayableSource_RejectsUnknownProfile(t *testing.T) {
	video := models.Video{
		TranscodedPath: "/tmp/video-avc.mp4",
	}

	if _, err := resolveProfiledPlayableSource(video, "unknown"); err == nil {
		t.Fatal("expected error for unknown profile")
	}
}

func TestChooseMovieBackdropVariantPathUsesOnlyLocalDownloadedBackdrop(t *testing.T) {
	metadata := []byte(`{
		"backdrop_path": "/var/video-server/videos/movie-1/backdrop.jpg",
		"tmdb": {
			"backdrop_path": "/tmdb-remote-backdrop.jpg"
		}
	}`)

	got := chooseVideoThumbnailVariantPath("movie", metadata, "backdrop", "/var/video-server/videos/movie-1/poster.jpg")

	if got != "/var/video-server/videos/movie-1/backdrop.jpg" {
		t.Fatalf("expected local movie backdrop file, got=%s", got)
	}
}

func TestChooseMovieBackdropVariantPathRejectsTMDBRelativePath(t *testing.T) {
	metadata := []byte(`{
		"tmdb": {
			"backdrop_path": "/tmdb-remote-backdrop.jpg"
		}
	}`)
	fallback := "/var/video-server/videos/movie-1/poster.jpg"

	got := chooseVideoThumbnailVariantPath("movie", metadata, "backdrop", fallback)

	if got != fallback {
		t.Fatalf("expected fallback when only TMDB relative path exists, got=%s", got)
	}
}
