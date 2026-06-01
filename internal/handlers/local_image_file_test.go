package handlers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenLocalImageFileWithReturnsFileInfo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "thumb.jpg")
	if err := os.WriteFile(path, []byte("jpg"), 0o644); err != nil {
		t.Fatalf("write temp image: %v", err)
	}

	file, info, err := openLocalImageFileWith(context.Background(), path, time.Second, make(chan struct{}, 1), os.Open)
	if err != nil {
		t.Fatalf("openLocalImageFileWith() error = %v", err)
	}
	defer file.Close()

	if info == nil || info.Size() != 3 {
		t.Fatalf("expected file info size 3, got %#v", info)
	}
}

func TestOpenLocalImageFileWithTimesOut(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{})
	limiter := make(chan struct{}, 1)
	opener := func(string) (*os.File, error) {
		close(started)
		<-release
		return nil, os.ErrNotExist
	}

	_, _, err := openLocalImageFileWith(context.Background(), "thumb.jpg", 10*time.Millisecond, limiter, opener)
	if !errors.Is(err, errLocalImageOpenTimeout) {
		t.Fatalf("expected timeout error, got %v", err)
	}

	select {
	case <-started:
	default:
		t.Fatal("expected opener to start")
	}

	_, _, err = openLocalImageFileWith(context.Background(), "other.jpg", time.Second, limiter, os.Open)
	if !errors.Is(err, errLocalImageOpenBusy) {
		t.Fatalf("expected busy while timed-out opener is still blocked, got %v", err)
	}

	close(release)
	waitForLimiterRelease(t, limiter)
}

func TestOpenLocalImageFileWithRejectsWhenLimiterFull(t *testing.T) {
	limiter := make(chan struct{}, 1)
	limiter <- struct{}{}

	_, _, err := openLocalImageFileWith(context.Background(), "thumb.jpg", time.Second, limiter, os.Open)
	if !errors.Is(err, errLocalImageOpenBusy) {
		t.Fatalf("expected busy error, got %v", err)
	}
}

func waitForLimiterRelease(t *testing.T, limiter chan struct{}) {
	t.Helper()

	deadline := time.After(time.Second)
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for limiter release")
		case <-ticker.C:
			if len(limiter) == 0 {
				return
			}
		}
	}
}
