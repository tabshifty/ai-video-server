package services

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestChunkUploadCompleteFileSurvivesAbort(t *testing.T) {
	t.Parallel()

	svc := NewChunkUploadService(t.TempDir())
	userID := uuid.New()

	session, err := svc.Init(context.Background(), userID, "demo.mp4", 6, 3, 2, "hash", "short", "t", "d", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("init session: %v", err)
	}

	if _, err := svc.SaveChunk(context.Background(), session.ID, 0, bytes.NewReader([]byte("abc"))); err != nil {
		t.Fatalf("save chunk 0: %v", err)
	}
	if _, err := svc.SaveChunk(context.Background(), session.ID, 1, bytes.NewReader([]byte("def"))); err != nil {
		t.Fatalf("save chunk 1: %v", err)
	}

	_, mergedPath, err := svc.Complete(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("complete session: %v", err)
	}

	if err := svc.Abort(session.ID); err != nil {
		t.Fatalf("abort session: %v", err)
	}

	raw, err := os.ReadFile(mergedPath)
	if err != nil {
		t.Fatalf("read merged file after abort: %v", err)
	}
	if string(raw) != "abcdef" {
		t.Fatalf("merged content = %q, want %q", string(raw), "abcdef")
	}
}
