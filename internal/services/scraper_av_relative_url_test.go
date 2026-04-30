package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestResolveRelativeAVURLMatchesBrowserURLSemantics(t *testing.T) {
	baseURL := "https://gg5.co/tw/search"

	cases := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "relative path with query",
			raw:  "/cn/video/detail/359635?tab=preview",
			want: "https://gg5.co/cn/video/detail/359635?tab=preview",
		},
		{
			name: "pure query",
			raw:  "?tab=preview",
			want: "https://gg5.co/tw/search?tab=preview",
		},
		{
			name: "protocol relative url",
			raw:  "//cdn.example.com/poster.jpg?size=2",
			want: "https://cdn.example.com/poster.jpg?size=2",
		},
		{
			name: "relative filename with query",
			raw:  "detail.html?tab=preview",
			want: "https://gg5.co/tw/detail.html?tab=preview",
		},
		{
			name: "absolute url",
			raw:  "https://example.com/watch?id=1",
			want: "https://example.com/watch?id=1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveRelativeAVURL(baseURL, tc.raw); got != tc.want {
				t.Fatalf("resolveRelativeAVURL(%q, %q) = %q, want %q", baseURL, tc.raw, got, tc.want)
			}
		})
	}
}

func TestPreviewAVKeepsAVSexDetailURLQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/tw/search" && strings.EqualFold(r.URL.Query().Get("query"), "cawd-582"):
			_, _ = w.Write([]byte(`
<html>
  <body>
    <a href="/cn/video/detail/359635?from=search&tab=preview">
      <span class="truncate p-2 text-primary font-bold dark:text-primary-200">CAWD-582 AVSex First Impression</span>
      <img src="/covers/avsex-cover.jpg" />
    </a>
  </body>
</html>
`))
		case r.URL.Path == "/cn/video/detail/359635":
			_, _ = w.Write([]byte(`
<html>
  <head>
    <title>AVSex First Impression</title>
  </head>
  <body>
    <span class="truncate p-2 text-primary font-bold dark:text-primary-200">CAWD-582 AVSex First Impression</span>
    <dd class="flex gap-2 flex-wrap"><a href="/actor/a1">演员甲</a></dd>
  </body>
</html>
`))
		case r.URL.Path == "/covers/avsex-cover.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("fake-image"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "avsex-relative-url-test", time.Second)

	got, err := svc.PreviewAV(context.Background(), "CAWD-582")
	if err != nil {
		t.Fatalf("PreviewAV returned error: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected av candidates, got none")
	}
	wantDetailURL := server.URL + "/cn/video/detail/359635?from=search&tab=preview"
	if got[0]["detail_url"] != wantDetailURL {
		t.Fatalf("expected detail_url %q, got=%v", wantDetailURL, got[0]["detail_url"])
	}
}
