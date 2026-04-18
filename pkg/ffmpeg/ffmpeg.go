package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
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
	SourceDuration   int
	ProgressHandler  func(TranscodeProgress)
}

// TranscodeProgress represents ffmpeg realtime progress.
type TranscodeProgress struct {
	SourceDurationSeconds int     `json:"source_duration_seconds"`
	ProcessedSeconds      int     `json:"processed_seconds"`
	RemainingSeconds      int     `json:"remaining_seconds"`
	ProgressPercent       float64 `json:"progress_percent"`
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
		"-progress", "pipe:1",
		"-nostats",
		outputPath,
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg setup stdout pipe: %w", err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg start failed: %w", err)
	}

	progressErrCh := make(chan error, 1)
	go func() {
		progressErrCh <- readTranscodeProgress(stdout, options.SourceDuration, options.ProgressHandler)
	}()

	waitErr := cmd.Wait()
	progressErr := <-progressErrCh
	if progressErr != nil {
		return fmt.Errorf("ffmpeg progress read failed: %w", progressErr)
	}
	if waitErr != nil {
		return fmt.Errorf("ffmpeg transcode failed: %w, output=%s", waitErr, stderr.String())
	}
	return nil
}

func readTranscodeProgress(r io.Reader, sourceDuration int, handler func(TranscodeProgress)) error {
	if handler == nil {
		_, err := io.Copy(io.Discard, r)
		return err
	}
	scanner := bufio.NewScanner(r)
	buffer := make([]byte, 0, 128*1024)
	scanner.Buffer(buffer, 1024*1024)

	processedSeconds := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		seconds, parsed := parseProgressValueToSeconds(key, value)
		if parsed {
			processedSeconds = seconds
			continue
		}
		if key == "progress" {
			handler(buildTranscodeProgress(processedSeconds, sourceDuration))
		}
	}
	return scanner.Err()
}

func parseProgressValueToSeconds(key, value string) (int, bool) {
	switch key {
	case "out_time_ms", "out_time_us":
		raw, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil || raw < 0 {
			return 0, false
		}
		return int(math.Round(float64(raw) / 1_000_000.0)), true
	case "out_time":
		return parseFFmpegTimestamp(value)
	default:
		return 0, false
	}
}

func parseFFmpegTimestamp(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		return 0, false
	}
	hours, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || hours < 0 {
		return 0, false
	}
	minutes, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || minutes < 0 {
		return 0, false
	}
	secondsFloat, err := strconv.ParseFloat(parts[2], 64)
	if err != nil || secondsFloat < 0 {
		return 0, false
	}
	total := float64(hours*3600+minutes*60) + secondsFloat
	return int(math.Round(total)), true
}

func buildTranscodeProgress(processedSeconds, sourceDuration int) TranscodeProgress {
	if processedSeconds < 0 {
		processedSeconds = 0
	}
	remainingSeconds := 0
	progressPercent := 0.0
	if sourceDuration > 0 {
		if processedSeconds > sourceDuration {
			processedSeconds = sourceDuration
		}
		remainingSeconds = sourceDuration - processedSeconds
		progressPercent = (float64(processedSeconds) / float64(sourceDuration)) * 100
		if progressPercent < 0 {
			progressPercent = 0
		}
		if progressPercent > 100 {
			progressPercent = 100
		}
	}
	return TranscodeProgress{
		SourceDurationSeconds: sourceDuration,
		ProcessedSeconds:      processedSeconds,
		RemainingSeconds:      remainingSeconds,
		ProgressPercent:       math.Round(progressPercent*100) / 100,
	}
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
