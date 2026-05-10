package services

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"video-server/internal/models"
)

type fakeAVCrawlerProviderForExecution struct {
	crawlers []avCrawler
}

func (p *fakeAVCrawlerProviderForExecution) Crawlers() []avCrawler {
	out := make([]avCrawler, len(p.crawlers))
	copy(out, p.crawlers)
	return out
}

func (p *fakeAVCrawlerProviderForExecution) Default() avCrawler {
	if len(p.crawlers) == 0 {
		return nil
	}
	return p.crawlers[0]
}

type fakeAVCrawlerForExecution struct {
	name         string
	searchResult []avScrapeCandidate
	searchErr    error
	searchCalls  int
	searchQueries []string
	detailCalls   int
	detailURLs    []string
}

func (c *fakeAVCrawlerForExecution) Name() string {
	return c.name
}

func (c *fakeAVCrawlerForExecution) SearchCandidates(_ context.Context, _ *avScrapeRunContext, query string, _ int) ([]avScrapeCandidate, error) {
	c.searchCalls++
	c.searchQueries = append(c.searchQueries, query)
	if c.searchErr != nil {
		return nil, c.searchErr
	}
	out := make([]avScrapeCandidate, len(c.searchResult))
	copy(out, c.searchResult)
	return out, nil
}

func (c *fakeAVCrawlerForExecution) FetchByDetailURL(_ context.Context, _ *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
	c.detailCalls++
	c.detailURLs = append(c.detailURLs, detailURL)
	if c.searchErr != nil {
		return avScrapeCandidate{}, c.searchErr
	}
	if len(c.searchResult) > 0 {
		return c.searchResult[0], nil
	}
	return avScrapeCandidate{Source: c.name, DetailURL: detailURL}, nil
}

func TestPreviewAVSearchUsesExplicitSiteSourceOnly(t *testing.T) {
	t.Parallel()

	javLibrary := &fakeAVCrawlerForExecution{
		name: "javlibrary",
		searchResult: []avScrapeCandidate{{
			Source:     "javlibrary",
			ExternalID: "javlib-123",
			Code:       "SSIS-123",
			Title:      "JavLibrary Result",
			DetailURL:  "https://www.javlibrary.com/cn/?v=javlib-123",
		}},
	}
	javDB := &fakeAVCrawlerForExecution{
		name: "javdb",
		searchResult: []avScrapeCandidate{{
			Source:     "javdb",
			ExternalID: "javdb-123",
			Code:       "SSIS-123",
			Title:      "JavDB Result",
			DetailURL:  "https://javdb.com/v/javdb-123",
		}},
	}

	svc := NewScraperService(nil, "", "", "", "", 0)
	svc.avProvider = &fakeAVCrawlerProviderForExecution{crawlers: []avCrawler{javLibrary, javDB}}

	got, err := svc.PreviewAVSearch(context.Background(), "SSIS-123", AVPreviewOptions{
		BypassCache: true,
		SiteSource:  "javlibrary",
	})
	if err != nil {
		t.Fatalf("PreviewAVSearch returned error: %v", err)
	}

	if javLibrary.searchCalls == 0 {
		t.Fatalf("expected explicit crawler to be called")
	}
	if javDB.searchCalls != 0 {
		t.Fatalf("expected javdb crawler to stay unused, calls=%d", javDB.searchCalls)
	}
	if got.UsedSource != "javlibrary" {
		t.Fatalf("expected used_source javlibrary, got=%s", got.UsedSource)
	}
	if len(got.Candidates) == 0 {
		t.Fatalf("expected av candidates, got none")
	}
	if got.Candidates[0]["scrape_source"] != "javlibrary" {
		t.Fatalf("expected first candidate source javlibrary, got=%v", got.Candidates[0]["scrape_source"])
	}
}

