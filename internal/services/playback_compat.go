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
	transcodeMetadata ...map[string]any,
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
	if isTrustedDVSdrOutput(source, sourceErr, output, outputErr, transcodeMetadata...) {
		metadata["trusted_compat_output"] = trustedToneMapDVSdr
		metadata["tone_mapped_sdr"] = true
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

func isTrustedDVSdrOutput(
	source ffmpeg.PlaybackCompatibilityProbe,
	sourceErr error,
	output ffmpeg.PlaybackCompatibilityProbe,
	outputErr error,
	transcodeMetadata ...map[string]any,
) bool {
	if sourceErr != nil || outputErr != nil {
		return false
	}
	if !source.VideoStreamFound || !output.VideoStreamFound {
		return false
	}
	if !source.DolbyVision || output.DolbyVision {
		return false
	}
	for _, metadata := range transcodeMetadata {
		if metadata == nil {
			continue
		}
		if stringFromMetadata(metadata["trusted_tone_map"]) != trustedToneMapDVSdr {
			continue
		}
		if boolFromMetadata(metadata["tone_mapped_sdr"]) != true {
			continue
		}
		if stringFromMetadata(metadata["tone_map_source"]) != "dolby_vision" {
			continue
		}
		if stringFromMetadata(metadata["tone_map_target"]) != "sdr_bt709" {
			continue
		}
		return true
	}
	return false
}

func playbackProbeMetadata(probe ffmpeg.PlaybackCompatibilityProbe) map[string]any {
	return map[string]any{
		"video_stream_found":     probe.VideoStreamFound,
		"codec":                  probe.Codec,
		"profile":                probe.Profile,
		"codec_tag":              probe.CodecTag,
		"pixel_format":           probe.PixelFormat,
		"color_range":            probe.ColorRange,
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

func stringFromMetadata(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func boolFromMetadata(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	default:
		return false
	}
}
