package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/utils"
)

type AVScraperSiteConfig = models.AVScraperSiteConfig

type avScraperConfigStore interface {
	GetAVScraperConfig(ctx context.Context) (AVScraperSiteConfig, error)
}

type AVPreviewOptions struct {
	BypassCache  bool
	SiteCategory string
	SiteSource   string
	FilePath     string
	DetailURL    string
}

type AVPreviewResult struct {
	Candidates        []map[string]any
	SiteCategory      string
	RecommendedSource string
	UsedSource        string
	EnabledSources    []string
}

type avSearchPlan struct {
	SiteCategory      string
	RecommendedSource string
	Sources           []string
	ExplicitSource    bool
	Config            AVScraperSiteConfig
	FilePath          string
	DetailURL         string
}

const (
	avSiteCategoryFC2      = "fc2"
	avSiteCategoryWestern  = "western"
	avSiteCategoryJapanese = "japanese"

	avPosterVariantOriginal = "original"
	avPosterVariantCropped  = "cropped"
	avPosterCropModeCenter  = "portrait_center"
	avPosterCropModeLeft    = "portrait_left"
	avPosterCropModeRight   = "portrait_right"
)

var errAVScraperConfigNotFound = errors.New("av scraper config not found")

func defaultAVScraperSiteConfig() AVScraperSiteConfig {
	return AVScraperSiteConfig{
		EnabledSites: defaultAVEnabledSites(),
		CategorySiteOrder: map[string][]string{
			avSiteCategoryFC2:      {"fc2", "fc2club", "fc2hub", "fc2ppvdb", "javdb"},
			avSiteCategoryWestern:  {"theporndb", "javdb", "mywife"},
			avSiteCategoryJapanese: {"javdb", "javbus", "javlibrary", "airav_cc", "avsex", "theporndb", "getchu"},
		},
		PosterCropEnabled: true,
		PosterCropMode:    avPosterCropModeCenter,
	}
}

func defaultAVEnabledSites() []string {
	return []string{
		"javdb",
		"javbus",
		"javlibrary",
		"airav_cc",
		"avsex",
		"theporndb",
		"getchu",
		"mywife",
		"fc2",
		"fc2club",
		"fc2hub",
		"fc2ppvdb",
	}
}

func normalizeAVSiteCategory(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case avSiteCategoryFC2:
		return avSiteCategoryFC2
	case avSiteCategoryWestern:
		return avSiteCategoryWestern
	case avSiteCategoryJapanese, "default", "jp":
		return avSiteCategoryJapanese
	case "":
		return ""
	default:
		return ""
	}
}

func detectAVSiteCategory(title string) string {
	keyword := strings.TrimSpace(title)
	lower := strings.ToLower(keyword)
	if strings.Contains(lower, "fc2") || normalizeFC2NumericID(keyword) != "" {
		return avSiteCategoryFC2
	}
	if looksLikeWesternAVTitle(keyword) {
		return avSiteCategoryWestern
	}
	return avSiteCategoryJapanese
}

func looksLikeWesternAVTitle(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	lower := strings.ToLower(raw)
	if strings.Contains(lower, "fc2") {
		return false
	}
	alphaCount := 0
	wordCount := 0
	hasDigit := false
	cjkCount := 0
	inWord := false
	for _, r := range raw {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z':
			alphaCount++
			if !inWord {
				wordCount++
				inWord = true
			}
		case r >= '0' && r <= '9':
			hasDigit = true
			inWord = false
		case r == ' ' || r == '-' || r == '_':
			inWord = false
		case r >= 0x4E00 && r <= 0x9FFF:
			cjkCount++
			inWord = false
		default:
			inWord = false
		}
	}
	if cjkCount > 0 {
		return false
	}
	if hasDigit && extractAVCode(raw) != "" {
		return false
	}
	return alphaCount >= 8 && wordCount >= 2
}

