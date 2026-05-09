package queue

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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
	videoByID  map[uuid.UUID]models.Video
	lastStatus string
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

func (r *queueRetagAVTestRepo) UpdateVideoScrapeResult(_ context.Context, _ uuid.UUID, _ *int, _ string, _ string, _ string, _ map[string]any, status string) error {
	r.lastStatus = status
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

func TestAutoScrapeAVPreservesCurrentStatus(t *testing.T) {
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
	if err := processor.autoScrapeAV(context.Background(), videoID, "SSIS-123", repo.videoByID[videoID].OriginalPath, repo.videoByID[videoID].Status); err != nil {
		t.Fatalf("autoScrapeAV returned error: %v", err)
	}
	if repo.lastStatus != "scraping" {
		t.Fatalf("expected retag av to preserve current status scraping, got=%s", repo.lastStatus)
	}
}
