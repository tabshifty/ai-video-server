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

	"video-server/internal/hashutil"
	"video-server/internal/models"
	"video-server/pkg/ffmpeg"
)

type fakeCloudImportTaskRepo struct {
	task models.CloudDownloadTask
	err  error
}

func (f *fakeCloudImportTaskRepo) GetByID(context.Context, int64) (models.CloudDownloadTask, error) {
	if f.err != nil {
		return models.CloudDownloadTask{}, f.err
	}
	if f.task.ID == 0 {
		return models.CloudDownloadTask{}, pgx.ErrNoRows
	}
	return f.task, nil
}

type cloudImportFinishCall struct {
	status     string
	videoID    *uuid.UUID
	batchID    *uuid.UUID
	importedAt time.Time
}

type fakeCloudImportRepo struct {
	record       models.CloudDownloadImport
	createErr    error
	getErr       error
	updateErr    error
	finishErr    error
	resetErr     error
	resetWon     bool
	resetCalls   int
	statuses     []string
	messages     []string
	finishCalls  []cloudImportFinishCall
	createdCount int
	sha256       string
}

func (f *fakeCloudImportRepo) CreateOrGet(_ context.Context, item models.CloudDownloadImport) (models.CloudDownloadImport, bool, error) {
	if f.createErr != nil {
		return models.CloudDownloadImport{}, false, f.createErr
	}
	if f.record.ID != 0 {
		return f.record, false, nil
	}
	item.ID = 1
	if item.Status == "" {
		item.Status = models.CloudImportQueued
	}
	f.record = item
	f.createdCount++
	return f.record, true, nil
}

func (f *fakeCloudImportRepo) GetByID(context.Context, int64) (models.CloudDownloadImport, error) {
	if f.getErr != nil {
		return models.CloudDownloadImport{}, f.getErr
	}
	if f.record.ID == 0 {
		return models.CloudDownloadImport{}, pgx.ErrNoRows
	}
	return f.record, nil
}

func (f *fakeCloudImportRepo) GetByTaskID(context.Context, int64) (models.CloudDownloadImport, error) {
	if f.getErr != nil {
		return models.CloudDownloadImport{}, f.getErr
	}
	if f.record.ID == 0 {
		return models.CloudDownloadImport{}, pgx.ErrNoRows
	}
	return f.record, nil
}

func (f *fakeCloudImportRepo) UpdateProgress(_ context.Context, _ int64, status, message string) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.statuses = append(f.statuses, status)
	f.messages = append(f.messages, message)
	f.record.Status = status
	f.record.ErrorMessage = message
	return nil
}

func (f *fakeCloudImportRepo) SetHash(_ context.Context, _ int64, sha256 string) error {
	f.sha256 = sha256
	return nil
}

func (f *fakeCloudImportRepo) SetVideo(_ context.Context, _ int64, videoID uuid.UUID) error {
	f.record.VideoID = &videoID
	return nil
}

func (f *fakeCloudImportRepo) SetArchiveBatch(_ context.Context, _ int64, batchID uuid.UUID) error {
	f.record.ArchiveBatchID = &batchID
	return nil
}

func (f *fakeCloudImportRepo) Finish(_ context.Context, _ int64, status string, videoID, batchID *uuid.UUID, importedAt time.Time) error {
	if f.finishErr != nil {
		return f.finishErr
	}
	f.finishCalls = append(f.finishCalls, cloudImportFinishCall{status: status, videoID: videoID, batchID: batchID, importedAt: importedAt})
	f.record.Status = status
	if videoID != nil {
		f.record.VideoID = videoID
	}
	if batchID != nil {
		f.record.ArchiveBatchID = batchID
	}
	f.record.ImportedAt = &importedAt
	return nil
}

func (f *fakeCloudImportRepo) ResetForRetry(context.Context, int64) (models.CloudDownloadImport, bool, error) {
	if f.resetErr != nil {
		return models.CloudDownloadImport{}, false, f.resetErr
	}
	f.resetCalls++
	if !f.resetWon {
		return f.record, false, nil
	}
	f.record.Attempt++
	f.record.Status = models.CloudImportQueued
	f.record.ErrorMessage = ""
	return f.record, true, nil
}

func (f *fakeCloudImportRepo) ListSummariesByTaskIDs(context.Context, []int64) (map[int64]models.CloudDownloadImportSummary, error) {
	return map[int64]models.CloudDownloadImportSummary{}, nil
}

type fakeCloudImportVideoLookup struct {
	videoID  uuid.UUID
	found    bool
	err      error
	video    models.Video
	videoErr error
}

func (f *fakeCloudImportVideoLookup) FindVideoByHash(context.Context, string, int64) (uuid.UUID, bool, error) {
	return f.videoID, f.found, f.err
}

func (f *fakeCloudImportVideoLookup) GetVideoByID(context.Context, uuid.UUID) (models.Video, error) {
	if f.videoErr != nil {
		return models.Video{}, f.videoErr
	}
	return f.video, nil
}

type fakeCloudImportUpload struct {
	previewResult UploadPreview
	previewErr    error
	saveResult    UploadResult
	saveErr       error
	previewInputs []LocalUploadInput
	saveInputs    []LocalUploadInput
}

func (f *fakeCloudImportUpload) PreviewUploadedFile(_ context.Context, in LocalUploadInput) (UploadPreview, error) {
	f.previewInputs = append(f.previewInputs, in)
	return f.previewResult, f.previewErr
}

func (f *fakeCloudImportUpload) SaveImportedFile(_ context.Context, in LocalUploadInput, _ int64) (UploadResult, error) {
	f.saveInputs = append(f.saveInputs, in)
	return f.saveResult, f.saveErr
}

type fakeCloudImportArchive struct {
	batch      models.ArchiveImportBatch
	err        error
	gotInput   ArchiveImportUploadInput
	gotSource  string
	calls      int
	retryCalls int
	retryID    uuid.UUID
}

