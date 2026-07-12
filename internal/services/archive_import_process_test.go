package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"video-server/internal/models"
)

func TestProcessFilePreservesDescriptionWhitespaceForVideoUpload(t *testing.T) {
	t.Parallel()

	const description = "旧标题\n  原说明\n\n"
	saveErr := errors.New("stop after save")
	fileID := uuid.New()
	batchID := uuid.New()
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "source.mp4")
	if err := os.WriteFile(sourcePath, []byte("video fixture"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	db := &processFileReloadFailureDB{
		file: models.ArchiveImportFileListItem{
			ID:           fileID,
			BatchID:      batchID,
			EntryType:    "file",
			MediaKind:    "video",
			VideoType:    "short",
			RelativePath: "目录/视频.mp4",
			FilePath:     sourcePath,
			FileSize:     int64(len("video fixture")),
			Title:        "视频",
			Description:  description,
			Status:       "pending",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		},
	}
	upload := &recordingArchiveImportUploadService{
		previewResult: UploadPreview{VideoID: uuid.New(), Status: "uploaded"},
		saveErr:       saveErr,
	}
	service := &ArchiveImportService{
		db:          db,
		uploadSvc:   upload,
		uploadTemp:  tempDir,
		storageRoot: tempDir,
	}

	_, err := service.ProcessFile(context.Background(), fileID)
	if !errors.Is(err, saveErr) {
		t.Fatalf("ProcessFile() error = %v, want %v", err, saveErr)
	}
	if len(upload.previewInputs) != 1 {
		t.Fatalf("PreviewUploadedFile() calls = %d, want 1", len(upload.previewInputs))
	}
	if len(upload.saveInputs) != 1 {
		t.Fatalf("SaveUploadedFile() calls = %d, want 1", len(upload.saveInputs))
	}
	if got := upload.saveInputs[0].Desc; got != description {
		t.Fatalf("saved description = %q, want %q", got, description)
	}
}

func TestUpdateFileTrimsDescriptionAtInputBoundary(t *testing.T) {
	t.Parallel()

	fileID := uuid.New()
	batchID := uuid.New()
	tx := &recordingArchiveImportFileTx{}
	db := &processFileReloadFailureDB{
		file: models.ArchiveImportFileListItem{
			ID:           fileID,
			BatchID:      batchID,
			EntryType:    "file",
			MediaKind:    "video",
			VideoType:    "short",
			RelativePath: "目录/视频.mp4",
			Title:        "视频",
			Status:       "pending",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		},
		tx: tx,
	}
	service := &ArchiveImportService{db: db}

	_, err := service.UpdateFile(context.Background(), fileID, ArchiveImportFileUpdateInput{
		Title:       "视频",
		Description: "  管理员说明\n\n ",
		VideoType:   "short",
	})
	if err != nil {
		t.Fatalf("UpdateFile() error = %v", err)
	}
	if len(tx.execArgs) != 1 {
		t.Fatalf("Exec() calls = %d, want 1", len(tx.execArgs))
	}
	if got := tx.execArgs[0][3]; got != "管理员说明" {
		t.Fatalf("description SQL arg = %q, want trimmed input", got)
	}
}

func TestProcessFileMarksFailedWhenMetadataReloadFails(t *testing.T) {
	t.Parallel()

	reloadErr := errors.New("forced metadata reload failure")
	fileID := uuid.New()
	batchID := uuid.New()
	db := &processFileReloadFailureDB{
		file: models.ArchiveImportFileListItem{
			ID:           fileID,
			BatchID:      batchID,
			EntryType:    "file",
			MediaKind:    "video",
			VideoType:    "short",
			RelativePath: "目录/视频.mp4",
			FilePath:     "/tmp/视频.mp4",
			Status:       "pending",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		},
		reloadErr: reloadErr,
	}
	service := &ArchiveImportService{db: db}

	_, err := service.ProcessFile(context.Background(), fileID)
	if !errors.Is(err, reloadErr) {
		t.Fatalf("ProcessFile() error = %v, want %v", err, reloadErr)
	}
	if len(db.execQueries) < 2 {
		t.Fatalf("Exec() calls = %d, want processing and failed updates", len(db.execQueries))
	}
	if !strings.Contains(db.execQueries[0], "SET status='processing'") {
		t.Fatalf("first Exec() query does not mark processing:\n%s", db.execQueries[0])
	}
	if !strings.Contains(db.execQueries[1], "SET status='failed'") {
		t.Fatalf("second Exec() query does not mark failed:\n%s", db.execQueries[1])
	}
	if len(db.execArgs[1]) != 2 || !strings.Contains(fmt.Sprint(db.execArgs[1][1]), reloadErr.Error()) {
		t.Fatalf("failure update args = %#v, want reload error", db.execArgs[1])
	}
}

type processFileReloadFailureDB struct {
	file        models.ArchiveImportFileListItem
	reloadErr   error
	tx          pgx.Tx
	queryCalls  int
	execQueries []string
	execArgs    [][]any
}

func (db *processFileReloadFailureDB) Begin(context.Context) (pgx.Tx, error) {
	if db.tx != nil {
		return db.tx, nil
	}
	return nil, errors.New("unexpected Begin call")
}

func (db *processFileReloadFailureDB) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	db.execQueries = append(db.execQueries, query)
	db.execArgs = append(db.execArgs, append([]any{}, args...))
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (db *processFileReloadFailureDB) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	db.queryCalls++
	if db.reloadErr != nil && db.queryCalls > 1 {
		return nil, db.reloadErr
	}
	return &archiveImportFileRows{files: []models.ArchiveImportFileListItem{db.file}}, nil
}

