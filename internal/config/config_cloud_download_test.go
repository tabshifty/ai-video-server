package config

import (
	"strings"
	"testing"
)

func TestLoadCloudDownloadDisabledIgnoresEndpoint(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("CLOUD_DOWNLOAD_ENABLED", "false")
	t.Setenv("CLOUD_BROWSER_CDP", "not-a-url")

	if _, err := Load(); err != nil {
		t.Fatalf("Load returned error for a disabled feature: %v", err)
	}
}

func TestLoadCloudDownloadRejectsInvalidEndpoint(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("CLOUD_DOWNLOAD_ENABLED", "true")
	t.Setenv("CLOUD_BROWSER_CDP", "not-a-url")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "CLOUD_BROWSER_CDP") {
		t.Fatalf("error = %v, want CLOUD_BROWSER_CDP validation error", err)
	}
}

func TestLoadCloudDownloadRequiresPortWhenBrowserPathIsSet(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("CLOUD_DOWNLOAD_ENABLED", "true")
	t.Setenv("CLOUD_BROWSER_CDP", "http://127.0.0.1")
	t.Setenv("CLOUD_BROWSER_APP", "/Applications/115Browser.app")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "port") {
		t.Fatalf("error = %v, want port validation error", err)
	}
}

func TestLoadCloudDownloadQueueRejectsInvalidName(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("CLOUD_DOWNLOAD_QUEUE", "cloud queue")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "CLOUD_DOWNLOAD_QUEUE") {
		t.Fatalf("error = %v, want CLOUD_DOWNLOAD_QUEUE name validation error", err)
	}
}

func TestLoadCloudDownloadQueueRejectsCollision(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("CLOUD_DOWNLOAD_QUEUE", "transcode")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "CLOUD_DOWNLOAD_QUEUE") {
		t.Fatalf("error = %v, want CLOUD_DOWNLOAD_QUEUE validation error", err)
	}
}

func TestLoadCloudDownloadAcceptsValidEndpoint(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("CLOUD_DOWNLOAD_ENABLED", "true")
	t.Setenv("CLOUD_BROWSER_CDP", "http://127.0.0.1:9333/")
	t.Setenv("CLOUD_BROWSER_APP", "/Applications/115Browser.app")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !cfg.CloudDownloadEnabled {
		t.Fatalf("CloudDownloadEnabled = false, want true")
	}
	if cfg.CloudBrowserCDP != "http://127.0.0.1:9333" {
		t.Fatalf("CloudBrowserCDP = %q", cfg.CloudBrowserCDP)
	}
}

func TestLoadCloudIngestQueueRejectsDownloadCollision(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("CLOUD_INGEST_QUEUE", "cloud")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "CLOUD_INGEST_QUEUE") {
		t.Fatalf("error = %v, want CLOUD_INGEST_QUEUE collision error", err)
	}
}

func TestLoadCloudIngestQueueDefaultsToDedicatedQueue(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// 导入不能和下载同队列：worker 并发是全局的，下载会把导入饿死。
	if cfg.CloudIngestQueue == cfg.CloudDownloadQueue {
		t.Fatalf("ingest queue must differ from download queue, both %q", cfg.CloudIngestQueue)
	}
	if cfg.CloudIngestQueue != "cloud-ingest" {
		t.Fatalf("CloudIngestQueue = %q, want cloud-ingest", cfg.CloudIngestQueue)
	}
}