func (f *fakeCloudImportArchive) ImportLocalArchive(_ context.Context, in ArchiveImportUploadInput, sourcePath string) (models.ArchiveImportBatch, error) {
	f.calls++
	f.gotInput = in
	f.gotSource = sourcePath
	return f.batch, f.err
}

func (f *fakeCloudImportArchive) GetBatch(context.Context, uuid.UUID) (models.ArchiveImportBatch, error) {
	return f.batch, f.err
}

func (f *fakeCloudImportArchive) RetryExtract(_ context.Context, batchID uuid.UUID, _, _ string) (models.ArchiveImportBatch, error) {
	f.retryCalls++
	f.retryID = batchID
	return f.batch, f.err
}

type cloudIngestCall struct {
	importID int64
	attempt  int
	input    models.CloudDownloadImportInput
}

type fakeCloudImportEnqueuer struct {
	ingestCalls []cloudIngestCall
	ingestErr   error

	transcodeCalls int
	transcodeErr   error
	transcodePath  string
	transcodeVideo string

	scrapeCalls int
	scrapeErr   error
	scrapeType  string
}

func (f *fakeCloudImportEnqueuer) EnqueueCloudIngest(importID int64, attempt int, input models.CloudDownloadImportInput) error {
	f.ingestCalls = append(f.ingestCalls, cloudIngestCall{importID: importID, attempt: attempt, input: input})
	return f.ingestErr
}

func (f *fakeCloudImportEnqueuer) EnqueueTranscode(videoID, inputPath, outputDir, targetFormat string, force bool) error {
	f.transcodeCalls++
	f.transcodePath = inputPath
	f.transcodeVideo = videoID
	return f.transcodeErr
}

func (f *fakeCloudImportEnqueuer) EnqueueScrapeMovie(videoID, filePath, filename string) error {
	f.scrapeCalls++
	f.scrapeType = "movie"
	return f.scrapeErr
}

func (f *fakeCloudImportEnqueuer) EnqueueScrapeTV(videoID, filePath, filename string) error {
	f.scrapeCalls++
	f.scrapeType = "episode"
	return f.scrapeErr
}

func (f *fakeCloudImportEnqueuer) EnqueueScrapeAV(videoID, filePath, filename string) error {
	f.scrapeCalls++
	f.scrapeType = "av"
	return f.scrapeErr
}

type cloudImportHarness struct {
	svc      *CloudDownloadImportService
	tasks    *fakeCloudImportTaskRepo
	repo     *fakeCloudImportRepo
	videos   *fakeCloudImportVideoLookup
	upload   *fakeCloudImportUpload
	archive  *fakeCloudImportArchive
	enqueuer *fakeCloudImportEnqueuer
	root     string
	source   string
}

func newCloudImportHarness(t *testing.T) *cloudImportHarness {
	t.Helper()
	root := t.TempDir()
	downloadDir := filepath.Join(root, "downloads")
	storageRoot := filepath.Join(root, "storage")
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		t.Fatalf("create download dir: %v", err)
	}
	// macOS t.TempDir() lives under /var, a symlink to /private/var; the
	// service resolves symlinks before comparing containment.
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
		downloadDir = filepath.Join(root, "downloads")
		storageRoot = filepath.Join(root, "storage")
	}
	source := filepath.Join(downloadDir, "demo.mkv")
	if err := os.WriteFile(source, []byte("video-body"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	h := &cloudImportHarness{
		tasks:    &fakeCloudImportTaskRepo{task: models.CloudDownloadTask{ID: 7, Status: models.CloudDownloadCompleted, LocalPath: source, InfoHash: "hash-1"}},
		repo:     &fakeCloudImportRepo{},
		videos:   &fakeCloudImportVideoLookup{},
		upload:   &fakeCloudImportUpload{},
		archive:  &fakeCloudImportArchive{},
		enqueuer: &fakeCloudImportEnqueuer{},
		root:     root,
		source:   source,
	}
	h.svc = NewCloudDownloadImportService(h.tasks, h.repo, h.videos, h.upload, h.archive, h.enqueuer, downloadDir, storageRoot, nil)
	h.svc.probePlayback = func(context.Context, string) (ffmpeg.PlaybackCompatibilityProbe, error) {
		return ffmpeg.PlaybackCompatibilityProbe{VideoStreamFound: true}, nil
	}
	return h
}

func (h *cloudImportHarness) archiveSource(t *testing.T) string {
	t.Helper()
	path := filepath.Join(filepath.Dir(h.source), "demo.zip")
	if err := writeTestZipArchive(path, []testZipArchiveEntry{{Name: "inside.txt", Body: "zip-body"}}); err != nil {
		t.Fatalf("write zip: %v", err)
	}
	h.source = path
	h.tasks.task.LocalPath = path
	return path
}

func (h *cloudImportHarness) resetSource(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(filepath.Dir(h.source), name)
	if err := os.WriteFile(path, []byte("body"), 0o644); err != nil {
		t.Fatalf("write source %s: %v", name, err)
	}
	h.source = path
	h.tasks.task.LocalPath = path
	return path
}

func TestCloudImportPreviewRejectsMissingAndNotCompleted(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)

	h.tasks.task.LocalPath = filepath.Join(filepath.Dir(h.source), "missing.mkv")
	if _, err := h.svc.Preview(context.Background(), 7); !errors.Is(err, models.ErrCloudImportFileMissing) {
		t.Fatalf("missing file error = %v, want ErrCloudImportFileMissing", err)
	}

	h.tasks.task.LocalPath = h.source
	h.tasks.task.Status = models.CloudDownloadDownloading
	if _, err := h.svc.Preview(context.Background(), 7); !errors.Is(err, models.ErrCloudImportNotCompleted) {
		t.Fatalf("not completed error = %v, want ErrCloudImportNotCompleted", err)
	}

	h.tasks.task.Status = models.CloudDownloadCompleted
	outside := filepath.Join(h.root, "outside.mkv")
	if err := os.WriteFile(outside, []byte("x"), 0o644); err != nil {
		t.Fatalf("write outside: %v", err)
	}
	h.tasks.task.LocalPath = outside
	if _, err := h.svc.Preview(context.Background(), 7); !errors.Is(err, models.ErrCloudImportFileMissing) {
		t.Fatalf("outside dir error = %v, want ErrCloudImportFileMissing", err)
	}
}

