package handlers

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestCollectionTypeValidationAV(t *testing.T) {
	t.Parallel()

	if err := validateCollectionsForType("av", []uuid.UUID{uuid.New()}); !errors.Is(err, errCollectionsOnlyForShort) {
		t.Fatalf("av should reject collections, got=%v", err)
	}
}
