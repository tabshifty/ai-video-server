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
		"printf '{\"task_id\":\"%s\",\"title\":\"%s\",\"resource_hash\":\"%s\",\"filename\":\"%s\",\"declared_size\":\"%s\",\"source_link\":\"%s\"}' \"$ED2K_TASK_ID\" \"$ED2K_TASK_TITLE\" \"$ED2K_RESOURCE_HASH\" \"$ED2K_FILENAME\" \"$ED2K_DECLARED_SIZE\" \"$ED2K_SOURCE_LINK\" > \"$1\"",
		"printf '{\"OutputDir\":\"/tmp/ed2k\",\"DownloadedPath\":\"/tmp/ed2k/file.mkv\",\"ProgressText\":\"下载已完成\",\"Files\":[{\"name\":\"file.mkv\",\"path\":\"file.mkv\",\"size\":12}]}'",
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

	result, err := executor.Run(context.Background(), task)
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
	}
	executor := ed2kDownloadExecutorStub{err: errors.New("executor failed")}
	task := asynq.NewTask(TypeEd2kDownload, mustMarshalEd2kPayload(t, Ed2kDownloadPayload{TaskID: taskID.String()}))
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	err := handleEd2kDownload(context.Background(), task, repo, executor, logger)
	if err != nil {
		t.Fatalf("handleEd2kDownload() error = %v, want nil", err)
	}
	if !repo.runningCalled {
		t.Fatal("expected task to enter running before executor failure")
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

type ed2kDownloadRepoStub struct {
	task            models.AdminEd2kDownloadTask
	runningCalled   bool
	failedCalled    bool
	completedCalled bool
	failedMessage   string
}

func (s *ed2kDownloadRepoStub) GetEd2kDownloadTask(context.Context, uuid.UUID) (models.AdminEd2kDownloadTask, error) {
	return s.task, nil
}

func (s *ed2kDownloadRepoStub) MarkEd2kDownloadTaskRunning(context.Context, uuid.UUID, string, models.AdminEd2kDownloadTaskHistoryItem) error {
	s.runningCalled = true
	return nil
}

func (s *ed2kDownloadRepoStub) MarkEd2kDownloadTaskFailed(_ context.Context, _ uuid.UUID, errorMessage string, _ models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	s.failedCalled = true
	s.failedMessage = errorMessage
	return s.task, nil
}

func (s *ed2kDownloadRepoStub) MarkEd2kDownloadTaskCompleted(context.Context, uuid.UUID, string, string, []models.AdminEd2kDownloadTaskFile, string, models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error) {
	s.completedCalled = true
	return s.task, nil
}

type ed2kDownloadExecutorStub struct {
	result Ed2kDownloadResult
	err    error
}

func (s ed2kDownloadExecutorStub) Run(context.Context, models.AdminEd2kDownloadTask) (Ed2kDownloadResult, error) {
	if s.err != nil {
		return Ed2kDownloadResult{}, s.err
	}
	return s.result, nil
}

func mustMarshalEd2kPayload(t *testing.T, payload Ed2kDownloadPayload) []byte {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return raw
}
