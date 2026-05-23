package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/utils"
)

type scraperRepo interface {
	GetVideoByID(ctx context.Context, videoID uuid.UUID) (models.Video, error)
	UpdateVideoScrapeResult(ctx context.Context, videoID uuid.UUID, tmdbID *int, title, description, thumbnailPath string, metadata map[string]any, status string) error
	UpsertSeries(ctx context.Context, tmdbID int, title, overview, poster, backdrop string, firstAirDate *time.Time, seasons, episodes int) (int64, error)
	UpsertSeason(ctx context.Context, seriesID int64, seasonNumber int, title, overview, poster string, airDate *time.Time) (int64, error)
	UpsertEpisode(ctx context.Context, seasonID int64, episodeNumber int, title, overview, stillPath string, runtime int, airDate *time.Time, videoID uuid.UUID, bindVideo bool) error
	FindVideoByTypeTMDB(ctx context.Context, typ string, tmdbID int, excludeVideoID uuid.UUID) (uuid.UUID, bool, error)
	ResolveActorIDs(ctx context.Context, actorIDs []uuid.UUID, actorNames []string, source string) ([]uuid.UUID, error)
	AddVideoActors(ctx context.Context, videoID uuid.UUID, actorIDs []uuid.UUID, source string) error
	ListVideoActors(ctx context.Context, videoID uuid.UUID) ([]models.AdminVideoActor, error)
	UpdateActorAvatar(ctx context.Context, actorID uuid.UUID, avatarURL, source, externalID string) error
	UpsertScrapedActorProfile(ctx context.Context, input models.AdminActorInput) (models.AdminActor, error)
}

// ScraperService handles TMDB search and metadata syncing.
type ScraperService struct {
	repo                scraperRepo
	avConfigStore       avScraperConfigStore
	apiKey              string
	baseURL             string
	avBaseURL           string
	avUserAgent         string
	avSiteURLs          map[string]string
	avJavDBCookie       string
	avJavBusCookie      string
	avThePornDBAPIToken string
	avThePornDBNoHash   bool
	avProvider          avCrawlerProvider
	textTranslator      contentTranslator
	storageRoot         string
	posterRoot          string
	httpClient          *http.Client
	avHTTPClient        *http.Client
	cacheTTL            time.Duration
	cacheMu             sync.RWMutex
	previewCache        map[string]previewCacheEntry
}

type previewCacheEntry struct {
	ExpireAt   time.Time
	Candidates []map[string]any
}

type AVScraperConfig struct {
	BaseURL           string
	UserAgent         string
	Timeout           time.Duration
	SiteURLs          map[string]string
	JavDBCookie       string
	JavBusCookie      string
	ThePornDBAPIToken string
	ThePornDBNoHash   bool
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
		avSiteURLs:  map[string]string{},
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
	if store, ok := repo.(avScraperConfigStore); ok {
		svc.avConfigStore = store
	}
	svc.avProvider = newAVCrawlerProvider(svc)
	return svc
}

func (s *ScraperService) ConfigureAVScraper(baseURL, userAgent string, timeout time.Duration) {
	s.ConfigureAVScraperConfig(AVScraperConfig{
		BaseURL:   baseURL,
		UserAgent: userAgent,
		Timeout:   timeout,
	})
}

func (s *ScraperService) ConfigureAVScraperConfig(cfg AVScraperConfig) {
	if strings.TrimSpace(cfg.BaseURL) != "" {
		s.avBaseURL = strings.TrimSuffix(strings.TrimSpace(cfg.BaseURL), "/")
	}
	if strings.TrimSpace(cfg.UserAgent) != "" {
		s.avUserAgent = strings.TrimSpace(cfg.UserAgent)
	}
	if cfg.Timeout > 0 {
		s.avHTTPClient = &http.Client{Timeout: cfg.Timeout}
	}
	if len(cfg.SiteURLs) > 0 {
		if s.avSiteURLs == nil {
			s.avSiteURLs = map[string]string{}
		}
		for rawSite, rawURL := range cfg.SiteURLs {
			site := normalizeAVSourceName(rawSite)
			if site == "" {
				site = strings.ToLower(strings.TrimSpace(rawSite))
			}
			rawURL = strings.TrimSuffix(strings.TrimSpace(rawURL), "/")
			if site == "" || rawURL == "" {
				continue
			}
			s.avSiteURLs[site] = rawURL
		}
	}
	if strings.TrimSpace(s.avBaseURL) != "" {
		if s.avSiteURLs == nil {
			s.avSiteURLs = map[string]string{}
		}
		if strings.TrimSpace(s.avSiteURLs["javdb"]) == "" {
			s.avSiteURLs["javdb"] = strings.TrimSuffix(strings.TrimSpace(s.avBaseURL), "/")
		}
	}
	s.avJavDBCookie = strings.TrimSpace(cfg.JavDBCookie)
	s.avJavBusCookie = strings.TrimSpace(cfg.JavBusCookie)
	s.avThePornDBAPIToken = strings.TrimSpace(cfg.ThePornDBAPIToken)
	s.avThePornDBNoHash = cfg.ThePornDBNoHash
}

func (s *ScraperService) ConfigureContentTranslation(cfg TranslationConfig) {
	if strings.TrimSpace(cfg.APIURL) == "" {
		s.textTranslator = nil
		return
	}
	s.textTranslator = NewOpenAITextTranslator(cfg)
}

type localizedScrapeFields struct {
	titleOriginal       string
	descriptionOriginal string
	titleZH             string
	descriptionZH       string
}

func (s *ScraperService) localizeScrapeFields(ctx context.Context, title, description string) localizedScrapeFields {
	fields := localizedScrapeFields{
		titleOriginal:       strings.TrimSpace(title),
		descriptionOriginal: strings.TrimSpace(description),
	}
	fields.titleZH = fields.titleOriginal
	fields.descriptionZH = fields.descriptionOriginal
	if s.textTranslator == nil || (fields.titleOriginal == "" && fields.descriptionOriginal == "") {
		return fields
	}
	translated, err := s.textTranslator.TranslateScrapeContent(ctx, fields.titleOriginal, fields.descriptionOriginal)
	if err != nil {
		return fields
	}
	if strings.TrimSpace(translated.Title) != "" {
		fields.titleZH = strings.TrimSpace(translated.Title)
	}
	if strings.TrimSpace(translated.Description) != "" {
		fields.descriptionZH = strings.TrimSpace(translated.Description)
	}
	return fields
}

func withLocalizedScrapeMetadata(metadata map[string]any, fields localizedScrapeFields) map[string]any {
	out := make(map[string]any, len(metadata)+4)
	for k, v := range metadata {
		out[k] = v
	}
	out["title_original"] = fields.titleOriginal
	out["description_original"] = fields.descriptionOriginal
	out["title_zh"] = fields.titleZH
	out["description_zh"] = fields.descriptionZH
	return out
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
	Status        string
	TMDBID        int
	ExternalID    string
	Title         string
	Overview      string
	PosterURL     string
	BackdropURL   string
	ThumbURL      string
	ReleaseDate   string
	Metadata      map[string]any
	SeasonNumber  int
	EpisodeNumber int
}

