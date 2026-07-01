package handlers

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/models"
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
		"DELETE /api/v1/admin/ed2k-download/tasks/:id",
	} {
		if _, ok := routes[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
	if _, ok := routes["POST /api/v1/admin/ed2k-download/tasks/:id/retry"]; ok {
		t.Fatal("did not expect retry route to remain registered")
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
	if got := buildEd2kDownloadTaskTitle("主角.mkv", "  "); got != "主角.mkv" {
		t.Fatalf("unexpected fallback title: %s", got)
	}
	if got := buildEd2kDownloadTaskTitle("主角.mkv", "  自定义 标题  "); got != "自定义 标题" {
		t.Fatalf("unexpected normalized title: %s", got)
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

func assertPathMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected %s to be removed, got err=%v", path, err)
	}
}
