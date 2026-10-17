package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/cloud"
	"video-server/internal/models"
)

type fakeCloudRepo struct {
	task              models.CloudDownloadTask
	getErr            error
	createErr         error
	createCreated     bool
	lastCreate        models.CloudDownloadTask
	statuses          []string
	reasons           []string
	messages          []string
	progress          []float64
	readyPickCodes    []string
	localPaths        []string
	bumpRetryCalls    int
	bumpRetryDenied   bool
	triggerClaims     int
	triggerClaimed    bool
	clearTriggerCalls int
	unfinished        []models.CloudDownloadTask
	statusByHash      map[string]models.CloudDownloadStatus
	statusErr         error
	listTasks         []models.CloudDownloadTask
	setStatusHook     func(ctx context.Context, status string) error
}

type fakeCloudImportSummaryRepo struct {
	summaries map[int64]models.CloudDownloadImportSummary
	err       error
	gotIDs    []int64
}

func (f *fakeCloudImportSummaryRepo) ListSummariesByTaskIDs(_ context.Context, taskIDs []int64) (map[int64]models.CloudDownloadImportSummary, error) {
	f.gotIDs = append([]int64(nil), taskIDs...)
	return f.summaries, f.err
}

func (f *fakeCloudRepo) CreateOrGet(_ context.Context, task models.CloudDownloadTask) (models.CloudDownloadTask, bool, error) {
	if f.createErr != nil {
		return models.CloudDownloadTask{}, false, f.createErr
	}
	f.lastCreate = task
	if f.createCreated {
		f.task = task
		return task, true, nil
	}
	return f.task, false, nil
}

func (f *fakeCloudRepo) GetByID(_ context.Context, _ int64) (models.CloudDownloadTask, error) {
	if f.getErr != nil {
		return models.CloudDownloadTask{}, f.getErr
	}
	return f.task, nil
}

func (f *fakeCloudRepo) List(_ context.Context, _ string, _, _ int) ([]models.CloudDownloadTask, int, error) {
	return f.listTasks, len(f.listTasks), nil
}

func (f *fakeCloudRepo) StatusByHashes(_ context.Context, _ []string) (map[string]models.CloudDownloadStatus, error) {
	if f.statusErr != nil {
		return nil, f.statusErr
	}
	if f.statusByHash == nil {
		return map[string]models.CloudDownloadStatus{}, nil
	}
	return f.statusByHash, nil
}

func (f *fakeCloudRepo) SetStatus(ctx context.Context, _ int64, status, reason, message string) error {
	if f.setStatusHook != nil {
		if err := f.setStatusHook(ctx, status); err != nil {
			return err
		}
	}
	f.statuses = append(f.statuses, status)
	f.reasons = append(f.reasons, reason)
	f.messages = append(f.messages, message)
	f.task.Status = status
	f.task.BlockedReason = reason
	f.task.ErrorMessage = message
	return nil
}

func (f *fakeCloudRepo) SetProgress(_ context.Context, _ int64, percent float64, status string) error {
	f.progress = append(f.progress, percent)
	f.task.ProgressPercent = &percent
	f.task.CloudStatus = status
	return nil
}

func (f *fakeCloudRepo) SetReady(_ context.Context, _ int64, pickCode string) error {
	f.readyPickCodes = append(f.readyPickCodes, pickCode)
	f.task.PickCode = pickCode
	f.task.Status = models.CloudDownloadReady
	return nil
}

func (f *fakeCloudRepo) SetLocalPath(_ context.Context, _ int64, localPath string) error {
	f.localPaths = append(f.localPaths, localPath)
	f.task.LocalPath = localPath
	f.task.Status = models.CloudDownloadCompleted
	return nil
}

func (f *fakeCloudRepo) BumpRetry(_ context.Context, _ int64) (bool, error) {
	f.bumpRetryCalls++
	if f.bumpRetryDenied {
		// Simulate another request having already moved the row to pending.
		f.task.Status = models.CloudDownloadPending
		return false, nil
	}
	f.task.RetryCount++
	f.task.Status = models.CloudDownloadPending
	f.task.BlockedReason = ""
	f.task.ErrorMessage = ""
	f.task.FinishedAt = nil
	f.task.DownloadTriggeredAt = nil
	return true, nil
}

func (f *fakeCloudRepo) ClaimDownloadTrigger(_ context.Context, _ int64) (bool, error) {
	f.triggerClaims++
	if !f.triggerClaimed {
		return false, nil
	}
	now := time.Now()
	f.task.DownloadTriggeredAt = &now
	return true, nil
}

func (f *fakeCloudRepo) ClearDownloadTrigger(_ context.Context, _ int64) error {
	f.clearTriggerCalls++
	f.task.DownloadTriggeredAt = nil
	return nil
}

func (f *fakeCloudRepo) ListUnfinished(_ context.Context, _ int) ([]models.CloudDownloadTask, error) {
	return f.unfinished, nil
}

