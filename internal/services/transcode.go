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

	transcodeProfileHEVCLongform = "hevc_longform"
	transcodeProfileAVCCompat    = "avc_compat"
	transcodeProfileDVSdrCompat  = "dv_sdr_compat"

	trustedToneMapDVSdr = "dv_sdr_bt709"

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
	TranscodeProfile    string
}

type transcodeOutputProfile struct {
	Path             string
	PlaybackCodec    string
	TranscodeProfile string
	FFmpegProfile    ffmpeg.TranscodeProfile
	SpatialAQ        bool
	TrustedToneMap   string
	CompatOutput     bool
}

type TranscodeProcessOptions struct {
	SourcePlaybackProbe   ffmpeg.PlaybackCompatibilityProbe
	SourcePlaybackProbeOK bool
}

func NewTranscodeService(storageRoot string) *TranscodeService {
	return &TranscodeService{storageRoot: storageRoot}
}

func (s *TranscodeService) Process(ctx context.Context, videoID uuid.UUID, inputPath, typ string, progressHandler func(TranscodeProgress)) (TranscodeResult, error) {
	return s.ProcessWithOptions(ctx, videoID, inputPath, typ, TranscodeProcessOptions{}, progressHandler)
}

func (s *TranscodeService) ProcessWithOptions(ctx context.Context, videoID uuid.UUID, inputPath, typ string, options TranscodeProcessOptions, progressHandler func(TranscodeProgress)) (TranscodeResult, error) {
	outputDir := filepath.Join(s.storageRoot, "videos", videoID.String())
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return TranscodeResult{}, fmt.Errorf("create output dir: %w", err)
	}

	outputProfile := chooseTranscodeOutputProfile(outputDir, typ, options)
	outputPath := outputProfile.Path
	thumbPath := filepath.Join(outputDir, "thumb.jpg")
	sourcePath := inputPath
	cleanupSourceCopy := func() {}
	if isSameFilePath(inputPath, outputPath) {
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

	inputProbe, inputProbeErr := ffmpeg.Probe(ctx, sourcePath)
	sourceDurationSeconds := probeDurationSeconds(inputProbe.Duration)
	plan := buildTranscodePlan(inputProbe, inputProbeErr, typ)
	plan.TranscodeProfile = outputProfile.TranscodeProfile
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
		SpatialAQ:        outputProfile.SpatialAQ,
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
	if err := ffmpeg.TranscodeVideo(ctx, sourcePath, transcodeOutputPath, outputProfile.FFmpegProfile, transcodeOptions); err != nil {
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
	if err := ffmpeg.Thumbnail(ctx, outputPath, thumbPath); err != nil {
		return TranscodeResult{}, err
	}
	probe, probeErr := ffmpeg.Probe(ctx, outputPath)
	duration, width, height, metadata := resolveProbeFields(probe, probeErr, plan)
	outputProfile.Path = outputPath
	for key, value := range buildPlaybackMetadata(outputProfile) {
		metadata[key] = value
	}

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
		"transcode_profile":   resolveTranscodeProfileMetadata(plan),
	}
	audioCodec, audioChannels, audioDownmixed := resolvePlaybackAudioMetadata(probe, plan)
	metadata["audio_codec"] = audioCodec
	metadata["audio_channels"] = audioChannels
	metadata["audio_track_count"] = probe.AudioTrackCount
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
	channels = 0
	if value := strings.TrimSpace(probe.AudioCodec); value != "" {
		codec = value
	}
	if probe.AudioChannels > 0 {
		channels = probe.AudioChannels
	}
	downmixed = false
	return codec, channels, downmixed
}

