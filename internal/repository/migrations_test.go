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

func TestIPTVPlaylistMigration(t *testing.T) {
	t.Parallel()

	up := readMigrationForTest(t, "0020_iptv_playlist.up.sql")
	down := readMigrationForTest(t, "0020_iptv_playlist.down.sql")

	assertSQLPattern(t, up, `(?is)create\s+table\s+if\s+not\s+exists\s+iptv_playlists`)
	assertSQLPattern(t, up, `(?is)constraint\s+iptv_playlists_singleton\s+check\s*\(\s*id\s*=\s*1\s*\)`)
	assertSQLPattern(t, up, `(?is)create\s+table\s+if\s+not\s+exists\s+iptv_channels`)
	assertSQLPattern(t, up, `(?is)sort_order\s+int\s+not\s+null`)
	assertSQLPattern(t, up, `(?is)constraint\s+iptv_channels_http_url\s+check`)
	assertSQLPattern(t, down, `(?is)drop\s+table\s+if\s+exists\s+iptv_channels`)
	assertSQLPattern(t, down, `(?is)drop\s+table\s+if\s+exists\s+iptv_playlists`)
}

func TestWesternAVOshashMigration(t *testing.T) {
	t.Parallel()

	up := readMigrationForTest(t, "0021_western_av_oshash_gate.up.sql")
	down := readMigrationForTest(t, "0021_western_av_oshash_gate.down.sql")

	assertSQLPattern(t, up, `(?is)alter\s+table\s+videos\s+add\s+column\s+os_hash\s+char\(16\)`)
	assertSQLPattern(t, up, `(?is)create\s+index\s+if\s+not\s+exists\s+idx_videos_os_hash`)
	assertSQLPattern(t, up, `(?is)check\s*\(\s*status\s+in\s*\(`)
	assertSQLPattern(t, up, `(?is)'av_scrape_pending'`)

	assertSQLPattern(t, down, `(?is)update\s+videos\s+set\s+status\s*=\s*'uploaded'\s+where\s+status\s*=\s*'av_scrape_pending'`)
	assertSQLPattern(t, down, `(?is)drop\s+index\s+if\s+exists\s+idx_videos_os_hash`)
	assertSQLPattern(t, down, `(?is)alter\s+table\s+videos\s+drop\s+column\s+if\s+exists\s+os_hash`)
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