type fakeCloudSession struct {
	quota         cloud.Quota
	quotaErr      error
	dirID         string
	dirErr        error
	submitResult  cloud.SubmitResult
	submitErr     error
	submitted     int
	findTask      *cloud.Task
	findErr       error
	finds         int
	waitTask      *cloud.Task
	waitErr       error
	cancelOnWait  context.CancelFunc
	triggerErr    error
	triggered     int
	triggeredName string
}

func (f *fakeCloudSession) Close() {}

func (f *fakeCloudSession) Quota(_ context.Context) (cloud.Quota, error) {
	return f.quota, f.quotaErr
}

func (f *fakeCloudSession) TargetDirID(_ context.Context) (string, error) {
	return f.dirID, f.dirErr
}

func (f *fakeCloudSession) SubmitURLs(_ context.Context, _ []string, _ string) (cloud.SubmitResult, error) {
	f.submitted++
	return f.submitResult, f.submitErr
}

func (f *fakeCloudSession) FindTask(_ context.Context, _ string) (*cloud.Task, error) {
	f.finds++
	return f.findTask, f.findErr
}

func (f *fakeCloudSession) WaitForTask(_ context.Context, _ string, _ time.Duration, _ func(cloud.Task)) (*cloud.Task, error) {
	if f.cancelOnWait != nil {
		f.cancelOnWait()
	}
	return f.waitTask, f.waitErr
}

func (f *fakeCloudSession) CreateDownloadTask(_ context.Context, name, _ string) error {
	f.triggered++
	f.triggeredName = name
	return f.triggerErr
}

type fakeCloudEnqueuer struct {
	ids    []int64
	retrys []int
	err    error
}

func (f *fakeCloudEnqueuer) EnqueueCloudDownload(taskID int64, retryCount int) error {
	f.ids = append(f.ids, taskID)
	f.retrys = append(f.retrys, retryCount)
	return f.err
}

func newCloudTestService(repo *fakeCloudRepo, session *fakeCloudSession, enq *fakeCloudEnqueuer, dir string) *CloudDownloadService {
	return &CloudDownloadService{
		repo:             repo,
		downloadDir:      dir,
		enqueuer:         enq,
		pollInterval:     time.Millisecond,
		filePollInterval: 5 * time.Millisecond,
		fileWaitTimeout:  2 * time.Second,
		ensureSession: func(context.Context) (cloudSession, cloud.State, error) {
			return session, cloud.State{Up: true, LoggedIn: true}, nil
		},
	}
}

func int64Ptr(value int64) *int64 { return &value }

func testED2KLink(name, hash string) string {
	return fmt.Sprintf("ed2k://|file|%s|1|%s|/", name, hash)
}

func writeDownloadFile(t *testing.T, dir, name string, size int) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), make([]byte, size), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestProcessResubmitsSubmittingTask(t *testing.T) {
	dir := t.TempDir()
	writeDownloadFile(t, dir, "movie.zip", 5)

	repo := &fakeCloudRepo{
		task: models.CloudDownloadTask{
			ID:           1,
			Status:       models.CloudDownloadSubmitting,
			InfoHash:     "0123456789abcdef0123456789abcdef",
			SourceURL:    "ed2k://|file|movie.zip|5|0123456789abcdef0123456789abcdef|/",
			DeclaredName: "movie.zip",
			DeclaredSize: int64Ptr(5),
		},
		triggerClaimed: true,
	}
	session := &fakeCloudSession{
		quota:        cloud.Quota{Known: true, SurplusKnown: true, Surplus: 10},
		dirID:        "123",
		submitResult: cloud.SubmitResult{State: true},
		waitTask: &cloud.Task{
			InfoHash: "0123456789abcdef0123456789abcdef",
			Name:     "movie.zip",
			Size:     5,
			Status:   "finished",
			PickCode: "pick-code",
			Percent:  100,
		},
	}
	svc := newCloudTestService(repo, session, &fakeCloudEnqueuer{}, dir)

	if err := svc.Process(context.Background(), 1); err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	if session.submitted != 1 {
		t.Fatalf("submitted = %d, want 1 (submitting must be re-submitted)", session.submitted)
	}
	if repo.triggerClaims != 1 || session.triggered != 1 {
		t.Fatalf("trigger claims = %d, triggered = %d, want 1/1", repo.triggerClaims, session.triggered)
	}
	if repo.task.Status != models.CloudDownloadCompleted {
		t.Fatalf("status = %q, want completed", repo.task.Status)
	}
	if len(repo.localPaths) != 1 || repo.localPaths[0] != filepath.Join(dir, "movie.zip") {
		t.Fatalf("local paths = %v", repo.localPaths)
	}
}

