package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"log/slog"

	"video-server/internal/models"
	"video-server/internal/queue"
)

type orphanFileScanRepoStub struct {
	beginPrev      models.AdminOrphanFileScan
	beginErr       error
	restorePrev    models.AdminOrphanFileScan
	restoreErr     error
	latest         models.AdminOrphanFileScan
	latestErr      error
	deletedCount   int64
	beginCalls     int
	restoreCalls   int
	latestCalls    int
	deleteCalls    int
	markDeletedErr error
}

func (s *orphanFileScanRepoStub) GetOrphanFileScan(context.Context) (models.AdminOrphanFileScan, error) {
	s.latestCalls++
	return s.latest, s.latestErr
}

func (s *orphanFileScanRepoStub) BeginOrphanFileScan(context.Context) (models.AdminOrphanFileScan, error) {
	s.beginCalls++
	return s.beginPrev, s.beginErr
}

func (s *orphanFileScanRepoStub) RestoreOrphanFileScan(_ context.Context, prev models.AdminOrphanFileScan) error {
	s.restoreCalls++
	s.restorePrev = prev
	return s.restoreErr
}

func (s *orphanFileScanRepoStub) MarkOrphanFileScanDeleted(_ context.Context, deletedCount int64) error {
	s.deleteCalls++
	s.deletedCount = deletedCount
	return s.markDeletedErr
}

type orphanScanEnqueuerStub struct {
	orphanScanCalls int
	orphanScanErr   error
}

func (s *orphanScanEnqueuerStub) EnqueueTranscode(queue.TranscodePayload) error {
	return nil
}

func (s *orphanScanEnqueuerStub) EnqueueScrapeMovie(queue.ScrapePayload) error {
	return nil
}

func (s *orphanScanEnqueuerStub) EnqueueScrapeTV(queue.ScrapePayload) error {
	return nil
}

func (s *orphanScanEnqueuerStub) EnqueueScrapeAV(queue.ScrapePayload) error {
	return nil
}

func (s *orphanScanEnqueuerStub) EnqueueScrapeRetag(queue.RetagScrapePayload) error {
	return nil
}

func (s *orphanScanEnqueuerStub) EnqueueOrphanFileScan() error {
	s.orphanScanCalls++
	return s.orphanScanErr
}

