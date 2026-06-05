package queue

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

func TestBuildOrphanFileScanTaskOptionsIncludesFixedRetryAndTimeout(t *testing.T) {
	t.Parallel()

	opts := buildOrphanFileScanTaskOptions("system")

	assertOption(t, opts, asynq.QueueOpt, "system")
	assertOption(t, opts, asynq.ProcessInOpt, 2*time.Second)
	assertOption(t, opts, asynq.TimeoutOpt, 6*time.Hour)
	assertOption(t, opts, asynq.MaxRetryOpt, 3)
}

func TestOrphanFileScanRootsReturnsKnownBusinessTrees(t *testing.T) {
	t.Parallel()

	roots := orphanFileScanRoots("/storage")
	if len(roots) != 6 {
		t.Fatalf("expected 6 scan roots, got %d: %v", len(roots), roots)
	}
	if roots[0] != filepath.Join("/storage", "videos") {
		t.Fatalf("unexpected first root: %v", roots[0])
	}
	if roots[len(roots)-1] != filepath.Join("/storage", "tv") {
		t.Fatalf("unexpected last root: %v", roots[len(roots)-1])
	}
}

func TestScanOrphanFilesSkipsReferencedFilesAndKeepsRelativePath(t *testing.T) {
	t.Parallel()

	storageRoot := t.TempDir()
	referencedPath := filepath.Join(storageRoot, "videos", "referenced.mp4")
	orphanPath := filepath.Join(storageRoot, "images", "orphan.jpg")

	if err := os.MkdirAll(filepath.Dir(referencedPath), 0o755); err != nil {
		t.Fatalf("mkdir referenced dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(orphanPath), 0o755); err != nil {
		t.Fatalf("mkdir orphan dir: %v", err)
	}
	if err := os.WriteFile(referencedPath, []byte("referenced"), 0o644); err != nil {
		t.Fatalf("write referenced file: %v", err)
	}
	if err := os.WriteFile(orphanPath, []byte("orphan-file"), 0o644); err != nil {
		t.Fatalf("write orphan file: %v", err)
	}

	items, total, err := scanOrphanFiles(storageRoot, map[string]struct{}{
		filepath.Clean(referencedPath): {},
	})
	if err != nil {
		t.Fatalf("scanOrphanFiles() error = %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total files=2, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 orphan item, got %d: %+v", len(items), items)
	}
	item := items[0]
	if item.FilePath != filepath.Clean(orphanPath) {
		t.Fatalf("unexpected orphan file path: %s", item.FilePath)
	}
	if item.RelativePath != filepath.Join("images", "orphan.jpg") {
		t.Fatalf("unexpected relative path: %s", item.RelativePath)
	}
	if item.SizeBytes != int64(len("orphan-file")) {
		t.Fatalf("unexpected size bytes: %d", item.SizeBytes)
	}
	if item.ModTime.IsZero() {
		t.Fatal("expected mod time to be populated")
	}
}
