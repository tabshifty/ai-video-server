package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestVideoPlayURL(t *testing.T) {
	videoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	got := VideoPlayURL(videoID)
	want := "/api/v1/videos/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee/source"
	if got != want {
		t.Fatalf("unexpected play url: got=%s want=%s", got, want)
	}
}
