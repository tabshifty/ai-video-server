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

	req = httptest.NewRequest(http.MethodGet, "/admin/assets/app.css", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /admin/assets/app.css = %d, want 200", w.Code)
	}
	if got := w.Body.String(); got != "/* css */" {
		t.Fatalf("css body = %q", got)
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
