package config

import "testing"

func TestTelegramForumConfiguration(t *testing.T) {
	for _, tt := range []struct {
		name, owner, token string
		bad                bool
	}{
		{"requires owner", "", "download", true},
		{"reject negative owner", "-1", "download", true},
		{"dedicated credential", "123", "ingestion", true},
		{"configured", "123", "download", false},
		{"disabled", "", "", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			setRequiredEnv(t)
			t.Setenv("HERMES_API_TOKEN", "ingestion")
			t.Setenv("TELEGRAM_FORUM_DOWNLOAD_TOKEN", tt.token)
			t.Setenv("TELEGRAM_FORUM_OWNER_ID", tt.owner)
			cfg, err := Load()
			if (err != nil) != tt.bad {
				t.Fatalf("err=%v", err)
			}
			if !tt.bad && tt.token != "" && cfg.TelegramForumOwnerID != 123 {
				t.Fatalf("owner=%d", cfg.TelegramForumOwnerID)
			}
		})
	}
}
