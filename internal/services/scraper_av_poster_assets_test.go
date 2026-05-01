package services

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"image"
	"image/color"
	"image/jpeg"
	"video-server/internal/models"
)

func TestConfirmAVStoresOriginalAndCroppedPosterAssets(t *testing.T) {
	t.Parallel()

	videoID := uuid.New()
	repo := &fakeScraperRepo{
		videoByID: map[uuid.UUID]models.Video{
			videoID: {
				ID:            videoID,
				Title:         "SSIS-123",
				Type:          "av",
				Status:        "ready",
				ThumbnailPath: "/keep/legacy.jpg",
			},
		},
	}
	root := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v/ssis-123":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`
				<html>
					<head>
						<title>SSIS-123 Example Title - JavDB</title>
						<meta property="og:image" content="` + serverURLFromRequest(r) + `/poster/ssis-123.jpg" />
					</head>
					<body>
						<h2 class="title is-4">SSIS-123 Example Title</h2>
						<strong>番號</strong><span>SSIS-123</span>
						<strong>日期</strong><span>2024-01-02</span>
						<div class="video-detail">作品简介</div>
					</body>
				</html>
			`))
		case "/poster/ssis-123.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write(testJPEGPortraitBytes(t))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(repo, "", "", root, filepath.Join(root, "posters"), 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "poster-assets-test", 2*time.Second)

	err := svc.ConfirmAV(context.Background(), ConfirmScrapeInput{
		VideoID:    videoID,
		ExternalID: "ssis-123",
		Metadata: map[string]any{
			"scrape_source": "javdb",
			"detail_url":    server.URL + "/v/ssis-123",
		},
	})
	if err != nil {
		t.Fatalf("ConfirmAV returned error: %v", err)
	}

	originalURL := asString(repo.lastUpdate.metadata["poster_original_path"])
	croppedURL := asString(repo.lastUpdate.metadata["poster_cropped_path"])
	originalFilePath := asString(repo.lastUpdate.metadata["poster_original_file_path"])
	croppedFilePath := asString(repo.lastUpdate.metadata["poster_cropped_file_path"])
	if strings.TrimSpace(originalURL) == "" {
		t.Fatalf("expected poster_original_path, metadata=%v", repo.lastUpdate.metadata)
	}
	if strings.TrimSpace(croppedURL) == "" {
		t.Fatalf("expected poster_cropped_path, metadata=%v", repo.lastUpdate.metadata)
	}
	if repo.lastUpdate.metadata["poster_variant"] != "cropped" {
		t.Fatalf("expected poster_variant cropped, got=%v", repo.lastUpdate.metadata["poster_variant"])
	}
	if repo.lastUpdate.thumbnailPath != croppedFilePath {
		t.Fatalf("expected thumbnail_path to use cropped asset, got=%s want=%s", repo.lastUpdate.thumbnailPath, croppedFilePath)
	}
	if _, err := os.Stat(originalFilePath); err != nil {
		t.Fatalf("expected original poster file: %v", err)
	}
	if _, err := os.Stat(croppedFilePath); err != nil {
		t.Fatalf("expected cropped poster file: %v", err)
	}
}

func TestConfirmAVFallsBackToOriginalPosterWhenCropFails(t *testing.T) {
	t.Parallel()

	videoID := uuid.New()
	repo := &fakeScraperRepo{
		videoByID: map[uuid.UUID]models.Video{
			videoID: {
				ID:     videoID,
				Title:  "SSIS-124",
				Type:   "av",
				Status: "ready",
			},
		},
	}
	root := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v/ssis-124":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`
				<html>
					<head>
						<title>SSIS-124 Broken Poster - JavDB</title>
						<meta property="og:image" content="` + serverURLFromRequest(r) + `/poster/ssis-124.jpg" />
					</head>
					<body>
						<h2 class="title is-4">SSIS-124 Broken Poster</h2>
					</body>
				</html>
			`))
		case "/poster/ssis-124.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("not-a-jpeg"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(repo, "", "", root, filepath.Join(root, "posters"), 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "poster-fallback-test", 2*time.Second)

	err := svc.ConfirmAV(context.Background(), ConfirmScrapeInput{
		VideoID:    videoID,
		ExternalID: "ssis-124",
		Metadata: map[string]any{
			"scrape_source": "javdb",
			"detail_url":    server.URL + "/v/ssis-124",
		},
	})
	if err != nil {
		t.Fatalf("ConfirmAV returned error: %v", err)
	}

	originalURL := asString(repo.lastUpdate.metadata["poster_original_path"])
	croppedURL := asString(repo.lastUpdate.metadata["poster_cropped_path"])
	originalFilePath := asString(repo.lastUpdate.metadata["poster_original_file_path"])
	if strings.TrimSpace(originalURL) == "" {
		t.Fatalf("expected poster_original_path, metadata=%v", repo.lastUpdate.metadata)
	}
	if croppedURL != "" {
		t.Fatalf("expected empty poster_cropped_path on crop failure, got=%s", croppedURL)
	}
	if repo.lastUpdate.metadata["poster_variant"] != "original" {
		t.Fatalf("expected poster_variant original, got=%v", repo.lastUpdate.metadata["poster_variant"])
	}
	if repo.lastUpdate.thumbnailPath != originalFilePath {
		t.Fatalf("expected thumbnail path to fall back to original, got=%s want=%s", repo.lastUpdate.thumbnailPath, originalFilePath)
	}
}

func testJPEGPortraitBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 240, 360))
	for y := 0; y < 360; y++ {
		for x := 0; x < 240; x++ {
			img.Set(x, y, color.RGBA{R: uint8((x * 255) / 239), G: uint8((y * 255) / 359), B: 160, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode jpeg fixture: %v", err)
	}
	return buf.Bytes()
}

func serverURLFromRequest(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
