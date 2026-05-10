package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeContentTranslator struct {
	result TranslationResult
	err    error
	calls  int
}

func (f *fakeContentTranslator) TranslateScrapeContent(_ context.Context, title, description string) (TranslationResult, error) {
	f.calls++
	if f.err != nil {
		return TranslationResult{}, f.err
	}
	return f.result, nil
}

func TestScrapeMovieUploadStoresLocalizedFields(t *testing.T) {
	t.Parallel()

	var searchLangs []string
	var detailLangs []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search/movie":
			searchLangs = append(searchLangs, r.URL.Query().Get("language"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{{"id": 301}},
			})
		case "/movie/301":
			lang := r.URL.Query().Get("language")
			detailLangs = append(detailLangs, lang)
			if lang == "zh-CN" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":           301,
					"title":        "",
					"overview":     "",
					"release_date": "2021-01-01",
					"genres": []map[string]any{
						{"id": 18, "name": "剧情"},
					},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":           301,
				"title":        "English Title",
				"overview":     "English overview",
				"release_date": "2021-01-01",
				"genres": []map[string]any{
					{"id": 18, "name": "Drama"},
				},
			})
		case "/movie/301/credits":
			_ = json.NewEncoder(w).Encode(map[string]any{"cast": []map[string]any{}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	repo := &fakeScraperRepo{}
	svc := NewScraperService(repo, "demo-key", server.URL, t.TempDir(), "", 2*time.Second)
	svc.textTranslator = &fakeContentTranslator{
		result: TranslationResult{
			Title:       "中文片名",
			Description: "中文简介",
		},
	}

	videoID := uuid.New()
	got, err := svc.ScrapeMovieUpload(context.Background(), videoID, "/tmp/English.Title.2021.mkv", "English.Title.2021.mkv")
	if err != nil {
		t.Fatalf("ScrapeMovieUpload returned error: %v", err)
	}

	if !slices.Contains(searchLangs, "zh-CN") {
		t.Fatalf("expected search/movie to request language=zh-CN, got=%v", searchLangs)
	}
	if !slices.Contains(detailLangs, "zh-CN") {
		t.Fatalf("expected movie detail to request language=zh-CN, got=%v", detailLangs)
	}
	if !slices.Contains(detailLangs, "") {
		t.Fatalf("expected fallback movie detail request without language, got=%v", detailLangs)
	}
	if got.Title != "中文片名" {
		t.Fatalf("unexpected title: %s", got.Title)
	}
	if got.Description != "中文简介" {
		t.Fatalf("unexpected description: %s", got.Description)
	}
	if repo.lastUpdate.title != "中文片名" || repo.lastUpdate.description != "中文简介" {
		t.Fatalf("unexpected repo update title/description: %q / %q", repo.lastUpdate.title, repo.lastUpdate.description)
	}
	if repo.lastUpdate.metadata["title_original"] != "English Title" {
		t.Fatalf("unexpected title_original: %v", repo.lastUpdate.metadata["title_original"])
	}
	if repo.lastUpdate.metadata["description_original"] != "English overview" {
		t.Fatalf("unexpected description_original: %v", repo.lastUpdate.metadata["description_original"])
	}
	if repo.lastUpdate.metadata["title_zh"] != "中文片名" {
		t.Fatalf("unexpected title_zh: %v", repo.lastUpdate.metadata["title_zh"])
	}
	if repo.lastUpdate.metadata["description_zh"] != "中文简介" {
		t.Fatalf("unexpected description_zh: %v", repo.lastUpdate.metadata["description_zh"])
	}
	tmdbRaw, ok := repo.lastUpdate.metadata["tmdb"].(map[string]any)
	if !ok {
		t.Fatalf("expected metadata.tmdb object, got=%T", repo.lastUpdate.metadata["tmdb"])
	}
	if tmdbRaw["title"] != "English Title" {
		t.Fatalf("expected metadata.tmdb.title fallback, got=%v", tmdbRaw["title"])
	}
}

func TestScrapeAVUploadStoresLocalizedFieldsAndKeepsActors(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search":
			_, _ = w.Write([]byte(`
<html>
  <body>
    <a href="/v/code-hit"><span>ABP-123 作品标题</span></a>
  </body>
</html>
`))
		case "/v/code-hit":
			_, _ = w.Write([]byte(`
<html>
  <head>
    <title>ABP-123 作品标题 - JavDB</title>
    <meta property="og:image" content="/covers/abp123.jpg" />
    <meta name="description" content="作品简介" />
  </head>
  <body>
    <h2 class="title is-4">ABP-123 作品标题</h2>
    <div><strong>番號:</strong><span>ABP-123</span></div>
    <div><strong>日期:</strong><span>2024-01-02</span></div>
    <a href="/actors/a1">演员甲</a>
    <a href="/actors/a2">演员乙</a>
  </body>
</html>
`))
		case "/covers/abp123.jpg", "/thumbs/abp123.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("fake-image"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	repo := &fakeScraperRepo{
		nextActorIDs: []uuid.UUID{uuid.New(), uuid.New()},
	}
	svc := NewScraperService(repo, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "av-scraper-test", time.Second)
	svc.textTranslator = &fakeContentTranslator{
		result: TranslationResult{
			Title:       "ABP-123 作品标题",
			Description: "作品简介中文",
		},
	}

	videoID := uuid.New()
	got, err := svc.ScrapeAVUpload(context.Background(), videoID, "/tmp/ABP-123.sample.mkv", "ABP-123.sample.mkv")
	if err != nil {
		t.Fatalf("ScrapeAVUpload returned error: %v", err)
	}

	if got.Title != "ABP-123 作品标题" {
		t.Fatalf("unexpected title: %s", got.Title)
	}
	if got.Description != "作品简介中文" {
		t.Fatalf("unexpected description: %s", got.Description)
	}
	if repo.resolveActorSource != "scrape_av" {
		t.Fatalf("expected resolve actor source scrape_av, got=%s", repo.resolveActorSource)
	}
	if !reflect.DeepEqual(repo.resolveActorNames, []string{"演员甲", "演员乙"}) {
		t.Fatalf("unexpected resolved actor names: %v", repo.resolveActorNames)
	}
	if repo.lastUpdate.metadata["title_original"] != "ABP-123 作品标题" {
		t.Fatalf("unexpected title_original: %v", repo.lastUpdate.metadata["title_original"])
	}
	if repo.lastUpdate.metadata["description_original"] != "作品简介" {
		t.Fatalf("unexpected description_original: %v", repo.lastUpdate.metadata["description_original"])
	}
	if repo.lastUpdate.metadata["title_zh"] != "ABP-123 作品标题" {
		t.Fatalf("unexpected title_zh: %v", repo.lastUpdate.metadata["title_zh"])
	}
	if repo.lastUpdate.metadata["description_zh"] != "作品简介中文" {
		t.Fatalf("unexpected description_zh: %v", repo.lastUpdate.metadata["description_zh"])
	}
}