func TestProcessSkipsSubmitWhenSubmittingTaskExistsOn115(t *testing.T) {
	dir := t.TempDir()
	writeDownloadFile(t, dir, "movie.zip", 5)
	const hash = "0123456789abcdef0123456789abcdef"

	repo := &fakeCloudRepo{
		task: models.CloudDownloadTask{
			ID:           12,
			Status:       models.CloudDownloadSubmitting,
			InfoHash:     hash,
			SourceURL:    "ed2k://|file|movie.zip|5|0123456789abcdef0123456789abcdef|/",
			DeclaredName: "movie.zip",
			DeclaredSize: int64Ptr(5),
		},
		triggerClaimed: true,
	}
	finished := &cloud.Task{
		InfoHash: hash,
		Name:     "movie.zip",
		Size:     5,
		Status:   "finished",
		PickCode: "pick-code",
		Percent:  100,
	}
	session := &fakeCloudSession{
		quota:    cloud.Quota{Known: true, SurplusKnown: true, Surplus: 10},
		findTask: finished,
		waitTask: finished,
	}
	svc := newCloudTestService(repo, session, &fakeCloudEnqueuer{}, dir)

	if err := svc.Process(context.Background(), 12); err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	if session.finds != 1 {
		t.Fatalf("FindTask calls = %d, want 1", session.finds)
	}
	if session.submitted != 0 {
		t.Fatalf("submitted = %d, want 0 when 115 already has the task", session.submitted)
	}
	if repo.task.Status != models.CloudDownloadCompleted {
		t.Fatalf("status = %q, want completed", repo.task.Status)
	}
}

func TestProcessSkipsSubmitWhenRetriedPendingTaskExistsOn115(t *testing.T) {
	dir := t.TempDir()
	writeDownloadFile(t, dir, "movie.zip", 5)
	const hash = "0123456789abcdef0123456789abcdef"

	repo := &fakeCloudRepo{
		task: models.CloudDownloadTask{
			ID:           13,
			Status:       models.CloudDownloadPending,
			RetryCount:   1,
			InfoHash:     hash,
			SourceURL:    "ed2k://|file|movie.zip|5|0123456789abcdef0123456789abcdef|/",
			DeclaredName: "movie.zip",
			DeclaredSize: int64Ptr(5),
		},
		triggerClaimed: true,
	}
	finished := &cloud.Task{
		InfoHash: hash,
		Name:     "movie.zip",
		Size:     5,
		Status:   "finished",
		PickCode: "pick-code",
		Percent:  100,
	}
	session := &fakeCloudSession{
		quota:    cloud.Quota{Known: true, SurplusKnown: true, Surplus: 10},
		findTask: finished,
		waitTask: finished,
	}
	svc := newCloudTestService(repo, session, &fakeCloudEnqueuer{}, dir)

	if err := svc.Process(context.Background(), 13); err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	if session.finds != 1 {
		t.Fatalf("FindTask calls = %d, want 1", session.finds)
	}
	if session.submitted != 0 {
		t.Fatalf("submitted = %d, want 0 when 115 already has the retried task", session.submitted)
	}
	if repo.task.Status != models.CloudDownloadCompleted {
		t.Fatalf("status = %q, want completed", repo.task.Status)
	}
}

