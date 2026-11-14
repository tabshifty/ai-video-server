package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/models"
)

type fakeCloudLocalFileService struct {
	preview    models.CloudLocalFilePreview
	task       models.CloudDownloadTask
	err        error
	gotTaskID  int64
	gotActor   *uuid.UUID
	deleteHits int
}

func (f *fakeCloudLocalFileService) Preview(_ context.Context, taskID int64) (models.CloudLocalFilePreview, error) {
	f.gotTaskID = taskID
	return f.preview, f.err
}

func (f *fakeCloudLocalFileService) Delete(_ context.Context, taskID int64, requestedBy *uuid.UUID) (models.CloudDownloadTask, error) {
	f.gotTaskID = taskID
	f.gotActor = requestedBy
	f.deleteHits++
	return f.task, f.err
}

func newCloudLocalFileTestRouter(svc cloudLocalFileDeletionService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	api := &API{cloudDownload: &fakeBatchCloudDownloadService{}, cloudLocalFileDeletion: svc}
	router := gin.New()
	router.GET("/cloud-downloads/:id/local-file", api.AdminCloudDownloadLocalFilePreview)
	router.POST("/cloud-downloads/:id/local-file/delete", api.AdminCloudDownloadDeleteLocalFile)
	return router
}

func TestAdminCloudDownloadLocalFilePreviewReturnsFileShape(t *testing.T) {
	t.Parallel()

	fake := &fakeCloudLocalFileService{preview: models.CloudLocalFilePreview{
		Path:   "/downloads/demo.rar",
		Size:   128,
		Exists: true,
	}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cloud-downloads/7/local-file", nil)
	newCloudLocalFileTestRouter(fake).ServeHTTP(rec, req)

	envelope := decodeCloudImportEnvelope(t, rec)
	if envelope.Code != 0 {
		t.Fatalf("code = %d body = %s", envelope.Code, rec.Body.String())
	}
	var preview models.CloudLocalFilePreview
	if err := json.Unmarshal(envelope.Data, &preview); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if preview.Path != "/downloads/demo.rar" || preview.Size != 128 || !preview.Exists {
		t.Fatalf("preview = %+v", preview)
	}
	if fake.gotTaskID != 7 {
		t.Fatalf("task id = %d, want 7", fake.gotTaskID)
	}
}

func TestAdminCloudDownloadDeleteLocalFileReturnsTask(t *testing.T) {
	t.Parallel()

	deletedAt := time.Now().UTC()
	fake := &fakeCloudLocalFileService{task: models.CloudDownloadTask{
		ID:                 7,
		Status:             models.CloudDownloadCompleted,
		LocalFileDeletedAt: &deletedAt,
	}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cloud-downloads/7/local-file/delete", strings.NewReader(`{"confirm":true}`))
	req.Header.Set("Content-Type", "application/json")
	newCloudLocalFileTestRouter(fake).ServeHTTP(rec, req)

	envelope := decodeCloudImportEnvelope(t, rec)
	if envelope.Code != 0 {
		t.Fatalf("code = %d body = %s", envelope.Code, rec.Body.String())
	}
	var payload struct {
		Task models.CloudDownloadTask `json:"task"`
	}
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		t.Fatalf("decode task: %v", err)
	}
	if payload.Task.LocalFileDeletedAt == nil || payload.Task.Status != models.CloudDownloadCompleted {
		t.Fatalf("task = %+v", payload.Task)
	}
	if fake.deleteHits != 1 || fake.gotTaskID != 7 {
		t.Fatalf("delete calls = %d id = %d", fake.deleteHits, fake.gotTaskID)
	}
}

func TestAdminCloudDownloadDeleteLocalFileRequiresExplicitConfirm(t *testing.T) {
	t.Parallel()

	fake := &fakeCloudLocalFileService{task: models.CloudDownloadTask{ID: 7, Status: models.CloudDownloadCompleted}}
	router := newCloudLocalFileTestRouter(fake)

	for name, body := range map[string]string{"empty": "", "not confirmed": `{"confirm":false}`} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/cloud-downloads/7/local-file/delete", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rec, req)
		if got := decodeCloudImportEnvelope(t, rec).Code; got != 400 {
			t.Fatalf("%s: code = %d, want 400", name, got)
		}
	}
	if fake.deleteHits != 0 {
		t.Fatalf("unconfirmed request must not reach the service, hits = %d", fake.deleteHits)
	}
}

func TestAdminCloudDownloadLocalFileMapsErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want int
	}{
		{name: "not found", err: models.ErrCloudLocalFileNotFound, want: 404},
		{name: "not completed", err: models.ErrCloudLocalFileNotCompleted, want: 409},
		{name: "import active", err: models.ErrCloudLocalFileImportActive, want: 409},
		{name: "unmanaged path", err: models.ErrCloudLocalFileUnavailable, want: 409},
		{name: "path not recorded", err: models.ErrCloudLocalFileNotRecorded, want: 409},
		{name: "internal", err: errors.New("boom"), want: 500},
	}
	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCloudLocalFileService{err: tt.err}
			router := newCloudLocalFileTestRouter(fake)

			previewRec := httptest.NewRecorder()
			router.ServeHTTP(previewRec, httptest.NewRequest(http.MethodGet, "/cloud-downloads/7/local-file", nil))
			if got := decodeCloudImportEnvelope(t, previewRec).Code; got != tt.want {
				t.Fatalf("preview code = %d, want %d body=%s", got, tt.want, previewRec.Body.String())
			}

			deleteRec := httptest.NewRecorder()
			deleteReq := httptest.NewRequest(http.MethodPost, "/cloud-downloads/7/local-file/delete", strings.NewReader(`{"confirm":true}`))
			deleteReq.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(deleteRec, deleteReq)
			if got := decodeCloudImportEnvelope(t, deleteRec).Code; got != tt.want {
				t.Fatalf("delete code = %d, want %d body=%s", got, tt.want, deleteRec.Body.String())
			}
		})
	}
}
