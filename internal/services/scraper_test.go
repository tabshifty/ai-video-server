package services

import (
	"context"
	"encoding/json"
	"io"
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

type fakeEpisodeUpsert struct {
	seasonID      int64
	episodeNumber int
	title         string
	overview      string
	stillPath     string
	runtime       int
	airDate       *time.Time
	videoID       uuid.UUID
	bindVideo     bool
}

type fakeScraperRepo struct {
	lastUpdate         fakeScrapeUpdate
	resolveActorNames  []string
	resolveActorSource string
	addedActorIDs      []uuid.UUID
	addedActorSource   string
	nextActorIDs       []uuid.UUID
	videoByID          map[uuid.UUID]models.Video
	getVideoErr        error
	episodeUpserts     []fakeEpisodeUpsert
	findVideoTMDBCalls []string
	findVideoExists    bool
	findVideoID        uuid.UUID
}

func (f *fakeScraperRepo) GetVideoByID(_ context.Context, videoID uuid.UUID) (models.Video, error) {
	if f.getVideoErr != nil {
		return models.Video{}, f.getVideoErr
	}
	if f.videoByID != nil {
		if video, ok := f.videoByID[videoID]; ok {
			if video.ID == uuid.Nil {
				video.ID = videoID
			}
			return video, nil
		}
	}
	return models.Video{ID: videoID}, nil
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

func (f *fakeScraperRepo) UpsertEpisode(_ context.Context, seasonID int64, episodeNumber int, title, overview, stillPath string, runtime int, airDate *time.Time, videoID uuid.UUID, bindVideo bool) error {
	f.episodeUpserts = append(f.episodeUpserts, fakeEpisodeUpsert{
		seasonID:      seasonID,
		episodeNumber: episodeNumber,
		title:         title,
		overview:      overview,
		stillPath:     stillPath,
		runtime:       runtime,
		airDate:       airDate,
		videoID:       videoID,
		bindVideo:     bindVideo,
	})
	return nil
}

func (f *fakeScraperRepo) FindVideoByTypeTMDB(_ context.Context, typ string, _ int, _ uuid.UUID) (uuid.UUID, bool, error) {
	f.findVideoTMDBCalls = append(f.findVideoTMDBCalls, typ)
	if f.findVideoExists {
		return f.findVideoID, true, nil
	}
	return uuid.Nil, false, nil
}

func (f *fakeScraperRepo) ResolveActorIDs(_ context.Context, _ []uuid.UUID, actorNames []string, source string) ([]uuid.UUID, error) {
	f.resolveActorNames = append([]string(nil), actorNames...)
	f.resolveActorSource = source
	if len(f.nextActorIDs) > 0 {
		return append([]uuid.UUID(nil), f.nextActorIDs...), nil
	}
	return nil, nil
}

func (f *fakeScraperRepo) AddVideoActors(_ context.Context, _ uuid.UUID, actorIDs []uuid.UUID, source string) error {
	f.addedActorIDs = append([]uuid.UUID(nil), actorIDs...)
	f.addedActorSource = source
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

func TestScrapeAVUploadCodeFirstAndActorSync(t *testing.T) {
	var searchKeywords []string
	var searchLocales []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search":
			searchKeywords = append(searchKeywords, r.URL.Query().Get("q"))
			searchLocales = append(searchLocales, r.URL.Query().Get("locale"))
			_, _ = w.Write([]byte(`
<html>
  <body>
    <a href="/v/code-hit"><span>ABP-123 作品标题</span></a>
    <a href="/v/title-hit"><span>RANDOM-999 备用标题</span></a>
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
		case "/v/title-hit":
			_, _ = w.Write([]byte(`
<html>
  <head>
    <title>RANDOM-999 备用标题 - JavDB</title>
    <meta property="og:image" content="/covers/random999.jpg" />
  </head>
  <body>
    <h2 class="title is-4">RANDOM-999 备用标题</h2>
  </body>
</html>
`))
		case "/covers/abp123.jpg", "/covers/random999.jpg":
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

	videoID := uuid.New()
	got, err := svc.ScrapeAVUpload(context.Background(), videoID, "/tmp/ABP-123.sample.mkv", "ABP-123.sample.mkv")
	if err != nil {
		t.Fatalf("ScrapeAVUpload returned error: %v", err)
	}

	if len(searchKeywords) == 0 || strings.ToUpper(searchKeywords[0]) != "ABP-123" {
		t.Fatalf("expected code-first search keyword ABP-123, got=%v", searchKeywords)
	}
	if len(searchLocales) == 0 || searchLocales[0] != "zh" {
		t.Fatalf("expected locale=zh for av search, got=%v", searchLocales)
	}
	if got.Title != "ABP-123 作品标题" {
		t.Fatalf("unexpected title: %s", got.Title)
	}
	if got.Description != "作品简介" {
		t.Fatalf("unexpected overview: %s", got.Description)
	}
	if repo.lastUpdate.tmdbID != nil {
		t.Fatalf("expected tmdb id nil for av scrape, got=%v", repo.lastUpdate.tmdbID)
	}
	if repo.lastUpdate.status != "uploaded" {
		t.Fatalf("expected status uploaded, got=%s", repo.lastUpdate.status)
	}
	if repo.resolveActorSource != "scrape_av" {
		t.Fatalf("expected resolve actor source scrape_av, got=%s", repo.resolveActorSource)
	}
	if !reflect.DeepEqual(repo.resolveActorNames, []string{"演员甲", "演员乙"}) {
		t.Fatalf("unexpected resolved actor names: %v", repo.resolveActorNames)
	}
	if repo.addedActorSource != "scrape_av" {
		t.Fatalf("expected add actor source scrape_av, got=%s", repo.addedActorSource)
	}
	if !reflect.DeepEqual(repo.addedActorIDs, repo.nextActorIDs) {
		t.Fatalf("unexpected added actor ids: %v", repo.addedActorIDs)
	}
	if repo.lastUpdate.metadata["scrape_source"] != "javdb" {
		t.Fatalf("expected scrape_source javdb, got=%v", repo.lastUpdate.metadata["scrape_source"])
	}
	if repo.lastUpdate.metadata["external_id"] != "code-hit" {
		t.Fatalf("expected external_id code-hit, got=%v", repo.lastUpdate.metadata["external_id"])
	}
	trace, ok := repo.lastUpdate.metadata["scrape_trace"].(map[string]any)
	if !ok {
		t.Fatalf("expected scrape_trace metadata object, got=%T", repo.lastUpdate.metadata["scrape_trace"])
	}
	if trace["source"] != "javdb" {
		t.Fatalf("expected scrape_trace.source=javdb, got=%v", trace["source"])
	}
	queries, ok := trace["search_queries"].([]string)
	if !ok || len(queries) == 0 || queries[0] != "ABP-123" {
		t.Fatalf("unexpected scrape_trace.search_queries: %v", trace["search_queries"])
	}
}

func TestPreviewAVFallbackByTitle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search":
			_, _ = w.Write([]byte(`
<html>
  <body>
    <a href="/v/title-only"><span>无番号作品</span></a>
  </body>
</html>
`))
		case "/v/title-only":
			_, _ = w.Write([]byte(`
<html>
  <head>
    <title>无番号作品 - JavDB</title>
    <meta property="og:image" content="/covers/title-only.jpg" />
    <meta name="description" content="标题检索简介" />
  </head>
  <body>
    <h2 class="title is-4">无番号作品</h2>
    <div><strong>日期:</strong><span>2023-11-09</span></div>
    <a href="/actors/a1">演员丙</a>
  </body>
</html>
`))
		case "/covers/title-only.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("fake-image"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "av-preview-test", time.Second)

	got, err := svc.PreviewAV(context.Background(), "无番号标题")
	if err != nil {
		t.Fatalf("PreviewAV returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected one av candidate, got=%d", len(got))
	}
	candidate := got[0]
	if candidate["title"] != "无番号作品" {
		t.Fatalf("unexpected title: %v", candidate["title"])
	}
	if candidate["media_type_hint"] != "av" {
		t.Fatalf("expected media_type_hint av, got=%v", candidate["media_type_hint"])
	}
	if candidate["release_date"] != "2023-11-09" {
		t.Fatalf("unexpected release date: %v", candidate["release_date"])
	}
	actors, ok := candidate["actors"].([]string)
	if !ok || len(actors) != 1 || actors[0] != "演员丙" {
		t.Fatalf("unexpected actors: %v", candidate["actors"])
	}
}

func TestExtractAVCodeVariants(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "standard code",
			input: "ABP-123 sample",
			want:  "ABP-123",
		},
		{
			name:  "fc2 ppv merged",
			input: "FC2PPV-123456 title",
			want:  "FC2-123456",
		},
		{
			name:  "fc2 with hyphen",
			input: "FC2-PPV-654321",
			want:  "FC2-654321",
		},
		{
			name:  "heyzo no hyphen",
			input: "HEYZO1234 trailer",
			want:  "HEYZO-1234",
		},
		{
			name:  "suren numeric prefix",
			input: "259LUXU-1456 uncensored",
			want:  "259LUXU-1456",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := extractAVCode(tc.input)
			if got != tc.want {
				t.Fatalf("extractAVCode(%q)=%q, want=%q", tc.input, got, tc.want)
			}
		})
	}
}

func TestPreviewAVJavDBPrefersCoverAndDetailOverview(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search":
			_, _ = w.Write([]byte(`
<html>
  <body>
    <a href="/v/logo-case"><span>ABP-456 示例标题</span></a>
  </body>
</html>
`))
		case "/v/logo-case":
			_, _ = w.Write([]byte(`
<html>
  <head>
    <title>ABP-456 示例标题 - JavDB</title>
    <meta property="og:image" content="/static/site-logo.png" />
    <meta name="description" content="JavDB 在线番号资源" />
  </head>
  <body>
    <h2 class="title is-4">ABP-456 示例标题</h2>
    <img class="video-cover" src="/covers/real-cover.jpg" />
    <div><strong>番號:</strong><span>ABP-456</span></div>
    <div><strong>日期:</strong><span>2024-04-05</span></div>
    <div><strong>簡介:</strong><span>这是正确的剧情简介</span></div>
    <span><strong>演員:</strong><span><a href="/actors/a1">演员甲</a><a href="/actors/a2">演员乙</a></span></span>
    <a href="/actors/rank">更多</a>
  </body>
</html>
`))
		case "/covers/real-cover.jpg", "/static/site-logo.png":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("fake-image"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "av-cover-test", time.Second)

	got, err := svc.PreviewAV(context.Background(), "ABP-456")
	if err != nil {
		t.Fatalf("PreviewAV returned error: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected av candidates, got none")
	}
	candidate := got[0]
	if candidate["poster_url"] != server.URL+"/covers/real-cover.jpg" {
		t.Fatalf("expected real cover, got=%v", candidate["poster_url"])
	}
	if candidate["overview"] != "这是正确的剧情简介" {
		t.Fatalf("expected detail overview, got=%v", candidate["overview"])
	}
	actors, ok := candidate["actors"].([]string)
	if !ok {
		t.Fatalf("expected actors []string, got=%T", candidate["actors"])
	}
	if !reflect.DeepEqual(actors, []string{"演员甲", "演员乙"}) {
		t.Fatalf("unexpected actors: %v", actors)
	}
}

func TestPreviewAVMergeUsesBestFieldsAcrossSources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/search":
			_, _ = w.Write([]byte(`
<html>
  <body>
    <a href="/v/javdb-hit"><span>ABP-999 合并样例</span></a>
  </body>
</html>
`))
		case r.URL.Path == "/search/ABP-999":
			_, _ = w.Write([]byte(`
<html>
  <body>
    <a class="movie-box" href="/abp-999"></a>
  </body>
</html>
`))
		case r.URL.Path == "/cn/vl_searchbyid.php":
			_, _ = w.Write([]byte(`<html><body></body></html>`))
		case r.URL.Path == "/v/javdb-hit":
			_, _ = w.Write([]byte(`
<html>
  <head>
    <title>ABP-999 合并样例 - JavDB</title>
    <meta property="og:image" content="/static/site-logo.png" />
    <meta name="description" content="JavDB 在线番号资源" />
  </head>
  <body>
    <h2 class="title is-4">ABP-999 合并样例</h2>
    <div><strong>番號:</strong><span>ABP-999</span></div>
    <div><strong>日期:</strong><span>2024-05-06</span></div>
  </body>
</html>
`))
		case r.URL.Path == "/abp-999":
			_, _ = w.Write([]byte(`
<html>
  <head>
    <title>ABP-999 合并样例 - JavBus</title>
    <meta name="description" content="更完整的剧情简介，应该用于合并结果" />
  </head>
  <body>
    <h3>ABP-999 合并样例</h3>
    <span class="header">識別碼:</span> ABP-999
    <a class="bigImage" href="/covers/javbus-cover.jpg"></a>
    <span class="header">發行日期:</span> 2024-05-06
    <div class="star-name"><a href="/star/a1">演员甲</a></div>
    <div class="star-name"><a href="/star/a2">演员乙</a></div>
  </body>
</html>
`))
		case r.URL.Path == "/covers/javbus-cover.jpg", r.URL.Path == "/static/site-logo.png":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("fake-image"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "av-merge-test", time.Second)

	got, err := svc.PreviewAV(context.Background(), "ABP-999")
	if err != nil {
		t.Fatalf("PreviewAV returned error: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected av candidates, got none")
	}
	first := got[0]
	if first["poster_url"] != server.URL+"/covers/javbus-cover.jpg" {
		t.Fatalf("expected merged poster from javbus, got=%v", first["poster_url"])
	}
	actors, ok := first["actors"].([]string)
	if !ok {
		t.Fatalf("expected actors []string, got=%T", first["actors"])
	}
	if !reflect.DeepEqual(actors, []string{"演员甲", "演员乙"}) {
		t.Fatalf("unexpected merged actors: %v", actors)
	}
	metadata, ok := first["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("expected metadata object, got=%T", first["metadata"])
	}
	fieldSources, ok := metadata["field_sources"].(map[string]string)
	if !ok {
		t.Fatalf("expected field_sources map[string]string, got=%T", metadata["field_sources"])
	}
	if fieldSources["poster_url"] != "javbus" {
		t.Fatalf("expected poster_url from javbus, got=%v", fieldSources["poster_url"])
	}
	if fieldSources["actors"] != "javbus" {
		t.Fatalf("expected actors from javbus, got=%v", fieldSources["actors"])
	}
}

func TestPreviewAVFallsBackToAVSexWhenPrimarySitesHaveNoResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/search":
			_, _ = w.Write([]byte(`<html><body></body></html>`))
		case r.URL.Path == "/search/CAWD-582":
			_, _ = w.Write([]byte(`<html><body></body></html>`))
		case r.URL.Path == "/cn/vl_searchbyid.php":
			_, _ = w.Write([]byte(`<html><body></body></html>`))
		case r.URL.Path == "/tw/search" && strings.EqualFold(r.URL.Query().Get("query"), "cawd-582"):
			_, _ = w.Write([]byte(`
<html>
  <body>
    <ul class="grid">
      <li>
        <a href="/cn/video/detail/359635">
          <div><h4 class="truncate">CAWD-582 AVSex First Impression</h4></div>
          <div class="relative overflow-hidden rounded-t-md"><img src="/covers/avsex-cover.jpg" /></div>
        </a>
      </li>
    </ul>
  </body>
</html>
`))
		case r.URL.Path == "/cn/video/detail/359635":
			_, _ = w.Write([]byte(`
<html>
  <head>
    <title>AVSex First Impression</title>
  </head>
  <body>
    <span class="truncate p-2 text-primary font-bold dark:text-primary-200">CAWD-582 AVSex First Impression</span>
    <dd class="flex gap-2 flex-wrap"><a href="/actor/a1">演员甲</a></dd>
  </body>
</html>
`))
		case r.URL.Path == "/covers/avsex-cover.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("fake-image"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "avsex-preview-test", time.Second)

	got, err := svc.PreviewAV(context.Background(), "CAWD-582")
	if err != nil {
		t.Fatalf("PreviewAV returned error: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected av candidates, got none")
	}
	first := got[0]
	if first["scrape_source"] != "avsex" {
		t.Fatalf("expected scrape_source avsex, got=%v", first["scrape_source"])
	}
	if first["title"] != "AVSex First Impression" {
		t.Fatalf("expected avsex title, got=%v", first["title"])
	}
	if first["detail_url"] != server.URL+"/cn/video/detail/359635" {
		t.Fatalf("expected avsex detail_url, got=%v", first["detail_url"])
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestConfirmAVKeepsExistingThumbnailWhenOnlyFallbackPoster(t *testing.T) {
	videoID := uuid.New()
	repo := &fakeScraperRepo{
		videoByID: map[uuid.UUID]models.Video{
			videoID: {
				ID:            videoID,
				Title:         "旧标题",
				ThumbnailPath: "/keep/existing.jpg",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v/fallback-only":
			_, _ = w.Write([]byte(`
<html>
  <head>
    <title>ABP-111 标题 - JavDB</title>
    <meta property="og:image" content="/thumbs/fallback.jpg" />
  </head>
  <body>
    <h2 class="title is-4">ABP-111 标题</h2>
    <div><strong>番號:</strong><span>ABP-111</span></div>
  </body>
</html>
`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(repo, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "av-confirm-fallback-test", time.Second)

	err := svc.ConfirmAV(context.Background(), ConfirmScrapeInput{
		VideoID:     videoID,
		ExternalID:  "fallback-only",
		ReleaseDate: "2024-06-01",
		Metadata: map[string]any{
			"scrape_source": "javdb",
			"detail_url":    server.URL + "/v/fallback-only",
		},
	})
	if err != nil {
		t.Fatalf("ConfirmAV returned error: %v", err)
	}
	if repo.lastUpdate.thumbnailPath != "/keep/existing.jpg" {
		t.Fatalf("expected keep existing thumbnail, got=%s", repo.lastUpdate.thumbnailPath)
	}
	if repo.lastUpdate.metadata["poster_decision"] != "fallback_kept_old" {
		t.Fatalf("expected poster_decision fallback_kept_old, got=%v", repo.lastUpdate.metadata["poster_decision"])
	}
}

func TestConfirmAVUsesFallbackPosterWhenNoExistingThumbnail(t *testing.T) {
	videoID := uuid.New()
	repo := &fakeScraperRepo{
		videoByID: map[uuid.UUID]models.Video{
			videoID: {
				ID:    videoID,
				Title: "新视频",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v/fallback-only":
			_, _ = w.Write([]byte(`
<html>
  <head>
    <title>ABP-222 标题 - JavDB</title>
    <meta property="og:image" content="/thumbs/fallback.jpg" />
  </head>
  <body>
    <h2 class="title is-4">ABP-222 标题</h2>
    <div><strong>番號:</strong><span>ABP-222</span></div>
  </body>
</html>
`))
		case "/thumbs/fallback.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("fake-image"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(repo, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "av-confirm-fallback-no-existing-test", time.Second)

	err := svc.ConfirmAV(context.Background(), ConfirmScrapeInput{
		VideoID:     videoID,
		ExternalID:  "fallback-only",
		ReleaseDate: "2024-06-02",
		Metadata: map[string]any{
			"scrape_source": "javdb",
			"detail_url":    server.URL + "/v/fallback-only",
		},
	})
	if err != nil {
		t.Fatalf("ConfirmAV returned error: %v", err)
	}
	if strings.TrimSpace(repo.lastUpdate.thumbnailPath) == "" {
		t.Fatalf("expected thumbnail path downloaded from fallback poster")
	}
	if repo.lastUpdate.metadata["poster_decision"] != "fallback_used_no_existing" {
		t.Fatalf("expected poster_decision fallback_used_no_existing, got=%v", repo.lastUpdate.metadata["poster_decision"])
	}
}

func TestConfirmAVRejectsRelativePosterURLWithoutTMDBFallback(t *testing.T) {
	videoID := uuid.New()
	repo := &fakeScraperRepo{
		videoByID: map[uuid.UUID]models.Video{
			videoID: {
				ID:            videoID,
				Title:         "已有标题",
				ThumbnailPath: "/keep/existing.jpg",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v/relative-poster":
			_, _ = w.Write([]byte(`
<html>
  <head>
    <title>ABP-333 标题 - JavDB</title>
  </head>
  <body>
    <h2 class="title is-4">ABP-333 标题</h2>
    <div><strong>番號:</strong><span>ABP-333</span></div>
  </body>
</html>
`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(repo, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "av-confirm-relative-poster-test", time.Second)

	requested := make([]string, 0, 1)
	svc.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requested = append(requested, req.URL.String())
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("fake-image")),
				Request:    req,
			}, nil
		}),
	}

	err := svc.ConfirmAV(context.Background(), ConfirmScrapeInput{
		VideoID:     videoID,
		ExternalID:  "relative-poster",
		PosterURL:   "/relative-av.jpg",
		ReleaseDate: "2024-06-03",
		Metadata: map[string]any{
			"scrape_source": "javdb",
			"detail_url":    server.URL + "/v/relative-poster",
		},
	})
	if err != nil {
		t.Fatalf("ConfirmAV returned error: %v", err)
	}
	if len(requested) != 0 {
		t.Fatalf("expected no poster download request for relative av poster url, got=%v", requested)
	}
	if repo.lastUpdate.thumbnailPath != "/keep/existing.jpg" {
		t.Fatalf("expected keep existing thumbnail, got=%s", repo.lastUpdate.thumbnailPath)
	}
	if repo.lastUpdate.metadata["poster_decision"] != "invalid_keep_old" {
		t.Fatalf("expected poster_decision invalid_keep_old, got=%v", repo.lastUpdate.metadata["poster_decision"])
	}
}

func TestConfirmAVBuildsAVSexDetailURLFromSourceAndExternalID(t *testing.T) {
	videoID := uuid.New()
	repo := &fakeScraperRepo{
		videoByID: map[uuid.UUID]models.Video{
			videoID: {
				ID:    videoID,
				Title: "旧标题",
			},
		},
		nextActorIDs: []uuid.UUID{uuid.New()},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cn/video/detail/359635":
			_, _ = w.Write([]byte(`
<html>
  <head>
    <title>AVSex First Impression</title>
  </head>
  <body>
    <span class="truncate p-2 text-primary font-bold dark:text-primary-200">CAWD-582 AVSex First Impression</span>
    <dd class="flex gap-2 flex-wrap"><a href="/actor/a1">演员甲</a></dd>
  </body>
</html>
`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(repo, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "avsex-confirm-test", time.Second)

	err := svc.ConfirmAV(context.Background(), ConfirmScrapeInput{
		VideoID:    videoID,
		ExternalID: "359635",
		Metadata: map[string]any{
			"scrape_source": "avsex",
		},
	})
	if err != nil {
		t.Fatalf("ConfirmAV returned error: %v", err)
	}
	if repo.lastUpdate.title != "AVSex First Impression" {
		t.Fatalf("expected avsex title, got=%s", repo.lastUpdate.title)
	}
	if repo.lastUpdate.metadata["scrape_source"] != "avsex" {
		t.Fatalf("expected scrape_source avsex, got=%v", repo.lastUpdate.metadata["scrape_source"])
	}
	if repo.lastUpdate.metadata["external_id"] != "359635" {
		t.Fatalf("expected external_id 359635, got=%v", repo.lastUpdate.metadata["external_id"])
	}
	if repo.lastUpdate.metadata["detail_url"] != server.URL+"/cn/video/detail/359635" {
		t.Fatalf("expected avsex detail_url, got=%v", repo.lastUpdate.metadata["detail_url"])
	}
	if repo.lastUpdate.metadata["av_code"] != "CAWD-582" {
		t.Fatalf("expected av_code CAWD-582, got=%v", repo.lastUpdate.metadata["av_code"])
	}
	if repo.resolveActorSource != "scrape_av" {
		t.Fatalf("expected resolve actor source scrape_av, got=%s", repo.resolveActorSource)
	}
	if !reflect.DeepEqual(repo.resolveActorNames, []string{"演员甲"}) {
		t.Fatalf("unexpected resolved actor names: %v", repo.resolveActorNames)
	}
}

func TestConfirmAVBuildsMDCXDetailURLsForAdditionalSites(t *testing.T) {
	cases := []struct {
		name         string
		source       string
		externalID   string
		path         string
		responseBody string
		wantTitle    string
		wantCode     string
	}{
		{
			name:       "dmm",
			source:     "dmm",
			externalID: "ssis001",
			path:       "/mono/dvd/-/detail/=/cid=ssis001/",
			responseBody: `<!doctype html>
<html>
<head>
  <meta property="og:image" content="https://pics.dmm.co.jp/mono/movie/adult/ssis001/ssis001ps.jpg">
  <script type="application/ld+json">
  {
    "name":"DMM First Impression",
    "description":"DMM outline.",
    "brand":{"name":"S1"},
    "subjectOf":{"contentUrl":"https://video.example.test/ssis001.mp4","uploadDate":"2024-05-02","actor":[{"name":"Yua Mikami"}],"genre":["Drama"]}
  }
  </script>
</head>
<body>
  <h1 id="title">DMM First Impression</h1>
  <table><tr><th>Number</th><td>SSIS-001</td></tr></table>
</body>
</html>`,
			wantTitle: "DMM First Impression",
			wantCode:  "SSIS-001",
		},
		{
			name:       "mgstage",
			source:     "mgstage",
			externalID: "300MIUM-382",
			path:       "/product/product_detail/300MIUM-382/",
			responseBody: `<!doctype html>
<html><body>
  <div id="center_column"><div><h1>MGStage First Impression</h1></div></div>
  <table><tr><th>Number</th><td>300MIUM-382</td></tr></table>
</body></html>`,
			wantTitle: "MGStage First Impression",
			wantCode:  "300MIUM-382",
		},
		{
			name:       "prestige",
			source:     "prestige",
			externalID: "2e4a2de8-7275-4803-bb07-7585fd4f2ff3",
			path:       "/api/product/2e4a2de8-7275-4803-bb07-7585fd4f2ff3",
			responseBody: `{
  "title": "Prestige First Impression",
  "body": "Prestige outline.",
  "actress": [{"name": "Yua Mikami"}],
  "sku": [{"salesStartAt": "2024-07-01T00:00:00"}],
  "playTime": 160,
  "series": {"name": "Prestige Series"},
  "genre": [{"name": "Drama"}],
  "maker": {"name": "Prestige"},
  "label": {"name": "PRESTIGE LABEL"}
}`,
			wantTitle: "Prestige First Impression",
			wantCode:  "",
		},
		{
			name:       "xcity",
			source:     "xcity",
			externalID: "147036",
			path:       "/avod/detail/",
			responseBody: `<!doctype html>
<html><body>
  <span id="program_detail_title">XCity First Impression XC-1280</span>
  <span id="hinban">XC-1280</span>
</body></html>`,
			wantTitle: "XCity First Impression",
			wantCode:  "XC-1280",
		},
		{
			name:       "getchu",
			source:     "getchu",
			externalID: "1180483",
			path:       "/soft.phtml",
			responseBody: `<!doctype html>
<html><head><meta property="og:image" content="/images/covers/istu5391.jpg"></head>
<body>
  <h1 id="soft-title">Getchu First Impression</h1>
  <table><tr><td>Item Code</td><td>ISTU-5391</td></tr></table>
</body></html>`,
			wantTitle: "Getchu First Impression",
			wantCode:  "ISTU-5391",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			videoID := uuid.New()
			repo := &fakeScraperRepo{
				videoByID: map[uuid.UUID]models.Video{
					videoID: {ID: videoID, Title: "旧标题"},
				},
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch tc.name {
				case "xcity":
					if r.URL.Path == tc.path && r.URL.RawQuery == "id=147036" {
						_, _ = w.Write([]byte(tc.responseBody))
						return
					}
				case "getchu":
					if r.URL.Path == tc.path && r.URL.RawQuery == "id=1180483&gc=gc" {
						_, _ = w.Write([]byte(tc.responseBody))
						return
					}
				default:
					if r.URL.Path == tc.path {
						_, _ = w.Write([]byte(tc.responseBody))
						return
					}
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			svc := NewScraperService(repo, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
			svc.ConfigureAVScraperConfig(AVScraperConfig{
				BaseURL: server.URL,
				SiteURLs: map[string]string{
					tc.source: server.URL,
				},
				UserAgent: "mdcx-confirm-test",
				Timeout:   time.Second,
			})
			svc.httpClient = &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     make(http.Header),
						Body:       io.NopCloser(strings.NewReader("fake-image")),
						Request:    req,
					}, nil
				}),
			}

			err := svc.ConfirmAV(context.Background(), ConfirmScrapeInput{
				VideoID:    videoID,
				ExternalID: tc.externalID,
				Metadata: map[string]any{
					"scrape_source": tc.source,
				},
			})
			if err != nil {
				t.Fatalf("ConfirmAV returned error: %v", err)
			}
			if repo.lastUpdate.title != tc.wantTitle {
				t.Fatalf("expected title %q, got=%q", tc.wantTitle, repo.lastUpdate.title)
			}
			if repo.lastUpdate.metadata["scrape_source"] != tc.source {
				t.Fatalf("expected scrape_source %s, got=%v", tc.source, repo.lastUpdate.metadata["scrape_source"])
			}
			if repo.lastUpdate.metadata["av_code"] != tc.wantCode {
				t.Fatalf("expected av_code %s, got=%v", tc.wantCode, repo.lastUpdate.metadata["av_code"])
			}
		})
	}
}

func TestConfirmAVBuildsMDCXDetailURLsForThirdBatchSites(t *testing.T) {
	cases := []struct {
		name         string
		source       string
		externalID   string
		path         string
		responseBody string
		wantTitle    string
		wantCode     string
	}{
		{
			name:       "fc2club",
			source:     "fc2club",
			externalID: "fc2-ppv-1234567",
			path:       "/html/FC2-PPV-1234567.html",
			responseBody: `<!doctype html>
<html><body>
  <h1>FC2Club First Impression</h1>
  <div>番號：FC2-PPV-1234567</div>
</body></html>`,
			wantTitle: "FC2Club First Impression",
			wantCode:  "FC2-PPV-1234567",
		},
		{
			name:       "fc2hub",
			source:     "fc2hub",
			externalID: "1234567",
			path:       "/detail/1234567",
			responseBody: `<!doctype html>
<html><body>
  <h1>FC2Hub First Impression</h1>
  <div>番号: FC2-PPV-1234567</div>
</body></html>`,
			wantTitle: "FC2Hub First Impression",
			wantCode:  "FC2-PPV-1234567",
		},
		{
			name:       "fc2ppvdb",
			source:     "fc2ppvdb",
			externalID: "fc2-ppv-1234567",
			path:       "/articles/FC2-PPV-1234567",
			responseBody: `<!doctype html>
<html><body>
  <h1>FC2PPVDB First Impression</h1>
  <div>品番 FC2-PPV-1234567</div>
</body></html>`,
			wantTitle: "FC2PPVDB First Impression",
			wantCode:  "FC2-PPV-1234567",
		},
		{
			name:       "airav",
			source:     "airav",
			externalID: "fc2-ppv-1234567",
			path:       "/video/FC2-PPV-1234567",
			responseBody: `<!doctype html>
<html><body>
  <h1>AIRAV First Impression</h1>
  <div>番號: FC2-PPV-1234567</div>
</body></html>`,
			wantTitle: "AIRAV First Impression",
			wantCode:  "FC2-PPV-1234567",
		},
		{
			name:       "jav321",
			source:     "jav321",
			externalID: "ymdd00173",
			path:       "/video/ymdd00173",
			responseBody: `<!doctype html>
<html><body>
  <h3>JAV321 First Impression</h3>
  <div><b>番号</b><span>YMDD-173</span></div>
</body></html>`,
			wantTitle: "JAV321 First Impression",
			wantCode:  "YMDD-173",
		},
		{
			name:       "mywife",
			source:     "mywife",
			externalID: "1307",
			path:       "/teigaku/model/no/1307",
			responseBody: `<!doctype html>
<html>
<head>
  <title>No.1307 Mywife First Impression</title>
</head>
<body>
  <div class="modelsamplephototop">Mywife outline text.</div>
  <div class="modelwaku0"><img alt="Yua Mikami"></div>
  <video id="video" poster="https://cdn.mywife.test/movie/1307/topview.jpg" src="https://cdn.mywife.test/movie/1307/preview.mp4"></video>
</body>
</html>`,
			wantTitle: "Mywife First Impression",
			wantCode:  "Mywife No.1307",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			videoID := uuid.New()
			repo := &fakeScraperRepo{
				videoByID: map[uuid.UUID]models.Video{
					videoID: {ID: videoID, Title: "旧标题"},
				},
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == tc.path {
					_, _ = w.Write([]byte(tc.responseBody))
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			svc := NewScraperService(repo, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
			svc.ConfigureAVScraperConfig(AVScraperConfig{
				BaseURL: server.URL,
				SiteURLs: map[string]string{
					tc.source: server.URL,
				},
				UserAgent: "mdcx-confirm-third-batch-test",
				Timeout:   time.Second,
			})
			svc.httpClient = &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     make(http.Header),
						Body:       io.NopCloser(strings.NewReader("fake-image")),
						Request:    req,
					}, nil
				}),
			}

			err := svc.ConfirmAV(context.Background(), ConfirmScrapeInput{
				VideoID:    videoID,
				ExternalID: tc.externalID,
				Metadata: map[string]any{
					"scrape_source": tc.source,
				},
			})
			if err != nil {
				t.Fatalf("ConfirmAV returned error: %v", err)
			}
			if repo.lastUpdate.title != tc.wantTitle {
				t.Fatalf("expected title %q, got=%q", tc.wantTitle, repo.lastUpdate.title)
			}
			if repo.lastUpdate.metadata["scrape_source"] != tc.source {
				t.Fatalf("expected scrape_source %s, got=%v", tc.source, repo.lastUpdate.metadata["scrape_source"])
			}
			if repo.lastUpdate.metadata["external_id"] != strings.ToLower(tc.externalID) {
				t.Fatalf("expected external_id %s, got=%v", strings.ToLower(tc.externalID), repo.lastUpdate.metadata["external_id"])
			}
			if repo.lastUpdate.metadata["detail_url"] != server.URL+tc.path {
				t.Fatalf("expected detail_url %s, got=%v", server.URL+tc.path, repo.lastUpdate.metadata["detail_url"])
			}
			if repo.lastUpdate.metadata["av_code"] != tc.wantCode {
				t.Fatalf("expected av_code %s, got=%v", tc.wantCode, repo.lastUpdate.metadata["av_code"])
			}
		})
	}
}

func TestConfirmAVBuildsMDCXDetailURLsForFourthBatchSites(t *testing.T) {
	cases := []struct {
		name         string
		source       string
		siteURLKey   string
		externalID   string
		path         string
		responseBody string
		wantSource   string
		wantTitle    string
		wantCode     string
	}{
		{
			name:       "avsox",
			source:     "avsox",
			siteURLKey: "avsox",
			externalID: "abc123",
			path:       "/cn/movie/abc123",
			responseBody: `<!doctype html>
<html><body>
  <h3>AVSOX First Impression</h3>
  <div><span>识别码:</span><span>ABC-123</span></div>
</body></html>`,
			wantSource: "avsox",
			wantTitle:  "AVSOX First Impression",
			wantCode:   "ABC-123",
		},
		{
			name:       "freejavbt",
			source:     "freejavbt",
			siteURLKey: "freejavbt",
			externalID: "fc2-ppv-1234567",
			path:       "/detail/FC2-PPV-1234567",
			responseBody: `<!doctype html>
<html><body>
  <h1>FreeJAVBT First Impression</h1>
  <div>番号 FC2-PPV-1234567</div>
</body></html>`,
			wantSource: "freejavbt",
			wantTitle:  "FreeJAVBT First Impression",
			wantCode:   "FC2-PPV-1234567",
		},
		{
			name:       "madouqu",
			source:     "madouqu",
			siteURLKey: "madouqu",
			externalID: "54321",
			path:       "/archives/54321",
			responseBody: `<!doctype html>
<html><body>
  <h1>MadouQu First Impression</h1>
  <div>编号: MD-54321</div>
</body></html>`,
			wantSource: "madouqu",
			wantTitle:  "MadouQu First Impression",
			wantCode:   "MD-54321",
		},
		{
			name:       "mdtv",
			source:     "mdtv",
			siteURLKey: "mdtv",
			externalID: "98765",
			path:       "/video/98765",
			responseBody: `<!doctype html>
<html><body>
  <h1>MDTV First Impression</h1>
  <div>编号: MDTV-98765</div>
</body></html>`,
			wantSource: "mdtv.com",
			wantTitle:  "MDTV First Impression",
			wantCode:   "MDTV-98765",
		},
		{
			name:       "cnmdb",
			source:     "cnmdb",
			siteURLKey: "cnmdb",
			externalID: "112233",
			path:       "/video/112233",
			responseBody: `<!doctype html>
<html><body>
  <h1>CNMDB First Impression</h1>
  <div>编号: CNMDB-112233</div>
</body></html>`,
			wantSource: "cnmdb",
			wantTitle:  "CNMDB First Impression",
			wantCode:   "CNMDB-112233",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			videoID := uuid.New()
			repo := &fakeScraperRepo{
				videoByID: map[uuid.UUID]models.Video{
					videoID: {ID: videoID, Title: "旧标题"},
				},
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == tc.path {
					_, _ = w.Write([]byte(tc.responseBody))
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			svc := NewScraperService(repo, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
			svc.ConfigureAVScraperConfig(AVScraperConfig{
				BaseURL: server.URL,
				SiteURLs: map[string]string{
					tc.siteURLKey: server.URL,
				},
				UserAgent: "mdcx-confirm-fourth-batch-test",
				Timeout:   time.Second,
			})
			svc.httpClient = &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     make(http.Header),
						Body:       io.NopCloser(strings.NewReader("fake-image")),
						Request:    req,
					}, nil
				}),
			}

			err := svc.ConfirmAV(context.Background(), ConfirmScrapeInput{
				VideoID:    videoID,
				ExternalID: tc.externalID,
				Metadata: map[string]any{
					"scrape_source": tc.source,
				},
			})
			if err != nil {
				t.Fatalf("ConfirmAV returned error: %v", err)
			}
			if repo.lastUpdate.title != tc.wantTitle {
				t.Fatalf("expected title %q, got=%q", tc.wantTitle, repo.lastUpdate.title)
			}
			if repo.lastUpdate.metadata["scrape_source"] != tc.wantSource {
				t.Fatalf("expected scrape_source %s, got=%v", tc.wantSource, repo.lastUpdate.metadata["scrape_source"])
			}
			if repo.lastUpdate.metadata["detail_url"] != server.URL+tc.path {
				t.Fatalf("expected detail_url %s, got=%v", server.URL+tc.path, repo.lastUpdate.metadata["detail_url"])
			}
			if repo.lastUpdate.metadata["av_code"] != tc.wantCode {
				t.Fatalf("expected av_code %s, got=%v", tc.wantCode, repo.lastUpdate.metadata["av_code"])
			}
		})
	}
}

func TestConfirmAVBuildsMDCXDetailURLsForFifthBatchSites(t *testing.T) {
	cases := []struct {
		name         string
		source       string
		externalID   string
		path         string
		rawQuery     string
		responseBody string
		wantSource   string
		wantTitle    string
		wantCode     string
	}{
		{
			name:       "faleno",
			source:     "faleno",
			externalID: "fsdss564",
			path:       "/top/works/fsdss564/",
			responseBody: `<!doctype html>
<html><body>
  <h1>Faleno First Impression</h1>
  <div class="number">FSDSS-564</div>
</body></html>`,
			wantSource: "faleno",
			wantTitle:  "Faleno First Impression",
			wantCode:   "FSDSS-564",
		},
		{
			name:       "fantastica",
			source:     "fantastica",
			externalID: "fakwm-001",
			path:       "/items/detail/FAKWM-001",
			responseBody: `<!doctype html>
<html><body>
  <h1>Fantastica First Impression</h1>
  <div class="sku">FAKWM-001</div>
</body></html>`,
			wantSource: "fantastica",
			wantTitle:  "Fantastica First Impression",
			wantCode:   "FAKWM-001",
		},
		{
			name:       "giga",
			source:     "giga",
			externalID: "6841",
			path:       "/product/index.php",
			rawQuery:   "product_id=6841",
			responseBody: `<!doctype html>
<html><body>
  <h1>GIGA First Impression</h1>
  <div class="product-code">GHOV-28</div>
</body></html>`,
			wantSource: "giga",
			wantTitle:  "GIGA First Impression",
			wantCode:   "GHOV-28",
		},
		{
			name:       "javday",
			source:     "javday",
			externalID: "ssis-001",
			path:       "/videos/ssis-001/",
			responseBody: `<!doctype html>
<html><body>
  <h1>Javday First Impression</h1>
  <div>番号: SSIS-001</div>
</body></html>`,
			wantSource: "javday",
			wantTitle:  "Javday First Impression",
			wantCode:   "SSIS-001",
		},
		{
			name:       "kin8",
			source:     "kin8",
			externalID: "3681",
			path:       "/moviepages/3681/index.html",
			responseBody: `<!doctype html>
<html><body>
  <h1>Kin8 First Impression</h1>
  <div>KIN8-3681</div>
</body></html>`,
			wantSource: "kin8",
			wantTitle:  "Kin8 First Impression",
			wantCode:   "KIN8-3681",
		},
		{
			name:       "love6",
			source:     "love6",
			externalID: "NDI2Mw==",
			path:       "/albums/view/NDI2Mw==",
			responseBody: `<!doctype html>
<html><body>
  <h1>Love6 First Impression</h1>
  <div>番号: LOVE6-4263</div>
</body></html>`,
			wantSource: "love6",
			wantTitle:  "Love6 First Impression",
			wantCode:   "LOVE6-4263",
		},
		{
			name:       "lulubar",
			source:     "lulubar",
			externalID: "340460",
			path:       "/video/detail",
			rawQuery:   "id=340460",
			responseBody: `<!doctype html>
<html><body>
  <h1>Lulubar First Impression</h1>
  <div>番号: LULU-340460</div>
</body></html>`,
			wantSource: "lulubar",
			wantTitle:  "Lulubar First Impression",
			wantCode:   "LULU-340460",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			videoID := uuid.New()
			repo := &fakeScraperRepo{
				videoByID: map[uuid.UUID]models.Video{
					videoID: {ID: videoID, Title: "旧标题"},
				},
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == tc.path && r.URL.RawQuery == tc.rawQuery {
					_, _ = w.Write([]byte(tc.responseBody))
					return
				}
				if r.URL.Path == tc.path && tc.rawQuery == "" {
					_, _ = w.Write([]byte(tc.responseBody))
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			svc := NewScraperService(repo, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
			svc.ConfigureAVScraperConfig(AVScraperConfig{
				BaseURL: server.URL,
				SiteURLs: map[string]string{
					tc.source: server.URL,
				},
				UserAgent: "mdcx-confirm-fifth-batch-test",
				Timeout:   time.Second,
			})
			svc.httpClient = &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     make(http.Header),
						Body:       io.NopCloser(strings.NewReader("fake-image")),
						Request:    req,
					}, nil
				}),
			}

			err := svc.ConfirmAV(context.Background(), ConfirmScrapeInput{
				VideoID:    videoID,
				ExternalID: tc.externalID,
				Metadata: map[string]any{
					"scrape_source": tc.source,
				},
			})
			if err != nil {
				t.Fatalf("ConfirmAV returned error: %v", err)
			}
			if repo.lastUpdate.title != tc.wantTitle {
				t.Fatalf("expected title %q, got=%q", tc.wantTitle, repo.lastUpdate.title)
			}
			if repo.lastUpdate.metadata["scrape_source"] != tc.wantSource {
				t.Fatalf("expected scrape_source %s, got=%v", tc.wantSource, repo.lastUpdate.metadata["scrape_source"])
			}
			wantDetailURL := server.URL + tc.path
			if tc.rawQuery != "" {
				wantDetailURL += "?" + tc.rawQuery
			}
			if repo.lastUpdate.metadata["detail_url"] != wantDetailURL {
				t.Fatalf("expected detail_url %s, got=%v", wantDetailURL, repo.lastUpdate.metadata["detail_url"])
			}
			if repo.lastUpdate.metadata["av_code"] != tc.wantCode {
				t.Fatalf("expected av_code %s, got=%v", tc.wantCode, repo.lastUpdate.metadata["av_code"])
			}
		})
	}
}

func TestPreviewAVFallsBackToThePornDBWhenConfigured(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/search":
			_, _ = w.Write([]byte(`<html><body></body></html>`))
		case r.URL.Path == "/search/ABW-123":
			_, _ = w.Write([]byte(`<html><body></body></html>`))
		case r.URL.Path == "/cn/vl_searchbyid.php":
			_, _ = w.Write([]byte(`<html><body></body></html>`))
		case r.URL.Path == "/scenes" && r.URL.Query().Get("parse") == "ABW 123":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"slug":"abw-123-scene"}]}`))
		case r.URL.Path == "/scenes/abw-123-scene":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
  "data":{
    "slug":"abw-123-scene",
    "title":"ThePornDB First Impression",
    "description":"ThePornDB outline.",
    "date":"2024-10-01",
    "poster":"https://image.example/abw-123-poster.jpg",
    "performers":[{"name":"Yua Mikami","parent":{"extras":{"gender":"female"}}}],
    "tags":[{"name":"Drama"}],
    "site":{"name":"Prestige","network":{"name":"Prestige"}}
  }
}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraperConfig(AVScraperConfig{
		BaseURL: server.URL,
		SiteURLs: map[string]string{
			"theporndb": server.URL,
		},
		ThePornDBAPIToken: "token-123",
		ThePornDBNoHash:   true,
		UserAgent:         "tpdb-preview-test",
		Timeout:           time.Second,
	})

	got, err := svc.PreviewAV(context.Background(), "ABW-123")
	if err != nil {
		t.Fatalf("PreviewAV returned error: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected av candidates, got none")
	}
	first := got[0]
	if first["scrape_source"] != "theporndb" {
		t.Fatalf("expected scrape_source theporndb, got=%v", first["scrape_source"])
	}
	if first["title"] != "ThePornDB First Impression" {
		t.Fatalf("expected title ThePornDB First Impression, got=%v", first["title"])
	}
	if first["detail_url"] != server.URL+"/scenes/abw-123-scene" {
		t.Fatalf("expected theporndb detail url, got=%v", first["detail_url"])
	}
}

func TestPreviewAVFallsBackToGetchuWhenConfigured(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/search":
			_, _ = w.Write([]byte(`<html><body></body></html>`))
		case r.URL.Path == "/search/ISTU-5391":
			_, _ = w.Write([]byte(`<html><body></body></html>`))
		case r.URL.Path == "/cn/vl_searchbyid.php":
			_, _ = w.Write([]byte(`<html><body></body></html>`))
		case r.URL.Path == "/php/search.phtml" && r.URL.Query().Get("search_keyword") == "ISTU-5391":
			_, _ = w.Write([]byte(`<!doctype html><html><body><a class="blueb" href="../soft.phtml?id=1180483">Getchu Hit</a></body></html>`))
		case r.URL.Path == "/soft.phtml" && r.URL.RawQuery == "id=1180483&gc=gc":
			_, _ = w.Write([]byte(`<!doctype html>
<html><head><meta property="og:image" content="/images/covers/istu5391.jpg"></head>
<body>
  <h1 id="soft-title">Getchu First Impression</h1>
  <table><tr><td>Item Code</td><td>ISTU-5391</td></tr></table>
</body></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraperConfig(AVScraperConfig{
		BaseURL: server.URL,
		SiteURLs: map[string]string{
			"getchu": server.URL,
		},
		UserAgent: "getchu-preview-test",
		Timeout:   time.Second,
	})

	got, err := svc.PreviewAV(context.Background(), "ISTU-5391")
	if err != nil {
		t.Fatalf("PreviewAV returned error: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected av candidates, got none")
	}
	first := got[0]
	if first["scrape_source"] != "getchu" {
		t.Fatalf("expected scrape_source getchu, got=%v", first["scrape_source"])
	}
	if first["title"] != "Getchu First Impression" {
		t.Fatalf("expected title Getchu First Impression, got=%v", first["title"])
	}
	if first["detail_url"] != server.URL+"/soft.phtml?id=1180483&gc=gc" {
		t.Fatalf("expected getchu detail url, got=%v", first["detail_url"])
	}
}
