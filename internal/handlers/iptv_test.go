package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"video-server/internal/models"
	"video-server/internal/services"
)

func TestRegisterIncludesIPTVRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := &API{}
	api.Register(router)

	routes := map[string]struct{}{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}

	for _, want := range []string{
		"GET /api/v1/admin/iptv/playlist",
		"POST /api/v1/admin/iptv/playlist/upload",
		"PUT /api/v1/admin/iptv/playlist/source",
		"POST /api/v1/admin/iptv/playlist/refresh",
		"GET /api/v1/tv/iptv/channels",
	} {
		if _, ok := routes[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
}

func TestAdminIPTVPlaylistReturnsEmptyStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/iptv/playlist", nil)
	api := &API{iptvSvc: &fakeIPTVService{status: models.IPTVPlaylistStatus{
		Groups:   []string{},
		Channels: []models.IPTVChannel{},
	}}}

	api.AdminIPTVPlaylist(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{`"code":0`, `"channel_count":0`, `"channels":[]`, `"groups":[]`} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected response to contain %s, got %s", want, body)
		}
	}
}

func TestAdminIPTVUploadPlaylistParsesSuccessfully(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "channels.m3u")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	_, _ = part.Write([]byte("#EXTM3U\n#EXTINF:-1 group-title=\"新闻\",新闻频道\nhttps://live.example/news.m3u8\n"))
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/iptv/playlist/upload", &body)
	ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())
	api := &API{iptvSvc: &fakeIPTVService{}}

	api.AdminIPTVUploadPlaylist(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	responseBody := rec.Body.String()
	for _, want := range []string{`"code":0`, `"channel_count":1`, `"name":"新闻频道"`, `"group":"新闻"`} {
		if !strings.Contains(responseBody, want) {
			t.Fatalf("expected response to contain %s, got %s", want, responseBody)
		}
	}
}

func TestAdminIPTVRefreshWithoutSourceURLFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/iptv/playlist/refresh", nil)
	api := &API{iptvSvc: &fakeIPTVService{refreshErr: services.ErrIPTVSourceURLRequired}}

	api.AdminIPTVRefreshPlaylist(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"code":2304`) || !strings.Contains(body, "请先设置 IPTV 远程地址") {
		t.Fatalf("unexpected response: %s", body)
	}
}

func TestTVIPTVChannelsReturnsFullPlaylistStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	updatedAt := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tv/iptv/channels", nil)
	api := &API{iptvSvc: &fakeIPTVService{status: models.IPTVPlaylistStatus{
		SourceURL:    "https://example.com/live.m3u",
		UpdatedAt:    &updatedAt,
		ChannelCount: 1,
		SkippedCount: 2,
		Groups:       []string{"新闻"},
		Channels: []models.IPTVChannel{
			{ID: "c1", Name: "新闻频道", URL: "https://live.example/news.m3u8", Group: "新闻", SortOrder: 0},
		},
	}}}

	api.TVIPTVChannels(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`"source_url":"https://example.com/live.m3u"`,
		`"updated_at":"2026-05-18T12:00:00Z"`,
		`"channel_count":1`,
		`"skipped_count":2`,
		`"groups":["新闻"]`,
		`"name":"新闻频道"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected response to contain %s, got %s", want, body)
		}
	}
}

type fakeIPTVService struct {
	status     models.IPTVPlaylistStatus
	refreshErr error
}

func (s *fakeIPTVService) Status(context.Context) (models.IPTVPlaylistStatus, error) {
	return s.status, nil
}

func (s *fakeIPTVService) Upload(_ context.Context, _ string, r io.Reader) (models.IPTVPlaylistStatus, error) {
	channels, skipped := services.ParseM3UPlaylist(r)
	return services.BuildIPTVPlaylistStatus(models.IPTVPlaylistMeta{SkippedCount: skipped}, channels), nil
}

func (s *fakeIPTVService) SaveSourceURL(_ context.Context, sourceURL string) (models.IPTVPlaylistStatus, error) {
	s.status.SourceURL = sourceURL
	return s.status, nil
}

func (s *fakeIPTVService) Refresh(context.Context) (models.IPTVPlaylistStatus, error) {
	if s.refreshErr != nil {
		return models.IPTVPlaylistStatus{}, s.refreshErr
	}
	return s.status, nil
}

func TestIPTVHandlerMapsGenericServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/iptv/playlist", nil)
	api := &API{iptvSvc: &fakeIPTVServiceWithStatusError{err: errors.New("boom")}}

	api.AdminIPTVPlaylist(ctx)

	if !strings.Contains(rec.Body.String(), `"code":2301`) {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
}

type fakeIPTVServiceWithStatusError struct {
	err error
}

func (s *fakeIPTVServiceWithStatusError) Status(context.Context) (models.IPTVPlaylistStatus, error) {
	return models.IPTVPlaylistStatus{}, s.err
}

func (s *fakeIPTVServiceWithStatusError) Upload(context.Context, string, io.Reader) (models.IPTVPlaylistStatus, error) {
	return models.IPTVPlaylistStatus{}, nil
}

func (s *fakeIPTVServiceWithStatusError) SaveSourceURL(context.Context, string) (models.IPTVPlaylistStatus, error) {
	return models.IPTVPlaylistStatus{}, nil
}

func (s *fakeIPTVServiceWithStatusError) Refresh(context.Context) (models.IPTVPlaylistStatus, error) {
	return models.IPTVPlaylistStatus{}, nil
}
