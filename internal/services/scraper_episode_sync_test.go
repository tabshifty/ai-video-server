package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestScrapeEpisodeUploadSyncsWholeSeriesAndBindsOnlyTargetEpisode(t *testing.T) {
	t.Parallel()

	var baseURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search/tv":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": 88},
				},
			})
		case "/tv/88":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                 88,
				"name":               "中文剧名",
				"original_name":      "中文剧名",
				"overview":           "整剧简介",
				"first_air_date":     "2020-01-01",
				"number_of_seasons":  2,
				"number_of_episodes": 3,
				"poster_path":        baseURL + "/images/show.jpg",
				"seasons": []map[string]any{
					{"season_number": 1},
					{"season_number": 2},
				},
			})
		case "/tv/88/season/1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":     "第一季",
				"overview": "第一季简介",
				"air_date": "2020-01-01",
				"episodes": []map[string]any{
					{
						"id":             5001,
						"episode_number": 1,
						"name":           "第一集",
						"overview":       "第一集简介",
						"runtime":        45,
						"air_date":       "2020-01-01",
						"still_path":     baseURL + "/images/s1e1.jpg",
					},
					{
						"id":             5002,
						"episode_number": 2,
						"name":           "第二集",
						"overview":       "第二集简介",
						"runtime":        46,
						"air_date":       "2020-01-08",
						"still_path":     baseURL + "/images/s1e2.jpg",
					},
				},
			})
		case "/tv/88/season/2":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":     "第二季",
				"overview": "第二季简介",
				"air_date": "2021-01-01",
				"episodes": []map[string]any{
					{
						"id":             6001,
						"episode_number": 1,
						"name":           "第二季第一集",
						"overview":       "第二季第一集简介",
						"runtime":        44,
						"air_date":       "2021-01-01",
						"still_path":     baseURL + "/images/s2e1.jpg",
					},
				},
			})
		case "/images/show.jpg", "/images/s1e1.jpg", "/images/s1e2.jpg", "/images/s2e1.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("fake-image"))
		case "/tv/88/season/1/episode/2/credits", "/tv/88/credits":
			_ = json.NewEncoder(w).Encode(map[string]any{"cast": []map[string]any{}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	baseURL = server.URL

	repo := &fakeScraperRepo{}
	svc := NewScraperService(repo, "demo-key", server.URL, t.TempDir(), "", 2*time.Second)

	videoID := uuid.New()
	got, err := svc.ScrapeEpisodeUpload(context.Background(), videoID, "/tmp/中文剧名.S01E02.mkv", "中文剧名.S01E02.mkv")
	if err != nil {
		t.Fatalf("ScrapeEpisodeUpload returned error: %v", err)
	}

	if got.TMDBID != 5002 {
		t.Fatalf("expected tmdb id 5002, got=%d", got.TMDBID)
	}
	if len(repo.episodeUpserts) != 3 {
		t.Fatalf("expected 3 episode upserts, got=%d", len(repo.episodeUpserts))
	}

	var boundCount int
	for _, upsert := range repo.episodeUpserts {
		if upsert.bindVideo {
			boundCount++
			if upsert.episodeNumber != 2 {
				t.Fatalf("expected only target episode bound, got episode=%d", upsert.episodeNumber)
			}
			if upsert.videoID != videoID {
				t.Fatalf("expected bound video id %s, got=%s", videoID, upsert.videoID)
			}
			continue
		}
		if upsert.videoID != uuid.Nil {
			t.Fatalf("expected placeholder episodes to keep nil video id, got=%s", upsert.videoID)
		}
	}
	if boundCount != 1 {
		t.Fatalf("expected exactly one bound episode, got=%d", boundCount)
	}
}

func TestScrapeEpisodeUploadDownloadsSeriesArtworkLocally(t *testing.T) {
	t.Parallel()

	var baseURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search/tv":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": 88},
				},
			})
		case "/tv/88":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                 88,
				"name":               "中文剧名",
				"original_name":      "中文剧名",
				"overview":           "整剧简介",
				"first_air_date":     "2020-01-01",
				"number_of_seasons":  1,
				"number_of_episodes": 1,
				"poster_path":        baseURL + "/images/show-poster.jpg",
				"backdrop_path":      baseURL + "/images/show-backdrop.jpg",
				"seasons": []map[string]any{
					{"season_number": 1},
				},
			})
		case "/tv/88/season/1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":     "第一季",
				"overview": "第一季简介",
				"air_date": "2020-01-01",
				"episodes": []map[string]any{
					{
						"id":             5002,
						"episode_number": 2,
						"name":           "第二集",
						"overview":       "第二集简介",
						"runtime":        46,
						"air_date":       "2020-01-08",
						"still_path":     baseURL + "/images/s1e2.jpg",
					},
				},
			})
		case "/images/show-poster.jpg", "/images/show-backdrop.jpg", "/images/s1e2.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("fake-image"))
		case "/tv/88/season/1/episode/2/credits", "/tv/88/credits":
			_ = json.NewEncoder(w).Encode(map[string]any{"cast": []map[string]any{}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	baseURL = server.URL

	storageRoot := t.TempDir()
	svc := NewScraperService(&fakeScraperRepo{}, "demo-key", server.URL, storageRoot, "", 2*time.Second)

	_, err := svc.ScrapeEpisodeUpload(context.Background(), uuid.New(), "/tmp/中文剧名.S01E02.mkv", "中文剧名.S01E02.mkv")
	if err != nil {
		t.Fatalf("ScrapeEpisodeUpload returned error: %v", err)
	}

	for _, target := range []string{
		filepath.Join(storageRoot, "tv", "series", "1", "poster.jpg"),
		filepath.Join(storageRoot, "tv", "series", "1", "backdrop.jpg"),
		filepath.Join(storageRoot, "tv", "series", "1", "episodes", "s01e02.jpg"),
	} {
		if _, statErr := os.Stat(target); statErr != nil {
			t.Fatalf("expected series artwork downloaded to %s, stat err=%v", target, statErr)
		}
	}
}

func TestScrapeEpisodeUploadSkipsExistingSeriesArtworkOnNextEpisode(t *testing.T) {
	t.Parallel()

	var baseURL string
	var mu sync.Mutex
	imageRequests := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search/tv":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": 88},
				},
			})
		case "/tv/88":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                 88,
				"name":               "中文剧名",
				"original_name":      "中文剧名",
				"overview":           "整剧简介",
				"first_air_date":     "2020-01-01",
				"number_of_seasons":  1,
				"number_of_episodes": 3,
				"poster_path":        baseURL + "/images/show-poster.jpg",
				"backdrop_path":      baseURL + "/images/show-backdrop.jpg",
				"seasons": []map[string]any{
					{"season_number": 1},
				},
			})
		case "/tv/88/season/1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":     "第一季",
				"overview": "第一季简介",
				"air_date": "2020-01-01",
				"episodes": []map[string]any{
					{
						"id":             5001,
						"episode_number": 1,
						"name":           "第一集",
						"overview":       "第一集简介",
						"runtime":        45,
						"air_date":       "2020-01-01",
						"still_path":     baseURL + "/images/s1e1.jpg",
					},
					{
						"id":             5002,
						"episode_number": 2,
						"name":           "第二集",
						"overview":       "第二集简介",
						"runtime":        46,
						"air_date":       "2020-01-08",
						"still_path":     baseURL + "/images/s1e2.jpg",
					},
					{
						"id":             5003,
						"episode_number": 3,
						"name":           "第三集",
						"overview":       "第三集简介",
						"runtime":        47,
						"air_date":       "2020-01-15",
						"still_path":     baseURL + "/images/s1e3.jpg",
					},
				},
			})
		case "/images/show-poster.jpg", "/images/show-backdrop.jpg", "/images/s1e1.jpg", "/images/s1e2.jpg", "/images/s1e3.jpg":
			mu.Lock()
			imageRequests[r.URL.Path]++
			mu.Unlock()
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("fake-image"))
		case "/tv/88/season/1/episode/1/credits", "/tv/88/season/1/episode/2/credits", "/tv/88/credits":
			_ = json.NewEncoder(w).Encode(map[string]any{"cast": []map[string]any{}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	baseURL = server.URL

	storageRoot := t.TempDir()
	svc := NewScraperService(&fakeScraperRepo{}, "demo-key", server.URL, storageRoot, "", 2*time.Second)

	if _, err := svc.ScrapeEpisodeUpload(context.Background(), uuid.New(), "/tmp/中文剧名.S01E01.mkv", "中文剧名.S01E01.mkv"); err != nil {
		t.Fatalf("first ScrapeEpisodeUpload returned error: %v", err)
	}
	if _, err := svc.ScrapeEpisodeUpload(context.Background(), uuid.New(), "/tmp/中文剧名.S01E02.mkv", "中文剧名.S01E02.mkv"); err != nil {
		t.Fatalf("second ScrapeEpisodeUpload returned error: %v", err)
	}

	for _, path := range []string{
		"/images/show-poster.jpg",
		"/images/show-backdrop.jpg",
		"/images/s1e3.jpg",
	} {
		if got := imageRequests[path]; got != 1 {
			t.Fatalf("expected image %s to be requested once, got=%d requests=%v", path, got, imageRequests)
		}
	}
	for _, path := range []string{"/images/s1e1.jpg", "/images/s1e2.jpg"} {
		if got := imageRequests[path]; got != 2 {
			t.Fatalf("expected target still %s once for video thumbnail and once for series still, got=%d requests=%v", path, got, imageRequests)
		}
	}
}

func TestScrapeEpisodeUploadReturnsPendingErrorWhenFilenameCannotBeParsed(t *testing.T) {
	t.Parallel()

	svc := NewScraperService(&fakeScraperRepo{}, "demo-key", "https://example.invalid", t.TempDir(), "", time.Second)

	_, err := svc.ScrapeEpisodeUpload(context.Background(), uuid.New(), "/tmp/bad-name.mkv", "bad-name.mkv")
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}

	var pendingErr *EpisodeAutoScrapeError
	if !errors.As(err, &pendingErr) {
		t.Fatalf("expected EpisodeAutoScrapeError, got=%T %v", err, err)
	}
	if pendingErr.Stage != EpisodeAutoScrapeStageParseFailed {
		t.Fatalf("expected stage parse_failed, got=%s", pendingErr.Stage)
	}
	if pendingErr.ParsedTitle != "" || pendingErr.ParsedSeasonNumber != 0 || pendingErr.ParsedEpisodeNumber != 0 {
		t.Fatalf("expected empty parsed fields, got title=%q season=%d episode=%d", pendingErr.ParsedTitle, pendingErr.ParsedSeasonNumber, pendingErr.ParsedEpisodeNumber)
	}
}

