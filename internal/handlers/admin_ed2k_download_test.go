package handlers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
	"video-server/internal/queue"
)

func TestRegisterIncludesEd2kDownloadRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := &API{}
	api.Register(router)

	routes := map[string]struct{}{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}

	for _, want := range []string{
		"GET /api/v1/admin/ed2k-download/tasks",
		"POST /api/v1/admin/ed2k-download/tasks",
		"GET /api/v1/admin/ed2k-download/tasks/:id",
		"POST /api/v1/admin/ed2k-download/tasks/:id/retry",
		"POST /api/v1/admin/ed2k-download/tasks/:id/clean-files",
		"POST /api/v1/admin/ed2k-download/tasks/:id/retry-cleanup",
		"DELETE /api/v1/admin/ed2k-download/tasks/:id",
	} {
		if _, ok := routes[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
}

func TestParseEd2kDownloadLinkAndBuildTitle(t *testing.T) {
	sourceLink, title, filename, hash, size, ok := parseEd2kDownloadLink("ed2k://|file|%E4%B8%BB%E8%A7%92.mkv|123|ABCDEF0123456789ABCDEF0123456789|/")
	if !ok {
		t.Fatal("expected ed2k link to parse")
	}
	if sourceLink != "ed2k://|file|%E4%B8%BB%E8%A7%92.mkv|123|ABCDEF0123456789ABCDEF0123456789|/" {
		t.Fatalf("unexpected source link: %s", sourceLink)
	}
	if title != "主角.mkv" || filename != "主角.mkv" {
		t.Fatalf("unexpected title or filename: %s %s", title, filename)
	}
	if hash != "ABCDEF0123456789ABCDEF0123456789" {
		t.Fatalf("unexpected hash: %s", hash)
	}
	if size != 123 {
		t.Fatalf("unexpected size: %d", size)
	}
	if got := buildEd2kDownloadTaskTitle("主角.mkv"); got != "主角.mkv" {
		t.Fatalf("unexpected fallback title: %s", got)
	}
	if got := buildEd2kDownloadTaskTitle("  "); got != "ED2K 下载任务" {
		t.Fatalf("unexpected default title: %s", got)
	}
}

func TestDeleteFailedEd2kDownloadArtifactsRemovesWorkspaceAndTempFiles(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	downloadRoot := filepath.Join(rootDir, "storage")
	downloadSubdir := "ed2k-downloads"
	baseDownloadDir := filepath.Join(downloadRoot, downloadSubdir)
	logPath := filepath.Join(rootDir, "cancel.log")
	amulecmdPath := filepath.Join(rootDir, "amulecmd")
	if err := os.MkdirAll(baseDownloadDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(download): %v", err)
	}
	if err := os.WriteFile(amulecmdPath, []byte(strings.Join([]string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		"printf '%s\\n' \"$*\" >> \"" + logPath + "\"",
	}, "\n")), 0o755); err != nil {
		t.Fatalf("WriteFile(amulecmd): %v", err)
	}

	taskID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	hash := "ABCDEF0123456789ABCDEF0123456789"
	outputDir := filepath.Join(baseDownloadDir, hash)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(outputDir): %v", err)
	}
	outputFile := filepath.Join(outputDir, "demo.mkv")
	if err := os.WriteFile(outputFile, []byte("demo"), 0o644); err != nil {
		t.Fatalf("WriteFile(output): %v", err)
	}

	task := makeFailedTask(taskID, hash, outputDir, outputFile)
	err := deleteFailedEd2kDownloadArtifacts(task, ed2kDeletePaths{
		downloadRoot:   downloadRoot,
		downloadSubdir: downloadSubdir,
		amulecmdBin:    amulecmdPath,
		remoteHost:     "127.0.0.1",
		remotePort:     "4712",
		remotePassword: "secret",
	})
	if err != nil {
		t.Fatalf("deleteFailedEd2kDownloadArtifacts() error = %v", err)
	}
	assertPathMissing(t, outputDir)
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile(cancel.log): %v", err)
	}
	if !strings.Contains(string(raw), "cancel "+hash) {
		t.Fatalf("expected cancel command to contain resource hash, got %q", string(raw))
	}
}

func TestDeleteFailedEd2kDownloadArtifactsDoesNotDeleteOutsideConfiguredRoots(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	downloadRoot := filepath.Join(rootDir, "storage")
	downloadSubdir := "ed2k-downloads"
	if err := os.MkdirAll(filepath.Join(downloadRoot, downloadSubdir), 0o755); err != nil {
		t.Fatalf("MkdirAll(download): %v", err)
	}

	taskID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	hash := "ABCDEF0123456789ABCDEF0123456789"
	outsideDir := filepath.Join(rootDir, "outside")
	if err := os.MkdirAll(outsideDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(outside): %v", err)
	}
	outsideFile := filepath.Join(outsideDir, "demo.mkv")
	if err := os.WriteFile(outsideFile, []byte("demo"), 0o644); err != nil {
		t.Fatalf("WriteFile(outside): %v", err)
	}

	task := makeFailedTask(taskID, hash, outsideDir, outsideFile)
	err := deleteFailedEd2kDownloadArtifacts(task, ed2kDeletePaths{
		downloadRoot:   downloadRoot,
		downloadSubdir: downloadSubdir,
	})
	if err != nil {
		t.Fatalf("deleteFailedEd2kDownloadArtifacts() error = %v", err)
	}
	if _, statErr := os.Stat(outsideFile); statErr != nil {
		t.Fatalf("expected outside file to remain, got %v", statErr)
	}
}

