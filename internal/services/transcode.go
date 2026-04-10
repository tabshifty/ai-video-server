package services

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

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
	probe, probeErr := ffmpeg.Probe(ctx, outputPath)
	duration, width, height, metadata := resolveProbeFields(probe, probeErr, crf)

	return TranscodeResult{
		TranscodedPath: outputPath,
		ThumbnailPath:  thumbPath,
		Duration:       duration,
		Width:          width,
		Height:         height,
		Metadata:       metadata,
	}, nil
}

func resolveProbeFields(probe ffmpeg.VideoProbe, probeErr error, crf string) (duration, width, height int, metadata map[string]any) {
	duration = 0
	width = 0
	height = 0
	metadata = map[string]any{
		"codec": "unknown",
		"crf":   crf,
	}
	if probeErr != nil {
		metadata["probe_error"] = probeErr.Error()
		return duration, width, height, metadata
	}

	if probe.Duration > 0 {
		duration = int(math.Round(probe.Duration))
	}
	if probe.Width > 0 {
		width = probe.Width
	}
	if probe.Height > 0 {
		height = probe.Height
	}
	if codec := strings.TrimSpace(probe.Codec); codec != "" {
		metadata["codec"] = codec
	}
	return duration, width, height, metadata
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
