package services

import (
	"errors"
	"testing"

	"video-server/pkg/ffmpeg"
)

func TestResolveProbeFieldsWithProbeErrorUsesDefaults(t *testing.T) {
	duration, width, height, metadata := resolveProbeFields(ffmpeg.VideoProbe{}, errors.New("ffprobe failed"), "23")

	if duration != 0 || width != 0 || height != 0 {
		t.Fatalf("expected zero defaults, got duration=%d width=%d height=%d", duration, width, height)
	}
	if metadata["codec"] != "unknown" {
		t.Fatalf("expected codec unknown, got %v", metadata["codec"])
	}
	if metadata["crf"] != "23" {
		t.Fatalf("expected crf=23, got %v", metadata["crf"])
	}
	if _, ok := metadata["probe_error"]; !ok {
		t.Fatalf("expected probe_error metadata")
	}
}

func TestResolveProbeFieldsWithValidProbeParsesValues(t *testing.T) {
	duration, width, height, metadata := resolveProbeFields(ffmpeg.VideoProbe{
		Duration: 12.6,
		Width:    1920,
		Height:   1080,
		Codec:    "h265",
	}, nil, "21")

	if duration != 13 || width != 1920 || height != 1080 {
		t.Fatalf("unexpected parsed values duration=%d width=%d height=%d", duration, width, height)
	}
	if metadata["codec"] != "h265" {
		t.Fatalf("expected codec h265, got %v", metadata["codec"])
	}
	if metadata["crf"] != "21" {
		t.Fatalf("expected crf=21, got %v", metadata["crf"])
	}
	if _, ok := metadata["probe_error"]; ok {
		t.Fatalf("did not expect probe_error metadata")
	}
}
