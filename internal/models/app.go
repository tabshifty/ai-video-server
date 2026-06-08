package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// VideoDetail aggregates video info for detail page.
type VideoDetail struct {
	ID              uuid.UUID             `json:"id"`
	Title           string                `json:"title"`
	Description     string                `json:"description"`
	PlayURL         string                `json:"play_url"`
	TranscodedPath  string                `json:"transcoded_path"`
	ThumbnailPath   string                `json:"thumbnail_path"`
	Duration        int                   `json:"duration"`
	ViewsCount      int64                 `json:"views_count"`
	LikesCount      int64                 `json:"likes_count"`
	FavoritesCount  int64                 `json:"favorites_count"`
	Tags            []string              `json:"tags"`
	Actors          []VideoActor          `json:"actors"`
	Collections     []VideoCollection     `json:"collections"`
	ImageCollection *VideoImageCollection `json:"image_collection,omitempty"`
	SubtitleTracks  []SubtitleTrack       `json:"subtitle_tracks"`
	Metadata        json.RawMessage       `json:"metadata"`
	UserState       VideoUserState        `json:"user_state"`
}

// VideoActor is a lightweight DTO for video-actor relation.
type VideoActor struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
}

type ActorDetail struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Aliases   []string  `json:"aliases"`
	Gender    string    `json:"gender"`
	Country   string    `json:"country"`
	BirthDate string    `json:"birth_date"`
	AvatarURL string    `json:"avatar_url"`
}

type ActorWorksPayload struct {
	Actor      ActorDetail     `json:"actor"`
	Items      []VideoListItem `json:"items"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
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
	Type          string    `json:"type"`
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
	ID             uuid.UUID         `json:"id"`
	Title          string            `json:"title"`
	Type           string            `json:"type"`
	ThumbnailPath  string            `json:"thumbnail_path"`
	TranscodedPath string            `json:"transcoded_path"`
	Duration       int               `json:"duration"`
	Collections    []VideoCollection `json:"collections"`
	CreatedAt      time.Time         `json:"created_at"`
	Metadata       json.RawMessage   `json:"metadata"`
}

type VideoImageCollection struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	CoverURL string    `json:"cover_url"`
}

type ImageCollectionListItem struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CoverURL    string    `json:"cover_url"`
	ImageCount  int       `json:"image_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ImageCollectionImage struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ThumbnailURL string    `json:"thumbnail_url"`
	ViewURL      string    `json:"view_url"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
}

type ImageCollectionDetail struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	CoverURL    string                 `json:"cover_url"`
	ImageCount  int                    `json:"image_count"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Images      []ImageCollectionImage `json:"images"`
}

type TvHomePayload struct {
	Kind             string                 `json:"kind,omitempty"`
	Featured         *TvHomeVideoDto        `json:"featured,omitempty"`
	RecentWatching   []TvHomeVideoDto       `json:"recent_watching,omitempty"`
	RecentUpdates    []TvHomeVideoDto       `json:"recent_updates,omitempty"`
	ContinueWatching *TvContinueWatchingDto `json:"continue_watching,omitempty"`
	Sections         []TvSectionDto         `json:"sections"`
	SearchResults    []TvSeriesSummaryDto   `json:"search_results"`
	TvSeries         []TvHomeVideoDto       `json:"tv_series"`
	Movies           []TvHomeVideoDto       `json:"movies"`
	AV               []TvHomeVideoDto       `json:"av"`
	Page             int                    `json:"page"`
	PageSize         int                    `json:"page_size"`
}

type TvContinueWatchingDto struct {
	Type            string `json:"type"`
	SeriesID        int64  `json:"series_id"`
	SeriesTitle     string `json:"series_title"`
	SeasonNumber    int    `json:"season_number"`
	EpisodeNumber   int    `json:"episode_number"`
	EpisodeTitle    string `json:"episode_title"`
	VideoID         string `json:"video_id"`
	PosterURL       string `json:"poster_url"`
	BackdropURL     string `json:"backdrop_url"`
	WatchSeconds    int    `json:"watch_seconds"`
	DurationSeconds int    `json:"duration_seconds"`
	ProgressPercent int    `json:"progress_percent"`
}