func TestCloudImportPreviewDetectsArchiveAndVideo(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	archivePath := h.resetSource(t, "demo.zip")
	if err := writeTestZipArchive(archivePath, []testZipArchiveEntry{{Name: "inside.txt", Body: "zip-body"}}); err != nil {
		t.Fatalf("write zip: %v", err)
	}
	preview, err := h.svc.Preview(context.Background(), 7)
	if err != nil {
		t.Fatalf("archive preview error = %v", err)
	}
	if preview.Kind != models.CloudImportKindArchive || preview.Name != "demo.zip" {
		t.Fatalf("archive preview = %+v", preview)
	}
	info, err := os.Stat(archivePath)
	if err != nil {
		t.Fatalf("stat archive: %v", err)
	}
	if preview.Size != info.Size() {
		t.Fatalf("archive size = %d, want %d", preview.Size, info.Size())
	}
	if preview.Mime == "" {
		t.Fatal("expected a mime type")
	}

	h.resetSource(t, "demo.mkv")
	preview, err = h.svc.Preview(context.Background(), 7)
	if err != nil {
		t.Fatalf("video preview error = %v", err)
	}
	if preview.Kind != models.CloudImportKindVideo || preview.Name != "demo.mkv" {
		t.Fatalf("video preview = %+v", preview)
	}
}

func TestCloudImportPreviewUsesContentNotExtension(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	// A video file with a .zip name must be classified as video by probing,
	// not as an archive by its extension.
	h.resetSource(t, "actually-video.zip")
	preview, err := h.svc.Preview(context.Background(), 7)
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	if preview.Kind != models.CloudImportKindVideo {
		t.Fatalf("kind = %s, want video", preview.Kind)
	}

	// The reverse: an archive name/extension with non-archive content is not
	// importable as an archive.
	videoID := uuid.New()
	h.upload.previewResult = UploadPreview{VideoID: videoID}
	h.repo.record = models.CloudDownloadImport{ID: 21, TaskID: 7, Kind: models.CloudImportKindArchive, Status: models.CloudImportQueued}
	if err := h.svc.Process(context.Background(), 21, 0, models.CloudDownloadImportInput{Kind: models.CloudImportKindArchive}); err == nil {
		t.Fatal("expected kind mismatch to fail")
	}
	if h.repo.record.Status != models.CloudImportFailed {
		t.Fatalf("status = %s, want failed", h.repo.record.Status)
	}
}

func TestCloudImportPreviewRejectsUnsupportedFile(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	h.svc.probePlayback = func(context.Context, string) (ffmpeg.PlaybackCompatibilityProbe, error) {
		return ffmpeg.PlaybackCompatibilityProbe{}, nil
	}
	if _, err := h.svc.Preview(context.Background(), 7); !errors.Is(err, models.ErrCloudImportUnsupported) {
		t.Fatalf("no video stream error = %v, want ErrCloudImportUnsupported", err)
	}

	h.svc.probePlayback = func(context.Context, string) (ffmpeg.PlaybackCompatibilityProbe, error) {
		return ffmpeg.PlaybackCompatibilityProbe{}, errors.New("ffprobe failed")
	}
	if _, err := h.svc.Preview(context.Background(), 7); !errors.Is(err, models.ErrCloudImportUnsupported) {
		t.Fatalf("probe failure error = %v, want ErrCloudImportUnsupported", err)
	}
}

func TestCloudImportCreateEnqueuesOnce(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	created, err := h.svc.Create(context.Background(), 7, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}, nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID != 1 || created.Status != models.CloudImportQueued || created.Kind != models.CloudImportKindVideo {
		t.Fatalf("created = %+v", created)
	}
	if len(h.enqueuer.ingestCalls) != 1 {
		t.Fatalf("ingest calls = %d, want 1", len(h.enqueuer.ingestCalls))
	}
	if h.enqueuer.ingestCalls[0].attempt != 0 || h.enqueuer.ingestCalls[0].input.Type != "short" {
		t.Fatalf("ingest call = %+v", h.enqueuer.ingestCalls[0])
	}
	if h.repo.record.SourcePath != h.source || h.repo.record.SourceSize == 0 {
		t.Fatalf("record source = %+v", h.repo.record)
	}

	h.repo.record.Status = models.CloudImportImported
	again, err := h.svc.Create(context.Background(), 7, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}, nil)
	if err != nil {
		t.Fatalf("second Create() error = %v", err)
	}
	if again.Status != models.CloudImportImported {
		t.Fatalf("second status = %s", again.Status)
	}
	if len(h.enqueuer.ingestCalls) != 1 {
		t.Fatalf("terminal import enqueued again: %d calls", len(h.enqueuer.ingestCalls))
	}
}

func TestCloudImportCreateRequeuesActiveRecordWithSameAttempt(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	h.repo.record = models.CloudDownloadImport{
		ID:         52,
		TaskID:     7,
		Kind:       models.CloudImportKindVideo,
		Status:     models.CloudImportQueued,
		SourcePath: h.source,
		SourceSize: 10,
	}

	record, err := h.svc.Create(context.Background(), 7, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}, nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if record.Status != models.CloudImportQueued || record.Attempt != 0 {
		t.Fatalf("record = %+v, want queued attempt 0", record)
	}
	if h.repo.resetCalls != 0 {
		t.Fatalf("active re-enqueue must not reset attempt: %d", h.repo.resetCalls)
	}
	if len(h.enqueuer.ingestCalls) != 1 || h.enqueuer.ingestCalls[0].attempt != 0 {
		t.Fatalf("ingest calls = %+v", h.enqueuer.ingestCalls)
	}
}

