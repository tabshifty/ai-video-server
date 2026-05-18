package models

import "time"

type IPTVChannel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	Group     string `json:"group"`
	LogoURL   string `json:"logo_url"`
	TVGID     string `json:"tvg_id"`
	SortOrder int    `json:"sort_order"`
}

type IPTVPlaylistMeta struct {
	SourceURL    string
	UpdatedAt    *time.Time
	SkippedCount int
}

type IPTVPlaylistStatus struct {
	SourceURL    string        `json:"source_url"`
	UpdatedAt    *time.Time    `json:"updated_at"`
	ChannelCount int           `json:"channel_count"`
	SkippedCount int           `json:"skipped_count"`
	Groups       []string      `json:"groups"`
	Channels     []IPTVChannel `json:"channels"`
}