func TestProcessWritesBlockedStatusAfterCanceledWait(t *testing.T) {
	repo := &fakeCloudRepo{
		task: models.CloudDownloadTask{
			ID:       2,
			Status:   models.CloudDownloadQueued,
			InfoHash: "0123456789abcdef0123456789abcdef",
		},
		setStatusHook: func(ctx context.Context, _ string) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			return nil
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	session := &fakeCloudSession{
		quota:        cloud.Quota{Known: true, SurplusKnown: true, Surplus: 10},
		waitErr:      context.Canceled,
		cancelOnWait: cancel,
	}
	svc := newCloudTestService(repo, session, &fakeCloudEnqueuer{}, t.TempDir())

	if err := svc.Process(ctx, 2); err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	if repo.task.Status != models.CloudDownloadBlocked {
		t.Fatalf("status = %q, want blocked", repo.task.Status)
	}
	if repo.task.BlockedReason != models.CloudBlockedNeedsAttention {
		t.Fatalf("blocked reason = %q", repo.task.BlockedReason)
	}
}

func TestProcessBlocksWhenQuotaIsKnownExhausted(t *testing.T) {
	repo := &fakeCloudRepo{
		task: models.CloudDownloadTask{
			ID:        10,
			Status:    models.CloudDownloadPending,
			InfoHash:  "0123456789abcdef0123456789abcdef",
			SourceURL: "ed2k://|file|a.zip|1|0123456789abcdef0123456789abcdef|/",
		},
	}
	session := &fakeCloudSession{quota: cloud.Quota{Known: true, SurplusKnown: true, Surplus: 0}}
	svc := newCloudTestService(repo, session, &fakeCloudEnqueuer{}, t.TempDir())

	if err := svc.Process(context.Background(), 10); err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	if repo.task.Status != models.CloudDownloadBlocked || repo.task.BlockedReason != models.CloudBlockedQuotaExhausted {
		t.Fatalf("task = %+v, want blocked quota_exhausted", repo.task)
	}
	if session.submitted != 0 {
		t.Fatalf("submitted = %d, want 0", session.submitted)
	}
}

func TestProcessDoesNotBlockWhenSurplusIsUnknown(t *testing.T) {
	dir := t.TempDir()
	writeDownloadFile(t, dir, "a.zip", 1)
	repo := &fakeCloudRepo{
		task: models.CloudDownloadTask{
			ID:           11,
			Status:       models.CloudDownloadPending,
			InfoHash:     "0123456789abcdef0123456789abcdef",
			SourceURL:    "ed2k://|file|a.zip|1|0123456789abcdef0123456789abcdef|/",
			DeclaredName: "a.zip",
			DeclaredSize: int64Ptr(1),
		},
		triggerClaimed: true,
	}
	session := &fakeCloudSession{
		quota:        cloud.Quota{Known: true, SurplusKnown: false},
		dirID:        "123",
		submitResult: cloud.SubmitResult{State: true},
		waitTask: &cloud.Task{
			InfoHash: "0123456789abcdef0123456789abcdef",
			Name:     "a.zip",
			Size:     1,
			Status:   "finished",
			PickCode: "pick-code",
			Percent:  100,
		},
	}
	svc := newCloudTestService(repo, session, &fakeCloudEnqueuer{}, dir)

	if err := svc.Process(context.Background(), 11); err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	if session.submitted != 1 {
		t.Fatalf("submitted = %d, want 1 (unknown surplus must not block)", session.submitted)
	}
	if repo.task.Status != models.CloudDownloadCompleted {
		t.Fatalf("status = %q, want completed", repo.task.Status)
	}
}

func TestProcessSkipsTriggerWhenAlreadyClaimed(t *testing.T) {
	dir := t.TempDir()
	writeDownloadFile(t, dir, "movie.zip", 5)
	now := time.Now()

	repo := &fakeCloudRepo{
		task: models.CloudDownloadTask{
			ID:                  3,
			Status:              models.CloudDownloadDownloading,
			InfoHash:            "0123456789abcdef0123456789abcdef",
			DeclaredName:        "movie.zip",
			DeclaredSize:        int64Ptr(5),
			DownloadTriggeredAt: &now,
		},
		triggerClaimed: false,
	}
	session := &fakeCloudSession{
		quota: cloud.Quota{Known: true, Surplus: 10},
		waitTask: &cloud.Task{
			InfoHash: "0123456789abcdef0123456789abcdef",
			Name:     "movie.zip",
			Size:     5,
			Status:   "finished",
			PickCode: "pick-code",
			Percent:  100,
		},
	}
	svc := newCloudTestService(repo, session, &fakeCloudEnqueuer{}, dir)

	if err := svc.Process(context.Background(), 3); err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	if session.triggered != 0 {
		t.Fatalf("triggered = %d, want 0 when the claim is already taken", session.triggered)
	}
	if repo.task.Status != models.CloudDownloadCompleted {
		t.Fatalf("status = %q, want completed", repo.task.Status)
	}
}

func TestRetryRejectsCompletedTask(t *testing.T) {
	repo := &fakeCloudRepo{task: models.CloudDownloadTask{ID: 4, Status: models.CloudDownloadCompleted}}
	enq := &fakeCloudEnqueuer{}
	svc := newCloudTestService(repo, &fakeCloudSession{}, enq, t.TempDir())

	_, err := svc.Retry(context.Background(), 4)
	if !errors.Is(err, models.ErrCloudDownloadNotRetryable) {
		t.Fatalf("error = %v, want ErrCloudDownloadNotRetryable", err)
	}
	if len(enq.ids) != 0 || repo.bumpRetryCalls != 0 {
		t.Fatalf("completed task must not be retried")
	}
}

func TestRetryMapsMissingTask(t *testing.T) {
	repo := &fakeCloudRepo{getErr: fmt.Errorf("get: %w", pgx.ErrNoRows)}
	svc := newCloudTestService(repo, &fakeCloudSession{}, &fakeCloudEnqueuer{}, t.TempDir())

	_, err := svc.Retry(context.Background(), 99)
	if !errors.Is(err, models.ErrCloudDownloadNotFound) {
		t.Fatalf("error = %v, want ErrCloudDownloadNotFound", err)
	}
}

func TestSubmitResetsBlockedTask(t *testing.T) {
	repo := &fakeCloudRepo{
		task: models.CloudDownloadTask{ID: 5, Status: models.CloudDownloadBlocked, BlockedReason: models.CloudBlockedLoginRequired, RetryCount: 1},
	}
	enq := &fakeCloudEnqueuer{}
	svc := newCloudTestService(repo, &fakeCloudSession{}, enq, t.TempDir())

	task, created, err := svc.Submit(context.Background(), nil, "ed2k://|file|a.zip|1|0123456789abcdef0123456789abcdef|/", nil, models.CloudDownloadSourceManual)
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if created {
		t.Fatalf("created = true, want false for an existing task")
	}
	if repo.bumpRetryCalls != 1 {
		t.Fatalf("bumpRetryCalls = %d, want 1", repo.bumpRetryCalls)
	}
	if task.Status != models.CloudDownloadPending || task.RetryCount != 2 {
		t.Fatalf("task = %+v, want pending retry_count=2", task)
	}
	if len(enq.ids) != 1 || enq.ids[0] != 5 || enq.retrys[0] != 2 {
		t.Fatalf("enqueue calls = %+v, want id=5 retry=2", enq)
	}
}

func TestSubmitKeepsForumPostUUID(t *testing.T) {
	repo := &fakeCloudRepo{createCreated: true}
	enq := &fakeCloudEnqueuer{}
	svc := newCloudTestService(repo, &fakeCloudSession{}, enq, t.TempDir())
	postID := uuid.New()

	task, created, err := svc.Submit(context.Background(), &postID, "ed2k://|file|a.zip|1|0123456789abcdef0123456789abcdef|/", nil, models.CloudDownloadSourceForumPost)
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if !created {
		t.Fatalf("created = false, want true")
	}
	if repo.lastCreate.PostID == nil || *repo.lastCreate.PostID != postID {
		t.Fatalf("stored post_id = %v, want %s", repo.lastCreate.PostID, postID)
	}
	if repo.lastCreate.Source != models.CloudDownloadSourceForumPost {
		t.Fatalf("stored source = %q, want %q", repo.lastCreate.Source, models.CloudDownloadSourceForumPost)
	}
	if task.PostID == nil || *task.PostID != postID {
		t.Fatalf("returned post_id = %v, want %s", task.PostID, postID)
	}
}

func TestSubmitBatchDedupesAndCreatesIndependentTasks(t *testing.T) {
	const (
		hashA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	)
	repo := &fakeCloudRepo{createCreated: true}
	enq := &fakeCloudEnqueuer{}
	svc := newCloudTestService(repo, &fakeCloudSession{}, enq, t.TempDir())

	result, err := svc.SubmitBatch(context.Background(), []string{
		testED2KLink("a.mkv", hashA),
		testED2KLink("a-copy.mkv", hashA),
		"not-an-ed2k-link",
	}, nil)
	if err != nil {
		t.Fatalf("SubmitBatch returned error: %v", err)
	}
	if result.Summary.Total != 3 || result.Summary.Submitted != 1 ||
		result.Summary.Duplicate != 1 || result.Summary.Invalid != 1 {
		t.Fatalf("summary = %+v, want 3 total / 1 submitted / 1 duplicate / 1 invalid", result.Summary)
	}
	if len(result.Items) != 3 {
		t.Fatalf("items = %d, want 3", len(result.Items))
	}
	if result.Items[0].Status != models.CloudBatchSubmitted {
		t.Fatalf("item 0 status = %q, want submitted", result.Items[0].Status)
	}
	if result.Items[1].Status != models.CloudBatchDuplicate || result.Items[1].InfoHash != hashA {
		t.Fatalf("item 1 = %+v, want duplicate with info hash", result.Items[1])
	}
	if result.Items[2].Status != models.CloudBatchInvalid {
		t.Fatalf("item 2 status = %q, want invalid", result.Items[2].Status)
	}
	if repo.lastCreate.Source != models.CloudDownloadSourceToolbox {
		t.Fatalf("created source = %q, want %q", repo.lastCreate.Source, models.CloudDownloadSourceToolbox)
	}
	if len(enq.ids) != 1 {
		t.Fatalf("enqueue calls = %v, want exactly 1", enq.ids)
	}
}

func TestSubmitBatchSkipsTerminalAndRetriesFailed(t *testing.T) {
	const (
		hashA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		hashB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		hashC = "cccccccccccccccccccccccccccccccc"
	)
	repo := &fakeCloudRepo{
		task: models.CloudDownloadTask{
			ID:         7,
			InfoHash:   hashB,
			Status:     models.CloudDownloadFailed,
			RetryCount: 1,
		},
		statusByHash: map[string]models.CloudDownloadStatus{
			hashA: {TaskID: 1, Status: models.CloudDownloadCompleted},
			hashB: {TaskID: 7, Status: models.CloudDownloadFailed},
			hashC: {TaskID: 3, Status: models.CloudDownloadQueued},
		},
	}
	enq := &fakeCloudEnqueuer{}
	svc := newCloudTestService(repo, &fakeCloudSession{}, enq, t.TempDir())

	result, err := svc.SubmitBatch(context.Background(), []string{
		testED2KLink("a.mkv", hashA),
		testED2KLink("b.mkv", hashB),
		testED2KLink("c.mkv", hashC),
	}, nil)
	if err != nil {
		t.Fatalf("SubmitBatch returned error: %v", err)
	}
	want := []string{
		models.CloudBatchSkippedCompleted,
		models.CloudBatchRetried,
		models.CloudBatchSkippedActive,
	}
	for index, status := range want {
		if result.Items[index].Status != status {
			t.Fatalf("item %d status = %q, want %q", index, result.Items[index].Status, status)
		}
	}
	if result.Items[0].TaskID == nil || *result.Items[0].TaskID != 1 ||
		result.Items[1].TaskID == nil || *result.Items[1].TaskID != 7 ||
		result.Items[2].TaskID == nil || *result.Items[2].TaskID != 3 {
		t.Fatalf("task ids = %+v", result.Items)
	}
	if repo.bumpRetryCalls != 1 {
		t.Fatalf("bumpRetryCalls = %d, want 1", repo.bumpRetryCalls)
	}
	if len(enq.ids) != 1 || enq.ids[0] != 7 {
		t.Fatalf("enqueue calls = %v, want [7]", enq.ids)
	}
}

func TestSubmitBatchContinuesAfterItemFailure(t *testing.T) {
	const (
		hashA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		hashB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	)
	repo := &fakeCloudRepo{
		createCreated: true,
		statusByHash: map[string]models.CloudDownloadStatus{
			hashB: {TaskID: 2, Status: models.CloudDownloadCompleted},
		},
	}
	enq := &fakeCloudEnqueuer{err: errors.New("redis down")}
	svc := newCloudTestService(repo, &fakeCloudSession{}, enq, t.TempDir())

	result, err := svc.SubmitBatch(context.Background(), []string{
		testED2KLink("a.mkv", hashA),
		testED2KLink("b.mkv", hashB),
	}, nil)
	if err != nil {
		t.Fatalf("SubmitBatch returned error for a per-item failure: %v", err)
	}
	if result.Items[0].Status != models.CloudBatchFailed || result.Items[0].Message == "" {
		t.Fatalf("item 0 = %+v, want failed with message", result.Items[0])
	}
	if result.Items[1].Status != models.CloudBatchSkippedCompleted {
		t.Fatalf("item 1 status = %q, want skipped_completed", result.Items[1].Status)
	}
	if result.Summary.Failed != 1 || result.Summary.SkippedCompleted != 1 {
		t.Fatalf("summary = %+v", result.Summary)
	}
}

func TestSubmitMarksFailedWhenEnqueueFails(t *testing.T) {
	repo := &fakeCloudRepo{createCreated: true}
	enq := &fakeCloudEnqueuer{err: errors.New("redis down")}
	svc := newCloudTestService(repo, &fakeCloudSession{}, enq, t.TempDir())

	_, _, err := svc.Submit(context.Background(), nil,
		"ed2k://|file|a.zip|1|0123456789abcdef0123456789abcdef|/", nil,
		models.CloudDownloadSourceManual)
	if err == nil {
		t.Fatalf("Submit returned nil error")
	}
	if repo.task.Status != models.CloudDownloadFailed {
		t.Fatalf("persisted status = %q, want failed", repo.task.Status)
	}
	if repo.task.ErrorMessage == "" {
		t.Fatalf("persisted error message is empty")
	}
}

func TestSubmitBatchHandlesConcurrentCreate(t *testing.T) {
	const hash = "dddddddddddddddddddddddddddddddd"
	repo := &fakeCloudRepo{
		task: models.CloudDownloadTask{ID: 9, InfoHash: hash, Status: models.CloudDownloadQueued},
	}
	enq := &fakeCloudEnqueuer{}
	svc := newCloudTestService(repo, &fakeCloudSession{}, enq, t.TempDir())

	result, err := svc.SubmitBatch(context.Background(), []string{
		testED2KLink("d.mkv", hash),
	}, nil)
	if err != nil {
		t.Fatalf("SubmitBatch returned error: %v", err)
	}
	if result.Items[0].Status != models.CloudBatchSkippedActive {
		t.Fatalf("status = %q, want skipped_active", result.Items[0].Status)
	}
	if result.Items[0].TaskID == nil || *result.Items[0].TaskID != 9 {
		t.Fatalf("task id = %v, want 9", result.Items[0].TaskID)
	}
	if len(enq.ids) != 0 {
		t.Fatalf("enqueue calls = %v, want none", enq.ids)
	}
}

func TestSubmitBatchAcceptsFiftyLinks(t *testing.T) {
	repo := &fakeCloudRepo{createCreated: true}
	enq := &fakeCloudEnqueuer{}
	svc := newCloudTestService(repo, &fakeCloudSession{}, enq, t.TempDir())
	links := make([]string, models.CloudDownloadBatchLimit)
	for index := range links {
		links[index] = testED2KLink(fmt.Sprintf("f%d.mkv", index), fmt.Sprintf("%032x", index+1))
	}

	result, err := svc.SubmitBatch(context.Background(), links, nil)
	if err != nil {
		t.Fatalf("SubmitBatch returned error: %v", err)
	}
	if result.Summary.Submitted != models.CloudDownloadBatchLimit || len(enq.ids) != models.CloudDownloadBatchLimit {
		t.Fatalf("summary = %+v, enqueues = %d", result.Summary, len(enq.ids))
	}
}

func TestSubmitBatchReturnsStatusLookupError(t *testing.T) {
	repo := &fakeCloudRepo{statusErr: errors.New("database down")}
	enq := &fakeCloudEnqueuer{}
	svc := newCloudTestService(repo, &fakeCloudSession{}, enq, t.TempDir())

	_, err := svc.SubmitBatch(context.Background(), []string{
		testED2KLink("a.mkv", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
	}, nil)
	if err == nil {
		t.Fatalf("SubmitBatch returned nil error")
	}
	if len(enq.ids) != 0 {
		t.Fatalf("enqueue calls = %v, want none", enq.ids)
	}
}

func TestSubmitBatchDoesNotDuplicateAlreadyClaimedRetry(t *testing.T) {
	const hash = "ffffffffffffffffffffffffffffffff"
	repo := &fakeCloudRepo{
		task: models.CloudDownloadTask{
			ID:         8,
			InfoHash:   hash,
			Status:     models.CloudDownloadFailed,
			RetryCount: 2,
		},
		statusByHash: map[string]models.CloudDownloadStatus{
			hash: {TaskID: 8, Status: models.CloudDownloadFailed},
		},
		bumpRetryDenied: true,
	}
	enq := &fakeCloudEnqueuer{}
	svc := newCloudTestService(repo, &fakeCloudSession{}, enq, t.TempDir())

	result, err := svc.SubmitBatch(context.Background(), []string{
		testED2KLink("f.mkv", hash),
	}, nil)
	if err != nil {
		t.Fatalf("SubmitBatch returned error: %v", err)
	}
	if result.Items[0].Status != models.CloudBatchRetried {
		t.Fatalf("status = %q, want retried", result.Items[0].Status)
	}
	if len(enq.ids) != 0 {
		t.Fatalf("enqueue calls = %v, want none for an already-claimed retry", enq.ids)
	}
}

func TestSubmitBatchDoesNotRewriteExistingSource(t *testing.T) {
	const hash = "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	repo := &fakeCloudRepo{
		task: models.CloudDownloadTask{
			ID:         7,
			InfoHash:   hash,
			Status:     models.CloudDownloadFailed,
			RetryCount: 1,
			Source:     models.CloudDownloadSourceForumPost,
		},
		statusByHash: map[string]models.CloudDownloadStatus{
			hash: {TaskID: 7, Status: models.CloudDownloadFailed},
		},
	}
	enq := &fakeCloudEnqueuer{}
	svc := newCloudTestService(repo, &fakeCloudSession{}, enq, t.TempDir())

	result, err := svc.SubmitBatch(context.Background(), []string{
		testED2KLink("e.mkv", hash),
	}, nil)
	if err != nil {
		t.Fatalf("SubmitBatch returned error: %v", err)
	}
	if result.Items[0].Status != models.CloudBatchRetried {
		t.Fatalf("status = %q, want retried", result.Items[0].Status)
	}
	if repo.task.Source != models.CloudDownloadSourceForumPost {
		t.Fatalf("source = %q, want forum_post", repo.task.Source)
	}
}

func TestSubmitBatchRejectsEmptyAndOversizedRequests(t *testing.T) {
	repo := &fakeCloudRepo{}
	svc := newCloudTestService(repo, &fakeCloudSession{}, &fakeCloudEnqueuer{}, t.TempDir())

	if _, err := svc.SubmitBatch(context.Background(), nil, nil); err == nil {
		t.Fatalf("SubmitBatch(nil) returned nil error")
	}
	links := make([]string, models.CloudDownloadBatchLimit+1)
	for index := range links {
		links[index] = testED2KLink("x.mkv", fmt.Sprintf("%032x", index+1))
	}
	if _, err := svc.SubmitBatch(context.Background(), links, nil); err == nil {
		t.Fatalf("SubmitBatch(%d links) returned nil error", len(links))
	}
}

func TestListAttachesImportSummary(t *testing.T) {
	t.Parallel()

	videoID := uuid.New()
	repo := &fakeCloudRepo{listTasks: []models.CloudDownloadTask{{ID: 1}, {ID: 2}}}
	summaries := &fakeCloudImportSummaryRepo{summaries: map[int64]models.CloudDownloadImportSummary{
		1: {ID: 9, Kind: models.CloudImportKindVideo, Status: models.CloudImportImported, VideoID: &videoID},
	}}
	svc := &CloudDownloadService{repo: repo}
	svc.SetImportSummaryRepository(summaries)

	tasks, total, err := svc.List(context.Background(), "", 1, 20)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 2 || len(tasks) != 2 {
		t.Fatalf("total = %d tasks = %d", total, len(tasks))
	}
	if tasks[0].Import == nil || tasks[0].Import.Status != models.CloudImportImported {
		t.Fatalf("task 1 import = %+v", tasks[0].Import)
	}
	if tasks[0].Import.VideoID == nil || *tasks[0].Import.VideoID != videoID {
		t.Fatalf("task 1 import video = %v", tasks[0].Import.VideoID)
	}
	if tasks[1].Import != nil {
		t.Fatalf("task 2 import = %+v, want nil", tasks[1].Import)
	}
	if len(summaries.gotIDs) != 2 || summaries.gotIDs[0] != 1 || summaries.gotIDs[1] != 2 {
		t.Fatalf("summary ids = %v", summaries.gotIDs)
	}
}

func TestReconcileEnqueuesUnfinishedTasks(t *testing.T) {
	repo := &fakeCloudRepo{unfinished: []models.CloudDownloadTask{
		{ID: 6, RetryCount: 0},
		{ID: 7, RetryCount: 3},
	}}
	enq := &fakeCloudEnqueuer{}
	svc := newCloudTestService(repo, &fakeCloudSession{}, enq, t.TempDir())

	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}
	if len(enq.ids) != 2 || enq.ids[0] != 6 || enq.ids[1] != 7 {
		t.Fatalf("enqueue ids = %v", enq.ids)
	}
	if enq.retrys[0] != 0 || enq.retrys[1] != 3 {
		t.Fatalf("enqueue retries = %v", enq.retrys)
	}
}

