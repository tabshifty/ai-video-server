package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

// RotationProbe distinguishes coded dimensions from the displayed orientation.
type RotationProbe struct {
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	Rotation    int     `json:"rotation"` // Counter-clockwise, following FFmpeg display matrices.
	Duration    float64 `json:"duration"`
	Codec       string  `json:"codec"`
	Format      string  `json:"format"`
	Bitrate     int     `json:"bitrate"`
	AudioTracks int     `json:"audio_tracks"`
	Lossless    bool    `json:"lossless"`
}

// DisplaySize returns dimensions after applying the display matrix.
func (p RotationProbe) DisplaySize() (int, int) {
	if p.Rotation%180 != 0 {
		return p.Height, p.Width
	}
	return p.Width, p.Height
}

// ProbeRotation rejects ambiguous stream layouts and non-quarter-turn matrices.
func ProbeRotation(ctx context.Context, path string) (RotationProbe, error) {
	raw, err := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-show_streams", "-show_format", "-of", "json", path).Output()
	if err != nil {
		return RotationProbe{}, fmt.Errorf("探测视频方向失败: %w", err)
	}
	var data struct {
		Streams []struct {
			Type        string `json:"codec_type"`
			Codec       string `json:"codec_name"`
			Width       int    `json:"width"`
			Height      int    `json:"height"`
			SAR         string `json:"sample_aspect_ratio"`
			Bitrate     string `json:"bit_rate"`
			Disposition struct {
				Attached int `json:"attached_pic"`
			} `json:"disposition"`
			Side []struct {
				Type     string  `json:"side_data_type"`
				Rotation float64 `json:"rotation"`
				Matrix   string  `json:"displaymatrix"`
			} `json:"side_data_list"`
		} `json:"streams"`
		Format struct {
			Name     string `json:"format_name"`
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		return RotationProbe{}, err
	}
	p := RotationProbe{Format: data.Format.Name, Lossless: true}
	p.Duration, _ = strconv.ParseFloat(data.Format.Duration, 64)
	videos := 0
	for _, s := range data.Streams {
		switch s.Type {
		case "video":
			if s.SAR != "" && s.SAR != "N/A" && s.SAR != "1:1" {
				return p, fmt.Errorf("暂不支持非方形像素视频")
			}
			if s.Disposition.Attached != 0 {
				return p, fmt.Errorf("暂不支持带内嵌封面的视频")
			}
			videos++
			p.Width = s.Width
			p.Height = s.Height
			p.Codec = s.Codec
			p.Bitrate, _ = strconv.Atoi(s.Bitrate)
			for _, side := range s.Side {
				if strings.Contains(strings.ToLower(side.Type), "dovi") {
					return p, fmt.Errorf("杜比视界源暂不支持方向修正")
				}
				if side.Type == "Display Matrix" {
					if math.Abs(side.Rotation-math.Round(side.Rotation/90)*90) > 0.01 {
						return p, fmt.Errorf("暂不支持非直角显示矩阵")
					}
					p.Rotation = ((int(math.Round(side.Rotation)) % 360) + 360) % 360
					// Reflections, shears and non-unit scaling cannot be represented by a rotation tag.
					if !pureRotationMatrix(side.Matrix) {
						return p, fmt.Errorf("暂不支持镜像或缩放显示矩阵")
					}
				}
			}
			if s.Codec != "h264" && s.Codec != "hevc" {
				p.Lossless = false
			}
		case "audio":
			p.AudioTracks++
			if s.Codec != "aac" && s.Codec != "mp3" && s.Codec != "ac3" && s.Codec != "eac3" && s.Codec != "alac" {
				p.Lossless = false
			}
		case "subtitle":
			if s.Codec != "mov_text" {
				return p, fmt.Errorf("暂不支持该内嵌字幕格式，请先转换为外置字幕")
			}
		default:
			return p, fmt.Errorf("暂不支持包含额外数据流的视频")
		}
	}
	if videos != 1 || p.Width <= 0 || p.Height <= 0 || p.Duration <= 0 {
		return p, fmt.Errorf("视频流或时长无效")
	}
	// Only advertise metadata rotation for MP4/MOV-compatible media.
	if !strings.Contains(p.Format, "mov") && !strings.Contains(p.Format, "mp4") {
		p.Lossless = false
	}
	return p, nil
}

func pureRotationMatrix(raw string) bool {
	var values []int64
	for _, line := range strings.Split(raw, "\n") {
		_, right, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		for _, s := range strings.Fields(right) {
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return false
			}
			values = append(values, v)
		}
	}
	if len(values) != 9 || values[2] != 0 || values[5] != 0 || values[6] != 0 || values[7] != 0 || values[8] != 1073741824 {
		return false
	}
	a, b, c, d := values[0], values[1], values[3], values[4]
	return a*a+b*b == 65536*65536 && c*c+d*d == 65536*65536 && a*d-b*c == 65536*65536
}

// RotateVideo writes a separate output; it never replaces the playable input.
// angle is clockwise relative to the current display orientation.
func RotateVideo(ctx context.Context, input, output string, angle int, mode string, p RotationProbe, progress func(TranscodeProgress)) error {
	if angle != 90 && angle != 180 && angle != 270 {
		return fmt.Errorf("旋转角度无效")
	}
	args := []string{"-v", "error", "-y", "-noautorotate"}
	if mode == "lossless" {
		if !p.Lossless {
			return fmt.Errorf("该视频不能使用无损旋转，请确认重新编码")
		}
		ccw := (p.Rotation - angle + 360) % 360
		args = append(args, "-display_rotation", strconv.Itoa(ccw), "-i", input, "-map", "0", "-c", "copy")
	} else if mode == "reencode" {
		cw := (angle - p.Rotation + 360) % 360
		filter := "null"
		switch cw {
		case 90:
			filter = "transpose=clock"
		case 180:
			filter = "hflip,vflip"
		case 270:
			filter = "transpose=cclock"
		}
		encoder, tag := "h264_videotoolbox", "avc1"
		if p.Codec == "hevc" {
			encoder, tag = "hevc_videotoolbox", "hvc1"
		}
		bitrate := p.Bitrate
		if bitrate < 500000 {
			bitrate = 4000000
		}
		args = append(args, "-display_rotation", "0", "-i", input, "-map", "0:V:0", "-map", "0:a?", "-map", "0:s?", "-vf", filter, "-c:v", encoder, "-allow_sw", "0", "-pix_fmt", "yuv420p", "-b:v", strconv.Itoa(bitrate), "-tag:v", tag, "-c:a", "copy", "-c:s", "copy", "-metadata:s:v:0", "rotate=0")
	} else {
		return fmt.Errorf("处理方式无效")
	}
	args = append(args, "-movflags", "+faststart", "-progress", "pipe:1", "-nostats", output)
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("启动方向修正失败: %w", err)
	}
	progressErr := readTranscodeProgress(stdout, int(math.Ceil(p.Duration)), progress)
	if err = cmd.Wait(); err != nil {
		return fmt.Errorf("方向修正失败: %w: %s", err, stderr.String())
	}
	return progressErr
}
