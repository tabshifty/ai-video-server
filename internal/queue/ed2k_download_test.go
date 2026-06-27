package queue

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"

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
