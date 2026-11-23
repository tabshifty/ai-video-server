package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrCloudIngestBatchInvalidInput means the explicit batch ingest request is
// malformed (title mode, video type, title length or collection/type mismatch).
var ErrCloudIngestBatchInvalidInput = errors.New("cloud: invalid batch ingest input")

// ErrCloudIngestBatchNotFound means the batch id does not exist.
var ErrCloudIngestBatchNotFound = errors.New("cloud: ingest batch not found")

// Batch ingest title modes. A batch either shares one人工 title or derives a
// title per link from its declared filename.
const (
	CloudIngestTitleUniform  = "uniform"
	CloudIngestTitleFilename = "filename"
)

// CloudIngestBatch is the batch-level ingest preset created by one explicit
// 批量入库 action. It holds no execution state: download progress and import
// outcome stay on the member tasks and their cloud_download_imports rows.
type CloudIngestBatch struct {
	ID                uuid.UUID   `json:"id"`
	TitleMode         string      `json:"title_mode"`
	UniformTitle      string      `json:"uniform_title,omitempty"`
	VideoType         string      `json:"video_type"`
	SiteCategory      string      `json:"site_category,omitempty"`
	Description       string      `json:"description,omitempty"`
	Tags              []string    `json:"tags,omitempty"`
	CollectionIDs     []uuid.UUID `json:"collection_ids,omitempty"`
	ActorIDs          []uuid.UUID `json:"actor_ids,omitempty"`
	ActorNames        []string    `json:"actor_names,omitempty"`
	ImageCollectionID *uuid.UUID  `json:"image_collection_id,omitempty"`
	ForcedCount       int         `json:"forced_count"`
	CreatedBy         *uuid.UUID  `json:"created_by,omitempty"`
	CreatedAt         time.Time   `json:"created_at"`
}

// CloudIngestBatchLink is one link in an explicit batch ingest request. Title
// is the final per-link title the admin already saw in the confirmation list;
// Forced marks a link whose extension could not be recognised as a video.
type CloudIngestBatchLink struct {
	Link   string `json:"link"`
	Title  string `json:"title,omitempty"`
	Forced bool   `json:"forced,omitempty"`
}

// CloudIngestBatchInput is the explicit 批量入库 request: batch-level ingest
// preset plus the confirmed links in display order.
type CloudIngestBatchInput struct {
	TitleMode         string                 `json:"title_mode"`
	UniformTitle      string                 `json:"uniform_title,omitempty"`
	VideoType         string                 `json:"video_type"`
	SiteCategory      string                 `json:"site_category,omitempty"`
	Description       string                 `json:"description,omitempty"`
	Tags              []string               `json:"tags,omitempty"`
	ActorIDs          []uuid.UUID            `json:"actor_ids,omitempty"`
	ActorNames        []string               `json:"actor_names,omitempty"`
	CollectionIDs     []uuid.UUID            `json:"collection_ids,omitempty"`
	ImageCollectionID *uuid.UUID             `json:"image_collection_id,omitempty"`
	Links             []CloudIngestBatchLink `json:"links"`
}

// CloudIngestBatchCounts aggregates member outcomes without storing them.
type CloudIngestBatchCounts struct {
	Members     int `json:"members"`
	Imported    int `json:"imported"`
	Skipped     int `json:"skipped_existing"`
	Failed      int `json:"failed"`
	Importing   int `json:"importing"`
	NotReady    int `json:"not_ready"`
	AwaitingRun int `json:"awaiting_import"`
}

// CloudIngestBatchSummary is one row of the batch filter list.
type CloudIngestBatchSummary struct {
	CloudIngestBatch
	Counts CloudIngestBatchCounts `json:"counts"`
}