func (s *ScraperService) loadAVScraperSiteConfig(ctx context.Context) AVScraperSiteConfig {
	cfg := defaultAVScraperSiteConfig()
	if s == nil || s.avConfigStore == nil {
		return cfg
	}
	stored, err := s.avConfigStore.GetAVScraperConfig(ctx)
	if err != nil {
		return cfg
	}
	if len(stored.EnabledSites) > 0 {
		cfg.EnabledSites = normalizeSourceList(stored.EnabledSites)
	}
	if len(stored.CategorySiteOrder) > 0 {
		cfg.CategorySiteOrder = map[string][]string{}
		for rawCategory, rawSources := range stored.CategorySiteOrder {
			category := normalizeAVSiteCategory(rawCategory)
			if category == "" {
				continue
			}
			cfg.CategorySiteOrder[category] = normalizeSourceList(rawSources)
		}
	}
	cfg.PosterCropMode = normalizeAVPosterCropMode(stored.PosterCropMode)
	cfg.PosterCropEnabled = stored.PosterCropEnabled || !stored.PosterCropConfigured
	cfg.PosterCropConfigured = true
	return cfg
}

func normalizeAVPosterCropMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case avPosterCropModeLeft:
		return avPosterCropModeLeft
	case avPosterCropModeRight:
		return avPosterCropModeRight
	case avPosterCropModeCenter:
		return avPosterCropModeCenter
	default:
		return avPosterCropModeCenter
	}
}

func normalizeSourceList(raw []string) []string {
	out := make([]string, 0, len(raw))
	seen := map[string]struct{}{}
	for _, item := range raw {
		source := normalizeAVSourceName(item)
		if source == "" {
			source = strings.ToLower(strings.TrimSpace(item))
		}
		if source == "" {
			continue
		}
		if _, ok := seen[source]; ok {
			continue
		}
		seen[source] = struct{}{}
		out = append(out, source)
	}
	return out
}

func (s *ScraperService) resolveAVSearchPlan(ctx context.Context, title string, opts AVPreviewOptions) (avSearchPlan, error) {
	cfg := s.loadAVScraperSiteConfig(ctx)
	category := normalizeAVSiteCategory(opts.SiteCategory)
	if category == "" {
		category = detectAVSiteCategory(title)
	}

	explicitSource := normalizeAVSourceName(opts.SiteSource)
	if explicitSource == "" {
		explicitSource = strings.ToLower(strings.TrimSpace(opts.SiteSource))
	}
	if explicitSource != "" {
		return avSearchPlan{
			SiteCategory:      category,
			RecommendedSource: explicitSource,
			Sources:           []string{explicitSource},
			ExplicitSource:    true,
			Config:            cfg,
			FilePath:          strings.TrimSpace(opts.FilePath),
			DetailURL:         strings.TrimSpace(opts.DetailURL),
		}, nil
	}

	enabled := normalizeSourceList(cfg.EnabledSites)
	enabledSet := map[string]struct{}{}
	for _, site := range enabled {
		enabledSet[site] = struct{}{}
	}

	sources := normalizeSourceList(cfg.CategorySiteOrder[category])
	if len(sources) == 0 {
		sources = normalizeSourceList(defaultAVScraperSiteConfig().CategorySiteOrder[category])
	}
	filtered := make([]string, 0, len(sources))
	for _, source := range sources {
		if len(enabledSet) > 0 {
			if _, ok := enabledSet[source]; !ok {
				continue
			}
		}
		filtered = append(filtered, source)
	}
	if len(filtered) == 0 {
		filtered = sources
	}
	if len(filtered) == 0 {
		filtered = normalizeSourceList(defaultAVEnabledSites())
	}

	return avSearchPlan{
		SiteCategory:      category,
		RecommendedSource: filtered[0],
		Sources:           filtered,
		Config:            cfg,
		FilePath:          strings.TrimSpace(opts.FilePath),
		DetailURL:         strings.TrimSpace(opts.DetailURL),
	}, nil
}

func posterVariantURL(videoID uuid.UUID, filePath, variant string) string {
	if strings.TrimSpace(filePath) == "" {
		return ""
	}
	return utils.VideoThumbnailURLWithVariant(videoID, variant)
}
