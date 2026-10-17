package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/cloud"
	"video-server/internal/models"
)

type fakeBatchCloudDownloadService struct {
	result models.CloudDownloadBatchResult
	err    error
	links  []string
}

func (f *fakeBatchCloudDownloadService) Submit(context.Context, *uuid.UUID, string, *uuid.UUID, string) (models.CloudDownloadTask, bool, error) {
	return models.CloudDownloadTask{}, false, nil
}

func (f *fakeBatchCloudDownloadService) SubmitBatch(_ context.Context, links []string, _ *uuid.UUID) (models.CloudDownloadBatchResult, error) {
	f.links = append([]string(nil), links...)
	return f.result, f.err
}

func (f *fakeBatchCloudDownloadService) Retry(context.Context, int64) (models.CloudDownloadTask, error) {
	return models.CloudDownloadTask{}, nil
}

func (f *fakeBatchCloudDownloadService) List(context.Context, string, int, int) ([]models.CloudDownloadTask, int, error) {
	return nil, 0, nil
}

func (f *fakeBatchCloudDownloadService) StatusByHashes(context.Context, []string) (map[string]models.CloudDownloadStatus, error) {
	return map[string]models.CloudDownloadStatus{}, nil
}

func (f *fakeBatchCloudDownloadService) BrowserState(context.Context) cloud.State {
	return cloud.State{}
}

func (f *fakeBatchCloudDownloadService) LoginQR(context.Context) (string, error) {
	return "", nil
}

func newBatchTestRouter(svc cloudDownloadService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	api := &API{cloudDownload: svc}
	router := gin.New()
	router.POST("/batch", api.AdminCloudDownloadBatchCreate)
	return router
}

func postBatch(t *testing.T, router *gin.Engine, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAdminCloudDownloadBatchCreateKeepsEmptyEntries(t *testing.T) {
	fake := &fakeBatchCloudDownloadService{
		result: models.CloudDownloadBatchResult{
			Items: []models.CloudDownloadBatchItem{
				{Link: "", Status: models.CloudBatchInvalid},
				{Link: "ed2k://|file|a.mkv|1|0123456789abcdef0123456789abcdef|/", Status: models.CloudBatchSubmitted},
			},
			Summary: models.CloudDownloadBatchSummary{Total: 2, Submitted: 1, Invalid: 1},
		},
	}
	router := newBatchTestRouter(fake)

	rec := postBatch(t, router, `{"links":["","ed2k://|file|a.mkv|1|0123456789abcdef0123456789abcdef|/"]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("http status = %d, want 200", rec.Code)
	}
	var envelope struct {
		Code int                             `json:"code"`
		Data models.CloudDownloadBatchResult `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Code != 0 {
		t.Fatalf("code = %d, msg = %s", envelope.Code, rec.Body.String())
	}
	if len(fake.links) != 2 || fake.links[0] != "" {
		t.Fatalf("service links = %#v, want the empty entry preserved", fake.links)
	}
	if envelope.Data.Summary.Total != 2 || envelope.Data.Summary.Invalid != 1 {
		t.Fatalf("summary = %+v", envelope.Data.Summary)
	}
}

func TestAdminCloudDownloadBatchCreateAcceptsFifty(t *testing.T) {
	fake := &fakeBatchCloudDownloadService{}
	router := newBatchTestRouter(fake)
	links := make([]string, models.CloudDownloadBatchLimit)
	for index := range links {
		links[index] = "ed2k://|file|a.mkv|1|0123456789abcdef0123456789abcdef|/"
	}
	body, err := json.Marshal(map[string]any{"links": links})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	rec := postBatch(t, router, string(body))
	var envelope struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Code != 0 {
		t.Fatalf("code = %d, want 0", envelope.Code)
	}
	if len(fake.links) != models.CloudDownloadBatchLimit {
		t.Fatalf("service links = %d, want %d", len(fake.links), models.CloudDownloadBatchLimit)
	}
}

func TestAdminCloudDownloadBatchCreateReturns503WhenDisabled(t *testing.T) {
	router := newBatchTestRouter(nil)

	rec := postBatch(t, router, `{"links":["ed2k://|file|a.mkv|1|0123456789abcdef0123456789abcdef|/"]}`)
	var envelope struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Code != 503 {
		t.Fatalf("code = %d, want 503", envelope.Code)
	}
}

func TestAdminCloudDownloadBatchCreateRejectsEmptyAndOversizedRequests(t *testing.T) {
	fake := &fakeBatchCloudDownloadService{}
	router := newBatchTestRouter(fake)

	rec := postBatch(t, router, `{"links":[]}`)
	var envelope struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode empty response: %v", err)
	}
	if envelope.Code != 400 {
		t.Fatalf("empty links code = %d, want 400", envelope.Code)
	}
	if len(fake.links) != 0 {
		t.Fatalf("service called for empty links: %#v", fake.links)
	}

	links := make([]string, models.CloudDownloadBatchLimit+1)
	body, err := json.Marshal(map[string]any{"links": links})
	if err != nil {
		t.Fatalf("marshal oversized request: %v", err)
	}
	rec = postBatch(t, router, string(body))
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode oversized response: %v", err)
	}
	if envelope.Code != 400 {
		t.Fatalf("oversized links code = %d, want 400", envelope.Code)
	}
	if len(fake.links) != 0 {
		t.Fatalf("service called for oversized links: %#v", fake.links)
	}
}

func TestAdminCloudDownloadBatchCreateMapsServiceError(t *testing.T) {
	fake := &fakeBatchCloudDownloadService{err: errors.New("database down")}
	router := newBatchTestRouter(fake)

	rec := postBatch(t, router, `{"links":["ed2k://|file|a.mkv|1|0123456789abcdef0123456789abcdef|/"]}`)
	var envelope struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Code != 500 {
		t.Fatalf("code = %d, want 500", envelope.Code)
	}
}