func TestRemoveFailedEd2kPathIgnoresNotExist(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing")
	if err := removeFailedEd2kPath(path); err != nil {
		t.Fatalf("removeFailedEd2kPath() error = %v", err)
	}
}

func TestCancelFailedEd2kDownloadIgnoresHashNotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "amulecmd")
	if err := os.WriteFile(scriptPath, []byte(strings.Join([]string{
		"#!/usr/bin/env bash",
		"echo 'FileHash not found: ABCDEF0123456789ABCDEF0123456789'",
		"exit 1",
	}, "\n")), 0o755); err != nil {
		t.Fatalf("WriteFile(script): %v", err)
	}

	err := cancelFailedEd2kDownload(makeFailedTask(uuid.New(), "ABCDEF0123456789ABCDEF0123456789", "", ""), ed2kDeletePaths{
		amulecmdBin:    scriptPath,
		remoteHost:     "127.0.0.1",
		remotePort:     "4712",
		remotePassword: "secret",
	})
	if err != nil {
		t.Fatalf("cancelFailedEd2kDownload() error = %v", err)
	}
}

func TestBuildEd2kCreateResultsCoversDuplicateInvalidReuseAndEnqueueFailure(t *testing.T) {
	t.Parallel()

	existingID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo := &ed2kCreateRepoStub{
		byHash: map[string]models.AdminEd2kDownloadTask{
			"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA": {
				ID:           existingID,
				ResourceHash: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
				Status:       "completed",
			},
		},
		createErrByHash: map[string]error{
			"CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC": errors.New("db create failed"),
		},
		createTaskIDByHash: map[string]uuid.UUID{
			"DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD": uuid.MustParse("44444444-4444-4444-4444-444444444444"),
		},
	}
	enqueuer := &ed2kCreateEnqueuerStub{
		errByTaskID: map[string]error{
			"44444444-4444-4444-4444-444444444444": errors.New("enqueue boom"),
		},
	}
	entries := []ed2kCreateLine{
		{lineNumber: 7, sourceLink: "ed2k://|file|first.mkv|1|AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA|/"},
		{lineNumber: 8, sourceLink: "bad-link"},
		{lineNumber: 9, sourceLink: "ed2k://|file|dup.mkv|1|BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB|/"},
		{lineNumber: 10, sourceLink: "ed2k://|file|dup.mkv|1|BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB|/"},
		{lineNumber: 11, sourceLink: "ed2k://|file|create-fail.mkv|1|CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC|/"},
		{lineNumber: 12, sourceLink: "ed2k://|file|enqueue-fail.mkv|1|DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD|/"},
	}

	results, err := buildEd2kCreateResults(context.Background(), repo, enqueuer, entries)
	if err != nil {
		t.Fatalf("buildEd2kCreateResults() error = %v", err)
	}
	if got := collectCreateStatuses(results); got != "7:reused,8:invalid,9:created,10:duplicate,11:create_failed,12:enqueue_failed" {
		t.Fatalf("unexpected statuses: %s", got)
	}
	if repo.appendHistoryCalls != 1 {
		t.Fatalf("expected one history append, got %d", repo.appendHistoryCalls)
	}
	if results[0].Task == nil || results[0].Task.ID != existingID {
		t.Fatalf("expected reused task target, got %#v", results[0].Task)
	}
	if results[5].Task == nil || results[5].Task.Status != "failed" {
		t.Fatalf("expected enqueue_failed result to carry failed task, got %#v", results[5].Task)
	}
}

func TestRetryEd2kDownloadTaskRequeuesFilesCleanedTask(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	repo := &ed2kRetryRepoStub{
		task: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "files_cleaned",
		},
		requeued: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "queued",
		},
	}
	enqueuer := &ed2kCreateEnqueuerStub{errByTaskID: map[string]error{}}

	item, err := retryEd2kDownloadTask(context.Background(), taskID, repo, enqueuer, time.Now())
	if err != nil {
		t.Fatalf("retryEd2kDownloadTask() error = %v", err)
	}
	if item.Status != "queued" {
		t.Fatalf("expected queued task, got %s", item.Status)
	}
	if repo.requeueCalls != 1 {
		t.Fatalf("expected one requeue call, got %d", repo.requeueCalls)
	}
	if enqueuer.calls != 1 {
		t.Fatalf("expected one enqueue call, got %d", enqueuer.calls)
	}
}

func TestRetryEd2kDownloadTaskRollsBackRequeuedStateWhenEnqueueFails(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("23222222-2222-2222-2222-222222222222")
	repo := &ed2kRetryRepoStub{
		task: models.AdminEd2kDownloadTask{
			ID:        taskID,
			Status:    "files_cleaned",
			StartedAt: timePtrNow(),
			CleanedAt: timePtrNow(),
		},
		requeued: models.AdminEd2kDownloadTask{ID: taskID, Status: "queued"},
	}
	enqueuer := &ed2kCreateEnqueuerStub{errByTaskID: map[string]error{
		taskID.String(): errors.New("enqueue boom"),
	}}

	_, err := retryEd2kDownloadTask(context.Background(), taskID, repo, enqueuer, time.Now())
	if err == nil || !strings.Contains(err.Error(), "enqueue boom") {
		t.Fatalf("expected enqueue error, got %v", err)
	}
	if enqueuer.calls != 1 {
		t.Fatalf("expected one enqueue call, got %d", enqueuer.calls)
	}
	if repo.restoreCalls != 1 {
		t.Fatalf("expected one restore call, got %d", repo.restoreCalls)
	}
}

