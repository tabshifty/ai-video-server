package models

import (
	"time"

	"github.com/google/uuid"
)

// Video is the primary media entity.
type Video struct {
	ID                uuid.UUID
	UserID            *uuid.UUID
	TMDBID            *int
	ImageCollectionID *uuid.UUID
	Title             string
	Description       string
	Type              string
	Status            string
	DurationSeconds   int
	Width             int
	Height            int
	OriginalPath      string
	OSHash            string
	TranscodedPath    string
	ThumbnailPath     string
	Metadata          []byte
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// RecommendedVideo is a lightweight DTO for feed APIs.
type RecommendedVideo struct {
	ID             uuid.UUID         `json:"id"`
	Title          string            `json:"title"`
	Type           string            `json:"type"`
	ThumbnailPath  string            `json:"thumbnail_path"`
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

// Image is the primary image media entity.
type Image struct {
	ID           uuid.UUID
	UserID       *uuid.UUID
	Title        string
	Description  string
	Status       string
	Active       bool
	OriginalPath string
	StoredPath   string
	OriginalMIME string
	StoredMIME   string
	OriginalExt  string
	StoredExt    string
	FileSize     int64
	Width        int
	Height       int
	Metadata     []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type VideoSubtitle struct {
	ID           uuid.UUID
	VideoID      uuid.UUID
	SourceType   string
	Status       string
	LanguageCode string
	Label        string
	Format       string
	MIMEType     string
	StoredPath   string
	FileSize     int64
	IsDefault    bool
	SortOrder    int
	Metadata     []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ArchiveImportBatch struct {
	ID                        uuid.UUID
	UserID                    *uuid.UUID
	Title                     string
	OriginalFilename          string
	ArchiveFormat             string
	OriginalPath              string
	ExtractedDir              string
	Status                    string
	LastError                 string
	TotalEntries              int
	ProcessableEntries        int
	ProcessedEntries          int
	SkippedEntries            int
	FailedEntries             int
	DefaultTitlePrefix        string
	DefaultDescription        string
	DefaultTags               []string
	DefaultVideoCollectionIDs []uuid.UUID
	DefaultImageCollectionIDs []uuid.UUID
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	CompletedAt               *time.Time
}

type ArchiveImportBatchListItem struct {
	ID                 uuid.UUID  `json:"id"`
	Title              string     `json:"title"`
	OriginalFilename   string     `json:"original_filename"`
	ArchiveFormat      string     `json:"archive_format"`
	Status             string     `json:"status"`
	LastError          string     `json:"last_error"`
	TotalEntries       int        `json:"total_entries"`
	ProcessableEntries int        `json:"processable_entries"`
	ProcessedEntries   int        `json:"processed_entries"`
	SkippedEntries     int        `json:"skipped_entries"`
	FailedEntries      int        `json:"failed_entries"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	CompletedAt        *time.Time `json:"completed_at"`
}

type ArchiveImportFile struct {
	ID                 uuid.UUID
	BatchID            uuid.UUID
	GroupID            *uuid.UUID
	RelativePath       string
	FilePath           string
	EntryType          string
	MediaKind          string
	VideoType          string
	FileSize           int64
	MIMEType           string
	Status             string
	Reason             string
	Title              string
	Description        string
	Tags               []string
	VideoCollectionIDs []uuid.UUID
	ImageCollectionIDs []uuid.UUID
	FieldOverrides     map[string]bool
	LinkedVideoID      *uuid.UUID
	LinkedImageID      *uuid.UUID
	Metadata           map[string]any
	CreatedAt          time.Time
	UpdatedAt          time.Time
	ProcessedAt        *time.Time
}

type ArchiveImportFileListItem struct {
	ID                 uuid.UUID       `json:"id"`
	BatchID            uuid.UUID       `json:"batch_id"`
	GroupID            *uuid.UUID      `json:"group_id"`
	GroupName          string          `json:"group_name"`
	RelativePath       string          `json:"relative_path"`
	FilePath           string          `json:"file_path"`
	EntryType          string          `json:"entry_type"`
	MediaKind          string          `json:"media_kind"`
	VideoType          string          `json:"video_type"`
	FileSize           int64           `json:"file_size"`
	MIMEType           string          `json:"mime_type"`
	Status             string          `json:"status"`
	Reason             string          `json:"reason"`
	Title              string          `json:"title"`
	Description        string          `json:"description"`
	Tags               []string        `json:"tags"`
	VideoCollectionIDs []uuid.UUID     `json:"video_collection_ids"`
	ImageCollectionIDs []uuid.UUID     `json:"image_collection_ids"`
	FieldOverrides     map[string]bool `json:"field_overrides"`
	LinkedVideoID      *uuid.UUID      `json:"linked_video_id"`
	LinkedImageID      *uuid.UUID      `json:"linked_image_id"`
	Metadata           map[string]any  `json:"metadata"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	ProcessedAt        *time.Time      `json:"processed_at"`
}

type ArchiveImportGroup struct {
	ID                 uuid.UUID   `json:"id"`
	BatchID            uuid.UUID   `json:"batch_id"`
	Name               string      `json:"name"`
	Note               string      `json:"note"`
	MediaKind          string      `json:"media_kind"`
	Title              *string     `json:"title"`
	Description        *string     `json:"description"`
	Tags               []string    `json:"tags"`
	VideoType          *string     `json:"video_type"`
	VideoCollectionIDs []uuid.UUID `json:"video_collection_ids"`
	ImageCollectionIDs []uuid.UUID `json:"image_collection_ids"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

type ArchiveImportBatchDetail struct {
	ArchiveImportBatch
	Files  []ArchiveImportFileListItem `json:"files"`
	Groups []ArchiveImportGroup        `json:"groups"`
}
