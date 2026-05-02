package services

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/repository"
	"video-server/pkg/ffmpeg"
)

type fakeFlickImportRepo struct {
	findVideoID     uuid.UUID
	findExists      bool
	findErr         error
	created         []repository.ImportedReadyVideo
	createErr       error
	findHashCalls   int
	createVideoCall int
}

func (f *fakeFlickImportRepo) FindVideoByHash(_ context.Context, _ string, _ int64) (uuid.UUID, bool, error) {
	f.findHashCalls++
	return f.findVideoID, f.findExists, f.findErr
}

func (f *fakeFlickImportRepo) CreateImportedReadyVideo(_ context.Context, spec repository.ImportedReadyVideo) error {
	f.createVideoCall++
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, spec)
	return nil
}

func TestFlickImportServiceImportPlayableVideoSkipsWhenTagsEmpty(t *testing.T) {
	t.Parallel()

	svc := NewFlickImportService(&fakeFlickImportRepo{})
	outcome, err := svc.ImportPlayableVideo(context.Background(), FlickSourceVideo{
		SourceID:  "source-1",
		MD5:       "abc123",
		CanPlay:   true,
		Tags:      []string{" ", ""},
		CreatedAt: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
	}, FlickImportOptions{})
	if err != nil {
		t.Fatalf("ImportPlayableVideo returned error: %v", err)
	}
	if outcome.Status != FlickImportStatusSkipped {
		t.Fatalf("expected skipped outcome, got %s", outcome.Status)
	}
	if outcome.Reason != FlickImportReasonNoTags {
		t.Fatalf("expected no-tags reason, got %s", outcome.Reason)
	}
}

func TestFlickImportServiceImportPlayableVideoSkipsDuplicateHash(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	sourceVideo := filepath.Join(tmpDir, "source", "video")
	sourceCover := filepath.Join(tmpDir, "source", "cover")
	storageRoot := filepath.Join(tmpDir, "storage")
	if err := os.MkdirAll(sourceVideo, 0o755); err != nil {
		t.Fatalf("mkdir source video dir: %v", err)
	}
	if err := os.MkdirAll(sourceCover, 0o755); err != nil {
		t.Fatalf("mkdir source cover dir: %v", err)
	}
	videoPath := filepath.Join(sourceVideo, "abc123.mp4")
	if err := os.WriteFile(videoPath, []byte("video"), 0o644); err != nil {
		t.Fatalf("write source video: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceCover, "abc123.jpg"), []byte("cover"), 0o644); err != nil {
		t.Fatalf("write source cover: %v", err)
	}

	existingID := uuid.New()
	repo := &fakeFlickImportRepo{
		findVideoID: existingID,
		findExists:  true,
	}
	svc := NewFlickImportService(repo)
	svc.probe = func(context.Context, string) (ffmpeg.VideoProbe, error) {
		return ffmpeg.VideoProbe{Duration: 12.4, Width: 720, Height: 1280}, nil
	}
	svc.hashFile = func(string) (string, error) { return "hash-1", nil }

	outcome, err := svc.ImportPlayableVideo(context.Background(), FlickSourceVideo{
		SourceID:  "source-2",
		MD5:       "abc123",
		CanPlay:   true,
		Tags:      []string{"Dance"},
		CreatedAt: time.Date(2024, 2, 3, 4, 5, 6, 0, time.UTC),
	}, FlickImportOptions{
		SourceVideoDir: sourceVideo,
		SourceCoverDir: sourceCover,
		StorageRoot:    storageRoot,
	})
	if err != nil {
		t.Fatalf("ImportPlayableVideo returned error: %v", err)
	}
	if outcome.Status != FlickImportStatusSkipped {
		t.Fatalf("expected skipped outcome, got %s", outcome.Status)
	}
	if outcome.Reason != FlickImportReasonDuplicateHash {
		t.Fatalf("expected duplicate-hash reason, got %s", outcome.Reason)
	}
	if outcome.VideoID != existingID {
		t.Fatalf("expected existing video id %s, got %s", existingID, outcome.VideoID)
	}
	if repo.createVideoCall != 0 {
		t.Fatalf("expected create not called, got %d", repo.createVideoCall)
	}
}

func TestFlickImportServiceImportPlayableVideoCreatesReadyVideo(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	sourceVideo := filepath.Join(tmpDir, "source", "video")
	sourceCover := filepath.Join(tmpDir, "source", "cover")
	storageRoot := filepath.Join(tmpDir, "storage")
	if err := os.MkdirAll(sourceVideo, 0o755); err != nil {
		t.Fatalf("mkdir source video dir: %v", err)
	}
	if err := os.MkdirAll(sourceCover, 0o755); err != nil {
		t.Fatalf("mkdir source cover dir: %v", err)
	}
	videoPath := filepath.Join(sourceVideo, "abc123.mp4")
	if err := os.WriteFile(videoPath, []byte("video-data"), 0o644); err != nil {
		t.Fatalf("write source video: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceCover, "abc123.jpg"), []byte("cover-data"), 0o644); err != nil {
		t.Fatalf("write source cover: %v", err)
	}

	repo := &fakeFlickImportRepo{}
	svc := NewFlickImportService(repo)
	svc.probe = func(context.Context, string) (ffmpeg.VideoProbe, error) {
		return ffmpeg.VideoProbe{Duration: 18.6, Width: 1080, Height: 1920}, nil
	}
	svc.hashFile = func(string) (string, error) { return "hash-2", nil }

	createdAt := time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC)
	updatedAt := createdAt.Add(2 * time.Hour)
	outcome, err := svc.ImportPlayableVideo(context.Background(), FlickSourceVideo{
		SourceID:    "source-3",
		MD5:         "abc123",
		CanPlay:     true,
		Tags:        []string{" Dance ", "dance", "MUSIC"},
		Description: "test description",
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, FlickImportOptions{
		SourceVideoDir: sourceVideo,
		SourceCoverDir: sourceCover,
		StorageRoot:    storageRoot,
	})
	if err != nil {
		t.Fatalf("ImportPlayableVideo returned error: %v", err)
	}
	if outcome.Status != FlickImportStatusImported {
		t.Fatalf("expected imported outcome, got %s", outcome.Status)
	}
	if repo.createVideoCall != 1 {
		t.Fatalf("expected create called once, got %d", repo.createVideoCall)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected one created record, got %d", len(repo.created))
	}

	created := repo.created[0]
	if created.Type != "short" {
		t.Fatalf("expected short type, got %s", created.Type)
	}
	if created.Status != "ready" {
		t.Fatalf("expected ready status, got %s", created.Status)
	}
	if !created.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected created_at %s, got %s", createdAt, created.CreatedAt)
	}
	if !created.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected updated_at %s, got %s", updatedAt, created.UpdatedAt)
	}
	if created.DurationSeconds != 19 {
		t.Fatalf("expected rounded duration 19, got %d", created.DurationSeconds)
	}
	if !reflect.DeepEqual(created.Tags, []string{"dance", "music"}) {
		t.Fatalf("unexpected normalized tags: %#v", created.Tags)
	}
	if created.Hash != "hash-2" {
		t.Fatalf("expected stored hash hash-2, got %s", created.Hash)
	}
	if created.TranscodedPath == "" || created.ThumbnailPath == "" {
		t.Fatalf("expected target paths to be populated: %#v", created)
	}
	if _, err := os.Stat(created.TranscodedPath); err != nil {
		t.Fatalf("expected copied transcoded file: %v", err)
	}
	if _, err := os.Stat(created.ThumbnailPath); err != nil {
		t.Fatalf("expected copied thumbnail file: %v", err)
	}
	if sourceMD5, ok := created.Metadata["source_md5"].(string); !ok || sourceMD5 != "abc123" {
		t.Fatalf("expected source_md5 metadata, got %#v", created.Metadata["source_md5"])
	}
}

