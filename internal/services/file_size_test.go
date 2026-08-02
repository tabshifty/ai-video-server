package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRequiredRegularFileSize(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "video.mp4")
	if err := os.WriteFile(filePath, []byte("video-data"), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	size, err := RequiredRegularFileSize(filePath)
	if err != nil {
		t.Fatalf("RequiredRegularFileSize() error = %v", err)
	}
	if size != int64(len("video-data")) {
		t.Fatalf("RequiredRegularFileSize() = %d, want %d", size, len("video-data"))
	}
}

func TestRequiredRegularFileSizeRejectsInvalidTargets(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	emptyFile := filepath.Join(dir, "empty.mp4")
	if err := os.WriteFile(emptyFile, nil, 0o600); err != nil {
		t.Fatalf("write empty file: %v", err)
	}

	for _, path := range []string{
		filepath.Join(dir, "missing.mp4"),
		dir,
		emptyFile,
	} {
		if _, err := RequiredRegularFileSize(path); err == nil {
			t.Fatalf("RequiredRegularFileSize(%q) expected error", path)
		}
	}
}
