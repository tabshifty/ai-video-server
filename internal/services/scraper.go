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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/utils"
)

type scraperRepo interface {
	GetVideoByID(ctx context.Context, videoID uuid.UUID) (models.Video, error)
	UpdateVideoScrapeResult(ctx context.Context, videoID uuid.UUID, tmdbID *int, title, description, thumbnailPath string, metadata map[string]any, status string) error
	UpsertSeries(ctx context.Context, tmdbID int, title, overview, poster, backdrop string, firstAirDate *time.Time, seasons, episodes int) (int64, error)
	UpsertSeason(ctx context.Context, seriesID int64, seasonNumber int, title, overview, poster string, airDate *time.Time) (int64, error)
	UpsertEpisode(ctx context.Context, seasonID int64, episodeNumber int, title, overview, stillPath string, runtime int, airDate *time.Time, videoID uuid.UUID) error
	FindVideoByTypeTMDB(ctx context.Context, typ string, tmdbID int, excludeVideoID uuid.UUID) (uuid.UUID, bool, error)
	ResolveActorIDs(ctx context.Context, actorIDs []uuid.UUID, actorNames []string, source string) ([]uuid.UUID, error)
	AddVideoActors(ctx context.Context, videoID uuid.UUID, actorIDs []uuid.UUID, source string) error
}

// ScraperService handles TMDB search and metadata syncing.
type ScraperService struct {
	repo         scraperRepo
	apiKey       string
	baseURL      string
	avBaseURL    string
	avUserAgent  string
	avProvider   avCrawlerProvider
	storageRoot  string
	posterRoot   string
	httpClient   *http.Client
	avHTTPClient *http.Client
	cacheTTL     time.Duration
	cacheMu      sync.RWMutex
	previewCache map[string]previewCacheEntry
}

type previewCacheEntry struct {
	ExpireAt   time.Time
	Candidates []map[string]any
}

func NewScraperService(repo scraperRepo, apiKey, baseURL, storageRoot, posterRoot string, timeout time.Duration) *ScraperService {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	svc := &ScraperService{
		repo:        repo,
		apiKey:      apiKey,
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		avBaseURL:   "https://javdb.com",
		avUserAgent: "Mozilla/5.0 (compatible; VideoServerBot/1.0; +https://example.invalid/bot)",
		storageRoot: storageRoot,
		posterRoot:  posterRoot,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		avHTTPClient: &http.Client{
			Timeout: timeout,
		},
		cacheTTL:     5 * time.Minute,
		previewCache: map[string]previewCacheEntry{},
	}
	svc.avProvider = newAVCrawlerProvider(svc)
	return svc
}

