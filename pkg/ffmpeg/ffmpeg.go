package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var ErrWebPEncodingUnavailable = errors.New("webp encoding unavailable")

// VideoProbe holds basic media attributes from ffprobe.
type VideoProbe struct {
	Duration        float64 `json:"duration"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	Codec           string  `json:"codec"`
	BitrateKbps     int     `json:"bitrate_kbps"`
	AudioCodec      string  `json:"audio_codec"`
	AudioChannels   int     `json:"audio_channels"`
	AudioTrackCount int     `json:"audio_track_count"`
}

type SubtitleProbe struct {
	Index     int
	Codec     string
	Language  string
	Title     string
	IsDefault bool
}

// TranscodeOptions controls H.265 transcoding behavior.
type TranscodeOptions struct {
	CRF              string
	VideoBitrateKbps int
	SourceDuration   int
	SpatialAQ        bool
	ProgressHandler  func(TranscodeProgress)
}

type TranscodeProfile string

const (
	TranscodeProfileHEVCPrimary TranscodeProfile = "hevc_primary"
	TranscodeProfileAVCCompat   TranscodeProfile = "avc_compat"
)

// TranscodeProgress represents ffmpeg realtime progress.
type TranscodeProgress struct {
	SourceDurationSeconds int     `json:"source_duration_seconds"`
	ProcessedSeconds      int     `json:"processed_seconds"`
	RemainingSeconds      int     `json:"remaining_seconds"`
	ProgressPercent       float64 `json:"progress_percent"`
}

func preferredHardwareHevcEncoder() string { return "hevc_videotoolbox" }

func preferredHardwareAvcEncoder() string { return "h264_videotoolbox" }

func buildTranscodeVideoArgs(inputPath, outputPath string, profile TranscodeProfile, options TranscodeOptions) []string {
	encoder := preferredHardwareHevcEncoder()
	switch profile {
	case TranscodeProfileAVCCompat:
		encoder = preferredHardwareAvcEncoder()
	case TranscodeProfileHEVCPrimary:
	default:
		panic(fmt.Sprintf("unsupported transcode profile: %s", profile))
	}
	args := []string{
		"-y",
		"-i", inputPath,
		"-map", "0:v:0",
		"-map", "0:a?",
		"-c:v", encoder,
		"-pix_fmt", "yuv420p",
		"-allow_sw", "0",
	}
	switch profile {
	case TranscodeProfileHEVCPrimary:
		args = append(args, "-tag:v", "hvc1")
		if options.SpatialAQ {
			args = append(args, "-spatial_aq", "1")
		}
	}
	if options.VideoBitrateKbps > 0 {
		bitrate := strconv.Itoa(options.VideoBitrateKbps) + "k"
		args = append(args,
			"-b:v", bitrate,
			"-maxrate", strconv.Itoa(options.VideoBitrateKbps*2)+"k",
			"-bufsize", strconv.Itoa(options.VideoBitrateKbps*4)+"k",
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
		"-movflags", "+faststart",
		"-progress", "pipe:1",
		"-nostats",
		outputPath,
	)
	return args
}

func TranscodeVideo(ctx context.Context, inputPath, outputPath string, profile TranscodeProfile, options TranscodeOptions) error {
	args := buildTranscodeVideoArgs(inputPath, outputPath, profile, options)
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
		output := stderr.String()
		encoder := preferredHardwareHevcEncoder()
		if profile == TranscodeProfileAVCCompat {
			encoder = preferredHardwareAvcEncoder()
		}
		if isEncoderUnavailableOutput(output, encoder) {
			return fmt.Errorf("hardware encoder unavailable for %s: %s", profile, encoder)
		}
		return fmt.Errorf("ffmpeg transcode failed: %w, output=%s", waitErr, output)
	}
	return nil
}

// TranscodeHEVC transcodes source into H.265/AAC output using videotoolbox.
func TranscodeHEVC(ctx context.Context, inputPath, outputPath string, options TranscodeOptions) error {
	return TranscodeVideo(ctx, inputPath, outputPath, TranscodeProfileHEVCPrimary, options)
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
		"-show_entries", "stream=codec_type,width,height,codec_name,bit_rate,channels:format=duration,bit_rate",
		"-of", "json",
		inputPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return VideoProbe{}, fmt.Errorf("ffprobe failed: %w", err)
	}
	return parseProbeOutput(out)
}

func parseProbeOutput(raw []byte) (VideoProbe, error) {
	var payload struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
			CodecName string `json:"codec_name"`
			BitRate   string `json:"bit_rate"`
			Channels  int    `json:"channels"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return VideoProbe{}, fmt.Errorf("unmarshal ffprobe output: %w", err)
	}

	probe := VideoProbe{}
	for _, stream := range payload.Streams {
		switch strings.TrimSpace(stream.CodecType) {
		case "video":
			if probe.Width == 0 {
				probe.Width = stream.Width
			}
			if probe.Height == 0 {
				probe.Height = stream.Height
			}
			if probe.Codec == "" {
				probe.Codec = strings.TrimSpace(stream.CodecName)
			}
			if probe.BitrateKbps == 0 {
				probe.BitrateKbps = parseBitrateKbps(stream.BitRate, payload.Format.BitRate)
			}
		case "audio":
			probe.AudioTrackCount++
			if probe.AudioCodec == "" {
				probe.AudioCodec = strings.TrimSpace(stream.CodecName)
			}
			if probe.AudioChannels == 0 {
				probe.AudioChannels = stream.Channels
			}
		}
	}
	if payload.Format.Duration != "" {
		var duration float64
		if _, err := fmt.Sscanf(payload.Format.Duration, "%f", &duration); err == nil {
			probe.Duration = duration
		}
	}
	if probe.BitrateKbps == 0 {
		probe.BitrateKbps = parseBitrateKbps("", payload.Format.BitRate)
	}
	return probe, nil
}

