package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestPreviewAVJavDBUsesMDCXStyleDOMDetailParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search":
			_, _ = w.Write([]byte(`
<html>
  <body>
    <a class="box" href="/v/ssis-001">
      <div class="video-title"><strong>SSIS-001 First Impression</strong></div>
      <div class="meta">2024-01-30</div>
    </a>
    <a class="box" href="/v/noise-002">
      <div class="video-title"><strong>NOISE-002 Other Title</strong></div>
    </a>
  </body>
</html>
`))
		case "/v/ssis-001":
			_, _ = w.Write([]byte(`
<!doctype html>
<html>
<body>
  <a class="button is-white copy-to-clipboard" data-clipboard-text="SSIS-001"></a>
  <h2 class="title is-4">
    <strong class="current-title">First Impression 001</strong>
    <span class="origin-title">SSIS-001 Original</span>
  </h2>
  <div class="panel-block"><strong>Released Date:</strong><span>2024-01-30</span></div>
  <img src="/covers/ssis-001.jpg" class="video-cover">
  <span><a>Yua Mikami</a><strong class="female"></strong></span>
  <span><a>Ena Satsuki</a><strong class="female"></strong></span>
  <script type="application/ld+json">
  {
    "@context":"https://schema.org",
    "@type":"Movie",
    "description":"Elite office affair."
  }
  </script>
</body>
</html>`))
		case "/v/noise-002":
			_, _ = w.Write([]byte(`
<!doctype html>
<html><body><h2 class="title is-4">NOISE-002 Other Title</h2></body></html>
`))
		case "/covers/ssis-001.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("fake-image"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "javdb-mdcx-dom-test", time.Second)

	got, err := svc.PreviewAV(context.Background(), "SSIS-001")
	if err != nil {
		t.Fatalf("PreviewAV returned error: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected av candidates, got none")
	}

	first := got[0]
	if first["title"] != "First Impression 001" {
		t.Fatalf("expected current-title only, got=%v", first["title"])
	}
	if first["overview"] != "Elite office affair." {
		t.Fatalf("expected ld+json overview, got=%v", first["overview"])
	}
	if first["poster_url"] != server.URL+"/covers/ssis-001.jpg" {
		t.Fatalf("expected video-cover poster, got=%v", first["poster_url"])
	}
	actors, ok := first["actors"].([]string)
	if !ok {
		t.Fatalf("expected actors []string, got=%T", first["actors"])
	}
	if !reflect.DeepEqual(actors, []string{"Yua Mikami", "Ena Satsuki"}) {
		t.Fatalf("unexpected actors: %v", actors)
	}
}

func TestPreviewAVJavDBReturnsExplicitErrorOnCloudflareBlock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search" {
			_, _ = w.Write([]byte(`<html><body><div>ray-id</div><div>please wait</div></body></html>`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	svc := NewScraperService(nil, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "javdb-blocked-test", time.Second)

	_, err := svc.PreviewAV(context.Background(), "SSIS-001")
	if err == nil {
		t.Fatalf("expected explicit blocked-page error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "cloudflare") {
		t.Fatalf("expected cloudflare error, got=%v", err)
	}
}
