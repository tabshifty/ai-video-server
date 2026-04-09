package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"video-server/internal/repository"
	"video-server/internal/utils"
)

// ScraperService handles TMDB search and metadata syncing.
type ScraperService struct {
	repo         *repository.VideoRepository
	apiKey       string
	baseURL      string
	storageRoot  string
	posterRoot   string
	httpClient   *http.Client
	cacheTTL     time.Duration
	cacheMu      sync.RWMutex
	previewCache map[string]previewCacheEntry
}

type previewCacheEntry struct {
	ExpireAt   time.Time
	Candidates []map[string]any
}

func NewScraperService(repo *repository.VideoRepository, apiKey, baseURL, storageRoot, posterRoot string, timeout time.Duration) *ScraperService {
	return &ScraperService{
		repo:        repo,
		apiKey:      apiKey,
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		storageRoot: storageRoot,
		posterRoot:  posterRoot,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		cacheTTL:     5 * time.Minute,
		previewCache: map[string]previewCacheEntry{},
	}
}

type ScrapeResult struct {
	TMDBID        int
	Title         string
	Description   string
	ThumbnailPath string
	Metadata      map[string]any
}

type ConfirmScrapeInput struct {
	VideoID       uuid.UUID
	Type          string // movie | tv | episode
	TMDBID        int
	Title         string
	Overview      string
	PosterURL     string
	ReleaseDate   string
	Metadata      map[string]any
	SeasonNumber  int
	EpisodeNumber int
}

func (s *ScraperService) PreviewMovie(ctx context.Context, title string, year int) ([]map[string]any, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("TMDB_API_KEY is empty")
	}
	cacheKey := fmt.Sprintf("movie|%s|%d", normalizeCacheKey(title), year)
	if c, ok := s.getPreviewCache(cacheKey); ok {
		return c, nil
	}

	q := url.Values{}
	q.Set("api_key", s.apiKey)
	q.Set("query", strings.TrimSpace(title))
	if year > 0 {
		q.Set("year", fmt.Sprintf("%d", year))
	}
	raw, err := s.getJSON(ctx, fmt.Sprintf("%s/search/movie?%s", s.baseURL, q.Encode()))
	if err != nil {
		return nil, err
	}
	rows, _ := raw["results"].([]any)
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		item, ok := row.(map[string]any)
		if !ok {
			continue
		}
		id := asInt(item["id"])
		if id <= 0 {
			continue
		}
		detail, dErr := s.getJSON(ctx, fmt.Sprintf("%s/movie/%d?%s", s.baseURL, id, url.Values{"api_key": []string{s.apiKey}}.Encode()))
		if dErr != nil {
			detail = item
		}
		out = append(out, map[string]any{
			"tmdb_id":         id,
			"title":           asString(detail["title"]),
			"original_title":  asString(detail["original_title"]),
			"overview":        asString(detail["overview"]),
			"poster_path":     asString(detail["poster_path"]),
			"release_date":    asString(detail["release_date"]),
			"vote_average":    detail["vote_average"],
			"genres":          extractGenres(detail["genres"]),
			"metadata":        detail,
			"media_type_hint": "movie",
		})
	}
	s.setPreviewCache(cacheKey, out)
	return out, nil
}

func (s *ScraperService) PreviewTV(ctx context.Context, title string, year int) ([]map[string]any, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("TMDB_API_KEY is empty")
	}
	cacheKey := fmt.Sprintf("tv|%s|%d", normalizeCacheKey(title), year)
	if c, ok := s.getPreviewCache(cacheKey); ok {
		return c, nil
	}

	q := url.Values{}
	q.Set("api_key", s.apiKey)
	q.Set("query", strings.TrimSpace(title))
	raw, err := s.getJSON(ctx, fmt.Sprintf("%s/search/tv?%s", s.baseURL, q.Encode()))
	if err != nil {
		return nil, err
	}
	rows, _ := raw["results"].([]any)
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		item, ok := row.(map[string]any)
		if !ok {
			continue
		}
		id := asInt(item["id"])
		if id <= 0 {
			continue
		}
		detail, dErr := s.getJSON(ctx, fmt.Sprintf("%s/tv/%d?%s", s.baseURL, id, url.Values{"api_key": []string{s.apiKey}}.Encode()))
		if dErr != nil {
			detail = item
		}
		firstYear := 0
		if fd := asString(detail["first_air_date"]); len(fd) >= 4 {
			firstYear, _ = strconv.Atoi(fd[:4])
		}
		if year > 0 && firstYear > 0 && firstYear != year {
			continue
		}
		out = append(out, map[string]any{
			"tmdb_id":         id,
			"title":           asString(detail["name"]),
			"original_title":  asString(detail["original_name"]),
			"overview":        asString(detail["overview"]),
			"poster_path":     asString(detail["poster_path"]),
			"release_date":    asString(detail["first_air_date"]),
			"vote_average":    detail["vote_average"],
			"genres":          extractGenres(detail["genres"]),
			"metadata":        detail,
			"media_type_hint": "tv",
		})
	}
	s.setPreviewCache(cacheKey, out)
	return out, nil
}

