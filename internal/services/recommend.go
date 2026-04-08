package services

import (
	"context"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
)

// RecommendService builds random and personalized feeds.
type RecommendService struct {
	repo *repository.VideoRepository
	rng  *rand.Rand
}

func NewRecommendService(repo *repository.VideoRepository) *RecommendService {
	return &RecommendService{
		repo: repo,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// RandomShortFeed returns random short videos.
func (s *RecommendService) RandomShortFeed(ctx context.Context, pageSize int) ([]models.RecommendedVideo, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return s.repo.RandomShorts(ctx, pageSize)
}

// Recommend returns personalized recommendations for a user.
func (s *RecommendService) Recommend(ctx context.Context, userID uuid.UUID, pageSize int) ([]models.RecommendedVideo, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	tagAffinity, err := s.repo.FetchUserTagAffinity(ctx, userID, time.Now().Add(-30*24*time.Hour), 30)
	if err != nil {
		return nil, err
	}

	if len(tagAffinity) == 0 {
		return s.repo.FetchHotVideos(ctx, pageSize)
	}

	tags := sortTagKeys(tagAffinity)
	candidates, err := s.repo.FetchCandidateVideos(ctx, userID, tags, pageSize*3)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return s.repo.FetchHotVideos(ctx, pageSize)
	}

	for i := range candidates {
		tagMatch := normalize(tagAffinity[bestTagGuess(candidates[i].Title, tags)])
		hot := s.rng.Float64()
		freshness := s.rng.Float64()
		randomness := s.rng.Float64()
		candidates[i].Score = calculateScore(tagMatch, hot, freshness, randomness)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	if len(candidates) > pageSize {
		candidates = candidates[:pageSize]
	}
	return candidates, nil
}

func sortTagKeys(m map[string]float64) []string {
	type pair struct {
		tag   string
		score float64
	}
	items := make([]pair, 0, len(m))
	for k, v := range m {
		items = append(items, pair{tag: k, score: v})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].score > items[j].score
	})
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item.score > 0 {
			out = append(out, item.tag)
		}
	}
	return out
}

func bestTagGuess(title string, tags []string) string {
	for _, tag := range tags {
		if tag != "" {
			return tag
		}
	}
	return ""
}

func normalize(v float64) float64 {
	if v <= 0 {
		return 0
	}
	if v > 100 {
		return 1
	}
	return v / 100
}

func calculateScore(tagMatch, hot, freshness, randomFactor float64) float64 {
	return 0.5*tagMatch + 0.3*hot + 0.1*freshness + 0.1*randomFactor
}
