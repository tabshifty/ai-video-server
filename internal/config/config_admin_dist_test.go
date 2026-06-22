package config

import (
	"path/filepath"
	"testing"
)

func TestLoadAdminWebDistPathFromEnv(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("ADMIN_WEB_DIST_PATH", "/Volumes/large/ai-video-server/current/admin-web-dist")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.AdminWebDistPath != "/Volumes/large/ai-video-server/current/admin-web-dist" {
		t.Fatalf("unexpected AdminWebDistPath: %q", cfg.AdminWebDistPath)
	}
}

func TestLoadAdminWebDistPathDefaultsToRelative(t *testing.T) {
	setRequiredEnv(t)
	// 不设 ADMIN_WEB_DIST_PATH

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	want := filepath.Join("admin-web", "dist")
	if cfg.AdminWebDistPath != want {
		t.Fatalf("default AdminWebDistPath = %q, want %q", cfg.AdminWebDistPath, want)
	}
}
