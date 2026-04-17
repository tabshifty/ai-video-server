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

	"github.com/google/uuid"

	"video-server/internal/models"
)

type fakeScrapeUpdate struct {
	videoID       uuid.UUID
	tmdbID        *int
	title         string
	description   string
	thumbnailPath string
	metadata      map[string]any
	status        string
}

type fakeScraperRepo struct {
	lastUpdate fakeScrapeUpdate
}

func (f *fakeScraperRepo) GetVideoByID(context.Context, uuid.UUID) (models.Video, error) {
	return models.Video{}, nil
}

func (f *fakeScraperRepo) UpdateVideoScrapeResult(_ context.Context, videoID uuid.UUID, tmdbID *int, title, description, thumbnailPath string, metadata map[string]any, status string) error {
	f.lastUpdate = fakeScrapeUpdate{
		videoID:       videoID,
		tmdbID:        tmdbID,
		title:         title,
		description:   description,
		thumbnailPath: thumbnailPath,
		metadata:      metadata,
		status:        status,
	}
	return nil
}

func (f *fakeScraperRepo) UpsertSeries(context.Context, int, string, string, string, string, *time.Time, int, int) (int64, error) {
	return 1, nil
}

func (f *fakeScraperRepo) UpsertSeason(context.Context, int64, int, string, string, string, *time.Time) (int64, error) {
	return 1, nil
}

func (f *fakeScraperRepo) UpsertEpisode(context.Context, int64, int, string, string, string, int, *time.Time, uuid.UUID) error {
	return nil
}

func (f *fakeScraperRepo) FindVideoByTypeTMDB(context.Context, string, int, uuid.UUID) (uuid.UUID, bool, error) {
	return uuid.Nil, false, nil
}

func (f *fakeScraperRepo) ResolveActorIDs(context.Context, []uuid.UUID, []string, string) ([]uuid.UUID, error) {
	return nil, nil
}

func (f *fakeScraperRepo) AddVideoActors(context.Context, uuid.UUID, []uuid.UUID, string) error {
	return nil
}

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

func TestScrapeMovieUploadUsesChineseLanguageAndFallback(t *testing.T) {
	var searchLangs []string
	var detailLangs []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search/movie":
			searchLangs = append(searchLangs, r.URL.Query().Get("language"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": 301},
				},
			})
		case "/movie/301":
			lang := r.URL.Query().Get("language")
			detailLangs = append(detailLangs, lang)
			if lang == "zh-CN" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":           301,
					"title":        "中文片名",
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
			_ = json.NewEncoder(w).Encode(map[string]any{
				"cast": []map[string]any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	repo := &fakeScraperRepo{}
	svc := NewScraperService(repo, "demo-key", server.URL, t.TempDir(), "", 2*time.Second)

	videoID := uuid.New()
	got, err := svc.ScrapeMovieUpload(context.Background(), videoID, "/tmp/中文片名.2021.mkv", "中文片名.2021.mkv")
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
		t.Fatalf("expected chinese title, got=%s", got.Title)
	}
	if got.Description != "English overview" {
		t.Fatalf("expected english fallback overview, got=%s", got.Description)
	}
	if repo.lastUpdate.videoID != videoID {
		t.Fatalf("unexpected update video id: %s", repo.lastUpdate.videoID)
	}
	if repo.lastUpdate.status != "uploaded" {
		t.Fatalf("expected status uploaded, got=%s", repo.lastUpdate.status)
	}
	if repo.lastUpdate.tmdbID == nil || *repo.lastUpdate.tmdbID != 301 {
		t.Fatalf("unexpected tmdb id in update: %v", repo.lastUpdate.tmdbID)
	}
	tmdbRaw, ok := repo.lastUpdate.metadata["tmdb"].(map[string]any)
	if !ok {
		t.Fatalf("expected metadata.tmdb object, got=%T", repo.lastUpdate.metadata["tmdb"])
	}
	if asString(tmdbRaw["title"]) != "中文片名" {
		t.Fatalf("expected metadata.tmdb.title chinese, got=%v", tmdbRaw["title"])
	}
	if asString(tmdbRaw["overview"]) != "English overview" {
		t.Fatalf("expected metadata.tmdb.overview english fallback, got=%v", tmdbRaw["overview"])
	}
}

