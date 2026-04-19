package handlers

import (
	"testing"

	"github.com/google/uuid"
)

func TestParseUUIDCSV(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()

	got, err := parseUUIDCSV(id1.String() + "," + id2.String() + "," + id1.String())
	if err != nil {
		t.Fatalf("parseUUIDCSV returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 unique ids, got %d", len(got))
	}
	if got[0] != id1 || got[1] != id2 {
		t.Fatalf("unexpected parse result: %#v", got)
	}
}

func TestParseUUIDCSVRejectsInvalidID(t *testing.T) {
	if _, err := parseUUIDCSV("not-a-uuid"); err == nil {
		t.Fatal("expected invalid uuid error")
	}
}
