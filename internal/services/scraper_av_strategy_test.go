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
	if plan.RecommendedSource != "fc2" {
		t.Fatalf("expected recommended source fc2, got=%s", plan.RecommendedSource)
	}
	if len(plan.Sources) < 4 {
		t.Fatalf("expected fc2 fallback sources, got=%v", plan.Sources)
	}
	if plan.Sources[0] != "fc2" || plan.Sources[1] != "fc2club" {
		t.Fatalf("unexpected fc2 source order: %v", plan.Sources)
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
