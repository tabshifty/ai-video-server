package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// ConvertToWebP converts an input image to WebP format.
func ConvertToWebP(ctx context.Context, inputPath, outputPath string, quality int) error {
	if quality <= 0 || quality > 100 {
		quality = 82
	}
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-i", inputPath,
		"-c:v", "libwebp",
		"-q:v", strconv.Itoa(quality),
		"-compression_level", "6",
		"-preset", "picture",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg convert webp failed: %w, output=%s", err, string(out))
	}
	return nil
}

// ResizeImage resizes an image with optional fit mode and output format.
func ResizeImage(ctx context.Context, inputPath, outputPath string, width, height int, fit, format string, quality int) error {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	if width == 0 && height == 0 {
		return fmt.Errorf("width or height is required")
	}
	if quality <= 0 || quality > 100 {
		quality = 82
	}

	scaleExpr := buildScaleExpr(width, height, fit)
	args := []string{
		"-y",
		"-i", inputPath,
		"-vf", scaleExpr,
	}
	switch format {
	case "gif":
		args = append(args, "-f", "gif")
	case "webp":
		args = append(args,
			"-c:v", "libwebp",
			"-q:v", strconv.Itoa(quality),
			"-compression_level", "6",
			"-preset", "picture",
		)
	}
	args = append(args, outputPath)
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg resize image failed: %w, output=%s", err, string(out))
	}
	return nil
}

// CopyFile copies a file using OS file IO.
func CopyFile(dstPath, srcPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err := dst.ReadFrom(src); err != nil {
		return err
	}
	return nil
}

func buildScaleExpr(width, height int, fit string) string {
	fit = filepath.Clean("/" + fit)
	fit = fit[1:]
	if fit == "" {
		fit = "inside"
	}
	switch fit {
	case "cover":
		wExpr := "iw"
		hExpr := "ih"
		if width > 0 {
			wExpr = strconv.Itoa(width)
		}
		if height > 0 {
			hExpr = strconv.Itoa(height)
		}
		return fmt.Sprintf("scale=%s:%s:force_original_aspect_ratio=increase,crop=%s:%s", wExpr, hExpr, wExpr, hExpr)
	case "contain":
		wExpr := "-1"
		hExpr := "-1"
		if width > 0 {
			wExpr = strconv.Itoa(width)
		}
		if height > 0 {
			hExpr = strconv.Itoa(height)
		}
		return fmt.Sprintf("scale=%s:%s:force_original_aspect_ratio=decrease,pad=%s:%s:(ow-iw)/2:(oh-ih)/2", wExpr, hExpr, wExpr, hExpr)
	default:
		wExpr := "-2"
		hExpr := "-2"
		if width > 0 {
			wExpr = strconv.Itoa(width)
		}
		if height > 0 {
			hExpr = strconv.Itoa(height)
		}
		return fmt.Sprintf("scale=%s:%s:force_original_aspect_ratio=decrease", wExpr, hExpr)
	}
}
