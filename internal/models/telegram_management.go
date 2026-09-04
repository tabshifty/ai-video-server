package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	TelegramAccountStatusUnconfigured  = "unconfigured"
	TelegramAccountStatusAuthorizing   = "authorizing"
	TelegramAccountStatusAuthorized    = "authorized"
	TelegramAccountStatusReauthorizing = "reauthorizing"
	TelegramAccountStatusError         = "error"
)

const (
	TelegramIngestorStatusRunning     = "running"
	TelegramIngestorStatusAuthorizing = "authorizing"
	TelegramIngestorStatusDraining    = "draining"
	TelegramIngestorStatusError       = "error"
	TelegramIngestorStatusStopped     = "stopped"
)

const (
	TelegramAuthorizationKindPhone = "phone"
	TelegramAuthorizationKindQR    = "qr"
)

const (
	TelegramAuthorizationStatusPending          = "pending"
	TelegramAuthorizationStatusAwaitingCode     = "awaiting_code"
	TelegramAuthorizationStatusAwaitingPassword = "awaiting_password"
	TelegramAuthorizationStatusScanning         = "scanning"
	TelegramAuthorizationStatusSucceeded        = "succeeded"
	TelegramAuthorizationStatusFailed           = "failed"
	TelegramAuthorizationStatusCancelled        = "cancelled"
	TelegramAuthorizationStatusExpired          = "expired"
)

// TelegramAccountState stores non-sensitive metadata for the single system account.
// Session bytes and temporary authorization secrets are intentionally absent.
type TelegramAccountState struct {
	ID             int        `json:"id"`
	TelegramUserID *int64     `json:"telegram_user_id,omitempty"`
	Username       string     `json:"username"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	PhoneMasked    string     `json:"phone_masked"`
	Status         string     `json:"status"`
	LastError      string     `json:"last_error"`
	AuthorizedAt   *time.Time `json:"authorized_at,omitempty"`
	LastSeenAt     *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TelegramIngestorHeartbeat describes the latest observed collector process.
type TelegramIngestorHeartbeat struct {
	ID            int       `json:"id"`
	Status        string    `json:"status"`
	AccountStatus string    `json:"account_status"`
	Version       string    `json:"version"`
	ErrorSummary  string    `json:"error_summary"`
	LastSeenAt    time.Time `json:"last_seen_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TelegramAuthorization stores the public lifecycle of one short-lived login attempt.
type TelegramAuthorization struct {
	ID             uuid.UUID  `json:"id"`
	Kind           string     `json:"kind"`
	Status         string     `json:"status"`
	ActorUserID    uuid.UUID  `json:"actor_user_id"`
	TelegramUserID *int64     `json:"telegram_user_id,omitempty"`
	ErrorSummary   string     `json:"error_summary"`
	StartedAt      time.Time  `json:"started_at"`
	ExpiresAt      time.Time  `json:"expires_at"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TelegramAuditLog is an append-only, sanitized administrative event.
type TelegramAuditLog struct {
	ID          int64           `json:"id"`
	ActorUserID uuid.UUID       `json:"actor_user_id"`
	Action      string          `json:"action"`
	TargetType  string          `json:"target_type"`
	TargetID    string          `json:"target_id"`
	Result      string          `json:"result"`
	Summary     json.RawMessage `json:"summary"`
	CreatedAt   time.Time       `json:"created_at"`
}

// TelegramAuditFilter limits the read-only audit query.
type TelegramAuditFilter struct {
	Page      int
	PageSize  int
	Action    string
	Result    string
	StartTime *time.Time
	EndTime   *time.Time
}
