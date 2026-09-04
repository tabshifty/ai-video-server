package repository

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestTelegramManagementMigration(t *testing.T) {
	t.Parallel()

	up := readTelegramManagementMigration(t, "0039_telegram_management.up.sql")
	down := readTelegramManagementMigration(t, "0039_telegram_management.down.sql")

	for _, table := range []string{
		"telegram_account_state",
		"telegram_ingestor_heartbeats",
		"telegram_authorizations",
		"telegram_audit_logs",
	} {
		assertTelegramManagementSQL(t, up, `(?is)create\s+table\s+if\s+not\s+exists\s+`+table)
		assertTelegramManagementSQL(t, down, `(?is)drop\s+table\s+if\s+exists\s+`+table)
	}
	assertTelegramManagementSQL(t, up, `(?is)telegram_account_state[\s\S]*check\s*\(\s*id\s*=\s*1\s*\)`)
	assertTelegramManagementSQL(t, up, `(?is)telegram_ingestor_heartbeats[\s\S]*check\s*\(\s*id\s*=\s*1\s*\)`)
	assertTelegramManagementSQL(t, up, `(?is)'unconfigured'[\s\S]*'authorizing'[\s\S]*'authorized'[\s\S]*'reauthorizing'[\s\S]*'error'`)
	assertTelegramManagementSQL(t, up, `(?is)'pending'[\s\S]*'awaiting_code'[\s\S]*'awaiting_password'[\s\S]*'scanning'[\s\S]*'succeeded'[\s\S]*'failed'[\s\S]*'cancelled'[\s\S]*'expired'`)
	assertTelegramManagementSQL(t, up, `(?is)create\s+unique\s+index[\s\S]*telegram_authorizations[\s\S]*where\s+status\s+in`)
	assertTelegramManagementSQL(t, up, `(?is)actor_user_id\s+uuid\s+not\s+null\s+references\s+users\s*\(\s*id\s*\)`)
	assertTelegramManagementSQL(t, up, `(?is)telegram_audit_logs[\s\S]*created_at`)
	assertTelegramManagementSQL(t, up, `(?is)telegram_audit_logs[\s\S]*check\s*\(\s*jsonb_typeof\s*\(\s*summary\s*\)\s*=\s*'object'`)
}

func readTelegramManagementMigration(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
	if err != nil {
		t.Fatalf("读取 migration %s: %v", name, err)
	}
	return strings.TrimSpace(string(raw))
}

func assertTelegramManagementSQL(t *testing.T, sql, pattern string) {
	t.Helper()
	if !regexp.MustCompile(pattern).MatchString(sql) {
		t.Fatalf("管理台 migration 未匹配 %q:\n%s", pattern, sql)
	}
}