func TestScrapeEpisodeUploadReturnsPendingErrorWhenCandidateIsAmbiguous(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search/tv":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": 88},
					{"id": 89},
				},
			})
		case "/tv/88":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":             88,
				"name":           "迷雾剧场",
				"original_name":  "迷雾剧场",
				"overview":       "版本一",
				"first_air_date": "2020-01-01",
			})
		case "/tv/89":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":             89,
				"name":           "迷雾剧场",
				"original_name":  "迷雾剧场",
				"overview":       "版本二",
				"first_air_date": "2022-01-01",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(&fakeScraperRepo{}, "demo-key", server.URL, t.TempDir(), "", 2*time.Second)

	_, err := svc.ScrapeEpisodeUpload(context.Background(), uuid.New(), "/tmp/迷雾剧场.S01E02.mkv", "迷雾剧场.S01E02.mkv")
	if err == nil {
		t.Fatal("expected ambiguous candidate error, got nil")
	}

	var pendingErr *EpisodeAutoScrapeError
	if !errors.As(err, &pendingErr) {
		t.Fatalf("expected EpisodeAutoScrapeError, got=%T %v", err, err)
	}
	if pendingErr.Stage != EpisodeAutoScrapeStageCandidateAmbiguous {
		t.Fatalf("expected stage candidate_ambiguous, got=%s", pendingErr.Stage)
	}
	if pendingErr.CandidateCount != 2 {
		t.Fatalf("expected candidate count 2, got=%d", pendingErr.CandidateCount)
	}
	if pendingErr.ParsedTitle != "迷雾剧场" || pendingErr.ParsedSeasonNumber != 1 || pendingErr.ParsedEpisodeNumber != 2 {
		t.Fatalf("unexpected parsed fields: %+v", pendingErr)
	}
	if len(pendingErr.CandidatePreview) != 2 {
		t.Fatalf("expected 2 candidate previews, got=%d", len(pendingErr.CandidatePreview))
	}
}