func buildPlaybackMetadata(output transcodeOutputProfile) map[string]any {
	metadata := map[string]any{
		"playback_path":  output.Path,
		"playback_codec": output.PlaybackCodec,
	}
	if output.CompatOutput {
		metadata["compat_transcoded_path"] = output.Path
	}
	if output.TrustedToneMap != "" {
		metadata["trusted_tone_map"] = output.TrustedToneMap
		metadata["tone_map_source"] = "dolby_vision"
		metadata["tone_map_target"] = "sdr_bt709"
		metadata["tone_mapped_sdr"] = true
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
	profile := chooseTranscodeProfileName(videoType)
	defaultPlan := transcodePlan{
		Mode:                transcodeModeCRFFallback,
		CRF:                 chooseCRF(videoType),
		ResolutionTier:      classifyResolutionTier(probe.Width, probe.Height),
		SourceBitrateKbps:   probe.BitrateKbps,
		TargetBitrateKbps:   0,
		BitrateCapped:       false,
		SourceAudioChannels: probe.AudioChannels,
		TranscodeProfile:    profile,
	}
	if probeErr != nil || probe.BitrateKbps <= 0 {
		return defaultPlan
	}
	target, tier, capped := decideVideoBitrate(videoType, probe.Width, probe.Height, probe.BitrateKbps)
	return transcodePlan{
		Mode:                transcodeModeBitrate,
		CRF:                 chooseCRF(videoType),
		ResolutionTier:      tier,
		SourceBitrateKbps:   probe.BitrateKbps,
		TargetBitrateKbps:   target,
		BitrateCapped:       capped,
		SourceAudioChannels: probe.AudioChannels,
		TranscodeProfile:    profile,
	}
}

func decideVideoBitrate(videoType string, width, height, sourceBitrateKbps int) (targetBitrateKbps int, tier string, capped bool) {
	tier = classifyResolutionTier(width, height)
	capped = false
	targetBitrateKbps = sourceBitrateKbps
	if sourceBitrateKbps <= 0 {
		return targetBitrateKbps, tier, capped
	}

	cap4K := 8000
	cap1080 := 4000
	if isLongformVideoType(videoType) {
		cap4K = 10000
		cap1080 = 5000
	}

	switch tier {
	case resolutionTier4K:
		if sourceBitrateKbps > cap4K {
			targetBitrateKbps = cap4K
			capped = true
		}
	case resolutionTier1080:
		if sourceBitrateKbps > cap1080 {
			targetBitrateKbps = cap1080
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
	switch normalizeVideoType(videoType) {
	case "short":
		return "26"
	default:
		return "23"
	}
}

func chooseTranscodeOutputProfile(outputDir, videoType string, options TranscodeProcessOptions) transcodeOutputProfile {
	if shouldUseDVSdrCompatOutput(options) {
		return transcodeOutputProfile{
			Path:             filepath.Join(outputDir, "video-dv-sdr.mp4"),
			PlaybackCodec:    "h264",
			TranscodeProfile: transcodeProfileDVSdrCompat,
			FFmpegProfile:    ffmpeg.TranscodeProfileDVSdrCompat,
			SpatialAQ:        false,
			TrustedToneMap:   trustedToneMapDVSdr,
			CompatOutput:     true,
		}
	}
	if isLongformVideoType(videoType) {
		return transcodeOutputProfile{
			Path:             filepath.Join(outputDir, "video-hevc.mp4"),
			PlaybackCodec:    "hevc",
			TranscodeProfile: transcodeProfileHEVCLongform,
			FFmpegProfile:    ffmpeg.TranscodeProfileHEVCPrimary,
			SpatialAQ:        true,
		}
	}
	return transcodeOutputProfile{
		Path:             filepath.Join(outputDir, "video-avc.mp4"),
		PlaybackCodec:    "h264",
		TranscodeProfile: transcodeProfileAVCCompat,
		FFmpegProfile:    ffmpeg.TranscodeProfileAVCCompat,
		SpatialAQ:        false,
	}
}

func shouldUseDVSdrCompatOutput(options TranscodeProcessOptions) bool {
	return options.SourcePlaybackProbeOK &&
		options.SourcePlaybackProbe.VideoStreamFound &&
		options.SourcePlaybackProbe.DolbyVision
}

func chooseTranscodeProfileName(videoType string) string {
	if isLongformVideoType(videoType) {
		return transcodeProfileHEVCLongform
	}
	return transcodeProfileAVCCompat
}

func resolveTranscodeProfileMetadata(plan transcodePlan) string {
	if profile := strings.TrimSpace(plan.TranscodeProfile); profile != "" {
		return profile
	}
	return transcodeProfileAVCCompat
}

func isLongformVideoType(videoType string) bool {
	switch normalizeVideoType(videoType) {
	case "movie", "episode":
		return true
	default:
		return false
	}
}

func normalizeVideoType(videoType string) string {
	return strings.ToLower(strings.TrimSpace(videoType))
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
