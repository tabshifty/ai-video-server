package queue

import (
	"errors"
	"testing"

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