func TestRetryEd2kDownloadTaskReturnsCurrentTaskWhenWorkerAlreadyTookOver(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("24222222-2222-2222-2222-222222222222")
	repo := &ed2kRetryRepoStub{
		task: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "files_cleaned",
		},
		refreshedTask: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "running",
		},
		requeueErr: errors.New("requeue files_cleaned ed2k download task: no rows in result set"),
	}
	enqueuer := &ed2kCreateEnqueuerStub{errByTaskID: map[string]error{}}

	item, err := retryEd2kDownloadTask(context.Background(), taskID, repo, enqueuer, time.Now())
	if err != nil {
		t.Fatalf("retryEd2kDownloadTask() error = %v", err)
	}
	if item.Status != "running" {
		t.Fatalf("expected running task, got %s", item.Status)
	}
	if enqueuer.inspectedTaskID != "" {
		t.Fatalf("did not expect queue inspection for running task, got %s", enqueuer.inspectedTaskID)
	}
}

func TestRetryEd2kDownloadTaskReturnsQueuedOnlyWhenQueueStillHasTask(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("25222222-2222-2222-2222-222222222222")
	repo := &ed2kRetryRepoStub{
		task: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "files_cleaned",
		},
		refreshedTask: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "queued",
		},
		requeueErr: errors.New("requeue files_cleaned ed2k download task: no rows in result set"),
	}
	enqueuer := &ed2kCreateEnqueuerStub{
		errByTaskID: map[string]error{},
		hasTaskByID: map[string]bool{taskID.String(): true},
	}

	item, err := retryEd2kDownloadTask(context.Background(), taskID, repo, enqueuer, time.Now())
	if err != nil {
		t.Fatalf("retryEd2kDownloadTask() error = %v", err)
	}
	if item.Status != "queued" {
		t.Fatalf("expected queued task, got %s", item.Status)
	}
	if enqueuer.inspectedTaskID != taskID.String() {
		t.Fatalf("expected queue inspection for %s, got %s", taskID.String(), enqueuer.inspectedTaskID)
	}
}

func TestRetryEd2kDownloadTaskTreatsEnqueueErrorWithPersistedJobAsSuccess(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("26222222-2222-2222-2222-222222222222")
	repo := &ed2kRetryRepoStub{
		task: models.AdminEd2kDownloadTask{
			ID:        taskID,
			Status:    "files_cleaned",
			StartedAt: timePtrNow(),
			CleanedAt: timePtrNow(),
		},
		requeued: models.AdminEd2kDownloadTask{ID: taskID, Status: "queued"},
		refreshedTask: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "queued",
		},
	}
	enqueuer := &ed2kCreateEnqueuerStub{
		errByTaskID: map[string]error{taskID.String(): errors.New("transient enqueue error")},
		hasTaskByID: map[string]bool{taskID.String(): true},
	}

	item, err := retryEd2kDownloadTask(context.Background(), taskID, repo, enqueuer, time.Now())
	if err != nil {
		t.Fatalf("retryEd2kDownloadTask() error = %v", err)
	}
	if item.Status != "queued" {
		t.Fatalf("expected queued task, got %s", item.Status)
	}
	if repo.restoreCalls != 0 {
		t.Fatalf("did not expect rollback when job already exists, got %d", repo.restoreCalls)
	}
}

func TestDeleteQueuedEd2kDownloadTaskDeletesOrdinaryQueuedTaskAndQueueJob(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("27222222-2222-2222-2222-222222222222")
	repo := &ed2kQueuedDeleteRepoStub{}
	enqueuer := &ed2kQueuedDeleteEnqueuerStub{}

	item, deleted, err := deleteQueuedEd2kDownloadTask(context.Background(), models.AdminEd2kDownloadTask{
		ID:     taskID,
		Status: "queued",
	}, repo, enqueuer, time.Now())
	if err != nil {
		t.Fatalf("deleteQueuedEd2kDownloadTask() error = %v", err)
	}
	if !deleted {
		t.Fatal("expected ordinary queued task to be physically deleted")
	}
	if item.ID != uuid.Nil {
		t.Fatalf("expected no restored task payload, got %#v", item)
	}
	if repo.deleteCalls != 1 {
		t.Fatalf("expected one repo delete call, got %d", repo.deleteCalls)
	}
	if enqueuer.deletedTaskID != taskID.String() {
		t.Fatalf("expected queue delete for %s, got %s", taskID.String(), enqueuer.deletedTaskID)
	}
}

