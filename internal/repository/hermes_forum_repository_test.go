package repository

import (
	"regexp"
	"strings"
	"testing"
)

func TestBuildAdminForumPostListSQLKeepsReadModelContract(t *testing.T) {
	t.Parallel()

	countSQL, listSQL, countArgs, listArgs := buildAdminForumPostListSQL(2, 20, "  资源%_\\  ")
	for _, query := range []string{countSQL, listSQL} {
		normalized := strings.ToLower(query)
		for _, clause := range []string{
			"inspection_status in ('inspected', 'restricted')",
			"filter_decision = 'included'",
			"p.inspection_status = 'restricted'",
			"p.created_at >= now() - interval '30 days'",
			"or exists",
			"collected_forum_post_resources",
			"rp.post_id = p.id",
			"rp.kind in ('attachment', 'ed2k')",
			`lower(p.title) like $1 escape '\'`,
		} {
			if !strings.Contains(normalized, clause) {
				t.Fatalf("query missing %q: %s", clause, query)
			}
		}
	}

	normalizedList := strings.ToLower(listSQL)
	for _, clause := range []string{
		"p.inspection_status",
		"r.kind = 'attachment'",
		"r.kind = 'ed2k'",
		"array_agg(r.value order by r.position)",
		"order by p.observed_at desc, p.id desc",
		"limit $2 offset $3",
	} {
		if !strings.Contains(normalizedList, clause) {
			t.Fatalf("list query missing %q: %s", clause, listSQL)
		}
	}
	if len(countArgs) != 1 || countArgs[0] != `%资源\%\_\\%` {
		t.Fatalf("count args=%#v, want escaped literal title pattern", countArgs)
	}
	if len(listArgs) != 3 || listArgs[0] != `%资源\%\_\\%` || listArgs[1] != 20 || listArgs[2] != 20 {
		t.Fatalf("list args=%#v, want [escaped-pattern 20 20]", listArgs)
	}
}

func TestBuildAdminForumPostListSQLWithoutSearchKeepsPaginationArguments(t *testing.T) {
	t.Parallel()

	countSQL, listSQL, countArgs, listArgs := buildAdminForumPostListSQL(1, 20, "   ")
	if strings.Contains(strings.ToLower(countSQL+listSQL), "lower(p.title) like") {
		t.Fatalf("empty search unexpectedly added title predicate: %s", listSQL)
	}
	if len(countArgs) != 0 {
		t.Fatalf("count args=%#v, want none", countArgs)
	}
	if len(listArgs) != 2 || listArgs[0] != 20 || listArgs[1] != 0 {
		t.Fatalf("list args=%#v, want [20 0]", listArgs)
	}
}

func TestDiscoverForumPostsCleanupSkipsPostsWithResources(t *testing.T) {
	t.Parallel()

	cleanupPattern := regexp.MustCompile(`(?is)delete\s+from\s+collected_forum_posts\s+p\s+where\s+p\.created_at\s+<\s+now\(\)\s+-\s+interval\s+'30 days'\s+and\s+not\s+exists\s*\(\s*select\s+1\s+from\s+collected_forum_post_resources\s+r\s+where\s+r\.post_id\s*=\s*p\.id\s+and\s+r\.kind\s+in\s*\(\s*'attachment'\s*,\s*'ed2k'\s*\)`)
	if !cleanupPattern.MatchString(cleanupExpiredForumPostsSQL) {
		t.Fatalf("expired forum cleanup must skip posts that have attachment or ED2K resources")
	}
	normalized := strings.ToLower(cleanupExpiredForumPostsSQL)
	if !strings.Contains(normalized, "p.inspection_status not in ('pending', 'restricted')") {
		t.Fatalf("expired forum cleanup must preserve pending and restricted posts: %s", cleanupExpiredForumPostsSQL)
	}
}
