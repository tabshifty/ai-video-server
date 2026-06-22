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

func TestOrphanFileScanMigration(t *testing.T) {
	t.Parallel()

	up := readMigrationForTest(t, "0022_orphan_file_scan.up.sql")
	down := readMigrationForTest(t, "0022_orphan_file_scan.down.sql")

	assertSQLPattern(t, up, `(?is)create\s+table\s+if\s+not\s+exists\s+orphan_file_scans`)
	assertSQLPattern(t, up, `(?is)check\s*\(\s*id\s*=\s*1\s*\)`)
	assertSQLPattern(t, up, `(?is)status\s+varchar\(20\)\s+not\s+null\s+default\s+'idle'`)
	assertSQLPattern(t, up, `(?is)'pending'.*'running'.*'completed'.*'failed'.*'deleted'`)
	assertSQLPattern(t, up, `(?is)create\s+table\s+if\s+not\s+exists\s+orphan_file_scan_items`)
	assertSQLPattern(t, up, `(?is)scan_id\s+bigint\s+not\s+null\s+references\s+orphan_file_scans\(id\)\s+on\s+delete\s+cascade`)
	assertSQLPattern(t, down, `(?is)drop\s+table\s+if\s+exists\s+orphan_file_scan_items`)
	assertSQLPattern(t, down, `(?is)drop\s+table\s+if\s+exists\s+orphan_file_scans`)
}

func TestTVApkDistributionMigration(t *testing.T) {
	t.Parallel()

	up := readMigrationForTest(t, "0023_tv_apk_distribution.up.sql")
	down := readMigrationForTest(t, "0023_tv_apk_distribution.down.sql")

	assertSQLPattern(t, up, `(?is)create\s+table\s+if\s+not\s+exists\s+tv_app_releases`)
	assertSQLPattern(t, up, `(?is)publish_status\s+varchar\(32\)\s+not\s+null\s+default\s+'draft'`)
	assertSQLPattern(t, up, `(?is)'draft'.*'published_complete'.*'published_missing_abi'.*'offline'`)
	assertSQLPattern(t, up, `(?is)unique\s*\(\s*package_name\s*,\s*version_code\s*\)`)
	assertSQLPattern(t, up, `(?is)create\s+table\s+if\s+not\s+exists\s+tv_app_release_apks`)
	assertSQLPattern(t, up, `(?is)references\s+tv_app_releases\(id\)\s+on\s+delete\s+cascade`)
	assertSQLPattern(t, up, `(?is)abi\s+in\s+\('armeabi-v7a',\s*'arm64-v8a'\)`)
	assertSQLPattern(t, up, `(?is)unique\s*\(\s*release_id\s*,\s*abi\s*\)`)

	assertSQLPattern(t, down, `(?is)drop\s+table\s+if\s+exists\s+tv_app_release_apks`)
	assertSQLPattern(t, down, `(?is)drop\s+table\s+if\s+exists\s+tv_app_releases`)
}

func TestAppAPKClientTypeMigration(t *testing.T) {
	t.Parallel()

	up := readMigrationForTest(t, "0024_app_apk_distribution_client_type.up.sql")
	down := readMigrationForTest(t, "0024_app_apk_distribution_client_type.down.sql")

	assertSQLPattern(t, up, `(?is)alter\s+table\s+tv_app_releases\s+add\s+column\s+if\s+not\s+exists\s+client_type`)
	assertSQLPattern(t, up, `(?is)client_type\s+in\s+\('android_tv',\s*'android_phone'\)`)
	assertSQLPattern(t, up, `(?is)package_name\s*=\s*'com\.chee\.videos\.tv'`)
	assertSQLPattern(t, up, `(?is)package_name\s*=\s*'com\.chee\.videos'`)
	assertSQLPattern(t, up, `(?is)create\s+index\s+if\s+not\s+exists\s+idx_tv_app_releases_client_type_version_code`)
	assertSQLPattern(t, up, `(?is)abi\s+in\s+\('armeabi-v7a',\s*'arm64-v8a',\s*'single'\)`)

	assertSQLPattern(t, down, `(?is)delete\s+from\s+tv_app_releases\s+where\s+client_type\s*=\s*'android_phone'`)
	assertSQLPattern(t, down, `(?is)drop\s+index\s+if\s+exists\s+idx_tv_app_releases_client_type_publish_status`)
	assertSQLPattern(t, down, `(?is)drop\s+column\s+if\s+exists\s+client_type`)
}

func TestArchiveImportBatchMigration(t *testing.T) {
	t.Parallel()

	up := readMigrationForTest(t, "0025_archive_import_batches.up.sql")
	down := readMigrationForTest(t, "0025_archive_import_batches.down.sql")

	assertSQLPattern(t, up, `(?is)create\s+table\s+if\s+not\s+exists\s+archive_import_batches`)
	assertSQLPattern(t, up, `(?is)archive_format\s+varchar\(10\)\s+not\s+null`)
	assertSQLPattern(t, up, `(?is)status\s+varchar\(24\)\s+not\s+null\s+default\s+'uploaded'`)
	assertSQLPattern(t, up, `(?is)create\s+table\s+if\s+not\s+exists\s+archive_import_files`)
	assertSQLPattern(t, up, `(?is)batch_id\s+uuid\s+not\s+null\s+references\s+archive_import_batches\(id\)\s+on\s+delete\s+cascade`)
	assertSQLPattern(t, up, `(?is)unique\s*\(\s*batch_id\s*,\s*relative_path\s*\)`)
	assertSQLPattern(t, down, `(?is)drop\s+table\s+if\s+exists\s+archive_import_files`)
	assertSQLPattern(t, down, `(?is)drop\s+table\s+if\s+exists\s+archive_import_batches`)
}

func TestPasswordVaultEntriesMigration(t *testing.T) {
	t.Parallel()

	up := readMigrationForTest(t, "0026_password_vault_entries.up.sql")
	down := readMigrationForTest(t, "0026_password_vault_entries.down.sql")

	assertSQLPattern(t, up, `(?is)create\s+table\s+if\s+not\s+exists\s+password_vault_entries`)
	assertSQLPattern(t, up, `(?is)password_ciphertext\s+text\s+not\s+null`)
	assertSQLPattern(t, up, `(?is)create\s+index\s+if\s+not\s+exists\s+idx_password_vault_entries_updated`)
	assertSQLPattern(t, down, `(?is)drop\s+table\s+if\s+exists\s+password_vault_entries`)
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