func TestDeleteQueuedEd2kDownloadTaskRestoresRetriedFilesCleanedHistory(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("28222222-2222-2222-2222-222222222222")
	startedAt := timePtrNow()
	cleanedAt := timePtrNow()
	finishedAt := timePtrNow()
	repo := &ed2kQueuedDeleteRepoStub{
		restored: models.AdminEd2kDownloadTask{
			ID:         taskID,
			Status:     "files_cleaned",
			StartedAt:  startedAt,
			FinishedAt: finishedAt,
			CleanedAt:  cleanedAt,
		},
	}
	enqueuer := &ed2kQueuedDeleteEnqueuerStub{}

	item, deleted, err := deleteQueuedEd2kDownloadTask(context.Background(), models.AdminEd2kDownloadTask{
		ID:         taskID,
		Status:     "queued",
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		CleanedAt:  cleanedAt,
	}, repo, enqueuer, time.Now())
	if err != nil {
		t.Fatalf("deleteQueuedEd2kDownloadTask() error = %v", err)
	}
	if deleted {
		t.Fatal("expected retried files_cleaned task to be restored instead of deleted")
	}
	if item.Status != "files_cleaned" {
		t.Fatalf("expected files_cleaned status, got %s", item.Status)
	}
	if repo.restoreCalls != 1 {
		t.Fatalf("expected one restore call, got %d", repo.restoreCalls)
	}
	if repo.deleteCalls != 0 {
		t.Fatalf("did not expect physical delete, got %d", repo.deleteCalls)
	}
	if enqueuer.deletedTaskID != taskID.String() {
		t.Fatalf("expected queue delete for %s, got %s", taskID.String(), enqueuer.deletedTaskID)
	}
}

func TestDeleteQueuedEd2kDownloadTaskReturnsStaleStateWhenRestoreMisses(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("2c222222-2222-2222-2222-222222222222")
	startedAt := timePtrNow()
	cleanedAt := timePtrNow()
	repo := &ed2kQueuedDeleteRepoStub{
		restoreErr: errors.New("restore files_cleaned ed2k download task: no rows in result set"),
	}
	enqueuer := &ed2kQueuedDeleteEnqueuerStub{}

	_, deleted, err := deleteQueuedEd2kDownloadTask(context.Background(), models.AdminEd2kDownloadTask{
		ID:         taskID,
		Status:     "queued",
		StartedAt:  startedAt,
		FinishedAt: timePtrNow(),
		CleanedAt:  cleanedAt,
	}, repo, enqueuer, time.Now())
	if !errors.Is(err, errEd2kQueuedDeleteStaleState) {
		t.Fatalf("expected stale state error, got %v", err)
	}
	if deleted {
		t.Fatal("did not expect deleted=true on stale restore")
	}
	if enqueuer.enqueuedTaskID != "" {
		t.Fatalf("did not expect queue job restore on stale rollback, got %s", enqueuer.enqueuedTaskID)
	}
}

func TestDeleteQueuedEd2kDownloadTaskRejectsWhenQueueJobAlreadyActive(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("29222222-2222-2222-2222-222222222222")
	repo := &ed2kQueuedDeleteRepoStub{}
	enqueuer := &ed2kQueuedDeleteEnqueuerStub{
		deleteErr: errors.New("delete enqueued ed2k download task: asynq: cannot delete task in active state. use CancelProcessing instead."),
	}

	_, deleted, err := deleteQueuedEd2kDownloadTask(context.Background(), models.AdminEd2kDownloadTask{
		ID:     taskID,
		Status: "queued",
	}, repo, enqueuer, time.Now())
	if !errors.Is(err, errEd2kQueuedDeleteActiveJob) {
		t.Fatalf("expected active job error, got %v", err)
	}
	if deleted {
		t.Fatal("did not expect deleted=true on active job rejection")
	}
	if repo.deleteCalls != 0 || repo.restoreCalls != 0 {
		t.Fatalf("did not expect repo mutation, got delete=%d restore=%d", repo.deleteCalls, repo.restoreCalls)
	}
}

func TestDeleteQueuedEd2kDownloadTaskRequeuesJobWhenRepoDeleteFails(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("2a222222-2222-2222-2222-222222222222")
	repo := &ed2kQueuedDeleteRepoStub{
		deleteErr: errors.New("delete queued ed2k download task: db boom"),
		current: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "queued",
		},
	}
	enqueuer := &ed2kQueuedDeleteEnqueuerStub{}

	_, deleted, err := deleteQueuedEd2kDownloadTask(context.Background(), models.AdminEd2kDownloadTask{
		ID:     taskID,
		Status: "queued",
	}, repo, enqueuer, time.Now())
	if err == nil || !strings.Contains(err.Error(), "db boom") {
		t.Fatalf("expected repo delete error, got %v", err)
	}
	if deleted {
		t.Fatal("did not expect deleted=true when repo delete fails")
	}
	if enqueuer.deletedTaskID != taskID.String() {
		t.Fatalf("expected queue delete for %s, got %s", taskID.String(), enqueuer.deletedTaskID)
	}
	if enqueuer.enqueuedTaskID != taskID.String() {
		t.Fatalf("expected queue job restore for %s, got %s", taskID.String(), enqueuer.enqueuedTaskID)
	}
}

func TestDeleteQueuedEd2kDownloadTaskReturnsStaleStateWhenQueuedDeleteMisses(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("2b222222-2222-2222-2222-222222222222")
	repo := &ed2kQueuedDeleteRepoStub{
		deleteErr: errors.New("delete queued ed2k download task: no rows in result set"),
	}
	enqueuer := &ed2kQueuedDeleteEnqueuerStub{}

	_, deleted, err := deleteQueuedEd2kDownloadTask(context.Background(), models.AdminEd2kDownloadTask{
		ID:     taskID,
		Status: "queued",
	}, repo, enqueuer, time.Now())
	if !errors.Is(err, errEd2kQueuedDeleteStaleState) {
		t.Fatalf("expected stale state error, got %v", err)
	}
	if deleted {
		t.Fatal("did not expect deleted=true on stale queued delete")
	}
	if enqueuer.enqueuedTaskID != "" {
		t.Fatalf("did not expect queue job restore on stale delete, got %s", enqueuer.enqueuedTaskID)
	}
}

