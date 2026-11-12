package models

import (
	"github.com/google/uuid"
	"time"
)

// VideoRotationJob is durable rotation and post-publication cleanup state.
type VideoRotationJob struct {
	ID            uuid.UUID `json:"id"`
	VideoID       uuid.UUID `json:"video_id"`
	RequestedBy   uuid.UUID `json:"-"`
	Angle         int       `json:"angle"`
	Mode          string    `json:"mode"`
	Status        string    `json:"status"`
	InputPath     string    `json:"-"`
	InputSize     int64     `json:"-"`
	InputMtime    int64     `json:"-"`
	OriginalPath  string    `json:"-"`
	OldThumbnail  string    `json:"-"`
	OutputPath    string    `json:"-"`
	ThumbnailPath string    `json:"-"`
	Progress      float64   `json:"progress"`
	ErrorMessage  string    `json:"error_message"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
