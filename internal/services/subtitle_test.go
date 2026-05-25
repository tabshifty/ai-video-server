package services

import (
	"strings"
	"testing"

	"video-server/pkg/ffmpeg"
)

func TestSubtitleUploadPlanKeepsAssSource(t *testing.T) {
	for _, filename := range []string{"episode.ass", "episode.ssa"} {
		t.Run(filename, func(t *testing.T) {
			plan, err := subtitleUploadPlanForFilename(filename)
			if err != nil {
				t.Fatalf("subtitleUploadPlanForFilename() error = %v", err)
			}
			if plan.StoredFormat != "ass" {
				t.Fatalf("StoredFormat = %q, want ass", plan.StoredFormat)
			}
			if plan.StoredMIMEType != "text/x-ssa" {
				t.Fatalf("StoredMIMEType = %q, want text/x-ssa", plan.StoredMIMEType)
			}
			if plan.NeedsWebVTT {
				t.Fatalf("NeedsWebVTT = true, want false")
			}
		})
	}
}

func TestSanitizeAssContentRemovesUnsafeFontPathAndClampsFade(t *testing.T) {
	raw := `Dialogue: 0,0:00:01.00,0:00:02.00,Default,,0,0,0,,{\fn/Users/me/Fonts/Evil.ttf\fad(-20,999999)}测试`

	got := sanitizeAssContent(raw)

	if strings.Contains(got, "/Users/me/Fonts") || strings.Contains(got, "Evil.ttf") {
		t.Fatalf("sanitizeAssContent kept unsafe font path: %q", got)
	}
	if !strings.Contains(got, `\fad(0,60000)`) {
		t.Fatalf("sanitizeAssContent did not clamp fad params: %q", got)
	}
}

func TestEmbeddedSubtitleOutputPlanKeepsAssSource(t *testing.T) {
	plan := embeddedSubtitleOutputPlanForProbe(ffmpeg.SubtitleProbe{Index: 3, Codec: "ass"})

	if plan.StoredFormat != "ass" {
		t.Fatalf("StoredFormat = %q, want ass", plan.StoredFormat)
	}
	if plan.MIMEType != "text/x-ssa" {
		t.Fatalf("MIMEType = %q, want text/x-ssa", plan.MIMEType)
	}
	if plan.Extension != "ass" {
		t.Fatalf("Extension = %q, want ass", plan.Extension)
	}
	if plan.ConvertToWebVTT {
		t.Fatalf("ConvertToWebVTT = true, want false")
	}
}
