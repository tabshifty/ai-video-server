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

	"video-server/internal/cloud"
	"video-server/internal/models"
)

type fakeCloudFileDeletionTasks struct {
	task    models.CloudDownloadTask
	getErr  error
	setIDs  bool
	setCall int
}

func (f *fakeCloudFileDeletionTasks) GetByID(_ context.Context, _ int64) (models.CloudDownloadTask, error) {
	if f.getErr != nil {
		return models.CloudDownloadTask{}, f.getErr
	}
	return f.task, nil
}

func (f *fakeCloudFileDeletionTasks) SetCloudFileIDs(_ context.Context, _ int64, fileID, parentID, userID string, category int) (bool, error) {
	f.setCall++
	f.task.CloudFileID = fileID
	f.task.CloudParentID = parentID
	f.task.CloudUserID = userID
	f.task.CloudFileCategory = category
	return f.setIDs, nil
}

type fakeCloudFileDeletionImports struct {
	item      models.CloudDownloadImport
	err       error
	callCount int
}

func (f *fakeCloudFileDeletionImports) GetByTaskID(_ context.Context, _ int64) (models.CloudDownloadImport, error) {
	f.callCount++
	return f.item, f.err
}

type fakeCloudFileDeletionRecords struct {
	item         models.CloudDownloadFileDeletion
	startCalls   int
	markedDelete int
	markedFail   int
}

func (f *fakeCloudFileDeletionRecords) GetByTaskID(_ context.Context, _ int64) (models.CloudDownloadFileDeletion, error) {
	if f.item.TaskID == 0 {
		return models.CloudDownloadFileDeletion{}, models.ErrCloudFileDeletionNotFound
	}
	return f.item, nil
}

func (f *fakeCloudFileDeletionRecords) Start(_ context.Context, taskID int64, fileID, parentID string, requestedBy *uuid.UUID) (models.CloudDownloadFileDeletion, bool, error) {
	f.startCalls++
	if f.item.TaskID == 0 {
		f.item = models.CloudDownloadFileDeletion{
			TaskID:      taskID,
			FileID:      fileID,
			ParentID:    parentID,
			Status:      models.CloudFileDeleting,
			Attempt:     1,
			RequestedBy: requestedBy,
			RequestedAt: time.Now(),
			UpdatedAt:   time.Now(),
		}
		return f.item, true, nil
	}
	if f.item.Status == models.CloudFileDeleted {
		return f.item, false, nil
	}
	if f.item.Status == models.CloudFileDeleting && time.Since(f.item.UpdatedAt) < cloudFileDeletionStaleAfter {
		return f.item, false, nil
	}
	f.item.Status = models.CloudFileDeleting
	f.item.Attempt++
	f.item.FileID = fileID
	f.item.ParentID = parentID
	f.item.UpdatedAt = time.Now()
	f.item.FinishedAt = nil
	f.item.ErrorMessage = ""
	f.item.Note = ""
	return f.item, true, nil
}

func (f *fakeCloudFileDeletionRecords) SetFileIDs(_ context.Context, _ int64, _ int, fileID, parentID string) (bool, error) {
	if f.item.Status != models.CloudFileDeleting {
		return false, nil
	}
	f.item.FileID = fileID
	f.item.ParentID = parentID
	return true, nil
}

func (f *fakeCloudFileDeletionRecords) MarkDeleted(_ context.Context, _ int64, _ int, note string) (bool, error) {
	if f.item.Status != models.CloudFileDeleting {
		return false, nil
	}
	f.markedDelete++
	f.item.Status = models.CloudFileDeleted
	f.item.Note = note
	now := time.Now()
	f.item.FinishedAt = &now
	f.item.UpdatedAt = now
	return true, nil
}

func (f *fakeCloudFileDeletionRecords) MarkFailed(_ context.Context, _ int64, _ int, message string) (bool, error) {
	if f.item.Status != models.CloudFileDeleting {
		return false, nil
	}
	f.markedFail++
	f.item.Status = models.CloudFileFailed
	f.item.ErrorMessage = message
	now := time.Now()
	f.item.FinishedAt = &now
	f.item.UpdatedAt = now
	return true, nil
}

