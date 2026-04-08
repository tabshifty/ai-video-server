package services

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"video-server/pkg/ffmpeg"
)

// TranscodeResult captures generated assets and parsed metadata.
type TranscodeResult struct {
	TranscodedPath string
	ThumbnailPath  string
	Duration       int
	Width          int
	Height         int
	Metadata       map[string]any
}

// TranscodeService performs ffmpeg processing for uploaded videos.
type TranscodeService struct {
	storageRoot string
}

func NewTranscodeService(storageRoot string) *TranscodeService {
	return &TranscodeService{storageRoot: storageRoot}
}

func (s *TranscodeService) Process(ctx context.Context, videoID uuid.UUID, inputPath, typ string) (TranscodeResult, error) {
	outputDir := filepath.Join(s.storageRoot, "videos", videoID.String())
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return TranscodeResult{}, fmt.Errorf("create output dir: %w", err)
	}

	outputPath := filepath.Join(outputDir, "video.mp4")
	thumbPath := filepath.Join(outputDir, "thumb.jpg")

	crf := chooseCRF(typ)
	if err := ffmpeg.TranscodeHEVC(ctx, inputPath, outputPath, crf); err != nil {
		return TranscodeResult{}, err
	}
	if err := ffmpeg.Thumbnail(ctx, outputPath, thumbPath); err != nil {
		return TranscodeResult{}, err
	}
	probe, err := ffmpeg.Probe(ctx, outputPath)
	if err != nil {
		return TranscodeResult{}, err
	}

	return TranscodeResult{
		TranscodedPath: outputPath,
		ThumbnailPath:  thumbPath,
		Duration:       int(math.Round(probe.Duration)),
		Width:          probe.Width,
		Height:         probe.Height,
		Metadata: map[string]any{
			"codec": probe.Codec,
			"crf":   crf,
		},
	}, nil
}

func chooseCRF(videoType string) string {
	switch videoType {
	case "short":
		return "26"
	case "movie":
		return "21"
	default:
		return "23"
	}
}
