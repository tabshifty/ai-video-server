package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestTranscodedFileSizeWritesRejectNonPositiveValues(t *testing.T) {
	t.Parallel()

	repo := NewVideoRepository(nil)
	videoID := uuid.New()
	if err := repo.UpdateTranscodeResult(context.Background(), videoID, "/storage/videos/video.mp4", "/storage/videos/thumb.jpg", 0, 10, 1920, 1080, map[string]any{}); err == nil {
		t.Fatal("UpdateTranscodeResult() accepted zero transcoded file size")
	}
	if err := repo.CreateImportedReadyVideo(context.Background(), ImportedReadyVideo{ID: videoID}); err == nil {
		t.Fatal("CreateImportedReadyVideo() accepted zero transcoded file size")
	}
}
