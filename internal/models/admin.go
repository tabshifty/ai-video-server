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
	ID              uuid.UUID       `json:"id"`
	UserID          *uuid.UUID      `json:"user_id"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Type            string          `json:"type"`
	Status          string          `json:"status"`
	DurationSeconds int             `json:"duration_seconds"`
	Width           int             `json:"width"`
	Height          int             `json:"height"`
	OriginalPath    string          `json:"original_path"`
	TranscodedPath  string          `json:"transcoded_path"`
	ThumbnailPath   string          `json:"thumbnail_path"`
	Metadata        json.RawMessage `json:"metadata"`
	Tags            []string        `json:"tags"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
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
	ID         int64      `json:"id"`
	VideoID    *uuid.UUID `json:"video_id"`
	UserID     *uuid.UUID `json:"user_id"`
	Status     string     `json:"status"`
	RetryCount int        `json:"retry_count"`
	Error      string     `json:"error"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
}
