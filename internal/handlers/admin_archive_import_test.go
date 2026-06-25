package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/services"
)

type archiveImportServiceStub struct {
	uploadBatch   models.ArchiveImportBatch
	uploadErr     error
	uploadCalls   int
	uploadInput   services.ArchiveImportUploadInput
	uploadHeader  *multipart.FileHeader
	retryBatch    models.ArchiveImportBatch
	retryErr      error
	retryCalls    int
	retryBatchID  uuid.UUID
	retryPassword string
	retryMode     string
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
			EncodingRequestedMode:  "gbk",
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