func ProbeSubtitles(ctx context.Context, inputPath string) ([]SubtitleProbe, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "stream=index,codec_name,codec_type:stream_tags=language,title:stream_disposition=default",
		"-of", "json",
		inputPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe subtitle probe failed: %w", err)
	}
	return parseSubtitleProbeOutput(out)
}

func parseSubtitleProbeOutput(raw []byte) ([]SubtitleProbe, error) {
	var payload struct {
		Streams []struct {
			Index     int    `json:"index"`
			CodecName string `json:"codec_name"`
			CodecType string `json:"codec_type"`
			Tags      struct {
				Language string `json:"language"`
				Title    string `json:"title"`
			} `json:"tags"`
			Disposition struct {
				Default int `json:"default"`
			} `json:"disposition"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal subtitle probe output: %w", err)
	}
	tracks := make([]SubtitleProbe, 0, len(payload.Streams))
	for _, stream := range payload.Streams {
		if strings.TrimSpace(stream.CodecType) != "subtitle" {
			continue
		}
		tracks = append(tracks, SubtitleProbe{
			Index:     stream.Index,
			Codec:     strings.TrimSpace(stream.CodecName),
			Language:  strings.TrimSpace(stream.Tags.Language),
			Title:     strings.TrimSpace(stream.Tags.Title),
			IsDefault: stream.Disposition.Default == 1,
		})
	}
	return tracks, nil
}

func ExtractSubtitleToWebVTT(ctx context.Context, inputPath string, streamIndex int, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create subtitle output dir: %w", err)
	}
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-i", inputPath,
		"-map", fmt.Sprintf("0:%d", streamIndex),
		"-vn",
		"-an",
		"-c:s", "webvtt",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg extract subtitle failed: %w, output=%s", err, string(out))
	}
	return nil
}

func buildExtractSubtitleToAssArgs(inputPath string, streamIndex int, outputPath string) []string {
	return []string{
		"-y",
		"-i", inputPath,
		"-map", fmt.Sprintf("0:%d", streamIndex),
		"-vn",
		"-an",
		"-c:s", "copy",
		outputPath,
	}
}

func ExtractSubtitleToAss(ctx context.Context, inputPath string, streamIndex int, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create subtitle output dir: %w", err)
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", buildExtractSubtitleToAssArgs(inputPath, streamIndex, outputPath)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg extract ass subtitle failed: %w, output=%s", err, string(out))
	}
	return nil
}

func buildConvertSubtitleToWebVTTArgs(inputPath, outputPath string) []string {
	return []string{
		"-y",
		"-i", inputPath,
		"-vn",
		"-an",
		"-c:s", "webvtt",
		"-f", "webvtt",
		outputPath,
	}
}

func ConvertSubtitleToWebVTT(ctx context.Context, inputPath, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create subtitle output dir: %w", err)
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", buildConvertSubtitleToWebVTTArgs(inputPath, outputPath)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg convert subtitle failed: %w, output=%s", err, string(out))
	}
	return nil
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
	primaryArgs := append(
		[]string{"-y", "-i", inputPath},
		append(webpEncodeArgs(true, quality), outputPath)...,
	)
	if out, err := runFFmpegCombinedOutput(ctx, primaryArgs...); err != nil {
		primaryOutput := string(out)
		if !isEncoderUnavailableOutput(primaryOutput, "libwebp") {
			return fmt.Errorf("ffmpeg convert webp failed: %w, output=%s", err, primaryOutput)
		}
		fallbackArgs := append(
			[]string{"-y", "-i", inputPath},
			append(webpEncodeArgs(false, quality), outputPath)...,
		)
		fallbackOut, fallbackErr := runFFmpegCombinedOutput(ctx, fallbackArgs...)
		if fallbackErr == nil {
			return nil
		}
		fallbackOutput := string(fallbackOut)
		if !isEncoderUnavailableOutput(fallbackOutput, "webp") {
			return fmt.Errorf(
				"ffmpeg convert webp failed: %w, output=%s, fallback_output=%s",
				fallbackErr,
				primaryOutput,
				fallbackOutput,
			)
		}
		cwebpOut, cwebpErr := runCWebPEncode(ctx, inputPath, outputPath, quality)
		if cwebpErr != nil {
			return fmt.Errorf(
				"%w: ffmpeg convert webp failed: %v, output=%s, fallback_output=%s, cwebp_output=%s",
				ErrWebPEncodingUnavailable,
				cwebpErr,
				primaryOutput,
				fallbackOutput,
				string(cwebpOut),
			)
		}
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
		primaryArgs := append(append([]string{}, args...), webpEncodeArgs(true, quality)...)
		primaryArgs = append(primaryArgs, outputPath)
		primaryOut, primaryErr := runFFmpegCombinedOutput(ctx, primaryArgs...)
		if primaryErr == nil {
			return nil
		}
		primaryOutput := string(primaryOut)
		if !isEncoderUnavailableOutput(primaryOutput, "libwebp") {
			return fmt.Errorf("ffmpeg resize image failed: %w, output=%s", primaryErr, string(primaryOut))
		}
		fallbackArgs := append(append([]string{}, args...), webpEncodeArgs(false, quality)...)
		fallbackArgs = append(fallbackArgs, outputPath)
		fallbackOut, fallbackErr := runFFmpegCombinedOutput(ctx, fallbackArgs...)
		if fallbackErr == nil {
			return nil
		}
		fallbackOutput := string(fallbackOut)
		if !isEncoderUnavailableOutput(fallbackOutput, "webp") {
			return fmt.Errorf(
				"ffmpeg resize image failed: %w, output=%s, fallback_output=%s",
				fallbackErr,
				primaryOutput,
				fallbackOutput,
			)
		}
		if cwebpErr := resizeImageToWebPViaCWebP(ctx, inputPath, outputPath, scaleExpr, quality); cwebpErr != nil {
			return fmt.Errorf(
				"ffmpeg resize image failed: %w, output=%s, fallback_output=%s, cwebp_error=%v",
				fallbackErr,
				primaryOutput,
				fallbackOutput,
				cwebpErr,
			)
		}
		return nil
	}
	args = append(args, outputPath)
	if out, err := runFFmpegCombinedOutput(ctx, args...); err != nil {
		return fmt.Errorf("ffmpeg resize image failed: %w, output=%s", err, string(out))
	}
	return nil
}

func runFFmpegCombinedOutput(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	return cmd.CombinedOutput()
}

func runCWebPEncode(ctx context.Context, inputPath, outputPath string, quality int) ([]byte, error) {
	cwebpPath, err := exec.LookPath("cwebp")
	if err != nil {
		return nil, fmt.Errorf("cwebp not found: %w", err)
	}
	args := []string{
		"-q", strconv.Itoa(quality),
		inputPath,
		"-o", outputPath,
	}
	cmd := exec.CommandContext(ctx, cwebpPath, args...)
	return cmd.CombinedOutput()
}

func resizeImageToWebPViaCWebP(ctx context.Context, inputPath, outputPath, scaleExpr string, quality int) error {
	tmpFile, err := os.CreateTemp(filepath.Dir(outputPath), ".resize-webp-*.png")
	if err != nil {
		return fmt.Errorf("create resize temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	if closeErr := tmpFile.Close(); closeErr != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close resize temp file: %w", closeErr)
	}
	defer os.Remove(tmpPath)

	ffmpegArgs := []string{
		"-y",
		"-i", inputPath,
		"-vf", scaleExpr,
		"-frames:v", "1",
		tmpPath,
	}
	ffmpegOut, ffmpegErr := runFFmpegCombinedOutput(ctx, ffmpegArgs...)
	if ffmpegErr != nil {
		return fmt.Errorf("prepare resize frame failed: %w, output=%s", ffmpegErr, string(ffmpegOut))
	}
	cwebpOut, cwebpErr := runCWebPEncode(ctx, tmpPath, outputPath, quality)
	if cwebpErr != nil {
		return fmt.Errorf("cwebp encode failed: %w, output=%s", cwebpErr, string(cwebpOut))
	}
	return nil
}

func webpEncodeArgs(preferLibWebP bool, quality int) []string {
	if preferLibWebP {
		return []string{
			"-c:v", "libwebp",
			"-q:v", strconv.Itoa(quality),
			"-compression_level", "6",
			"-preset", "picture",
		}
	}
	return []string{
		"-c:v", "webp",
		"-q:v", strconv.Itoa(quality),
	}
}

func isEncoderUnavailableOutput(output, encoder string) bool {
	normalized := strings.ToLower(output)
	encoder = strings.ToLower(strings.TrimSpace(encoder))
	if encoder == "" {
		return false
	}
	return strings.Contains(normalized, fmt.Sprintf("unknown encoder '%s'", encoder)) ||
		strings.Contains(normalized, fmt.Sprintf(`unknown encoder "%s"`, encoder)) ||
		(strings.Contains(normalized, "encoder not found") && strings.Contains(normalized, encoder))
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
