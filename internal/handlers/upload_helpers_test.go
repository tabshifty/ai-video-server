package handlers

import (
	"reflect"
	"testing"
)

func TestParseUploadTags_JSON(t *testing.T) {
	got := parseUploadTags(`["搞笑", "舞蹈", "  "]`)
	want := []string{"搞笑", "舞蹈"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected tags: got=%v want=%v", got, want)
	}
}

func TestParseUploadTags_CSVFallback(t *testing.T) {
	got := parseUploadTags("funny, dance, ,sport")
	want := []string{"funny", "dance", "sport"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected tags: got=%v want=%v", got, want)
	}
}

func TestIsSHA256Hex(t *testing.T) {
	valid := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if !isSHA256Hex(valid) {
		t.Fatalf("expected valid hash")
	}
	if isSHA256Hex("abc123") {
		t.Fatalf("expected invalid short hash")
	}
}

func TestIsAllowedVideoExt(t *testing.T) {
	if !isAllowedVideoExt("movie.MP4") {
		t.Fatalf("expected mp4 allowed")
	}
	if isAllowedVideoExt("archive.zip") {
		t.Fatalf("expected zip disallowed")
	}
}