const tmdbLangChinese = "zh-CN"
const tmdbCastLimit = 20
const avPreviewLimitDefault = 10

const (
	EpisodeAutoScrapeStageParseFailed        = "parse_failed"
	EpisodeAutoScrapeStageCandidateAmbiguous = "candidate_ambiguous"
	EpisodeAutoScrapeStageCandidateNotFound  = "candidate_not_found"
	EpisodeAutoScrapeStageTMDBMissingEpisode = "tmdb_missing_episode"
	EpisodeAutoScrapeStageAPIError           = "api_error"
)

type EpisodeAutoScrapeError struct {
	Stage               string
	ParsedTitle         string
	ParsedSeasonNumber  int
	ParsedEpisodeNumber int
	CandidateCount      int
	CandidatePreview    []map[string]any
	Err                 error
}

func (e *EpisodeAutoScrapeError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	if strings.TrimSpace(e.Stage) != "" {
		return e.Stage
	}
	return "episode auto scrape failed"
}

func (e *EpisodeAutoScrapeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type syncedEpisodeResult struct {
	tvRaw         map[string]any
	seasonRaw     map[string]any
	episodeRaw    map[string]any
	episodeTMDBID int
}

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
	Source      string
	ExternalID  string
	Code        string
	Title       string
	Overview    string
	PosterURL   string
	ThumbURL    string
	ReleaseDate string
	Actors      []string
	DetailURL   string
	Raw         map[string]any
}

func (s *ScraperService) PreviewMovie(ctx context.Context, title string, year int) ([]map[string]any, error) {
	return s.PreviewMovieWithOptions(ctx, title, year, MoviePreviewOptions{})
}

func (s *ScraperService) PreviewMovieWithOptions(ctx context.Context, title string, year int, opts MoviePreviewOptions) ([]map[string]any, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("TMDB_API_KEY is empty")
	}
	cacheKey := fmt.Sprintf("movie|%s|%d", normalizeCacheKey(title), year)
	if !opts.BypassCache {
		if c, ok := s.getPreviewCache(cacheKey); ok {
			return c, nil
		}
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
			"backdrop_path":   asString(detail["backdrop_path"]),
			"release_date":    asString(detail["release_date"]),
			"vote_average":    detail["vote_average"],
			"genres":          extractGenres(detail["genres"]),
			"metadata":        detail,
			"media_type_hint": "movie",
		})
	}
	if !opts.BypassCache {
		s.setPreviewCache(cacheKey, out)
	}
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
			"title":           chooseStr(asString(detail["name"]), asString(detail["original_name"])),
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
	result, err := s.PreviewAVSearch(ctx, title, AVPreviewOptions{})
	if err != nil {
		return nil, err
	}
	return result.Candidates, nil
}

func (s *ScraperService) PreviewAVWithOptions(ctx context.Context, title string, opts AVPreviewOptions) ([]map[string]any, error) {
	result, err := s.PreviewAVSearch(ctx, title, opts)
	if err != nil {
		return nil, err
	}
	return result.Candidates, nil
}

func (s *ScraperService) PreviewAVSearch(ctx context.Context, title string, opts AVPreviewOptions) (AVPreviewResult, error) {
	keyword := resolveAVPreviewKeyword(title, opts)
	if keyword == "" {
		return AVPreviewResult{}, fmt.Errorf("title is required")
	}
	plan, err := s.resolveAVSearchPlan(ctx, keyword, opts)
	if err != nil {
		return AVPreviewResult{}, err
	}
	cacheKey := s.buildAVPreviewCacheKey(keyword, plan)
	if !opts.BypassCache {
		if c, ok := s.getPreviewCache(cacheKey); ok {
			return AVPreviewResult{
				Candidates:        c,
				SiteCategory:      plan.SiteCategory,
				RecommendedSource: plan.RecommendedSource,
				UsedSource:        previewAVUsedSourceFromCandidates(c, plan.RecommendedSource),
				EnabledSources:    append([]string(nil), plan.Config.EnabledSites...),
			}, nil
		}
	}
	candidates, trace, err := s.searchAVCandidatesWithTrace(ctx, keyword, avPreviewLimitDefault, plan)
	if err != nil {
		return AVPreviewResult{}, err
	}
	out := make([]map[string]any, 0, len(candidates))
	usedSource := plan.RecommendedSource
	for _, candidate := range candidates {
		scrapeSource := normalizeAVSourceName(candidate.Source)
		if scrapeSource == "" {
			scrapeSource = plan.RecommendedSource
		}
		if usedSource == "" {
			usedSource = scrapeSource
		}
		metadata := candidate.Raw
		if metadata == nil {
			metadata = map[string]any{}
		}
		metadata["site_category"] = plan.SiteCategory
		metadata["scrape_source"] = scrapeSource
		if trace != nil {
			metadata["scrape_trace"] = trace
		}
		out = append(out, map[string]any{
			"external_id":     candidate.ExternalID,
			"av_code":         candidate.Code,
			"title":           candidate.Title,
			"original_title":  candidate.Title,
			"overview":        candidate.Overview,
			"poster_url":      candidate.PosterURL,
			"thumb_url":       candidate.ThumbURL,
			"release_date":    candidate.ReleaseDate,
			"actors":          candidate.Actors,
			"detail_url":      candidate.DetailURL,
			"metadata":        metadata,
			"media_type_hint": "av",
			"scrape_source":   scrapeSource,
			"site_category":   plan.SiteCategory,
		})
	}
	if !opts.BypassCache {
		s.setPreviewCache(cacheKey, out)
	}
	return AVPreviewResult{
		Candidates:        out,
		SiteCategory:      plan.SiteCategory,
		RecommendedSource: plan.RecommendedSource,
		UsedSource:        usedSource,
		EnabledSources:    append([]string(nil), plan.Config.EnabledSites...),
	}, nil
}

func (s *ScraperService) buildAVPreviewCacheKey(keyword string, plan avSearchPlan) string {
	return fmt.Sprintf(
		"av|%s|%s|%s|%s|%s|%s",
		normalizeCacheKey(keyword),
		normalizeCacheKey(plan.SiteCategory),
		normalizeCacheKey(strings.Join(plan.Sources, ",")),
		normalizeCacheKey(plan.FilePath),
		normalizeCacheKey(plan.DetailURL),
		normalizeCacheKey(plan.OSHash),
	)
}

func previewAVUsedSourceFromCandidates(candidates []map[string]any, fallback string) string {
	for _, candidate := range candidates {
		source := normalizeAVSourceName(asString(candidate["scrape_source"]))
		if source != "" {
			return source
		}
	}
	return fallback
}

