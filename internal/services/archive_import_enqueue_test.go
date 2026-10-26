package services

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"video-server/internal/models"
)

type fakeArchiveImportDB struct {
	execArgs [][]any
	row      models.ArchiveImportBatch
	rowErr   error
}

func (f *fakeArchiveImportDB) Begin(context.Context) (pgx.Tx, error) {
	return &fakeArchiveLifecycleTx{db: f}, nil
}

func (f *fakeArchiveImportDB) Exec(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
	f.execArgs = append(f.execArgs, append([]any(nil), args...))
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (f *fakeArchiveImportDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeArchiveImportDB) QueryRow(context.Context, string, ...any) pgx.Row {
	if f.rowErr != nil {
		return errRow{err: f.rowErr}
	}
	return archiveBatchRow{batch: f.row}
}

type errRow struct{ err error }

func (r errRow) Scan(...any) error { return r.err }

type archiveBatchRow struct{ batch models.ArchiveImportBatch }

func (r archiveBatchRow) Scan(dest ...any) error {
	batch := r.batch
	*(dest[0].(*uuid.UUID)) = batch.ID
	*(dest[1].(**uuid.UUID)) = batch.UserID
	*(dest[2].(*string)) = batch.Title
	*(dest[3].(*string)) = batch.OriginalFilename
	*(dest[4].(*string)) = batch.ArchiveFormat
	*(dest[5].(*string)) = batch.OriginalPath
	*(dest[6].(*string)) = batch.ExtractedDir
	*(dest[7].(*string)) = batch.Status
	*(dest[8].(*string)) = batch.LastError
	*(dest[9].(*string)) = batch.EncodingMode
	*(dest[10].(*string)) = batch.EncodingRequestedMode
	*(dest[11].(*int)) = batch.TotalEntries
	*(dest[12].(*int)) = batch.ProcessableEntries
	*(dest[13].(*int)) = batch.ProcessedEntries
	*(dest[14].(*int)) = batch.SkippedEntries
	*(dest[15].(*int)) = batch.FailedEntries
	*(dest[16].(*string)) = batch.DefaultTitlePrefix
	*(dest[17].(*string)) = batch.DefaultDescription
	*(dest[18].(*[]byte)) = []byte("[]")
	*(dest[19].(*[]byte)) = []byte("[]")
	*(dest[20].(*[]byte)) = []byte("[]")
	*(dest[21].(*time.Time)) = batch.CreatedAt
	*(dest[22].(*time.Time)) = batch.UpdatedAt
	*(dest[23].(**time.Time)) = batch.CompletedAt
	return nil
}

type fakeArchiveEnqueuer struct {
	extractCalls int
	batchID      uuid.UUID
	password     string
	err          error
}

func (f *fakeArchiveEnqueuer) EnqueueTranscode(string, string, string, string, bool) error {
	return nil
}
func (f *fakeArchiveEnqueuer) EnqueueScrapeMovie(string, string, string) error { return nil }
func (f *fakeArchiveEnqueuer) EnqueueScrapeTV(string, string, string) error    { return nil }
func (f *fakeArchiveEnqueuer) EnqueueScrapeAV(string, string, string) error    { return nil }
func (f *fakeArchiveEnqueuer) EnqueueArchiveExtract(batchID uuid.UUID, password string) error {
	f.extractCalls++
	f.batchID = batchID
	f.password = password
	return f.err
}

func TestEnqueueArchiveExtractPersistsRequestedEncodingMode(t *testing.T) {
	t.Parallel()

	db := &fakeArchiveImportDB{}
	enqueuer := &fakeArchiveEnqueuer{}
	svc := &ArchiveImportService{db: db, enqueuer: enqueuer}
	batchID := uuid.New()
	batch := models.ArchiveImportBatch{ID: batchID, EncodingRequestedMode: archiveEncodingAuto}

	if err := svc.enqueueArchiveExtract(context.Background(), batchID, batch, "secret", archiveEncodingGBK); err != nil {
		t.Fatalf("enqueueArchiveExtract() error = %v", err)
	}
	if enqueuer.extractCalls != 1 || enqueuer.batchID != batchID || enqueuer.password != "secret" {
		t.Fatalf("enqueuer call = %d/%s/%q", enqueuer.extractCalls, enqueuer.batchID, enqueuer.password)
	}
	if len(db.execArgs) != 1 {
		t.Fatalf("exec calls = %d, want 1", len(db.execArgs))
	}
	args := db.execArgs[0]
	if len(args) != 4 || args[1] != "extracting" || args[3] != archiveEncodingGBK {
		t.Fatalf("exec args = %#v, want requested mode %q persisted", args, archiveEncodingGBK)
	}
}

func TestImportLocalArchiveCopiesIntoStorageAndEnqueues(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	downloadDir := filepath.Join(root, "downloads")
	storageRoot := filepath.Join(root, "storage")
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		t.Fatalf("mkdir downloads: %v", err)
	}
	sourcePath := filepath.Join(downloadDir, "演示.zip")
	if err := writeTestZipArchive(sourcePath, []testZipArchiveEntry{{Name: "inside.txt", Body: "zip-body"}}); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	db := &fakeArchiveImportDB{row: models.ArchiveImportBatch{Status: "extracting"}}
	enqueuer := &fakeArchiveEnqueuer{}
	svc := &ArchiveImportService{db: db, enqueuer: enqueuer, storageRoot: storageRoot}

	batch, err := svc.ImportLocalArchive(context.Background(), ArchiveImportUploadInput{
		UserID:      uuid.New(),
		Title:       "压缩包标题",
		HasPassword: true,
		Password:    "secret",
	}, sourcePath)
	if err != nil {
		t.Fatalf("ImportLocalArchive() error = %v", err)
	}
	if enqueuer.extractCalls != 1 || enqueuer.password != "secret" {
		t.Fatalf("enqueuer calls = %d password = %q", enqueuer.extractCalls, enqueuer.password)
	}
	if enqueuer.batchID == uuid.Nil || batch.ID != db.row.ID {
		t.Fatalf("batch id = %s", enqueuer.batchID)
	}

	if len(db.execArgs) < 2 {
		t.Fatalf("exec calls = %d, want insert + extracting update", len(db.execArgs))
	}
	insertArgs := db.execArgs[0]
	batchID, okID := insertArgs[0].(uuid.UUID)
	if !okID {
		t.Fatalf("insert batch id type = %T", insertArgs[0])
	}
	if insertArgs[2] != "压缩包标题" || insertArgs[4] != "zip" {
		t.Fatalf("insert args = %#v", insertArgs)
	}
	originalPath, okPath := insertArgs[5].(string)
	if !okPath {
		t.Fatalf("original path type = %T", insertArgs[5])
	}
	wantOriginal := filepath.Join(storageRoot, "archive-imports", batchID.String(), "original", "演示.zip")
	if originalPath != wantOriginal {
		t.Fatalf("original path = %q, want %q", originalPath, wantOriginal)
	}
	body, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatalf("read batch original: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("batch original is empty")
	}
	if _, err := os.Stat(sourcePath); err != nil {
		t.Fatalf("source should remain: %v", err)
	}
	updateArgs := db.execArgs[1]
	if len(updateArgs) != 4 || updateArgs[1] != "extracting" {
		t.Fatalf("extracting update args = %#v", updateArgs)
	}
}

