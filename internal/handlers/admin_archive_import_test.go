package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/services"
)

type archiveImportServiceStub struct {
	uploadBatch      models.ArchiveImportBatch
	uploadErr        error
	uploadCalls      int
	uploadInput      services.ArchiveImportUploadInput
	uploadHeader     *multipart.FileHeader
	retryBatch       models.ArchiveImportBatch
	retryErr         error
	retryCalls       int
	retryBatchID     uuid.UUID
	retryPassword    string
	retryMode        string
	batchUpdateInput services.ArchiveImportBatchUpdateInput
	batchUpdateFiles []models.ArchiveImportFileListItem
	batchUpdateErr   error
	batchUpdateCalls int
}

func (s *archiveImportServiceStub) ListBatches(context.Context, int, int) ([]models.ArchiveImportBatchListItem, int, error) {
	return nil, 0, nil
}

func (s *archiveImportServiceStub) GetBatchWithFiles(context.Context, uuid.UUID) (models.ArchiveImportBatch, []models.ArchiveImportFileListItem, error) {
	return models.ArchiveImportBatch{}, nil, nil
}

func (s *archiveImportServiceStub) ListGroups(context.Context, uuid.UUID) ([]models.ArchiveImportGroup, error) {
	return nil, nil
}

func (s *archiveImportServiceStub) GetFile(context.Context, uuid.UUID) (models.ArchiveImportFileListItem, error) {
	return models.ArchiveImportFileListItem{}, nil
}

func (s *archiveImportServiceStub) UploadArchive(_ context.Context, in services.ArchiveImportUploadInput, fileHeader *multipart.FileHeader) (models.ArchiveImportBatch, error) {
	s.uploadCalls++
	s.uploadInput = in
	s.uploadHeader = fileHeader
	return s.uploadBatch, s.uploadErr
}

func (s *archiveImportServiceStub) UpdateFile(context.Context, uuid.UUID, services.ArchiveImportFileUpdateInput) (models.ArchiveImportFileListItem, error) {
	return models.ArchiveImportFileListItem{}, nil
}

func (s *archiveImportServiceStub) BatchUpdateFiles(_ context.Context, in services.ArchiveImportBatchUpdateInput) ([]models.ArchiveImportFileListItem, error) {
	s.batchUpdateCalls++
	s.batchUpdateInput = in
	return s.batchUpdateFiles, s.batchUpdateErr
}

func (s *archiveImportServiceStub) CreateGroup(context.Context, uuid.UUID, services.ArchiveImportGroupCreateInput) (models.ArchiveImportGroup, error) {
	return models.ArchiveImportGroup{}, nil
}

func (s *archiveImportServiceStub) UpdateGroup(context.Context, uuid.UUID, services.ArchiveImportGroupUpdateInput) (models.ArchiveImportGroup, error) {
	return models.ArchiveImportGroup{}, nil
}

func (s *archiveImportServiceStub) DeleteGroup(context.Context, uuid.UUID) error {
	return nil
}

func (s *archiveImportServiceStub) AssignFilesToGroup(context.Context, uuid.UUID, []uuid.UUID) error {
	return nil
}

func (s *archiveImportServiceStub) RemoveFilesFromGroup(context.Context, uuid.UUID, []uuid.UUID) error {
	return nil
}

func (s *archiveImportServiceStub) ProcessGroup(context.Context, uuid.UUID) ([]models.ArchiveImportFileListItem, error) {
	return nil, nil
}

func (s *archiveImportServiceStub) ProcessFile(context.Context, uuid.UUID) (models.ArchiveImportFileListItem, error) {
	return models.ArchiveImportFileListItem{}, nil
}

func (s *archiveImportServiceStub) ProcessAllFiles(context.Context, uuid.UUID) ([]models.ArchiveImportFileListItem, error) {
	return nil, nil
}

func (s *archiveImportServiceStub) RetryExtract(_ context.Context, batchID uuid.UUID, password, requestedMode string) (models.ArchiveImportBatch, error) {
	s.retryCalls++
	s.retryBatchID = batchID
	s.retryPassword = password
	s.retryMode = requestedMode
	return s.retryBatch, s.retryErr
}

func (s *archiveImportServiceStub) DeleteBatch(context.Context, uuid.UUID) error {
	return nil
}

