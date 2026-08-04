package repository

import (
	"strings"
	"testing"
)

func TestBuildAdminForumPostListSQLKeepsReadModelContract(t *testing.T) {
	t.Parallel()

	countSQL, listSQL, listArgs := buildAdminForumPostListSQL(2, 20)
	for _, query := range []string{countSQL, listSQL} {
		normalized := strings.ToLower(query)
		for _, clause := range []string{
			"inspection_status = 'inspected'",
			"filter_decision = 'included'",
			"created_at >= now() - interval '30 days'",
		} {
			if !strings.Contains(normalized, clause) {
				t.Fatalf("query missing %q: %s", clause, query)
			}
		}
	}

	normalizedList := strings.ToLower(listSQL)
	for _, clause := range []string{
		"r.kind = 'attachment'",
		"r.kind = 'ed2k'",
		"array_agg(r.value order by r.position)",
		"order by p.observed_at desc, p.id desc",
		"limit $1 offset $2",
	} {
		if !strings.Contains(normalizedList, clause) {
			t.Fatalf("list query missing %q: %s", clause, listSQL)
		}
	}
	if len(listArgs) != 2 || listArgs[0] != 20 || listArgs[1] != 20 {
		t.Fatalf("list args=%#v, want [20 20]", listArgs)
	}
}
