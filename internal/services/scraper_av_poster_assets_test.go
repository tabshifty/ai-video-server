package services

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

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
			_, _ = w.Write(testJPEGLandscapeBytes(t))
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
	posterFilePath := asString(repo.lastUpdate.metadata["poster"])
	thumbFilePath := asString(repo.lastUpdate.metadata["thumb"])
	if strings.TrimSpace(originalURL) == "" {
		t.Fatalf("expected poster_original_path, metadata=%v", repo.lastUpdate.metadata)
	}
	if strings.TrimSpace(croppedURL) == "" {
		t.Fatalf("expected poster_cropped_path, metadata=%v", repo.lastUpdate.metadata)
	}
	if posterFilePath != originalFilePath {
		t.Fatalf("expected poster to keep original file, got=%s want=%s", posterFilePath, originalFilePath)
	}
	if thumbFilePath != croppedFilePath {
		t.Fatalf("expected thumb to use cropped file, got=%s want=%s", thumbFilePath, croppedFilePath)
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
	posterFilePath := asString(repo.lastUpdate.metadata["poster"])
	thumbFilePath := asString(repo.lastUpdate.metadata["thumb"])
	if strings.TrimSpace(originalURL) == "" {
		t.Fatalf("expected poster_original_path, metadata=%v", repo.lastUpdate.metadata)
	}
	if posterFilePath != originalFilePath {
		t.Fatalf("expected poster to keep original file, got=%s want=%s", posterFilePath, originalFilePath)
	}
	if thumbFilePath != originalFilePath {
		t.Fatalf("expected thumb to fall back to original file, got=%s want=%s", thumbFilePath, originalFilePath)
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

func TestConfirmAVPrefersCroppedPosterOverSeparateThumbWhenCropEnabled(t *testing.T) {
	t.Parallel()

	videoID := uuid.New()
	repo := &fakeScraperRepo{
		videoByID: map[uuid.UUID]models.Video{
			videoID: {
				ID:     videoID,
				Title:  "SSIS-125",
				Type:   "av",
				Status: "ready",
			},
		},
	}
	root := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v/ssis-125":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`
				<html>
					<head>
						<title>SSIS-125 Separate Thumb - JavDB</title>
						<meta property="og:image" content="` + serverURLFromRequest(r) + `/covers/ssis-125.jpg" />
					</head>
					<body>
						<h2 class="title is-4">SSIS-125 Separate Thumb</h2>
						<strong>番號</strong><span>SSIS-125</span>
						<strong>日期</strong><span>2024-01-03</span>
					</body>
				</html>
			`))
		case "/thumbs/ssis-125.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write(testJPEGPortraitBytes(t))
		case "/covers/ssis-125.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write(testJPEGLandscapeBytes(t))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(repo, "", "", root, filepath.Join(root, "posters"), 2*time.Second)
	svc.ConfigureAVScraper(server.URL, "poster-thumb-test", 2*time.Second)

	err := svc.ConfirmAV(context.Background(), ConfirmScrapeInput{
		VideoID:    videoID,
		ExternalID: "ssis-125",
		ThumbURL:   server.URL + "/thumbs/ssis-125.jpg",
		Metadata: map[string]any{
			"scrape_source": "javdb",
			"detail_url":    server.URL + "/v/ssis-125",
		},
	})
	if err != nil {
		t.Fatalf("ConfirmAV returned error: %v", err)
	}

	posterFilePath := asString(repo.lastUpdate.metadata["poster"])
	thumbFilePath := asString(repo.lastUpdate.metadata["thumb"])
	croppedFilePath := asString(repo.lastUpdate.metadata["poster_cropped_file_path"])
	if posterFilePath == "" || thumbFilePath == "" {
		t.Fatalf("expected poster and thumb paths, metadata=%v", repo.lastUpdate.metadata)
	}
	if thumbFilePath != croppedFilePath {
		t.Fatalf("expected thumb to use cropped poster, got=%s want=%s", thumbFilePath, croppedFilePath)
	}
	if asString(repo.lastUpdate.metadata["poster_variant"]) != avPosterVariantCropped {
		t.Fatalf("expected poster_variant cropped, got=%v", repo.lastUpdate.metadata["poster_variant"])
	}
	posterImg := decodeJPEGFile(t, posterFilePath)
	thumbImg := decodeJPEGFile(t, thumbFilePath)
	if posterImg.Bounds().Dx() <= posterImg.Bounds().Dy() {
		t.Fatalf("expected poster to stay landscape, got=%dx%d", posterImg.Bounds().Dx(), posterImg.Bounds().Dy())
	}
	if thumbImg.Bounds().Dx() >= thumbImg.Bounds().Dy() {
		t.Fatalf("expected thumb to use portrait crop, got=%dx%d", thumbImg.Bounds().Dx(), thumbImg.Bounds().Dy())
	}
	if _, err := os.Stat(posterFilePath); err != nil {
		t.Fatalf("expected poster file: %v", err)
	}
	if _, err := os.Stat(thumbFilePath); err != nil {
		t.Fatalf("expected thumb file: %v", err)
	}
}

func TestResolveAVPosterAssetsCropsWhenThumbURLIsNotPortrait(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/poster/landscape.jpg", "/thumb/landscape.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write(testJPEGLandscapeBytes(t))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "", "", t.TempDir(), "", 2*time.Second)
	assets, _, err := svc.resolveAVPosterAssets(
		context.Background(),
		uuid.New(),
		"",
		server.URL+"/poster/landscape.jpg",
		server.URL+"/thumb/landscape.jpg",
		"video_cover",
		avPosterQualityPrimary,
		AVScraperSiteConfig{
			PosterCropEnabled: true,
			PosterCropMode:    avPosterCropModeCenter,
		},
	)
	if err != nil {
		t.Fatalf("resolveAVPosterAssets returned error: %v", err)
	}
	if assets.OriginalPath == "" || assets.CroppedPath == "" {
		t.Fatalf("expected original and cropped paths, assets=%+v", assets)
	}
	if assets.ThumbPath != assets.CroppedPath {
		t.Fatalf("expected non-portrait thumb URL to be ignored in favor of crop, got thumb=%s cropped=%s", assets.ThumbPath, assets.CroppedPath)
	}
	if assets.Variant != avPosterVariantCropped {
		t.Fatalf("expected cropped variant, got=%s", assets.Variant)
	}

	posterImg := decodeJPEGFile(t, assets.OriginalPath)
	thumbImg := decodeJPEGFile(t, assets.ThumbPath)
	if posterImg.Bounds().Dx() <= posterImg.Bounds().Dy() {
		t.Fatalf("expected poster to stay landscape, got=%dx%d", posterImg.Bounds().Dx(), posterImg.Bounds().Dy())
	}
	if thumbImg.Bounds().Dx() >= thumbImg.Bounds().Dy() {
		t.Fatalf("expected thumb to be portrait crop, got=%dx%d", thumbImg.Bounds().Dx(), thumbImg.Bounds().Dy())
	}
}

func TestResolveAVPosterAssetsUsesConfiguredHorizontalCropAnchor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/poster/wide-anchor.jpg" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(testJPEGWideAnchorBytes(t))
	}))
	defer server.Close()

	cases := []struct {
		name          string
		mode          string
		wantLeftEdge  string
		wantRightEdge string
	}{
		{name: "center", mode: "portrait_center", wantLeftEdge: "red", wantRightEdge: "blue"},
		{name: "left", mode: "portrait_left", wantLeftEdge: "red", wantRightEdge: "green"},
		{name: "right", mode: "portrait_right", wantLeftEdge: "green", wantRightEdge: "blue"},
		{name: "invalid_falls_back_to_center", mode: "unexpected", wantLeftEdge: "red", wantRightEdge: "blue"},
		{name: "empty_falls_back_to_center", mode: "", wantLeftEdge: "red", wantRightEdge: "blue"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc := NewScraperService(nil, "", "", t.TempDir(), "", 2*time.Second)
			assets, _, err := svc.resolveAVPosterAssets(
				context.Background(),
				uuid.New(),
				"",
				server.URL+"/poster/wide-anchor.jpg",
				"",
				"manual",
				avPosterQualityPrimary,
				AVScraperSiteConfig{
					PosterCropEnabled: true,
					PosterCropMode:    tc.mode,
				},
			)
			if err != nil {
				t.Fatalf("resolveAVPosterAssets returned error: %v", err)
			}
			if assets.Variant != avPosterVariantCropped {
				t.Fatalf("expected cropped variant, got=%s", assets.Variant)
			}

			cropped := decodeJPEGFile(t, assets.CroppedPath)
			assertNamedEdgeColor(t, cropped, 10, cropped.Bounds().Dy()/2, tc.wantLeftEdge)
			assertNamedEdgeColor(t, cropped, cropped.Bounds().Dx()-11, cropped.Bounds().Dy()/2, tc.wantRightEdge)
		})
	}
}

func testJPEGLandscapeBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 360, 240))
	for y := 0; y < 240; y++ {
		for x := 0; x < 360; x++ {
			img.Set(x, y, color.RGBA{R: uint8((x * 255) / 359), G: uint8((y * 255) / 239), B: 160, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode jpeg fixture: %v", err)
	}
	return buf.Bytes()
}

func testJPEGPortraitBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 240, 360))
	for y := 0; y < 360; y++ {
		for x := 0; x < 240; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: uint8((y * 255) / 359), B: uint8((x * 255) / 239), A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode jpeg fixture: %v", err)
	}
	return buf.Bytes()
}

func testJPEGWideAnchorBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 900, 900))
	for y := 0; y < 900; y++ {
		for x := 0; x < 900; x++ {
			switch {
			case x < 300:
				img.Set(x, y, color.RGBA{R: 255, A: 255})
			case x < 600:
				img.Set(x, y, color.RGBA{G: 255, A: 255})
			default:
				img.Set(x, y, color.RGBA{B: 255, A: 255})
			}
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode wide jpeg fixture: %v", err)
	}
	return buf.Bytes()
}

func decodeJPEGFile(t *testing.T, path string) image.Image {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open jpeg file: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("decode jpeg file: %v", err)
	}
	return img
}

func assertNamedEdgeColor(t *testing.T, img image.Image, x, y int, want string) {
	t.Helper()
	r16, g16, b16, _ := img.At(x, y).RGBA()
	r := int(r16 >> 8)
	g := int(g16 >> 8)
	b := int(b16 >> 8)

	switch want {
	case "red":
		if !(r > g+40 && r > b+40) {
			t.Fatalf("expected red edge at (%d,%d), got rgb=(%d,%d,%d)", x, y, r, g, b)
		}
	case "green":
		if !(g > r+40 && g > b+40) {
			t.Fatalf("expected green edge at (%d,%d), got rgb=(%d,%d,%d)", x, y, r, g, b)
		}
	case "blue":
		if !(b > r+40 && b > g+40) {
			t.Fatalf("expected blue edge at (%d,%d), got rgb=(%d,%d,%d)", x, y, r, g, b)
		}
	default:
		t.Fatalf("unsupported expected color %q", want)
	}
}

func serverURLFromRequest(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