func (f *fakeCloudFileDeletionRecords) ListSummariesByTaskIDs(_ context.Context, _ []int64) (map[int64]models.CloudDownloadFileDeletionSummary, error) {
	return nil, nil
}

type fakeCloudFileDeletionSession struct {
	userID      string
	findTask    *cloud.Task
	findErr     error
	fileResult  cloud.DeleteFileResult
	fileErr     error
	findCalls   int
	deleteCalls int
}

func (f *fakeCloudFileDeletionSession) Close()         {}
func (f *fakeCloudFileDeletionSession) UserID() string { return f.userID }
func (f *fakeCloudFileDeletionSession) FindUniqueTaskDeep(_ context.Context, _ string, _ int) (*cloud.Task, error) {
	f.findCalls++
	return f.findTask, f.findErr
}
func (f *fakeCloudFileDeletionSession) DeleteFile(_ context.Context, _, _ string) (cloud.DeleteFileResult, error) {
	f.deleteCalls++
	return f.fileResult, f.fileErr
}

func newCloudFileDeletionTestService(t *testing.T, task models.CloudDownloadTask, session *fakeCloudFileDeletionSession, imports *fakeCloudFileDeletionImports) (*CloudFileDeletionService, *fakeCloudFileDeletionRecords, *fakeCloudFileDeletionTasks) {
	t.Helper()
	dir := t.TempDir()
	if task.LocalPath == "" {
		task.LocalPath = filepath.Join(dir, "source.zip")
	}
	if err := os.WriteFile(task.LocalPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write local file: %v", err)
	}
	taskRepo := &fakeCloudFileDeletionTasks{task: task, setIDs: true}
	records := &fakeCloudFileDeletionRecords{}
	svc := &CloudFileDeletionService{
		tasks:       taskRepo,
		imports:     imports,
		deletions:   records,
		downloadDir: dir,
		ensureSession: func(context.Context) (cloudFileDeletionSession, error) {
			if session == nil {
				return nil, errors.New("no session")
			}
			return session, nil
		},
	}
	return svc, records, taskRepo
}

func completedCloudDeletionTask() models.CloudDownloadTask {
	size := int64(4)
	return models.CloudDownloadTask{
		ID:                7,
		Status:            models.CloudDownloadCompleted,
		InfoHash:          "abc",
		CloudFileID:       "3544176998614042374",
		CloudParentID:     "2726135308640032912",
		CloudUserID:       "user-1",
		CloudFileCategory: 1,
		DeclaredSize:      &size,
	}
}

func TestCloudFileDeletionSuccessMarksDeleted(t *testing.T) {
	session := &fakeCloudFileDeletionSession{
		userID:     "user-1",
		fileResult: cloud.DeleteFileResult{State: true, HTTPStatus: 200, JSONParsed: true},
	}
	svc, records, _ := newCloudFileDeletionTestService(t, completedCloudDeletionTask(), session, nil)

	task, err := svc.DeleteCloudFile(context.Background(), 7, nil)
	if err != nil {
		t.Fatalf("DeleteCloudFile() error = %v", err)
	}
	if task.FileDeletion == nil || task.FileDeletion.Status != models.CloudFileDeleted {
		t.Fatalf("file deletion = %+v, want deleted", task.FileDeletion)
	}
	if records.markedDelete != 1 || session.deleteCalls != 1 {
		t.Fatalf("markedDelete=%d deleteCalls=%d, want 1/1", records.markedDelete, session.deleteCalls)
	}
	if task.Status != models.CloudDownloadCompleted {
		t.Fatalf("task status = %q, want completed", task.Status)
	}
}

