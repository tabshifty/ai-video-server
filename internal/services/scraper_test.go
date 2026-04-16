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
)

func TestPreviewMovieUsesChineseLanguageAndEnglishFallback(t *testing.T) {
	var searchLangs []string
	var detailLangs []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search/movie":
			searchLangs = append(searchLangs, r.URL.Query().Get("language"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": 79277},
				},
			})
		case "/movie/79277":
			lang := r.URL.Query().Get("language")
			detailLangs = append(detailLangs, lang)
			if lang == "zh-CN" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":           79277,
					"title":        "闯关东",
					"overview":     "",
					"release_date": "2008-01-02",
					"genres": []map[string]any{
						{"id": 18, "name": "剧情"},
						{"id": 10751, "name": "家庭"},
					},
					"poster_path": "/zh.jpg",
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":           79277,
				"title":        "Pathfinding to the Northeast",
				"overview":     "English overview fallback.",
				"release_date": "2008-01-02",
				"genres": []map[string]any{
					{"id": 18, "name": "Drama"},
					{"id": 10751, "name": "Family"},
				},
				"poster_path": "/en.jpg",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "demo-key", server.URL, "", "", 2*time.Second)
	got, err := svc.PreviewMovie(context.Background(), "闯关东", 2008)
	if err != nil {
		t.Fatalf("PreviewMovie returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got=%d", len(got))
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

	candidate := got[0]
	if candidate["title"] != "闯关东" {
		t.Fatalf("expected chinese title, got=%v", candidate["title"])
	}
	if candidate["overview"] != "English overview fallback." {
		t.Fatalf("expected english overview fallback, got=%v", candidate["overview"])
	}
	genres, ok := candidate["genres"].([]string)
	if !ok {
		t.Fatalf("expected genres []string, got=%T", candidate["genres"])
	}
	if len(genres) != 2 || genres[0] != "剧情" || genres[1] != "家庭" {
		t.Fatalf("expected chinese genres, got=%v", genres)
	}
	metadata, ok := candidate["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("expected metadata object, got=%T", candidate["metadata"])
	}
	if asString(metadata["title"]) != "闯关东" {
		t.Fatalf("expected metadata.title chinese, got=%v", metadata["title"])
	}
	if asString(metadata["overview"]) != "English overview fallback." {
		t.Fatalf("expected metadata.overview fallback, got=%v", metadata["overview"])
	}
}

func TestPreviewTVUsesChineseLanguageAndFallback(t *testing.T) {
	var searchLangs []string
	var detailLangs []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search/tv":
			searchLangs = append(searchLangs, r.URL.Query().Get("language"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": 79277},
				},
			})
		case "/tv/79277":
			lang := r.URL.Query().Get("language")
			detailLangs = append(detailLangs, lang)
			if lang == "zh-CN" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":             79277,
					"name":           "",
					"original_name":  "闯关东",
					"overview":       "中文简介",
					"first_air_date": "2008-01-02",
					"genres": []map[string]any{
						{"id": 18, "name": "剧情"},
					},
					"poster_path": "/zh.jpg",
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":             79277,
				"name":           "Pathfinding to the Northeast",
				"original_name":  "闯关东",
				"overview":       "English overview.",
				"first_air_date": "2008-01-02",
				"genres": []map[string]any{
					{"id": 18, "name": "Drama"},
				},
				"poster_path": "/en.jpg",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "demo-key", server.URL, "", "", 2*time.Second)
	got, err := svc.PreviewTV(context.Background(), "闯关东", 2008)
	if err != nil {
		t.Fatalf("PreviewTV returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got=%d", len(got))
	}

	if !slices.Contains(searchLangs, "zh-CN") {
		t.Fatalf("expected search/tv to request language=zh-CN, got=%v", searchLangs)
	}
	if !slices.Contains(detailLangs, "zh-CN") {
		t.Fatalf("expected tv detail to request language=zh-CN, got=%v", detailLangs)
	}
	if !slices.Contains(detailLangs, "") {
		t.Fatalf("expected fallback tv detail request without language, got=%v", detailLangs)
	}

	candidate := got[0]
	if candidate["title"] != "Pathfinding to the Northeast" {
		t.Fatalf("expected title fallback from english detail, got=%v", candidate["title"])
	}
	if candidate["overview"] != "中文简介" {
		t.Fatalf("expected chinese overview, got=%v", candidate["overview"])
	}
	genres, ok := candidate["genres"].([]string)
	if !ok {
		t.Fatalf("expected genres []string, got=%T", candidate["genres"])
	}
	if len(genres) != 1 || genres[0] != "剧情" {
		t.Fatalf("expected chinese genres, got=%v", genres)
	}
}

func TestExtractCastNames(t *testing.T) {
	raw := map[string]any{
		"cast": []any{
			map[string]any{"name": "演员甲"},
			map[string]any{"name": "演员乙"},
			map[string]any{"name": "演员甲"},
			map[string]any{"name": "", "original_name": "Actor C"},
			map[string]any{"original_name": "  "},
		},
	}

	got := extractCastNames(raw, 3)
	want := []string{"演员甲", "演员乙", "Actor C"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractCastNames() = %v, want %v", got, want)
	}
}
