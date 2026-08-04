package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ForumPostDiscoverModeNormal     = "normal"
	ForumPostDiscoverModeDedupeOnly = "dedupe_only"

	ForumPostStatusDedupeOnly = "dedupe_only"
	ForumPostStatusPending    = "pending"
	ForumPostStatusInspected  = "inspected"
	ForumPostStatusRestricted = "restricted"
	ForumPostStatusFailed     = "failed"

	ForumPostFilterIncluded = "included"
	ForumPostFilterExcluded = "excluded"

	ForumPostDispositionCreated   = "created"
	ForumPostDispositionPending   = "pending"
	ForumPostDispositionDuplicate = "duplicate"
)

type ForumPostCandidate struct {
	TID   string `json:"tid"`
	Title string `json:"title,omitempty"`
	URL   string `json:"url,omitempty"`
}

type ForumPostDiscoverInput struct {
	Mode       string               `json:"mode"`
	Source     string               `json:"source"`
	BoardKey   string               `json:"board_key"`
	ObservedAt *time.Time           `json:"observed_at,omitempty"`
	Posts      []ForumPostCandidate `json:"posts"`
}

type ForumPostDiscoverResult struct {
	TID              string    `json:"tid"`
	ID               uuid.UUID `json:"id"`
	Disposition      string    `json:"disposition"`
	InspectionStatus string    `json:"inspection_status"`
}

type ForumPostInspectionInput struct {
	Status         string   `json:"status"`
	FilterDecision string   `json:"filter_decision"`
	FilterReasons  []string `json:"filter_reasons"`
	FetchMethod    string   `json:"fetch_method,omitempty"`
	ErrorSummary   string   `json:"error_summary,omitempty"`
	Attachments    []string `json:"attachments"`
	ED2KLinks      []string `json:"ed2k_links"`
}

type ForumPostInspectionResult struct {
	Status      string    `json:"inspection_status"`
	Idempotent  bool      `json:"idempotent"`
	InspectedAt time.Time `json:"inspected_at"`
}
