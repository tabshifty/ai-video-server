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
	ID                uuid.UUID              `json:"id"`
	UserID            *uuid.UUID             `json:"user_id"`
	ImageCollectionID *uuid.UUID             `json:"image_collection_id"`
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	Type              string                 `json:"type"`
	Status            string                 `json:"status"`
	DurationSeconds   int                    `json:"duration_seconds"`
	Width             int                    `json:"width"`
	Height            int                    `json:"height"`
	OriginalPath      string                 `json:"original_path"`
	TranscodedPath    string                 `json:"transcoded_path"`
	ThumbnailPath     string                 `json:"thumbnail_path"`
	Metadata          json.RawMessage        `json:"metadata"`
	Tags              []string               `json:"tags"`
	Actors            []AdminVideoActor      `json:"actors"`
	Collections       []AdminVideoCollection `json:"collections"`
	ImageCollection   *AdminImageCollection  `json:"image_collection,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type AdminVideoSubtitle struct {
	ID            uuid.UUID       `json:"id"`
	VideoID       uuid.UUID       `json:"video_id"`
	SourceType    string          `json:"source_type"`
	Status        string          `json:"status"`
	LanguageCode  string          `json:"language_code"`
	Label         string          `json:"label"`
	Format        string          `json:"format"`
	MIMEType      string          `json:"mime_type"`
	StoredPath    string          `json:"stored_path"`
	FileSize      int64           `json:"file_size"`
	IsDefault     bool            `json:"is_default"`
	SortOrder     int             `json:"sort_order"`
	Metadata      json.RawMessage `json:"metadata"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	EmbeddedIndex int             `json:"embedded_index"`
}

type VideoTagStat struct {
	Tag       string `json:"tag"`
	UsedCount int64  `json:"used_count"`
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
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	CoverURL       string     `json:"cover_url"`
	ManualCoverURL string     `json:"manual_cover_url"`
	CoverImageID   *uuid.UUID `json:"cover_image_id"`
	SortOrder      int        `json:"sort_order"`
	Active         bool       `json:"active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
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
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	CoverURL     string     `json:"cover_url"`
	CoverImageID *uuid.UUID `json:"cover_image_id"`
	SortOrder    int        `json:"sort_order"`
	Active       bool       `json:"active"`
}

type AdminTvSeriesListItem struct {
	ID               int64  `json:"id"`
	Title            string `json:"title"`
	Overview         string `json:"overview"`
	PosterURL        string `json:"poster_url"`
	BackdropURL      string `json:"backdrop_url"`
	FirstAirDate     string `json:"first_air_date"`
	TotalSeasons     int    `json:"total_seasons"`
	TotalEpisodes    int    `json:"total_episodes"`
	PlayableEpisodes int    `json:"playable_episodes"`
	Active           bool   `json:"active"`
}

type AdminTvSeriesDetail struct {
	ID               int64                 `json:"id"`
	Title            string                `json:"title"`
	Overview         string                `json:"overview"`
	PosterURL        string                `json:"poster_url"`
	BackdropURL      string                `json:"backdrop_url"`
	FirstAirDate     string                `json:"first_air_date"`
	TotalSeasons     int                   `json:"total_seasons"`
	TotalEpisodes    int                   `json:"total_episodes"`
	PlayableEpisodes int                   `json:"playable_episodes"`
	Active           bool                  `json:"active"`
	Seasons          []AdminTvSeasonDetail `json:"seasons"`
}

type AdminTvSeasonDetail struct {
	ID           int64                  `json:"id"`
	SeriesID     int64                  `json:"series_id"`
	SeasonNumber int                    `json:"season_number"`
	Title        string                 `json:"title"`
	Overview     string                 `json:"overview"`
	PosterURL    string                 `json:"poster_url"`
	AirDate      string                 `json:"air_date"`
	Episodes     []AdminTvEpisodeDetail `json:"episodes"`
}

type AdminTvEpisodeDetail struct {
	ID            int64  `json:"id"`
	SeasonID      int64  `json:"season_id"`
	EpisodeNumber int    `json:"episode_number"`
	Title         string `json:"title"`
	Overview      string `json:"overview"`
	Runtime       int    `json:"runtime"`
	AirDate       string `json:"air_date"`
	StillURL      string `json:"still_url"`
	VideoID       string `json:"video_id"`
	VideoTitle    string `json:"video_title"`
	VideoStatus   string `json:"video_status"`
	Playable      bool   `json:"playable"`
}

type AdminTvSeriesInput struct {
	Title        string `json:"title"`
	Overview     string `json:"overview"`
	PosterURL    string `json:"poster_url"`
	BackdropURL  string `json:"backdrop_url"`
	FirstAirDate string `json:"first_air_date"`
	Active       bool   `json:"active"`
}

type AdminTvSeasonInput struct {
	SeasonNumber int    `json:"season_number"`
	Title        string `json:"title"`
	Overview     string `json:"overview"`
	PosterURL    string `json:"poster_url"`
	AirDate      string `json:"air_date"`
}

type AdminTvEpisodeInput struct {
	EpisodeNumber int    `json:"episode_number"`
	Title         string `json:"title"`
	Overview      string `json:"overview"`
	Runtime       int    `json:"runtime"`
	AirDate       string `json:"air_date"`
	StillURL      string `json:"still_url"`
	VideoID       string `json:"video_id"`
}
