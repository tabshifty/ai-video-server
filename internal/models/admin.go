package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AdminStats struct {
	TotalVideos       int64          `json:"total_videos"`
	ShortVideos       int64          `json:"short_videos"`
	MovieVideos       int64          `json:"movie_videos"`
	EpisodeVideos     int64          `json:"episode_videos"`
	AVVideos          int64          `json:"av_videos"`
	TotalUsers        int64          `json:"total_users"`
	TodayUploads      int64          `json:"today_uploads"`
	QueueLength       int64          `json:"queue_length"`
	DiskTotalBytes    uint64         `json:"disk_total_bytes"`
	DiskFreeBytes     uint64         `json:"disk_free_bytes"`
	WeeklyUploadTrend []DailyUploads `json:"weekly_upload_trend"`
}

type DailyUploads struct {
	Day   string `json:"day"`
	Count int64  `json:"count"`
}

type AdminVideoListItem struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	Thumbnail    string     `json:"thumbnail"`
	UploadUserID *uuid.UUID `json:"upload_user_id"`
	UploadUser   string     `json:"upload_user"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type AdminVideoDetail struct {
	ID              uuid.UUID              `json:"id"`
	UserID          *uuid.UUID             `json:"user_id"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Type            string                 `json:"type"`
	Status          string                 `json:"status"`
	DurationSeconds int                    `json:"duration_seconds"`
	Width           int                    `json:"width"`
	Height          int                    `json:"height"`
	OriginalPath    string                 `json:"original_path"`
	TranscodedPath  string                 `json:"transcoded_path"`
	ThumbnailPath   string                 `json:"thumbnail_path"`
	Metadata        json.RawMessage        `json:"metadata"`
	Tags            []string               `json:"tags"`
	Actors          []AdminVideoActor      `json:"actors"`
	Collections     []AdminVideoCollection `json:"collections"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type AdminVideoActor struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	AvatarURL  string    `json:"avatar_url"`
	Active     bool      `json:"active"`
	BindSource string    `json:"bind_source"`
}

type AdminVideoCollection struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CoverURL  string    `json:"cover_url"`
	SortOrder int       `json:"sort_order"`
	Active    bool      `json:"active"`
}

type AdminVideoFilter struct {
	Page      int
	PageSize  int
	Keyword   string
	Type      string
	Status    string
	User      string
	Tag       string
	StartTime *time.Time
	EndTime   *time.Time
}

type AdminUserListItem struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminTaskListItem struct {
	ID                    int64      `json:"id"`
	VideoID               *uuid.UUID `json:"video_id"`
	UserID                *uuid.UUID `json:"user_id"`
	Status                string     `json:"status"`
	RetryCount            int        `json:"retry_count"`
	Error                 string     `json:"error"`
	StartedAt             *time.Time `json:"started_at"`
	FinishedAt            *time.Time `json:"finished_at"`
	SourceDurationSeconds *int       `json:"source_duration_seconds"`
	ProcessedSeconds      *int       `json:"processed_seconds"`
	RemainingSeconds      *int       `json:"remaining_seconds"`
	ProgressPercent       *float64   `json:"progress_percent"`
	ProgressUpdatedAt     *time.Time `json:"progress_updated_at"`
}

type AdminActor struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Aliases    []string  `json:"aliases"`
	Gender     string    `json:"gender"`
	Country    string    `json:"country"`
	BirthDate  string    `json:"birth_date"`
	AvatarURL  string    `json:"avatar_url"`
	Source     string    `json:"source"`
	ExternalID string    `json:"external_id"`
	Notes      string    `json:"notes"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type AdminActorInput struct {
	Name       string   `json:"name"`
	Aliases    []string `json:"aliases"`
	Gender     string   `json:"gender"`
	Country    string   `json:"country"`
	BirthDate  string   `json:"birth_date"`
	AvatarURL  string   `json:"avatar_url"`
	Source     string   `json:"source"`
	ExternalID string   `json:"external_id"`
	Notes      string   `json:"notes"`
	Active     bool     `json:"active"`
}

type AdminCollection struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CoverURL    string    `json:"cover_url"`
	SortOrder   int       `json:"sort_order"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AdminCollectionInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
	SortOrder   int    `json:"sort_order"`
	Active      bool   `json:"active"`
}

type AdminImageListItem struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Status       string     `json:"status"`
	Active       bool       `json:"active"`
	StoredPath   string     `json:"stored_path"`
	StoredMIME   string     `json:"stored_mime"`
	FileSize     int64      `json:"file_size"`
	Width        int        `json:"width"`
	Height       int        `json:"height"`
	UploadUser   string     `json:"upload_user"`
	UploadUserID *uuid.UUID `json:"upload_user_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type AdminImageActor struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	AvatarURL  string    `json:"avatar_url"`
	Active     bool      `json:"active"`
	BindSource string    `json:"bind_source"`
}

type AdminImageCollection struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CoverURL    string    `json:"cover_url"`
	SortOrder   int       `json:"sort_order"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AdminImageDetail struct {
	ID           uuid.UUID              `json:"id"`
	UserID       *uuid.UUID             `json:"user_id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Status       string                 `json:"status"`
	Active       bool                   `json:"active"`
	OriginalPath string                 `json:"original_path"`
	StoredPath   string                 `json:"stored_path"`
	OriginalMIME string                 `json:"original_mime"`
	StoredMIME   string                 `json:"stored_mime"`
	OriginalExt  string                 `json:"original_ext"`
	StoredExt    string                 `json:"stored_ext"`
	FileSize     int64                  `json:"file_size"`
	Width        int                    `json:"width"`
	Height       int                    `json:"height"`
	Metadata     json.RawMessage        `json:"metadata"`
	Actors       []AdminImageActor      `json:"actors"`
	Collections  []AdminImageCollection `json:"collections"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type AdminImageFilter struct {
	Page         int
	PageSize     int
	Keyword      string
	Status       string
	Active       *bool
	ActorID      *uuid.UUID
	CollectionID *uuid.UUID
}

type AdminImageCollectionInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
	SortOrder   int    `json:"sort_order"`
	Active      bool   `json:"active"`
}
