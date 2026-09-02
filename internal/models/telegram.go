package models

import (
	"time"

	"github.com/google/uuid"
)

// TelegramSource describes one administrator-selected Telegram chat.
// ChatID is zero until the ingestor resolves ChatRef with the personal account.
type TelegramSource struct {
	ID                     uuid.UUID  `json:"id"`
	ChatID                 int64      `json:"chat_id"`
	ChatRef                string     `json:"chat_ref"`
	Title                  string     `json:"title"`
	Username               string     `json:"username"`
	Enabled                bool       `json:"enabled"`
	SyncStatus             string     `json:"sync_status"`
	HistoryCursorMessageID int64      `json:"history_cursor_message_id"`
	LastError              string     `json:"last_error"`
	NextRetryAt            *time.Time `json:"next_retry_at,omitempty"`
	BackfillCompletedAt    *time.Time `json:"backfill_completed_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// TelegramMedia records a Telegram video message and its local import state.
type TelegramMedia struct {
	ID               uuid.UUID  `json:"id"`
	SourceID         uuid.UUID  `json:"source_id"`
	ChatID           int64      `json:"chat_id"`
	MessageID        int64      `json:"message_id"`
	DocumentID       *int64     `json:"document_id,omitempty"`
	DocumentDCID     *int       `json:"document_dc_id,omitempty"`
	Filename         string     `json:"filename"`
	MIMEType         string     `json:"mime_type"`
	FileSize         int64      `json:"file_size"`
	Caption          string     `json:"caption"`
	MessageURL       string     `json:"message_url"`
	MessageCreatedAt time.Time  `json:"message_created_at"`
	ProcessingStatus string     `json:"processing_status"`
	TranscodeStatus  string     `json:"transcode_status"`
	VideoID          *uuid.UUID `json:"video_id,omitempty"`
	SHA256           string     `json:"sha256,omitempty"`
	TempPath         string     `json:"temp_path,omitempty"`
	Attempts         int        `json:"attempts"`
	NextRetryAt      *time.Time `json:"next_retry_at,omitempty"`
	LastError        string     `json:"last_error"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
