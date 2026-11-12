package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRotationRoundTrip(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg unavailable")
	}
	ctx := context.Background()
	dir := t.TempDir()
	src := filepath.Join(dir, "source.mp4")
	if out, err := exec.CommandContext(ctx, "ffmpeg", "-v", "error", "-f", "lavfi", "-i", "testsrc2=size=160x96:rate=10", "-f", "lavfi", "-i", "sine=frequency=440", "-t", "1", "-c:v", "libx264", "-c:a", "aac", "-shortest", src).CombinedOutput(); err != nil {
		t.Fatalf("sample: %v %s", err, out)
	}
	original, err := ProbeRotation(ctx, src)
	if err != nil {
		t.Fatal(err)
	}
	for _, angle := range []int{90, 180, 270} {
		t.Run(fmt.Sprintf("%d", angle), func(t *testing.T) {
			dst := filepath.Join(dir, string(rune('a'+angle/90))+".mp4")
			if err := RotateVideo(ctx, src, dst, angle, "lossless", original, nil); err != nil {
				t.Fatal(err)
			}
			got, err := ProbeRotation(ctx, dst)
			if err != nil {
				t.Fatal(err)
			}
			w, h := got.DisplaySize()
			ew, eh := 160, 96
			if angle != 180 {
				ew, eh = eh, ew
			}
			if w != ew || h != eh {
				t.Fatalf("display %dx%d, rotation %v", w, h, got.Rotation)
			}
			hash := func(path string) string {
				out, err := exec.Command("ffmpeg", "-v", "error", "-i", path, "-map", "0:v:0", "-c", "copy", "-f", "hash", "-").Output()
				if err != nil {
					t.Fatal(err)
				}
				return string(out)
			}
			if hash(src) != hash(dst) {
				t.Fatal("lossless media packets changed")
			}
			filters := map[int]string{90: "transpose=clock", 180: "hflip,vflip", 270: "transpose=cclock"}
			frameHash := func(path, filter string) string {
				args := []string{"-v", "error", "-i", path, "-map", "0:v:0", "-an"}
				if filter != "" {
					args = append(args, "-vf", filter)
				}
				args = append(args, "-f", "framemd5", "-")
				out, e := exec.Command("ffmpeg", args...).Output()
				if e != nil {
					t.Fatal(e)
				}
				return string(out)
			}
			if frameHash(src, filters[angle]) != frameHash(dst, "") {
				t.Fatal("display rotation differs from requested clockwise direction")
			}

			restored := filepath.Join(dir, string(rune('a'+angle/90))+"-restored.mp4")
			if err := RotateVideo(ctx, dst, restored, 360-angle, "lossless", got, nil); err != nil {
				t.Fatal(err)
			}
			final, err := ProbeRotation(ctx, restored)
			if err != nil {
				t.Fatal(err)
			}
			if final.Rotation != 0 {
				t.Fatalf("reverse rotation = %v", final.Rotation)
			}
		})
	}
}

func TestRotationReencode(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("VideoToolbox requires macOS")
	}
	dir := t.TempDir()
	ctx := context.Background()
	src := filepath.Join(dir, "source.mp4")
	if out, err := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "testsrc2=size=160x96:rate=10", "-t", "1", "-c:v", "libx264", src).CombinedOutput(); err != nil {
		t.Fatalf("sample %v %s", err, out)
	}
	probe, err := ProbeRotation(ctx, src)
	if err != nil {
		t.Fatal(err)
	}
	tagged := filepath.Join(dir, "tagged.mp4")
	if err = RotateVideo(ctx, src, tagged, 90, "lossless", probe, nil); err != nil {
		t.Fatal(err)
	}
	p, err := ProbeRotation(ctx, tagged)
	if err != nil {
		t.Fatal(err)
	}
	for _, angle := range []int{90, 180, 270} {
		dst := filepath.Join(dir, fmt.Sprintf("encoded-%d.mp4", angle))
		if err = RotateVideo(ctx, tagged, dst, angle, "reencode", p, nil); err != nil {
			t.Fatal(err)
		}
		out, err := ProbeRotation(ctx, dst)
		if err != nil {
			t.Fatal(err)
		}
		ew, eh := p.DisplaySize()
		if angle%180 != 0 {
			ew, eh = eh, ew
		}
		if out.Width != ew || out.Height != eh || out.Rotation != 0 {
			t.Fatalf("angle %d got %+v", angle, out)
		}
	}
}
