package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestDeleteVideosIndividuallySummarizesPartialFailures(t *testing.T) {
	t.Parallel()

	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()
	calls := make([]uuid.UUID, 0, 3)

	result := deleteVideosIndividually([]uuid.UUID{id1, id2, id3}, func(videoID uuid.UUID) error {
		calls = append(calls, videoID)
		if videoID == id2 {
			return errors.New("video not found")
		}
		return nil
	})

	if len(calls) != 3 {
		t.Fatalf("expected 3 delete calls, got %d", len(calls))
	}
	if result.RequestedCount != 3 {
		t.Fatalf("expected requested_count=3, got %d", result.RequestedCount)
	}
	if result.SuccessCount != 2 {
		t.Fatalf("expected success_count=2, got %d", result.SuccessCount)
	}
	if result.FailureCount != 1 {
		t.Fatalf("expected failure_count=1, got %d", result.FailureCount)
	}
	if len(result.Results) != 3 {
		t.Fatalf("expected 3 per-item results, got %d", len(result.Results))
	}
	if !result.Results[0].Deleted || result.Results[0].Message != "" {
		t.Fatalf("expected first item success, got %+v", result.Results[0])
	}
	if result.Results[1].Deleted || result.Results[1].Message != "video not found" {
		t.Fatalf("expected second item failure message, got %+v", result.Results[1])
	}
	if result.Results[1].VideoID != id2 {
		t.Fatalf("expected second item id=%s, got %s", id2, result.Results[1].VideoID)
	}
}

func TestAdminBatchDeleteVideosRejectsEmptyVideoIDs(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	api := &API{}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/videos/batch-delete", bytes.NewBufferString(`{"video_ids":[]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	api.AdminBatchDeleteVideos(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 1 {
		t.Fatalf("expected code=1, got=%d body=%s", resp.Code, rec.Body.String())
	}
}

func TestAdminBatchDeleteVideosRejectsInvalidUUID(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	api := &API{}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/videos/batch-delete", bytes.NewBufferString(`{"video_ids":["bad-id"]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	api.AdminBatchDeleteVideos(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 1 {
		t.Fatalf("expected code=1, got=%d body=%s", resp.Code, rec.Body.String())
	}
}

func TestAdminBatchUpdateVideosRejectsEmptyVideoIDs(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	api := &API{}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/videos/batch-update", bytes.NewBufferString(`{"video_ids":[],"update_title":true,"title":"新标题"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	api.AdminBatchUpdateVideos(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 1 {
		t.Fatalf("expected code=1, got=%d body=%s", resp.Code, rec.Body.String())
	}
}

func TestAdminBatchUpdateVideosRejectsInvalidTagsMode(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	api := &API{}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/videos/batch-update", bytes.NewBufferString(`{"video_ids":["11111111-1111-1111-1111-111111111111"],"update_tags":true,"tags_mode":"oops","tags":["tag1"]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	api.AdminBatchUpdateVideos(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 1 {
		t.Fatalf("expected code=1, got=%d body=%s", resp.Code, rec.Body.String())
	}
}

func TestAdminBatchUpdateVideosRejectsEmptyPatch(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	api := &API{}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/videos/batch-update", bytes.NewBufferString(`{"video_ids":["11111111-1111-1111-1111-111111111111"]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	api.AdminBatchUpdateVideos(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 1 {
		t.Fatalf("expected code=1, got=%d body=%s", resp.Code, rec.Body.String())
	}
}

func TestAdminBatchUpdateVideosRejectsInvalidImageCollectionID(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	api := &API{}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/videos/batch-update", bytes.NewBufferString(`{"video_ids":["11111111-1111-1111-1111-111111111111"],"update_image_collection_id":true,"image_collection_id":"bad-id"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	api.AdminBatchUpdateVideos(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 1 {
		t.Fatalf("expected code=1, got=%d body=%s", resp.Code, rec.Body.String())
	}
}
