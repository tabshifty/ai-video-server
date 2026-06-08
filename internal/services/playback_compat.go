package services

import (
	"context"

	"video-server/pkg/ffmpeg"
)

const (
	// PlaybackCompatibilityMetadataKey is the video metadata key for TV playback compatibility probes.
	PlaybackCompatibilityMetadataKey = "playback_compat"

	playbackCompatibilityVersion           = 1
	playbackCompatibilityStatusOK          = "ok"
	playbackCompatibilityStatusProbeFailed = "probe_failed"
)

// ProbePlaybackCompatibility reads playback compatibility metadata without modifying media files.
func ProbePlaybackCompatibility(ctx context.Context, inputPath string) (ffmpeg.PlaybackCompatibilityProbe, error) {
	return ffmpeg.ProbePlaybackCompatibility(ctx, inputPath)
}

// BuildPlaybackCompatibilityMetadata builds the persisted videos.metadata.playback_compat object.
func BuildPlaybackCompatibilityMetadata(
	source ffmpeg.PlaybackCompatibilityProbe,
	sourceErr error,
	output ffmpeg.PlaybackCompatibilityProbe,
	outputErr error,
) map[string]any {
	metadata := map[string]any{
		"version":            playbackCompatibilityVersion,
		"status":             playbackCompatibilityStatusOK,
		"dolby_vision_risk":  source.DolbyVision || output.DolbyVision,
		"source":             playbackProbeMetadata(source),
		"output":             playbackProbeMetadata(output),
		"source_probe_ok":    sourceErr == nil && source.VideoStreamFound,
		"output_probe_ok":    outputErr == nil && output.VideoStreamFound,
		"source_video_found": source.VideoStreamFound,
		"output_video_found": output.VideoStreamFound,
	}
	if sourceErr != nil {
		metadata["status"] = playbackCompatibilityStatusProbeFailed
		metadata["source_probe_error"] = sourceErr.Error()
	} else if !source.VideoStreamFound {
		metadata["status"] = playbackCompatibilityStatusProbeFailed
		metadata["source_probe_error"] = "source video stream not found"
	}
	if outputErr != nil {
		metadata["status"] = playbackCompatibilityStatusProbeFailed
		metadata["output_probe_error"] = outputErr.Error()
	} else if !output.VideoStreamFound {
		metadata["status"] = playbackCompatibilityStatusProbeFailed
		metadata["output_probe_error"] = "output video stream not found"
	}
	return metadata
}

func playbackProbeMetadata(probe ffmpeg.PlaybackCompatibilityProbe) map[string]any {
	return map[string]any{
		"video_stream_found":     probe.VideoStreamFound,
		"codec":                  probe.Codec,
		"profile":                probe.Profile,
		"codec_tag":              probe.CodecTag,
		"pixel_format":           probe.PixelFormat,
		"color_transfer":         probe.ColorTransfer,
		"color_primaries":        probe.ColorPrimaries,
		"color_space":            probe.ColorSpace,
		"dolby_vision":           probe.DolbyVision,
		"dolby_vision_profile":   probe.DolbyVisionProfile,
		"dolby_vision_level":     probe.DolbyVisionLevel,
		"dolby_vision_compat_id": probe.DolbyVisionCompatID,
		"hdr":                    probe.HDR,
	}
}
