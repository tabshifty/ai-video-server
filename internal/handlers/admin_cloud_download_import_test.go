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

	"video-server/internal/models"
)

type fakeCloudImportService struct {
	task         models.CloudDownloadTask
	taskErr      error
	preview      models.CloudDownloadImportPreview
	previewErr   error
	created      models.CloudDownloadImport
	createErr    error
	gotTaskID    int64
	gotInput     models.CloudDownloadImportInput
	gotCreatedBy *uuid.UUID
}

func (f *fakeCloudImportService) GetImportTask(_ context.Context, id int64) (models.CloudDownloadTask, error) {
	f.gotTaskID = id
	return f.task, f.taskErr
}

func (f *fakeCloudImportService) Preview(context.Context, int64) (models.CloudDownloadImportPreview, error) {
	return f.preview, f.previewErr
}

func (f *fakeCloudImportService) Create(_ context.Context, taskID int64, input models.CloudDownloadImportInput, createdBy *uuid.UUID) (models.CloudDownloadImport, error) {
	f.gotTaskID = taskID
	f.gotInput = input
	f.gotCreatedBy = createdBy
	return f.created, f.createErr
}

func newCloudImportTestRouter(importSvc cloudDownloadImportService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	api := &API{cloudDownload: &fakeBatchCloudDownloadService{}, cloudDownloadImport: importSvc}
	router := gin.New()
	router.GET("/cloud-downloads/:id/import", api.AdminCloudDownloadImportTask)
	router.POST("/cloud-downloads/:id/import/preview", api.AdminCloudDownloadImportPreview)
	router.POST("/cloud-downloads/:id/import", api.AdminCloudDownloadImport)
	return router
}

func doCloudImportRequest(t *testing.T, router *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeCloudImportEnvelope(t *testing.T, rec *httptest.ResponseRecorder) struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
} {
	t.Helper()
	var envelope struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rec.Body.String())
	}
	return envelope
}

func TestAdminCloudDownloadImportPreviewReturnsFileShape(t *testing.T) {
	t.Parallel()

	fake := &fakeCloudImportService{preview: models.CloudDownloadImportPreview{
		Kind: models.CloudImportKindVideo,
		Name: "demo.mkv",
		Size: 2048,
		Mime: "video/x-matroska",
	}}
	rec := doCloudImportRequest(t, newCloudImportTestRouter(fake), "/cloud-downloads/7/import/preview", `{}`)
	envelope := decodeCloudImportEnvelope(t, rec)
	if envelope.Code != 0 {
		t.Fatalf("code = %d, body = %s", envelope.Code, rec.Body.String())
	}
	var preview models.CloudDownloadImportPreview
	if err := json.Unmarshal(envelope.Data, &preview); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if preview.Kind != models.CloudImportKindVideo || preview.Name != "demo.mkv" || preview.Size != 2048 || preview.Mime != "video/x-matroska" {
		t.Fatalf("preview = %+v", preview)
	}
}

