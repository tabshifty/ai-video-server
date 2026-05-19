package queue

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/services"
)

func TestBuildScrapeFailureDecisionMarksEpisodeAsTVPending(t *testing.T) {
	t.Parallel()

	err := &services.EpisodeAutoScrapeError{
		Stage:               services.EpisodeAutoScrapeStageCandidateAmbiguous,
		ParsedTitle:         "三体",
		ParsedSeasonNumber:  1,
		ParsedEpisodeNumber: 2,
		CandidateCount:      2,
		CandidatePreview: []map[string]any{
			{"tmdb_id": 1, "title": "三体"},
			{"tmdb_id": 2, "title": "三体"},
		},
		Err: errors.New("ambiguous tv candidate"),
	}

	decision := buildScrapeFailureDecision("episode", err)

	if decision.status != "tv_pending" {
		t.Fatalf("expected status tv_pending, got=%s", decision.status)
	}
	if decision.enqueueTranscode {
		t.Fatal("expected tv_pending decision to skip transcode enqueue")
	}
	if decision.metadata["scrape_stage"] != services.EpisodeAutoScrapeStageCandidateAmbiguous {
		t.Fatalf("unexpected scrape_stage: %v", decision.metadata["scrape_stage"])
	}
	if decision.metadata["parsed_title"] != "三体" {
		t.Fatalf("unexpected parsed_title: %v", decision.metadata["parsed_title"])
	}
	if decision.metadata["parsed_season_number"] != 1 {
		t.Fatalf("unexpected parsed_season_number: %v", decision.metadata["parsed_season_number"])
	}
	if decision.metadata["parsed_episode_number"] != 2 {
		t.Fatalf("unexpected parsed_episode_number: %v", decision.metadata["parsed_episode_number"])
	}
	if decision.metadata["candidate_count"] != 2 {
		t.Fatalf("unexpected candidate_count: %v", decision.metadata["candidate_count"])
	}
}

func TestBuildScrapeFailureDecisionKeepsMovieFallbackBehavior(t *testing.T) {
	t.Parallel()

	decision := buildScrapeFailureDecision("movie", errors.New("no movie candidate"))

	if decision.status != "uploaded" {
		t.Fatalf("expected status uploaded, got=%s", decision.status)
	}
	if !decision.enqueueTranscode {
		t.Fatal("expected movie failure to continue into transcode")
	}
	if decision.metadata["scrape_error"] != "no movie candidate" {
		t.Fatalf("unexpected scrape_error: %v", decision.metadata["scrape_error"])
	}
	if _, ok := decision.metadata["scrape_stage"]; ok {
		t.Fatalf("movie fallback should not add scrape_stage, got=%v", decision.metadata["scrape_stage"])
	}
}

type queueRetagAVTestRepo struct {
	videoByID         map[uuid.UUID]models.Video
	lastStatus        string
	lastTitle         string
	lastDescription   string
	lastMetadata      map[string]any
	lastThumbnailPath string
}

func (r *queueRetagAVTestRepo) GetVideoByID(_ context.Context, videoID uuid.UUID) (models.Video, error) {
	if r.videoByID != nil {
		if video, ok := r.videoByID[videoID]; ok {
			if video.ID == uuid.Nil {
				video.ID = videoID
			}
			return video, nil
		}
	}
	return models.Video{ID: videoID}, nil
}

func (r *queueRetagAVTestRepo) UpdateVideoScrapeResult(_ context.Context, _ uuid.UUID, _ *int, title, description, thumbnailPath string, metadata map[string]any, status string) error {
	r.lastStatus = status
	r.lastTitle = title
	r.lastDescription = description
	r.lastThumbnailPath = thumbnailPath
	r.lastMetadata = metadata
	return nil
}

func (r *queueRetagAVTestRepo) UpsertSeries(context.Context, int, string, string, string, string, *time.Time, int, int) (int64, error) {
	return 0, nil
}

func (r *queueRetagAVTestRepo) UpsertSeason(context.Context, int64, int, string, string, string, *time.Time) (int64, error) {
	return 0, nil
}

func (r *queueRetagAVTestRepo) UpsertEpisode(context.Context, int64, int, string, string, string, int, *time.Time, uuid.UUID, bool) error {
	return nil
}

func (r *queueRetagAVTestRepo) FindVideoByTypeTMDB(context.Context, string, int, uuid.UUID) (uuid.UUID, bool, error) {
	return uuid.Nil, false, nil
}

