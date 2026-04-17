package handlers

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildCaptureTempPath_KeepExtension(t *testing.T) {
	target := filepath.Join("/tmp", "videos", "abc", "thumb.jpg")
	got := buildCaptureTempPath(target)

	if filepath.Ext(got) != ".jpg" {
		t.Fatalf("expected .jpg extension, got %s", filepath.Ext(got))
	}
	if filepath.Dir(got) != filepath.Dir(target) {
		t.Fatalf("expected temp dir %s, got %s", filepath.Dir(target), filepath.Dir(got))
	}
	if strings.HasSuffix(got, ".tmp") {
		t.Fatalf("unexpected temp suffix format: %s", got)
	}
}

func TestBuildCaptureTempPath_FallbackExtension(t *testing.T) {
	target := filepath.Join("/tmp", "videos", "abc", "thumb")
	got := buildCaptureTempPath(target)

	if filepath.Ext(got) != ".jpg" {
		t.Fatalf("expected fallback .jpg extension, got %s", filepath.Ext(got))
	}
}
