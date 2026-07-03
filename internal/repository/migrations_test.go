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

func TestEd2kDownloadTasksMigration(t *testing.T) {
	t.Parallel()

	up := readMigrationForTest(t, "0029_ed2k_download_tasks.up.sql")
	down := readMigrationForTest(t, "0029_ed2k_download_tasks.down.sql")

	assertSQLPattern(t, up, `(?is)create\s+table\s+if\s+not\s+exists\s+ed2k_download_tasks`)
	assertSQLPattern(t, up, `(?is)resource_hash\s+varchar\(64\)\s+not\s+null`)
	assertSQLPattern(t, up, `(?is)status\s+varchar\(24\)\s+not\s+null\s+default\s+'queued'`)
	assertSQLPattern(t, up, `(?is)'queued'.*'running'.*'completed'.*'failed'.*'deleted'`)
	assertSQLPattern(t, up, `(?is)create\s+unique\s+index\s+if\s+not\s+exists\s+idx_ed2k_download_tasks_resource_hash_unique`)
	assertSQLPattern(t, up, `(?is)where\s+status\s*<>\s*'deleted'`)
	assertSQLPattern(t, up, `(?is)create\s+index\s+if\s+not\s+exists\s+idx_ed2k_download_tasks_status_updated`)
	assertSQLPattern(t, down, `(?is)drop\s+table\s+if\s+exists\s+ed2k_download_tasks`)
}

func TestEd2kDownloadTaskLifecycleMigration(t *testing.T) {
	t.Parallel()

	up := readMigrationForTest(t, "0030_ed2k_download_task_lifecycle.up.sql")
	down := readMigrationForTest(t, "0030_ed2k_download_task_lifecycle.down.sql")

	assertSQLPattern(t, up, `(?is)drop\s+constraint\s+if\s+exists\s+ed2k_download_tasks_status_check`)
	assertSQLPattern(t, up, `(?is)'queued'.*'running'.*'canceling'.*'cancelled'.*'completed'.*'failed'.*'files_cleaned'.*'deleted'`)
	assertSQLPattern(t, up, `(?is)add\s+column\s+if\s+not\s+exists\s+cleaned_at\s+timestamptz`)
	assertSQLPattern(t, up, `(?is)drop\s+index\s+if\s+exists\s+idx_ed2k_download_tasks_resource_hash_unique`)
	assertSQLPattern(t, up, `(?is)where\s+status\s*<>\s*'deleted'\s+and\s+status\s*<>\s*'cancelled'`)

	assertSQLPattern(t, down, `(?is)update\s+ed2k_download_tasks\s+set\s+status\s*=\s*'running'\s+where\s+status\s*=\s*'canceling'`)
	assertSQLPattern(t, down, `(?is)update\s+ed2k_download_tasks\s+set\s+status\s*=\s*'deleted'\s+where\s+status\s*=\s*'cancelled'`)
	assertSQLPattern(t, down, `(?is)update\s+ed2k_download_tasks\s+set\s+status\s*=\s*'completed'\s+where\s+status\s*=\s*'files_cleaned'`)
	assertSQLPattern(t, down, `(?is)drop\s+column\s+if\s+exists\s+cleaned_at`)
	assertSQLPattern(t, down, `(?is)'queued'.*'running'.*'completed'.*'failed'.*'deleted'`)
	assertSQLPattern(t, down, `(?is)where\s+status\s*<>\s*'deleted'`)
}

func TestShortPendingDeleteMigration(t *testing.T) {
	t.Parallel()

	up := readMigrationForTest(t, "0031_short_pending_delete.up.sql")
	down := readMigrationForTest(t, "0031_short_pending_delete.down.sql")

	assertSQLPattern(t, up, `(?is)alter\s+table\s+videos\s+add\s+column\s+if\s+not\s+exists\s+pending_delete_at\s+timestamptz`)
	assertSQLPattern(t, up, `(?is)drop\s+constraint\s+if\s+exists\s+videos_status_check`)
	assertSQLPattern(t, up, `(?is)'uploaded'.*'scraping'.*'tv_pending'.*'av_scrape_pending'.*'processing'.*'ready'.*'pending_delete'.*'failed'`)
	assertSQLPattern(t, up, `(?is)add\s+constraint\s+videos_pending_delete_short_check\s+check\s*\(\s*status\s*<>\s*'pending_delete'\s+or\s+type\s*=\s*'short'\s*\)`)
	assertSQLPattern(t, up, `(?is)create\s+index\s+if\s+not\s+exists\s+idx_videos_short_pending_delete_at`)
	assertSQLPattern(t, up, `(?is)where\s+type\s*=\s*'short'\s+and\s+status\s*=\s*'pending_delete'`)

	assertSQLPattern(t, down, `(?is)drop\s+index\s+if\s+exists\s+idx_videos_short_pending_delete_at`)
	assertSQLPattern(t, down, `(?is)update\s+videos\s+set\s+status\s*=\s*'ready'.*pending_delete_at\s*=\s*null.*where\s+status\s*=\s*'pending_delete'`)
	assertSQLPattern(t, down, `(?is)drop\s+constraint\s+if\s+exists\s+videos_pending_delete_short_check`)
	assertSQLPattern(t, down, `(?is)'uploaded'.*'scraping'.*'tv_pending'.*'av_scrape_pending'.*'processing'.*'ready'.*'failed'`)
	assertSQLPattern(t, down, `(?is)alter\s+table\s+videos\s+drop\s+column\s+if\s+exists\s+pending_delete_at`)
}

