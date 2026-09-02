package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

type telegramTestRow struct {
	values []any
}

func (r telegramTestRow) Scan(dest ...any) error {
	for i, value := range r.values {
		switch target := dest[i].(type) {
		case *uuid.UUID:
			*target = value.(uuid.UUID)
		case *int64:
			*target = value.(int64)
		case *int:
			*target = int(value.(int64))
		case *string:
			*target = value.(string)
		case *bool:
			*target = value.(bool)
		case *time.Time:
			*target = value.(time.Time)
		case *sql.NullInt64:
			*target = value.(sql.NullInt64)
		case *sql.NullInt32:
			*target = value.(sql.NullInt32)
		case *sql.NullString:
			*target = value.(sql.NullString)
		case *sql.NullTime:
			*target = value.(sql.NullTime)
		default:
			panic("unsupported Telegram test scan target")
		}
	}
	return nil
}

func TestScanTelegramSourcePreservesUnresolvedChatAndNullableTimes(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 9, 2, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	item, err := scanTelegramSource(telegramTestRow{values: []any{
		uuid.New(), int64(0), "invite-link", "待解析群组", "", true, "pending", int64(0), "",
		sql.NullTime{}, sql.NullTime{}, createdAt, updatedAt,
	}})
	if err != nil {
		t.Fatalf("scanTelegramSource returned error: %v", err)
	}
	if item.ChatID != 0 {
		t.Fatalf("ChatID = %d, want unresolved zero value", item.ChatID)
	}
	if item.NextRetryAt != nil || item.BackfillCompletedAt != nil {
		t.Fatalf("nullable source timestamps should remain nil: next=%v completed=%v", item.NextRetryAt, item.BackfillCompletedAt)
	}
	if item.ChatRef != "invite-link" || item.SyncStatus != "pending" {
		t.Fatalf("unexpected source: %+v", item)
	}
}

func TestScanTelegramMediaPreservesNullableIdentityAndRetryFields(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 9, 2, 7, 0, 0, 0, time.UTC)
	item, err := scanTelegramMedia(telegramTestRow{values: []any{
		uuid.New(), uuid.New(), int64(-100123), int64(42), sql.NullInt64{}, sql.NullInt32{},
		"clip.mp4", "video/mp4", int64(128), "caption", "https://t.me/c/123/42", createdAt,
		"discovered", "not_required", sql.NullString{}, sql.NullString{}, sql.NullString{}, int64(0),
		sql.NullTime{}, sql.NullString{}, createdAt, createdAt,
	}})
	if err != nil {
		t.Fatalf("scanTelegramMedia returned error: %v", err)
	}
	if item.DocumentID != nil || item.DocumentDCID != nil || item.VideoID != nil {
		t.Fatalf("nullable Telegram identity fields should remain nil: %+v", item)
	}
	if item.SHA256 != "" || item.TempPath != "" || item.NextRetryAt != nil || item.LastError != "" {
		t.Fatalf("nullable processing fields should remain empty: %+v", item)
	}
	if item.ChatID != -100123 || item.MessageID != 42 || item.FileSize != 128 {
		t.Fatalf("unexpected media identity fields: %+v", item)
	}
}

func TestNullableTelegramValues(t *testing.T) {
	t.Parallel()

	if got := nullableTelegramChatID(0); got != nil {
		t.Fatalf("nullableTelegramChatID(0) = %v, want nil", got)
	}
	if got := nullableTelegramChatID(-100123); got != int64(-100123) {
		t.Fatalf("nullableTelegramChatID(-100123) = %v, want -100123", got)
	}
	if got := nullableTelegramString("  "); got != nil {
		t.Fatalf("nullableTelegramString(whitespace) = %v, want nil", got)
	}
	if got := nullableTelegramString("error"); got != "error" {
		t.Fatalf("nullableTelegramString(error) = %v, want error", got)
	}
}
