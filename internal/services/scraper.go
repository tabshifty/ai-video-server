package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"video-server/internal/repository"
	"video-server/internal/utils"
)

// ScraperService handles TMDB search and metadata syncing.
type ScraperService struct {
	repo       *repository.VideoRepository
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewScraperService(repo *repository.VideoRepository, apiKey, baseURL string, timeout time.Duration) *ScraperService {
	return &ScraperService{
		repo:    repo,
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// SearchMovieCandidates searches TMDB by file title/year pattern.
func (s *ScraperService) SearchMovieCandidates(ctx context.Context, filePath string) ([]map[string]any, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("TMDB_API_KEY is empty")
	}

	filename := filepath.Base(filePath)
	title, year, ok := utils.ExtractTitleYear(filename)
	if !ok {
		title = strings.TrimSuffix(filename, filepath.Ext(filename))
	}

	query := url.Values{}
	query.Set("api_key", s.apiKey)
	query.Set("query", title)
	if year > 0 {
		query.Set("year", fmt.Sprintf("%d", year))
	}

	api := fmt.Sprintf("%s/search/movie?%s", s.baseURL, query.Encode())
	raw, err := s.getJSON(ctx, api)
	if err != nil {
		return nil, err
	}
	results, ok := raw["results"].([]any)
	if !ok {
		return []map[string]any{}, nil
	}

	out := make([]map[string]any, 0, len(results))
	for _, row := range results {
		item, ok := row.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, map[string]any{
			"id":           item["id"],
			"title":        item["title"],
			"overview":     item["overview"],
			"poster_path":  item["poster_path"],
			"release_date": item["release_date"],
		})
	}
	return out, nil
}

// SyncMovieMetadata fetches movie details and updates videos metadata JSONB.
func (s *ScraperService) SyncMovieMetadata(ctx context.Context, videoID uuid.UUID, tmdbID int) error {
	if s.apiKey == "" {
		return fmt.Errorf("TMDB_API_KEY is empty")
	}
	q := url.Values{}
	q.Set("api_key", s.apiKey)
	api := fmt.Sprintf("%s/movie/%d?%s", s.baseURL, tmdbID, q.Encode())
	raw, err := s.getJSON(ctx, api)
	if err != nil {
		return err
	}

	title, _ := raw["title"].(string)
	overview, _ := raw["overview"].(string)
	meta := map[string]any{
		"tmdb": raw,
	}
	return s.repo.UpdateVideoMetadata(ctx, videoID, title, overview, meta)
}

// SyncTVEpisode syncs tv metadata and binds season/episode to the given video.
func (s *ScraperService) SyncTVEpisode(ctx context.Context, videoID uuid.UUID, tmdbID, seasonNum, episodeNum int) error {
	if s.apiKey == "" {
		return fmt.Errorf("TMDB_API_KEY is empty")
	}
	q := url.Values{}
	q.Set("api_key", s.apiKey)

	tvAPI := fmt.Sprintf("%s/tv/%d?%s", s.baseURL, tmdbID, q.Encode())
	tvRaw, err := s.getJSON(ctx, tvAPI)
	if err != nil {
		return err
	}

	seriesID, err := s.repo.UpsertSeries(
		ctx,
		tmdbID,
		asString(tvRaw["name"]),
		asString(tvRaw["overview"]),
		asString(tvRaw["poster_path"]),
		asString(tvRaw["backdrop_path"]),
		parseDate(tvRaw["first_air_date"]),
		asInt(tvRaw["number_of_seasons"]),
		asInt(tvRaw["number_of_episodes"]),
	)
	if err != nil {
		return err
	}

	seasonAPI := fmt.Sprintf("%s/tv/%d/season/%d?%s", s.baseURL, tmdbID, seasonNum, q.Encode())
	seasonRaw, err := s.getJSON(ctx, seasonAPI)
	if err != nil {
		return err
	}

	seasonID, err := s.repo.UpsertSeason(ctx, seriesID, seasonNum, asString(seasonRaw["name"]), asString(seasonRaw["overview"]), asString(seasonRaw["poster_path"]), parseDate(seasonRaw["air_date"]))
	if err != nil {
		return err
	}

	episodes, _ := seasonRaw["episodes"].([]any)
	for _, ep := range episodes {
		epRaw, ok := ep.(map[string]any)
		if !ok || asInt(epRaw["episode_number"]) != episodeNum {
			continue
		}
		if err := s.repo.UpsertEpisode(ctx, seasonID, episodeNum, asString(epRaw["name"]), asString(epRaw["overview"]), asString(epRaw["still_path"]), asInt(epRaw["runtime"]), parseDate(epRaw["air_date"]), videoID); err != nil {
			return err
		}
		break
	}

	return nil
}

func (s *ScraperService) getJSON(ctx context.Context, endpoint string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tmdb request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("tmdb status=%d body=%s", resp.StatusCode, string(body))
	}

	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode tmdb json: %w", err)
	}
	return out, nil
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asInt(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	default:
		return 0
	}
}

func parseDate(v any) *time.Time {
	s, _ := v.(string)
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}
