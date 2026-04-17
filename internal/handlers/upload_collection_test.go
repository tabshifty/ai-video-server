package handlers

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestParseUploadCollectionIDs(t *testing.T) {
	t.Parallel()

	id1 := uuid.New()
	id2 := uuid.New()
	raw := `["` + id1.String() + `","` + id2.String() + `","` + id1.String() + `"]`

	got, err := parseUploadCollectionIDs(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []uuid.UUID{id1, id2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected ids: got=%v want=%v", got, want)
	}
}

func TestParseUploadCollectionIDsInvalid(t *testing.T) {
	t.Parallel()

	_, err := parseUploadCollectionIDs("invalid-id")
	if !errors.Is(err, errInvalidCollectionID) {
		t.Fatalf("expected errInvalidCollectionID, got=%v", err)
	}
}

func TestCollectionTypeValidation(t *testing.T) {
	t.Parallel()

	if err := validateCollectionsForType("short", []uuid.UUID{uuid.New()}); err != nil {
		t.Fatalf("short should allow collections, got=%v", err)
	}

	if err := validateCollectionsForType("movie", []uuid.UUID{uuid.New()}); !errors.Is(err, errCollectionsOnlyForShort) {
		t.Fatalf("movie should reject collections, got=%v", err)
	}

	if err := validateCollectionsForType("episode", []uuid.UUID{uuid.New()}); !errors.Is(err, errCollectionsOnlyForShort) {
		t.Fatalf("episode should reject collections, got=%v", err)
	}
	if err := validateCollectionsForType("av", []uuid.UUID{uuid.New()}); !errors.Is(err, errCollectionsOnlyForShort) {
		t.Fatalf("av should reject collections, got=%v", err)
	}

	if err := validateCollectionsForType("movie", nil); err != nil {
		t.Fatalf("movie with empty collections should pass, got=%v", err)
	}
}