func TestPreviewAVSearchHonorsConfiguredEnabledSourcesOnly(t *testing.T) {
	t.Parallel()

	javLibrary := &fakeAVCrawlerForExecution{
		name: "javlibrary",
		searchResult: []avScrapeCandidate{{
			Source:     "javlibrary",
			ExternalID: "javlib-456",
			Code:       "SSIS-456",
			Title:      "Configured JavLibrary Result",
			DetailURL:  "https://www.javlibrary.com/cn/?v=javlib-456",
		}},
	}
	javDB := &fakeAVCrawlerForExecution{name: "javdb"}

	svc := NewScraperService(nil, "", "", "", "", 0)
	svc.avConfigStore = fakeAVConfigStore{
		config: AVScraperSiteConfig{
			EnabledSites: []string{"javlibrary"},
			CategorySiteOrder: map[string][]string{
				avSiteCategoryJapanese: {"javlibrary"},
			},
		},
	}
	svc.avProvider = &fakeAVCrawlerProviderForExecution{crawlers: []avCrawler{javLibrary, javDB}}

	got, err := svc.PreviewAVSearch(context.Background(), "SSIS-456", AVPreviewOptions{
		BypassCache: true,
	})
	if err != nil {
		t.Fatalf("PreviewAVSearch returned error: %v", err)
	}

	if javLibrary.searchCalls == 0 {
		t.Fatalf("expected configured crawler to be called")
	}
	if javDB.searchCalls != 0 {
		t.Fatalf("expected disabled crawler to stay unused, calls=%d", javDB.searchCalls)
	}
	if got.UsedSource != "javlibrary" {
		t.Fatalf("expected used_source javlibrary, got=%s", got.UsedSource)
	}
}

func TestPreviewAVSearchCacheKeyIncludesSiteSource(t *testing.T) {
	t.Parallel()

	javLibrary := &fakeAVCrawlerForExecution{
		name: "javlibrary",
		searchResult: []avScrapeCandidate{{
			Source:     "javlibrary",
			ExternalID: "javlib-789",
			Code:       "SSIS-789",
			Title:      "JavLibrary Cache Result",
			DetailURL:  "https://www.javlibrary.com/cn/?v=javlib-789",
		}},
	}
	javDB := &fakeAVCrawlerForExecution{
		name: "javdb",
		searchResult: []avScrapeCandidate{{
			Source:     "javdb",
			ExternalID: "javdb-789",
			Code:       "SSIS-789",
			Title:      "JavDB Cache Result",
			DetailURL:  "https://javdb.com/v/javdb-789",
		}},
	}

	svc := NewScraperService(nil, "", "", "", "", 0)
	svc.avProvider = &fakeAVCrawlerProviderForExecution{crawlers: []avCrawler{javLibrary, javDB}}

	first, err := svc.PreviewAVSearch(context.Background(), "SSIS-789", AVPreviewOptions{
		BypassCache: false,
		SiteSource:  "javdb",
	})
	if err != nil {
		t.Fatalf("first PreviewAVSearch returned error: %v", err)
	}
	second, err := svc.PreviewAVSearch(context.Background(), "SSIS-789", AVPreviewOptions{
		BypassCache: false,
		SiteSource:  "javlibrary",
	})
	if err != nil {
		t.Fatalf("second PreviewAVSearch returned error: %v", err)
	}

	if len(first.Candidates) == 0 || first.Candidates[0]["scrape_source"] != "javdb" {
		t.Fatalf("expected first result from javdb, got=%v", first.Candidates)
	}
	if len(second.Candidates) == 0 || second.Candidates[0]["scrape_source"] != "javlibrary" {
		t.Fatalf("expected second result from javlibrary, got=%v", second.Candidates)
	}
}

