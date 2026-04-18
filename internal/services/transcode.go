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

// TranscodeProgress represents realtime transcode progress.
type TranscodeProgress struct {
	SourceDurationSeconds int
	ProcessedSeconds      int
	RemainingSeconds      int
	ProgressPercent       float64
}

const (
	transcodeModeBitrate     = "bitrate"
	transcodeModeCRFFallback = "crf_fallback"

	resolutionTier4K    = "4k"
	resolutionTier1080  = "1080"
	resolutionTierOther = "other"
)

type transcodePlan struct {
	Mode              string
	CRF               string
	ResolutionTier    string
	SourceBitrateKbps int
	TargetBitrateKbps int
	BitrateCapped     bool
}

func NewTranscodeService(storageRoot string) *TranscodeService {
	return &TranscodeService{storageRoot: storageRoot}
}

func (s *TranscodeService) Process(ctx context.Context, videoID uuid.UUID, inputPath, typ string, progressHandler func(TranscodeProgress)) (TranscodeResult, error) {
	outputDir := filepath.Join(s.storageRoot, "videos", videoID.String())
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return TranscodeResult{}, fmt.Errorf("create output dir: %w", err)
	}

	outputPath := filepath.Join(outputDir, "video.mp4")
	thumbPath := filepath.Join(outputDir, "thumb.jpg")

	inputProbe, inputProbeErr := ffmpeg.Probe(ctx, inputPath)
	sourceDurationSeconds := probeDurationSeconds(inputProbe.Duration)
	plan := buildTranscodePlan(inputProbe, inputProbeErr, typ)
	if progressHandler != nil && sourceDurationSeconds > 0 {
		progressHandler(TranscodeProgress{
			SourceDurationSeconds: sourceDurationSeconds,
			ProcessedSeconds:      0,
			RemainingSeconds:      sourceDurationSeconds,
			ProgressPercent:       0,
		})
	}
	transcodeOptions := ffmpeg.TranscodeOptions{
		CRF:              plan.CRF,
		VideoBitrateKbps: plan.TargetBitrateKbps,
		SourceDuration:   sourceDurationSeconds,
		ProgressHandler: func(progress ffmpeg.TranscodeProgress) {
			if progressHandler == nil {
				return
			}
			progressHandler(TranscodeProgress{
				SourceDurationSeconds: progress.SourceDurationSeconds,
				ProcessedSeconds:      progress.ProcessedSeconds,
				RemainingSeconds:      progress.RemainingSeconds,
				ProgressPercent:       progress.ProgressPercent,
			})
		},
	}
	if plan.Mode == transcodeModeCRFFallback {
		transcodeOptions.VideoBitrateKbps = 0
	}
	if err := ffmpeg.TranscodeHEVC(ctx, inputPath, outputPath, transcodeOptions); err != nil {
		return TranscodeResult{}, err
	}
	if err := ffmpeg.Thumbnail(ctx, outputPath, thumbPath); err != nil {
		return TranscodeResult{}, err
	}
	probe, probeErr := ffmpeg.Probe(ctx, outputPath)
	duration, width, height, metadata := resolveProbeFields(probe, probeErr, plan)

	return TranscodeResult{
		TranscodedPath: outputPath,
		ThumbnailPath:  thumbPath,
		Duration:       duration,
		Width:          width,
		Height:         height,
		Metadata:       metadata,
	}, nil
}

func resolveProbeFields(probe ffmpeg.VideoProbe, probeErr error, plan transcodePlan) (duration, width, height int, metadata map[string]any) {
	duration = 0
	width = 0
	height = 0
	metadata = map[string]any{
		"codec":               "unknown",
		"transcode_mode":      plan.Mode,
		"resolution_tier":     plan.ResolutionTier,
		"source_bitrate_kbps": plan.SourceBitrateKbps,
		"target_bitrate_kbps": plan.TargetBitrateKbps,
		"bitrate_capped":      plan.BitrateCapped,
	}
	if plan.Mode == transcodeModeCRFFallback {
		metadata["crf"] = plan.CRF
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

func probeDurationSeconds(duration float64) int {
	if duration <= 0 {
		return 0
	}
	return int(math.Round(duration))
}

func buildTranscodePlan(probe ffmpeg.VideoProbe, probeErr error, videoType string) transcodePlan {
	defaultPlan := transcodePlan{
		Mode:              transcodeModeCRFFallback,
		CRF:               chooseCRF(videoType),
		ResolutionTier:    classifyResolutionTier(probe.Width, probe.Height),
		SourceBitrateKbps: probe.BitrateKbps,
		TargetBitrateKbps: 0,
		BitrateCapped:     false,
	}
	if probeErr != nil || probe.BitrateKbps <= 0 {
		return defaultPlan
	}
	target, tier, capped := decideVideoBitrate(probe.Width, probe.Height, probe.BitrateKbps)
	return transcodePlan{
		Mode:              transcodeModeBitrate,
		CRF:               chooseCRF(videoType),
		ResolutionTier:    tier,
		SourceBitrateKbps: probe.BitrateKbps,
		TargetBitrateKbps: target,
		BitrateCapped:     capped,
	}
}

func decideVideoBitrate(width, height, sourceBitrateKbps int) (targetBitrateKbps int, tier string, capped bool) {
	tier = classifyResolutionTier(width, height)
	capped = false
	targetBitrateKbps = sourceBitrateKbps
	if sourceBitrateKbps <= 0 {
		return targetBitrateKbps, tier, capped
	}
	switch tier {
	case resolutionTier4K:
		if sourceBitrateKbps > 8000 {
			targetBitrateKbps = 8000
			capped = true
		}
	case resolutionTier1080:
		if sourceBitrateKbps > 4000 {
			targetBitrateKbps = 4000
			capped = true
		}
	}
	return targetBitrateKbps, tier, capped
}

func classifyResolutionTier(width, height int) string {
	if width >= 3840 || height >= 2160 {
		return resolutionTier4K
	}
	if width >= 1920 || height >= 1080 {
		return resolutionTier1080
	}
	return resolutionTierOther
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