func TestCloudImportCreateReturnsTerminalRecordWithoutSource(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	videoID := uuid.New()
	h.repo.record = models.CloudDownloadImport{ID: 51, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportImported, VideoID: &videoID}
	if err := os.Remove(h.source); err != nil {
		t.Fatalf("remove source: %v", err)
	}

	record, err := h.svc.Create(context.Background(), 7, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}, nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if record.Status != models.CloudImportImported || record.VideoID == nil || *record.VideoID != videoID {
		t.Fatalf("record = %+v", record)
	}
	if len(h.enqueuer.ingestCalls) != 0 {
		t.Fatalf("terminal record enqueued again")
	}
	if len(h.upload.previewInputs) != 0 || len(h.upload.saveInputs) != 0 {
		t.Fatal("terminal record must not inspect/import the file again")
	}
}

func TestCloudImportCreateRetriesFailedRecord(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	h.repo.record = models.CloudDownloadImport{ID: 5, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportFailed, Attempt: 1}
	h.repo.resetWon = true

	updated, err := h.svc.Create(context.Background(), 7, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}, nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if updated.Attempt != 2 || updated.Status != models.CloudImportQueued {
		t.Fatalf("updated = %+v", updated)
	}
	if len(h.enqueuer.ingestCalls) != 1 || h.enqueuer.ingestCalls[0].attempt != 2 {
		t.Fatalf("ingest calls = %+v", h.enqueuer.ingestCalls)
	}
}

func TestCloudImportCreateEnqueueFailureMarksFailed(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	h.enqueuer.ingestErr = errors.New("redis down")

	record, err := h.svc.Create(context.Background(), 7, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}, nil)
	if err == nil {
		t.Fatal("expected enqueue error")
	}
	if record.Status != models.CloudImportFailed {
		t.Fatalf("record status = %s, want failed", record.Status)
	}
	if h.repo.record.Status != models.CloudImportFailed {
		t.Fatalf("persisted status = %s, want failed", h.repo.record.Status)
	}
}

func TestCloudImportCreateRejectsKindMismatchAndBadInput(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	if _, err := h.svc.Create(context.Background(), 7, models.CloudDownloadImportInput{Kind: models.CloudImportKindArchive}, nil); !errors.Is(err, models.ErrCloudImportUnsupported) {
		t.Fatalf("archive ext error = %v, want ErrCloudImportUnsupported", err)
	}
	if _, err := h.svc.Create(context.Background(), 7, models.CloudDownloadImportInput{Kind: "image"}, nil); !errors.Is(err, models.ErrCloudImportInvalidInput) {
		t.Fatalf("kind error = %v, want ErrCloudImportInvalidInput", err)
	}
	if _, err := h.svc.Create(context.Background(), 7, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "telegram"}, nil); !errors.Is(err, models.ErrCloudImportInvalidInput) {
		t.Fatalf("type error = %v, want ErrCloudImportInvalidInput", err)
	}
}

