package services

import (
	"context"
	"testing"
)

type fakeAVConfigStore struct {
	config AVScraperSiteConfig
	err    error
}

func (f fakeAVConfigStore) GetAVScraperConfig(context.Context) (AVScraperSiteConfig, error) {
	return f.config, f.err
}

func TestResolveAVSearchPlanUsesDefaultFC2Sites(t *testing.T) {
	t.Parallel()

	svc := NewScraperService(nil, "", "", "", "", 0)

	plan, err := svc.resolveAVSearchPlan(context.Background(), "FC2-PPV-123456", AVPreviewOptions{})
	if err != nil {
		t.Fatalf("resolveAVSearchPlan returned error: %v", err)
	}

	if plan.SiteCategory != avSiteCategoryFC2 {
		t.Fatalf("expected fc2 category, got=%s", plan.SiteCategory)
	}
	if plan.RecommendedSource != "fc2ppvdb" {
		t.Fatalf("expected recommended source fc2ppvdb, got=%s", plan.RecommendedSource)
	}
	if len(plan.Sources) < 4 {
		t.Fatalf("expected fc2 fallback sources, got=%v", plan.Sources)
	}
	if plan.Sources[0] != "fc2ppvdb" || plan.Sources[1] != "fc2club" {
		t.Fatalf("unexpected fc2 source order: %v", plan.Sources)
	}
}

func TestDefaultAVScraperSiteConfigIncludesMDCxMigratedSites(t *testing.T) {
	t.Parallel()

	cfg := defaultAVScraperSiteConfig()
	for _, site := range []string{"theporndb", "dmm", "javdb", "jav321", "mgstage", "fc2ppvdb", "fc2club", "fc2", "fc2hub"} {
		if !stringSliceContains(cfg.EnabledSites, site) {
			t.Fatalf("expected default enabled sites to include %s, got=%v", site, cfg.EnabledSites)
		}
	}

	wantJapanesePrefix := []string{"theporndb", "dmm", "javdb", "jav321", "mgstage"}
	if got := cfg.CategorySiteOrder[avSiteCategoryJapanese]; len(got) < len(wantJapanesePrefix) {
		t.Fatalf("expected japanese order to include MDCx prefix %v, got=%v", wantJapanesePrefix, got)
	} else {
		for i, want := range wantJapanesePrefix {
			if got[i] != want {
				t.Fatalf("expected japanese source %d to be %s, got order=%v", i, want, got)
			}
		}
	}

	wantFC2Prefix := []string{"fc2ppvdb", "fc2club", "fc2", "fc2hub"}
	if got := cfg.CategorySiteOrder[avSiteCategoryFC2]; len(got) < len(wantFC2Prefix) {
		t.Fatalf("expected fc2 order to include MDCx prefix %v, got=%v", wantFC2Prefix, got)
	} else {
		for i, want := range wantFC2Prefix {
			if got[i] != want {
				t.Fatalf("expected fc2 source %d to be %s, got order=%v", i, want, got)
			}
		}
	}
}

func TestResolveAVSearchPlanUsesStoredWesternConfigOrder(t *testing.T) {
	t.Parallel()

	svc := NewScraperService(nil, "", "", "", "", 0)
	svc.avConfigStore = fakeAVConfigStore{
		config: AVScraperSiteConfig{
			EnabledSites: []string{"javdb", "theporndb", "javlibrary"},
			CategorySiteOrder: map[string][]string{
				avSiteCategoryWestern: {"theporndb", "javdb"},
			},
		},
	}

	plan, err := svc.resolveAVSearchPlan(context.Background(), "Brazzers office affair", AVPreviewOptions{})
	if err != nil {
		t.Fatalf("resolveAVSearchPlan returned error: %v", err)
	}

	if plan.SiteCategory != avSiteCategoryWestern {
		t.Fatalf("expected western category, got=%s", plan.SiteCategory)
	}
	if plan.RecommendedSource != "theporndb" {
		t.Fatalf("expected recommended source theporndb, got=%s", plan.RecommendedSource)
	}
	if len(plan.Sources) != 2 || plan.Sources[0] != "theporndb" || plan.Sources[1] != "javdb" {
		t.Fatalf("unexpected stored config order: %v", plan.Sources)
	}
}

func TestResolveAVSearchPlanAllowsExplicitSiteSourceOverride(t *testing.T) {
	t.Parallel()

	svc := NewScraperService(nil, "", "", "", "", 0)
	svc.avConfigStore = fakeAVConfigStore{
		config: AVScraperSiteConfig{
			EnabledSites: []string{"javdb", "javlibrary"},
		},
	}

	plan, err := svc.resolveAVSearchPlan(context.Background(), "SSIS-123", AVPreviewOptions{
		SiteCategory: avSiteCategoryJapanese,
		SiteSource:   "javlibrary",
	})
	if err != nil {
		t.Fatalf("resolveAVSearchPlan returned error: %v", err)
	}

	if plan.SiteCategory != avSiteCategoryJapanese {
		t.Fatalf("expected japanese category, got=%s", plan.SiteCategory)
	}
	if plan.RecommendedSource != "javlibrary" {
		t.Fatalf("expected explicit recommended source javlibrary, got=%s", plan.RecommendedSource)
	}
	if len(plan.Sources) != 1 || plan.Sources[0] != "javlibrary" {
		t.Fatalf("expected explicit single source list, got=%v", plan.Sources)
	}
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