func TestMarkFailedSkipsBlockedTask(t *testing.T) {
	repo := &fakeCloudRepo{task: models.CloudDownloadTask{ID: 8, Status: models.CloudDownloadBlocked}}
	svc := newCloudTestService(repo, &fakeCloudSession{}, &fakeCloudEnqueuer{}, t.TempDir())

	if err := svc.MarkFailed(context.Background(), 8, "boom"); err != nil {
		t.Fatalf("MarkFailed returned error: %v", err)
	}
	if len(repo.statuses) != 0 {
		t.Fatalf("blocked task must not be overwritten: %v", repo.statuses)
	}
}

func TestMarkFailedRecordsTerminalFailure(t *testing.T) {
	repo := &fakeCloudRepo{task: models.CloudDownloadTask{ID: 9, Status: models.CloudDownloadQueued}}
	svc := newCloudTestService(repo, &fakeCloudSession{}, &fakeCloudEnqueuer{}, t.TempDir())

	if err := svc.MarkFailed(context.Background(), 9, "queue gave up"); err != nil {
		t.Fatalf("MarkFailed returned error: %v", err)
	}
	if repo.task.Status != models.CloudDownloadFailed || repo.task.ErrorMessage != "queue gave up" {
		t.Fatalf("task = %+v", repo.task)
	}
}