func TestScrapeEpisodeUploadUsesChineseLanguageAndFallback(t *testing.T) {
	var searchLangs []string
	var tvDetailLangs []string
	var seasonLangs []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search/tv":
			searchLangs = append(searchLangs, r.URL.Query().Get("language"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": 88},
				},
			})
		case "/tv/88":
			lang := r.URL.Query().Get("language")
			tvDetailLangs = append(tvDetailLangs, lang)
			if lang == "zh-CN" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":                 88,
					"name":               "",
					"original_name":      "中文剧名",
					"overview":           "中文剧简介",
					"first_air_date":     "2020-01-01",
					"number_of_seasons":  1,
					"number_of_episodes": 10,
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                 88,
				"name":               "English Show",
				"original_name":      "中文剧名",
				"overview":           "English show overview",
				"first_air_date":     "2020-01-01",
				"number_of_seasons":  1,
				"number_of_episodes": 10,
			})
		case "/tv/88/season/1":
			lang := r.URL.Query().Get("language")
			seasonLangs = append(seasonLangs, lang)
			if lang == "zh-CN" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"name":     "第一季",
					"overview": "",
					"air_date": "2020-02-01",
					"episodes": []map[string]any{
						{
							"episode_number": 2,
							"id":             5002,
							"name":           "",
							"overview":       "",
							"runtime":        45,
							"air_date":       "2020-02-02",
						},
					},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":     "Season 1",
				"overview": "Season overview EN",
				"air_date": "2020-02-01",
				"episodes": []map[string]any{
					{
						"episode_number": 2,
						"id":             5002,
						"name":           "Episode Two",
						"overview":       "Episode overview EN",
						"runtime":        45,
						"air_date":       "2020-02-02",
					},
				},
			})
		case "/tv/88/season/1/episode/2/credits":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"cast": []map[string]any{},
			})
		case "/tv/88/credits":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"cast": []map[string]any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	repo := &fakeScraperRepo{}
	svc := NewScraperService(repo, "demo-key", server.URL, t.TempDir(), "", 2*time.Second)

	videoID := uuid.New()
	got, err := svc.ScrapeEpisodeUpload(context.Background(), videoID, "/tmp/中文剧名.S01E02.mkv", "中文剧名.S01E02.mkv")
	if err != nil {
		t.Fatalf("ScrapeEpisodeUpload returned error: %v", err)
	}

	if !slices.Contains(searchLangs, "zh-CN") {
		t.Fatalf("expected search/tv to request language=zh-CN, got=%v", searchLangs)
	}
	if !slices.Contains(tvDetailLangs, "zh-CN") {
		t.Fatalf("expected tv detail to request language=zh-CN, got=%v", tvDetailLangs)
	}
	if !slices.Contains(tvDetailLangs, "") {
		t.Fatalf("expected fallback tv detail request without language, got=%v", tvDetailLangs)
	}
	if !slices.Contains(seasonLangs, "zh-CN") {
		t.Fatalf("expected season detail to request language=zh-CN, got=%v", seasonLangs)
	}
	if !slices.Contains(seasonLangs, "") {
		t.Fatalf("expected fallback season detail request without language, got=%v", seasonLangs)
	}

	if got.Title != "Episode Two" {
		t.Fatalf("expected episode title from fallback, got=%s", got.Title)
	}
	if got.Description != "Episode overview EN" {
		t.Fatalf("expected episode overview from fallback, got=%s", got.Description)
	}
	if repo.lastUpdate.videoID != videoID {
		t.Fatalf("unexpected update video id: %s", repo.lastUpdate.videoID)
	}
	if repo.lastUpdate.tmdbID == nil || *repo.lastUpdate.tmdbID != 5002 {
		t.Fatalf("unexpected tmdb id in update: %v", repo.lastUpdate.tmdbID)
	}
	tvRaw, ok := repo.lastUpdate.metadata["tmdb_tv"].(map[string]any)
	if !ok {
		t.Fatalf("expected metadata.tmdb_tv object, got=%T", repo.lastUpdate.metadata["tmdb_tv"])
	}
	if asString(tvRaw["name"]) != "English Show" {
		t.Fatalf("expected metadata.tmdb_tv.name fallback, got=%v", tvRaw["name"])
	}
	epRaw, ok := repo.lastUpdate.metadata["tmdb_episode"].(map[string]any)
	if !ok {
		t.Fatalf("expected metadata.tmdb_episode object, got=%T", repo.lastUpdate.metadata["tmdb_episode"])
	}
	if asString(epRaw["name"]) != "Episode Two" {
		t.Fatalf("expected metadata.tmdb_episode.name fallback, got=%v", epRaw["name"])
	}
}
