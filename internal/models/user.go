package models

import (
	"time"

	"github.com/google/uuid"
)

// User is the account identity entity.
type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AuthTokens holds issued access/refresh tokens.
type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TvAuthSession struct {
	ID               uuid.UUID  `json:"session_id"`
	PairCode         string     `json:"pair_code"`
	DeviceID         string     `json:"device_id"`
	DeviceName       string     `json:"device_name"`
	Platform         string     `json:"platform"`
	Status           string     `json:"status"`
	UserID           *uuid.UUID `json:"-"`
	AccessToken      string     `json:"-"`
	RefreshToken     string     `json:"-"`
	ApprovedUsername string     `json:"-"`
	ApprovedRole     string     `json:"-"`
	ApprovedAt       *time.Time `json:"-"`
	ExpiresAt        time.Time  `json:"expires_at"`
	CreatedAt        time.Time  `json:"-"`
	UpdatedAt        time.Time  `json:"-"`
}

type TvAuthSessionCreateResult struct {
	SessionID           uuid.UUID `json:"session_id"`
	PairCode            string    `json:"pair_code"`
	QRContent           string    `json:"qr_content"`
	ExpiresAt           time.Time `json:"expires_at"`
	PollIntervalSeconds int       `json:"poll_interval_seconds"`
}

type TvAuthSessionPollResult struct {
	SessionID     uuid.UUID        `json:"session_id"`
	Status        string           `json:"status"`
	ExpiresAt     time.Time        `json:"expires_at"`
	AccessToken   string           `json:"access_token,omitempty"`
	RefreshToken  string           `json:"refresh_token,omitempty"`
	User          *TvAuthUserBrief `json:"user,omitempty"`
	DeviceName    string           `json:"device_name,omitempty"`
	PairCode      string           `json:"pair_code,omitempty"`
	ServerBaseURL string           `json:"server_base_url,omitempty"`
}

type TvAuthUserBrief struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
}

type TvDeviceRecord struct {
	ID               uuid.UUID  `json:"id"`
	DeviceID         string     `json:"device_id"`
	DeviceName       string     `json:"device_name"`
	Platform         string     `json:"platform"`
	UserID           *uuid.UUID `json:"user_id,omitempty"`
	LastAuthorizedAt *time.Time `json:"last_authorized_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