func TestFindDownloadedFilePrefersExactAndSkipsPartial(t *testing.T) {
	dir := t.TempDir()
	writeDownloadFile(t, dir, "movie.zip.crdownload", 5)
	writeDownloadFile(t, dir, "movie.zip.part", 5)
	writeDownloadFile(t, dir, "movie (1).zip", 5)

	path, info, ok := findDownloadedFile(dir, "movie.zip", "movie", int64Ptr(5))
	if !ok || info.Size() != 5 || filepath.Base(path) != "movie (1).zip" {
		t.Fatalf("unexpected duplicate match: %q %v %v", path, info, ok)
	}

	writeDownloadFile(t, dir, "movie.zip", 5)
	path, _, ok = findDownloadedFile(dir, "movie.zip", "movie", int64Ptr(5))
	if !ok || filepath.Base(path) != "movie.zip" {
		t.Fatalf("exact file must win: %q", path)
	}

	if _, _, ok := findDownloadedFile(dir, "movie.zip", "movie", int64Ptr(6)); ok {
		t.Fatalf("size mismatch must not match")
	}
	writeDownloadFile(t, dir, "only-part.zip.part", 5)
	if _, _, ok := findDownloadedFile(dir, "only-part.zip", "only-part", int64Ptr(5)); ok {
		t.Fatalf("only a partial file must not match")
	}
}

func TestExpectedDownloadSizePrefersCloudTaskSize(t *testing.T) {
	declared := int64(5)
	task := &cloud.Task{Size: 7}
	if got := expectedDownloadSize(task, &declared); got == nil || *got != 7 {
		t.Fatalf("got %v, want 7", got)
	}
	if got := expectedDownloadSize(&cloud.Task{}, &declared); got == nil || *got != 5 {
		t.Fatalf("got %v, want 5", got)
	}
	if got := expectedDownloadSize(nil, nil); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}