func TestCloudImportProcessVideoSkipsHashHitAndDeletesSource(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	existingVideoID := uuid.New()
	// FindVideoByHash reads file_hashes without a status filter, so a
	// pending_delete video is a valid hash hit too.
	h.videos.found = true
	h.videos.videoID = existingVideoID
	h.repo.record = models.CloudDownloadImport{ID: 3, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportQueued}

	err := h.svc.Process(context.Background(), 3, 0, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if h.repo.record.Status != models.CloudImportSkippedExisting {
		t.Fatalf("status = %s, want skipped_existing", h.repo.record.Status)
	}
	if h.repo.record.VideoID == nil || *h.repo.record.VideoID != existingVideoID {
		t.Fatalf("video id = %v, want %s", h.repo.record.VideoID, existingVideoID)
	}
	if len(h.upload.previewInputs) != 0 || len(h.upload.saveInputs) != 0 {
		t.Fatal("hash hit must not copy or create a video")
	}
	if _, statErr := os.Stat(h.source); !os.IsNotExist(statErr) {
		t.Fatalf("download source still exists, stat err = %v", statErr)
	}
}

func TestCloudImportProcessVideoResumesUnlinkedVideoFromSameTask(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	videoID := uuid.New()
	targetPath := filepath.Join(h.root, "storage", "videos", videoID.String(), "source-original.mkv")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("create target dir: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("video-body"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	h.videos.found = true
	h.videos.videoID = videoID
	h.videos.video = models.Video{
		ID:           videoID,
		Type:         "short",
		Status:       "uploaded",
		OriginalPath: targetPath,
		Metadata:     []byte(`{"source":"cloud_download","cloud_task_id":7}`),
	}
	h.repo.record = models.CloudDownloadImport{ID: 54, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportQueued}

	if err := h.svc.Process(context.Background(), 54, 0, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if h.repo.record.Status != models.CloudImportImported {
		t.Fatalf("status = %s, want imported", h.repo.record.Status)
	}
	if h.repo.record.VideoID == nil || *h.repo.record.VideoID != videoID {
		t.Fatalf("video id = %v, want %s", h.repo.record.VideoID, videoID)
	}
	if h.enqueuer.transcodeCalls != 1 {
		t.Fatalf("transcode calls = %d, want 1", h.enqueuer.transcodeCalls)
	}
	if _, statErr := os.Stat(h.source); !os.IsNotExist(statErr) {
		t.Fatalf("download source still exists, stat err = %v", statErr)
	}
}

func TestCloudImportProcessVideoCreatesCopiesAndEnqueues(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	videoID := uuid.New()
	h.upload.previewResult = UploadPreview{VideoID: videoID}
	h.upload.saveResult = UploadResult{VideoID: videoID, Status: "uploaded", Enqueue: true}
	h.repo.record = models.CloudDownloadImport{ID: 4, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportQueued}

	input := models.CloudDownloadImportInput{
		Kind:          models.CloudImportKindVideo,
		Type:          "short",
		Title:         "标题",
		Description:   "描述",
		Tags:          []string{"tag"},
		CollectionIDs: []uuid.UUID{uuid.New()},
	}
	expectedHash, err := hashutil.SHA256(h.source)
	if err != nil {
		t.Fatalf("hash source: %v", err)
	}
	if err := h.svc.Process(context.Background(), 4, 0, input); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if h.repo.record.Status != models.CloudImportImported {
		t.Fatalf("status = %s, want imported", h.repo.record.Status)
	}
	if h.repo.record.VideoID == nil || *h.repo.record.VideoID != videoID {
		t.Fatalf("video id = %v", h.repo.record.VideoID)
	}
	if h.repo.sha256 != expectedHash {
		t.Fatalf("persisted sha256 = %q, want %q", h.repo.sha256, expectedHash)
	}
	if len(h.upload.saveInputs) != 1 {
		t.Fatalf("save calls = %d, want 1", len(h.upload.saveInputs))
	}
	saved := h.upload.saveInputs[0]
	targetPath := filepath.Join(h.root, "storage", "videos", videoID.String(), "source-original.mkv")
	if saved.FilePath != targetPath {
		t.Fatalf("save path = %s, want %s", saved.FilePath, targetPath)
	}
	if saved.VideoID != videoID || !saved.HashVerified || saved.Hash == "" {
		t.Fatalf("save input = %+v", saved)
	}
	if saved.Metadata["source"] != "cloud_download" || saved.Metadata["delete_source_after_transcode"] != true {
		t.Fatalf("metadata = %+v", saved.Metadata)
	}
	if saved.Metadata["cloud_task_id"] != int64(7) || saved.Metadata["info_hash"] != "hash-1" {
		t.Fatalf("trace metadata = %+v", saved.Metadata)
	}
	if _, statErr := os.Stat(targetPath); statErr != nil {
		t.Fatalf("copied target missing: %v", statErr)
	}
	if _, statErr := os.Stat(h.source); !os.IsNotExist(statErr) {
		t.Fatalf("download source still exists")
	}
	if h.enqueuer.transcodeCalls != 1 || h.enqueuer.transcodePath != targetPath {
		t.Fatalf("transcode calls = %d path = %s", h.enqueuer.transcodeCalls, h.enqueuer.transcodePath)
	}
	if len(h.repo.finishCalls) != 1 || h.repo.finishCalls[0].status != models.CloudImportImported {
		t.Fatalf("finish calls = %+v", h.repo.finishCalls)
	}
}

func TestCloudImportProcessVideoEnqueueFailureKeepsTargetAndResumes(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	videoID := uuid.New()
	h.upload.previewResult = UploadPreview{VideoID: videoID}
	h.upload.saveResult = UploadResult{VideoID: videoID, Status: "uploaded"}
	h.enqueuer.transcodeErr = errors.New("queue down")
	h.repo.record = models.CloudDownloadImport{ID: 6, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportQueued}
	input := models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}

	if err := h.svc.Process(context.Background(), 6, 0, input); err == nil {
		t.Fatal("expected enqueue error")
	}
	targetPath := filepath.Join(h.root, "storage", "videos", videoID.String(), "source-original.mkv")
	if h.repo.record.Status != models.CloudImportFailed {
		t.Fatalf("status = %s, want failed", h.repo.record.Status)
	}
	if _, statErr := os.Stat(targetPath); statErr != nil {
		t.Fatalf("target should be kept for retry: %v", statErr)
	}
	if _, statErr := os.Stat(h.source); statErr != nil {
		t.Fatalf("download source should be kept for retry: %v", statErr)
	}

	// Retry: the video row now exists, so Process resumes without copying or
	// creating it again.
	h.enqueuer.transcodeErr = nil
	h.videos.found = true
	h.videos.videoID = videoID
	h.videos.video = models.Video{ID: videoID, Status: "processing", Type: "short"}
	if err := h.svc.Process(context.Background(), 6, 0, input); err != nil {
		t.Fatalf("retry Process() error = %v", err)
	}
	if h.repo.record.Status != models.CloudImportImported {
		t.Fatalf("retry status = %s, want imported", h.repo.record.Status)
	}
	if len(h.upload.saveInputs) != 1 {
		t.Fatalf("save calls = %d, want 1 (retry must reuse target)", len(h.upload.saveInputs))
	}
	if h.enqueuer.transcodeCalls != 2 {
		t.Fatalf("transcode calls = %d, want 2", h.enqueuer.transcodeCalls)
	}
	if _, statErr := os.Stat(h.source); !os.IsNotExist(statErr) {
		t.Fatalf("download source should be deleted after retry success")
	}
	if _, statErr := os.Stat(targetPath); statErr != nil {
		t.Fatalf("source-original must survive the retry: %v", statErr)
	}
}

func TestCloudImportResumeUsesPersistedVideoType(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	videoID := uuid.New()
	targetDir := filepath.Join(h.root, "storage", "videos", videoID.String())
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}
	targetPath := filepath.Join(targetDir, "source-original.mkv")
	if err := os.WriteFile(targetPath, []byte("video-body"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	h.videos.found = true
	h.videos.videoID = videoID
	h.videos.video = models.Video{ID: videoID, Status: "scraping", Type: "movie"}
	h.repo.record = models.CloudDownloadImport{ID: 41, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportFailed, VideoID: &videoID}

	// The retry form says short, but the persisted video row says movie; the
	// movie scrape queue must win.
	if err := h.svc.Process(context.Background(), 41, 0, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if h.enqueuer.transcodeCalls != 0 {
		t.Fatalf("transcode calls = %d, want 0", h.enqueuer.transcodeCalls)
	}
	if h.enqueuer.scrapeCalls != 1 || h.enqueuer.scrapeType != "movie" {
		t.Fatalf("scrape calls = %d type = %q", h.enqueuer.scrapeCalls, h.enqueuer.scrapeType)
	}
	if h.repo.record.Status != models.CloudImportImported {
		t.Fatalf("status = %s, want imported", h.repo.record.Status)
	}
}

func TestCloudImportCreateRejectsActualKindMismatch(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	zipPath := h.resetSource(t, "demo.zip")
	if err := writeTestZipArchive(zipPath, []testZipArchiveEntry{{Name: "inside.txt", Body: "zip-body"}}); err != nil {
		t.Fatalf("write zip: %v", err)
	}
	if _, err := h.svc.Create(context.Background(), 7, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}, nil); !errors.Is(err, models.ErrCloudImportInvalidInput) {
		t.Fatalf("Create() error = %v, want ErrCloudImportInvalidInput", err)
	}
	if h.repo.record.ID != 0 {
		t.Fatalf("mismatched create must not persist a record: %+v", h.repo.record)
	}
}

func TestCloudImportProcessVideoSaveFailureRollsBackCopy(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	videoID := uuid.New()
	h.upload.previewResult = UploadPreview{VideoID: videoID}
	h.upload.saveErr = errors.New("db write failed")
	h.repo.record = models.CloudDownloadImport{ID: 42, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportQueued}

	if err := h.svc.Process(context.Background(), 42, 0, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}); err == nil {
		t.Fatal("expected save error")
	}
	if h.repo.record.Status != models.CloudImportFailed {
		t.Fatalf("status = %s, want failed", h.repo.record.Status)
	}
	targetPath := filepath.Join(h.root, "storage", "videos", videoID.String(), "source-original.mkv")
	if _, statErr := os.Stat(targetPath); !os.IsNotExist(statErr) {
		t.Fatalf("copied target should be rolled back, stat err = %v", statErr)
	}
	if _, statErr := os.Stat(h.source); statErr != nil {
		t.Fatalf("download source should be kept: %v", statErr)
	}
	if h.enqueuer.transcodeCalls != 0 {
		t.Fatalf("transcode calls = %d, want 0", h.enqueuer.transcodeCalls)
	}
}

func TestCloudImportProcessArchiveCreatesBatchAndDeletesSource(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	batchID := uuid.New()
	h.archive.batch = models.ArchiveImportBatch{ID: batchID, Status: "extracting"}
	h.archiveSource(t)
	h.repo.record = models.CloudDownloadImport{ID: 8, TaskID: 7, Kind: models.CloudImportKindArchive, Status: models.CloudImportQueued}
	input := models.CloudDownloadImportInput{
		Kind:                      models.CloudImportKindArchive,
		Title:                     "压缩包",
		DefaultDescription:        "说明",
		DefaultTags:               []string{"tag"},
		DefaultVideoCollectionIDs: []uuid.UUID{uuid.New()},
		DefaultImageCollectionIDs: []uuid.UUID{uuid.New()},
		HasPassword:               true,
		Password:                  "secret",
	}

	if err := h.svc.Process(context.Background(), 8, 0, input); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if h.archive.calls != 1 || h.archive.gotSource != h.source {
		t.Fatalf("archive calls = %d source = %s", h.archive.calls, h.archive.gotSource)
	}
	if !h.archive.gotInput.HasPassword || h.archive.gotInput.Password != "secret" || h.archive.gotInput.Title != "压缩包" {
		t.Fatalf("archive input = %+v", h.archive.gotInput)
	}
	if h.repo.record.Status != models.CloudImportImported {
		t.Fatalf("status = %s, want imported", h.repo.record.Status)
	}
	if h.repo.record.ArchiveBatchID == nil || *h.repo.record.ArchiveBatchID != batchID {
		t.Fatalf("batch id = %v, want %s", h.repo.record.ArchiveBatchID, batchID)
	}
	if _, statErr := os.Stat(h.source); !os.IsNotExist(statErr) {
		t.Fatalf("download source still exists")
	}
}

func TestCloudImportProcessArchiveFailureKeepsDownload(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	batchID := uuid.New()
	h.archive.batch = models.ArchiveImportBatch{ID: batchID, Status: "failed"}
	h.archive.err = errors.New("enqueue failed")
	h.archiveSource(t)
	h.repo.record = models.CloudDownloadImport{ID: 9, TaskID: 7, Kind: models.CloudImportKindArchive, Status: models.CloudImportQueued}

	err := h.svc.Process(context.Background(), 9, 0, models.CloudDownloadImportInput{Kind: models.CloudImportKindArchive})
	if err == nil {
		t.Fatal("expected archive error")
	}
	if h.repo.record.Status != models.CloudImportFailed {
		t.Fatalf("status = %s, want failed", h.repo.record.Status)
	}
	if h.repo.record.ArchiveBatchID == nil || *h.repo.record.ArchiveBatchID != batchID {
		t.Fatalf("batch id = %v, want %s", h.repo.record.ArchiveBatchID, batchID)
	}
	if _, statErr := os.Stat(h.source); statErr != nil {
		t.Fatalf("download source should be kept: %v", statErr)
	}
}

func TestCloudImportProcessRejectsNonCompletedTask(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	h.tasks.task.Status = models.CloudDownloadReady
	h.repo.record = models.CloudDownloadImport{ID: 11, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportQueued}

	if err := h.svc.Process(context.Background(), 11, 0, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}); err == nil {
		t.Fatal("expected not-completed error")
	}
	if h.repo.record.Status != models.CloudImportFailed {
		t.Fatalf("status = %s, want failed", h.repo.record.Status)
	}
}

func TestCloudImportProcessRejectsActualKindMismatch(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	h.svc.probePlayback = func(context.Context, string) (ffmpeg.PlaybackCompatibilityProbe, error) {
		return ffmpeg.PlaybackCompatibilityProbe{}, nil
	}
	h.repo.record = models.CloudDownloadImport{ID: 33, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportQueued}

	if err := h.svc.Process(context.Background(), 33, 0, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}); err == nil {
		t.Fatal("expected unsupported file error")
	}
	if h.repo.record.Status != models.CloudImportFailed {
		t.Fatalf("status = %s, want failed", h.repo.record.Status)
	}
	if len(h.upload.saveInputs) != 0 {
		t.Fatalf("save calls = %d, want 0", len(h.upload.saveInputs))
	}
}

