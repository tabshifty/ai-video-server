package handlers

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/queue"
)

func TestRegisterIncludesEd2kDownloadRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := &API{}
	api.Register(router)

	routes := map[string]struct{}{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}

	for _, want := range []string{
		"GET /api/v1/admin/ed2k-download/tasks",
		"POST /api/v1/admin/ed2k-download/tasks",
		"GET /api/v1/admin/ed2k-download/tasks/:id",
		"POST /api/v1/admin/ed2k-download/tasks/:id/retry",
		"DELETE /api/v1/admin/ed2k-download/tasks/:id",
	} {
		if _, ok := routes[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
}

func TestParseEd2kDownloadLinkAndBuildTitle(t *testing.T) {
	sourceLink, title, filename, hash, size, ok := parseEd2kDownloadLink("ed2k://|file|%E4%B8%BB%E8%A7%92.mkv|123|ABCDEF0123456789ABCDEF0123456789|/")
	if !ok {
		t.Fatal("expected ed2k link to parse")
	}
	if sourceLink != "ed2k://|file|%E4%B8%BB%E8%A7%92.mkv|123|ABCDEF0123456789ABCDEF0123456789|/" {
		t.Fatalf("unexpected source link: %s", sourceLink)
	}
	if title != "主角.mkv" || filename != "主角.mkv" {
		t.Fatalf("unexpected title or filename: %s %s", title, filename)
	}
	if hash != "ABCDEF0123456789ABCDEF0123456789" {
		t.Fatalf("unexpected hash: %s", hash)
	}
	if size != 123 {
		t.Fatalf("unexpected size: %d", size)
	}
	if got := buildEd2kDownloadTaskTitle("主角.mkv", "  "); got != "主角.mkv" {
		t.Fatalf("unexpected fallback title: %s", got)
	}
	if got := buildEd2kDownloadTaskTitle("主角.mkv", "  自定义 标题  "); got != "自定义 标题" {
		t.Fatalf("unexpected normalized title: %s", got)
	}
}

func TestRetryEd2kDownloadTaskRequeuesAfterStatusUpdate(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	task := models.AdminEd2kDownloadTask{
		ID:     taskID,
		Status: "failed",
	}
	repo := &ed2kRetryRepoStub{
		task: task,
		updated: models.AdminEd2kDownloadTask{
			ID:           taskID,
			Status:       "queued",
			ProgressText: "等待执行器接管",
			RetryCount:   1,
		},
	}
	enq := &ed2kRetryEnqueuerStub{}

	item, err := retryEd2kDownloadTask(context.Background(), taskID, repo, enq, time.Now())
	if err != nil {
		t.Fatalf("retryEd2kDownloadTask() error = %v", err)
	}
	if item.Status != "queued" {
		t.Fatalf("expected queued item, got %s", item.Status)
	}
	if !repo.updateCalled {
		t.Fatal("expected status update")
	}
	if len(enq.payloads) != 1 {
		t.Fatalf("expected one enqueue call, got %d", len(enq.payloads))
	}
	if enq.payloads[0].TaskID != taskID.String() {
		t.Fatalf("unexpected enqueue task id: %s", enq.payloads[0].TaskID)
	}
}

func TestRetryEd2kDownloadTaskReturnsEnqueueError(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo := &ed2kRetryRepoStub{
		task: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "failed",
		},
		updated: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "queued",
		},
	}
	enq := &ed2kRetryEnqueuerStub{err: errors.New("enqueue failed")}

	_, err := retryEd2kDownloadTask(context.Background(), taskID, repo, enq, time.Now())
	if err == nil {
		t.Fatal("expected enqueue error")
	}
	if !strings.Contains(err.Error(), "enqueue failed") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.updateCalled {
		t.Fatal("expected status update before enqueue")
	}
}

func TestRetryEd2kDownloadTaskReturnsInFlightConflict(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo := &ed2kRetryRepoStub{
		task: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "failed",
		},
		updated: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "queued",
		},
	}
	enq := &ed2kRetryEnqueuerStub{err: queue.ErrEd2kDownloadTaskInFlight}

	_, err := retryEd2kDownloadTask(context.Background(), taskID, repo, enq, time.Now())
	if !errors.Is(err, queue.ErrEd2kDownloadTaskInFlight) {
		t.Fatalf("expected in-flight conflict, got %v", err)
	}
}

type ed2kRetryRepoStub struct {
	task         models.AdminEd2kDownloadTask
	updated      models.AdminEd2kDownloadTask
	updateCalled bool
}

func (s *ed2kRetryRepoStub) GetEd2kDownloadTask(context.Context, uuid.UUID) (models.AdminEd2kDownloadTask, error) {
	return s.task, nil
}

func (s *ed2kRetryRepoStub) UpdateEd2kDownloadTaskStatus(ctx context.Context, id uuid.UUID, status, progressText, errorMessage string, startedAt, finishedAt, deletedAt *time.Time, outputDir, downloadedPath string, files []models.AdminEd2kDownloadTaskFile, retryDelta int, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	s.updateCalled = true
	if id != s.task.ID {
		return models.AdminEd2kDownloadTask{}, errors.New("unexpected task id")
	}
	if status != "queued" || progressText != "等待执行器接管" || retryDelta != 1 || history.Kind != "retry" {
		return models.AdminEd2kDownloadTask{}, errors.New("unexpected retry update")
	}
	return s.updated, nil
}

type ed2kRetryEnqueuerStub struct {
	payloads []queue.Ed2kDownloadPayload
	err      error
}

func (s *ed2kRetryEnqueuerStub) EnqueueEd2kDownload(payload queue.Ed2kDownloadPayload) error {
	s.payloads = append(s.payloads, payload)
	return s.err
}