func TestCloudFileDeletionAlreadyDeletedIsIdempotent(t *testing.T) {
	session := &fakeCloudFileDeletionSession{userID: "user-1"}
	svc, records, _ := newCloudFileDeletionTestService(t, completedCloudDeletionTask(), session, nil)
	records.item = models.CloudDownloadFileDeletion{
		TaskID:    7,
		Status:    models.CloudFileDeleted,
		FileID:    "3544176998614042374",
		UpdatedAt: time.Now(),
	}

	task, err := svc.DeleteCloudFile(context.Background(), 7, nil)
	if err != nil {
		t.Fatalf("DeleteCloudFile() error = %v", err)
	}
	if task.FileDeletion == nil || task.FileDeletion.Status != models.CloudFileDeleted {
		t.Fatalf("file deletion = %+v", task.FileDeletion)
	}
	if session.deleteCalls != 0 || records.startCalls != 0 {
		t.Fatalf("deleteCalls=%d startCalls=%d, want 0/0", session.deleteCalls, records.startCalls)
	}
}

func TestCloudFileDeletionRequiresLocalOrTerminalImport(t *testing.T) {
	imports := &fakeCloudFileDeletionImports{item: models.CloudDownloadImport{Status: models.CloudImportHashing}}
	session := &fakeCloudFileDeletionSession{userID: "user-1"}
	task := completedCloudDeletionTask()
	svc, records, taskRepo := newCloudFileDeletionTestService(t, task, session, imports)
	// Remove the local file after the helper created it.
	if err := os.Remove(taskRepo.task.LocalPath); err != nil {
		t.Fatalf("remove local: %v", err)
	}
	_, err := svc.DeleteCloudFile(context.Background(), 7, nil)
	if !errors.Is(err, models.ErrCloudFileDeleteLocalMissing) {
		t.Fatalf("error = %v, want local missing", err)
	}
	if records.startCalls != 0 || session.deleteCalls != 0 {
		t.Fatalf("must not claim/delete without local or imported source")
	}
}

func TestCloudFileDeletionAllowsTerminalImportWithoutLocalFile(t *testing.T) {
	imports := &fakeCloudFileDeletionImports{item: models.CloudDownloadImport{Status: models.CloudImportImported}}
	session := &fakeCloudFileDeletionSession{
		userID:     "user-1",
		fileResult: cloud.DeleteFileResult{State: true, HTTPStatus: 200, JSONParsed: true},
	}
	task := completedCloudDeletionTask()
	svc, _, taskRepo := newCloudFileDeletionTestService(t, task, session, imports)
	if err := os.Remove(taskRepo.task.LocalPath); err != nil {
		t.Fatalf("remove local: %v", err)
	}
	task2, err := svc.DeleteCloudFile(context.Background(), 7, nil)
	if err != nil {
		t.Fatalf("DeleteCloudFile() error = %v", err)
	}
	if task2.FileDeletion == nil || task2.FileDeletion.Status != models.CloudFileDeleted {
		t.Fatalf("file deletion = %+v", task2.FileDeletion)
	}
}

func TestCloudFileDeletionHistoricalLazyResolveUsesActualIDs(t *testing.T) {
	task := completedCloudDeletionTask()
	task.CloudFileID = ""
	task.CloudParentID = ""
	task.CloudFileCategory = 0
	task.CloudUserID = ""
	task.DeclaredName = "source.zip"
	session := &fakeCloudFileDeletionSession{
		userID: "user-1",
		findTask: &cloud.Task{
			Name:         "source.zip",
			DeleteFileID: "1111111111111111111",
			WPPathID:     "2222222222222222222",
			FileCategory: 1,
			Size:         4,
		},
		fileResult: cloud.DeleteFileResult{State: true, HTTPStatus: 200, JSONParsed: true},
	}
	svc, _, tasksRepo := newCloudFileDeletionTestService(t, task, session, nil)
	_, err := svc.DeleteCloudFile(context.Background(), 7, nil)
	if err != nil {
		t.Fatalf("DeleteCloudFile() error = %v", err)
	}
	if tasksRepo.setCall != 1 || tasksRepo.task.CloudFileID != "1111111111111111111" || tasksRepo.task.CloudParentID != "2222222222222222222" {
		t.Fatalf("lazy ids not persisted: call=%d task=%+v", tasksRepo.setCall, tasksRepo.task)
	}
}

