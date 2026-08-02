package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"video-server/internal/repository"
)

type fakeBackfillRepository struct {
	items   []repository.TranscodedVideoFile
	updates map[uuid.UUID]int64
}

func (f *fakeBackfillRepository) ListTranscodedVideoFiles(_ context.Context, afterID uuid.UUID, limit int) ([]repository.TranscodedVideoFile, error) {
	items := make([]repository.TranscodedVideoFile, 0, limit)
	for _, item := range f.items {
		if item.ID.String() <= afterID.String() {
			continue
		}
		items = append(items, item)
		if len(items) == limit {
			break
		}
	}
	return items, nil
}

func (f *fakeBackfillRepository) UpdateVideoTranscodedFileSize(_ context.Context, videoID uuid.UUID, size int64) error {
	if f.updates == nil {
		f.updates = map[uuid.UUID]int64{}
	}
	f.updates[videoID] = size
	return nil
}

func TestRunBackfillDryRunReportsMissingFileWithoutWriting(t *testing.T) {
	t.Parallel()

	storageRoot := t.TempDir()
	videoOneID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	videoTwoID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	videoOnePath := filepath.Join(storageRoot, "videos", videoOneID.String(), "video.mp4")
	if err := os.MkdirAll(filepath.Dir(videoOnePath), 0o755); err != nil {
		t.Fatalf("create video directory: %v", err)
	}
	if err := os.WriteFile(videoOnePath, []byte("playable-video"), 0o600); err != nil {
		t.Fatalf("write video: %v", err)
	}

	repo := &fakeBackfillRepository{items: []repository.TranscodedVideoFile{
		{ID: videoOneID, TranscodedPath: videoOnePath},
		{ID: videoTwoID, TranscodedPath: filepath.Join(storageRoot, "videos", videoTwoID.String(), "video.mp4")},
	}}

	report, err := runBackfill(context.Background(), repo, backfillOptions{
		StorageRoot: storageRoot,
		BatchSize:   1,
		Apply:       false,
	})
	if err == nil {
		t.Fatal("runBackfill() expected incomplete result error")
	}
	if report.Scanned != 2 || report.WouldUpdate != 1 || report.Updated != 0 || len(report.Failures) != 1 {
		t.Fatalf("unexpected report: %+v", report)
	}
	if report.StartedAt.IsZero() || report.FinishedAt.IsZero() {
		t.Fatalf("report timestamps must be populated: %+v", report)
	}
	if len(repo.updates) != 0 {
		t.Fatalf("dry run wrote updates: %+v", repo.updates)
	}
}

func TestRunBackfillApplyWritesPrimaryPlaybackSize(t *testing.T) {
	t.Parallel()

	storageRoot := t.TempDir()
	videoID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	videoPath := filepath.Join(storageRoot, "videos", videoID.String(), "video.mp4")
	if err := os.MkdirAll(filepath.Dir(videoPath), 0o755); err != nil {
		t.Fatalf("create video directory: %v", err)
	}
	if err := os.WriteFile(videoPath, []byte("video"), 0o600); err != nil {
		t.Fatalf("write video: %v", err)
	}

	repo := &fakeBackfillRepository{items: []repository.TranscodedVideoFile{{
		ID: videoID, TranscodedPath: videoPath,
	}}}
	report, err := runBackfill(context.Background(), repo, backfillOptions{
		StorageRoot: storageRoot,
		BatchSize:   100,
		Apply:       true,
	})
	if err != nil {
		t.Fatalf("runBackfill() error = %v", err)
	}
	if report.Scanned != 1 || report.Updated != 1 || report.WouldUpdate != 0 || len(report.Failures) != 0 {
		t.Fatalf("unexpected report: %+v", report)
	}
	if repo.updates[videoID] != int64(len("video")) {
		t.Fatalf("stored size = %d, want %d", repo.updates[videoID], len("video"))
	}
}

func TestResolveBackfillFileSizeRejectsPathOutsideVideoStorage(t *testing.T) {
	t.Parallel()

	storageRoot := t.TempDir()
	outsidePath := filepath.Join(t.TempDir(), "outside.mp4")
	if err := os.WriteFile(outsidePath, []byte("video"), 0o600); err != nil {
		t.Fatalf("write outside video: %v", err)
	}

	if _, err := resolveBackfillFileSize(storageRoot, outsidePath); err == nil {
		t.Fatal("resolveBackfillFileSize() expected outside-root error")
	}
}
