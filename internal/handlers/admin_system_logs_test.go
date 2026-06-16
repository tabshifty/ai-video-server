package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdminSystemLogsReturnsEmptyLinesWhenLogFileMissing(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	api := &API{serverLogPath: filepath.Join(t.TempDir(), "server.log")}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/system/logs?lines=300", nil)

	api.AdminSystemLogs(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code=0, got=%d body=%s", resp.Code, rec.Body.String())
	}
	if resp.Data["line_count"] != float64(0) {
		t.Fatalf("expected line_count=0, got=%v", resp.Data["line_count"])
	}
	lines, ok := resp.Data["lines"].([]any)
	if !ok {
		t.Fatalf("expected lines array, got=%T", resp.Data["lines"])
	}
	if len(lines) != 0 {
		t.Fatalf("expected empty lines, got=%v", lines)
	}
}

func TestAdminSystemLogsDownloadReturnsEmptyFileWhenLogFileMissing(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	api := &API{serverLogPath: filepath.Join(t.TempDir(), "server.log")}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/system/logs?download=1", nil)

	api.AdminSystemLogs(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got=%d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("expected text/plain content type, got=%q", got)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty download body, got=%q", rec.Body.String())
	}
}