func TestPreviewAVSearchRejectsUnknownExplicitSiteSource(t *testing.T) {
	t.Parallel()

	svc := NewScraperService(nil, "", "", "", "", 0)
	svc.avProvider = &fakeAVCrawlerProviderForExecution{crawlers: []avCrawler{
		&fakeAVCrawlerForExecution{name: "javdb"},
	}}

	_, err := svc.PreviewAVSearch(context.Background(), "SSIS-999", AVPreviewOptions{
		BypassCache: true,
		SiteSource:  "unknown-site",
	})
	if err == nil {
		t.Fatalf("expected error for unknown explicit site source")
	}
}

func TestPreviewAVSearchDoesNotFallBackToDisabledCategorySource(t *testing.T) {
	t.Parallel()

	javDB := &fakeAVCrawlerForExecution{
		name: "javdb",
		searchResult: []avScrapeCandidate{{
			Source:     "javdb",
			ExternalID: "javdb-disabled",
			Code:       "SSIS-001",
			Title:      "Disabled JavDB Result",
			DetailURL:  "https://javdb.com/v/javdb-disabled",
		}},
	}
	dmm := &fakeAVCrawlerForExecution{
		name: "dmm",
		searchResult: []avScrapeCandidate{{
			Source:     "dmm",
			ExternalID: "dmm-enabled",
			Code:       "SSIS-001",
			Title:      "Enabled DMM Result",
			DetailURL:  "https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=dmm-enabled/",
		}},
	}

	svc := NewScraperService(nil, "", "", "", "", 0)
	svc.avConfigStore = fakeAVConfigStore{
		config: AVScraperSiteConfig{
			EnabledSites: []string{"dmm"},
			CategorySiteOrder: map[string][]string{
				avSiteCategoryJapanese: {"javdb"},
			},
		},
	}
	svc.avProvider = &fakeAVCrawlerProviderForExecution{crawlers: []avCrawler{javDB, dmm}}

	got, err := svc.PreviewAVSearch(context.Background(), "SSIS-001", AVPreviewOptions{
		BypassCache: true,
	})
	if err != nil {
		t.Fatalf("PreviewAVSearch returned error: %v", err)
	}
	if javDB.searchCalls != 0 {
		t.Fatalf("expected disabled category source to stay unused, calls=%d", javDB.searchCalls)
	}
	if dmm.searchCalls == 0 {
		t.Fatal("expected enabled source dmm to be used")
	}
	if got.UsedSource != "dmm" {
		t.Fatalf("expected used_source dmm, got=%s", got.UsedSource)
	}
}

func TestScrapeAVUploadPassesFilePathToThePornDBCrawler(t *testing.T) {
	t.Parallel()

	filePath := "/videos/x-art.19.11.03.A.Kitten.For.Christmas.mp4"
	videoID := uuid.New()
	repo := &fakeScraperRepo{
		videoByID: map[uuid.UUID]models.Video{
			videoID: {
				ID:           videoID,
				Title:        "XART-123",
				OriginalPath: filePath,
			},
		},
	}
	thePornDB := &fakeAVCrawlerForExecution{
		name: "theporndb",
		searchResult: []avScrapeCandidate{{
			Source:     "theporndb",
			ExternalID: "x-art-kitten-for-christmas",
			Code:       "XART-123",
			Title:      "ThePornDB Result",
			DetailURL:  "https://api.theporndb.net/scenes/x-art-kitten-for-christmas",
		}},
	}

	svc := NewScraperService(repo, "", "", t.TempDir(), "", 0)
	svc.avProvider = &fakeAVCrawlerProviderForExecution{crawlers: []avCrawler{thePornDB}}

	_, err := svc.ScrapeAVUpload(context.Background(), videoID, filePath, "")
	if err != nil {
		t.Fatalf("ScrapeAVUpload returned error: %v", err)
	}
	if thePornDB.searchCalls == 0 {
		t.Fatal("expected ThePornDB crawler to be used")
	}
	if len(thePornDB.searchQueries) == 0 || thePornDB.searchQueries[0] != filePath {
		t.Fatalf("expected first ThePornDB query to use raw file path, got=%v", thePornDB.searchQueries)
	}
}