func TestIsSupportedArchiveFileRequiresExtensionAndMIME(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		mime string
		want bool
	}{
		{name: "demo.zip", mime: "application/zip", want: true},
		{name: "demo.rar", mime: "application/vnd.rar", want: true},
		{name: "demo.7z", mime: "application/x-7z-compressed", want: true},
		{name: "demo.zip", mime: "application/octet-stream", want: true},
		{name: "demo.zip", mime: "video/x-matroska", want: false},
		{name: "demo.mkv", mime: "application/zip", want: false},
		{name: "demo.txt", mime: "text/plain", want: false},
	}
	for _, tt := range cases {
		if got := isSupportedArchiveFile(tt.name, tt.mime); got != tt.want {
			t.Fatalf("isSupportedArchiveFile(%q, %q) = %v, want %v", tt.name, tt.mime, got, tt.want)
		}
	}
}

func TestCloudImportProcessArchiveRetryReusesExistingBatch(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	batchID := uuid.New()
	h.archive.batch = models.ArchiveImportBatch{ID: batchID, Status: "failed"}
	h.archiveSource(t)
	h.repo.record = models.CloudDownloadImport{ID: 31, TaskID: 7, Kind: models.CloudImportKindArchive, Status: models.CloudImportFailed, ArchiveBatchID: &batchID}

	if err := h.svc.Process(context.Background(), 31, 0, models.CloudDownloadImportInput{Kind: models.CloudImportKindArchive, Password: "secret"}); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if h.archive.calls != 0 {
		t.Fatalf("ImportLocalArchive calls = %d, want 0", h.archive.calls)
	}
	if h.archive.retryCalls != 1 || h.archive.retryID != batchID {
		t.Fatalf("retry calls = %d id = %s", h.archive.retryCalls, h.archive.retryID)
	}
	if h.repo.record.Status != models.CloudImportImported {
		t.Fatalf("status = %s, want imported", h.repo.record.Status)
	}
	if _, statErr := os.Stat(h.source); !os.IsNotExist(statErr) {
		t.Fatalf("download source should be deleted")
	}
}

