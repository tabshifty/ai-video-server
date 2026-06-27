package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
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
		"POST /api/v1/admin/ed2k-download/tasks/:id/retry",
		"DELETE /api/v1/admin/ed2k-download/tasks/:id",
	} {
		if _, ok := routes[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
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