func TestCleanCompletedEd2kDownloadArtifactsRemovesSymlinkAndRealFile(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	downloadRoot := filepath.Join(rootDir, "storage")
	downloadSubdir := "ed2k-downloads"
	outputDir := filepath.Join(downloadRoot, downloadSubdir, "task-1")
	realFile := filepath.Join(downloadRoot, downloadSubdir, "done.mkv")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir output: %v", err)
	}
	if err := os.WriteFile(realFile, []byte("demo"), 0o644); err != nil {
		t.Fatalf("write real file: %v", err)
	}
	if err := os.Symlink(realFile, filepath.Join(outputDir, "done.mkv")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	taskID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	task := models.AdminEd2kDownloadTask{
		ID:         taskID,
		Status:     "completed",
		OutputDir:  outputDir,
		Filename:   "done.mkv",
		FinishedAt: timePtrNow(),
		Files: []models.AdminEd2kDownloadTaskFile{
			{Name: "done.mkv", Path: "done.mkv", Size: 4},
		},
	}
	if err := cleanCompletedEd2kDownloadArtifacts(task, ed2kDeletePaths{
		downloadRoot:   downloadRoot,
		downloadSubdir: downloadSubdir,
	}); err != nil {
		t.Fatalf("cleanCompletedEd2kDownloadArtifacts() error = %v", err)
	}
	assertPathMissing(t, realFile)
	assertPathMissing(t, outputDir)
}

func TestCleanCompletedEd2kDownloadArtifactsRejectsSymlinkedParentDirectoryEscape(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	downloadRoot := filepath.Join(rootDir, "storage")
	downloadSubdir := "ed2k-downloads"
	outputDir := filepath.Join(downloadRoot, downloadSubdir, "task-escape")
	outsideDir := filepath.Join(rootDir, "outside")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir output: %v", err)
	}
	if err := os.MkdirAll(outsideDir, 0o755); err != nil {
		t.Fatalf("mkdir outside: %v", err)
	}
	if err := os.Symlink(outsideDir, filepath.Join(outputDir, "nested")); err != nil {
		t.Fatalf("symlink nested dir: %v", err)
	}
	outsideFile := filepath.Join(outsideDir, "done.mkv")
	if err := os.WriteFile(outsideFile, []byte("demo"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}

	task := models.AdminEd2kDownloadTask{
		ID:         uuid.MustParse("34333333-3333-3333-3333-333333333333"),
		Status:     "completed",
		OutputDir:  outputDir,
		Filename:   "done.mkv",
		FinishedAt: timePtrNow(),
		Files: []models.AdminEd2kDownloadTaskFile{
			{Name: "done.mkv", Path: "nested/done.mkv", Size: 4},
		},
	}
	err := cleanCompletedEd2kDownloadArtifacts(task, ed2kDeletePaths{
		downloadRoot:   downloadRoot,
		downloadSubdir: downloadSubdir,
	})
	if err == nil || !strings.Contains(err.Error(), "父目录越出任务目录") {
		t.Fatalf("expected parent directory escape error, got %v", err)
	}
	if _, statErr := os.Stat(outsideFile); statErr != nil {
		t.Fatalf("expected outside file to remain, got %v", statErr)
	}
}

