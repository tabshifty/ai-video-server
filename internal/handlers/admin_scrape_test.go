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

func TestAdminScrapePreviewTVNormalizesSeasonEpisodeTitle(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	var requestedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search/tv":
			requestedQuery = r.URL.Query().Get("query")
			_, _ = w.Write([]byte(`{"results":[{"id":101}]}`))
		case "/tv/101":
			_, _ = w.Write([]byte(`{"id":101,"name":"庆余年","original_name":"庆余年","overview":"简介","first_air_date":"2019-11-26","number_of_seasons":1,"number_of_episodes":46}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	api := &API{scrapeSvc: services.NewScraperService(nil, "demo-key", server.URL, "", "", 2*time.Second)}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/scrape/preview", bytes.NewBufferString(`{"title":"庆余年第一季第一集","type":"tv"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	api.AdminScrapePreview(ctx)

	if requestedQuery != "庆余年" {
		t.Fatalf("expected normalized tmdb query, got=%q", requestedQuery)
	}

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code=0, got=%d body=%s", resp.Code, rec.Body.String())
	}

	rows, ok := resp.Data["candidates"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("expected one candidate, got=%v", resp.Data["candidates"])
	}
	candidate, ok := rows[0].(map[string]any)
	if !ok {
		t.Fatalf("expected candidate object, got=%T", rows[0])
	}
	if candidate["parsed_title"] != "庆余年" {
		t.Fatalf("expected parsed_title 庆余年, got=%v", candidate["parsed_title"])
	}
	if candidate["parsed_season_number"] != float64(1) {
		t.Fatalf("expected parsed_season_number=1, got=%v", candidate["parsed_season_number"])
	}
	if candidate["parsed_episode_number"] != float64(1) {
		t.Fatalf("expected parsed_episode_number=1, got=%v", candidate["parsed_episode_number"])
	}
}