func (r *queueRetagAVTestRepo) ResolveActorIDs(context.Context, []uuid.UUID, []string, string) ([]uuid.UUID, error) {
	return nil, nil
}

func (r *queueRetagAVTestRepo) AddVideoActors(context.Context, uuid.UUID, []uuid.UUID, string) error {
	return nil
}

func (r *queueRetagAVTestRepo) ListVideoActors(context.Context, uuid.UUID) ([]models.AdminVideoActor, error) {
	return nil, nil
}

func (r *queueRetagAVTestRepo) UpdateActorAvatar(context.Context, uuid.UUID, string, string, string) error {
	return nil
}

func (r *queueRetagAVTestRepo) UpsertScrapedActorProfile(_ context.Context, input models.AdminActorInput) (models.AdminActor, error) {
	return models.AdminActor{
		ID:         uuid.New(),
		Name:       input.Name,
		Aliases:    input.Aliases,
		Gender:     input.Gender,
		Country:    input.Country,
		BirthDate:  input.BirthDate,
		AvatarURL:  input.AvatarURL,
		Source:     input.Source,
		ExternalID: input.ExternalID,
		Notes:      input.Notes,
		Active:     input.Active,
	}, nil
}

func TestAutoScrapeAVMarksReadyOnSuccess(t *testing.T) {
	videoID := uuid.New()
	repo := &queueRetagAVTestRepo{
		videoByID: map[uuid.UUID]models.Video{
			videoID: {
				ID:           videoID,
				Title:        "SSIS-123",
				Status:       "scraping",
				OriginalPath: "/videos/SSIS-123.mp4",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/search":
			_, _ = w.Write([]byte(`
<html>
  <body>
    <a href="/v/ssis-123">
      <h2 class="title">SSIS-123 Retag Title</h2>
    </a>
  </body>
</html>
`))
		case r.URL.Path == "/v/ssis-123":
			_, _ = w.Write([]byte(`
<!doctype html>
<html>
  <body>
    <a class="button is-white copy-to-clipboard" data-clipboard-text="SSIS-123"></a>
    <h2 class="title is-4">SSIS-123 Retag Title</h2>
    <div><strong>番號:</strong><span>SSIS-123</span></div>
  </body>
</html>
`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := services.NewScraperService(repo, "", "https://api.themoviedb.org/3", t.TempDir(), "", time.Second)
	svc.ConfigureAVScraper(server.URL, "queue-retag-av-test", time.Second)

	processor := &Processor{scrape: svc}
	if err := processor.autoScrapeAV(context.Background(), videoID, "SSIS-123", repo.videoByID[videoID].OriginalPath); err != nil {
		t.Fatalf("autoScrapeAV returned error: %v", err)
	}
	if repo.lastStatus != "ready" {
		t.Fatalf("expected retag av to mark ready on success, got=%s", repo.lastStatus)
	}
}

func TestAutoScrapeAVStoresLocalizedFieldsAndMarksReady(t *testing.T) {
	videoID := uuid.New()
	repo := &queueRetagAVTestRepo{
		videoByID: map[uuid.UUID]models.Video{
			videoID: {
				ID:           videoID,
				Title:        "SSIS-123",
				Status:       "scraping",
				Type:         "av",
				OriginalPath: "/videos/SSIS-123.mp4",
			},
		},
	}

	var translationCalls int32
	translationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}
		atomic.AddInt32(&translationCalls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices":[{"message":{"content":"{\"title_zh\":\"中文标题\",\"description_zh\":\"中文简介\"}"}}]
		}`))
	}))
	defer translationServer.Close()

	avServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search":
			_, _ = w.Write([]byte(`
<html>
  <body>
    <a href="/v/ssis-123">
      <h2 class="title">SSIS-123 English Title</h2>
    </a>
  </body>
</html>
`))
		case "/v/ssis-123":
			_, _ = w.Write([]byte(`
<!doctype html>
<html>
  <body>
    <a class="button is-white copy-to-clipboard" data-clipboard-text="SSIS-123"></a>
    <h2 class="title is-4">SSIS-123 English Title</h2>
    <div><strong>番號:</strong><span>SSIS-123</span></div>
    <meta name="description" content="English overview" />
  </body>
</html>
`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer avServer.Close()

	svc := services.NewScraperService(repo, "", "https://api.themoviedb.org/3", t.TempDir(), "", time.Second)
	svc.ConfigureAVScraper(avServer.URL, "queue-retag-av-translation-test", time.Second)
	svc.ConfigureContentTranslation(services.TranslationConfig{
		APIURL:  translationServer.URL + "/v1",
		Model:   "HY-MT1.5-1.8B",
		Timeout: time.Second,
	})

	processor := &Processor{scrape: svc}
	if err := processor.autoScrapeAV(context.Background(), videoID, "SSIS-123", repo.videoByID[videoID].OriginalPath); err != nil {
		t.Fatalf("autoScrapeAV returned error: %v", err)
	}
	if repo.lastStatus != "ready" {
		t.Fatalf("expected retag av to mark ready on success, got=%s", repo.lastStatus)
	}
	if atomic.LoadInt32(&translationCalls) == 0 {
		t.Fatal("expected translation service to be called")
	}
	if repo.lastTitle != "中文标题" {
		t.Fatalf("unexpected translated title: %s", repo.lastTitle)
	}
	if repo.lastDescription != "中文简介" {
		t.Fatalf("unexpected translated description: %s", repo.lastDescription)
	}
	if repo.lastMetadata["title_original"] != "SSIS-123 English Title" {
		t.Fatalf("unexpected title_original: %v", repo.lastMetadata["title_original"])
	}
	if repo.lastMetadata["description_original"] != "English overview" {
		t.Fatalf("unexpected description_original: %v", repo.lastMetadata["description_original"])
	}
	if repo.lastMetadata["title_zh"] != "中文标题" {
		t.Fatalf("unexpected title_zh: %v", repo.lastMetadata["title_zh"])
	}
	if repo.lastMetadata["description_zh"] != "中文简介" {
		t.Fatalf("unexpected description_zh: %v", repo.lastMetadata["description_zh"])
	}
}

func TestAutoScrapeAVUsesTitleToSelectSite(t *testing.T) {
	t.Parallel()

	videoID := uuid.New()
	repo := &queueRetagAVTestRepo{
		videoByID: map[uuid.UUID]models.Video{
			videoID: {
				ID:           videoID,
				Title:        "Brazzers office affair",
				Status:       "scraping",
				Type:         "av",
				OriginalPath: "/videos/FC2-PPV-123456.mp4",
			},
		},
	}

	var fc2Hits int32
	var searchHits int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/scenes") && r.URL.Query().Get("parse") != "":
			atomic.AddInt32(&searchHits, 1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"slug":"brazzers-office-affair"}]}`))
		case r.URL.Path == "/scenes/brazzers-office-affair":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"data": {
					"slug": "brazzers-office-affair",
					"title": "Brazzers Office Affair",
					"description": "Western synopsis",
					"date": "2024-01-02",
					"performers": []
				}
			}`))
		case strings.HasPrefix(r.URL.Path, "/article/"):
			atomic.AddInt32(&fc2Hits, 1)
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := services.NewScraperService(repo, "", server.URL, t.TempDir(), "", time.Second)
	svc.ConfigureAVScraperConfig(services.AVScraperConfig{
		BaseURL:           server.URL,
		UserAgent:         "queue-retag-av-test",
		Timeout:           time.Second,
		SiteURLs:          map[string]string{"theporndb": server.URL, "fc2": server.URL},
		ThePornDBAPIToken: "token-123",
	})

	processor := &Processor{scrape: svc}
	if err := processor.autoScrapeAV(context.Background(), videoID, "Brazzers office affair", repo.videoByID[videoID].OriginalPath); err != nil {
		t.Fatalf("autoScrapeAV returned error: %v", err)
	}
	if got := atomic.LoadInt32(&searchHits); got == 0 {
		t.Fatalf("expected theporndb search to be used")
	}
	if got := atomic.LoadInt32(&fc2Hits); got != 0 {
		t.Fatalf("expected fc2 crawler to stay unused, hits=%d", got)
	}
	if repo.lastStatus != "ready" {
		t.Fatalf("expected retag av to mark ready on success, got=%s", repo.lastStatus)
	}
	if repo.lastMetadata["site_category"] != "western" {
		t.Fatalf("expected western site category, got=%v", repo.lastMetadata["site_category"])
	}
	if repo.lastMetadata["scrape_source"] != "theporndb" {
		t.Fatalf("expected theporndb scrape source, got=%v", repo.lastMetadata["scrape_source"])
	}
}