func TestCleanCompletedEd2kDownloadArtifactsRejectsSymlinkedTaskRootEscape(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	downloadRoot := filepath.Join(rootDir, "storage")
	downloadSubdir := "ed2k-downloads"
	managedRoot := filepath.Join(downloadRoot, downloadSubdir)
	realOutsideDir := filepath.Join(rootDir, "outside-real")
	if err := os.MkdirAll(managedRoot, 0o755); err != nil {
		t.Fatalf("mkdir managed root: %v", err)
	}
	if err := os.MkdirAll(realOutsideDir, 0o755); err != nil {
		t.Fatalf("mkdir real outside dir: %v", err)
	}
	symlinkedTaskDir := filepath.Join(managedRoot, "task-link")
	if err := os.Symlink(realOutsideDir, symlinkedTaskDir); err != nil {
		t.Fatalf("symlink task dir: %v", err)
	}
	outsideFile := filepath.Join(realOutsideDir, "done.mkv")
	if err := os.WriteFile(outsideFile, []byte("demo"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}

	task := models.AdminEd2kDownloadTask{
		ID:         uuid.MustParse("35333333-3333-3333-3333-333333333333"),
		Status:     "completed",
		OutputDir:  symlinkedTaskDir,
		Filename:   "done.mkv",
		FinishedAt: timePtrNow(),
		Files: []models.AdminEd2kDownloadTaskFile{
			{Name: "done.mkv", Path: "done.mkv", Size: 4},
		},
	}
	err := cleanCompletedEd2kDownloadArtifacts(task, ed2kDeletePaths{
		downloadRoot:   downloadRoot,
		downloadSubdir: downloadSubdir,
	})
	if err == nil || !strings.Contains(err.Error(), "任务目录超出受管目录") {
		t.Fatalf("expected symlinked task root escape error, got %v", err)
	}
	if _, statErr := os.Stat(outsideFile); statErr != nil {
		t.Fatalf("expected outside file to remain, got %v", statErr)
	}
}

func TestRetryCancelledEd2kDownloadCleanupSucceeds(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	downloadRoot := filepath.Join(rootDir, "storage")
	downloadSubdir := "ed2k-downloads"
	outputDir := filepath.Join(downloadRoot, downloadSubdir, "task-cancelled")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir output: %v", err)
	}
	residueFile := filepath.Join(outputDir, "part.tmp")
	if err := os.WriteFile(residueFile, []byte("tmp"), 0o644); err != nil {
		t.Fatalf("write residue file: %v", err)
	}
	taskID := uuid.MustParse("36333333-3333-3333-3333-333333333333")
	finishedAt := timePtrNow()
	repo := &ed2kCleanupRepoStub{
		task: models.AdminEd2kDownloadTask{
			ID:           taskID,
			Status:       "cancelled",
			ErrorMessage: "仍有残留待清理",
			OutputDir:    outputDir,
			FinishedAt:   finishedAt,
		},
		updated: models.AdminEd2kDownloadTask{
			ID:         taskID,
			Status:     "cancelled",
			FinishedAt: finishedAt,
		},
	}

	item, err := retryCancelledEd2kDownloadCleanup(context.Background(), taskID, repo, ed2kDeletePaths{
		downloadRoot:   downloadRoot,
		downloadSubdir: downloadSubdir,
	}, time.Now())
	if err != nil {
		t.Fatalf("retryCancelledEd2kDownloadCleanup() error = %v", err)
	}
	if item.Status != "cancelled" {
		t.Fatalf("expected cancelled status, got %s", item.Status)
	}
	if repo.updateCalls != 1 {
		t.Fatalf("expected one update call, got %d", repo.updateCalls)
	}
	assertPathMissing(t, outputDir)
}

func TestCancelEd2kDownloadTaskLeavesTaskCancelingWhenEngineStopFails(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	downloadRoot := filepath.Join(rootDir, "storage")
	downloadSubdir := "ed2k-downloads"
	outputDir := filepath.Join(downloadRoot, downloadSubdir, "task-running")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir output: %v", err)
	}
	residueFile := filepath.Join(outputDir, "part.tmp")
	if err := os.WriteFile(residueFile, []byte("tmp"), 0o644); err != nil {
		t.Fatalf("write residue file: %v", err)
	}
	amulecmdPath := filepath.Join(rootDir, "amulecmd")
	if err := os.WriteFile(amulecmdPath, []byte(strings.Join([]string{
		"#!/usr/bin/env bash",
		"echo 'cancel failed'",
		"exit 1",
	}, "\n")), 0o755); err != nil {
		t.Fatalf("write amulecmd: %v", err)
	}
	taskID := uuid.MustParse("37333333-3333-3333-3333-333333333333")
	task := models.AdminEd2kDownloadTask{
		ID:           taskID,
		Status:       "running",
		ResourceHash: "ABCDEF0123456789ABCDEF0123456789",
		OutputDir:    outputDir,
	}
	repo := &ed2kCancelRepoStub{
		cancelingItem: models.AdminEd2kDownloadTask{
			ID:           taskID,
			Status:       "canceling",
			ResourceHash: task.ResourceHash,
			OutputDir:    outputDir,
		},
		cancellationFailedItem: models.AdminEd2kDownloadTask{
			ID:           taskID,
			Status:       "canceling",
			ResourceHash: task.ResourceHash,
			OutputDir:    outputDir,
			ErrorMessage: "取消 aMule 下载失败: cancel failed",
		},
	}

	item, err := cancelEd2kDownloadTask(context.Background(), task, repo, ed2kDeletePaths{
		downloadRoot:   downloadRoot,
		downloadSubdir: downloadSubdir,
		amulecmdBin:    amulecmdPath,
		remoteHost:     "127.0.0.1",
		remotePort:     "4712",
		remotePassword: "secret",
	}, time.Now())
	if err != nil {
		t.Fatalf("cancelEd2kDownloadTask() error = %v", err)
	}
	if item.Status != "canceling" {
		t.Fatalf("expected canceling status, got %s", item.Status)
	}
	if repo.cancelingCalls != 1 {
		t.Fatalf("expected one canceling transition, got %d", repo.cancelingCalls)
	}
	if repo.cancellationFailedCalls != 1 {
		t.Fatalf("expected one cancellation failure transition, got %d", repo.cancellationFailedCalls)
	}
	if repo.cancelledCalls != 0 {
		t.Fatalf("did not expect cancelled transition, got %d", repo.cancelledCalls)
	}
	if _, err := os.Stat(residueFile); err != nil {
		t.Fatalf("expected residue file to remain after cancel failure, got %v", err)
	}
}

