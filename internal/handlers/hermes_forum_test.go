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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/services"
)

type hermesForumServiceStub struct {
	adminListPage     int
	adminListPageSize int
	adminListQuery    string
	adminListResult   []models.AdminForumPostListItem
	adminListTotal    int
	adminListErr      error
	discoverResult    []models.ForumPostDiscoverResult
	discoverErr       error
	inspectResult     models.ForumPostInspectionResult
	inspectErr        error
}

func (s *hermesForumServiceStub) ListAdmin(_ context.Context, page, pageSize int, query string) ([]models.AdminForumPostListItem, int, error) {
	s.adminListPage = page
	s.adminListPageSize = pageSize
	s.adminListQuery = query
	return s.adminListResult, s.adminListTotal, s.adminListErr
}

func (s *hermesForumServiceStub) Discover(context.Context, models.ForumPostDiscoverInput) ([]models.ForumPostDiscoverResult, error) {
	return s.discoverResult, s.discoverErr
}

func (s *hermesForumServiceStub) CompleteInspection(context.Context, uuid.UUID, models.ForumPostInspectionInput) (models.ForumPostInspectionResult, error) {
	return s.inspectResult, s.inspectErr
}

func TestAdminForumPostsReturnsPaginatedReadModel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	observedAt := time.Date(2026, 8, 4, 10, 30, 0, 0, time.UTC)
	service := &hermesForumServiceStub{
		adminListResult: []models.AdminForumPostListItem{{
			ID:               uuid.New(),
			Title:            "帖子标题",
			URL:              "https://sehuatang.org/thread-1",
			InspectionStatus: models.ForumPostStatusRestricted,
			Attachments:      []string{"https://sehuatang.org/attachment.php?aid=1"},
			ED2KLinks:        []string{"ed2k://|file|a.zip|1|0123456789ABCDEF0123456789ABCDEF|/"},
			ObservedAt:       observedAt,
		}},
		adminListTotal: 21,
	}
	api := &API{hermesForumSvc: service}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/forum-posts?page=2&page_size=20&q=%20%E8%B5%84%E6%BA%90%20", nil)

	api.AdminForumPosts(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d, want=%d body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if service.adminListPage != 2 || service.adminListPageSize != 20 || service.adminListQuery != " 资源 " {
		t.Fatalf("service query=(%d,%d,%q), want (2,20,%q)", service.adminListPage, service.adminListPageSize, service.adminListQuery, " 资源 ")
	}
	var envelope struct {
		Code int `json:"code"`
		Data struct {
			Items      []models.AdminForumPostListItem `json:"items"`
			TotalCount int                             `json:"total_count"`
			Page       int                             `json:"page"`
			PageSize   int                             `json:"page_size"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if envelope.Code != 0 || envelope.Data.TotalCount != 21 || envelope.Data.Page != 2 || envelope.Data.PageSize != 20 || len(envelope.Data.Items) != 1 || envelope.Data.Items[0].InspectionStatus != models.ForumPostStatusRestricted {
		t.Fatalf("unexpected response: %#v", envelope)
	}
}

func TestAdminForumPostsMapsInvalidSearchQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	api := &API{hermesForumSvc: &hermesForumServiceStub{adminListErr: services.ErrAdminForumSearchQueryTooLong}}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/forum-posts?q="+strings.Repeat("a", 201), nil)

	api.AdminForumPosts(ctx)

	var envelope struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if envelope.Code != 1 || envelope.Msg != services.ErrAdminForumSearchQueryTooLong.Error() {
		t.Fatalf("unexpected response: %#v", envelope)
	}
}

func TestRegisterIncludesAdminForumPostRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	api := &API{}
	router := gin.New()
	api.Register(router)

	for _, route := range router.Routes() {
		if route.Method == http.MethodGet && route.Path == "/api/v1/admin/forum-posts" {
			return
		}
	}
	t.Fatal("未注册管理员论坛资源列表路由")
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