func resolveAVPreviewKeyword(title string, opts AVPreviewOptions) string {
	if detailURL := strings.TrimSpace(opts.DetailURL); detailURL != "" {
		if keyword := normalizeAVKeyword(detailURL); keyword != "" {
			return keyword
		}
	}
	if filePath := strings.TrimSpace(opts.FilePath); filePath != "" {
		if keyword := normalizeAVKeyword(filePath); keyword != "" {
			return keyword
		}
	}
	return normalizeAVKeyword(title)
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
	backdropPath, err := s.downloadMovieBackdrop(ctx, chooseStr(in.BackdropURL, asString(detail["backdrop_path"])), in.VideoID)
	if err != nil {
		return err
	}
	if backdropPath != "" {
		detail = cloneMetadataMap(detail)
		detail["backdrop_path"] = backdropPath
		detail["backdrop_url"] = backdropPath
	}
	meta := map[string]any{
		"tmdb":         detail,
		"manual":       in.Metadata,
		"release_date": chooseStr(in.ReleaseDate, asString(detail["release_date"])),
	}
	if backdropPath != "" {
		meta["backdrop_path"] = backdropPath
		meta["backdrop_url"] = backdropPath
	}
	localized := s.localizeScrapeFields(ctx, title, overview)
	meta = withLocalizedScrapeMetadata(meta, localized)
	if err := s.repo.UpdateVideoScrapeResult(ctx, video.ID, &in.TMDBID, localized.titleZH, localized.descriptionZH, thumbPath, meta, "uploaded"); err != nil {
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

	synced, err := s.syncTVSeriesAndBindEpisode(ctx, in.VideoID, in.TMDBID, seasonNum, episodeNum)
	if err != nil {
		return err
	}
	title := chooseStr(in.Title, asString(synced.episodeRaw["name"]))
	overview := chooseStr(in.Overview, asString(synced.episodeRaw["overview"]))
	posterURL := chooseStr(in.PosterURL, chooseStr(asString(synced.episodeRaw["still_path"]), asString(synced.tvRaw["poster_path"])))
	thumbPath, err := s.DownloadPoster(ctx, posterURL, in.VideoID)
	if err != nil {
		return err
	}
	meta := map[string]any{
		"tmdb_tv":      synced.tvRaw,
		"tmdb_season":  synced.seasonRaw,
		"tmdb_episode": synced.episodeRaw,
		"manual":       in.Metadata,
		"release_date": chooseStr(in.ReleaseDate, asString(synced.episodeRaw["air_date"])),
	}
	localized := s.localizeScrapeFields(ctx, title, overview)
	meta = withLocalizedScrapeMetadata(meta, localized)
	if err := s.repo.UpdateVideoScrapeResult(ctx, video.ID, &synced.episodeTMDBID, localized.titleZH, localized.descriptionZH, thumbPath, meta, "uploaded"); err != nil {
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

	sourceHint := normalizeAVSourceName(asString(in.Metadata["scrape_source"]))
	detailURL := strings.TrimSpace(asString(in.Metadata["detail_url"]))
	if detailURL == "" {
		detailURL = s.buildAVDetailURLBySource(sourceHint, externalID)
	}
	if detailURL == "" {
		detailURL = toAbsoluteURL(strings.TrimSpace(s.avBaseURL), "/v/"+externalID)
	}
	candidate, trace, err := s.fetchAVCandidateBySourceAndDetailURLWithTrace(ctx, sourceHint, detailURL)
	if err != nil {
		return err
	}
	if strings.TrimSpace(candidate.ExternalID) == "" {
		candidate.ExternalID = externalID
	}
	if strings.TrimSpace(candidate.Source) == "" {
		candidate.Source = sourceHint
	}
	scrapeSource := normalizeAVSourceName(candidate.Source)
	if scrapeSource == "" {
		scrapeSource = "javdb"
	}

	title := chooseStr(in.Title, candidate.Title)
	if strings.TrimSpace(title) == "" {
		title = strings.TrimSpace(video.Title)
	}
	overview := chooseStr(in.Overview, candidate.Overview)
	posterURL := chooseStr(in.PosterURL, candidate.PosterURL)
	posterSource := avPosterSourceFromCandidate(candidate)
	posterQuality := candidatePosterQuality(candidate)
	thumbURL := chooseStr(in.ThumbURL, asString(in.Metadata["thumb_url"]))
	thumbURL = chooseStr(thumbURL, asString(in.Metadata["poster_thumb_url"]))
	thumbURL = chooseStr(thumbURL, candidate.ThumbURL)
	thumbURL = chooseStr(thumbURL, asString(candidate.Raw["thumb_url"]))
	thumbURL = chooseStr(thumbURL, asString(candidate.Raw["poster_thumb_url"]))
	if strings.TrimSpace(in.PosterURL) != "" {
		posterSource = "manual"
		posterQuality = classifyAVPosterURL(posterURL, "")
	}
	siteConfig := s.loadAVScraperSiteConfig(ctx)
	posterAssets, posterDecision, err := s.resolveAVPosterAssets(ctx, in.VideoID, video.ThumbnailPath, posterURL, thumbURL, posterSource, posterQuality, siteConfig)
	if err != nil {
		return err
	}
	if trace == nil {
		trace = map[string]any{}
	}
	trace["poster_source"] = posterSource
	trace["poster_quality"] = posterQuality
	trace["poster_decision"] = posterDecision

	posterPath := strings.TrimSpace(posterAssets.OriginalPath)
	thumbPath := strings.TrimSpace(posterAssets.SelectedPath)
	if thumbPath == "" {
		thumbPath = posterPath
	}
	thumbFilePath := strings.TrimSpace(posterAssets.ThumbPath)
	if thumbFilePath == "" {
		thumbFilePath = strings.TrimSpace(posterAssets.CroppedPath)
	}

	siteCategory := normalizeAVSiteCategory(asString(in.Metadata["site_category"]))
	if siteCategory == "" {
		siteCategory = avSiteCategoryFromVideo(video, title)
	}

	meta := map[string]any{
		"scrape_source":             scrapeSource,
		"site_category":             siteCategory,
		"external_id":               candidate.ExternalID,
		"av_code":                   candidate.Code,
		"actors":                    candidate.Actors,
		"release_date":              chooseStr(in.ReleaseDate, candidate.ReleaseDate),
		"detail_url":                candidate.DetailURL,
		"poster_url":                posterURL,
		"poster":                    posterPath,
		"thumb":                     thumbPath,
		"thumb_url":                 thumbURL,
		"poster_source":             posterSource,
		"poster_quality":            posterQuality,
		"poster_decision":           posterDecision,
		"poster_original_path":      utils.VideoThumbnailURLWithVariant(video.ID, avPosterVariantOriginal),
		"poster_thumb_path":         posterVariantURL(video.ID, thumbFilePath, "thumb"),
		"poster_cropped_path":       posterVariantURL(video.ID, posterAssets.CroppedPath, avPosterVariantCropped),
		"poster_variant":            posterAssets.Variant,
		"poster_original_file_path": posterAssets.OriginalPath,
		"poster_thumb_file_path":    posterAssets.ThumbPath,
		"poster_cropped_file_path":  posterAssets.CroppedPath,
		scrapeSource:                candidate.Raw,
		"manual":                    in.Metadata,
		"scrape_trace":              trace,
	}
	if sources, ok := candidate.Raw["merged_sources"]; ok {
		meta["scrape_sources"] = sources
	}
	localized := s.localizeScrapeFields(ctx, title, overview)
	meta = withLocalizedScrapeMetadata(meta, localized)
	targetStatus := strings.TrimSpace(in.Status)
	if targetStatus == "" {
		targetStatus = "uploaded"
	}
	if video.Status == "av_scrape_pending" || targetStatus == "av_scrape_pending" {
		targetStatus = "processing"
	}
	if err := s.repo.UpdateVideoScrapeResult(ctx, video.ID, nil, localized.titleZH, localized.descriptionZH, thumbPath, meta, targetStatus); err != nil {
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
	backdropPath, err := s.downloadMovieBackdrop(ctx, asString(raw["backdrop_path"]), videoID)
	if err != nil {
		return err
	}
	if backdropPath != "" {
		raw = cloneMetadataMap(raw)
		raw["backdrop_path"] = backdropPath
		raw["backdrop_url"] = backdropPath
	}
	meta := map[string]any{
		"tmdb": raw,
	}
	if backdropPath != "" {
		meta["backdrop_path"] = backdropPath
		meta["backdrop_url"] = backdropPath
	}
	localized := s.localizeScrapeFields(ctx, title, overview)
	meta = withLocalizedScrapeMetadata(meta, localized)
	if err := s.repo.UpdateVideoScrapeResult(ctx, videoID, &tmdbID, localized.titleZH, localized.descriptionZH, thumbPath, meta, "uploaded"); err != nil {
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
	synced, err := s.syncTVSeriesAndBindEpisode(ctx, videoID, tmdbID, seasonNum, episodeNum)
	if err != nil {
		return err
	}
	title := chooseStr(asString(synced.episodeRaw["name"]), fmt.Sprintf("S%02dE%02d", seasonNum, episodeNum))
	description := asString(synced.episodeRaw["overview"])
	thumbPath, err := s.downloadTMDBImage(ctx, chooseStr(asString(synced.episodeRaw["still_path"]), asString(synced.tvRaw["poster_path"])), videoID)
	if err != nil {
		return err
	}
	meta := map[string]any{
		"tmdb_tv":      synced.tvRaw,
		"tmdb_season":  synced.seasonRaw,
		"tmdb_episode": synced.episodeRaw,
	}
	localized := s.localizeScrapeFields(ctx, title, description)
	meta = withLocalizedScrapeMetadata(meta, localized)
	if err := s.repo.UpdateVideoScrapeResult(ctx, videoID, &synced.episodeTMDBID, localized.titleZH, localized.descriptionZH, thumbPath, meta, "uploaded"); err != nil {
		return err
	}
	s.syncEpisodeActors(ctx, videoID, tmdbID, seasonNum, episodeNum)
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
	backdropPath, err := s.downloadMovieBackdrop(ctx, asString(detailRaw["backdrop_path"]), videoID)
	if err != nil {
		return ScrapeResult{}, err
	}
	if backdropPath != "" {
		detailRaw = cloneMetadataMap(detailRaw)
		detailRaw["backdrop_path"] = backdropPath
		detailRaw["backdrop_url"] = backdropPath
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
	if backdropPath != "" {
		scrape.Metadata["backdrop_path"] = backdropPath
		scrape.Metadata["backdrop_url"] = backdropPath
	}
	if scrape.Title == "" {
		scrape.Title = title
	}
	localized := s.localizeScrapeFields(ctx, scrape.Title, scrape.Description)
	scrape.Title = localized.titleZH
	scrape.Description = localized.descriptionZH
	scrape.Metadata = withLocalizedScrapeMetadata(scrape.Metadata, localized)

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
	title, year, seasonNum, episodeNum, ok := utils.ParseFilename(filename)
	if !ok || seasonNum <= 0 || episodeNum <= 0 || strings.TrimSpace(title) == "" {
		return ScrapeResult{}, &EpisodeAutoScrapeError{
			Stage: EpisodeAutoScrapeStageParseFailed,
			Err:   fmt.Errorf("cannot parse episode info from filename: %s", filename),
		}
	}

	candidates, err := s.PreviewTV(ctx, title, year)
	if err != nil {
		return ScrapeResult{}, &EpisodeAutoScrapeError{
			Stage:               EpisodeAutoScrapeStageAPIError,
			ParsedTitle:         title,
			ParsedSeasonNumber:  seasonNum,
			ParsedEpisodeNumber: episodeNum,
			Err:                 err,
		}
	}
	selected, err := selectReliableTVCandidate(title, candidates)
	if err != nil {
		var pendingErr *EpisodeAutoScrapeError
		if errors.As(err, &pendingErr) {
			pendingErr.ParsedTitle = title
			pendingErr.ParsedSeasonNumber = seasonNum
			pendingErr.ParsedEpisodeNumber = episodeNum
			return ScrapeResult{}, pendingErr
		}
		return ScrapeResult{}, &EpisodeAutoScrapeError{
			Stage:               EpisodeAutoScrapeStageAPIError,
			ParsedTitle:         title,
			ParsedSeasonNumber:  seasonNum,
			ParsedEpisodeNumber: episodeNum,
			Err:                 err,
		}
	}
	tvID := asInt(selected["tmdb_id"])
	if tvID <= 0 {
		return ScrapeResult{}, &EpisodeAutoScrapeError{
			Stage:               EpisodeAutoScrapeStageCandidateNotFound,
			ParsedTitle:         title,
			ParsedSeasonNumber:  seasonNum,
			ParsedEpisodeNumber: episodeNum,
			CandidateCount:      len(candidates),
			CandidatePreview:    summarizeAutoTVCandidates(candidates),
			Err:                 fmt.Errorf("invalid tv id from tmdb candidate"),
		}
	}
	synced, err := s.syncTVSeriesAndBindEpisode(ctx, videoID, tvID, seasonNum, episodeNum)
	if err != nil {
		return ScrapeResult{}, wrapEpisodeSyncError(err, title, seasonNum, episodeNum, candidates)
	}
	posterRel := chooseStr(asString(synced.episodeRaw["still_path"]), asString(synced.tvRaw["poster_path"]))
	thumbPath, err := s.downloadTMDBImage(ctx, posterRel, videoID)
	if err != nil {
		return ScrapeResult{}, &EpisodeAutoScrapeError{
			Stage:               EpisodeAutoScrapeStageAPIError,
			ParsedTitle:         title,
			ParsedSeasonNumber:  seasonNum,
			ParsedEpisodeNumber: episodeNum,
			CandidateCount:      len(candidates),
			CandidatePreview:    summarizeAutoTVCandidates(candidates),
			Err:                 err,
		}
	}

	episodeTitle := asString(synced.episodeRaw["name"])
	if episodeTitle == "" {
		episodeTitle = fmt.Sprintf("%s S%02dE%02d", title, seasonNum, episodeNum)
	}
	scrape := ScrapeResult{
		TMDBID:        synced.episodeTMDBID,
		Title:         episodeTitle,
		Description:   asString(synced.episodeRaw["overview"]),
		ThumbnailPath: thumbPath,
		Metadata: map[string]any{
			"tmdb_tv":      synced.tvRaw,
			"tmdb_season":  synced.seasonRaw,
			"tmdb_episode": synced.episodeRaw,
		},
	}
	localized := s.localizeScrapeFields(ctx, scrape.Title, scrape.Description)
	scrape.Title = localized.titleZH
	scrape.Description = localized.descriptionZH
	scrape.Metadata = withLocalizedScrapeMetadata(scrape.Metadata, localized)
	if err := s.repo.UpdateVideoScrapeResult(ctx, videoID, &scrape.TMDBID, scrape.Title, scrape.Description, scrape.ThumbnailPath, scrape.Metadata, "uploaded"); err != nil {
		return ScrapeResult{}, err
	}
	s.syncEpisodeActors(ctx, videoID, tvID, seasonNum, episodeNum)
	return scrape, nil
}

func (s *ScraperService) syncTVSeriesAndBindEpisode(ctx context.Context, videoID uuid.UUID, tvID, targetSeasonNum, targetEpisodeNum int) (syncedEpisodeResult, error) {
	tvRaw, err := s.fetchLocalizedMediaDetail(ctx, fmt.Sprintf("/tv/%d", tvID), "tv")
	if err != nil {
		return syncedEpisodeResult{}, err
	}
	seriesID, err := s.repo.UpsertSeries(
		ctx,
		tvID,
		chooseStr(asString(tvRaw["name"]), asString(tvRaw["original_name"])),
		asString(tvRaw["overview"]),
		asString(tvRaw["poster_path"]),
		asString(tvRaw["backdrop_path"]),
		parseDate(tvRaw["first_air_date"]),
		asInt(tvRaw["number_of_seasons"]),
		asInt(tvRaw["number_of_episodes"]),
	)
	if err != nil {
		return syncedEpisodeResult{}, err
	}
	_ = s.downloadTVSeriesArtwork(ctx, asString(tvRaw["poster_path"]), seriesID, "poster")
	_ = s.downloadTVSeriesArtwork(ctx, asString(tvRaw["backdrop_path"]), seriesID, "backdrop")

	seasonNumbers := extractTVSeasonNumbers(tvRaw)
	if len(seasonNumbers) == 0 {
		for seasonNumber := 1; seasonNumber <= asInt(tvRaw["number_of_seasons"]); seasonNumber++ {
			seasonNumbers = append(seasonNumbers, seasonNumber)
		}
	}

	result := syncedEpisodeResult{tvRaw: tvRaw}
	for _, seasonNumber := range seasonNumbers {
		if seasonNumber <= 0 {
			continue
		}
		seasonRaw, seasonErr := s.fetchLocalizedSeasonDetail(ctx, tvID, seasonNumber)
		if seasonErr != nil {
			return syncedEpisodeResult{}, seasonErr
		}
		seasonID, seasonErr := s.repo.UpsertSeason(
			ctx,
			seriesID,
			seasonNumber,
			asString(seasonRaw["name"]),
			asString(seasonRaw["overview"]),
			asString(seasonRaw["poster_path"]),
			parseDate(seasonRaw["air_date"]),
		)
		if seasonErr != nil {
			return syncedEpisodeResult{}, seasonErr
		}

		episodes, _ := seasonRaw["episodes"].([]any)
		for _, row := range episodes {
			epRaw, ok := row.(map[string]any)
			if !ok {
				continue
			}
			episodeNumber := asInt(epRaw["episode_number"])
			if episodeNumber <= 0 {
				continue
			}
			bindVideo := seasonNumber == targetSeasonNum && episodeNumber == targetEpisodeNum
			boundVideoID := uuid.Nil
			if bindVideo {
				boundVideoID = videoID
			}
			if err := s.repo.UpsertEpisode(
				ctx,
				seasonID,
				episodeNumber,
				asString(epRaw["name"]),
				asString(epRaw["overview"]),
				asString(epRaw["still_path"]),
				asInt(epRaw["runtime"]),
				parseDate(epRaw["air_date"]),
				boundVideoID,
				bindVideo,
			); err != nil {
				return syncedEpisodeResult{}, err
			}
			if bindVideo {
				result.seasonRaw = seasonRaw
				result.episodeRaw = epRaw
				result.episodeTMDBID = asInt(epRaw["id"])
			}
		}
	}

	if result.episodeRaw == nil {
		return syncedEpisodeResult{}, fmt.Errorf("episode S%02dE%02d not found in tmdb", targetSeasonNum, targetEpisodeNum)
	}
	if result.episodeTMDBID <= 0 {
		return syncedEpisodeResult{}, fmt.Errorf("invalid episode tmdb id")
	}
	return result, nil
}

func selectReliableTVCandidate(parsedTitle string, candidates []map[string]any) (map[string]any, error) {
	if len(candidates) == 0 {
		return nil, &EpisodeAutoScrapeError{
			Stage: EpisodeAutoScrapeStageCandidateNotFound,
			Err:   fmt.Errorf("no tv candidate found"),
		}
	}
	normalizedParsed := normalizeAutoMatchTitle(parsedTitle)
	if normalizedParsed == "" {
		return nil, &EpisodeAutoScrapeError{
			Stage:            EpisodeAutoScrapeStageCandidateNotFound,
			CandidateCount:   len(candidates),
			CandidatePreview: summarizeAutoTVCandidates(candidates),
			Err:              fmt.Errorf("parsed title is empty"),
		}
	}

	exactMatches := make([]map[string]any, 0, len(candidates))
	for _, candidate := range candidates {
		if autoTVCandidateMatches(normalizedParsed, candidate) {
			exactMatches = append(exactMatches, candidate)
		}
	}
	if len(exactMatches) == 1 {
		return exactMatches[0], nil
	}
	stage := EpisodeAutoScrapeStageCandidateAmbiguous
	if len(exactMatches) == 0 {
		stage = EpisodeAutoScrapeStageCandidateNotFound
	}
	previewSource := exactMatches
	if len(previewSource) == 0 {
		previewSource = candidates
	}
	return nil, &EpisodeAutoScrapeError{
		Stage:            stage,
		CandidateCount:   len(previewSource),
		CandidatePreview: summarizeAutoTVCandidates(previewSource),
		Err:              fmt.Errorf("unable to find unique reliable tv candidate for %q", parsedTitle),
	}
}

func autoTVCandidateMatches(normalizedParsed string, candidate map[string]any) bool {
	for _, raw := range autoTVCandidateTitles(candidate) {
		if normalizeAutoMatchTitle(raw) == normalizedParsed {
			return true
		}
	}
	return false
}

func autoTVCandidateTitles(candidate map[string]any) []string {
	out := []string{
		asString(candidate["title"]),
		asString(candidate["original_title"]),
	}
	if metadata, ok := candidate["metadata"].(map[string]any); ok {
		out = append(out, asString(metadata["name"]), asString(metadata["original_name"]))
	}
	return out
}

func normalizeAutoMatchTitle(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(normalized))
	for _, r := range normalized {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func summarizeAutoTVCandidates(candidates []map[string]any) []map[string]any {
	limit := len(candidates)
	if limit > 5 {
		limit = 5
	}
	out := make([]map[string]any, 0, limit)
	for i := 0; i < limit; i++ {
		item := candidates[i]
		out = append(out, map[string]any{
			"tmdb_id":        item["tmdb_id"],
			"title":          chooseStr(asString(item["title"]), asString(item["original_title"])),
			"original_title": asString(item["original_title"]),
			"release_date":   asString(item["release_date"]),
		})
	}
	return out
}

func wrapEpisodeSyncError(err error, parsedTitle string, seasonNum, episodeNum int, candidates []map[string]any) error {
	if err == nil {
		return nil
	}
	stage := EpisodeAutoScrapeStageAPIError
	if strings.Contains(strings.ToLower(err.Error()), "episode s") || strings.Contains(strings.ToLower(err.Error()), "episode tmdb id") {
		stage = EpisodeAutoScrapeStageTMDBMissingEpisode
	}
	return &EpisodeAutoScrapeError{
		Stage:               stage,
		ParsedTitle:         parsedTitle,
		ParsedSeasonNumber:  seasonNum,
		ParsedEpisodeNumber: episodeNum,
		CandidateCount:      len(candidates),
		CandidatePreview:    summarizeAutoTVCandidates(candidates),
		Err:                 err,
	}
}

func extractTVSeasonNumbers(tvRaw map[string]any) []int {
	rows, _ := tvRaw["seasons"].([]any)
	seen := map[int]struct{}{}
	out := make([]int, 0, len(rows))
	for _, row := range rows {
		item, ok := row.(map[string]any)
		if !ok {
			continue
		}
		seasonNumber := asInt(item["season_number"])
		if seasonNumber <= 0 {
			continue
		}
		if _, exists := seen[seasonNumber]; exists {
			continue
		}
		seen[seasonNumber] = struct{}{}
		out = append(out, seasonNumber)
	}
	slices.Sort(out)
	return out
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

	video, err := s.repo.GetVideoByID(ctx, videoID)
	if err != nil {
		return ScrapeResult{}, err
	}
	siteCategory := avSiteCategoryFromVideo(video, keyword)
	plan, err := s.resolveAVSearchPlan(ctx, keyword, AVPreviewOptions{
		SiteCategory: siteCategory,
		FilePath:     filePath,
		OSHash:       video.OSHash,
	})
	if err != nil {
		return ScrapeResult{}, err
	}
	if plan.SiteCategory == avSiteCategoryWestern {
		return s.scrapeAVWesternUpload(ctx, video, keyword, plan)
	}
	candidates, trace, err := s.searchAVCandidatesWithTrace(ctx, keyword, 6, plan)
	if err != nil {
		return ScrapeResult{}, err
	}
	if len(candidates) == 0 {
		return ScrapeResult{}, fmt.Errorf("no av result for %q", keyword)
	}
	candidate := candidates[0]
	posterSource := avPosterSourceFromCandidate(candidate)
	posterQuality := candidatePosterQuality(candidate)
	thumbURL := chooseStr(candidate.ThumbURL, asString(candidate.Raw["thumb_url"]))
	thumbURL = chooseStr(thumbURL, asString(candidate.Raw["poster_thumb_url"]))
	siteConfig := s.loadAVScraperSiteConfig(ctx)
	posterAssets, posterDecision, err := s.resolveAVPosterAssets(ctx, videoID, video.ThumbnailPath, candidate.PosterURL, thumbURL, posterSource, posterQuality, siteConfig)
	if err != nil {
		return ScrapeResult{}, err
	}
	if trace == nil {
		trace = map[string]any{}
	}
	trace["poster_source"] = posterSource
	trace["poster_quality"] = posterQuality
	trace["poster_decision"] = posterDecision

	posterPath := strings.TrimSpace(posterAssets.OriginalPath)
	thumbPath := strings.TrimSpace(posterAssets.SelectedPath)
	if thumbPath == "" {
		thumbPath = posterPath
	}
	thumbFilePath := strings.TrimSpace(posterAssets.ThumbPath)
	if thumbFilePath == "" {
		thumbFilePath = strings.TrimSpace(posterAssets.CroppedPath)
	}

	title := strings.TrimSpace(candidate.Title)
	if title == "" {
		title = keyword
	}
	scrapeSource := normalizeAVSourceName(candidate.Source)
	if scrapeSource == "" {
		scrapeSource = "javdb"
	}
	meta := map[string]any{
		"scrape_source":             scrapeSource,
		"site_category":             plan.SiteCategory,
		"external_id":               candidate.ExternalID,
		"av_code":                   candidate.Code,
		"actors":                    candidate.Actors,
		"release_date":              candidate.ReleaseDate,
		"detail_url":                candidate.DetailURL,
		"poster_url":                candidate.PosterURL,
		"poster":                    posterPath,
		"thumb":                     thumbPath,
		"thumb_url":                 thumbURL,
		"poster_source":             posterSource,
		"poster_quality":            posterQuality,
		"poster_decision":           posterDecision,
		"poster_original_path":      utils.VideoThumbnailURLWithVariant(videoID, avPosterVariantOriginal),
		"poster_thumb_path":         posterVariantURL(videoID, thumbFilePath, "thumb"),
		"poster_cropped_path":       posterVariantURL(videoID, posterAssets.CroppedPath, avPosterVariantCropped),
		"poster_variant":            posterAssets.Variant,
		"poster_original_file_path": posterAssets.OriginalPath,
		"poster_thumb_file_path":    posterAssets.ThumbPath,
		"poster_cropped_file_path":  posterAssets.CroppedPath,
		"search_keyword":            keyword,
		scrapeSource:                candidate.Raw,
		"scrape_trace":              trace,
	}
	if sources, ok := candidate.Raw["merged_sources"]; ok {
		meta["scrape_sources"] = sources
	}
	scrape := ScrapeResult{
		TMDBID:        0,
		Title:         title,
		Description:   strings.TrimSpace(candidate.Overview),
		ThumbnailPath: thumbPath,
		Metadata:      meta,
	}
	localized := s.localizeScrapeFields(ctx, scrape.Title, scrape.Description)
	scrape.Title = localized.titleZH
	scrape.Description = localized.descriptionZH
	scrape.Metadata = withLocalizedScrapeMetadata(scrape.Metadata, localized)
	if err := s.repo.UpdateVideoScrapeResult(ctx, videoID, nil, scrape.Title, scrape.Description, scrape.ThumbnailPath, scrape.Metadata, "uploaded"); err != nil {
		return ScrapeResult{}, err
	}
	s.syncAVActors(ctx, videoID, candidate.Actors)
	return scrape, nil
}

func (s *ScraperService) scrapeAVWesternUpload(ctx context.Context, video models.Video, keyword string, plan avSearchPlan) (ScrapeResult, error) {
	candidates, trace, scrapeErr := s.searchAVCandidatesWithTrace(ctx, keyword, 6, plan)
	if trace == nil {
		trace = map[string]any{}
	}
	trace["site_category"] = avSiteCategoryWestern
	trace["hash_used"] = strings.TrimSpace(plan.OSHash) != ""
	if strings.TrimSpace(plan.OSHash) != "" {
		trace["os_hash"] = strings.TrimSpace(plan.OSHash)
	}
	if scrapeErr != nil {
		trace["error"] = scrapeErr.Error()
	}

	preview := make([]map[string]any, 0, len(candidates))
	for _, candidate := range candidates {
		preview = append(preview, avScrapePreviewCandidateMap(candidate, plan, trace))
	}

	meta := metadataMapFromRaw(video.Metadata)
	meta["site_category"] = avSiteCategoryWestern
	meta["scrape_preview"] = preview
	meta["scrape_attempt"] = trace

	title := strings.TrimSpace(video.Title)
	if title == "" {
		title = keyword
	}
	description := strings.TrimSpace(video.Description)
	if err := s.repo.UpdateVideoScrapeResult(ctx, video.ID, nil, title, description, video.ThumbnailPath, meta, "av_scrape_pending"); err != nil {
		return ScrapeResult{}, err
	}
	return ScrapeResult{
		TMDBID:        0,
		Title:         title,
		Description:   description,
		ThumbnailPath: video.ThumbnailPath,
		Metadata:      meta,
	}, nil
}

func avScrapePreviewCandidateMap(candidate avScrapeCandidate, plan avSearchPlan, trace map[string]any) map[string]any {
	scrapeSource := normalizeAVSourceName(candidate.Source)
	if scrapeSource == "" {
		scrapeSource = plan.RecommendedSource
	}
	matchSource := avCandidateMatchSource(candidate)
	metadata := cloneMetadataMap(candidate.Raw)
	metadata["site_category"] = plan.SiteCategory
	metadata["scrape_source"] = scrapeSource
	metadata["match_source"] = matchSource
	if trace != nil {
		metadata["scrape_trace"] = trace
	}
	return map[string]any{
		"external_id":     candidate.ExternalID,
		"av_code":         candidate.Code,
		"title":           candidate.Title,
		"original_title":  candidate.Title,
		"overview":        candidate.Overview,
		"poster_url":      candidate.PosterURL,
		"thumb_url":       candidate.ThumbURL,
		"release_date":    candidate.ReleaseDate,
		"actors":          candidate.Actors,
		"detail_url":      candidate.DetailURL,
		"metadata":        metadata,
		"media_type_hint": "av",
		"scrape_source":   scrapeSource,
		"site_category":   plan.SiteCategory,
		"match_source":    matchSource,
	}
}

func avCandidateMatchSource(candidate avScrapeCandidate) string {
	if candidate.Raw != nil {
		if raw := strings.TrimSpace(asString(candidate.Raw["match_source"])); raw != "" {
			return raw
		}
	}
	if normalizeAVSourceName(candidate.Source) == "theporndb" {
		detailPath := strings.ToLower(strings.TrimSpace(candidate.DetailURL))
		switch {
		case strings.Contains(detailPath, "/movies/"):
			return "keyword:movies"
		case strings.Contains(detailPath, "/scenes/"):
			return "keyword:scenes"
		}
	}
	source := normalizeAVSourceName(candidate.Source)
	if source == "" {
		source = "unknown"
	}
	return "keyword:" + source
}

func metadataMapFromRaw(raw []byte) map[string]any {
	out := map[string]any{}
	if len(raw) == 0 {
		return out
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return out
	}
	for key, value := range parsed {
		out[key] = value
	}
	return out
}

func avSiteCategoryFromVideo(video models.Video, fallbackTitle string) string {
	meta := metadataMapFromRaw(video.Metadata)
	if category := normalizeAVSiteCategory(asString(meta["site_category"])); category != "" {
		return category
	}
	if strings.TrimSpace(video.OSHash) != "" {
		return avSiteCategoryWestern
	}
	category := detectAVSiteCategory(fallbackTitle)
	if category == "" {
		return avSiteCategoryJapanese
	}
	return category
}

func (s *ScraperService) syncMovieActors(ctx context.Context, videoID uuid.UUID, movieTMDBID int) {
	if movieTMDBID <= 0 {
		return
	}
	path := fmt.Sprintf("/movie/%d/credits", movieTMDBID)
	cast := s.tryFetchCastMembers(ctx, path)
	s.syncTMDBCastActors(ctx, videoID, cast)
}

func (s *ScraperService) syncEpisodeActors(ctx context.Context, videoID uuid.UUID, tvTMDBID, seasonNum, episodeNum int) {
	if tvTMDBID <= 0 || seasonNum <= 0 || episodeNum <= 0 {
		return
	}
	episodeCreditsPath := fmt.Sprintf("/tv/%d/season/%d/episode/%d/credits", tvTMDBID, seasonNum, episodeNum)
	cast := s.tryFetchCastMembers(ctx, episodeCreditsPath)
	if len(cast) == 0 {
		tvCreditsPath := fmt.Sprintf("/tv/%d/credits", tvTMDBID)
		cast = s.tryFetchCastMembers(ctx, tvCreditsPath)
	}
	s.syncTMDBCastActors(ctx, videoID, cast)
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
	_ = s.completeAVActorAvatars(ctx, videoID)
}

func (s *ScraperService) searchAVCandidates(ctx context.Context, keyword string, limit int) ([]avScrapeCandidate, error) {
	plan, err := s.resolveAVSearchPlan(ctx, keyword, AVPreviewOptions{})
	if err != nil {
		return nil, err
	}
	candidates, _, err := s.searchAVCandidatesWithTrace(ctx, keyword, limit, plan)
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

type tmdbCastMember struct {
	TMDBID      int
	Name        string
	ProfilePath string
}

func (s *ScraperService) tryFetchCastMembers(ctx context.Context, path string) []tmdbCastMember {
	raw, err := s.getTMDBJSON(ctx, path, nil, tmdbLangChinese)
	if err == nil {
		if cast := extractTMDBCastMembers(raw); len(cast) > 0 {
			return cast
		}
	}
	fallback, fallbackErr := s.getTMDBJSON(ctx, path, nil, "")
	if fallbackErr != nil {
		return nil
	}
	return extractTMDBCastMembers(fallback)
}

func (s *ScraperService) syncTMDBCastActors(ctx context.Context, videoID uuid.UUID, cast []tmdbCastMember) {
	if len(cast) == 0 {
		return
	}
	actorIDs := make([]uuid.UUID, 0, len(cast))
	for _, member := range cast {
		actor, err := s.upsertTMDBActorProfile(ctx, member)
		if err != nil || actor.ID == uuid.Nil {
			continue
		}
		actorIDs = append(actorIDs, actor.ID)
	}
	if len(actorIDs) == 0 {
		return
	}
	_ = s.repo.AddVideoActors(ctx, videoID, actorIDs, "scrape_tmdb")
}

func (s *ScraperService) upsertTMDBActorProfile(ctx context.Context, member tmdbCastMember) (models.AdminActor, error) {
	detail := map[string]any{}
	if member.TMDBID > 0 {
		var err error
		detail, err = s.getTMDBJSON(ctx, fmt.Sprintf("/person/%d", member.TMDBID), nil, tmdbLangChinese)
		if err != nil {
			detail = map[string]any{}
		} else if needsPersonLocalizedFallback(detail) {
			if fallback, fallbackErr := s.getTMDBJSON(ctx, fmt.Sprintf("/person/%d", member.TMDBID), nil, ""); fallbackErr == nil {
				detail = mergeLocalizedPersonDetail(detail, fallback)
			}
		}
	}
	input := tmdbActorInputFromDetail(member, detail)
	if strings.TrimSpace(input.Name) == "" {
		return models.AdminActor{}, nil
	}
	avatarURL := strings.TrimSpace(input.AvatarURL)
	initialInput := input
	initialInput.AvatarURL = ""
	actor, err := s.repo.UpsertScrapedActorProfile(ctx, initialInput)
	if err != nil {
		return models.AdminActor{}, err
	}
	if strings.TrimSpace(actor.AvatarURL) != "" || avatarURL == "" {
		return actor, nil
	}
	localURL, err := s.downloadActorAvatar(ctx, actor.ID, avatarURL)
	if err != nil {
		return actor, nil
	}
	updated := input
	updated.AvatarURL = localURL
	actor, err = s.repo.UpsertScrapedActorProfile(ctx, updated)
	if err != nil {
		return models.AdminActor{}, err
	}
	return actor, nil
}

func tmdbActorInputFromDetail(member tmdbCastMember, detail map[string]any) models.AdminActorInput {
	name := strings.TrimSpace(asString(detail["name"]))
	if name == "" {
		name = member.Name
	}
	profilePath := strings.TrimSpace(asString(detail["profile_path"]))
	if profilePath == "" {
		profilePath = member.ProfilePath
	}
	avatarURL := ""
	if profilePath != "" {
		avatarURL = "https://image.tmdb.org/t/p/w500" + profilePath
	}
	externalID := ""
	if id := asInt(detail["id"]); id > 0 {
		externalID = strconv.Itoa(id)
	} else if member.TMDBID > 0 {
		externalID = strconv.Itoa(member.TMDBID)
	}
	return models.AdminActorInput{
		Name:       name,
		Aliases:    extractStringSlice(detail["also_known_as"]),
		Gender:     tmdbGenderText(asInt(detail["gender"])),
		Country:    normalizeTMDBPlace(asString(detail["place_of_birth"])),
		BirthDate:  strings.TrimSpace(asString(detail["birthday"])),
		AvatarURL:  avatarURL,
		Source:     "scrape_tmdb",
		ExternalID: externalID,
		Notes:      strings.TrimSpace(asString(detail["biography"])),
		Active:     true,
	}
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

func isAbsoluteHTTPURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	return strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://")
}

func (s *ScraperService) resolveAVThumbnailPath(ctx context.Context, videoID uuid.UUID, existingThumbPath, posterURL, thumbURL, posterSource, posterQuality string) (string, string, error) {
	assets, decision, err := s.resolveAVPosterAssets(ctx, videoID, existingThumbPath, posterURL, thumbURL, posterSource, posterQuality, defaultAVScraperSiteConfig())
	return assets.SelectedPath, decision, err
}

func (s *ScraperService) DownloadPoster(ctx context.Context, posterURL string, videoID uuid.UUID) (string, error) {
	posterURL = strings.TrimSpace(posterURL)
	if posterURL == "" {
		return "", nil
	}
	root := s.posterRoot
	if strings.TrimSpace(root) == "" {
		root = filepath.Join(s.storageRoot, "posters")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", fmt.Errorf("create poster root: %w", err)
	}
	outputPath := filepath.Join(root, videoID.String()+".jpg")
	if err := s.downloadPosterToPath(ctx, posterURL, outputPath); err != nil {
		return "", err
	}
	return outputPath, nil
}

func (s *ScraperService) downloadPosterToPath(ctx context.Context, posterURL, outputPath string) error {
	imageURL := strings.TrimSpace(posterURL)
	if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		imageURL = "https://image.tmdb.org/t/p/original" + imageURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return fmt.Errorf("create poster request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download poster failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("poster status=%d body=%s", resp.StatusCode, string(body))
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create poster file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("write poster file: %w", err)
	}
	return nil
}

func (s *ScraperService) downloadTMDBImage(ctx context.Context, relativePath string, videoID uuid.UUID) (string, error) {
	return s.downloadTMDBImageFile(ctx, relativePath, videoID, "poster.jpg")
}

func (s *ScraperService) downloadMovieBackdrop(ctx context.Context, relativePath string, videoID uuid.UUID) (string, error) {
	return s.downloadTMDBImageFile(ctx, relativePath, videoID, "backdrop.jpg")
}

func (s *ScraperService) downloadTMDBImageFile(ctx context.Context, relativePath string, videoID uuid.UUID, filename string) (string, error) {
	relativePath = strings.TrimSpace(relativePath)
	if relativePath == "" {
		return "", nil
	}
	filename = strings.TrimSpace(filename)
	if filename == "" {
		filename = "poster.jpg"
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
	outputPath := filepath.Join(outputDir, filename)
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

func cloneMetadataMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func (s *ScraperService) downloadTVSeriesArtwork(ctx context.Context, relativePath string, seriesID int64, kind string) error {
	relativePath = strings.TrimSpace(relativePath)
	if relativePath == "" || seriesID <= 0 {
		return nil
	}
	imageURL := relativePath
	if !strings.HasPrefix(relativePath, "http://") && !strings.HasPrefix(relativePath, "https://") {
		imageURL = "https://image.tmdb.org/t/p/original" + relativePath
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return fmt.Errorf("create tv artwork request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download tv artwork failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("tv artwork status=%d body=%s", resp.StatusCode, string(body))
	}

	outputDir := filepath.Join(s.storageRoot, "tv", "series", strconv.FormatInt(seriesID, 10))
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create tv artwork dir: %w", err)
	}
	outputPath := filepath.Join(outputDir, kind+".jpg")
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create tv artwork file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("write tv artwork file: %w", err)
	}
	return nil
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

func extractTMDBCastMembers(raw map[string]any) []tmdbCastMember {
	castRows, ok := raw["cast"].([]any)
	if !ok || len(castRows) == 0 {
		return nil
	}
	out := make([]tmdbCastMember, 0, len(castRows))
	seen := map[string]struct{}{}
	for _, row := range castRows {
		item, ok := row.(map[string]any)
		if !ok {
			continue
		}
		name := strings.TrimSpace(asString(item["name"]))
		if name == "" {
			name = strings.TrimSpace(asString(item["original_name"]))
		}
		personID := asInt(item["id"])
		if name == "" && personID <= 0 {
			continue
		}
		key := strings.ToLower(name)
		if personID > 0 {
			key = strconv.Itoa(personID)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, tmdbCastMember{
			TMDBID:      personID,
			Name:        name,
			ProfilePath: strings.TrimSpace(asString(item["profile_path"])),
		})
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