func (s *ScraperService) ConfirmMovie(ctx context.Context, in ConfirmScrapeInput) error {
	if in.TMDBID <= 0 {
		return fmt.Errorf("tmdb_id is required")
	}
	video, err := s.repo.GetVideoByID(ctx, in.VideoID)
	if err != nil {
		return err
	}
	detail, err := s.getJSON(ctx, fmt.Sprintf("%s/movie/%d?%s", s.baseURL, in.TMDBID, url.Values{"api_key": []string{s.apiKey}}.Encode()))
	if err != nil {
		return err
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = asString(detail["title"])
	}
	overview := strings.TrimSpace(in.Overview)
	if overview == "" {
		overview = asString(detail["overview"])
	}
	posterURL := strings.TrimSpace(in.PosterURL)
	if posterURL == "" {
		posterURL = asString(detail["poster_path"])
	}
	thumbPath, err := s.DownloadPoster(ctx, posterURL, in.VideoID)
	if err != nil {
		return err
	}
	meta := map[string]any{
		"tmdb":         detail,
		"manual":       in.Metadata,
		"release_date": chooseStr(in.ReleaseDate, asString(detail["release_date"])),
	}
	return s.repo.UpdateVideoScrapeResult(ctx, video.ID, &in.TMDBID, title, overview, thumbPath, meta, "uploaded")
}

