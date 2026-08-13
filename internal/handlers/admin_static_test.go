package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMountAdminStaticServesIndexFromGivenDir(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("hello-admin"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.css"), []byte("/* css */"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	mountAdminStatic(r, dir)

	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /admin/ = %d, want 200", w.Code)
	}
	if got := w.Body.String(); got != "hello-admin" {
		t.Fatalf("body = %q, want hello-admin", got)
	}
	if got := w.Header().Get("Cache-Control"); got != adminIndexCacheControl {
		t.Fatalf("index Cache-Control = %q, want %q", got, adminIndexCacheControl)
	}

	req = httptest.NewRequest(http.MethodHead, "/admin", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("HEAD /admin = %d, want 200", w.Code)
	}
	if got := w.Header().Get("Cache-Control"); got != adminIndexCacheControl {
		t.Fatalf("HEAD /admin Cache-Control = %q, want %q", got, adminIndexCacheControl)
	}

	req = httptest.NewRequest(http.MethodGet, "/admin/assets/app.css", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /admin/assets/app.css = %d, want 200", w.Code)
	}
	if got := w.Body.String(); got != "/* css */" {
		t.Fatalf("css body = %q", got)
	}
	if got := w.Header().Get("Cache-Control"); got != adminAssetCacheControl {
		t.Fatalf("asset Cache-Control = %q, want %q", got, adminAssetCacheControl)
	}
}

func TestMountAdminStaticMissingAssetIsNotLongCached(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("hello-admin"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	mountAdminStatic(r, dir)

	req := httptest.NewRequest(http.MethodGet, "/admin/assets/old-hash.css", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("GET missing asset = %d, want 404", w.Code)
	}
	if got := w.Header().Get("Cache-Control"); got != adminMissingAssetCacheControl {
		t.Fatalf("missing asset Cache-Control = %q, want %q", got, adminMissingAssetCacheControl)
	}
}

func TestMountAdminStaticServesChinaMapDataFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	mapDir := filepath.Join(dir, "china-map")
	if err := os.MkdirAll(filepath.Join(mapDir, "layers"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("hello-admin"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mapDir, "catalog.json"), []byte(`{"root_region_code":"CN"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mapDir, "layers", "CN.geojson"), []byte(`{"type":"FeatureCollection","features":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	mountAdminStatic(r, dir)

	tests := []struct {
		name         string
		path         string
		wantStatus   int
		wantBody     string
		contentType  string
		cacheControl string
	}{
		{
			name:         "catalog",
			path:         "/admin/china-map/catalog.json",
			wantStatus:   http.StatusOK,
			wantBody:     `{"root_region_code":"CN"}`,
			contentType:  "application/json",
			cacheControl: adminMapDataCacheControl,
		},
		{
			name:         "layer",
			path:         "/admin/china-map/layers/CN.geojson",
			wantStatus:   http.StatusOK,
			wantBody:     `{"type":"FeatureCollection","features":[]}`,
			contentType:  "application/geo+json",
			cacheControl: adminMapDataCacheControl,
		},
		{
			name:       "missing data",
			path:       "/admin/china-map/layers/missing.geojson",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "path traversal",
			path:       "/admin/china-map/%2e%2e/index.html",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unsupported file",
			path:       "/admin/china-map/readme.txt",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("GET %s = %d, want %d; body %q", tt.path, w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantBody != "" && w.Body.String() != tt.wantBody {
				t.Fatalf("GET %s body = %q, want %q", tt.path, w.Body.String(), tt.wantBody)
			}
			if tt.contentType != "" && w.Header().Get("Content-Type") != tt.contentType {
				t.Fatalf("GET %s Content-Type = %q, want %q", tt.path, w.Header().Get("Content-Type"), tt.contentType)
			}
			if tt.cacheControl != "" && w.Header().Get("Cache-Control") != tt.cacheControl {
				t.Fatalf("GET %s Cache-Control = %q, want %q", tt.path, w.Header().Get("Cache-Control"), tt.cacheControl)
			}
			if tt.wantStatus == http.StatusNotFound && w.Body.String() == "hello-admin" {
				t.Fatalf("GET %s unexpectedly returned the SPA index", tt.path)
			}
		})
	}
}

func TestMountAdminStaticSkipsWhenDirMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	missing := filepath.Join(t.TempDir(), "no-such-dir")

	r := gin.New()
	mountAdminStatic(r, missing)

	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Fatalf("GET /admin/ unexpectedly OK while dir missing")
	}
}
