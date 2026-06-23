package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"video-server/internal/models"
)

func TestArchiveImportShouldProcessInBatch(t *testing.T) {
	tests := []struct {
		name string
		file models.ArchiveImportFileListItem
		want bool
	}{
		{
			name: "pending video",
			file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "video", Status: "pending"},
			want: true,
		},
		{
			name: "failed image retry",
			file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "image", Status: "failed"},
			want: true,
		},
		{
			name: "processing video retry",
			file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "video", Status: "processing"},
			want: true,
		},
		{
			name: "ready video",
			file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "video", Status: "ready"},
			want: false,
		},
		{
			name: "existing image",
			file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "image", Status: "existing"},
			want: false,
		},
		{
			name: "skipped archive",
			file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "archive", Status: "skipped"},
			want: false,
		},
		{
			name: "directory",
			file: models.ArchiveImportFileListItem{EntryType: "directory", MediaKind: "directory", Status: "skipped"},
			want: false,
		},
		{
			name: "other pending file",
			file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "other", Status: "pending"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldProcessArchiveFileInBatch(tt.file); got != tt.want {
				t.Fatalf("shouldProcessArchiveFileInBatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateArchiveBatchDeletionRejectsProcessing(t *testing.T) {
	t.Parallel()

	err := validateArchiveBatchDeletion(models.ArchiveImportBatch{Status: "processing"})
	if err != ErrArchiveBatchBusy {
		t.Fatalf("validateArchiveBatchDeletion() error = %v, want %v", err, ErrArchiveBatchBusy)
	}
}

func TestArchiveBatchCleanupPaths(t *testing.T) {
	t.Parallel()

	batch := models.ArchiveImportBatch{
		OriginalPath: filepath.Join("/tmp", "archive-imports", "batch-1", "original", "demo.zip"),
		ExtractedDir: filepath.Join("/tmp", "archive-imports", "batch-1", "extracted"),
	}
	got := archiveBatchCleanupPaths(batch)
	if len(got) != 2 {
		t.Fatalf("archiveBatchCleanupPaths() len = %d, want 2", len(got))
	}
	if got[0] != filepath.Join("/tmp", "archive-imports", "batch-1", "original") {
		t.Fatalf("archiveBatchCleanupPaths()[0] = %q", got[0])
	}
	if got[1] != filepath.Join("/tmp", "archive-imports", "batch-1", "extracted") {
		t.Fatalf("archiveBatchCleanupPaths()[1] = %q", got[1])
	}
}

func TestArchiveBatchCleanupPathsDedupes(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	originalDir := filepath.Join(root, "original")
	if err := os.MkdirAll(originalDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	batch := models.ArchiveImportBatch{
		OriginalPath: filepath.Join(originalDir, "demo.zip"),
		ExtractedDir: originalDir,
	}
	got := archiveBatchCleanupPaths(batch)
	if len(got) != 1 {
		t.Fatalf("archiveBatchCleanupPaths() len = %d, want 1", len(got))
	}
	if got[0] != originalDir {
		t.Fatalf("archiveBatchCleanupPaths()[0] = %q, want %q", got[0], originalDir)
	}
}

func TestArchiveVideoSourceFilenameUsesStableNameAndExt(t *testing.T) {
	t.Parallel()

	file := models.ArchiveImportFileListItem{
		ID:           uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		RelativePath: "Season 1/Show.S01E02.mkv",
		FilePath:     "/tmp/archive-video.mov",
	}

	if got := archiveVideoSourceFilename(file); got != "source-original.mkv" {
		t.Fatalf("archiveVideoSourceFilename() = %q", got)
	}
}

func TestCopyArchiveVideoSourceFileUsesStablePathOutsideWorkDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workPath := filepath.Join(root, "archive-imports", "batch-1", "work", "clip.mp4")
	if err := os.MkdirAll(filepath.Dir(workPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	const body = "demo-video"
	if err := os.WriteFile(workPath, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	svc := &ArchiveImportService{uploadTemp: root}
	file := models.ArchiveImportFileListItem{
		ID:           uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		BatchID:      uuid.MustParse("44444444-4444-4444-4444-444444444444"),
		RelativePath: "nested/demo clip.mp4",
		FilePath:     workPath,
	}

	got, err := svc.copyArchiveVideoSourceFile(file.ID, file, workPath)
	if err != nil {
		t.Fatalf("copyArchiveVideoSourceFile() error = %v", err)
	}

	want := filepath.Join(root, "videos", file.ID.String(), "source-original.mp4")
	if got != want {
		t.Fatalf("copyArchiveVideoSourceFile() path = %q, want %q", got, want)
	}
	data, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != body {
		t.Fatalf("copied file body = %q, want %q", string(data), body)
	}
	if _, err := os.Stat(workPath); err != nil {
		t.Fatalf("work file should remain for caller cleanup, stat error = %v", err)
	}
}

func TestArchiveFileTitleForScannedFileUsesBatchTitleOnlyForVideos(t *testing.T) {
	t.Parallel()

	batch := models.ArchiveImportBatch{
		Title:              "压缩包标题",
		DefaultTitlePrefix: "视频默认标题",
	}

	if got := archiveFileTitleForScannedFile("video", "第一组/clip-01.mp4", batch); got != "视频默认标题" {
		t.Fatalf("video title = %q, want batch title", got)
	}
	if got := archiveFileTitleForScannedFile("image", "第一组/cover image.jpg", batch); got != "cover image" {
		t.Fatalf("image title = %q, want filename-derived title", got)
	}
}

func TestArchiveFileTitleForScannedFileFallsBackToBatchTitleAndFilename(t *testing.T) {
	t.Parallel()

	if got := archiveFileTitleForScannedFile(
		"video",
		"clips/fallback-name.mp4",
		models.ArchiveImportBatch{Title: "批次标题"},
	); got != "批次标题" {
		t.Fatalf("video title with batch title = %q, want batch title", got)
	}
	if got := archiveFileTitleForScannedFile("video", "clips/fallback-name.mp4", models.ArchiveImportBatch{}); got != "fallback-name" {
		t.Fatalf("video title without batch title = %q, want filename-derived fallback", got)
	}
}

func TestArchiveFileTitleForProcessingBlankVideoUsesBatchTitle(t *testing.T) {
	t.Parallel()

	batch := models.ArchiveImportBatch{DefaultTitlePrefix: "视频默认标题"}
	file := models.ArchiveImportFileListItem{
		MediaKind:    "video",
		RelativePath: "第二组/demo.mp4",
		Title:        "",
	}
	if got := archiveFileTitleForProcessing(file, batch); got != "视频默认标题" {
		t.Fatalf("processing title = %q, want batch title", got)
	}

	file.Title = "单文件标题"
	if got := archiveFileTitleForProcessing(file, batch); got != "单文件标题" {
		t.Fatalf("processing explicit title = %q, want explicit title", got)
	}
}

func TestArchiveFileTitleForUpdateBlankVideoUsesBatchTitle(t *testing.T) {
	t.Parallel()

	video := models.ArchiveImportFileListItem{
		MediaKind:    "video",
		RelativePath: "第三组/demo.mp4",
		Title:        "旧单文件标题",
	}
	if got := archiveFileTitleForUpdate("", video, "视频默认标题"); got != "视频默认标题" {
		t.Fatalf("blank video update title = %q, want batch title", got)
	}
	if got := archiveFileTitleForUpdate(" 新标题 ", video, "视频默认标题"); got != "新标题" {
		t.Fatalf("explicit video update title = %q, want trimmed input title", got)
	}

	image := models.ArchiveImportFileListItem{
		MediaKind:    "image",
		RelativePath: "第三组/cover.jpg",
		Title:        "图片原标题",
	}
	if got := archiveFileTitleForUpdate("", image, "视频默认标题"); got != "图片原标题" {
		t.Fatalf("blank image update title = %q, want existing image title", got)
	}
}

func TestNormalizeArchiveVideoImageCollectionIDs(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("55555555-5555-5555-8555-555555555555")
	got, err := normalizeArchiveVideoImageCollectionIDs([]uuid.UUID{id, id})
	if err != nil {
		t.Fatalf("normalizeArchiveVideoImageCollectionIDs() error = %v", err)
	}
	if len(got) != 1 || got[0] != id {
		t.Fatalf("normalizeArchiveVideoImageCollectionIDs() = %#v, want single %s", got, id)
	}

	empty, err := normalizeArchiveVideoImageCollectionIDs(nil)
	if err != nil {
		t.Fatalf("normalizeArchiveVideoImageCollectionIDs(nil) error = %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("normalizeArchiveVideoImageCollectionIDs(nil) = %#v, want empty", empty)
	}
}

func TestNormalizeArchiveVideoImageCollectionIDsRejectsMultiple(t *testing.T) {
	t.Parallel()

	_, err := normalizeArchiveVideoImageCollectionIDs([]uuid.UUID{
		uuid.MustParse("66666666-6666-4666-8666-666666666666"),
		uuid.MustParse("77777777-7777-4777-8777-777777777777"),
	})
	if err == nil {
		t.Fatal("expected error for multiple image collections on one video")
	}
}

func TestArchiveImageCollectionsForScannedFileOnlyAppliesToImages(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("88888888-8888-4888-8888-888888888888")
	if got := archiveImageCollectionsForScannedFile("video", []uuid.UUID{id}); len(got) != 0 {
		t.Fatalf("video default image collections = %#v, want empty", got)
	}
	got := archiveImageCollectionsForScannedFile("image", []uuid.UUID{id})
	if len(got) != 1 || got[0] != id {
		t.Fatalf("image default image collections = %#v, want %s", got, id)
	}
	got[0] = uuid.Nil
	next := archiveImageCollectionsForScannedFile("image", []uuid.UUID{id})
	if next[0] != id {
		t.Fatal("archiveImageCollectionsForScannedFile should return a copy")
	}
}

func TestArchiveVideoImageCollectionsForProcessingIgnoresInheritedBatchDefault(t *testing.T) {
	t.Parallel()

	defaultID := uuid.MustParse("99999999-9999-4999-8999-999999999999")
	got, err := archiveVideoImageCollectionsForProcessing(
		models.ArchiveImportFileListItem{MediaKind: "video", ImageCollectionIDs: []uuid.UUID{defaultID}},
		models.ArchiveImportBatch{DefaultImageCollectionIDs: []uuid.UUID{defaultID}},
	)
	if err != nil {
		t.Fatalf("archiveVideoImageCollectionsForProcessing() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("inherited default image collections = %#v, want empty", got)
	}
}

func TestArchiveVideoImageCollectionsForProcessingKeepsExplicitVideoLink(t *testing.T) {
	t.Parallel()

	defaultID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	explicitID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	got, err := archiveVideoImageCollectionsForProcessing(
		models.ArchiveImportFileListItem{MediaKind: "video", ImageCollectionIDs: []uuid.UUID{explicitID}},
		models.ArchiveImportBatch{DefaultImageCollectionIDs: []uuid.UUID{defaultID}},
	)
	if err != nil {
		t.Fatalf("archiveVideoImageCollectionsForProcessing() error = %v", err)
	}
	if len(got) != 1 || got[0] != explicitID {
		t.Fatalf("explicit video image collection = %#v, want %s", got, explicitID)
	}
}
