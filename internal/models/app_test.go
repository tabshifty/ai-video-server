package models

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestVideoListItemJSONIncludesMetadata(t *testing.T) {
	t.Parallel()

	item := VideoListItem{
		ID:            uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Title:         "SSIS-001",
		Type:          "av",
		ThumbnailPath: "/api/v1/videos/11111111-1111-1111-1111-111111111111/thumbnail",
		Metadata:      json.RawMessage(`{"poster_url":"https://img.example/poster.jpg","poster_decision":"primary_selected"}`),
	}

	raw, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal VideoListItem: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal VideoListItem json: %v", err)
	}

	metadata, ok := decoded["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("expected metadata object, got=%T", decoded["metadata"])
	}
	if metadata["poster_url"] != "https://img.example/poster.jpg" {
		t.Fatalf("expected poster_url in metadata, got=%v", metadata["poster_url"])
	}
	if metadata["poster_decision"] != "primary_selected" {
		t.Fatalf("expected poster_decision in metadata, got=%v", metadata["poster_decision"])
	}
}
