package services

import (
	"context"
	"encoding/json"
	"strings"

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

func DecodePlaybackCompatibilityMetadata(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil || payload == nil {
		return map[string]any{}
	}
	block, _ := payload[PlaybackCompatibilityMetadataKey].(map[string]any)
	if block == nil {
		return map[string]any{}
	}
	return block
}

func SourcePlaybackPathFromMetadata(raw []byte) string {
	return sourcePlaybackPathFromBlock(DecodePlaybackCompatibilityMetadata(raw))
}

func MergePlaybackCompatibilityMetadata(metadata map[string]any, patch map[string]any) map[string]any {
	merged := cloneAnyMap(metadata)
	if len(patch) == 0 {
		return merged
	}
	current, _ := merged[PlaybackCompatibilityMetadataKey].(map[string]any)
	current = cloneAnyMap(current)
	for key, value := range patch {
		current[key] = value
	}
	merged[PlaybackCompatibilityMetadataKey] = current
	return merged
}

func sourcePlaybackPathFromBlock(block map[string]any) string {
	if block == nil {
		return ""
	}
	value, _ := block["source_playback_path"].(string)
	return strings.TrimSpace(value)
}

func cloneAnyMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
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
