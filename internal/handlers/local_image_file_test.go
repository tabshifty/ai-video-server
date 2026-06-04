package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
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

func TestTryServeLocalImagePathServesFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	path := filepath.Join(dir, "thumb.jpg")
	if err := os.WriteFile(path, []byte("jpg"), 0o644); err != nil {
		t.Fatalf("write temp image: %v", err)
	}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/view", nil)

	if ok := tryServeLocalImagePath(ctx, path, "图片不存在", "图片暂时不可用"); !ok {
		t.Fatal("expected helper to serve local image")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != "jpg" {
		t.Fatalf("body = %q, want jpg", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=86400" {
		t.Fatalf("Cache-Control = %q, want public, max-age=86400", got)
	}
	if got := rec.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Fatalf("Content-Type = %q, want image/jpeg", got)
	}
}

func TestTryServeLocalImagePathReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/view", nil)

	if ok := tryServeLocalImagePath(ctx, filepath.Join(t.TempDir(), "missing.jpg"), "图片不存在", "图片暂时不可用"); ok {
		t.Fatal("expected missing image to be rejected")
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
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