func TestCloudFileDeletionRejectsAccountMismatchAndFolder(t *testing.T) {
	task := completedCloudDeletionTask()
	session := &fakeCloudFileDeletionSession{userID: "user-2", fileResult: cloud.DeleteFileResult{State: true, HTTPStatus: 200, JSONParsed: true}}
	svc, records, _ := newCloudFileDeletionTestService(t, task, session, nil)
	if _, err := svc.DeleteCloudFile(context.Background(), 7, nil); !errors.Is(err, models.ErrCloudFileDeleteUnavailable) {
		t.Fatalf("account mismatch error = %v", err)
	}
	if session.deleteCalls != 0 || records.markedFail != 1 {
		t.Fatalf("account mismatch deleteCalls=%d markedFail=%d", session.deleteCalls, records.markedFail)
	}

	task = completedCloudDeletionTask()
	task.CloudFileCategory = 2
	session = &fakeCloudFileDeletionSession{userID: "user-1", fileResult: cloud.DeleteFileResult{State: true, HTTPStatus: 200, JSONParsed: true}}
	svc, records, _ = newCloudFileDeletionTestService(t, task, session, nil)
	if _, err := svc.DeleteCloudFile(context.Background(), 7, nil); !errors.Is(err, models.ErrCloudFileDeleteUnavailable) {
		t.Fatalf("folder error = %v", err)
	}
	if session.deleteCalls != 0 || records.markedFail != 1 {
		t.Fatalf("folder deleteCalls=%d markedFail=%d", session.deleteCalls, records.markedFail)
	}
}

func TestCloudFileDeletionBrowserFailureIsRecorded(t *testing.T) {
	task := completedCloudDeletionTask()
	svc, records, _ := newCloudFileDeletionTestService(t, task, nil, nil)
	svc.ensureSession = func(context.Context) (cloudFileDeletionSession, error) {
		return nil, cloud.ErrLoginRequired
	}
	if _, err := svc.DeleteCloudFile(context.Background(), 7, nil); !errors.Is(err, models.ErrCloudFileDeleteLoginRequired) {
		t.Fatalf("error = %v, want login required", err)
	}
	if records.item.Status != models.CloudFileFailed || records.markedFail != 1 {
		t.Fatalf("record=%+v markedFail=%d, want failed record", records.item, records.markedFail)
	}
}

func TestCloudFileDeletionAlreadyMissingIsDeleted(t *testing.T) {
	session := &fakeCloudFileDeletionSession{
		userID: "user-1",
		fileResult: cloud.DeleteFileResult{
			State:      false,
			Error:      "文件不存在",
			Errno:      "not_exist",
			HTTPStatus: 200,
			JSONParsed: true,
		},
	}
	svc, records, _ := newCloudFileDeletionTestService(t, completedCloudDeletionTask(), session, nil)
	task, err := svc.DeleteCloudFile(context.Background(), 7, nil)
	if err != nil {
		t.Fatalf("DeleteCloudFile() error = %v", err)
	}
	if task.FileDeletion == nil || task.FileDeletion.Status != models.CloudFileDeleted || task.FileDeletion.Note == "" {
		t.Fatalf("file deletion = %+v", task.FileDeletion)
	}
	if records.markedDelete != 1 {
		t.Fatalf("markedDelete = %d, want 1", records.markedDelete)
	}
}

func TestCloudFileDeletionConcurrentAttemptDoesNotCall115(t *testing.T) {
	session := &fakeCloudFileDeletionSession{userID: "user-1"}
	svc, records, _ := newCloudFileDeletionTestService(t, completedCloudDeletionTask(), session, nil)
	records.item = models.CloudDownloadFileDeletion{
		TaskID:    7,
		Status:    models.CloudFileDeleting,
		UpdatedAt: time.Now(),
	}
	if _, err := svc.DeleteCloudFile(context.Background(), 7, nil); !errors.Is(err, models.ErrCloudFileDeleteInProgress) {
		t.Fatalf("error = %v, want in progress", err)
	}
	if session.deleteCalls != 0 {
		t.Fatalf("deleteCalls = %d, want 0", session.deleteCalls)
	}
}

