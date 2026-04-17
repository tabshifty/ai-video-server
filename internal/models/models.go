package models

import (
	"time"

	"github.com/google/uuid"
)

// Video is the primary media entity.
type Video struct {
	ID              uuid.UUID
	UserID          *uuid.UUID
	TMDBID          *int
	Title           string
	Description     string
	Type            string
	Status          string
	DurationSeconds int
	Width           int
	Height          int
	OriginalPath    string
	TranscodedPath  string
	ThumbnailPath   string
	Metadata        []byte
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// RecommendedVideo is a lightweight DTO for feed APIs.
type RecommendedVideo struct {
	ID             uuid.UUID         `json:"id"`
	Title          string            `json:"title"`
	Type           string            `json:"type"`
	TranscodedPath string            `json:"transcoded_path"`
	Duration       int               `json:"duration_seconds"`
	Collections    []VideoCollection `json:"collections"`
	Score          float64           `json:"score,omitempty"`
}

// VideoCollection is a lightweight DTO for video-collection relation.
type VideoCollection struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	CoverURL string    `json:"cover_url"`
}
