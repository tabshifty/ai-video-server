package services

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

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

	svc := &ArchiveImportService{storageRoot: root}
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

func TestSanitizeArchiveTextReplacesInvalidUTF8(t *testing.T) {
	t.Parallel()

	input := "解压失败: " + string([]byte{0xc9, 0xcf}) + " 文件"
	got := sanitizeArchiveText(input)
	if !utf8.ValidString(got) {
		t.Fatalf("sanitizeArchiveText() should return valid UTF-8, got %q", got)
	}
	if !strings.Contains(got, "解压失败:") || !strings.Contains(got, "文件") {
		t.Fatalf("sanitizeArchiveText() should preserve valid text, got %q", got)
	}
	if strings.Contains(got, string([]byte{0xc9, 0xcf})) {
		t.Fatalf("sanitizeArchiveText() should remove raw invalid bytes, got %q", got)
	}
}

func TestSanitizeArchiveErrorReplacesInvalidUTF8(t *testing.T) {
	t.Parallel()

	got := sanitizeArchiveError(errors.New("外部命令报错: " + string([]byte{0xc9, 0xcf})))
	if !utf8.ValidString(got) {
		t.Fatalf("sanitizeArchiveError() should return valid UTF-8, got %q", got)
	}
	if !strings.Contains(got, "外部命令报错:") {
		t.Fatalf("sanitizeArchiveError() should preserve valid text, got %q", got)
	}
}

func TestDecodeArchiveZipEntryNameAutoPrefersUTF8(t *testing.T) {
	t.Parallel()

	const rawName = "第一组/测试视频.mp4"
	got, mode, err := decodeArchiveZipEntryName(rawName, archiveEncodingAuto)
	if err != nil {
		t.Fatalf("decodeArchiveZipEntryName() error = %v", err)
	}
	if got != rawName {
		t.Fatalf("decodeArchiveZipEntryName() name = %q, want %q", got, rawName)
	}
	if mode != archiveEncodingUTF8 {
		t.Fatalf("decodeArchiveZipEntryName() mode = %q, want %q", mode, archiveEncodingUTF8)
	}
}

func TestDecodeArchiveZipEntryNameAutoFallsBackToGBK(t *testing.T) {
	t.Parallel()

	const want = "第一组/测试视频.mp4"
	rawName := mustEncodeGBKString(t, want)

	got, mode, err := decodeArchiveZipEntryName(rawName, archiveEncodingAuto)
	if err != nil {
		t.Fatalf("decodeArchiveZipEntryName() error = %v", err)
	}
	if got != want {
		t.Fatalf("decodeArchiveZipEntryName() name = %q, want %q", got, want)
	}
	if mode != archiveEncodingGBK {
		t.Fatalf("decodeArchiveZipEntryName() mode = %q, want %q", mode, archiveEncodingGBK)
	}
}

func TestDecodeArchiveZipEntryNameExplicitUTF8RejectsGBKBytes(t *testing.T) {
	t.Parallel()

	rawName := mustEncodeGBKString(t, "第一组/测试视频.mp4")
	if _, _, err := decodeArchiveZipEntryName(rawName, archiveEncodingUTF8); !errors.Is(err, ErrArchiveEncodingRequired) {
		t.Fatalf("decodeArchiveZipEntryName() error = %v, want %v", err, ErrArchiveEncodingRequired)
	}
}

func TestArchiveBatchFailureStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		batch         models.ArchiveImportBatch
		requestedMode string
		err           error
		want          string
	}{
		{
			name:          "password errors prefer needs password",
			requestedMode: archiveEncodingAuto,
			err:           ErrArchivePasswordRequired,
			want:          "needs_password",
		},
		{
			name:          "auto decode failure asks for encoding",
			requestedMode: archiveEncodingAuto,
			err:           ErrArchiveEncodingRequired,
			want:          "needs_encoding",
		},
		{
			name:          "first explicit decode failure still retryable",
			batch:         models.ArchiveImportBatch{EncodingRequestedMode: archiveEncodingAuto},
			requestedMode: archiveEncodingUTF8,
			err:           ErrArchiveEncodingRequired,
			want:          "needs_encoding",
		},
		{
			name:          "second distinct explicit decode failure is final",
			batch:         models.ArchiveImportBatch{EncodingRequestedMode: archiveEncodingUTF8},
			requestedMode: archiveEncodingGBK,
			err:           ErrArchiveEncodingRequired,
			want:          "failed",
		},
		{
			name:          "non encoding failures fail directly",
			requestedMode: archiveEncodingAuto,
			err:           errors.New("boom"),
			want:          "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := archiveBatchFailureStatus(tt.batch, tt.requestedMode, tt.err); got != tt.want {
				t.Fatalf("archiveBatchFailureStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractZipArchiveDecodesGBKPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	archivePath := filepath.Join(root, "gbk.zip")
	extractedDir := filepath.Join(root, "out")
	const decodedName = "第一组/测试视频.txt"
	if err := writeTestZipArchive(archivePath, []testZipArchiveEntry{
		{
			Name:    mustEncodeGBKString(t, decodedName),
			Body:    "demo-body",
			NonUTF8: true,
		},
	}); err != nil {
		t.Fatalf("writeTestZipArchive() error = %v", err)
	}

	mode, err := extractZipArchive(archivePath, extractedDir, archiveEncodingAuto)
	if err != nil {
		t.Fatalf("extractZipArchive() error = %v", err)
	}
	if mode != archiveEncodingGBK {
		t.Fatalf("extractZipArchive() mode = %q, want %q", mode, archiveEncodingGBK)
	}
	data, err := os.ReadFile(filepath.Join(extractedDir, "第一组", "测试视频.txt"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "demo-body" {
		t.Fatalf("ReadFile() body = %q, want %q", string(data), "demo-body")
	}
}

func TestExtractZipArchiveRejectsDecodedPathConflicts(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	archivePath := filepath.Join(root, "conflict.zip")
	extractedDir := filepath.Join(root, "out")
	const decodedName = "第一组/测试视频.txt"
	if err := writeTestZipArchive(archivePath, []testZipArchiveEntry{
		{
			Name: decodedName,
			Body: "utf8-body",
		},
		{
			Name:    mustEncodeGBKString(t, decodedName),
			Body:    "gbk-body",
			NonUTF8: true,
		},
	}); err != nil {
		t.Fatalf("writeTestZipArchive() error = %v", err)
	}

	_, err := extractZipArchive(archivePath, extractedDir, archiveEncodingAuto)
	if !errors.Is(err, ErrArchiveEntryNameConflict) {
		t.Fatalf("extractZipArchive() error = %v, want %v", err, ErrArchiveEntryNameConflict)
	}
}

func TestExtractZipArchiveWithPasswordUsesDecodedGBKPaths(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "password.zip")
	extractedDir := filepath.Join(root, "out")
	fakeBinDir := filepath.Join(root, "bin")
	logPath := filepath.Join(root, "unzip.log")
	const decodedName = "第一组/测试视频.txt"
	if err := os.MkdirAll(fakeBinDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := writeTestZipArchive(archivePath, []testZipArchiveEntry{
		{
			Name:    mustEncodeGBKString(t, decodedName),
			Body:    "encrypted-body",
			NonUTF8: true,
		},
	}); err != nil {
		t.Fatalf("writeTestZipArchive() error = %v", err)
	}
	if err := writeFakeUnzipScript(fakeBinDir, logPath, "encrypted-body"); err != nil {
		t.Fatalf("writeFakeUnzipScript() error = %v", err)
	}
	t.Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_UNZIP_LOG", logPath)

	svc := &ArchiveImportService{}
	mode, err := svc.extractZipArchiveWithPassword(context.Background(), archivePath, extractedDir, "secret", archiveEncodingAuto)
	if err != nil {
		t.Fatalf("extractZipArchiveWithPassword() error = %v", err)
	}
	if mode != archiveEncodingGBK {
		t.Fatalf("extractZipArchiveWithPassword() mode = %q, want %q", mode, archiveEncodingGBK)
	}

	data, err := os.ReadFile(filepath.Join(extractedDir, "第一组", "测试视频.txt"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "encrypted-body" {
		t.Fatalf("extracted file body = %q, want %q", string(data), "encrypted-body")
	}

	rawHex, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile(log) error = %v", err)
	}
	wantHex := fmt.Sprintf("%x", []byte(mustEncodeGBKString(t, decodedName)))
	if strings.TrimSpace(string(rawHex)) != wantHex {
		t.Fatalf("raw unzip entry bytes = %q, want %q", strings.TrimSpace(string(rawHex)), wantHex)
	}
}

func TestPlanZipExtractEntriesDecodesGBKForPasswordPathPlanning(t *testing.T) {
	t.Parallel()

	const decodedName = "第一组/测试视频.txt"
	files := []*zip.File{
		{
			FileHeader: zip.FileHeader{
				Name:    mustEncodeGBKString(t, decodedName),
				NonUTF8: true,
			},
		},
	}

	entries, mode, err := planZipExtractEntries(files, t.TempDir(), archiveEncodingAuto)
	if err != nil {
		t.Fatalf("planZipExtractEntries() error = %v", err)
	}
	if mode != archiveEncodingGBK {
		t.Fatalf("planZipExtractEntries() mode = %q, want %q", mode, archiveEncodingGBK)
	}
	if len(entries) != 1 {
		t.Fatalf("planZipExtractEntries() len = %d, want 1", len(entries))
	}
	if !strings.HasSuffix(entries[0].targetPath, filepath.Join("第一组", "测试视频.txt")) {
		t.Fatalf("planZipExtractEntries() targetPath = %q", entries[0].targetPath)
	}
}

func TestValidateArchiveImportGroupSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		files []models.ArchiveImportFileListItem
		want  string
		err   string
	}{
		{
			name: "accepts image selection",
			files: []models.ArchiveImportFileListItem{
				{MediaKind: "image", Status: "pending"},
			},
			want: "image",
		},
		{
			name: "rejects mixed media kinds",
			files: []models.ArchiveImportFileListItem{
				{MediaKind: "video", Status: "pending"},
				{MediaKind: "image", Status: "pending"},
			},
			err: "group selection cannot mix video and image files",
		},
		{
			name: "rejects frozen files",
			files: []models.ArchiveImportFileListItem{
				{MediaKind: "video", Status: "ready"},
			},
			err: "processed files can no longer move between groups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateArchiveImportGroupSelection(tt.files)
			if tt.err != "" {
				if err == nil || err.Error() != tt.err {
					t.Fatalf("validateArchiveImportGroupSelection() error = %v, want %q", err, tt.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateArchiveImportGroupSelection() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("validateArchiveImportGroupSelection() = %q, want %q", got, tt.want)
			}
		})
	}
}

type testZipArchiveEntry struct {
	Name    string
	Body    string
	NonUTF8 bool
}

func writeTestZipArchive(path string, entries []testZipArchiveEntry) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	for _, entry := range entries {
		header := &zip.FileHeader{
			Name:    entry.Name,
			Method:  zip.Store,
			NonUTF8: entry.NonUTF8,
		}
		w, err := writer.CreateHeader(header)
		if err != nil {
			_ = writer.Close()
			return err
		}
		if _, err := w.Write([]byte(entry.Body)); err != nil {
			_ = writer.Close()
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return file.Close()
}

func mustEncodeGBKString(t *testing.T, value string) string {
	t.Helper()

	encoded, _, err := transform.String(simplifiedchinese.GBK.NewEncoder(), value)
	if err != nil {
		t.Fatalf("transform.String() error = %v", err)
	}
	return encoded
}

func writeFakeUnzipScript(dir, logPath, body string) error {
	script := fmt.Sprintf(`#!/bin/sh
set -eu
if [ "$1" != "-p" ] || [ "$2" != "-qq" ] || [ "$3" != "-P" ]; then
  echo "unexpected unzip args: $*" >&2
  exit 2
fi
if [ "$4" != "secret" ]; then
  echo "unexpected password" >&2
  exit 2
fi
printf '%%s' "$6" | od -An -tx1 | tr -d ' \n' > "$FAKE_UNZIP_LOG"
printf %q
`, body)
	path := filepath.Join(dir, "unzip")
	return os.WriteFile(path, []byte(script), 0o755)
}

func TestArchiveImportApplyGroupToFileHonorsOverrides(t *testing.T) {
	t.Parallel()

	videoCollectionID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	groupVideoCollectionID := uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd")
	imageCollectionID := uuid.MustParse("eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee")
	groupTitle := "分组标题"
	groupDescription := "分组说明"
	groupVideoType := "movie"
	file := models.ArchiveImportFileListItem{
		MediaKind:          "video",
		RelativePath:       "demo/clip.mp4",
		Title:              "旧标题",
		Description:        "文件级说明",
		Tags:               []string{"old"},
		VideoType:          "short",
		VideoCollectionIDs: []uuid.UUID{videoCollectionID},
		ImageCollectionIDs: nil,
	}
	batch := models.ArchiveImportBatch{
		DefaultTitlePrefix:        "批次标题",
		DefaultDescription:        "批次说明",
		DefaultTags:               []string{"batch-tag"},
		DefaultVideoCollectionIDs: []uuid.UUID{videoCollectionID},
	}
	group := models.ArchiveImportGroup{
		MediaKind:          "video",
		Title:              &groupTitle,
		Description:        &groupDescription,
		Tags:               []string{"group-tag"},
		VideoType:          &groupVideoType,
		VideoCollectionIDs: []uuid.UUID{groupVideoCollectionID},
		ImageCollectionIDs: []uuid.UUID{imageCollectionID},
	}
	overrides := archiveImportFieldOverrides{
		Title:              false,
		Description:        true,
		Tags:               false,
		VideoType:          false,
		VideoCollectionIDs: true,
		ImageCollectionIDs: false,
	}

	archiveImportApplyGroupToFile(&file, batch, &group, overrides)

	if file.Title != groupTitle {
		t.Fatalf("file.Title = %q, want %q", file.Title, groupTitle)
	}
	if file.Description != "文件级说明" {
		t.Fatalf("file.Description = %q, want file-level override", file.Description)
	}
	if len(file.Tags) != 1 || file.Tags[0] != "group-tag" {
		t.Fatalf("file.Tags = %#v, want group defaults", file.Tags)
	}
	if file.VideoType != "movie" {
		t.Fatalf("file.VideoType = %q, want %q", file.VideoType, "movie")
	}
	if len(file.VideoCollectionIDs) != 1 || file.VideoCollectionIDs[0] != videoCollectionID {
		t.Fatalf("file.VideoCollectionIDs = %#v, want file-level override", file.VideoCollectionIDs)
	}
	if len(file.ImageCollectionIDs) != 1 || file.ImageCollectionIDs[0] != imageCollectionID {
		t.Fatalf("file.ImageCollectionIDs = %#v, want group default image collection", file.ImageCollectionIDs)
	}
}