func TestCloudImportProcessArchiveDoesNotClearQueuedBatch(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	batchID := uuid.New()
	h.archive.batch = models.ArchiveImportBatch{ID: batchID, Status: "extracting"}
	h.archiveSource(t)
	h.repo.record = models.CloudDownloadImport{ID: 34, TaskID: 7, Kind: models.CloudImportKindArchive, Status: models.CloudImportFailed, ArchiveBatchID: &batchID}

	if err := h.svc.Process(context.Background(), 34, 0, models.CloudDownloadImportInput{Kind: models.CloudImportKindArchive, Password: "secret"}); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if h.archive.retryCalls != 0 {
		t.Fatalf("retry calls = %d, want 0 for a queued/extracting batch", h.archive.retryCalls)
	}
	if h.repo.record.Status != models.CloudImportImported {
		t.Fatalf("status = %s, want imported", h.repo.record.Status)
	}
	if _, statErr := os.Stat(h.source); !os.IsNotExist(statErr) {
		t.Fatalf("download source should be deleted")
	}
}

func TestCloudImportProcessVideoReadyTargetRemovesOrphanSource(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	videoID := uuid.New()
	targetDir := filepath.Join(h.root, "storage", "videos", videoID.String())
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}
	orphanPath := filepath.Join(targetDir, "source-original.mkv")
	if err := os.WriteFile(orphanPath, []byte("orphan"), 0o644); err != nil {
		t.Fatalf("write orphan: %v", err)
	}
	h.videos.found = true
	h.videos.videoID = videoID
	h.videos.video = models.Video{ID: videoID, Status: "ready"}
	h.repo.record = models.CloudDownloadImport{ID: 32, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportFailed, VideoID: &videoID}

	if err := h.svc.Process(context.Background(), 32, 0, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if h.repo.record.Status != models.CloudImportImported {
		t.Fatalf("status = %s, want imported", h.repo.record.Status)
	}
	if len(h.upload.saveInputs) != 0 || h.enqueuer.transcodeCalls != 0 || h.enqueuer.scrapeCalls != 0 {
		t.Fatalf("ready target must not copy or enqueue: saves=%d transcode=%d scrape=%d", len(h.upload.saveInputs), h.enqueuer.transcodeCalls, h.enqueuer.scrapeCalls)
	}
	if _, statErr := os.Stat(orphanPath); !os.IsNotExist(statErr) {
		t.Fatalf("orphan source-original should be removed, stat err = %v", statErr)
	}
	if _, statErr := os.Stat(h.source); !os.IsNotExist(statErr) {
		t.Fatalf("download source should be deleted")
	}
}

