package repository

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestVideoTitleTextMigration(t *testing.T) {
	t.Parallel()

	up := readMigrationForTest(t, "0019_video_title_text.up.sql")
	down := readMigrationForTest(t, "0019_video_title_text.down.sql")

	assertSQLPattern(t, up, `(?is)alter\s+table\s+videos\s+alter\s+column\s+title\s+type\s+text`)
	assertSQLPattern(t, down, `(?is)alter\s+table\s+videos\s+alter\s+column\s+title\s+type\s+varchar\s*\(\s*200\s*\)\s+using\s+left\s*\(\s*title\s*,\s*200\s*\)`)
}

func readMigrationForTest(t *testing.T, name string) string {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
	if err != nil {
		t.Fatalf("read migration %s: %v", name, err)
	}
	return strings.TrimSpace(string(raw))
}

func assertSQLPattern(t *testing.T, sql, pattern string) {
	t.Helper()

	if !regexp.MustCompile(pattern).MatchString(sql) {
		t.Fatalf("migration SQL does not match %q:\n%s", pattern, sql)
	}
}
