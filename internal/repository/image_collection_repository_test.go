package repository

import (
	"testing"

	"github.com/google/uuid"

	"video-server/internal/utils"
)

func TestNormalizeSingleImageCollectionID(t *testing.T) {
	id := uuid.New()

	got, err := normalizeSingleImageCollectionID([]uuid.UUID{id, id})
	if err != nil {
		t.Fatalf("normalizeSingleImageCollectionID returned error: %v", err)
	}
	if got == nil {
		t.Fatal("expected single image collection id")
	}
	if *got != id {
		t.Fatalf("expected %s, got %s", id, *got)
	}
}

func TestNormalizeSingleImageCollectionIDRejectsMultipleDistinctIDs(t *testing.T) {
	_, err := normalizeSingleImageCollectionID([]uuid.UUID{uuid.New(), uuid.New()})
	if err == nil {
		t.Fatal("expected error for multiple distinct image collection ids")
	}
}

func TestResolveAdminImageCollectionCoverURLPrefersDerivedPreviewURL(t *testing.T) {
	imageID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")

	got := resolveAdminImageCollectionCoverURL(&imageID, "https://legacy.example/cover.jpg")
	want := utils.AdminImageViewURL(imageID, adminImageCollectionCoverWidth, adminImageCollectionCoverHeight, "cover", adminImageCollectionCoverQuality)
	if got != want {
		t.Fatalf("expected derived preview url, got=%s want=%s", got, want)
	}
}

func TestResolveAdminImageCollectionCoverURLFallsBackToStoredURL(t *testing.T) {
	got := resolveAdminImageCollectionCoverURL(nil, " https://legacy.example/cover.jpg ")
	want := "https://legacy.example/cover.jpg"
	if got != want {
		t.Fatalf("expected fallback cover url, got=%s want=%s", got, want)
	}
}

func TestResolveAppImageCollectionCoverURLPrefersDerivedPreviewURL(t *testing.T) {
	imageID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")

	got := resolveAppImageCollectionCoverURL(&imageID, "https://legacy.example/cover.jpg")
	want := utils.AppImageViewURL(imageID, appImageCollectionCoverWidth, appImageCollectionCoverHeight, "cover", appImageCollectionCoverQuality)
	if got != want {
		t.Fatalf("expected derived preview url, got=%s want=%s", got, want)
	}
}

func TestResolveAppImageCollectionCoverURLFallsBackToStoredURL(t *testing.T) {
	got := resolveAppImageCollectionCoverURL(nil, " https://legacy.example/cover.jpg ")
	want := "https://legacy.example/cover.jpg"
	if got != want {
		t.Fatalf("expected fallback cover url, got=%s want=%s", got, want)
	}
}
