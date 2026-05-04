package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"video-server/internal/services"
)

func TestBuildMongoFilterIncludesPlayableTagAndSince(t *testing.T) {
	t.Parallel()

	since := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)
	filter := buildMongoFilter("Dance", since, "")

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

func TestBuildMongoFilterSupportsResumeCheckpoint(t *testing.T) {
	t.Parallel()

	since := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)
	filter := buildMongoFilter("Dance", since, "66112233445566778899aabb")
	if _, ok := filter["createdAt"]; ok {
		t.Fatalf("createdAt direct filter should not exist when resume checkpoint is set")
	}
	orFilter, ok := filter["$or"].([]any)
	if !ok || len(orFilter) != 2 {
		t.Fatalf("expected $or resume filter, got %#v", filter["$or"])
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

func TestRunImportBatchCollectsStats(t *testing.T) {
	t.Parallel()

	cfg := importConfig{
		SourceVideoDir: "/source/video",
		SourceCoverDir: "/source/cover",
		StorageRoot:    "/storage",
		DryRun:         true,
	}
	docs := []mongoExportVideo{
		{ID: mongoObjectID("a"), MD5: "a", CanPlay: true, Tags: []string{"x"}},
		{ID: mongoObjectID("b"), MD5: "b", CanPlay: true, Tags: []string{"x"}},
		{ID: mongoObjectID("c"), MD5: "c", CanPlay: true, Tags: []string{"x"}},
		{ID: mongoObjectID("d"), MD5: "d", CanPlay: true, Tags: []string{"x"}},
	}

	importer := &fakeImporter{
		outcomes: []services.FlickImportOutcome{
			{Status: services.FlickImportStatusDryRun},
			{Status: services.FlickImportStatusSkipped, Reason: services.FlickImportReasonDuplicateHash},
			{Status: services.FlickImportStatusSkipped, Reason: services.FlickImportReasonMissingVideo},
		},
		errAt: map[int]error{
			3: errors.New("boom"),
		},
	}

	stats := runImportBatch(context.Background(), importer, cfg, docs)
	if stats.processed != 4 {
		t.Fatalf("expected processed=4, got=%d", stats.processed)
	}
	if stats.wouldImport != 1 {
		t.Fatalf("expected wouldImport=1, got=%d", stats.wouldImport)
	}
	if stats.skipped != 2 {
		t.Fatalf("expected skipped=2, got=%d", stats.skipped)
	}
	if stats.failed != 1 {
		t.Fatalf("expected failed=1, got=%d", stats.failed)
	}
	if stats.skipReasons[services.FlickImportReasonDuplicateHash] != 1 {
		t.Fatalf("expected duplicate_hash=1, got=%d", stats.skipReasons[services.FlickImportReasonDuplicateHash])
	}
	if stats.skipReasons[services.FlickImportReasonMissingVideo] != 1 {
		t.Fatalf("expected missing_video=1, got=%d", stats.skipReasons[services.FlickImportReasonMissingVideo])
	}
	if len(stats.failures) != 1 || stats.failures[0].SourceID != "d" {
		t.Fatalf("expected one failure for source d, got=%#v", stats.failures)
	}
}

func TestCheckpointReadWriteRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "checkpoint.json")
	in := importCheckpoint{
		CreatedAt: time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC),
		ObjectID:  "66112233445566778899aabb",
	}
	if err := writeCheckpoint(path, in); err != nil {
		t.Fatalf("write checkpoint failed: %v", err)
	}
	out, err := readCheckpoint(path)
	if err != nil {
		t.Fatalf("read checkpoint failed: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil checkpoint")
	}
	if !out.CreatedAt.Equal(in.CreatedAt) || out.ObjectID != in.ObjectID {
		t.Fatalf("checkpoint mismatch: got=%+v want=%+v", out, in)
	}
}

func TestReadCheckpointMissingFileReturnsNil(t *testing.T) {
	t.Parallel()
	out, err := readCheckpoint(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("read missing checkpoint should not fail: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil checkpoint for missing file, got %+v", out)
	}
}

func TestWriteCheckpointSkipsInvalidPayload(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "checkpoint.json")
	err := writeCheckpoint(path, importCheckpoint{})
	if err != nil {
		t.Fatalf("write invalid checkpoint should not fail: %v", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("expected no checkpoint file, stat err=%v", statErr)
	}
}

type fakeImporter struct {
	outcomes []services.FlickImportOutcome
	errAt    map[int]error
	index    int
}

func (f *fakeImporter) ImportPlayableVideo(_ context.Context, _ services.FlickSourceVideo, _ services.FlickImportOptions) (services.FlickImportOutcome, error) {
	err := f.errAt[f.index]
	if err != nil {
		f.index++
		return services.FlickImportOutcome{}, err
	}
	if f.index >= len(f.outcomes) {
		f.index++
		return services.FlickImportOutcome{}, nil
	}
	out := f.outcomes[f.index]
	f.index++
	return out, nil
}
