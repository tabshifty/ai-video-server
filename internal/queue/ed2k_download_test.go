package queue

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"video-server/internal/models"
)

func TestCommandEd2kDownloadExecutorPassesTaskMetadataByEnv(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	capturePath := filepath.Join(dir, "captured-env.json")
	scriptPath := filepath.Join(dir, "executor.sh")
	script := strings.Join([]string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		"printf '{\"task_id\":\"%s\",\"title\":\"%s\",\"resource_hash\":\"%s\",\"filename\":\"%s\",\"declared_size\":\"%s\",\"source_link\":\"%s\",\"mode\":\"%s\"}' \"$ED2K_TASK_ID\" \"$ED2K_TASK_TITLE\" \"$ED2K_RESOURCE_HASH\" \"$ED2K_FILENAME\" \"$ED2K_DECLARED_SIZE\" \"$ED2K_SOURCE_LINK\" \"$ED2K_EXECUTOR_MODE\" > \"$1\"",
		"printf '{\"Status\":\"completed\",\"OutputDir\":\"/tmp/ed2k\",\"DownloadedPath\":\"/tmp/ed2k/file.mkv\",\"ProgressText\":\"下载已完成\",\"Files\":[{\"name\":\"file.mkv\",\"path\":\"file.mkv\",\"size\":12}]}'",
	}, "\n")
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	task := models.AdminEd2kDownloadTask{
		ID:           uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Title:        "测试任务",
		SourceLink:   "ed2k://|file|demo.mkv|123|ABCDEF1234567890|/",
		ResourceHash: "ABCDEF1234567890",
		Filename:     "demo.mkv",
		DeclaredSize: 123,
	}

	executor := CommandEd2kDownloadExecutor{
		Command: scriptPath,
		Args:    []string{capturePath},
	}

	result, err := executor.Submit(context.Background(), task)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.OutputDir != "/tmp/ed2k" {
		t.Fatalf("unexpected output dir: %s", result.OutputDir)
	}
	if len(result.Files) != 1 || result.Files[0].Path != "file.mkv" {
		t.Fatalf("unexpected files: %+v", result.Files)
	}

	raw, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatalf("read capture: %v", err)
	}
	var captured struct {
		TaskID       string `json:"task_id"`
		Title        string `json:"title"`
		ResourceHash string `json:"resource_hash"`
		Filename     string `json:"filename"`
		DeclaredSize string `json:"declared_size"`
		SourceLink   string `json:"source_link"`
		Mode         string `json:"mode"`
	}
	if err := json.Unmarshal(raw, &captured); err != nil {
		t.Fatalf("unmarshal capture: %v", err)
	}
	if captured.TaskID != task.ID.String() {
		t.Fatalf("unexpected task id: %s", captured.TaskID)
	}
	if captured.Title != task.Title {
		t.Fatalf("unexpected title: %s", captured.Title)
	}
	if captured.ResourceHash != task.ResourceHash {
		t.Fatalf("unexpected resource hash: %s", captured.ResourceHash)
	}
	if captured.Filename != task.Filename {
		t.Fatalf("unexpected filename: %s", captured.Filename)
	}
	if captured.DeclaredSize != "123" {
		t.Fatalf("unexpected declared size: %s", captured.DeclaredSize)
	}
	if captured.SourceLink != task.SourceLink {
		t.Fatalf("unexpected source link: %s", captured.SourceLink)
	}
	if captured.Mode != ed2kExecutorModeSubmit {
		t.Fatalf("unexpected executor mode: %s", captured.Mode)
	}
}

func TestHandleEd2kDownloadMarksFailedAndStopsAsynqRetry(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo := &ed2kDownloadRepoStub{
		task: models.AdminEd2kDownloadTask{
			ID:           taskID,
			Status:       "queued",
			SourceLink:   "ed2k://|file|demo.mkv|123|ABCDEF1234567890|/",
			ResourceHash: "ABCDEF1234567890",
			Filename:     "demo.mkv",
		},
		refreshedTask: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "running",
		},
	}
	executor := ed2kDownloadExecutorStub{submitErr: errors.New("executor failed")}
	task := asynq.NewTask(TypeEd2kDownload, mustMarshalEd2kPayload(t, Ed2kDownloadPayload{TaskID: taskID.String()}))
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	err := handleEd2kDownload(context.Background(), task, repo, &executor, logger)
	if err != nil {
		t.Fatalf("handleEd2kDownload() error = %v, want nil", err)
	}
	if !repo.runningCalled {
		t.Fatal("expected task to enter running before executor failure")
	}
	if !executor.submitCalled {
		t.Fatal("expected queued task to call submit")
	}
	if !repo.failedCalled {
		t.Fatal("expected task to be marked failed")
	}
	if repo.completedCalled {
		t.Fatal("did not expect task to be marked completed")
	}
	if repo.failedMessage != "executor failed" {
		t.Fatalf("unexpected failed message: %s", repo.failedMessage)
	}
}

