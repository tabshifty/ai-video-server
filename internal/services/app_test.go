package services

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeShortCollectionSearchQuery(t *testing.T) {
	t.Parallel()

	query, err := normalizeShortCollectionSearchQuery("  海边 合集  ")
	if err != nil || query != "海边 合集" {
		t.Fatalf("query=%q err=%v, want trimmed query", query, err)
	}

	query, err = normalizeShortCollectionSearchQuery("   ")
	if err != nil || query != "" {
		t.Fatalf("blank query=%q err=%v, want empty query", query, err)
	}
}

func TestNormalizeShortCollectionSearchQueryRejectsMoreThan200Runes(t *testing.T) {
	t.Parallel()

	_, err := normalizeShortCollectionSearchQuery(strings.Repeat("合", 201))
	if !errors.Is(err, ErrShortCollectionSearchQueryTooLong) {
		t.Fatalf("error=%v, want ErrShortCollectionSearchQueryTooLong", err)
	}
}
