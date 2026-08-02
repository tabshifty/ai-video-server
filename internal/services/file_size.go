package services

import (
	"fmt"
	"os"
	"strings"
)

// RequiredRegularFileSize returns the positive size of a generated media file.
func RequiredRegularFileSize(path string) (int64, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return 0, fmt.Errorf("file path is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("stat file %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return 0, fmt.Errorf("file %s is not a regular file", path)
	}
	if info.Size() <= 0 {
		return 0, fmt.Errorf("file %s is empty", path)
	}
	return info.Size(), nil
}