func TestRetryExtractPersistsRequestedEncodingMode(t *testing.T) {
	t.Parallel()

	batchID := uuid.New()
	db := &fakeArchiveImportDB{row: models.ArchiveImportBatch{
		ID:                    batchID,
		ArchiveFormat:         "zip",
		OriginalPath:          "/tmp/archive.zip",
		ExtractedDir:          "/tmp/extracted",
		Status:                "needs_encoding",
		EncodingRequestedMode: archiveEncodingAuto,
	}}
	enqueuer := &fakeArchiveEnqueuer{}
	svc := &ArchiveImportService{db: db, enqueuer: enqueuer}

	if _, err := svc.RetryExtract(context.Background(), batchID, "secret", archiveEncodingGBK); err != nil {
		t.Fatalf("RetryExtract() error = %v", err)
	}
	if enqueuer.extractCalls != 1 || enqueuer.password != "secret" {
		t.Fatalf("enqueuer calls = %d password = %q", enqueuer.extractCalls, enqueuer.password)
	}
	if len(db.execArgs) < 2 {
		t.Fatalf("exec calls = %d, want clear + extracting updates", len(db.execArgs))
	}
	// The last UPDATE is the extracting transition issued by
	// enqueueArchiveExtract; it must carry the requested mode, not the stale
	// batch.EncodingRequestedMode.
	args := db.execArgs[len(db.execArgs)-1]
	if len(args) != 4 || args[3] != archiveEncodingGBK {
		t.Fatalf("last exec args = %#v, want encoding %q", args, archiveEncodingGBK)
	}
}

// fakeArchiveLifecycleTx exercises the new transaction-protected retry transition.
type fakeArchiveLifecycleTx struct {
	pgx.Tx
	db *fakeArchiveImportDB
}

func (t *fakeArchiveLifecycleTx) Exec(ctx context.Context, q string, args ...any) (pgconn.CommandTag, error) {
	return t.db.Exec(ctx, q, args...)
}
func (t *fakeArchiveLifecycleTx) QueryRow(_ context.Context, q string, args ...any) pgx.Row {
	return archiveLifecycleRow{batch: t.db.row}
}
func (t *fakeArchiveLifecycleTx) Commit(context.Context) error   { return nil }
func (t *fakeArchiveLifecycleTx) Rollback(context.Context) error { return nil }

type archiveLifecycleRow struct{ batch models.ArchiveImportBatch }

func (r archiveLifecycleRow) Scan(dest ...any) error {
	switch v := dest[0].(type) {
	case *uuid.UUID:
		*v = r.batch.ID
	case *bool:
		*v = false
	}
	return nil
}