func TestTVRemoteSessionMigration(t *testing.T) {
	t.Parallel()

	up := readMigrationForTest(t, "0032_tv_remote_sessions.up.sql")
	down := readMigrationForTest(t, "0032_tv_remote_sessions.down.sql")

	assertSQLPattern(t, up, `(?is)alter\s+table\s+tv_devices\s+add\s+column\s+if\s+not\s+exists\s+last_seen_at\s+timestamptz`)
	assertSQLPattern(t, up, `(?is)create\s+table\s+if\s+not\s+exists\s+tv_remote_sessions`)
	assertSQLPattern(t, up, `(?is)status\s+varchar\(16\)\s+not\s+null\s+default\s+'active'`)
	assertSQLPattern(t, up, `(?is)items\s+jsonb\s+not\s+null\s+default\s+'\[\]'\:\:jsonb`)
	assertSQLPattern(t, up, `(?is)current_index\s+int\s+not\s+null\s+default\s+0`)
	assertSQLPattern(t, up, `(?is)current_video_id\s+uuid`)
	assertSQLPattern(t, up, `(?is)'active'.*'ended'`)
	assertSQLPattern(t, up, `(?is)jsonb_typeof\(items\)\s*=\s*'array'`)
	assertSQLPattern(t, up, `(?is)create\s+unique\s+index\s+if\s+not\s+exists\s+idx_tv_remote_sessions_device_active_unique`)
	assertSQLPattern(t, up, `(?is)where\s+status\s*=\s*'active'`)

	assertSQLPattern(t, down, `(?is)drop\s+index\s+if\s+exists\s+idx_tv_remote_sessions_device_active_unique`)
	assertSQLPattern(t, down, `(?is)drop\s+index\s+if\s+exists\s+idx_tv_remote_sessions_user_updated_at`)
	assertSQLPattern(t, down, `(?is)drop\s+table\s+if\s+exists\s+tv_remote_sessions`)
	assertSQLPattern(t, down, `(?is)alter\s+table\s+tv_devices\s+drop\s+column\s+if\s+exists\s+last_seen_at`)
}

func TestMarkEd2kDownloadTaskRunningRefreshesStartedAt(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join(".", "ed2k_download_repository.go"))
	if err != nil {
		t.Fatalf("read repository source: %v", err)
	}
	source := string(raw)
	if !strings.Contains(source, "started_at = $3") {
		t.Fatal("expected running transition to refresh started_at directly")
	}
	if strings.Contains(source, "started_at = COALESCE($3, started_at)") {
		t.Fatal("expected running transition not to preserve stale started_at")
	}
}

func TestArchiveImportBatchEncodingMigration(t *testing.T) {
	t.Parallel()

	up := readMigrationForTest(t, "0028_archive_import_batch_encoding.up.sql")
	down := readMigrationForTest(t, "0028_archive_import_batch_encoding.down.sql")

	assertSQLPattern(t, up, `(?is)alter\s+table\s+archive_import_batches\s+add\s+column\s+if\s+not\s+exists\s+encoding_mode`)
	assertSQLPattern(t, up, `(?is)alter\s+table\s+archive_import_batches\s+add\s+column\s+if\s+not\s+exists\s+encoding_requested_mode`)
	assertSQLPattern(t, up, `(?is)'needs_password'.*'needs_encoding'.*'failed'`)
	assertSQLPattern(t, up, `(?is)encoding_mode\s+in\s+\('','utf8','gbk'\)`)
	assertSQLPattern(t, up, `(?is)encoding_requested_mode\s+in\s+\('auto','utf8','gbk'\)`)

	assertSQLPattern(t, down, `(?is)update\s+archive_import_batches\s+set\s+status\s*=\s*'failed'\s+where\s+status\s*=\s*'needs_encoding'`)
	assertSQLPattern(t, down, `(?is)drop\s+column\s+if\s+exists\s+encoding_requested_mode`)
	assertSQLPattern(t, down, `(?is)drop\s+column\s+if\s+exists\s+encoding_mode`)
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