func TestCloudFileDeletionMissingTask(t *testing.T) {
	svc := &CloudFileDeletionService{tasks: &fakeCloudFileDeletionTasks{getErr: pgx.ErrNoRows}, deletions: &fakeCloudFileDeletionRecords{}}
	if _, err := svc.DeleteCloudFile(context.Background(), 404, nil); !errors.Is(err, models.ErrCloudFileDeleteNotFound) {
		t.Fatalf("error = %v, want not found", err)
	}
}

func TestCloudFileDeletionRejectsParentMissingAndHTML(t *testing.T) {
	tests := []struct {
		name   string
		result cloud.DeleteFileResult
	}{
		{name: "parent directory missing", result: cloud.DeleteFileResult{State: false, Error: "父目录不存在", HTTPStatus: 200, JSONParsed: true}},
		{name: "html not found", result: cloud.DeleteFileResult{State: false, Error: "Not Found", HTTPStatus: 200, JSONParsed: false}},
		{name: "unknown errno", result: cloud.DeleteFileResult{State: false, Error: "unknown", Errno: "99999", HTTPStatus: 200, JSONParsed: true}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			session := &fakeCloudFileDeletionSession{userID: "user-1", fileResult: tc.result}
			svc, records, _ := newCloudFileDeletionTestService(t, completedCloudDeletionTask(), session, nil)
			if _, err := svc.DeleteCloudFile(context.Background(), 7, nil); err == nil {
				t.Fatal("expected failure, got success")
			}
			if records.markedDelete != 0 || records.item.Status != models.CloudFileFailed {
				t.Fatalf("record=%+v, want failed", records.item)
			}
		})
	}
}

func TestCloudFileDeletionMapsDeleteFileBrowserError(t *testing.T) {
	session := &fakeCloudFileDeletionSession{userID: "user-1", fileErr: cloud.ErrBrowserDown}
	svc, records, _ := newCloudFileDeletionTestService(t, completedCloudDeletionTask(), session, nil)
	if _, err := svc.DeleteCloudFile(context.Background(), 7, nil); !errors.Is(err, models.ErrCloudFileDeleteBrowserDown) {
		t.Fatalf("error = %v, want browser down", err)
	}
	if records.item.Status != models.CloudFileFailed {
		t.Fatalf("record = %+v, want failed", records.item)
	}
}

func TestCloudFileDeletionRejectsMalformedIDs(t *testing.T) {
	task := completedCloudDeletionTask()
	task.CloudFileID = "not-a-number"
	session := &fakeCloudFileDeletionSession{userID: "user-1", fileResult: cloud.DeleteFileResult{State: true, HTTPStatus: 200, JSONParsed: true}}
	svc, records, _ := newCloudFileDeletionTestService(t, task, session, nil)
	if _, err := svc.DeleteCloudFile(context.Background(), 7, nil); !errors.Is(err, models.ErrCloudFileDeleteUnavailable) {
		t.Fatalf("error = %v, want unavailable", err)
	}
	if session.deleteCalls != 0 || records.item.Status != models.CloudFileFailed {
		t.Fatalf("deleteCalls=%d record=%+v", session.deleteCalls, records.item)
	}
}

func TestCloudFileDeletionRejectsAmbiguousHistoricalTask(t *testing.T) {
	task := completedCloudDeletionTask()
	task.CloudFileID = ""
	task.CloudParentID = ""
	task.CloudFileCategory = 0
	task.CloudUserID = ""
	task.DeclaredName = "source.zip"
	session := &fakeCloudFileDeletionSession{userID: "user-1", findErr: cloud.ErrAmbiguousTask}
	svc, records, _ := newCloudFileDeletionTestService(t, task, session, nil)
	if _, err := svc.DeleteCloudFile(context.Background(), 7, nil); !errors.Is(err, models.ErrCloudFileDeleteUnavailable) {
		t.Fatalf("error = %v, want unavailable", err)
	}
	if session.deleteCalls != 0 || records.item.Status != models.CloudFileFailed {
		t.Fatalf("deleteCalls=%d record=%+v", session.deleteCalls, records.item)
	}
}