func TestFlickImportServiceImportPlayableVideoSkipsProbeFailure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	sourceVideo := filepath.Join(tmpDir, "source", "video")
	sourceCover := filepath.Join(tmpDir, "source", "cover")
	if err := os.MkdirAll(sourceVideo, 0o755); err != nil {
		t.Fatalf("mkdir source video dir: %v", err)
	}
	if err := os.MkdirAll(sourceCover, 0o755); err != nil {
		t.Fatalf("mkdir source cover dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceVideo, "abc123.mp4"), []byte("video"), 0o644); err != nil {
		t.Fatalf("write source video: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceCover, "abc123.jpg"), []byte("cover"), 0o644); err != nil {
		t.Fatalf("write source cover: %v", err)
	}

	svc := NewFlickImportService(&fakeFlickImportRepo{})
	svc.probe = func(context.Context, string) (ffmpeg.VideoProbe, error) {
		return ffmpeg.VideoProbe{}, errors.New("probe failed")
	}

	outcome, err := svc.ImportPlayableVideo(context.Background(), FlickSourceVideo{
		SourceID: "source-4",
		MD5:      "abc123",
		CanPlay:  true,
		Tags:     []string{"tag-a"},
	}, FlickImportOptions{
		SourceVideoDir: sourceVideo,
		SourceCoverDir: sourceCover,
		StorageRoot:    filepath.Join(tmpDir, "storage"),
	})
	if err != nil {
		t.Fatalf("ImportPlayableVideo returned error: %v", err)
	}
	if outcome.Status != FlickImportStatusSkipped {
		t.Fatalf("expected skipped outcome, got %s", outcome.Status)
	}
	if outcome.Reason != FlickImportReasonProbeFailed {
		t.Fatalf("expected probe-failed reason, got %s", outcome.Reason)
	}
}