func TestScrapeEpisodeUploadAllowsRebindingEpisodeWithoutTMDBDuplicateGuard(t *testing.T) {
	t.Parallel()

	var baseURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search/tv":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": 88},
				},
			})
		case "/tv/88":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                 88,
				"name":               "重绑定剧集",
				"original_name":      "重绑定剧集",
				"overview":           "简介",
				"first_air_date":     "2020-01-01",
				"number_of_seasons":  1,
				"number_of_episodes": 1,
				"poster_path":        baseURL + "/images/rebind-show.jpg",
			})
		case "/tv/88/season/1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":     "第一季",
				"overview": "简介",
				"air_date": "2020-01-01",
				"episodes": []map[string]any{
					{
						"id":             5002,
						"episode_number": 2,
						"name":           "第二集",
						"overview":       "第二集简介",
						"runtime":        46,
						"air_date":       "2020-01-08",
						"still_path":     baseURL + "/images/rebind-s1e2.jpg",
					},
				},
			})
		case "/images/rebind-show.jpg", "/images/rebind-s1e2.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("fake-image"))
		case "/tv/88/season/1/episode/2/credits", "/tv/88/credits":
			_ = json.NewEncoder(w).Encode(map[string]any{"cast": []map[string]any{}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	baseURL = server.URL

	repo := &fakeScraperRepo{
		findVideoExists: true,
		findVideoID:     uuid.New(),
	}
	svc := NewScraperService(repo, "demo-key", server.URL, t.TempDir(), "", 2*time.Second)

	_, err := svc.ScrapeEpisodeUpload(context.Background(), uuid.New(), "/tmp/重绑定剧集.S01E02.mkv", "重绑定剧集.S01E02.mkv")
	if err != nil {
		t.Fatalf("ScrapeEpisodeUpload returned error: %v", err)
	}
	for _, typ := range repo.findVideoTMDBCalls {
		if typ == "episode" {
			t.Fatalf("expected no episode tmdb duplicate guard call, got calls=%v", repo.findVideoTMDBCalls)
		}
	}
}
