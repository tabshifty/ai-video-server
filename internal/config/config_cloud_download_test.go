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
