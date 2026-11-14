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

	"video-server/internal/models"
)

type fakeLocalFileTaskRepo struct {
	task             models.CloudDownloadTask
	getErr           error
	failGetAfterMark bool
	marked           bool
	markedBy         *uuid.UUID
}

func (f *fakeLocalFileTaskRepo) GetByID(context.Context, int64) (models.CloudDownloadTask, error) {
	if f.getErr != nil {
		return models.CloudDownloadTask{}, f.getErr
	}
	if f.failGetAfterMark && f.marked {
		return models.CloudDownloadTask{}, errors.New("reload failed")
	}
	if f.task.ID == 0 {
		return models.CloudDownloadTask{}, pgx.ErrNoRows
	}
	return f.task, nil
}

func (f *fakeLocalFileTaskRepo) MarkLocalFileDeleted(_ context.Context, _ int64, deletedBy *uuid.UUID) error {
	f.marked = true
	f.markedBy = deletedBy
	now := time.Now().UTC()
	f.task.LocalFileDeletedAt = &now
	f.task.LocalFileDeletedBy = deletedBy
	return nil
}

type fakeLocalFileImportRepo struct {
	item *models.CloudDownloadImport
	err  error
}

func (f *fakeLocalFileImportRepo) GetByTaskID(context.Context, int64) (models.CloudDownloadImport, error) {
	if f.err != nil {
		return models.CloudDownloadImport{}, f.err
	}
	if f.item == nil {
		return models.CloudDownloadImport{}, pgx.ErrNoRows
	}
	return *f.item, nil
}

func newLocalFileHarness(t *testing.T) (string, *fakeLocalFileTaskRepo, *fakeLocalFileImportRepo, *CloudLocalFileDeletionService) {
	t.Helper()
	root := t.TempDir()
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	downloadDir := filepath.Join(root, "downloads")
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		t.Fatalf("mkdir downloads: %v", err)
	}
	tasks := &fakeLocalFileTaskRepo{}
	imports := &fakeLocalFileImportRepo{}
	svc := NewCloudLocalFileDeletionService(tasks, imports, downloadDir, nil)
	return downloadDir, tasks, imports, svc
}

func writeLocalFile(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestCloudLocalFileDeletionRemovesFileAndMarksTask(t *testing.T) {
	t.Parallel()

	downloadDir, tasks, _, svc := newLocalFileHarness(t)
	path := filepath.Join(downloadDir, "demo.mkv")
	writeLocalFile(t, path, "video-body")
	tasks.task = models.CloudDownloadTask{ID: 7, Status: models.CloudDownloadCompleted, LocalPath: path}
	actor := uuid.New()

	task, err := svc.Delete(context.Background(), 7, &actor)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("downloaded file still exists: %v", statErr)
	}
	if !tasks.marked || tasks.markedBy == nil || *tasks.markedBy != actor {
		t.Fatalf("mark = %v by %v, want %s", tasks.marked, tasks.markedBy, actor)
	}
	if task.LocalFileDeletedAt == nil {
		t.Fatal("returned task must carry local_file_deleted_at")
	}
	if task.Status != models.CloudDownloadCompleted {
		t.Fatalf("status = %s, want completed", task.Status)
	}
}

