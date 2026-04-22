package repository

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
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

func TestScanAppVideoImageCollectionPrefersDerivedPreviewURL(t *testing.T) {
	collectionID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	imageID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	now := time.Unix(1710000000, 0).UTC()

	got, err := scanAppVideoImageCollection(stubRowScanner{
		values: []any{
			collectionID,
			"剧照合集",
			"https://legacy.example/cover.jpg",
			imageID.String(),
			now,
			now,
		},
	})
	if err != nil {
		t.Fatalf("scanAppVideoImageCollection returned error: %v", err)
	}

	want := models.VideoImageCollection{
		ID:       collectionID,
		Name:     "剧照合集",
		CoverURL: utils.AppImageViewURL(imageID, appImageCollectionCoverWidth, appImageCollectionCoverHeight, "cover", appImageCollectionCoverQuality),
	}
	if !reflect.DeepEqual(*got, want) {
		t.Fatalf("unexpected video image collection: got=%#v want=%#v", *got, want)
	}
}

func TestScanAppVideoImageCollectionFallsBackToStoredURL(t *testing.T) {
	collectionID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	now := time.Unix(1710000000, 0).UTC()

	got, err := scanAppVideoImageCollection(stubRowScanner{
		values: []any{
			collectionID,
			"剧照合集",
			" https://legacy.example/cover.jpg ",
			"",
			now,
			now,
		},
	})
	if err != nil {
		t.Fatalf("scanAppVideoImageCollection returned error: %v", err)
	}
	if got == nil {
		t.Fatal("expected video image collection")
	}
	if got.CoverURL != "https://legacy.example/cover.jpg" {
		t.Fatalf("unexpected fallback cover url: %s", got.CoverURL)
	}
}

type stubRowScanner struct {
	values []any
}

func (s stubRowScanner) Scan(dest ...any) error {
	for idx := range dest {
		reflect.ValueOf(dest[idx]).Elem().Set(reflect.ValueOf(s.values[idx]))
	}
	return nil
}
