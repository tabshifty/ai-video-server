package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// VideoDetail aggregates video info for detail page.
type VideoDetail struct {
	ID             uuid.UUID       `json:"id"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	PlayURL        string          `json:"play_url"`
	TranscodedPath string          `json:"transcoded_path"`
	ThumbnailPath  string          `json:"thumbnail_path"`
	Duration       int             `json:"duration"`
	ViewsCount     int64           `json:"views_count"`
	LikesCount     int64           `json:"likes_count"`
	FavoritesCount int64           `json:"favorites_count"`
	Tags           []string        `json:"tags"`
	Metadata       json.RawMessage `json:"metadata"`
	UserState      VideoUserState  `json:"user_state"`
}

// VideoUserState is per-user interaction state for a video.
type VideoUserState struct {
	IsLiked      bool `json:"is_liked"`
	IsFavorited  bool `json:"is_favorited"`
	IsDisliked   bool `json:"is_disliked"`
	WatchSeconds int  `json:"watch_seconds"`
	IsCompleted  bool `json:"is_completed"`
}

// HistoryItem represents continue-watching entry.
type HistoryItem struct {
	VideoID       uuid.UUID `json:"video_id"`
	Title         string    `json:"title"`
	ThumbnailPath string    `json:"thumbnail_path"`
	Duration      int       `json:"duration"`
	WatchSeconds  int       `json:"watch_seconds"`
	Progress      float64   `json:"progress"`
	LastWatchedAt time.Time `json:"last_watched_at"`
}

// PageResult contains paginated list response data.
type PageResult[T any] struct {
	Items      []T `json:"items"`
	TotalCount int `json:"total_count"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
}

// UserProfileView is profile response payload.
type UserProfileView struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// VideoListItem is a generic user video list item.
type VideoListItem struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Type           string    `json:"type"`
	ThumbnailPath  string    `json:"thumbnail_path"`
	TranscodedPath string    `json:"transcoded_path"`
	Duration       int       `json:"duration"`
	CreatedAt      time.Time `json:"created_at"`
}
