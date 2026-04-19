package repository

import (
	"testing"

	"github.com/google/uuid"

	"video-server/internal/models"
)

func TestMergeRecommendedVideos(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()

	primary := []models.RecommendedVideo{
		{ID: id1, Title: "A"},
	}
	fallback := []models.RecommendedVideo{
		{ID: id1, Title: "A-dup"},
		{ID: id2, Title: "B"},
		{ID: id3, Title: "C"},
	}

	got := mergeRecommendedVideos(primary, fallback, 3)
	if len(got) != 3 {
		t.Fatalf("expected 3 videos after merge, got %d", len(got))
	}
	if got[0].ID != id1 || got[1].ID != id2 || got[2].ID != id3 {
		t.Fatalf("unexpected merge order: %#v", got)
	}
}
