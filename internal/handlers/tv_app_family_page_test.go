package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterIncludesTVAppFamilyPageRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := &API{}
	api.Register(router)

	routes := map[string]struct{}{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}

	for _, want := range []string{
		"GET /tv-app",
		"HEAD /tv-app",
		"GET /tv-app/",
		"HEAD /tv-app/",
	} {
		if _, ok := routes[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
}

func TestMountTVAppFamilyPageServesStandaloneHTML(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mountTVAppFamilyPage(router)

	req := httptest.NewRequest(http.MethodGet, "/tv-app/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /tv-app/ = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
	body := rec.Body.String()
	for _, want := range []string{
		"<title>TV 安装包下载</title>",
		`/api/v1/auth/login`,
		`/api/v1/auth/refresh`,
		`/api/v1/tv-app/releases`,
		`data-download-release-id="`,
		`buildDownloadURL(link.dataset.downloadReleaseId, link.dataset.downloadAbi, accessToken)`,
		"登录后查看下载列表",
		"最近三版安装包",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected body to contain %q", want)
		}
	}
	if strings.Contains(body, "/admin/assets/") || strings.Contains(body, "hello-admin") {
		t.Fatalf("family page should not include admin shell assets, got body %q", body)
	}
}

func TestMountTVAppFamilyPageSupportsHead(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mountTVAppFamilyPage(router)

	req := httptest.NewRequest(http.MethodHead, "/tv-app", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HEAD /tv-app = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}
