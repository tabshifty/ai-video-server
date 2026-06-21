package services

import (
	"path/filepath"
	"strings"
	"testing"

	"video-server/internal/models"
)

func TestParseTVAPKMetadataParsesReleaseAPK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileName    string
		wantABI     string
		wantVersion int64
	}{
		{
			name:        "arm64 release apk",
			fileName:    "tv-app-arm64-v8a-release.apk",
			wantABI:     models.TVABIArm64,
			wantVersion: 121,
		},
		{
			name:        "armeabi release apk",
			fileName:    "tv-app-armeabi-v7a-release.apk",
			wantABI:     models.TVABIArmV7,
			wantVersion: 121,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			apkPath := filepath.Join("..", "..", "android-tv-app", "tv-app", "release", tt.fileName)
			meta, err := ParseTVAPKMetadata(apkPath, tt.fileName)
			if err != nil {
				t.Fatalf("ParseTVAPKMetadata() error = %v", err)
			}
			if meta.PackageName != models.TVAppPackageName {
				t.Fatalf("package_name = %q, want %q", meta.PackageName, models.TVAppPackageName)
			}
			if meta.ABI != tt.wantABI {
				t.Fatalf("abi = %q, want %q", meta.ABI, tt.wantABI)
			}
			if meta.VersionCode != tt.wantVersion {
				t.Fatalf("version_code = %d, want %d", meta.VersionCode, tt.wantVersion)
			}
			if meta.VersionName != "0.1.121" {
				t.Fatalf("version_name = %q, want 0.1.121", meta.VersionName)
			}
			if meta.IsDebuggable {
				t.Fatal("expected release apk to be non-debuggable")
			}
			if meta.FileSize <= 0 {
				t.Fatalf("file_size = %d, want > 0", meta.FileSize)
			}
			if len(meta.SHA256) != 64 {
				t.Fatalf("sha256 length = %d, want 64", len(meta.SHA256))
			}
			if !strings.Contains(string(meta.RawManifest), `"package_name":"com.chee.videos.tv"`) {
				t.Fatalf("raw_manifest missing package_name, got %s", string(meta.RawManifest))
			}
		})
	}
}

func TestTVReleaseHelpers(t *testing.T) {
	t.Parallel()

	items := []models.AdminTvAppReleaseABIItem{
		{ABI: models.TVABIArm64},
	}
	if got := TVReleaseUploadedABIs(items); len(got) != 1 || got[0] != models.TVABIArm64 {
		t.Fatalf("TVReleaseUploadedABIs() = %#v", got)
	}
	if got := TVReleaseMissingABIs(items); len(got) != 1 || got[0] != models.TVABIArmV7 {
		t.Fatalf("TVReleaseMissingABIs() = %#v", got)
	}
	if TVReleaseABIComplete(items) {
		t.Fatal("expected ABI set to be incomplete")
	}
	if got := NormalizeTVReleaseStatus(true, items); got != models.TVReleaseStatusPublishedMissing {
		t.Fatalf("NormalizeTVReleaseStatus() = %q, want %q", got, models.TVReleaseStatusPublishedMissing)
	}
}
