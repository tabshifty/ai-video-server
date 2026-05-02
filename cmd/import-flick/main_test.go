package main

import (
	"reflect"
	"testing"
	"time"
)

func TestBuildMongoFilterIncludesPlayableTagAndSince(t *testing.T) {
	t.Parallel()

	since := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)
	filter := buildMongoFilter("Dance", since)

	if got := filter["canplay"]; got != true {
		t.Fatalf("expected canplay=true, got %#v", got)
	}
	tagFilter, ok := filter["tags"].(map[string]any)
	if !ok {
		t.Fatalf("expected tags filter map[string]any, got %T", filter["tags"])
	}
	if tagFilter["$options"] != "i" {
		t.Fatalf("expected case-insensitive tag filter, got %#v", tagFilter)
	}
	createdFilter, ok := filter["createdAt"].(map[string]any)
	if !ok {
		t.Fatalf("expected createdAt filter map[string]any, got %T", filter["createdAt"])
	}
	if !reflect.DeepEqual(createdFilter["$gte"], since) {
		t.Fatalf("expected createdAt >= since, got %#v", createdFilter)
	}
}

func TestParseSinceFlagSupportsDateAndRFC3339(t *testing.T) {
	t.Parallel()

	dateOnly, err := parseSinceFlag("2024-05-06")
	if err != nil {
		t.Fatalf("parse date-only since flag: %v", err)
	}
	if want := time.Date(2024, 5, 6, 0, 0, 0, 0, time.UTC); !dateOnly.Equal(want) {
		t.Fatalf("unexpected date-only value: got=%s want=%s", dateOnly, want)
	}

	withTime, err := parseSinceFlag("2024-05-06T07:08:09Z")
	if err != nil {
		t.Fatalf("parse rfc3339 since flag: %v", err)
	}
	if want := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC); !withTime.Equal(want) {
		t.Fatalf("unexpected rfc3339 value: got=%s want=%s", withTime, want)
	}
}