func TestAdminUploadArchiveImportReturnsRecoverableBatchForEncodingSelection(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	stub := &archiveImportServiceStub{
		uploadBatch: models.ArchiveImportBatch{
			ID:                    uuid.MustParse("11111111-1111-4111-8111-111111111111"),
			Status:                "needs_encoding",
			EncodingRequestedMode: "auto",
		},
		uploadErr: services.ErrArchiveEncodingRequired,
	}
	api := &API{archiveImportSvc: stub}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("title", "压缩包标题"); err != nil {
		t.Fatalf("WriteField() error = %v", err)
	}
	if err := writer.WriteField("has_password", "0"); err != nil {
		t.Fatalf("WriteField() error = %v", err)
	}
	part, err := writer.CreateFormFile("file", "中文.zip")
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write([]byte("demo")); err != nil {
		t.Fatalf("part.Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/archive-import/upload", &body)
	ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())
	ctx.Set("auth_user_id", uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
	ctx.Set("auth_role", "admin")

	api.AdminUploadArchiveImport(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 1077 {
		t.Fatalf("expected code=1077, got=%d body=%s", resp.Code, rec.Body.String())
	}
	if resp.Data["status"] != "needs_encoding" {
		t.Fatalf("expected needs_encoding batch, got=%v", resp.Data["status"])
	}
	if resp.Data["encoding_requested_mode"] != "auto" {
		t.Fatalf("expected requested mode auto, got=%v", resp.Data["encoding_requested_mode"])
	}
	if stub.uploadCalls != 1 {
		t.Fatalf("expected one upload call, got %d", stub.uploadCalls)
	}
}

func TestAdminRetryArchiveImportExtractPassesEncodingMode(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	stub := &archiveImportServiceStub{
		retryBatch: models.ArchiveImportBatch{
			ID:                    uuid.MustParse("22222222-2222-4222-8222-222222222222"),
			Status:                "completed",
			EncodingMode:          "gbk",
			EncodingRequestedMode: "gbk",
		},
	}
	api := &API{archiveImportSvc: stub}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/archive-import/batches/22222222-2222-4222-8222-222222222222/retry-extract", bytes.NewBufferString(`{"password":"secret","encoding_mode":"gbk"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: "22222222-2222-4222-8222-222222222222"}}
	ctx.Set("auth_user_id", uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
	ctx.Set("auth_role", "admin")

	api.AdminRetryArchiveImportExtract(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code=0, got=%d body=%s", resp.Code, rec.Body.String())
	}
	if stub.retryCalls != 1 {
		t.Fatalf("expected one retry call, got %d", stub.retryCalls)
	}
	if stub.retryPassword != "secret" {
		t.Fatalf("expected password to pass through, got %q", stub.retryPassword)
	}
	if stub.retryMode != "gbk" {
		t.Fatalf("expected encoding mode gbk, got %q", stub.retryMode)
	}
}

func TestAdminRetryArchiveImportExtractRejectsInvalidEncodingMode(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	stub := &archiveImportServiceStub{}
	api := &API{archiveImportSvc: stub}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/archive-import/batches/22222222-2222-4222-8222-222222222222/retry-extract", bytes.NewBufferString(`{"password":"secret","encoding_mode":"sjis"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: "22222222-2222-4222-8222-222222222222"}}
	ctx.Set("auth_user_id", uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
	ctx.Set("auth_role", "admin")

	api.AdminRetryArchiveImportExtract(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 1 {
		t.Fatalf("expected code=1, got=%d body=%s", resp.Code, rec.Body.String())
	}
	if stub.retryCalls != 0 {
		t.Fatalf("expected no retry call, got %d", stub.retryCalls)
	}
}

func TestAdminBatchUpdateArchiveImportFilesPassesCompleteInputAndReturnsItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	firstID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	secondID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	videoCollectionID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	imageCollectionID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	firstUpdatedAt := time.Date(2026, 7, 12, 5, 0, 0, 0, time.UTC)
	secondUpdatedAt := time.Date(2026, 7, 12, 5, 1, 0, 0, time.UTC)
	stub := &archiveImportServiceStub{
		batchUpdateFiles: []models.ArchiveImportFileListItem{
			{ID: firstID},
			{ID: secondID},
		},
	}
	api := &API{archiveImportSvc: stub}
	payload := `{
		"targets":[
			{"id":"11111111-1111-4111-8111-111111111111","updated_at":"2026-07-12T05:00:00Z"},
			{"id":"22222222-2222-4222-8222-222222222222","updated_at":"2026-07-12T05:01:00Z"}
		],
		"title_mode":"filename",
		"update_tags":true,
		"tags":["标签甲","标签乙"],
		"update_video_type":true,
		"video_type":"movie",
		"update_video_collection_ids":true,
		"video_collection_ids":["33333333-3333-4333-8333-333333333333"],
		"update_image_collection_ids":true,
		"image_collection_ids":["44444444-4444-4444-8444-444444444444"]
	}`

	rec := performAdminBatchUpdateArchiveImportFiles(api, payload)
	resp := decodeAdminArchiveImportResponse(t, rec)

	if resp.Code != 0 || stub.batchUpdateCalls != 1 {
		t.Fatalf("body=%s calls=%d", rec.Body.String(), stub.batchUpdateCalls)
	}
	if stub.batchUpdateInput.TitleMode != services.ArchiveImportTitleModeFilename {
		t.Fatalf("title mode=%q", stub.batchUpdateInput.TitleMode)
	}
	if len(stub.batchUpdateInput.Targets) != 2 {
		t.Fatalf("targets=%#v", stub.batchUpdateInput.Targets)
	}
	if stub.batchUpdateInput.Targets[0].ID != firstID || !stub.batchUpdateInput.Targets[0].UpdatedAt.Equal(firstUpdatedAt) {
		t.Fatalf("first target=%#v", stub.batchUpdateInput.Targets[0])
	}
	if stub.batchUpdateInput.Targets[1].ID != secondID || !stub.batchUpdateInput.Targets[1].UpdatedAt.Equal(secondUpdatedAt) {
		t.Fatalf("second target=%#v", stub.batchUpdateInput.Targets[1])
	}
	if !stub.batchUpdateInput.UpdateTags || !reflect.DeepEqual(stub.batchUpdateInput.Tags, []string{"标签甲", "标签乙"}) {
		t.Fatalf("tags patch=%#v", stub.batchUpdateInput)
	}
	if !stub.batchUpdateInput.UpdateVideoType || stub.batchUpdateInput.VideoType != "movie" {
		t.Fatalf("video type patch=%#v", stub.batchUpdateInput)
	}
	if !stub.batchUpdateInput.UpdateVideoCollectionIDs || !reflect.DeepEqual(stub.batchUpdateInput.VideoCollectionIDs, []uuid.UUID{videoCollectionID}) {
		t.Fatalf("video collection patch=%#v", stub.batchUpdateInput)
	}
	if !stub.batchUpdateInput.UpdateImageCollectionIDs || !reflect.DeepEqual(stub.batchUpdateInput.ImageCollectionIDs, []uuid.UUID{imageCollectionID}) {
		t.Fatalf("image collection patch=%#v", stub.batchUpdateInput)
	}
	if resp.Data["updated_count"] != float64(2) {
		t.Fatalf("updated_count=%v body=%s", resp.Data["updated_count"], rec.Body.String())
	}
	items, ok := resp.Data["items"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("items=%#v body=%s", resp.Data["items"], rec.Body.String())
	}
}

func TestAdminBatchUpdateArchiveImportFilesRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{
			name:    "空目标",
			payload: `{"targets":[],"title_mode":"filename"}`,
		},
		{
			name:    "目标时间为零值",
			payload: `{"targets":[{"id":"11111111-1111-4111-8111-111111111111"}],"title_mode":"filename"}`,
		},
		{
			name:    "目标ID非法",
			payload: `{"targets":[{"id":"not-a-uuid","updated_at":"2026-07-12T05:00:00Z"}],"title_mode":"filename"}`,
		},
		{
			name:    "视频合集ID非法",
			payload: `{"targets":[{"id":"11111111-1111-4111-8111-111111111111","updated_at":"2026-07-12T05:00:00Z"}],"title_mode":"filename","update_video_collection_ids":true,"video_collection_ids":["not-a-uuid"]}`,
		},
		{
			name:    "图片合集ID非法",
			payload: `{"targets":[{"id":"11111111-1111-4111-8111-111111111111","updated_at":"2026-07-12T05:00:00Z"}],"title_mode":"filename","update_image_collection_ids":true,"image_collection_ids":["not-a-uuid"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &archiveImportServiceStub{}
			rec := performAdminBatchUpdateArchiveImportFiles(&API{archiveImportSvc: stub}, tt.payload)
			resp := decodeAdminArchiveImportResponse(t, rec)

			if resp.Code != 1 {
				t.Fatalf("expected code=1, got=%d body=%s", resp.Code, rec.Body.String())
			}
			if stub.batchUpdateCalls != 0 {
				t.Fatalf("expected no batch update call, got %d", stub.batchUpdateCalls)
			}
		})
	}
}

func TestAdminBatchUpdateArchiveImportFilesPreservesStructuredError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	issueID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	stub := &archiveImportServiceStub{
		batchUpdateErr: &services.ArchiveImportBatchUpdateError{
			Reason: services.ArchiveImportBatchReasonStaleTarget,
			Issues: []services.ArchiveImportBatchUpdateIssue{
				{ID: issueID, RelativePath: "目录/视频.mp4", Message: "文件已被其他操作更新"},
			},
		},
	}
	payload := `{"targets":[{"id":"11111111-1111-4111-8111-111111111111","updated_at":"2026-07-12T05:00:00Z"}],"title_mode":"filename"}`

	rec := performAdminBatchUpdateArchiveImportFiles(&API{archiveImportSvc: stub}, payload)
	resp := decodeAdminArchiveImportResponse(t, rec)

	if resp.Code != 1078 || resp.Msg != "压缩包文件批量更新失败" {
		t.Fatalf("body=%s", rec.Body.String())
	}
	if resp.Data["reason"] != services.ArchiveImportBatchReasonStaleTarget {
		t.Fatalf("reason=%v body=%s", resp.Data["reason"], rec.Body.String())
	}
	issues, ok := resp.Data["issues"].([]any)
	if !ok || len(issues) != 1 {
		t.Fatalf("issues=%#v body=%s", resp.Data["issues"], rec.Body.String())
	}
	issue, ok := issues[0].(map[string]any)
	if !ok || issue["id"] != issueID.String() || issue["relative_path"] != "目录/视频.mp4" || issue["message"] != "文件已被其他操作更新" {
		t.Fatalf("issue=%#v body=%s", issues[0], rec.Body.String())
	}
}

func TestAdminBatchUpdateArchiveImportFilesUsesExistingServiceErrorResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	payload := `{"targets":[{"id":"11111111-1111-4111-8111-111111111111","updated_at":"2026-07-12T05:00:00Z"}],"title_mode":"filename"}`

	t.Run("服务不可用", func(t *testing.T) {
		rec := performAdminBatchUpdateArchiveImportFiles(&API{}, payload)
		resp := decodeAdminArchiveImportResponse(t, rec)
		if resp.Code != 1071 {
			t.Fatalf("expected code=1071, got=%d body=%s", resp.Code, rec.Body.String())
		}
	})

	t.Run("通用失败", func(t *testing.T) {
		stub := &archiveImportServiceStub{batchUpdateErr: errors.New("数据库写入失败")}
		rec := performAdminBatchUpdateArchiveImportFiles(&API{archiveImportSvc: stub}, payload)
		resp := decodeAdminArchiveImportResponse(t, rec)
		if resp.Code != 1072 || resp.Msg != "数据库写入失败" {
			t.Fatalf("body=%s", rec.Body.String())
		}
		if stub.batchUpdateCalls != 1 {
			t.Fatalf("expected one batch update call, got %d", stub.batchUpdateCalls)
		}
	})
}

func TestAdminBatchUpdateArchiveImportFilesRouteIsRegisteredBeforeFileID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := &API{}
	api.Register(router)

	batchRouteIndex := -1
	fileRouteIndex := -1
	for index, route := range router.Routes() {
		switch route.Method + " " + route.Path {
		case "PUT /api/v1/admin/archive-import/files/batch-update":
			batchRouteIndex = index
		case "PUT /api/v1/admin/archive-import/files/:id":
			fileRouteIndex = index
		}
	}
	if batchRouteIndex == -1 {
		t.Fatal("expected batch update route to be registered")
	}
	if fileRouteIndex == -1 {
		t.Fatal("expected file update route to be registered")
	}
	if batchRouteIndex > fileRouteIndex {
		t.Fatalf("expected static batch route before file ID route, indexes %d and %d", batchRouteIndex, fileRouteIndex)
	}
}

func performAdminBatchUpdateArchiveImportFiles(api *API, payload string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/archive-import/files/batch-update", bytes.NewBufferString(payload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	api.AdminBatchUpdateArchiveImportFiles(ctx)
	return rec
}

func decodeAdminArchiveImportResponse(t *testing.T, rec *httptest.ResponseRecorder) apiEnvelope {
	t.Helper()
	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v body=%s", err, rec.Body.String())
	}
	return resp
}