func TestAdminStartOrphanFileScanQueuesTaskAndReturnsPending(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	prev := models.AdminOrphanFileScan{
		ID:     1,
		Status: "completed",
	}
	repo := &orphanFileScanRepoStub{beginPrev: prev}
	enqueuer := &orphanScanEnqueuerStub{}
	api := &API{
		orphanFileScanRepo: repo,
		enqueuer:           enqueuer,
		logger:             slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/system/orphan-files/scan", nil)

	api.AdminStartOrphanFileScan(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code=0, got=%d body=%s", resp.Code, rec.Body.String())
	}
	if resp.Data["status"] != "pending" {
		t.Fatalf("expected pending response, got=%v", resp.Data["status"])
	}
	if repo.beginCalls != 1 || enqueuer.orphanScanCalls != 1 || repo.restoreCalls != 0 {
		t.Fatalf("unexpected call counts: begin=%d enqueue=%d restore=%d", repo.beginCalls, enqueuer.orphanScanCalls, repo.restoreCalls)
	}
}

func TestAdminStartOrphanFileScanRestoresPreviousStateOnQueueFailure(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	prev := models.AdminOrphanFileScan{
		ID:     1,
		Status: "completed",
	}
	repo := &orphanFileScanRepoStub{beginPrev: prev}
	enqueuer := &orphanScanEnqueuerStub{orphanScanErr: context.DeadlineExceeded}
	api := &API{
		orphanFileScanRepo: repo,
		enqueuer:           enqueuer,
		logger:             slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/system/orphan-files/scan", nil)

	api.AdminStartOrphanFileScan(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 1069 {
		t.Fatalf("expected code=1069, got=%d body=%s", resp.Code, rec.Body.String())
	}
	if repo.restoreCalls != 1 {
		t.Fatalf("expected restore call after enqueue failure, got %d", repo.restoreCalls)
	}
	if repo.restorePrev.Status != prev.Status {
		t.Fatalf("expected restore previous state to be preserved, got=%+v", repo.restorePrev)
	}
}

func TestAdminLatestOrphanFileScanReturnsLatestSnapshot(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	repo := &orphanFileScanRepoStub{
		latest: models.AdminOrphanFileScan{
			ID:          1,
			Status:      "completed",
			OrphanFiles: 2,
			Items: []models.AdminOrphanFileScanItem{
				{ID: 11, FilePath: "/storage/images/a.jpg", RelativePath: "images/a.jpg"},
			},
		},
	}
	api := &API{orphanFileScanRepo: repo}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/system/orphan-files/latest", nil)

	api.AdminLatestOrphanFileScan(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code=0, got=%d body=%s", resp.Code, rec.Body.String())
	}
	if resp.Data["status"] != "completed" {
		t.Fatalf("expected completed response, got=%v", resp.Data["status"])
	}
	if repo.latestCalls != 1 {
		t.Fatalf("expected one latest call, got %d", repo.latestCalls)
	}
}

func TestAdminDeleteLatestOrphanFileScanDeletesAllFiles(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	storageRoot := t.TempDir()
	first := filepath.Join(storageRoot, "videos", "orphan-1.mp4")
	second := filepath.Join(storageRoot, "images", "orphan-2.jpg")
	if err := os.MkdirAll(filepath.Dir(first), 0o755); err != nil {
		t.Fatalf("mkdir first path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(second), 0o755); err != nil {
		t.Fatalf("mkdir second path: %v", err)
	}
	if err := os.WriteFile(first, []byte("first"), 0o644); err != nil {
		t.Fatalf("write first file: %v", err)
	}
	if err := os.WriteFile(second, []byte("second"), 0o644); err != nil {
		t.Fatalf("write second file: %v", err)
	}

	repo := &orphanFileScanRepoStub{
		latest: models.AdminOrphanFileScan{
			ID:          1,
			Status:      "completed",
			OrphanFiles: 2,
			Items: []models.AdminOrphanFileScanItem{
				{ID: 11, FilePath: first, RelativePath: "videos/orphan-1.mp4"},
				{ID: 12, FilePath: second, RelativePath: "images/orphan-2.jpg"},
			},
		},
	}
	api := &API{orphanFileScanRepo: repo}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/admin/system/orphan-files/latest", nil)

	api.AdminDeleteLatestOrphanFileScan(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code=0, got=%d body=%s", resp.Code, rec.Body.String())
	}
	if resp.Data["status"] != "deleted" {
		t.Fatalf("expected deleted response, got=%v", resp.Data["status"])
	}
	if repo.deleteCalls != 1 || repo.deletedCount != 2 {
		t.Fatalf("unexpected delete calls/count: calls=%d deleted=%d", repo.deleteCalls, repo.deletedCount)
	}
	if _, err := os.Stat(first); !os.IsNotExist(err) {
		t.Fatalf("expected first file to be removed, stat err=%v", err)
	}
	if _, err := os.Stat(second); !os.IsNotExist(err) {
		t.Fatalf("expected second file to be removed, stat err=%v", err)
	}
}

func TestAdminDeleteLatestOrphanFileScanRejectsIncompleteScan(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	repo := &orphanFileScanRepoStub{
		latest: models.AdminOrphanFileScan{
			ID:     1,
			Status: "running",
		},
	}
	api := &API{orphanFileScanRepo: repo}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/admin/system/orphan-files/latest", nil)

	api.AdminDeleteLatestOrphanFileScan(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 1070 {
		t.Fatalf("expected code=1070, got=%d body=%s", resp.Code, rec.Body.String())
	}
	if repo.deleteCalls != 0 {
		t.Fatalf("expected no delete call, got %d", repo.deleteCalls)
	}
}