func TestCloudImportResumeMissingTargetUsesVideoStatus(t *testing.T) {
	t.Parallel()

	t.Run("processing finishes", func(t *testing.T) {
		h := newCloudImportHarness(t)
		videoID := uuid.New()
		h.videos.found = true
		h.videos.videoID = videoID
		h.videos.video = models.Video{ID: videoID, Status: "processing", Type: "short"}
		h.repo.record = models.CloudDownloadImport{ID: 52, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportFailed, VideoID: &videoID}

		if err := h.svc.Process(context.Background(), 52, 0, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}); err != nil {
			t.Fatalf("Process() error = %v", err)
		}
		if h.repo.record.Status != models.CloudImportImported {
			t.Fatalf("status = %s, want imported", h.repo.record.Status)
		}
		if h.enqueuer.transcodeCalls != 0 {
			t.Fatalf("transcode calls = %d, want 0", h.enqueuer.transcodeCalls)
		}
	})

	t.Run("uploaded fails and keeps download", func(t *testing.T) {
		h := newCloudImportHarness(t)
		videoID := uuid.New()
		h.videos.found = true
		h.videos.videoID = videoID
		h.videos.video = models.Video{ID: videoID, Status: "uploaded", Type: "short"}
		h.repo.record = models.CloudDownloadImport{ID: 53, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportFailed, VideoID: &videoID}

		if err := h.svc.Process(context.Background(), 53, 0, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}); err == nil {
			t.Fatal("expected missing source error")
		}
		if h.repo.record.Status != models.CloudImportFailed {
			t.Fatalf("status = %s, want failed", h.repo.record.Status)
		}
		if _, statErr := os.Stat(h.source); statErr != nil {
			t.Fatalf("download source should be kept: %v", statErr)
		}
		if h.enqueuer.transcodeCalls != 0 {
			t.Fatalf("transcode calls = %d, want 0", h.enqueuer.transcodeCalls)
		}
	})
}

func TestCloudImportMarkFailedSkipsTerminal(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	h.repo.record = models.CloudDownloadImport{ID: 12, TaskID: 7, Kind: models.CloudImportKindVideo, Status: models.CloudImportImported}
	if err := h.svc.MarkFailed(context.Background(), 12, 0, "boom"); err != nil {
		t.Fatalf("MarkFailed() error = %v", err)
	}
	if h.repo.record.Status != models.CloudImportImported {
		t.Fatalf("terminal status overwritten: %s", h.repo.record.Status)
	}
}

func TestCloudImportProcessIgnoresStaleAttempt(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	h.repo.record = models.CloudDownloadImport{
		ID:      55,
		TaskID:  7,
		Kind:    models.CloudImportKindVideo,
		Status:  models.CloudImportQueued,
		Attempt: 1,
	}
	if err := h.svc.Process(context.Background(), 55, 0, models.CloudDownloadImportInput{Kind: models.CloudImportKindVideo, Type: "short"}); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if len(h.repo.statuses) != 0 {
		t.Fatalf("stale attempt changed statuses: %v", h.repo.statuses)
	}
	if h.repo.record.Status != models.CloudImportQueued {
		t.Fatalf("stale attempt changed status: %s", h.repo.record.Status)
	}
}

func TestCloudImportMarkFailedIgnoresStaleAttempt(t *testing.T) {
	t.Parallel()

	h := newCloudImportHarness(t)
	h.repo.record = models.CloudDownloadImport{
		ID:      56,
		TaskID:  7,
		Kind:    models.CloudImportKindVideo,
		Status:  models.CloudImportHashing,
		Attempt: 2,
	}
	if err := h.svc.MarkFailed(context.Background(), 56, 1, "old attempt"); err != nil {
		t.Fatalf("MarkFailed() error = %v", err)
	}
	if h.repo.record.Status != models.CloudImportHashing {
		t.Fatalf("stale MarkFailed changed status: %s", h.repo.record.Status)
	}
}

func TestCloudImportGetTaskRestoresArchiveWithoutSource(t *testing.T) {
	t.Parallel()
	for _, status := range []string{models.CloudImportQueued, models.CloudImportImported, models.CloudImportFailed} {
		t.Run(status, func(t *testing.T) {
			h := newCloudImportHarness(t)
			if err := os.Remove(h.source); err != nil {
				t.Fatal(err)
			}
			batchID := uuid.New()
			h.repo.record = models.CloudDownloadImport{ID: 12, TaskID: h.tasks.task.ID, Kind: models.CloudImportKindArchive, Status: status, ArchiveBatchID: &batchID, ErrorMessage: "原导入错误"}
			task, err := h.svc.GetImportTask(context.Background(), h.tasks.task.ID)
			if err != nil {
				t.Fatal(err)
			}
			if task.Import == nil || task.Import.ArchiveBatchID == nil || *task.Import.ArchiveBatchID != batchID || task.Import.Status != status {
				t.Fatalf("未恢复导入状态: %+v", task.Import)
			}
			if h.repo.createdCount != 0 || len(h.enqueuer.ingestCalls) != 0 {
				t.Fatal("查询不应创建或重新入队")
			}
		})
	}
}

func TestCloudImportGetTaskHandlesMissingAndReadErrors(t *testing.T) {
	t.Parallel()
	h := newCloudImportHarness(t)
	task, err := h.svc.GetImportTask(context.Background(), h.tasks.task.ID)
	if err != nil || task.Import != nil {
		t.Fatalf("无导入记录应返回任务: %+v %v", task, err)
	}
	h.repo.getErr = errors.New("模拟数据库不可用")
	if _, err = h.svc.GetImportTask(context.Background(), h.tasks.task.ID); !errors.Is(err, h.repo.getErr) {
		t.Fatalf("查询错误未传递: %v", err)
	}
	h.tasks.err = pgx.ErrNoRows
	if _, err = h.svc.GetImportTask(context.Background(), h.tasks.task.ID); !errors.Is(err, models.ErrCloudDownloadNotFound) {
		t.Fatalf("任务不存在: %v", err)
	}
}
