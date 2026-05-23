package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

)

type stubQueryRower struct {
	query string
	args  []any
	row   pgx.Row
}

func (s *stubQueryRower) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	s.query = query
	s.args = append([]any(nil), args...)
	return s.row
}

type errRowScanner struct {
	err error
}

func (s errRowScanner) Scan(dest ...any) error {
	return s.err
}

type stubExecer struct {
	query string
	args  []any
}

func (s *stubExecer) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	s.query = query
	s.args = append([]any(nil), args...)
	return pgconn.CommandTag{}, nil
}

func TestScanVideoRecordReadsOSHash(t *testing.T) {
	t.Parallel()

	videoID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	userID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	tmdbID := 998877
	now := time.Unix(1710000000, 0).UTC()

	got, err := scanVideoRecord(stubRowScanner{
		values: []any{
			videoID,
			&userID,
			&tmdbID,
			"bbbbbbbb-cccc-dddd-eeee-ffffffffffff",
			"欧美样本",
			"测试描述",
			"av",
			"av_scrape_pending",
			sql.NullInt32{Int32: 123, Valid: true},
			sql.NullInt32{Int32: 1920, Valid: true},
			sql.NullInt32{Int32: 1080, Valid: true},
			"/videos/sample.mp4",
			"/videos/sample.mp4.transcoded.mp4",
			"/videos/sample.jpg",
			sql.NullString{String: "0123456789abcdef", Valid: true},
			[]byte(`{"site_category":"western"}`),
			now,
			now,
		},
	})
	if err != nil {
		t.Fatalf("scanVideoRecord() error = %v", err)
	}

	if got.ID != videoID {
		t.Fatalf("unexpected video id: %s", got.ID)
	}
	if got.UserID == nil || *got.UserID != userID {
		t.Fatalf("unexpected user id: %#v", got.UserID)
	}
	if got.TMDBID == nil || *got.TMDBID != tmdbID {
		t.Fatalf("unexpected tmdb id: %#v", got.TMDBID)
	}
	if got.OSHash != "0123456789abcdef" {
		t.Fatalf("unexpected os_hash: %q", got.OSHash)
	}
	if got.ImageCollectionID == nil || got.ImageCollectionID.String() != "bbbbbbbb-cccc-dddd-eeee-ffffffffffff" {
		t.Fatalf("unexpected image collection id: %#v", got.ImageCollectionID)
	}
	if got.DurationSeconds != 123 || got.Width != 1920 || got.Height != 1080 {
		t.Fatalf("unexpected dimensions: %+v", got)
	}
	if string(got.Metadata) != `{"site_category":"western"}` {
		t.Fatalf("unexpected metadata: %s", string(got.Metadata))
	}
}

func TestGetVideoOSHashReturnsEmptyForNoRows(t *testing.T) {
	t.Parallel()

	rower := &stubQueryRower{row: errRowScanner{err: pgx.ErrNoRows}}
	got, err := getVideoOSHash(context.Background(), rower, uuid.New())
	if err != nil {
		t.Fatalf("getVideoOSHash() error = %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty hash, got %q", got)
	}
	if !containsNormalizedSQL(rower.query, "SELECT os_hash FROM videos WHERE id = $1") {
		t.Fatalf("unexpected query: %s", rower.query)
	}
}

func TestUpdateVideoOSHashWritesNullForEmptyHash(t *testing.T) {
	t.Parallel()

	execer := &stubExecer{}
	videoID := uuid.New()

	if err := updateVideoOSHash(context.Background(), execer, videoID, ""); err != nil {
		t.Fatalf("updateVideoOSHash() error = %v", err)
	}
	if !containsNormalizedSQL(execer.query, "UPDATE videos SET os_hash = $2, updated_at = NOW() WHERE id = $1") {
		t.Fatalf("unexpected query: %s", execer.query)
	}
	if len(execer.args) != 2 {
		t.Fatalf("unexpected args length: %d", len(execer.args))
	}
	if execer.args[0] != videoID {
		t.Fatalf("unexpected video id arg: %#v", execer.args[0])
	}
	if execer.args[1] != nil {
		t.Fatalf("expected NULL hash arg, got %#v", execer.args[1])
	}
}

func containsNormalizedSQL(sql string, pattern string) bool {
	return strings.Join(strings.Fields(strings.ToLower(sql)), " ") == strings.Join(strings.Fields(strings.ToLower(pattern)), " ")
}