func TestHandleEd2kDownloadSkipsLateResultWhenTaskLeavesRunning(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	repo := &ed2kDownloadRepoStub{
		task: models.AdminEd2kDownloadTask{
			ID:           taskID,
			Status:       "running",
			SourceLink:   "ed2k://|file|demo.mkv|123|ABCDEF1234567890|/",
			ResourceHash: "ABCDEF1234567890",
			Filename:     "demo.mkv",
		},
		refreshedTask: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "canceling",
		},
	}
	executor := ed2kDownloadExecutorStub{
		statusResult: Ed2kDownloadResult{
			Status:         ed2kDownloadStatusCompleted,
			OutputDir:      "/tmp/ed2k",
			DownloadedPath: "/tmp/ed2k/demo.mkv",
			ProgressText:   "下载已完成",
			Files: []models.AdminEd2kDownloadTaskFile{
				{Name: "demo.mkv", Path: "demo.mkv", Size: 123},
			},
		},
	}
	task := asynq.NewTask(TypeEd2kDownload, mustMarshalEd2kPayload(t, Ed2kDownloadPayload{TaskID: taskID.String()}))
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := handleEd2kDownload(context.Background(), task, repo, &executor, logger); err != nil {
		t.Fatalf("handleEd2kDownload() error = %v", err)
	}
	if repo.completedCalled {
		t.Fatal("did not expect cancelled/canceling task to be marked completed")
	}
	if repo.failedCalled {
		t.Fatal("did not expect cancelled/canceling task to be marked failed")
	}
}

func TestHandleEd2kDownloadTreatsConditionalCompleteMissAsLateResult(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("23333333-2222-2222-2222-222222222222")
	repo := &ed2kDownloadRepoStub{
		task: models.AdminEd2kDownloadTask{
			ID:           taskID,
			Status:       "running",
			SourceLink:   "ed2k://|file|demo.mkv|123|ABCDEF1234567890|/",
			ResourceHash: "ABCDEF1234567890",
			Filename:     "demo.mkv",
		},
		refreshedTask: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "running",
		},
		completeErr: errors.New("mark ed2k download task completed: no rows in result set"),
	}
	executor := ed2kDownloadExecutorStub{
		statusResult: Ed2kDownloadResult{
			Status:         ed2kDownloadStatusCompleted,
			OutputDir:      "/tmp/ed2k",
			DownloadedPath: "/tmp/ed2k/demo.mkv",
			ProgressText:   "下载已完成",
			Files:          []models.AdminEd2kDownloadTaskFile{{Name: "demo.mkv", Path: "demo.mkv", Size: 123}},
		},
	}
	task := asynq.NewTask(TypeEd2kDownload, mustMarshalEd2kPayload(t, Ed2kDownloadPayload{TaskID: taskID.String()}))
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := handleEd2kDownload(context.Background(), task, repo, &executor, logger); err != nil {
		t.Fatalf("handleEd2kDownload() should swallow late-result conditional miss, got %v", err)
	}
	if !repo.completedCalled {
		t.Fatal("expected completion attempt")
	}
}

func TestHandleEd2kDownloadReturnsRetrySentinelWhileRunning(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("25555555-2222-2222-2222-222222222222")
	repo := &ed2kDownloadRepoStub{
		task: models.AdminEd2kDownloadTask{
			ID:           taskID,
			Status:       "running",
			SourceLink:   "ed2k://|file|demo.mkv|123|ABCDEF1234567890|/",
			ResourceHash: "ABCDEF1234567890",
			Filename:     "demo.mkv",
		},
		refreshedTask: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "running",
		},
	}
	executor := ed2kDownloadExecutorStub{
		statusResult: Ed2kDownloadResult{
			Status:       ed2kDownloadStatusRunning,
			ProgressText: "下载进行中（12.3%）",
		},
	}
	task := asynq.NewTask(TypeEd2kDownload, mustMarshalEd2kPayload(t, Ed2kDownloadPayload{TaskID: taskID.String()}))
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	err := handleEd2kDownload(context.Background(), task, repo, &executor, logger)
	if !errors.Is(err, ErrEd2kDownloadStillInProgress) {
		t.Fatalf("handleEd2kDownload() error = %v, want in-progress sentinel", err)
	}
	if !repo.progressCalled {
		t.Fatal("expected running task progress to update")
	}
	if repo.progressText != "下载进行中（12.3%）" {
		t.Fatalf("unexpected progress text: %s", repo.progressText)
	}
	if repo.failedCalled || repo.completedCalled {
		t.Fatal("did not expect non-terminal status to mark final state")
	}
}