func TestCloudLocalFileDeletionRejectsActiveImport(t *testing.T) {
	t.Parallel()

	downloadDir, tasks, imports, svc := newLocalFileHarness(t)
	path := filepath.Join(downloadDir, "demo.mkv")
	writeLocalFile(t, path, "video-body")
	tasks.task = models.CloudDownloadTask{ID: 7, Status: models.CloudDownloadCompleted, LocalPath: path}
	imports.item = &models.CloudDownloadImport{Status: models.CloudImportCopying}

	if _, err := svc.Delete(context.Background(), 7, nil); !errors.Is(err, models.ErrCloudLocalFileImportActive) {
		t.Fatalf("error = %v, want ErrCloudLocalFileImportActive", err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("file must survive a refused deletion: %v", statErr)
	}
	if tasks.marked {
		t.Fatal("refused deletion must not mark the task")
	}
}

func TestCloudLocalFileDeletionAllowsFailedImport(t *testing.T) {
	t.Parallel()

	downloadDir, tasks, imports, svc := newLocalFileHarness(t)
	path := filepath.Join(downloadDir, "demo.mkv")
	writeLocalFile(t, path, "video-body")
	tasks.task = models.CloudDownloadTask{ID: 7, Status: models.CloudDownloadCompleted, LocalPath: path}
	imports.item = &models.CloudDownloadImport{Status: models.CloudImportFailed}

	if _, err := svc.Delete(context.Background(), 7, nil); err != nil {
		t.Fatalf("failed import must not block cleanup: %v", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("file still exists: %v", statErr)
	}
}

func TestCloudLocalFileDeletionRejectsNonCompletedTask(t *testing.T) {
	t.Parallel()

	downloadDir, tasks, _, svc := newLocalFileHarness(t)
	path := filepath.Join(downloadDir, "demo.mkv")
	writeLocalFile(t, path, "video-body")
	tasks.task = models.CloudDownloadTask{ID: 7, Status: models.CloudDownloadDownloading, LocalPath: path}

	if _, err := svc.Delete(context.Background(), 7, nil); !errors.Is(err, models.ErrCloudLocalFileNotCompleted) {
		t.Fatalf("error = %v, want ErrCloudLocalFileNotCompleted", err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("file must survive an active download: %v", statErr)
	}
}

func TestCloudLocalFileDeletionIsIdempotentWhenFileMissing(t *testing.T) {
	t.Parallel()

	downloadDir, tasks, _, svc := newLocalFileHarness(t)
	path := filepath.Join(downloadDir, "gone.mkv")
	tasks.task = models.CloudDownloadTask{ID: 7, Status: models.CloudDownloadCompleted, LocalPath: path}

	if _, err := svc.Delete(context.Background(), 7, nil); err != nil {
		t.Fatalf("missing file must be an idempotent success: %v", err)
	}
	if !tasks.marked {
		t.Fatal("missing file must still be recorded as deleted")
	}
}

func TestCloudLocalFileDeletionRejectsUnmanagedPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	downloadDir := filepath.Join(root, "downloads")
	outsideDir := filepath.Join(root, "outside")
	for _, dir := range []string{downloadDir, outsideDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	outsideFile := filepath.Join(outsideDir, "demo.mkv")
	writeLocalFile(t, outsideFile, "video-body")

	cases := []struct {
		name string
		path string
	}{
		{name: "outside download dir", path: outsideFile},
		{name: "parent escape", path: filepath.Join(downloadDir, "..", "outside", "demo.mkv")},
		{name: "directory", path: downloadDir},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tasks := &fakeLocalFileTaskRepo{task: models.CloudDownloadTask{ID: 7, Status: models.CloudDownloadCompleted, LocalPath: tt.path}}
			svc := NewCloudLocalFileDeletionService(tasks, &fakeLocalFileImportRepo{}, downloadDir, nil)
			if _, err := svc.Delete(context.Background(), 7, nil); !errors.Is(err, models.ErrCloudLocalFileUnavailable) {
				t.Fatalf("error = %v, want ErrCloudLocalFileUnavailable", err)
			}
			if tasks.marked {
				t.Fatal("unmanaged path must not mark the task")
			}
		})
	}
	if _, statErr := os.Stat(outsideFile); statErr != nil {
		t.Fatalf("outside file must not be touched: %v", statErr)
	}
}

func TestCloudLocalFileDeletionKeepsSuccessWhenReloadFails(t *testing.T) {
	t.Parallel()

	downloadDir, tasks, _, svc := newLocalFileHarness(t)
	path := filepath.Join(downloadDir, "demo.mkv")
	writeLocalFile(t, path, "video-body")
	tasks.task = models.CloudDownloadTask{ID: 7, Status: models.CloudDownloadCompleted, LocalPath: path}
	tasks.failGetAfterMark = true
	actor := uuid.New()

	task, err := svc.Delete(context.Background(), 7, &actor)
	if err != nil {
		t.Fatalf("a completed deletion must not fail because of the reload: %v", err)
	}
	if task.LocalFileDeletedAt == nil || task.LocalFileDeletedBy == nil || *task.LocalFileDeletedBy != actor {
		t.Fatalf("fallback task must carry the deletion facts: %+v", task)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("file still exists: %v", statErr)
	}
}

func TestCloudLocalFileDeletionRejectsTaskWithoutRecordedPath(t *testing.T) {
	t.Parallel()

	_, tasks, _, svc := newLocalFileHarness(t)
	tasks.task = models.CloudDownloadTask{ID: 7, Status: models.CloudDownloadCompleted, LocalPath: ""}

	if _, err := svc.Delete(context.Background(), 7, nil); !errors.Is(err, models.ErrCloudLocalFileNotRecorded) {
		t.Fatalf("error = %v, want ErrCloudLocalFileNotRecorded", err)
	}
	if tasks.marked {
		t.Fatal("a task without a recorded path must not be marked")
	}
}

func TestCloudLocalFileDeletionIsIdempotentThroughSymlinkedDownloadRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	realDir := filepath.Join(root, "real-downloads")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatalf("mkdir real: %v", err)
	}
	linkedDir := filepath.Join(root, "downloads")
	if err := os.Symlink(realDir, linkedDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	tasks := &fakeLocalFileTaskRepo{task: models.CloudDownloadTask{
		ID:        7,
		Status:    models.CloudDownloadCompleted,
		LocalPath: filepath.Join(linkedDir, "already-gone.mkv"),
	}}
	svc := NewCloudLocalFileDeletionService(tasks, &fakeLocalFileImportRepo{}, linkedDir, nil)

	if _, err := svc.Delete(context.Background(), 7, nil); err != nil {
		t.Fatalf("a missing file under a symlinked root must stay idempotent: %v", err)
	}
	if !tasks.marked {
		t.Fatal("missing file under a symlinked root must be marked")
	}

	existing := filepath.Join(linkedDir, "still-here.mkv")
	writeLocalFile(t, existing, "video-body")
	tasks.task = models.CloudDownloadTask{ID: 8, Status: models.CloudDownloadCompleted, LocalPath: existing}
	if _, err := svc.Delete(context.Background(), 8, nil); err != nil {
		t.Fatalf("existing file under a symlinked root: %v", err)
	}
	if _, statErr := os.Stat(existing); !os.IsNotExist(statErr) {
		t.Fatalf("file still exists: %v", statErr)
	}
}