func TestAdminCloudDownloadImportPassesVideoFields(t *testing.T) {
	t.Parallel()

	videoID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	collectionID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	actorID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	fake := &fakeCloudImportService{created: models.CloudDownloadImport{
		ID:      9,
		TaskID:  7,
		Kind:    models.CloudImportKindVideo,
		Status:  models.CloudImportQueued,
		VideoID: &videoID,
	}}
	body := `{
		"kind":"video",
		"type":"movie",
		"site_category":"western",
		"title":"标题",
		"description":"描述",
		"tags":["a"],
		"actor_ids":["33333333-3333-4333-8333-333333333333"],
		"actor_names":["演员"],
		"collection_ids":["22222222-2222-4222-8222-222222222222"],
		"image_collection_id":"22222222-2222-4222-8222-222222222222"
	}`
	rec := doCloudImportRequest(t, newCloudImportTestRouter(fake), "/cloud-downloads/7/import", body)
	envelope := decodeCloudImportEnvelope(t, rec)
	if envelope.Code != 0 {
		t.Fatalf("code = %d, body = %s", envelope.Code, rec.Body.String())
	}
	if fake.gotTaskID != 7 {
		t.Fatalf("task id = %d, want 7", fake.gotTaskID)
	}
	if fake.gotInput.Type != "movie" || fake.gotInput.Title != "标题" || fake.gotInput.SiteCategory != "western" {
		t.Fatalf("input = %+v", fake.gotInput)
	}
	if len(fake.gotInput.ActorIDs) != 1 || fake.gotInput.ActorIDs[0] != actorID {
		t.Fatalf("actor ids = %v", fake.gotInput.ActorIDs)
	}
	if len(fake.gotInput.CollectionIDs) != 1 || fake.gotInput.CollectionIDs[0] != collectionID {
		t.Fatalf("collection ids = %v", fake.gotInput.CollectionIDs)
	}
	if fake.gotInput.ImageCollectionID == nil || *fake.gotInput.ImageCollectionID != collectionID {
		t.Fatalf("image collection = %v", fake.gotInput.ImageCollectionID)
	}
	var data struct {
		Import models.CloudDownloadImport `json:"import"`
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data.Import.ID != 9 || data.Import.Status != models.CloudImportQueued {
		t.Fatalf("import = %+v", data.Import)
	}
}

func TestAdminCloudDownloadImportHandlesEmptyImageCollectionID(t *testing.T) {
	t.Parallel()

	fake := &fakeCloudImportService{created: models.CloudDownloadImport{ID: 11, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportQueued}}
	router := newCloudImportTestRouter(fake)

	rec := doCloudImportRequest(t, router, "/cloud-downloads/7/import", `{"kind":"video","type":"short","image_collection_id":""}`)
	if code := decodeCloudImportEnvelope(t, rec).Code; code != 0 {
		t.Fatalf("empty image collection code = %d, body = %s", code, rec.Body.String())
	}
	if fake.gotInput.ImageCollectionID != nil {
		t.Fatalf("image collection = %v, want nil", fake.gotInput.ImageCollectionID)
	}

	rec = doCloudImportRequest(t, router, "/cloud-downloads/7/import", `{"kind":"video","type":"short","image_collection_id":"not-a-uuid"}`)
	if code := decodeCloudImportEnvelope(t, rec).Code; code != 400 {
		t.Fatalf("invalid image collection code = %d, want 400", code)
	}
}

func TestAdminCloudDownloadImportPassesArchiveFields(t *testing.T) {
	t.Parallel()

	fake := &fakeCloudImportService{created: models.CloudDownloadImport{
		ID:     10,
		TaskID: 8,
		Kind:   models.CloudImportKindArchive,
		Status: models.CloudImportQueued,
	}}
	body := `{
		"kind":"archive",
		"title":"压缩包",
		"default_description":"说明",
		"default_tags":["tag"],
		"default_video_collection_ids":["22222222-2222-4222-8222-222222222222"],
		"default_image_collection_ids":["33333333-3333-4333-8333-333333333333"],
		"has_password":true,
		"password":"secret"
	}`
	rec := doCloudImportRequest(t, newCloudImportTestRouter(fake), "/cloud-downloads/8/import", body)
	envelope := decodeCloudImportEnvelope(t, rec)
	if envelope.Code != 0 {
		t.Fatalf("code = %d, body = %s", envelope.Code, rec.Body.String())
	}
	if fake.gotInput.Kind != models.CloudImportKindArchive || fake.gotInput.Title != "压缩包" {
		t.Fatalf("input = %+v", fake.gotInput)
	}
	if !fake.gotInput.HasPassword || fake.gotInput.Password != "secret" {
		t.Fatalf("password = %+v", fake.gotInput)
	}
	if len(fake.gotInput.DefaultVideoCollectionIDs) != 1 || len(fake.gotInput.DefaultImageCollectionIDs) != 1 {
		t.Fatalf("default collections = %v / %v", fake.gotInput.DefaultVideoCollectionIDs, fake.gotInput.DefaultImageCollectionIDs)
	}
}

func TestAdminCloudDownloadImportMapsErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want int
	}{
		{name: "not found", err: models.ErrCloudImportNotFound, want: 404},
		{name: "not completed", err: models.ErrCloudImportNotCompleted, want: 409},
		{name: "file missing", err: models.ErrCloudImportFileMissing, want: 410},
		{name: "unsupported", err: models.ErrCloudImportUnsupported, want: 422},
		{name: "invalid input", err: models.ErrCloudImportInvalidInput, want: 400},
		{name: "internal", err: errors.New("boom"), want: 500},
	}
	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCloudImportService{previewErr: tt.err, createErr: tt.err}
			router := newCloudImportTestRouter(fake)

			previewRec := doCloudImportRequest(t, router, "/cloud-downloads/7/import/preview", `{}`)
			if got := decodeCloudImportEnvelope(t, previewRec).Code; got != tt.want {
				t.Fatalf("preview code = %d, want %d (body=%s)", got, tt.want, previewRec.Body.String())
			}

			createRec := doCloudImportRequest(t, router, "/cloud-downloads/7/import", `{"kind":"video","type":"short"}`)
			if got := decodeCloudImportEnvelope(t, createRec).Code; got != tt.want {
				t.Fatalf("create code = %d, want %d (body=%s)", got, tt.want, createRec.Body.String())
			}
		})
	}
}

