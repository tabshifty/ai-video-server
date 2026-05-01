package services

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type avPosterAssets struct {
	SelectedPath string
	OriginalPath string
	CroppedPath  string
	Variant      string
}

func (s *ScraperService) resolveAVPosterAssets(ctx context.Context, videoID uuid.UUID, existingThumbPath, posterURL, posterSource, posterQuality string, cropCfg AVScraperSiteConfig) (avPosterAssets, string, error) {
	posterURL = strings.TrimSpace(posterURL)
	existingThumbPath = strings.TrimSpace(existingThumbPath)
	posterSource = strings.ToLower(strings.TrimSpace(posterSource))
	posterQuality = strings.ToLower(strings.TrimSpace(posterQuality))
	if posterQuality == "" {
		posterQuality = classifyAVPosterURL(posterURL, posterSource)
	}

	keepExisting := func(decision string) (avPosterAssets, string, error) {
		return avPosterAssets{SelectedPath: existingThumbPath}, decision, nil
	}

	switch posterQuality {
	case avPosterQualityPrimary, avPosterQualityFallback:
		if !isAbsoluteHTTPURL(posterURL) {
			if existingThumbPath != "" {
				return keepExisting("invalid_keep_old")
			}
			return avPosterAssets{}, "invalid_no_existing", nil
		}
		originalPath, err := s.downloadPosterVariant(ctx, posterURL, videoID, "original")
		if err != nil {
			return avPosterAssets{}, "", err
		}
		assets := avPosterAssets{
			SelectedPath: originalPath,
			OriginalPath: originalPath,
			Variant:      avPosterVariantOriginal,
		}
		if cropCfg.PosterCropEnabled {
			croppedPath, cropErr := s.cropAVPosterImage(originalPath, videoID, cropCfg.PosterCropMode)
			if cropErr == nil {
				assets.SelectedPath = croppedPath
				assets.CroppedPath = croppedPath
				assets.Variant = avPosterVariantCropped
			}
		}
		if posterQuality == avPosterQualityPrimary {
			return assets, "primary_selected", nil
		}
		if existingThumbPath != "" {
			return assets, "fallback_replaced_existing", nil
		}
		return assets, "fallback_used_no_existing", nil
	default:
		if existingThumbPath != "" {
			return keepExisting("invalid_keep_old")
		}
		return avPosterAssets{}, "invalid_no_existing", nil
	}
}

func (s *ScraperService) downloadPosterVariant(ctx context.Context, posterURL string, videoID uuid.UUID, variant string) (string, error) {
	root := s.posterRoot
	if strings.TrimSpace(root) == "" {
		root = filepath.Join(s.storageRoot, "posters")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", fmt.Errorf("create poster root: %w", err)
	}
	filename := videoID.String() + ".jpg"
	if variant != "" {
		filename = videoID.String() + "-" + strings.TrimSpace(variant) + ".jpg"
	}
	outputPath := filepath.Join(root, filename)
	if err := s.downloadPosterToPath(ctx, posterURL, outputPath); err != nil {
		return "", err
	}
	return outputPath, nil
}

func (s *ScraperService) cropAVPosterImage(originalPath string, videoID uuid.UUID, cropMode string) (string, error) {
	file, err := os.Open(originalPath)
	if err != nil {
		return "", fmt.Errorf("open original poster: %w", err)
	}
	defer file.Close()

	src, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("decode original poster: %w", err)
	}
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return "", fmt.Errorf("invalid poster size")
	}

	targetRatio := 2.0 / 3.0
	cropWidth := width
	cropHeight := height
	if float64(width)/float64(height) > targetRatio {
		cropWidth = int(float64(height) * targetRatio)
		if cropWidth <= 0 {
			cropWidth = width
		}
	} else {
		cropHeight = int(float64(width) / targetRatio)
		if cropHeight <= 0 {
			cropHeight = height
		}
	}

	offsetX := bounds.Min.X
	if width > cropWidth {
		switch normalizeAVPosterCropMode(cropMode) {
		case avPosterCropModeLeft:
			offsetX = bounds.Min.X
		case avPosterCropModeRight:
			offsetX = bounds.Max.X - cropWidth
		default:
			offsetX = bounds.Min.X + (width-cropWidth)/2
		}
	}
	offsetY := bounds.Min.Y + maxInt((height-cropHeight)/10, 0)
	if offsetY+cropHeight > bounds.Max.Y {
		offsetY = bounds.Max.Y - cropHeight
	}
	rect := image.Rect(0, 0, cropWidth, cropHeight)
	dst := image.NewRGBA(rect)
	draw.Draw(dst, rect, src, image.Point{X: offsetX, Y: offsetY}, draw.Src)

	root := s.posterRoot
	if strings.TrimSpace(root) == "" {
		root = filepath.Join(s.storageRoot, "posters")
	}
	outputPath := filepath.Join(root, videoID.String()+"-cropped.jpg")
	if err := writeJPEG(outputPath, dst); err != nil {
		return "", err
	}
	return outputPath, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func writeJPEG(path string, src image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create cropped poster: %w", err)
	}
	defer file.Close()
	if err := jpeg.Encode(file, src, &jpeg.Options{Quality: 90}); err != nil {
		return fmt.Errorf("encode cropped poster: %w", err)
	}
	return nil
}
