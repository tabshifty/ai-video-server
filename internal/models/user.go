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
	LastSeenAt       *time.Time `json:"last_seen_at,omitempty"`
	IsOnline         bool       `json:"is_online"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type TvRemoteSessionItem struct {
	VideoID       uuid.UUID `json:"video_id"`
	Title         string    `json:"title"`
	ThumbnailPath string    `json:"thumbnail_path"`
	Duration      int       `json:"duration"`
	Type          string    `json:"type"`
}

type TvRemoteSearchContext struct {
	Query        string     `json:"query"`
	Type         string     `json:"type"`
	Page         int        `json:"page"`
	PageSize     int        `json:"page_size"`
	TotalCount   int        `json:"total_count"`
	CollectionID *uuid.UUID `json:"collection_id,omitempty"`
}

type TvRemoteSession struct {
	ID                  uuid.UUID              `json:"session_id"`
	UserID              uuid.UUID              `json:"-"`
	DeviceID            string                 `json:"device_id"`
	DeviceName          string                 `json:"device_name"`
	Platform            string                 `json:"platform"`
	Status              string                 `json:"status"`
	Items               []TvRemoteSessionItem  `json:"items"`
	SearchContext       *TvRemoteSearchContext `json:"search_context,omitempty"`
	CurrentIndex        int                    `json:"current_index"`
	CurrentVideoID      *uuid.UUID             `json:"current_video_id,omitempty"`
	CurrentItem         *TvRemoteSessionItem   `json:"current_item,omitempty"`
	HasPrevious         bool                   `json:"has_previous"`
	HasNext             bool                   `json:"has_next"`
	AutoplayNextEnabled bool                   `json:"autoplay_next_enabled"`
	EndedReason         string                 `json:"ended_reason,omitempty"`
	EndedAt             *time.Time             `json:"ended_at,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

type TvDeviceListPayload struct {
	Items []TvDeviceRecord `json:"items"`
}

type TvRemoteDeviceSessionPayload struct {
	Session *TvRemoteSession `json:"session"`
}