func (db *processFileReloadFailureDB) QueryRow(_ context.Context, query string, _ ...any) pgx.Row {
	switch {
	case strings.Contains(query, "FROM archive_import_batches"):
		return archiveImportScanRow(func(dest ...any) error {
			if len(dest) != 24 {
				return fmt.Errorf("batch Scan() destinations = %d, want 24", len(dest))
			}
			*(dest[0].(*uuid.UUID)) = db.file.BatchID
			*(dest[1].(**uuid.UUID)) = nil
			for _, index := range []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 16, 17} {
				*(dest[index].(*string)) = ""
			}
			for _, index := range []int{11, 12, 13, 14, 15} {
				*(dest[index].(*int)) = 0
			}
			for _, index := range []int{18, 19, 20} {
				*(dest[index].(*[]byte)) = []byte("[]")
			}
			*(dest[21].(*time.Time)) = db.file.CreatedAt
			*(dest[22].(*time.Time)) = db.file.UpdatedAt
			*(dest[23].(**time.Time)) = nil
			return nil
		})
	case strings.Contains(query, "SELECT batch_id FROM archive_import_files"):
		return archiveImportScanRow(func(dest ...any) error {
			*(dest[0].(*uuid.UUID)) = db.file.BatchID
			return nil
		})
	case strings.Contains(query, "COUNT(*)::INT AS total"):
		return archiveImportScanRow(func(dest ...any) error {
			for index, value := range []int{1, 1, 0, 0, 1} {
				*(dest[index].(*int)) = value
			}
			return nil
		})
	default:
		return archiveImportScanRow(func(...any) error {
			return fmt.Errorf("unexpected QueryRow: %s", query)
		})
	}
}

type archiveImportScanRow func(dest ...any) error

func (row archiveImportScanRow) Scan(dest ...any) error {
	return row(dest...)
}

type recordingArchiveImportUploadService struct {
	previewInputs []LocalUploadInput
	previewResult UploadPreview
	previewErr    error
	saveInputs    []LocalUploadInput
	saveErr       error
}

func (service *recordingArchiveImportUploadService) PreviewUploadedFile(_ context.Context, input LocalUploadInput) (UploadPreview, error) {
	service.previewInputs = append(service.previewInputs, input)
	return service.previewResult, service.previewErr
}

func (service *recordingArchiveImportUploadService) SaveUploadedFile(_ context.Context, input LocalUploadInput, _ int64) (UploadResult, error) {
	service.saveInputs = append(service.saveInputs, input)
	return UploadResult{}, service.saveErr
}

type recordingArchiveImportFileTx struct {
	pgx.Tx
	execArgs      [][]any
	commitCalls   int
	rollbackCalls int
}

func (tx *recordingArchiveImportFileTx) Exec(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
	tx.execArgs = append(tx.execArgs, append([]any{}, args...))
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (tx *recordingArchiveImportFileTx) Commit(context.Context) error {
	tx.commitCalls++
	return nil
}

func (tx *recordingArchiveImportFileTx) Rollback(context.Context) error {
	tx.rollbackCalls++
	return nil
}