func TestAdminCloudDownloadImportRejectsInvalidJSONAndID(t *testing.T) {
	t.Parallel()

	fake := &fakeCloudImportService{}
	router := newCloudImportTestRouter(fake)

	badID := doCloudImportRequest(t, router, "/cloud-downloads/not-a-number/import", `{"kind":"video","type":"short"}`)
	if got := decodeCloudImportEnvelope(t, badID).Code; got != 400 {
		t.Fatalf("bad id code = %d, want 400", got)
	}
	badJSON := doCloudImportRequest(t, router, "/cloud-downloads/7/import", `{`)
	if got := decodeCloudImportEnvelope(t, badJSON).Code; got != 400 {
		t.Fatalf("bad json code = %d, want 400", got)
	}
	missingKind := doCloudImportRequest(t, router, "/cloud-downloads/7/import", `{}`)
	if got := decodeCloudImportEnvelope(t, missingKind).Code; got != 400 {
		t.Fatalf("missing kind code = %d, want 400", got)
	}
}

func TestCloudDownloadImportRoutesAreRegistered(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	api := &API{
		cloudDownload:       &fakeBatchCloudDownloadService{},
		cloudDownloadImport: &fakeCloudImportService{},
	}
	router := gin.New()
	api.Register(router)

	got := map[string]bool{}
	for _, route := range router.Routes() {
		got[route.Method+" "+route.Path] = true
	}
	for _, want := range []string{
		"GET /api/v1/admin/cloud-downloads/:id/import",
		"POST /api/v1/admin/cloud-downloads/:id/import/preview",
		"POST /api/v1/admin/cloud-downloads/:id/import",
	} {
		if !got[want] {
			t.Fatalf("route %s is not registered; routes=%v", want, got)
		}
	}
}

func TestAdminCloudDownloadImportReturns503WhenDisabled(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	api := &API{cloudDownload: &fakeBatchCloudDownloadService{}}
	router := gin.New()
	router.GET("/cloud-downloads/:id/import", api.AdminCloudDownloadImportTask)
	router.POST("/cloud-downloads/:id/import/preview", api.AdminCloudDownloadImportPreview)
	router.POST("/cloud-downloads/:id/import", api.AdminCloudDownloadImport)

	rec := doCloudImportRequest(t, router, "/cloud-downloads/7/import", `{"kind":"video","type":"short"}`)
	if got := decodeCloudImportEnvelope(t, rec).Code; got != 503 {
		t.Fatalf("code = %d, want 503 (body=%s)", got, rec.Body.String())
	}
}

func TestAdminCloudDownloadImportTaskRestoresAndHandlesErrors(t *testing.T) {
	for _, tt := range []struct {
		name, path string
		err        error
		code       int
	}{
		{name: "成功", path: "7", code: 0}, {name: "无效 ID", path: "invalid", code: 400}, {name: "非正 ID", path: "0", code: 400}, {name: "不存在", path: "7", err: models.ErrCloudDownloadNotFound, code: 404}, {name: "查询失败", path: "7", err: errors.New("查询失败"), code: 500},
	} {
		t.Run(tt.name, func(t *testing.T) {
			batchID := uuid.New()
			f := &fakeCloudImportService{taskErr: tt.err, task: models.CloudDownloadTask{ID: 7, Import: &models.CloudDownloadImportSummary{Kind: "archive", Status: "imported", ArchiveBatchID: &batchID}}}
			r := newCloudImportTestRouter(f)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/cloud-downloads/"+tt.path+"/import", nil))
			envelope := decodeCloudImportEnvelope(t, rec)
			if envelope.Code != tt.code {
				t.Fatalf("code=%d, body=%s", envelope.Code, rec.Body.String())
			}
			if tt.code == 0 {
				var task models.CloudDownloadTask
				if err := json.Unmarshal(envelope.Data, &task); err != nil {
					t.Fatal(err)
				}
				if task.Import == nil || task.Import.ArchiveBatchID == nil || *task.Import.ArchiveBatchID != batchID {
					t.Fatal("批次标识丢失")
				}
			}
		})
	}
}