func (s *ScraperService) ConfirmEpisode(ctx context.Context, in ConfirmScrapeInput) error {
	if in.TMDBID <= 0 {
		return fmt.Errorf("tmdb_id is required")
	}
	video, err := s.repo.GetVideoByID(ctx, in.VideoID)
	if err != nil {
		return err
	}
	seasonNum := in.SeasonNumber
	episodeNum := in.EpisodeNumber
	if seasonNum <= 0 || episodeNum <= 0 {
		_, _, sn, en, ok := utils.ParseFilename(video.OriginalPath)
		if ok {
			if seasonNum <= 0 {
				seasonNum = sn
			}
			if episodeNum <= 0 {
				episodeNum = en
			}
		}
	}
	if seasonNum <= 0 || episodeNum <= 0 {
		return fmt.Errorf("season_number and episode_number are required for episode")
	}

	tvRaw, err := s.getJSON(ctx, fmt.Sprintf("%s/tv/%d?%s", s.baseURL, in.TMDBID, url.Values{"api_key": []string{s.apiKey}}.Encode()))
	if err != nil {
		return err
	}
	seriesID, err := s.repo.UpsertSeries(
		ctx,
		in.TMDBID,
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

	seasonRaw, err := s.getJSON(ctx, fmt.Sprintf("%s/tv/%d/season/%d?%s", s.baseURL, in.TMDBID, seasonNum, url.Values{"api_key": []string{s.apiKey}}.Encode()))
	if err != nil {
		return err
	}
	seasonID, err := s.repo.UpsertSeason(ctx, seriesID, seasonNum, asString(seasonRaw["name"]), asString(seasonRaw["overview"]), asString(seasonRaw["poster_path"]), parseDate(seasonRaw["air_date"]))
	if err != nil {
		return err
	}

	var epRaw map[string]any
	episodes, _ := seasonRaw["episodes"].([]any)
	for _, ep := range episodes {
		row, ok := ep.(map[string]any)
		if !ok {
			continue
		}
		if asInt(row["episode_number"]) == episodeNum {
			epRaw = row
			break
		}
	}
	if epRaw == nil {
		return fmt.Errorf("episode S%02dE%02d not found", seasonNum, episodeNum)
	}
	episodeTMDBID := asInt(epRaw["id"])
	if episodeTMDBID <= 0 {
		return fmt.Errorf("invalid episode tmdb id")
	}

	if err := s.repo.UpsertEpisode(ctx, seasonID, episodeNum, asString(epRaw["name"]), asString(epRaw["overview"]), asString(epRaw["still_path"]), asInt(epRaw["runtime"]), parseDate(epRaw["air_date"]), in.VideoID); err != nil {
		return err
	}
	title := chooseStr(in.Title, asString(epRaw["name"]))
	overview := chooseStr(in.Overview, asString(epRaw["overview"]))
	posterURL := chooseStr(in.PosterURL, chooseStr(asString(epRaw["still_path"]), asString(tvRaw["poster_path"])))
	thumbPath, err := s.DownloadPoster(ctx, posterURL, in.VideoID)
	if err != nil {
		return err
	}
	meta := map[string]any{
		"tmdb_tv":      tvRaw,
		"tmdb_season":  seasonRaw,
		"tmdb_episode": epRaw,
		"manual":       in.Metadata,
		"release_date": chooseStr(in.ReleaseDate, asString(epRaw["air_date"])),
	}
	return s.repo.UpdateVideoScrapeResult(ctx, video.ID, &episodeTMDBID, title, overview, thumbPath, meta, "uploaded")
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
	thumbPath, err := s.downloadTMDBImage(ctx, asString(raw["poster_path"]), videoID)
	if err != nil {
		return err
	}
	meta := map[string]any{
		"tmdb": raw,
	}
	return s.repo.UpdateVideoScrapeResult(ctx, videoID, &tmdbID, title, overview, thumbPath, meta, "uploaded")
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

	episodeTMDBID := 0
	title := ""
	description := ""
	thumbPath := ""
	episodes, _ := seasonRaw["episodes"].([]any)
	for _, ep := range episodes {
		epRaw, ok := ep.(map[string]any)
		if !ok || asInt(epRaw["episode_number"]) != episodeNum {
			continue
		}
		episodeTMDBID = asInt(epRaw["id"])
		title = asString(epRaw["name"])
		description = asString(epRaw["overview"])
		thumbPath, err = s.downloadTMDBImage(ctx, asString(epRaw["still_path"]), videoID)
		if err != nil {
			return err
		}
		if err := s.repo.UpsertEpisode(ctx, seasonID, episodeNum, title, description, asString(epRaw["still_path"]), asInt(epRaw["runtime"]), parseDate(epRaw["air_date"]), videoID); err != nil {
			return err
		}
		break
	}
	if episodeTMDBID > 0 {
		meta := map[string]any{"tmdb_tv": tvRaw, "tmdb_season": seasonRaw}
		if err := s.repo.UpdateVideoScrapeResult(ctx, videoID, &episodeTMDBID, title, description, thumbPath, meta, "uploaded"); err != nil {
			return err
		}
	}
	return nil
}

func (s *ScraperService) ScrapeMovieUpload(ctx context.Context, videoID uuid.UUID, filePath, filename string) (ScrapeResult, error) {
	if s.apiKey == "" {
		return ScrapeResult{}, fmt.Errorf("TMDB_API_KEY is empty")
	}
	if filename == "" {
		filename = filepath.Base(filePath)
	}
	title, year, _, _, ok := utils.ParseFilename(filename)
	if !ok || strings.TrimSpace(title) == "" {
		title = strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	}

	query := url.Values{}
	query.Set("api_key", s.apiKey)
	query.Set("query", title)
	if year > 0 {
		query.Set("year", fmt.Sprintf("%d", year))
	}
	searchAPI := fmt.Sprintf("%s/search/movie?%s", s.baseURL, query.Encode())
	searchRaw, err := s.getJSON(ctx, searchAPI)
	if err != nil {
		return ScrapeResult{}, err
	}
	results, _ := searchRaw["results"].([]any)
	if len(results) == 0 {
		return ScrapeResult{}, fmt.Errorf("no movie result for %q", title)
	}
	first, _ := results[0].(map[string]any)
	movieID := asInt(first["id"])
	if movieID <= 0 {
		return ScrapeResult{}, fmt.Errorf("invalid movie id from tmdb search")
	}
	if existingID, exists, err := s.repo.FindVideoByTypeTMDB(ctx, "movie", movieID, videoID); err != nil {
		return ScrapeResult{}, err
	} else if exists {
		return ScrapeResult{}, fmt.Errorf("duplicate movie metadata, existing video_id=%s", existingID)
	}

	detailAPI := fmt.Sprintf("%s/movie/%d?%s", s.baseURL, movieID, url.Values{"api_key": []string{s.apiKey}}.Encode())
	detailRaw, err := s.getJSON(ctx, detailAPI)
	if err != nil {
		return ScrapeResult{}, err
	}

	thumbPath, err := s.downloadTMDBImage(ctx, asString(detailRaw["poster_path"]), videoID)
	if err != nil {
		return ScrapeResult{}, err
	}
	scrape := ScrapeResult{
		TMDBID:        movieID,
		Title:         asString(detailRaw["title"]),
		Description:   asString(detailRaw["overview"]),
		ThumbnailPath: thumbPath,
		Metadata: map[string]any{
			"tmdb":        detailRaw,
			"tmdb_search": first,
		},
	}
	if scrape.Title == "" {
		scrape.Title = title
	}

	if err := s.repo.UpdateVideoScrapeResult(ctx, videoID, &scrape.TMDBID, scrape.Title, scrape.Description, scrape.ThumbnailPath, scrape.Metadata, "uploaded"); err != nil {
		return ScrapeResult{}, err
	}
	return scrape, nil
}

func (s *ScraperService) ScrapeEpisodeUpload(ctx context.Context, videoID uuid.UUID, filePath, filename string) (ScrapeResult, error) {
	if s.apiKey == "" {
		return ScrapeResult{}, fmt.Errorf("TMDB_API_KEY is empty")
	}
	if filename == "" {
		filename = filepath.Base(filePath)
	}
	title, _, seasonNum, episodeNum, ok := utils.ParseFilename(filename)
	if !ok || seasonNum <= 0 || episodeNum <= 0 || strings.TrimSpace(title) == "" {
		return ScrapeResult{}, fmt.Errorf("cannot parse episode info from filename: %s", filename)
	}

	q := url.Values{}
	q.Set("api_key", s.apiKey)
	q.Set("query", title)
	tvSearchAPI := fmt.Sprintf("%s/search/tv?%s", s.baseURL, q.Encode())
	tvSearchRaw, err := s.getJSON(ctx, tvSearchAPI)
	if err != nil {
		return ScrapeResult{}, err
	}
	tvResults, _ := tvSearchRaw["results"].([]any)
	if len(tvResults) == 0 {
		return ScrapeResult{}, fmt.Errorf("no tv result for %q", title)
	}
	tvFirst, _ := tvResults[0].(map[string]any)
	tvID := asInt(tvFirst["id"])
	if tvID <= 0 {
		return ScrapeResult{}, fmt.Errorf("invalid tv id from tmdb search")
	}

	tvDetailAPI := fmt.Sprintf("%s/tv/%d?%s", s.baseURL, tvID, url.Values{"api_key": []string{s.apiKey}}.Encode())
	tvRaw, err := s.getJSON(ctx, tvDetailAPI)
	if err != nil {
		return ScrapeResult{}, err
	}
	seriesID, err := s.repo.UpsertSeries(
		ctx,
		tvID,
		asString(tvRaw["name"]),
		asString(tvRaw["overview"]),
		asString(tvRaw["poster_path"]),
		asString(tvRaw["backdrop_path"]),
		parseDate(tvRaw["first_air_date"]),
		asInt(tvRaw["number_of_seasons"]),
		asInt(tvRaw["number_of_episodes"]),
	)
	if err != nil {
		return ScrapeResult{}, err
	}

	seasonAPI := fmt.Sprintf("%s/tv/%d/season/%d?%s", s.baseURL, tvID, seasonNum, url.Values{"api_key": []string{s.apiKey}}.Encode())
	seasonRaw, err := s.getJSON(ctx, seasonAPI)
	if err != nil {
		return ScrapeResult{}, err
	}
	seasonID, err := s.repo.UpsertSeason(ctx, seriesID, seasonNum, asString(seasonRaw["name"]), asString(seasonRaw["overview"]), asString(seasonRaw["poster_path"]), parseDate(seasonRaw["air_date"]))
	if err != nil {
		return ScrapeResult{}, err
	}

	var epRaw map[string]any
	episodes, _ := seasonRaw["episodes"].([]any)
	for _, ep := range episodes {
		item, ok := ep.(map[string]any)
		if !ok {
			continue
		}
		if asInt(item["episode_number"]) == episodeNum {
			epRaw = item
			break
		}
	}
	if epRaw == nil {
		return ScrapeResult{}, fmt.Errorf("episode S%02dE%02d not found in tmdb", seasonNum, episodeNum)
	}

	episodeTMDBID := asInt(epRaw["id"])
	if episodeTMDBID <= 0 {
		return ScrapeResult{}, fmt.Errorf("invalid episode tmdb id")
	}
	if existingID, exists, err := s.repo.FindVideoByTypeTMDB(ctx, "episode", episodeTMDBID, videoID); err != nil {
		return ScrapeResult{}, err
	} else if exists {
		return ScrapeResult{}, fmt.Errorf("duplicate episode metadata, existing video_id=%s", existingID)
	}

	if err := s.repo.UpsertEpisode(
		ctx,
		seasonID,
		episodeNum,
		asString(epRaw["name"]),
		asString(epRaw["overview"]),
		asString(epRaw["still_path"]),
		asInt(epRaw["runtime"]),
		parseDate(epRaw["air_date"]),
		videoID,
	); err != nil {
		return ScrapeResult{}, err
	}

	posterRel := asString(epRaw["still_path"])
	if posterRel == "" {
		posterRel = asString(tvRaw["poster_path"])
	}
	thumbPath, err := s.downloadTMDBImage(ctx, posterRel, videoID)
	if err != nil {
		return ScrapeResult{}, err
	}

	episodeTitle := asString(epRaw["name"])
	if episodeTitle == "" {
		episodeTitle = fmt.Sprintf("%s S%02dE%02d", title, seasonNum, episodeNum)
	}
	scrape := ScrapeResult{
		TMDBID:        episodeTMDBID,
		Title:         episodeTitle,
		Description:   asString(epRaw["overview"]),
		ThumbnailPath: thumbPath,
		Metadata: map[string]any{
			"tmdb_tv":      tvRaw,
			"tmdb_season":  seasonRaw,
			"tmdb_episode": epRaw,
		},
	}
	if err := s.repo.UpdateVideoScrapeResult(ctx, videoID, &scrape.TMDBID, scrape.Title, scrape.Description, scrape.ThumbnailPath, scrape.Metadata, "uploaded"); err != nil {
		return ScrapeResult{}, err
	}
	return scrape, nil
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

func (s *ScraperService) DownloadPoster(ctx context.Context, posterURL string, videoID uuid.UUID) (string, error) {
	posterURL = strings.TrimSpace(posterURL)
	if posterURL == "" {
		return "", nil
	}
	imageURL := posterURL
	if !strings.HasPrefix(posterURL, "http://") && !strings.HasPrefix(posterURL, "https://") {
		imageURL = "https://image.tmdb.org/t/p/original" + posterURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("create poster request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download poster failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("poster status=%d body=%s", resp.StatusCode, string(body))
	}

	root := s.posterRoot
	if strings.TrimSpace(root) == "" {
		root = filepath.Join(s.storageRoot, "posters")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", fmt.Errorf("create poster root: %w", err)
	}
	outputPath := filepath.Join(root, videoID.String()+".jpg")
	f, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("create poster file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("write poster file: %w", err)
	}
	return outputPath, nil
}

func (s *ScraperService) downloadTMDBImage(ctx context.Context, relativePath string, videoID uuid.UUID) (string, error) {
	relativePath = strings.TrimSpace(relativePath)
	if relativePath == "" {
		return "", nil
	}
	imageURL := relativePath
	if !strings.HasPrefix(relativePath, "http://") && !strings.HasPrefix(relativePath, "https://") {
		imageURL = "https://image.tmdb.org/t/p/original" + relativePath
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("create poster request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download poster failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("poster status=%d body=%s", resp.StatusCode, string(body))
	}

	outputDir := filepath.Join(s.storageRoot, "videos", videoID.String())
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create poster dir: %w", err)
	}
	outputPath := filepath.Join(outputDir, "poster.jpg")
	f, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("create poster file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("write poster file: %w", err)
	}
	return outputPath, nil
}

func extractGenres(v any) []string {
	rows, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		item, ok := row.(map[string]any)
		if !ok {
			continue
		}
		name := asString(item["name"])
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

func chooseStr(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return strings.TrimSpace(primary)
	}
	return strings.TrimSpace(fallback)
}

func normalizeCacheKey(raw string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(raw)), " "))
}

func (s *ScraperService) getPreviewCache(key string) ([]map[string]any, bool) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	entry, ok := s.previewCache[key]
	if !ok || time.Now().After(entry.ExpireAt) {
		return nil, false
	}
	return entry.Candidates, true
}

func (s *ScraperService) setPreviewCache(key string, candidates []map[string]any) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	now := time.Now()
	for k, v := range s.previewCache {
		if now.After(v.ExpireAt) {
			delete(s.previewCache, k)
		}
	}
	s.previewCache[key] = previewCacheEntry{
		ExpireAt:   now.Add(s.cacheTTL),
		Candidates: candidates,
	}
}
