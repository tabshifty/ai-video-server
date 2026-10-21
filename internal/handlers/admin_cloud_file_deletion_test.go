package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/models"
)

type fakeCloudFileDeletionService struct {
	task    models.CloudDownloadTask
	err     error
	calls   int
	gotID   int64
	gotUser *uuid.UUID
}

func (f *fakeCloudFileDeletionService) DeleteCloudFile(_ context.Context, taskID int64, requestedBy *uuid.UUID) (models.CloudDownloadTask, error) {
	f.calls++
	f.gotID = taskID
	f.gotUser = requestedBy
	return f.task, f.err
}

func newCloudFileDeletionTestRouter(svc cloudFileDeletionService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	api := &API{cloudDownload: &fakeBatchCloudDownloadService{}, cloudFileDeletion: svc}
	router := gin.New()
	router.POST("/cloud-downloads/:id/cloud-file/delete", api.AdminCloudDownloadDeleteCloudFile)
	return router
}

func doCloudFileDeletionRequest(t *testing.T, router *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAdminCloudDeleteFileRequiresConfirmation(t *testing.T) {
	t.Parallel()

	fake := &fakeCloudFileDeletionService{}
	router := newCloudFileDeletionTestRouter(fake)
	rec := doCloudFileDeletionRequest(t, router, "/cloud-downloads/7/cloud-file/delete", `{}`)
	if got := decodeCloudImportEnvelope(t, rec).Code; got != 400 {
		t.Fatalf("code = %d, want 400; body=%s", got, rec.Body.String())
	}
	if fake.calls != 0 {
		t.Fatalf("service calls = %d, want 0", fake.calls)
	}
}

func TestAdminCloudDeleteFileRejectsInvalidID(t *testing.T) {
	t.Parallel()

	fake := &fakeCloudFileDeletionService{}
	router := newCloudFileDeletionTestRouter(fake)
	rec := doCloudFileDeletionRequest(t, router, "/cloud-downloads/not-a-number/cloud-file/delete", `{"confirm":true}`)
	if got := decodeCloudImportEnvelope(t, rec).Code; got != 400 {
		t.Fatalf("code = %d, want 400", got)
	}
	if fake.calls != 0 {
		t.Fatalf("service calls = %d, want 0", fake.calls)
	}
}

func TestAdminCloudDeleteFileReturnsTask(t *testing.T) {
	t.Parallel()

	fake := &fakeCloudFileDeletionService{task: models.CloudDownloadTask{
		ID:     7,
		Status: models.CloudDownloadCompleted,
		FileDeletion: &models.CloudDownloadFileDeletionSummary{
			Status: models.CloudFileDeleted,
		},
	}}
	router := newCloudFileDeletionTestRouter(fake)
	rec := doCloudFileDeletionRequest(t, router, "/cloud-downloads/7/cloud-file/delete", `{"confirm":true}`)
	envelope := decodeCloudImportEnvelope(t, rec)
	if envelope.Code != 0 {
		t.Fatalf("code = %d, want 0; body=%s", envelope.Code, rec.Body.String())
	}
	if fake.calls != 1 || fake.gotID != 7 {
		t.Fatalf("calls=%d taskID=%d", fake.calls, fake.gotID)
	}
}

func TestAdminCloudDeleteFileMapsErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "not found", err: models.ErrCloudFileDeleteNotFound, want: 404},
		{name: "not completed", err: models.ErrCloudFileDeleteNotCompleted, want: 409},
		{name: "in progress", err: models.ErrCloudFileDeleteInProgress, want: 409},
		{name: "unavailable", err: models.ErrCloudFileDeleteUnavailable, want: 409},
		{name: "local missing", err: models.ErrCloudFileDeleteLocalMissing, want: 410},
		{name: "browser down", err: models.ErrCloudFileDeleteBrowserDown, want: 503},
		{name: "login required", err: models.ErrCloudFileDeleteLoginRequired, want: 503},
		{name: "timeout", err: models.ErrCloudFileDeleteTimeout, want: 504},
		{name: "115 rejected", err: models.ErrCloudFileDeleteFailed, want: 502},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeCloudFileDeletionService{err: tc.err}
			router := newCloudFileDeletionTestRouter(fake)
			rec := doCloudFileDeletionRequest(t, router, "/cloud-downloads/7/cloud-file/delete", `{"confirm":true}`)
			if got := decodeCloudImportEnvelope(t, rec).Code; got != tc.want {
				t.Fatalf("code = %d, want %d; body=%s", got, tc.want, rec.Body.String())
			}
		})
	}
}

func TestCloudFileDeletionRouteIsRegistered(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	api := &API{
		cloudDownload:     &fakeBatchCloudDownloadService{},
		cloudFileDeletion: &fakeCloudFileDeletionService{},
	}
	router := gin.New()
	api.Register(router)
	for _, route := range router.Routes() {
		if route.Method == http.MethodPost && route.Path == "/api/v1/admin/cloud-downloads/:id/cloud-file/delete" {
			return
		}
	}
	t.Fatalf("cloud file deletion route is not registered")
}

func TestAdminCloudDeleteFileReturns503WhenDisabled(t *testing.T) {
	t.Parallel()

	api := &API{cloudDownload: &fakeBatchCloudDownloadService{}}
	router := gin.New()
	router.POST("/cloud-downloads/:id/cloud-file/delete", api.AdminCloudDownloadDeleteCloudFile)
	rec := doCloudFileDeletionRequest(t, router, "/cloud-downloads/7/cloud-file/delete", `{"confirm":true}`)
	if got := decodeCloudImportEnvelope(t, rec).Code; got != 503 {
		t.Fatalf("code = %d, want 503", got)
	}
}
