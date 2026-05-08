package services

import (
	"context"
	"testing"
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
}

func (c *fakeAVCrawlerForExecution) Name() string {
	return c.name
}

func (c *fakeAVCrawlerForExecution) SearchCandidates(context.Context, *avScrapeRunContext, string, int) ([]avScrapeCandidate, error) {
	c.searchCalls++
	if c.searchErr != nil {
		return nil, c.searchErr
	}
	out := make([]avScrapeCandidate, len(c.searchResult))
	copy(out, c.searchResult)
	return out, nil
}

func (c *fakeAVCrawlerForExecution) FetchByDetailURL(context.Context, *avScrapeRunContext, string) (avScrapeCandidate, error) {
	return avScrapeCandidate{}, nil
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
