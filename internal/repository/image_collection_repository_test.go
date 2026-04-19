package repository

import (
	"testing"

	"github.com/google/uuid"
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
