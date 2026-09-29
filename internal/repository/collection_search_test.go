package repository

import "testing"

func TestShortCollectionSearchPatternEscapesLikeWildcards(t *testing.T) {
	t.Parallel()

	if got, want := shortCollectionSearchPattern(`  A%_\\B  `), `%A\%\_\\\\B%`; got != want {
		t.Fatalf("pattern=%q want=%q", got, want)
	}
	if got := shortCollectionSearchPattern("   "); got != "" {
		t.Fatalf("blank pattern=%q want empty", got)
	}
}
