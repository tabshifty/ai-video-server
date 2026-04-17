package repository

import (
	"testing"

	"github.com/google/uuid"
)

func TestNormalizeCollectionName(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"":                "",
		"   ":             "",
		" 热 门 合 集 ":       "热 门 合 集",
		"  Mix\t\nName  ": "mix name",
		"  A   B   C  ":   "a b c",
		"\n\t中文  合集  名  ": "中文 合集 名",
	}

	for raw, want := range cases {
		got := normalizeCollectionName(raw)
		if got != want {
			t.Fatalf("normalizeCollectionName(%q)=%q want=%q", raw, got, want)
		}
	}
}

func TestDedupeCollectionIDs(t *testing.T) {
	t.Parallel()

	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()

	got := dedupeCollectionIDs([]uuid.UUID{id1, id2, id1, id3, id2, id3})
	if len(got) != 3 {
		t.Fatalf("len=%d want=3", len(got))
	}
	if got[0] != id1 || got[1] != id2 || got[2] != id3 {
		t.Fatalf("unexpected dedupe order: %v", got)
	}
}