func TestHandleEd2kDownloadSkipsFilesCleanedTaskUntilApiRequeuesIt(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("24444444-2222-2222-2222-222222222222")
	repo := &ed2kDownloadRepoStub{
		task: models.AdminEd2kDownloadTask{
			ID:           taskID,
			Status:       "files_cleaned",
			SourceLink:   "ed2k://|file|demo.mkv|123|ABCDEF1234567890|/",
			ResourceHash: "ABCDEF1234567890",
			Filename:     "demo.mkv",
		},
		refreshedTask: models.AdminEd2kDownloadTask{
			ID:     taskID,
			Status: "running",
		},
	}
	executor := ed2kDownloadExecutorStub{
		statusResult: Ed2kDownloadResult{
			Status:         ed2kDownloadStatusCompleted,
			OutputDir:      "/tmp/ed2k",
			DownloadedPath: "/tmp/ed2k/demo.mkv",
			ProgressText:   "下载已完成",
			Files:          []models.AdminEd2kDownloadTaskFile{{Name: "demo.mkv", Path: "demo.mkv", Size: 123}},
		},
	}
	task := asynq.NewTask(TypeEd2kDownload, mustMarshalEd2kPayload(t, Ed2kDownloadPayload{TaskID: taskID.String()}))
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := handleEd2kDownload(context.Background(), task, repo, &executor, logger); err != nil {
		t.Fatalf("handleEd2kDownload() error = %v", err)
	}
	if repo.runningCalled {
		t.Fatal("did not expect files_cleaned task to enter running before api requeue")
	}
	if repo.completedCalled {
		t.Fatal("did not expect files_cleaned task to complete without queued transition")
	}
}

type ed2kDownloadRepoStub struct {
	task            models.AdminEd2kDownloadTask
	refreshedTask   models.AdminEd2kDownloadTask
	getCalls        int
	runningCalled   bool
	progressCalled  bool
	failedCalled    bool
	completedCalled bool
	progressText    string
	failedMessage   string
	failErr         error
	completeErr     error
}

func (s *ed2kDownloadRepoStub) GetEd2kDownloadTask(context.Context, uuid.UUID) (models.AdminEd2kDownloadTask, error) {
	s.getCalls++
	if s.getCalls > 1 && s.refreshedTask.ID != uuid.Nil {
		return s.refreshedTask, nil
	}
	return s.task, nil
}

func (s *ed2kDownloadRepoStub) MarkEd2kDownloadTaskRunning(context.Context, uuid.UUID, string, models.AdminEd2kDownloadTaskHistoryItem) error {
	s.runningCalled = true
	return nil
}

func (s *ed2kDownloadRepoStub) UpdateEd2kDownloadTaskProgress(_ context.Context, _ uuid.UUID, progressText string) error {
	s.progressCalled = true
	s.progressText = progressText
	return nil
}

func (s *ed2kDownloadRepoStub) MarkEd2kDownloadTaskFailed(_ context.Context, _ uuid.UUID, errorMessage string, _ models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	s.failedCalled = true
	s.failedMessage = errorMessage
	if s.failErr != nil {
		return models.AdminEd2kDownloadTask{}, s.failErr
	}
	return s.task, nil
}

func (s *ed2kDownloadRepoStub) MarkEd2kDownloadTaskCompleted(context.Context, uuid.UUID, string, string, []models.AdminEd2kDownloadTaskFile, string, models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	s.completedCalled = true
	if s.completeErr != nil {
		return models.AdminEd2kDownloadTask{}, s.completeErr
	}
	return s.task, nil
}

type ed2kDownloadExecutorStub struct {
	submitResult Ed2kDownloadResult
	statusResult Ed2kDownloadResult
	submitErr    error
	statusErr    error
	submitCalled bool
	statusCalled bool
}

func (s *ed2kDownloadExecutorStub) Submit(context.Context, models.AdminEd2kDownloadTask) (Ed2kDownloadResult, error) {
	s.submitCalled = true
	if s.submitErr != nil {
		return Ed2kDownloadResult{}, s.submitErr
	}
	return s.submitResult, nil
}

func (s *ed2kDownloadExecutorStub) Status(context.Context, models.AdminEd2kDownloadTask) (Ed2kDownloadResult, error) {
	s.statusCalled = true
	if s.statusErr != nil {
		return Ed2kDownloadResult{}, s.statusErr
	}
	return s.statusResult, nil
}

func mustMarshalEd2kPayload(t *testing.T, payload Ed2kDownloadPayload) []byte {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return raw
}
