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

func TestResolveProfiledPlayableSource_RejectsUnknownProfile(t *testing.T) {
	video := models.Video{
		TranscodedPath: "/tmp/video-avc.mp4",
	}

	if _, err := resolveProfiledPlayableSource(video, "unknown"); err == nil {
		t.Fatal("expected error for unknown profile")
	}
}
