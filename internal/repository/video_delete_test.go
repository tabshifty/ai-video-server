package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeExecer struct {
	statements []string
	failAt     int
	err        error
}

func (f *fakeExecer) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	f.statements = append(f.statements, sql)
	if f.failAt > 0 && len(f.statements) == f.failAt {
		return pgconn.CommandTag{}, f.err
	}
	return pgconn.CommandTag{}, nil
}

func TestDeleteVideoDependenciesOrder(t *testing.T) {
	t.Parallel()

	videoID := uuid.New()
	execer := &fakeExecer{}

	if err := deleteVideoDependencies(context.Background(), execer, videoID); err != nil {
		t.Fatalf("deleteVideoDependencies() error = %v", err)
	}

	if len(execer.statements) != 3 {
		t.Fatalf("statements count = %d, want 3", len(execer.statements))
	}
	if execer.statements[0] != "DELETE FROM transcoding_jobs WHERE video_id = $1" {
		t.Fatalf("statement[0] = %q", execer.statements[0])
	}
	if execer.statements[1] != "DELETE FROM user_video_actions WHERE video_id = $1" {
		t.Fatalf("statement[1] = %q", execer.statements[1])
	}
	if execer.statements[2] != "DELETE FROM videos WHERE id = $1" {
		t.Fatalf("statement[2] = %q", execer.statements[2])
	}
}

func TestDeleteVideoDependenciesError(t *testing.T) {
	t.Parallel()

	videoID := uuid.New()
	execer := &fakeExecer{
		failAt: 2,
		err:    errors.New("boom"),
	}

	err := deleteVideoDependencies(context.Background(), execer, videoID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "delete user actions by video id: boom" {
		t.Fatalf("error = %q", err.Error())
	}
}