type TvHomeVideoDto struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	Title           string `json:"title"`
	Overview        string `json:"overview"`
	PosterURL       string `json:"poster_url"`
	BackdropURL     string `json:"backdrop_url"`
	VideoID         string `json:"video_id,omitempty"`
	SeasonNumber    int    `json:"season_number,omitempty"`
	EpisodeNumber   int    `json:"episode_number,omitempty"`
	ProgressPercent int    `json:"progress_percent,omitempty"`
}

type TvCatalogWallItemDto struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	Title           string `json:"title"`
	Overview        string `json:"overview"`
	PosterURL       string `json:"poster_url"`
	BackdropURL     string `json:"backdrop_url"`
	VideoID         string `json:"video_id,omitempty"`
	SeasonNumber    int    `json:"season_number,omitempty"`
	EpisodeNumber   int    `json:"episode_number,omitempty"`
	ProgressPercent int    `json:"progress_percent,omitempty"`
}

type TvSearchPayload struct {
	Items      []TvSearchResultDto `json:"items"`
	TotalCount int                 `json:"total_count"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}

type TvSearchResultDto struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Overview    string `json:"overview"`
	PosterURL   string `json:"poster_url"`
	BackdropURL string `json:"backdrop_url"`
}

type TvSectionDto struct {
	Title    string               `json:"title"`
	Subtitle string               `json:"subtitle"`
	Items    []TvSeriesSummaryDto `json:"items"`
}

type TvSeriesSummaryDto struct {
	ID                   int64  `json:"id"`
	Title                string `json:"title"`
	Overview             string `json:"overview"`
	PosterURL            string `json:"poster_url"`
	BackdropURL          string `json:"backdrop_url"`
	FirstAirDate         string `json:"first_air_date"`
	TotalSeasons         int    `json:"total_seasons"`
	TotalEpisodes        int    `json:"total_episodes"`
	PlayableEpisodes     int    `json:"playable_episodes"`
	LatestEpisodeAirDate string `json:"latest_episode_air_date"`
}

type TvSeriesDetailDto struct {
	ID               int64         `json:"id"`
	Title            string        `json:"title"`
	Overview         string        `json:"overview"`
	PosterURL        string        `json:"poster_url"`
	BackdropURL      string        `json:"backdrop_url"`
	FirstAirDate     string        `json:"first_air_date"`
	TotalSeasons     int           `json:"total_seasons"`
	TotalEpisodes    int           `json:"total_episodes"`
	PlayableEpisodes int           `json:"playable_episodes"`
	Tags             []string      `json:"tags"`
	Cast             []string      `json:"cast"`
	Seasons          []TvSeasonDto `json:"seasons"`
}

type TvSeasonDto struct {
	ID           int64          `json:"id"`
	SeasonNumber int            `json:"season_number"`
	Title        string         `json:"title"`
	Overview     string         `json:"overview"`
	PosterURL    string         `json:"poster_url"`
	AirDate      string         `json:"air_date"`
	Episodes     []TvEpisodeDto `json:"episodes"`
}

type TvEpisodeDto struct {
	ID              int64           `json:"id"`
	EpisodeNumber   int             `json:"episode_number"`
	Title           string          `json:"title"`
	Overview        string          `json:"overview"`
	Runtime         int             `json:"runtime"`
	AirDate         string          `json:"air_date"`
	StillURL        string          `json:"still_url"`
	VideoID         string          `json:"video_id"`
	VideoTitle      string          `json:"video_title"`
	VideoStatus     string          `json:"video_status"`
	WatchSeconds    int             `json:"watch_seconds"`
	ProgressPercent int             `json:"progress_percent"`
	LastWatchedAt   string          `json:"last_watched_at"`
	Playable        bool            `json:"playable"`
	SubtitleTracks  []SubtitleTrack `json:"subtitle_tracks"`
	Metadata        json.RawMessage `json:"metadata"`
}

type SubtitleTrack struct {
	ID            uuid.UUID `json:"id"`
	SourceType    string    `json:"source_type"`
	LanguageCode  string    `json:"language_code"`
	LanguageLabel string    `json:"language_label"`
	Label         string    `json:"label"`
	Format        string    `json:"format"`
	URL           string    `json:"url"`
	MIMEType      string    `json:"mime_type"`
	IsDefault     bool      `json:"is_default"`
	IsEmbedded    bool      `json:"is_embedded"`
	EmbeddedIndex int       `json:"embedded_index"`
	Available     bool      `json:"available"`
}
