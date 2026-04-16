package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
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

func TestPreviewActorByNameTMDB(t *testing.T) {
	var searchLangs []string
	var detailLangs []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search/person":
			searchLangs = append(searchLangs, r.URL.Query().Get("language"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": 1001},
				},
			})
		case "/person/1001":
			lang := r.URL.Query().Get("language")
			detailLangs = append(detailLangs, lang)
			if lang == "zh-CN" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":           1001,
					"name":         "演员甲",
					"gender":       2,
					"biography":    "",
					"profile_path": "/actor-zh.jpg",
					"also_known_as": []string{
						"演员甲A",
					},
					"birthday":       "1995-01-01",
					"place_of_birth": "Japan, Tokyo",
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":           1001,
				"name":         "Actor A",
				"gender":       2,
				"biography":    "fallback bio",
				"profile_path": "/actor-en.jpg",
				"also_known_as": []string{
					"Actor A",
				},
				"birthday":       "1995-01-01",
				"place_of_birth": "Japan, Tokyo",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "demo-key", server.URL, "", "", 2*time.Second)
	got, err := svc.PreviewActorByName(context.Background(), "演员甲", "tmdb", 10)
	if err != nil {
		t.Fatalf("PreviewActorByName returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got=%d", len(got))
	}
	if !slices.Contains(searchLangs, "zh-CN") {
		t.Fatalf("expected search/person to request language=zh-CN, got=%v", searchLangs)
	}
	if !slices.Contains(detailLangs, "zh-CN") {
		t.Fatalf("expected person detail to request language=zh-CN, got=%v", detailLangs)
	}
	if !slices.Contains(detailLangs, "") {
		t.Fatalf("expected fallback person detail without language, got=%v", detailLangs)
	}

	item := got[0]
	if item.Source != "tmdb" {
		t.Fatalf("expected source tmdb, got=%s", item.Source)
	}
	if item.Name != "演员甲" {
		t.Fatalf("expected chinese name, got=%s", item.Name)
	}
	if item.Gender != "男" {
		t.Fatalf("expected gender 男, got=%s", item.Gender)
	}
	if item.Notes != "fallback bio" {
		t.Fatalf("expected fallback bio, got=%s", item.Notes)
	}
	if item.AvatarURL != "https://image.tmdb.org/t/p/w500/actor-zh.jpg" {
		t.Fatalf("unexpected avatar url: %s", item.AvatarURL)
	}
	if len(item.Aliases) != 1 || item.Aliases[0] != "演员甲A" {
		t.Fatalf("unexpected aliases: %v", item.Aliases)
	}
}

func TestPreviewActorByNameJavDB(t *testing.T) {
	var userAgents []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgents = append(userAgents, r.Header.Get("User-Agent"))
		if r.URL.Path != "/search" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("f") != "actor" {
			t.Fatalf("expected f=actor, got=%s", r.URL.Query().Get("f"))
		}
		_, _ = w.Write([]byte(`
<html>
  <body>
    <a href="/actors/abc123" title="演员甲"><img src="/avatars/a.jpg" /></a>
    <a href="/actors/def456"><span>演员乙</span></a>
    <a href="/actors/abc123"><span>演员甲</span></a>
  </body>
</html>
`))
	}))
	defer server.Close()

	svc := NewScraperService(nil, "", "https://api.themoviedb.org/3", "", "", 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "actor-scraper-test", time.Second)

	got, err := svc.PreviewActorByName(context.Background(), "演员", "javdb", 10)
	if err != nil {
		t.Fatalf("PreviewActorByName returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 candidates, got=%d", len(got))
	}
	if len(userAgents) == 0 || userAgents[0] != "actor-scraper-test" {
		t.Fatalf("expected custom user-agent, got=%v", userAgents)
	}
	if got[0].Source != "javdb" || got[0].ExternalID != "abc123" {
		t.Fatalf("unexpected first candidate: %+v", got[0])
	}
	if got[0].AvatarURL != server.URL+"/avatars/a.jpg" {
		t.Fatalf("unexpected first avatar: %s", got[0].AvatarURL)
	}
	if got[1].Name != "演员乙" {
		t.Fatalf("unexpected second name: %s", got[1].Name)
	}
}

func TestPreviewActorByNameTMDBNotesNotTruncated(t *testing.T) {
	fullBio := strings.Repeat("这是一段完整中文简介。", 120)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search/person":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": 2002},
				},
			})
		case "/person/2002":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":           2002,
				"name":         "演员乙",
				"gender":       1,
				"biography":    fullBio,
				"profile_path": "/actor-2.jpg",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "demo-key", server.URL, "", "", 2*time.Second)
	got, err := svc.PreviewActorByName(context.Background(), "演员乙", "tmdb", 10)
	if err != nil {
		t.Fatalf("PreviewActorByName returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got=%d", len(got))
	}

	if got[0].Notes != fullBio {
		t.Fatalf("expected full notes without truncation, got length=%d want=%d", len(got[0].Notes), len(fullBio))
	}
	if strings.Contains(got[0].Notes, "\uFFFD") {
		t.Fatalf("notes contains invalid replacement character: %q", got[0].Notes)
	}
}
