package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// VideoProbe holds basic media attributes from ffprobe.
type VideoProbe struct {
	Duration float64 `json:"duration"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Codec    string  `json:"codec"`
}

// TranscodeHEVC transcodes source into H.265/AAC output using videotoolbox.
func TranscodeHEVC(ctx context.Context, inputPath, outputPath, crf string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-i", inputPath,
		"-c:v", "hevc_videotoolbox",
		"-crf", crf,
		"-preset", "medium",
		"-c:a", "aac",
		"-b:a", "128k",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg transcode failed: %w, output=%s", err, string(out))
	}
	return nil
}

// Thumbnail captures first frame as thumbnail image.
func Thumbnail(ctx context.Context, inputPath, outputPath string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-i", inputPath,
		"-frames:v", "1",
		"-q:v", "2",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg thumbnail failed: %w, output=%s", err, string(out))
	}
	return nil
}

// Probe reads media metadata via ffprobe JSON output.
func Probe(ctx context.Context, inputPath string) (VideoProbe, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "stream=width,height,codec_name:format=duration",
		"-of", "json",
		inputPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return VideoProbe{}, fmt.Errorf("ffprobe failed: %w", err)
	}

	var raw struct {
		Streams []struct {
			Width     int    `json:"width"`
			Height    int    `json:"height"`
			CodecName string `json:"codec_name"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return VideoProbe{}, fmt.Errorf("unmarshal ffprobe output: %w", err)
	}

	probe := VideoProbe{}
	if len(raw.Streams) > 0 {
		probe.Width = raw.Streams[0].Width
		probe.Height = raw.Streams[0].Height
		probe.Codec = raw.Streams[0].CodecName
	}
	if raw.Format.Duration != "" {
		var duration float64
		if _, err := fmt.Sscanf(raw.Format.Duration, "%f", &duration); err == nil {
			probe.Duration = duration
		}
	}
	return probe, nil
}