func TestCloudLocalFileDeletionReturnsNotFound(t *testing.T) {
	t.Parallel()

	_, _, _, svc := newLocalFileHarness(t)
	if _, err := svc.Delete(context.Background(), 99, nil); !errors.Is(err, models.ErrCloudLocalFileNotFound) {
		t.Fatalf("error = %v, want ErrCloudLocalFileNotFound", err)
	}
}

func TestCloudLocalFilePreviewReportsSizeAndDeletion(t *testing.T) {
	t.Parallel()

	downloadDir, tasks, _, svc := newLocalFileHarness(t)
	path := filepath.Join(downloadDir, "demo.mkv")
	writeLocalFile(t, path, "video-body")
	tasks.task = models.CloudDownloadTask{ID: 7, Status: models.CloudDownloadCompleted, LocalPath: path}

	preview, err := svc.Preview(context.Background(), 7)
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	if !preview.Exists || !preview.Managed || preview.Size != int64(len("video-body")) || preview.Path != path || preview.Deleted {
		t.Fatalf("preview = %+v", preview)
	}

	deletedAt := time.Now().UTC()
	tasks.task.LocalFileDeletedAt = &deletedAt
	preview, err = svc.Preview(context.Background(), 7)
	if err != nil {
		t.Fatalf("Preview() after deletion error = %v", err)
	}
	if !preview.Deleted {
		t.Fatalf("preview must report the recorded deletion: %+v", preview)
	}
}

func TestCloudLocalFileDeletionPreviewKeepsUnmanagedPathInvisible(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	downloadDir := filepath.Join(root, "downloads")
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	outside := filepath.Join(root, "outside.mkv")
	writeLocalFile(t, outside, "x")
	tasks := &fakeLocalFileTaskRepo{task: models.CloudDownloadTask{ID: 7, Status: models.CloudDownloadCompleted, LocalPath: outside}}
	svc := NewCloudLocalFileDeletionService(tasks, &fakeLocalFileImportRepo{}, downloadDir, nil)

	preview, err := svc.Preview(context.Background(), 7)
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	if preview.Exists || preview.Managed || preview.Size != 0 {
		t.Fatalf("unmanaged path must not be reported as a managed file: %+v", preview)
	}
}

func TestCloudFileDeletionAllowsRemoteCleanupAfterExplicitLocalDeletion(t *testing.T) {
	t.Parallel()

	deletedAt := time.Now().UTC()
	svc := &CloudFileDeletionService{downloadDir: t.TempDir()}
	task := models.CloudDownloadTask{ID: 7, Status: models.CloudDownloadCompleted, LocalPath: "", LocalFileDeletedAt: &deletedAt}
	if !svc.localSourceOrImported(context.Background(), task) {
		t.Fatal("explicit local deletion must allow 115 cleanup")
	}
	task.LocalFileDeletedAt = nil
	if svc.localSourceOrImported(context.Background(), task) {
		t.Fatal("missing local file without import or explicit deletion must stay blocked")
	}
}
