package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
)

func TestMinimalHTMLExternalIDPreservesLove6Case(t *testing.T) {
	got := minimalHTMLExternalID("https://love6.tv/albums/view/NDI2Mw==", "love6")
	if got != "NDI2Mw==" {
		t.Fatalf("expected love6 external id to preserve case, got=%q", got)
	}
}

func TestConfirmAVPreservesLove6ExternalIDCaseInMetadata(t *testing.T) {
	videoID := uuid.New()
	repo := &fakeScraperRepo{
		videoByID: map[uuid.UUID]models.Video{
			videoID: {ID: videoID, Title: "旧标题"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/albums/view/NDI2Mw==" || r.URL.RawQuery != "" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`<!doctype html>
<html>
<head><title>Love6 Case Title</title></head>
<body><h1>Love6 Case Title</h1></body>
</html>`))
	}))
	defer server.Close()

	svc := NewScraperService(repo, "", "https://api.themoviedb.org/3", t.TempDir(), "", 2*time.Second)
	svc.ConfigureAVScraperConfig(AVScraperConfig{
		BaseURL: server.URL,
		SiteURLs: map[string]string{
			"love6": server.URL,
		},
		UserAgent: "love6-regression-test",
		Timeout:   time.Second,
	})

	err := svc.ConfirmAV(context.Background(), ConfirmScrapeInput{
		VideoID:    videoID,
		ExternalID: "NDI2Mw==",
		Metadata: map[string]any{
			"scrape_source": "love6",
		},
	})
	if err != nil {
		t.Fatalf("ConfirmAV returned error: %v", err)
	}

	if got := repo.lastUpdate.metadata["external_id"]; got != "NDI2Mw==" {
		t.Fatalf("expected metadata external_id to preserve case, got=%v", got)
	}
	love6Meta, ok := repo.lastUpdate.metadata["love6"].(map[string]any)
	if !ok {
		t.Fatalf("expected metadata.love6 to be a map, got=%T", repo.lastUpdate.metadata["love6"])
	}
	if got := love6Meta["external_id"]; got != "NDI2Mw==" {
		t.Fatalf("expected metadata.love6.external_id to preserve case, got=%v", got)
	}
	wantDetailURL := server.URL + "/albums/view/NDI2Mw=="
	if got := repo.lastUpdate.metadata["detail_url"]; got != wantDetailURL {
		t.Fatalf("expected detail_url %q, got=%v", wantDetailURL, got)
	}
}