func TestCancelEd2kDownloadTaskCanRetryExistingCancelingTask(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	downloadRoot := filepath.Join(rootDir, "storage")
	downloadSubdir := "ed2k-downloads"
	outputDir := filepath.Join(downloadRoot, downloadSubdir, "task-canceling")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir output: %v", err)
	}
	residueFile := filepath.Join(outputDir, "part.tmp")
	if err := os.WriteFile(residueFile, []byte("tmp"), 0o644); err != nil {
		t.Fatalf("write residue file: %v", err)
	}
	amulecmdPath := filepath.Join(rootDir, "amulecmd")
	if err := os.WriteFile(amulecmdPath, []byte(strings.Join([]string{
		"#!/usr/bin/env bash",
		"exit 0",
	}, "\n")), 0o755); err != nil {
		t.Fatalf("write amulecmd: %v", err)
	}
	taskID := uuid.MustParse("38333333-3333-3333-3333-333333333333")
	task := models.AdminEd2kDownloadTask{
		ID:           taskID,
		Status:       "canceling",
		ResourceHash: "ABCDEF0123456789ABCDEF0123456789",
		OutputDir:    outputDir,
	}
	repo := &ed2kCancelRepoStub{
		cancelledItem: models.AdminEd2kDownloadTask{
			ID:         taskID,
			Status:     "cancelled",
			OutputDir:  outputDir,
			FinishedAt: timePtrNow(),
			CleanedAt:  timePtrNow(),
		},
	}

	item, err := cancelEd2kDownloadTask(context.Background(), task, repo, ed2kDeletePaths{
		downloadRoot:   downloadRoot,
		downloadSubdir: downloadSubdir,
		amulecmdBin:    amulecmdPath,
		remoteHost:     "127.0.0.1",
		remotePort:     "4712",
		remotePassword: "secret",
	}, time.Now())
	if err != nil {
		t.Fatalf("cancelEd2kDownloadTask() error = %v", err)
	}
	if item.Status != "cancelled" {
		t.Fatalf("expected cancelled status, got %s", item.Status)
	}
	if repo.cancelingCalls != 0 {
		t.Fatalf("did not expect running->canceling transition, got %d", repo.cancelingCalls)
	}
	if repo.cancelledCalls != 1 {
		t.Fatalf("expected one cancelled transition, got %d", repo.cancelledCalls)
	}
	if repo.lastCancelledCleanedAt == nil {
		t.Fatal("expected cleaned_at to be recorded on successful cancel cleanup")
	}
	assertPathMissing(t, outputDir)
}

func makeFailedTask(taskID uuid.UUID, hash, outputDir, downloadedPath string) models.AdminEd2kDownloadTask {
	return models.AdminEd2kDownloadTask{
		ID:             taskID,
		Status:         "failed",
		ResourceHash:   hash,
		Filename:       "demo.mkv",
		OutputDir:      outputDir,
		DownloadedPath: downloadedPath,
	}
}

func collectCreateStatuses(results []models.AdminEd2kDownloadCreateResult) string {
	parts := make([]string, 0, len(results))
	for _, item := range results {
		parts = append(parts, strconv.Itoa(item.LineNumber)+":"+item.Status)
	}
	return strings.Join(parts, ",")
}

func timePtrNow() *time.Time {
	now := time.Now()
	return &now
}

type ed2kCreateRepoStub struct {
	byHash             map[string]models.AdminEd2kDownloadTask
	createErrByHash    map[string]error
	createTaskIDByHash map[string]uuid.UUID
	appendHistoryCalls int
}

func (s *ed2kCreateRepoStub) GetEd2kDownloadTaskByHash(_ context.Context, resourceHash string) (models.AdminEd2kDownloadTask, error) {
	if task, ok := s.byHash[resourceHash]; ok {
		return task, nil
	}
	return models.AdminEd2kDownloadTask{}, pgx.ErrNoRows
}

func (s *ed2kCreateRepoStub) AppendEd2kDownloadTaskHistory(context.Context, uuid.UUID, models.AdminEd2kDownloadTaskHistoryItem) error {
	s.appendHistoryCalls++
	return nil
}

func (s *ed2kCreateRepoStub) CreateEd2kDownloadTask(_ context.Context, input models.AdminEd2kDownloadTask) (models.AdminEd2kDownloadTask, error) {
	if err := s.createErrByHash[input.ResourceHash]; err != nil {
		return models.AdminEd2kDownloadTask{}, err
	}
	if taskID, ok := s.createTaskIDByHash[input.ResourceHash]; ok {
		input.ID = taskID
	}
	return input, nil
}

func (s *ed2kCreateRepoStub) UpdateEd2kDownloadTaskStatus(_ context.Context, id uuid.UUID, status, progressText, errorMessage string, startedAt, finishedAt, deletedAt *time.Time, outputDir, downloadedPath string, files []models.AdminEd2kDownloadTaskFile, retryDelta int, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	return models.AdminEd2kDownloadTask{
		ID:             id,
		Status:         status,
		ProgressText:   progressText,
		ErrorMessage:   errorMessage,
		OutputDir:      outputDir,
		DownloadedPath: downloadedPath,
		Files:          files,
	}, nil
}

type ed2kCreateEnqueuerStub struct {
	calls           int
	errByTaskID     map[string]error
	hasTaskByID     map[string]bool
	inspectedTaskID string
}

func (s *ed2kCreateEnqueuerStub) EnqueueEd2kDownload(payload queue.Ed2kDownloadPayload) error {
	s.calls++
	return s.errByTaskID[payload.TaskID]
}

func (s *ed2kCreateEnqueuerStub) HasEd2kDownloadTask(taskID string) (bool, error) {
	s.inspectedTaskID = taskID
	return s.hasTaskByID[taskID], nil
}

