// Package telegram defines the small Telegram client boundary used by ingestion services.
package telegram

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Message is the normalized video metadata needed by the ingestor.
type Message struct {
	ChatID       int64
	MessageID    int64
	DocumentID   int64
	DocumentDCID int
	Filename     string
	MIMEType     string
	Size         int64
	Caption      string
	MessageURL   string
	SentAt       time.Time
}

// Chat is the normalized identity of a Telegram chat.
type Chat struct {
	ID       int64
	Title    string
	Username string
}

const (
	// TelegramChatTypeGroup identifies a basic Telegram group.
	TelegramChatTypeGroup = "group"
	// TelegramChatTypeSupergroup identifies a Telegram supergroup.
	TelegramChatTypeSupergroup = "supergroup"
	// TelegramChatTypeChannel identifies a Telegram broadcast channel.
	TelegramChatTypeChannel = "channel"
)

// ChatPreviewResult describes a candidate Telegram source without exposing
// its invite token. A zero ChatID means the account must join before Telegram
// exposes a durable chat identity.
type ChatPreviewResult struct {
	ChatID           int64
	Title            string
	Username         string
	ChatType         string
	RequiresJoin     bool
	RequiresApproval bool
}

// Client contains only Telegram operations required by the ingestion service.
type Client interface {
	ResolveChat(ctx context.Context, ref string) (Chat, error)
	History(ctx context.Context, chatID, offsetID int64, limit int, fn func(Message) error) error
	Subscribe(ctx context.Context, fn func(Message) error) error
	RefreshMessage(ctx context.Context, chatID, messageID int64) (Message, error)
	Download(ctx context.Context, message Message, dst io.Writer) error
}

// FloodWaitError reports a Telegram rate limit without exposing gotd types.
type FloodWaitError struct {
	Seconds int
	Err     error
}

// Error implements error.
func (e *FloodWaitError) Error() string {
	if e == nil {
		return "Telegram FloodWait"
	}
	if e.Err == nil {
		return fmt.Sprintf("Telegram FloodWait: wait %d seconds", e.Seconds)
	}
	return fmt.Sprintf("Telegram FloodWait: wait %d seconds: %v", e.Seconds, e.Err)
}

// Unwrap returns the original Telegram error.
func (e *FloodWaitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
