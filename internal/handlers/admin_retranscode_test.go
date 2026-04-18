package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"video-server/internal/models"
)

func TestSelectRetranscodeInputPath_PreferOriginal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	originalPath := filepath.Join(dir, "video-original.mp4")
	transcodedPath := filepath.Join(dir, "video-transcoded.mp4")
	if err := os.WriteFile(originalPath, []byte("original"), 0o644); err != nil {
		t.Fatalf("write original: %v", err)
	}
	if err := os.WriteFile(transcodedPath, []byte("transcoded"), 0o644); err != nil {
		t.Fatalf("write transcoded: %v", err)
	}

	input, source := selectRetranscodeInputPath(models.Video{
		OriginalPath:   originalPath,
		TranscodedPath: transcodedPath,
	})
	if input != originalPath || source != "original" {
		t.Fatalf("expected original path selected, got input=%s source=%s", input, source)
	}
}

func TestSelectRetranscodeInputPath_FallbackTranscoded(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	transcodedPath := filepath.Join(dir, "video-transcoded.mp4")
	if err := os.WriteFile(transcodedPath, []byte("transcoded"), 0o644); err != nil {
		t.Fatalf("write transcoded: %v", err)
	}

	input, source := selectRetranscodeInputPath(models.Video{
		OriginalPath:   filepath.Join(dir, "missing-original.mp4"),
		TranscodedPath: transcodedPath,
	})
	if input != transcodedPath || source != "transcoded" {
		t.Fatalf("expected transcoded fallback selected, got input=%s source=%s", input, source)
	}
}

func TestSelectRetranscodeInputPath_NoFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	input, source := selectRetranscodeInputPath(models.Video{
		OriginalPath:   filepath.Join(dir, "missing-original.mp4"),
		TranscodedPath: filepath.Join(dir, "missing-transcoded.mp4"),
	})
	if input != "" || source != "" {
		t.Fatalf("expected empty input and source, got input=%s source=%s", input, source)
	}
}
