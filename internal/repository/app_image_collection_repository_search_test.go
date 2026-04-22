package repository

import "testing"

func TestBuildAppImageCollectionTitleSearchPattern(t *testing.T) {
	got := buildAppImageCollectionTitleSearchPattern("  夏日写真  ")
	if got == nil {
		t.Fatal("expected search pattern")
	}
	if *got != "%夏日写真%" {
		t.Fatalf("unexpected search pattern: %s", *got)
	}
}

func TestBuildAppImageCollectionTitleSearchPatternReturnsNilForBlank(t *testing.T) {
	if got := buildAppImageCollectionTitleSearchPattern("   "); got != nil {
		t.Fatalf("expected nil for blank keyword, got %v", *got)
	}
}
