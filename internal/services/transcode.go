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
	TranscodedPath       string
	CompatTranscodedPath string
	ThumbnailPath        string
	Duration             int
	Width                int
	Height               int
	Metadata             map[string]any
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
	Mode                string
	CRF                 string
	ResolutionTier      string
	SourceBitrateKbps   int
	TargetBitrateKbps   int
	BitrateCapped       bool
	SourceAudioChannels int
}

const (
	playbackProfilePrimary = "primary"
	playbackProfileCompat  = "compat"
)

type transcodeOutputProfile struct {
	Key   string
	Path  string
	Codec string
}

func NewTranscodeService(storageRoot string) *TranscodeService {
	return &TranscodeService{storageRoot: storageRoot}
}

func (s *TranscodeService) Process(ctx context.Context, videoID uuid.UUID, inputPath, typ string, progressHandler func(TranscodeProgress)) (TranscodeResult, error) {
	outputDir := filepath.Join(s.storageRoot, "videos", videoID.String())
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return TranscodeResult{}, fmt.Errorf("create output dir: %w", err)
	}

	outputPath := filepath.Join(outputDir, "video-hevc.mp4")
	compatPath := filepath.Join(outputDir, "video-avc.mp4")
	thumbPath := filepath.Join(outputDir, "thumb.jpg")
	sourcePath := inputPath
	cleanupSourceCopy := func() {}
	if isSameFilePath(inputPath, outputPath) || isSameFilePath(inputPath, compatPath) {
		sourcePath = buildTranscodeOutputTempPath(filepath.Join(outputDir, "source-copy.mp4"))
		if err := ffmpeg.CopyFile(sourcePath, inputPath); err != nil {
			return TranscodeResult{}, fmt.Errorf("copy stable transcode input: %w", err)
		}
		cleanupSourceCopy = func() {
			_ = os.Remove(sourcePath)
		}
	}
	defer cleanupSourceCopy()
	transcodeOutputPath := outputPath
	useTempOutput := isSameFilePath(inputPath, outputPath)
	if useTempOutput {
		transcodeOutputPath = buildTranscodeOutputTempPath(outputPath)
	}
	compatOutputPath := compatPath
	useCompatTempOutput := isSameFilePath(inputPath, compatPath)
	if useCompatTempOutput {
		compatOutputPath = buildTranscodeOutputTempPath(compatPath)
	}

	inputProbe, inputProbeErr := ffmpeg.Probe(ctx, sourcePath)
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
	if err := ffmpeg.TranscodeVideo(ctx, sourcePath, transcodeOutputPath, ffmpeg.TranscodeProfileHEVCPrimary, transcodeOptions); err != nil {
		if useTempOutput {
			_ = os.Remove(transcodeOutputPath)
		}
		return TranscodeResult{}, err
	}
	if useTempOutput {
		if err := replaceOutputFile(transcodeOutputPath, outputPath); err != nil {
			_ = os.Remove(transcodeOutputPath)
			return TranscodeResult{}, fmt.Errorf("replace transcode output file: %w", err)
		}
	}
	if err := ffmpeg.TranscodeVideo(ctx, sourcePath, compatOutputPath, ffmpeg.TranscodeProfileAVCCompat, transcodeOptions); err != nil {
		if useCompatTempOutput {
			_ = os.Remove(compatOutputPath)
		}
		return TranscodeResult{}, err
	}
	if useCompatTempOutput {
		if err := replaceOutputFile(compatOutputPath, compatPath); err != nil {
			_ = os.Remove(compatOutputPath)
			return TranscodeResult{}, fmt.Errorf("replace compat transcode output file: %w", err)
		}
	}
	if err := ffmpeg.Thumbnail(ctx, outputPath, thumbPath); err != nil {
		return TranscodeResult{}, err
	}
	probe, probeErr := ffmpeg.Probe(ctx, outputPath)
	duration, width, height, metadata := resolveProbeFields(probe, probeErr, plan)
	for key, value := range buildPlaybackProfilesMetadata(
		transcodeOutputProfile{
			Key:   playbackProfilePrimary,
			Path:  outputPath,
			Codec: "hevc",
		},
		transcodeOutputProfile{
			Key:   playbackProfileCompat,
			Path:  compatPath,
			Codec: "h264",
		},
	) {
		metadata[key] = value
	}

	return TranscodeResult{
		TranscodedPath:       outputPath,
		CompatTranscodedPath: compatPath,
		ThumbnailPath:        thumbPath,
		Duration:             duration,
		Width:                width,
		Height:               height,
		Metadata:             metadata,
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
	audioCodec, audioChannels, audioDownmixed := resolvePlaybackAudioMetadata(probe, plan)
	metadata["audio_codec"] = audioCodec
	metadata["audio_channels"] = audioChannels
	metadata["audio_downmixed"] = audioDownmixed
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

func resolvePlaybackAudioMetadata(probe ffmpeg.VideoProbe, plan transcodePlan) (codec string, channels int, downmixed bool) {
	codec = "aac"
	channels = 2
	if value := strings.TrimSpace(probe.AudioCodec); value != "" {
		codec = value
	}
	if probe.AudioChannels > 0 {
		channels = probe.AudioChannels
	}
	downmixed = plan.SourceAudioChannels > channels && channels > 0
	return codec, channels, downmixed
}

func buildPlaybackProfilesMetadata(primary, compat transcodeOutputProfile) map[string]any {
	profiles := map[string]any{
		primary.Key: map[string]any{
			"path":  primary.Path,
			"codec": primary.Codec,
		},
	}
	metadata := map[string]any{
		"primary_codec":          primary.Codec,
		"compat_available":       false,
		"playback_profiles":      profiles,
		"compat_transcoded_path": "",
	}
	if compat.Path != "" {
		profiles[compat.Key] = map[string]any{
			"path":  compat.Path,
			"codec": compat.Codec,
		}
		metadata["compat_available"] = true
		metadata["compat_codec"] = compat.Codec
		metadata["compat_transcoded_path"] = compat.Path
	}
	return metadata
}

func probeDurationSeconds(duration float64) int {
	if duration <= 0 {
		return 0
	}
	return int(math.Round(duration))
}

func buildTranscodePlan(probe ffmpeg.VideoProbe, probeErr error, videoType string) transcodePlan {
	defaultPlan := transcodePlan{
		Mode:                transcodeModeCRFFallback,
		CRF:                 chooseCRF(videoType),
		ResolutionTier:      classifyResolutionTier(probe.Width, probe.Height),
		SourceBitrateKbps:   probe.BitrateKbps,
		TargetBitrateKbps:   0,
		BitrateCapped:       false,
		SourceAudioChannels: probe.AudioChannels,
	}
	if probeErr != nil || probe.BitrateKbps <= 0 {
		return defaultPlan
	}
	target, tier, capped := decideVideoBitrate(probe.Width, probe.Height, probe.BitrateKbps)
	return transcodePlan{
		Mode:                transcodeModeBitrate,
		CRF:                 chooseCRF(videoType),
		ResolutionTier:      tier,
		SourceBitrateKbps:   probe.BitrateKbps,
		TargetBitrateKbps:   target,
		BitrateCapped:       capped,
		SourceAudioChannels: probe.AudioChannels,
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

func isSameFilePath(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" || b == "" {
		return false
	}
	aAbs, aErr := filepath.Abs(filepath.Clean(a))
	bAbs, bErr := filepath.Abs(filepath.Clean(b))
	if aErr != nil || bErr != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return aAbs == bAbs
}

func buildTranscodeOutputTempPath(target string) string {
	ext := filepath.Ext(target)
	if ext == "" {
		ext = ".mp4"
	}
	base := strings.TrimSuffix(filepath.Base(target), filepath.Ext(target))
	if base == "" {
		base = "video"
	}
	return filepath.Join(filepath.Dir(target), fmt.Sprintf(".%s.retranscode.%s%s", base, uuid.NewString(), ext))
}

func replaceOutputFile(tempPath, outputPath string) error {
	if err := os.Rename(tempPath, outputPath); err == nil {
		return nil
	}
	_ = os.Remove(outputPath)
	return os.Rename(tempPath, outputPath)
}
