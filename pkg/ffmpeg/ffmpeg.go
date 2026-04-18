package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// VideoProbe holds basic media attributes from ffprobe.
type VideoProbe struct {
	Duration    float64 `json:"duration"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	Codec       string  `json:"codec"`
	BitrateKbps int     `json:"bitrate_kbps"`
}

// TranscodeOptions controls H.265 transcoding behavior.
type TranscodeOptions struct {
	CRF              string
	VideoBitrateKbps int
}

// TranscodeHEVC transcodes source into H.265/AAC output using videotoolbox.
func TranscodeHEVC(ctx context.Context, inputPath, outputPath string, options TranscodeOptions) error {
	args := []string{
		"-y",
		"-i", inputPath,
		"-c:v", "hevc_videotoolbox",
		"-preset", "medium",
	}
	if options.VideoBitrateKbps > 0 {
		bitrate := strconv.Itoa(options.VideoBitrateKbps) + "k"
		args = append(args,
			"-b:v", bitrate,
			"-maxrate", bitrate,
			"-bufsize", strconv.Itoa(options.VideoBitrateKbps*2)+"k",
		)
	} else {
		crf := strings.TrimSpace(options.CRF)
		if crf == "" {
			crf = "23"
		}
		args = append(args, "-crf", crf)
	}
	args = append(args,
		"-c:a", "aac",
		"-b:a", "128k",
		outputPath,
	)

	cmd := exec.CommandContext(ctx, "ffmpeg",
		args...,
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

// ThumbnailAt captures a frame at target time (seconds) as thumbnail image.
func ThumbnailAt(ctx context.Context, inputPath, outputPath string, atSeconds float64) error {
	if atSeconds < 0 {
		atSeconds = 0
	}
	timeArg := strconv.FormatFloat(atSeconds, 'f', 3, 64)
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-ss", timeArg,
		"-i", inputPath,
		"-frames:v", "1",
		"-q:v", "2",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg thumbnail at failed: %w, output=%s", err, string(out))
	}
	return nil
}

// Probe reads media metadata via ffprobe JSON output.
func Probe(ctx context.Context, inputPath string) (VideoProbe, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,codec_name,bit_rate:format=duration,bit_rate",
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
			BitRate   string `json:"bit_rate"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
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
		probe.BitrateKbps = parseBitrateKbps(raw.Streams[0].BitRate, raw.Format.BitRate)
	}
	if raw.Format.Duration != "" {
		var duration float64
		if _, err := fmt.Sscanf(raw.Format.Duration, "%f", &duration); err == nil {
			probe.Duration = duration
		}
	}
	if probe.BitrateKbps == 0 {
		probe.BitrateKbps = parseBitrateKbps("", raw.Format.BitRate)
	}
	return probe, nil
}

func parseBitrateKbps(streamBitrate, formatBitrate string) int {
	bitsPerSecond := parseBitrateBits(streamBitrate)
	if bitsPerSecond <= 0 {
		bitsPerSecond = parseBitrateBits(formatBitrate)
	}
	if bitsPerSecond <= 0 {
		return 0
	}
	return int(bitsPerSecond / 1000)
}

func parseBitrateBits(raw string) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0
	}
	return value
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
