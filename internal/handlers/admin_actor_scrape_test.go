package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"video-server/internal/services"
)

type apiEnvelope struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data map[string]any `json:"data"`
}

func TestAdminActorScrapePreviewInvalidSource(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	api := &API{}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/actors/scrape/preview", bytes.NewBufferString(`{"name":"演员甲","source":"unknown"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	api.AdminActorScrapePreview(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 1 {
		t.Fatalf("expected code=1, got=%d body=%s", resp.Code, rec.Body.String())
	}
}

func TestAdminActorScrapePreviewTMDBSuccess(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search/person":
			_, _ = w.Write([]byte(`{"results":[{"id":101}]}`))
		case "/person/101":
			_, _ = w.Write([]byte(`{"id":101,"name":"演员甲","gender":2,"biography":"简介","profile_path":"/actor.jpg","also_known_as":["别名甲"]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := services.NewScraperService(nil, "demo-key", server.URL, "", "", 2*time.Second)
	api := &API{scrapeSvc: svc}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/actors/scrape/preview", bytes.NewBufferString(`{"name":"演员甲","source":"tmdb","limit":5}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	api.AdminActorScrapePreview(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code=0, got=%d body=%s", resp.Code, rec.Body.String())
	}
	items, ok := resp.Data["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected one item, got=%v", resp.Data["items"])
	}
}
