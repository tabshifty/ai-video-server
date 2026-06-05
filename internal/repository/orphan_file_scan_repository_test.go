package repository

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestNormalizePotentialStoragePathFiltersNonLocalPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "trim local path", in: "  /storage/videos/sample.mp4  ", want: "/storage/videos/sample.mp4"},
		{name: "skip http url", in: "https://img.example/poster.jpg", want: ""},
		{name: "skip api path", in: "/api/v1/videos/123/source", want: ""},
		{name: "skip path without extension", in: "/storage/videos/sample", want: ""},
		{name: "skip empty string", in: "   ", want: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := normalizePotentialStoragePath(tt.in); got != tt.want {
				t.Fatalf("normalizePotentialStoragePath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCollectLocalStoragePathsRecursesNestedJSONValues(t *testing.T) {
	t.Parallel()

	var found []string
	add := func(value string) {
		if normalized := normalizePotentialStoragePath(value); normalized != "" {
			found = append(found, normalized)
		}
	}

	payload := map[string]any{
		"poster_path": "/storage/posters/cover.jpg",
		"nested": map[string]any{
			"items": []any{
				"  /storage/videos/sample.mp4  ",
				"https://img.example/poster.jpg",
				map[string]any{
					"avatar": "/storage/actors/1/avatar.png",
				},
			},
		},
		"ignored": 123,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	collectLocalStoragePaths(decoded, add)

	for _, want := range []string{
		"/storage/posters/cover.jpg",
		"/storage/videos/sample.mp4",
		"/storage/actors/1/avatar.png",
	} {
		if !slices.Contains(found, want) {
			t.Fatalf("expected collected paths to contain %q, got %v", want, found)
		}
	}
	if strings.Contains(strings.Join(found, ","), "https://img.example/poster.jpg") {
		t.Fatalf("expected HTTP URL to be skipped, got %v", found)
	}
}