func (s *ScraperService) ConfigureAVScraper(baseURL, userAgent string, timeout time.Duration) {
	if strings.TrimSpace(baseURL) != "" {
		s.avBaseURL = strings.TrimSuffix(strings.TrimSpace(baseURL), "/")
	}
	if strings.TrimSpace(userAgent) != "" {
		s.avUserAgent = strings.TrimSpace(userAgent)
	}
	if timeout > 0 {
		s.avHTTPClient = &http.Client{Timeout: timeout}
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
	Type          string // movie | tv | episode | av
	TMDBID        int
	ExternalID    string
	Title         string
	Overview      string
	PosterURL     string
	ReleaseDate   string
	Metadata      map[string]any
	SeasonNumber  int
	EpisodeNumber int
}

const tmdbLangChinese = "zh-CN"
const tmdbCastLimit = 20
const avPreviewLimitDefault = 10

var (
	avCodePattern              = regexp.MustCompile(`(?i)\b([a-z]{2,10})[-_ ]?(\d{2,5})\b`)
	avFC2CodePattern           = regexp.MustCompile(`(?i)\bFC2(?:-?PPV)?[-_ ]?(\d{5,7})\b`)
	avHeyzoCodePattern         = regexp.MustCompile(`(?i)\bHEYZO[-_ ]?(\d{3,6})\b`)
	avSurenCodePattern         = regexp.MustCompile(`(?i)\b(\d{2,}[a-z]{2,})[-_ ]?(\d{2,6}[a-z]?)\b`)
	avGeneralCodePattern       = regexp.MustCompile(`(?i)\b([a-z]{2,12})[-_ ]?0*(\d{2,6}[a-z]?)\b`)
	javDBVideoAnchorRe         = regexp.MustCompile(`(?is)<a[^>]*href=["']([^"']*/v/[^"']+)["'][^>]*>(.*?)</a>`)
	javDBTitleRe               = regexp.MustCompile(`(?is)<h2[^>]*class=["'][^"']*title[^"']*["'][^>]*>(.*?)</h2>`)
	javDBPageTitleRe           = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	javDBOGImageRe             = regexp.MustCompile(`(?is)<meta[^>]+property=["']og:image["'][^>]*content=["']([^"']+)["']`)
	javDBDescriptionMetaRe     = regexp.MustCompile(`(?is)<meta[^>]+name=["']description["'][^>]*content=["']([^"']+)["']`)
	javDBReleaseDateFieldRe    = regexp.MustCompile(`(?is)(?:日期|発売日|發行日期|Release Date)[^<]{0,20}</strong>\s*<span[^>]*>(.*?)</span>`)
	javDBGenericDateInlineRe   = regexp.MustCompile(`\b(19|20)\d{2}[-/.](0?[1-9]|1[0-2])[-/.](0?[1-9]|[12]\d|3[01])\b`)
	javDBVideoCodeFieldRe      = regexp.MustCompile(`(?is)(?:番號|番号|識別碼|Code)[^<]{0,20}</strong>\s*<span[^>]*>(.*?)</span>`)
	javDBExternalIDPathPattern = regexp.MustCompile(`(?i)/v/([^/?#]+)`)
)

type avScrapeCandidate struct {
	ExternalID  string
	Code        string
	Title       string
	Overview    string
	PosterURL   string
	ReleaseDate string
	Actors      []string
	DetailURL   string
	Raw         map[string]any
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
	q.Set("query", strings.TrimSpace(title))
	if year > 0 {
		q.Set("year", fmt.Sprintf("%d", year))
	}
	raw, err := s.getTMDBJSON(ctx, "/search/movie", q, tmdbLangChinese)
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
		detail, dErr := s.getTMDBJSON(ctx, fmt.Sprintf("/movie/%d", id), nil, tmdbLangChinese)
		if dErr != nil {
			detail = item
		} else if needsLocalizedFallback(detail, "movie") {
			fallback, fErr := s.getTMDBJSON(ctx, fmt.Sprintf("/movie/%d", id), nil, "")
			if fErr == nil {
				detail = mergeLocalizedDetail(detail, fallback, "movie")
			}
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
	q.Set("query", strings.TrimSpace(title))
	raw, err := s.getTMDBJSON(ctx, "/search/tv", q, tmdbLangChinese)
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
		detail, dErr := s.getTMDBJSON(ctx, fmt.Sprintf("/tv/%d", id), nil, tmdbLangChinese)
		if dErr != nil {
			detail = item
		} else if needsLocalizedFallback(detail, "tv") {
			fallback, fErr := s.getTMDBJSON(ctx, fmt.Sprintf("/tv/%d", id), nil, "")
			if fErr == nil {
				detail = mergeLocalizedDetail(detail, fallback, "tv")
			}
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

func (s *ScraperService) PreviewAV(ctx context.Context, title string) ([]map[string]any, error) {
	keyword := normalizeAVKeyword(title)
	if keyword == "" {
		return nil, fmt.Errorf("title is required")
	}
	cacheKey := fmt.Sprintf("av|%s", normalizeCacheKey(keyword))
	if c, ok := s.getPreviewCache(cacheKey); ok {
		return c, nil
	}

	candidates, err := s.searchAVCandidates(ctx, keyword, avPreviewLimitDefault)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(candidates))
	for _, candidate := range candidates {
		out = append(out, map[string]any{
			"external_id":     candidate.ExternalID,
			"av_code":         candidate.Code,
			"title":           candidate.Title,
			"original_title":  candidate.Title,
			"overview":        candidate.Overview,
			"poster_url":      candidate.PosterURL,
			"release_date":    candidate.ReleaseDate,
			"actors":          candidate.Actors,
			"metadata":        candidate.Raw,
			"media_type_hint": "av",
			"scrape_source":   "javdb",
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
	detail, err := s.getTMDBJSON(ctx, fmt.Sprintf("/movie/%d", in.TMDBID), nil, tmdbLangChinese)
	if err != nil {
		return err
	}
	if needsLocalizedFallback(detail, "movie") {
		fallback, fErr := s.getTMDBJSON(ctx, fmt.Sprintf("/movie/%d", in.TMDBID), nil, "")
		if fErr == nil {
			detail = mergeLocalizedDetail(detail, fallback, "movie")
		}
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
	if err := s.repo.UpdateVideoScrapeResult(ctx, video.ID, &in.TMDBID, title, overview, thumbPath, meta, "uploaded"); err != nil {
		return err
	}
	s.syncMovieActors(ctx, video.ID, in.TMDBID)
	return nil
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

	tvRaw, err := s.getTMDBJSON(ctx, fmt.Sprintf("/tv/%d", in.TMDBID), nil, tmdbLangChinese)
	if err != nil {
		return err
	}
	if needsLocalizedFallback(tvRaw, "tv") {
		tvFallback, fErr := s.getTMDBJSON(ctx, fmt.Sprintf("/tv/%d", in.TMDBID), nil, "")
		if fErr == nil {
			tvRaw = mergeLocalizedDetail(tvRaw, tvFallback, "tv")
		}
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

	seasonRaw, err := s.getTMDBJSON(ctx, fmt.Sprintf("/tv/%d/season/%d", in.TMDBID, seasonNum), nil, tmdbLangChinese)
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
	if err := s.repo.UpdateVideoScrapeResult(ctx, video.ID, &episodeTMDBID, title, overview, thumbPath, meta, "uploaded"); err != nil {
		return err
	}
	s.syncEpisodeActors(ctx, video.ID, in.TMDBID, seasonNum, episodeNum)
	return nil
}

func (s *ScraperService) ConfirmAV(ctx context.Context, in ConfirmScrapeInput) error {
	video, err := s.repo.GetVideoByID(ctx, in.VideoID)
	if err != nil {
		return err
	}

	externalID := strings.TrimSpace(in.ExternalID)
	if externalID == "" {
		return fmt.Errorf("external_id is required")
	}

	detailURL := toAbsoluteURL(strings.TrimSpace(s.avBaseURL), "/v/"+externalID)
	candidate, trace, err := s.fetchAVCandidateByDetailURLWithTrace(ctx, detailURL)
	if err != nil {
		return err
	}

	title := chooseStr(in.Title, candidate.Title)
	if strings.TrimSpace(title) == "" {
		title = strings.TrimSpace(video.Title)
	}
	overview := chooseStr(in.Overview, candidate.Overview)
	posterURL := chooseStr(in.PosterURL, candidate.PosterURL)
	thumbPath, err := s.DownloadPoster(ctx, posterURL, in.VideoID)
	if err != nil {
		return err
	}

	meta := map[string]any{
		"scrape_source": "javdb",
		"external_id":   candidate.ExternalID,
		"av_code":       candidate.Code,
		"actors":        candidate.Actors,
		"release_date":  chooseStr(in.ReleaseDate, candidate.ReleaseDate),
		"javdb":         candidate.Raw,
		"manual":        in.Metadata,
		"scrape_trace":  trace,
	}
	if err := s.repo.UpdateVideoScrapeResult(ctx, video.ID, nil, title, overview, thumbPath, meta, "uploaded"); err != nil {
		return err
	}
	s.syncAVActors(ctx, video.ID, candidate.Actors)
	return nil
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
	if err := s.repo.UpdateVideoScrapeResult(ctx, videoID, &tmdbID, title, overview, thumbPath, meta, "uploaded"); err != nil {
		return err
	}
	s.syncMovieActors(ctx, videoID, tmdbID)
	return nil
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
		s.syncEpisodeActors(ctx, videoID, tmdbID, seasonNum, episodeNum)
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
	query.Set("query", title)
	if year > 0 {
		query.Set("year", fmt.Sprintf("%d", year))
	}
	searchRaw, err := s.getTMDBJSON(ctx, "/search/movie", query, tmdbLangChinese)
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

	detailRaw, err := s.fetchLocalizedMediaDetail(ctx, fmt.Sprintf("/movie/%d", movieID), "movie")
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
	s.syncMovieActors(ctx, videoID, scrape.TMDBID)
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
	q.Set("query", title)
	tvSearchRaw, err := s.getTMDBJSON(ctx, "/search/tv", q, tmdbLangChinese)
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

	tvRaw, err := s.fetchLocalizedMediaDetail(ctx, fmt.Sprintf("/tv/%d", tvID), "tv")
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

	seasonRaw, err := s.fetchLocalizedSeasonDetail(ctx, tvID, seasonNum)
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
	s.syncEpisodeActors(ctx, videoID, tvID, seasonNum, episodeNum)
	return scrape, nil
}

func (s *ScraperService) ScrapeAVUpload(ctx context.Context, videoID uuid.UUID, filePath, filename string) (ScrapeResult, error) {
	if filename == "" {
		filename = filepath.Base(filePath)
	}

	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	keyword := normalizeAVKeyword(base)
	if keyword == "" {
		keyword = normalizeAVKeyword(filePath)
	}
	if keyword == "" {
		return ScrapeResult{}, fmt.Errorf("invalid av filename")
	}

	candidates, trace, err := s.searchAVCandidatesWithTrace(ctx, keyword, 6)
	if err != nil {
		return ScrapeResult{}, err
	}
	if len(candidates) == 0 {
		return ScrapeResult{}, fmt.Errorf("no av result for %q", keyword)
	}
	candidate := candidates[0]

	thumbPath, err := s.DownloadPoster(ctx, candidate.PosterURL, videoID)
	if err != nil {
		return ScrapeResult{}, err
	}

	title := strings.TrimSpace(candidate.Title)
	if title == "" {
		title = keyword
	}
	meta := map[string]any{
		"scrape_source":  "javdb",
		"external_id":    candidate.ExternalID,
		"av_code":        candidate.Code,
		"actors":         candidate.Actors,
		"release_date":   candidate.ReleaseDate,
		"search_keyword": keyword,
		"javdb":          candidate.Raw,
		"scrape_trace":   trace,
	}
	scrape := ScrapeResult{
		TMDBID:        0,
		Title:         title,
		Description:   strings.TrimSpace(candidate.Overview),
		ThumbnailPath: thumbPath,
		Metadata:      meta,
	}
	if err := s.repo.UpdateVideoScrapeResult(ctx, videoID, nil, scrape.Title, scrape.Description, scrape.ThumbnailPath, scrape.Metadata, "uploaded"); err != nil {
		return ScrapeResult{}, err
	}
	s.syncAVActors(ctx, videoID, candidate.Actors)
	return scrape, nil
}

func (s *ScraperService) syncMovieActors(ctx context.Context, videoID uuid.UUID, movieTMDBID int) {
	if movieTMDBID <= 0 {
		return
	}
	path := fmt.Sprintf("/movie/%d/credits", movieTMDBID)
	credits, err := s.getTMDBJSON(ctx, path, nil, tmdbLangChinese)
	if err != nil {
		return
	}
	names := extractCastNames(credits, tmdbCastLimit)
	if len(names) == 0 {
		fallback, fallbackErr := s.getTMDBJSON(ctx, path, nil, "")
		if fallbackErr == nil {
			names = extractCastNames(fallback, tmdbCastLimit)
		}
	}
	if len(names) == 0 {
		return
	}
	actorIDs, err := s.repo.ResolveActorIDs(ctx, nil, names, "scrape_tmdb")
	if err != nil {
		return
	}
	_ = s.repo.AddVideoActors(ctx, videoID, actorIDs, "scrape_tmdb")
}

func (s *ScraperService) syncEpisodeActors(ctx context.Context, videoID uuid.UUID, tvTMDBID, seasonNum, episodeNum int) {
	if tvTMDBID <= 0 || seasonNum <= 0 || episodeNum <= 0 {
		return
	}
	episodeCreditsPath := fmt.Sprintf("/tv/%d/season/%d/episode/%d/credits", tvTMDBID, seasonNum, episodeNum)
	names := s.tryFetchCastNames(ctx, episodeCreditsPath, tmdbCastLimit)
	if len(names) == 0 {
		tvCreditsPath := fmt.Sprintf("/tv/%d/credits", tvTMDBID)
		names = s.tryFetchCastNames(ctx, tvCreditsPath, tmdbCastLimit)
	}
	if len(names) == 0 {
		return
	}
	actorIDs, err := s.repo.ResolveActorIDs(ctx, nil, names, "scrape_tmdb")
	if err != nil {
		return
	}
	_ = s.repo.AddVideoActors(ctx, videoID, actorIDs, "scrape_tmdb")
}

func (s *ScraperService) syncAVActors(ctx context.Context, videoID uuid.UUID, actorNames []string) {
	names := dedupeAVActorNames(actorNames)
	if len(names) == 0 {
		return
	}
	actorIDs, err := s.repo.ResolveActorIDs(ctx, nil, names, "scrape_av")
	if err != nil {
		return
	}
	_ = s.repo.AddVideoActors(ctx, videoID, actorIDs, "scrape_av")
}

func (s *ScraperService) searchAVCandidates(ctx context.Context, keyword string, limit int) ([]avScrapeCandidate, error) {
	candidates, _, err := s.searchAVCandidatesWithTrace(ctx, keyword, limit)
	return candidates, err
}

func (s *ScraperService) searchAVByQuery(ctx context.Context, keyword string, limit int) ([]avScrapeCandidate, error) {
	crawler := s.defaultAVCrawler()
	run := newAVScrapeRunContext("search_by_query", "")
	run.addSearchQuery(keyword)
	out, err := crawler.SearchCandidates(ctx, run, keyword, limit)
	if err != nil && len(out) == 0 {
		return nil, err
	}
	return out, nil
}

func (s *ScraperService) fetchAVCandidateByDetailURL(ctx context.Context, detailURL string) (avScrapeCandidate, error) {
	candidate, _, err := s.fetchAVCandidateByDetailURLWithTrace(ctx, detailURL)
	return candidate, err
}

func (s *ScraperService) tryFetchCastNames(ctx context.Context, path string, limit int) []string {
	raw, err := s.getTMDBJSON(ctx, path, nil, tmdbLangChinese)
	if err == nil {
		if names := extractCastNames(raw, limit); len(names) > 0 {
			return names
		}
	}
	fallback, fallbackErr := s.getTMDBJSON(ctx, path, nil, "")
	if fallbackErr != nil {
		return nil
	}
	return extractCastNames(fallback, limit)
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

func (s *ScraperService) getTMDBJSON(ctx context.Context, path string, q url.Values, language string) (map[string]any, error) {
	query := cloneURLValues(q)
	query.Set("api_key", s.apiKey)
	if strings.TrimSpace(language) != "" {
		query.Set("language", language)
	}
	endpoint := fmt.Sprintf("%s/%s?%s", s.baseURL, strings.TrimPrefix(path, "/"), query.Encode())
	return s.getJSON(ctx, endpoint)
}

func (s *ScraperService) fetchLocalizedMediaDetail(ctx context.Context, path, mediaType string) (map[string]any, error) {
	detail, err := s.getTMDBJSON(ctx, path, nil, tmdbLangChinese)
	if err != nil {
		return nil, err
	}
	if needsLocalizedFallback(detail, mediaType) {
		fallback, fallbackErr := s.getTMDBJSON(ctx, path, nil, "")
		if fallbackErr == nil {
			detail = mergeLocalizedDetail(detail, fallback, mediaType)
		}
	}
	return detail, nil
}

func (s *ScraperService) fetchLocalizedSeasonDetail(ctx context.Context, tvID, seasonNum int) (map[string]any, error) {
	path := fmt.Sprintf("/tv/%d/season/%d", tvID, seasonNum)
	seasonRaw, err := s.getTMDBJSON(ctx, path, nil, tmdbLangChinese)
	if err != nil {
		return nil, err
	}
	if needsSeasonLocalizedFallback(seasonRaw) {
		fallback, fallbackErr := s.getTMDBJSON(ctx, path, nil, "")
		if fallbackErr == nil {
			seasonRaw = mergeSeasonLocalizedDetail(seasonRaw, fallback)
		}
	}
	return seasonRaw, nil
}

func cloneURLValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vals := range in {
		cloned := make([]string, len(vals))
		copy(cloned, vals)
		out[k] = cloned
	}
	return out
}

func needsLocalizedFallback(detail map[string]any, mediaType string) bool {
	switch mediaType {
	case "movie":
		if isBlankAnyString(detail["title"]) || isBlankAnyString(detail["overview"]) || isBlankAnyString(detail["release_date"]) {
			return true
		}
		return len(extractGenres(detail["genres"])) == 0
	case "tv":
		if isBlankAnyString(detail["name"]) || isBlankAnyString(detail["overview"]) || isBlankAnyString(detail["first_air_date"]) {
			return true
		}
		return len(extractGenres(detail["genres"])) == 0
	default:
		return false
	}
}

func needsSeasonLocalizedFallback(detail map[string]any) bool {
	if isBlankAnyString(detail["name"]) || isBlankAnyString(detail["overview"]) || isBlankAnyString(detail["air_date"]) {
		return true
	}
	episodes, _ := detail["episodes"].([]any)
	for _, episode := range episodes {
		row, ok := episode.(map[string]any)
		if !ok {
			continue
		}
		if isBlankAnyString(row["name"]) || isBlankAnyString(row["overview"]) || isBlankAnyString(row["air_date"]) {
			return true
		}
	}
	return false
}

func mergeLocalizedDetail(primary, fallback map[string]any, mediaType string) map[string]any {
	if len(primary) == 0 {
		return fallback
	}
	if len(fallback) == 0 {
		return primary
	}

	switch mediaType {
	case "movie":
		fillBlankStringField(primary, fallback, "title")
		fillBlankStringField(primary, fallback, "original_title")
		fillBlankStringField(primary, fallback, "overview")
		fillBlankStringField(primary, fallback, "release_date")
		fillBlankGenresField(primary, fallback)
	case "tv":
		fillBlankStringField(primary, fallback, "name")
		fillBlankStringField(primary, fallback, "original_name")
		fillBlankStringField(primary, fallback, "overview")
		fillBlankStringField(primary, fallback, "first_air_date")
		fillBlankStringField(primary, fallback, "status")
		fillBlankStringField(primary, fallback, "type")
		fillBlankGenresField(primary, fallback)
		fillBlankNestedObjectFields(primary, fallback, "last_episode_to_air", "name", "overview")
		fillBlankNestedSliceObjectFields(primary, fallback, "seasons", "name", "overview")
	}

	return primary
}

func mergeSeasonLocalizedDetail(primary, fallback map[string]any) map[string]any {
	if len(primary) == 0 {
		return fallback
	}
	if len(fallback) == 0 {
		return primary
	}
	fillBlankStringField(primary, fallback, "name")
	fillBlankStringField(primary, fallback, "overview")
	fillBlankStringField(primary, fallback, "air_date")
	fillBlankStringField(primary, fallback, "poster_path")
	fillBlankEpisodes(primary, fallback)
	return primary
}

func fillBlankStringField(dst, src map[string]any, key string) {
	if !isBlankAnyString(dst[key]) {
		return
	}
	fallback := strings.TrimSpace(asString(src[key]))
	if fallback != "" {
		dst[key] = fallback
	}
}

func fillBlankGenresField(dst, src map[string]any) {
	if len(extractGenres(dst["genres"])) > 0 {
		return
	}
	if rows, ok := src["genres"].([]any); ok && len(rows) > 0 {
		dst["genres"] = rows
	}
}

func fillBlankNestedObjectFields(dst, src map[string]any, key string, fields ...string) {
	srcObj, ok := src[key].(map[string]any)
	if !ok {
		return
	}
	dstObj, _ := dst[key].(map[string]any)
	if dstObj == nil {
		dstObj = map[string]any{}
		dst[key] = dstObj
	}
	for _, field := range fields {
		if !isBlankAnyString(dstObj[field]) {
			continue
		}
		if value := strings.TrimSpace(asString(srcObj[field])); value != "" {
			dstObj[field] = value
		}
	}
}

func fillBlankNestedSliceObjectFields(dst, src map[string]any, key string, fields ...string) {
	srcRows, ok := src[key].([]any)
	if !ok || len(srcRows) == 0 {
		return
	}
	dstRows, _ := dst[key].([]any)
	if len(dstRows) == 0 {
		dst[key] = srcRows
		return
	}

	limit := len(dstRows)
	if len(srcRows) < limit {
		limit = len(srcRows)
	}
	for i := 0; i < limit; i++ {
		dstObj, okDst := dstRows[i].(map[string]any)
		srcObj, okSrc := srcRows[i].(map[string]any)
		if !okDst || !okSrc {
			continue
		}
		for _, field := range fields {
			if !isBlankAnyString(dstObj[field]) {
				continue
			}
			if value := strings.TrimSpace(asString(srcObj[field])); value != "" {
				dstObj[field] = value
			}
		}
	}
}

func fillBlankEpisodes(dst, src map[string]any) {
	srcRows, ok := src["episodes"].([]any)
	if !ok || len(srcRows) == 0 {
		return
	}
	dstRows, _ := dst["episodes"].([]any)
	if len(dstRows) == 0 {
		dst["episodes"] = srcRows
		return
	}

	fallbackByEpisode := map[int]map[string]any{}
	for _, raw := range srcRows {
		row, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		episodeNum := asInt(row["episode_number"])
		if episodeNum <= 0 {
			continue
		}
		fallbackByEpisode[episodeNum] = row
	}

	for _, raw := range dstRows {
		row, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		episodeNum := asInt(row["episode_number"])
		if episodeNum <= 0 {
			continue
		}
		fallback := fallbackByEpisode[episodeNum]
		if fallback == nil {
			continue
		}
		fillBlankStringField(row, fallback, "name")
		fillBlankStringField(row, fallback, "overview")
		fillBlankStringField(row, fallback, "air_date")
		fillBlankStringField(row, fallback, "still_path")
		if asInt(row["id"]) <= 0 && asInt(fallback["id"]) > 0 {
			row["id"] = fallback["id"]
		}
		if asInt(row["runtime"]) <= 0 && asInt(fallback["runtime"]) > 0 {
			row["runtime"] = fallback["runtime"]
		}
	}
}

func isBlankAnyString(v any) bool {
	return strings.TrimSpace(asString(v)) == ""
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

func normalizeAVKeyword(input string) string {
	base := strings.TrimSpace(strings.TrimSuffix(filepath.Base(input), filepath.Ext(input)))
	if base == "" {
		return ""
	}
	replaced := strings.NewReplacer(".", " ", "_", " ", "-", " ").Replace(base)
	return strings.Join(strings.Fields(replaced), " ")
}

func extractAVCode(input string) string {
	raw := strings.ToUpper(strings.TrimSpace(input))
	if raw == "" {
		return ""
	}
	raw = strings.NewReplacer(".", "-", "_", "-", " ", "-").Replace(raw)
	for strings.Contains(raw, "--") {
		raw = strings.ReplaceAll(raw, "--", "-")
	}

	if mm := avFC2CodePattern.FindStringSubmatch(raw); len(mm) > 1 {
		number := strings.TrimLeft(strings.TrimSpace(mm[1]), "0")
		if number == "" {
			number = "0"
		}
		return "FC2-" + number
	}
	if mm := avHeyzoCodePattern.FindStringSubmatch(raw); len(mm) > 1 {
		return "HEYZO-" + strings.TrimSpace(mm[1])
	}
	if mm := avSurenCodePattern.FindStringSubmatch(raw); len(mm) > 2 {
		prefix := strings.TrimSpace(mm[1])
		number := strings.TrimLeft(strings.TrimSpace(mm[2]), "0")
		if number == "" {
			number = "0"
		}
		return prefix + "-" + number
	}
	if mm := avGeneralCodePattern.FindStringSubmatch(raw); len(mm) > 2 {
		prefix := strings.TrimSpace(mm[1])
		number := strings.TrimLeft(strings.TrimSpace(mm[2]), "0")
		if number == "" {
			number = "0"
		}
		return prefix + "-" + number
	}
	if mm := avCodePattern.FindStringSubmatch(raw); len(mm) > 2 {
		prefix := strings.TrimSpace(mm[1])
		number := strings.TrimLeft(strings.TrimSpace(mm[2]), "0")
		if number == "" {
			number = "0"
		}
		return strings.ToUpper(prefix) + "-" + number
	}
	return ""
}

func normalizeAVCodeForCompare(input string) string {
	if strings.TrimSpace(input) == "" {
		return ""
	}
	code := strings.ToUpper(strings.TrimSpace(input))
	code = strings.NewReplacer("-", "", "_", "", ".", "", " ", "").Replace(code)
	return code
}

func extractJavDBTitle(content string) string {
	if mm := javDBTitleRe.FindStringSubmatch(content); len(mm) > 1 {
		title := stripHTMLText(mm[1])
		if title != "" {
			return title
		}
	}
	if mm := javDBPageTitleRe.FindStringSubmatch(content); len(mm) > 1 {
		title := stripHTMLText(mm[1])
		title = strings.TrimSpace(strings.TrimSuffix(title, "- JavDB"))
		if title != "" {
			return title
		}
	}
	return ""
}

func parseJavDBDetailActors(content string) []string {
	matches := javDBActorAnchorRe.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	out := make([]string, 0, len(matches))
	seen := map[string]struct{}{}
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		name := stripHTMLText(match[2])
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}
	return out
}

func dedupeAVActorNames(names []string) []string {
	if len(names) == 0 {
		return nil
	}
	out := make([]string, 0, len(names))
	seen := map[string]struct{}{}
	for _, raw := range names {
		name := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}
	return out
}

func extractJavDBVideoID(rawURL string) string {
	path := strings.TrimSpace(rawURL)
	if path == "" {
		return ""
	}
	if mm := javDBExternalIDPathPattern.FindStringSubmatch(path); len(mm) > 1 {
		return strings.TrimSpace(mm[1])
	}
	parsed, err := url.Parse(path)
	if err != nil {
		return ""
	}
	if mm := javDBExternalIDPathPattern.FindStringSubmatch(parsed.Path); len(mm) > 1 {
		return strings.TrimSpace(mm[1])
	}
	return ""
}

func normalizeAVDate(raw string) string {
	v := strings.TrimSpace(strings.NewReplacer("/", "-", ".", "-").Replace(raw))
	if len(v) >= 10 {
		v = v[:10]
	}
	if _, err := time.Parse("2006-01-02", v); err == nil {
		return v
	}
	return ""
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

func extractCastNames(raw map[string]any, limit int) []string {
	if limit <= 0 {
		limit = tmdbCastLimit
	}
	castRows, ok := raw["cast"].([]any)
	if !ok || len(castRows) == 0 {
		return nil
	}
	out := make([]string, 0, limit)
	seen := map[string]struct{}{}
	for _, row := range castRows {
		if len(out) >= limit {
			break
		}
		item, ok := row.(map[string]any)
		if !ok {
			continue
		}
		name := strings.TrimSpace(asString(item["name"]))
		if name == "" {
			name = strings.TrimSpace(asString(item["original_name"]))
		}
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}
	return out
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