type ed2kRetryRepoStub struct {
	task          models.AdminEd2kDownloadTask
	refreshedTask models.AdminEd2kDownloadTask
	requeued      models.AdminEd2kDownloadTask
	restored      models.AdminEd2kDownloadTask
	requeueCalls  int
	requeueErr    error
	restoreCalls  int
	getCalls      int
}

func (s *ed2kRetryRepoStub) GetEd2kDownloadTask(context.Context, uuid.UUID) (models.AdminEd2kDownloadTask, error) {
	s.getCalls++
	if s.getCalls > 1 && s.refreshedTask.ID != uuid.Nil {
		return s.refreshedTask, nil
	}
	return s.task, nil
}

func (s *ed2kRetryRepoStub) UpdateEd2kDownloadTaskStatus(context.Context, uuid.UUID, string, string, string, *time.Time, *time.Time, *time.Time, string, string, []models.AdminEd2kDownloadTaskFile, int, models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	return models.AdminEd2kDownloadTask{}, nil
}

func (s *ed2kRetryRepoStub) RequeueFilesCleanedEd2kDownloadTask(context.Context, uuid.UUID, models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	s.requeueCalls++
	if s.requeueErr != nil {
		return models.AdminEd2kDownloadTask{}, s.requeueErr
	}
	return s.requeued, nil
}

func (s *ed2kRetryRepoStub) RestoreFilesCleanedEd2kDownloadTask(context.Context, uuid.UUID, models.AdminEd2kDownloadTaskHistoryItem, *time.Time, *time.Time) (models.AdminEd2kDownloadTask, error) {
	s.restoreCalls++
	if s.restored.ID != uuid.Nil {
		return s.restored, nil
	}
	return s.task, nil
}

type ed2kQueuedDeleteRepoStub struct {
	deleteCalls  int
	deleteErr    error
	restoreCalls int
	restored     models.AdminEd2kDownloadTask
	restoreErr   error
	current      models.AdminEd2kDownloadTask
	getCalls     int
}

func (s *ed2kQueuedDeleteRepoStub) GetEd2kDownloadTask(context.Context, uuid.UUID) (models.AdminEd2kDownloadTask, error) {
	s.getCalls++
	if s.current.ID != uuid.Nil {
		return s.current, nil
	}
	return models.AdminEd2kDownloadTask{}, pgx.ErrNoRows
}

func (s *ed2kQueuedDeleteRepoStub) DeleteQueuedEd2kDownloadTask(context.Context, uuid.UUID) error {
	s.deleteCalls++
	return s.deleteErr
}

func (s *ed2kQueuedDeleteRepoStub) RestoreFilesCleanedEd2kDownloadTask(context.Context, uuid.UUID, models.AdminEd2kDownloadTaskHistoryItem, *time.Time, *time.Time) (models.AdminEd2kDownloadTask, error) {
	s.restoreCalls++
	if s.restoreErr != nil {
		return models.AdminEd2kDownloadTask{}, s.restoreErr
	}
	return s.restored, nil
}

type ed2kQueuedDeleteEnqueuerStub struct {
	deletedTaskID  string
	deleteErr      error
	enqueuedTaskID string
	enqueueErr     error
}

func (s *ed2kQueuedDeleteEnqueuerStub) DeleteEd2kDownloadTask(taskID string) error {
	s.deletedTaskID = taskID
	return s.deleteErr
}

func (s *ed2kQueuedDeleteEnqueuerStub) EnqueueEd2kDownload(payload queue.Ed2kDownloadPayload) error {
	s.enqueuedTaskID = payload.TaskID
	return s.enqueueErr
}

type ed2kCleanupRepoStub struct {
	task        models.AdminEd2kDownloadTask
	updated     models.AdminEd2kDownloadTask
	updateCalls int
}

func (s *ed2kCleanupRepoStub) GetEd2kDownloadTask(context.Context, uuid.UUID) (models.AdminEd2kDownloadTask, error) {
	return s.task, nil
}

func (s *ed2kCleanupRepoStub) MarkEd2kDownloadTaskCancelledCleanupResolved(context.Context, uuid.UUID, *time.Time, models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	s.updateCalls++
	return s.updated, nil
}

type ed2kCancelRepoStub struct {
	cancelingItem           models.AdminEd2kDownloadTask
	cancellationFailedItem  models.AdminEd2kDownloadTask
	cancelledItem           models.AdminEd2kDownloadTask
	cancelingCalls          int
	cancellationFailedCalls int
	cancelledCalls          int
	lastCancelledFinishedAt *time.Time
	lastCancelledCleanedAt  *time.Time
}

func (s *ed2kCancelRepoStub) MarkEd2kDownloadTaskCanceling(context.Context, uuid.UUID, string, models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	s.cancelingCalls++
	return s.cancelingItem, nil
}

func (s *ed2kCancelRepoStub) MarkEd2kDownloadTaskCancellationFailed(context.Context, uuid.UUID, string, string, models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	s.cancellationFailedCalls++
	return s.cancellationFailedItem, nil
}

func (s *ed2kCancelRepoStub) MarkEd2kDownloadTaskCancelled(_ context.Context, _ uuid.UUID, _ string, _ string, finishedAt, cleanedAt *time.Time, _ models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	s.cancelledCalls++
	s.lastCancelledFinishedAt = finishedAt
	s.lastCancelledCleanedAt = cleanedAt
	return s.cancelledItem, nil
}

func assertPathMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected %s to be removed, got err=%v", path, err)
	}
}
