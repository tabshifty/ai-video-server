package repository

import (
	"testing"
)

func TestNormalizeVideoListItemMetadataDefaultsToJSONObject(t *testing.T) {
	t.Parallel()

	got := normalizeVideoListItemMetadata(nil)
	if string(got) != "{}" {
		t.Fatalf("expected empty metadata object, got=%s", string(got))
	}
}

func TestNormalizeVideoListItemMetadataPreservesScrapedPosterFields(t *testing.T) {
	t.Parallel()

	got := normalizeVideoListItemMetadata([]byte(`{"poster_url":"https://img.example/poster.jpg","poster_source":"video_cover","poster_quality":"primary"}`))
	if string(got) != `{"poster_url":"https://img.example/poster.jpg","poster_source":"video_cover","poster_quality":"primary"}` {
		t.Fatalf("expected poster metadata preserved, got=%s", string(got))
	}
}
