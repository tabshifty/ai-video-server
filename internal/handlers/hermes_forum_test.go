package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/services"
)

type hermesForumServiceStub struct {
	discoverResult []models.ForumPostDiscoverResult
	discoverErr    error
	inspectResult  models.ForumPostInspectionResult
	inspectErr     error
}

func (s *hermesForumServiceStub) Discover(context.Context, models.ForumPostDiscoverInput) ([]models.ForumPostDiscoverResult, error) {
	return s.discoverResult, s.discoverErr
}

func (s *hermesForumServiceStub) CompleteInspection(context.Context, uuid.UUID, models.ForumPostInspectionInput) (models.ForumPostInspectionResult, error) {
	return s.inspectResult, s.inspectErr
}

func TestHermesDiscoverForumPostsReturnsRealHTTPStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	postID := uuid.New()
	tests := []struct {
		name       string
		body       string
		serviceErr error
		wantStatus int
	}{
		{name: "success", body: `{"mode":"dedupe_only","source":"sehuatang","board_key":"95","posts":[{"tid":"1"}]}`, wantStatus: http.StatusOK},
		{name: "invalid json", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "unknown field", body: `{"mode":"dedupe_only","source":"sehuatang","board_key":"95","posts":[{"tid":"1"}],"extra":true}`, wantStatus: http.StatusBadRequest},
		{name: "body too large", body: `{"padding":"` + strings.Repeat("x", hermesForumMaxRequestBytes) + `"}`, wantStatus: http.StatusRequestEntityTooLarge},
		{name: "invalid input", body: `{"mode":"normal"}`, serviceErr: services.ErrHermesForumInvalidInput, wantStatus: http.StatusBadRequest},
		{name: "internal error", body: `{"mode":"dedupe_only","source":"sehuatang","board_key":"95","posts":[{"tid":"1"}]}`, serviceErr: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &API{hermesForumSvc: &hermesForumServiceStub{
				discoverResult: []models.ForumPostDiscoverResult{{TID: "1", ID: postID, Disposition: models.ForumPostDispositionCreated, InspectionStatus: models.ForumPostStatusDedupeOnly}},
				discoverErr:    tt.serviceErr,
			}}
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/integrations/hermes/forum-posts/discover", bytes.NewBufferString(tt.body))
			ctx.Request.Header.Set("Content-Type", "application/json")

			api.HermesDiscoverForumPosts(ctx)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status=%d, want=%d body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			var envelope struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if tt.wantStatus == http.StatusOK && envelope.Code != 0 {
				t.Fatalf("code=%d, want=0", envelope.Code)
			}
			if tt.wantStatus != http.StatusOK && envelope.Code != tt.wantStatus {
				t.Fatalf("code=%d, want=%d", envelope.Code, tt.wantStatus)
			}
		})
	}
}

func TestHermesCompleteForumPostInspectionMapsDomainErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		pathID     string
		serviceErr error
		wantStatus int
	}{
		{name: "success", pathID: uuid.NewString(), wantStatus: http.StatusOK},
		{name: "invalid id", pathID: "not-a-uuid", wantStatus: http.StatusBadRequest},
		{name: "not found", pathID: uuid.NewString(), serviceErr: services.ErrHermesForumPostNotFound, wantStatus: http.StatusNotFound},
		{name: "conflict", pathID: uuid.NewString(), serviceErr: services.ErrHermesForumInspectionConflict, wantStatus: http.StatusConflict},
		{name: "internal error", pathID: uuid.NewString(), serviceErr: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &API{hermesForumSvc: &hermesForumServiceStub{
				inspectResult: models.ForumPostInspectionResult{Status: models.ForumPostStatusInspected},
				inspectErr:    tt.serviceErr,
			}}
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Params = gin.Params{{Key: "id", Value: tt.pathID}}
			ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/integrations/hermes/forum-posts/"+tt.pathID+"/inspection", bytes.NewBufferString(`{"status":"inspected","filter_decision":"included","filter_reasons":["ed2k"],"ed2k_links":["ed2k://|file|a.zip|1|0123456789ABCDEF0123456789ABCDEF|/"]}`))
			ctx.Request.Header.Set("Content-Type", "application/json")

			api.HermesCompleteForumPostInspection(ctx)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status=%d, want=%d body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
