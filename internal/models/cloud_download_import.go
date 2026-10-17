package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Cloud download import kinds. A task imports either one video file or one
// archive that enters the existing archive import batch flow.
const (
	CloudImportKindVideo   = "video"
	CloudImportKindArchive = "archive"
)

// Cloud download import statuses. queued -> hashing -> copying -> creating,
// then imported or skipped_existing; any step may end in failed.
const (
	CloudImportQueued          = "queued"
	CloudImportHashing         = "hashing"
	CloudImportCopying         = "copying"
	CloudImportCreating        = "creating"
	CloudImportImported        = "imported"
	CloudImportSkippedExisting = "skipped_existing"
	CloudImportFailed          = "failed"
)

var (
	// ErrCloudImportNotFound means the download task id does not exist.
	ErrCloudImportNotFound = errors.New("cloud: download task not found for import")
	// ErrCloudImportNotCompleted means the task has not finished downloading.
	ErrCloudImportNotCompleted = errors.New("cloud: download task is not completed")
	// ErrCloudImportFileMissing means local_path is absent, outside the
	// download directory, or not a regular file.
	ErrCloudImportFileMissing = errors.New("cloud: downloaded file is missing")
	// ErrCloudImportUnsupported means the file cannot enter either media flow.
	ErrCloudImportUnsupported = errors.New("cloud: downloaded file is not supported")
	// ErrCloudImportInvalidInput means the import form does not match kind.
	ErrCloudImportInvalidInput = errors.New("cloud: invalid import input")
)

// CloudImportTerminal reports whether an import record is finished.
func CloudImportTerminal(status string) bool {
	return status == CloudImportImported || status == CloudImportSkippedExisting
}

// CloudImportActive reports whether a worker job may still own the record.
func CloudImportActive(status string) bool {
	switch status {
	case CloudImportQueued, CloudImportHashing, CloudImportCopying, CloudImportCreating:
		return true
	default:
		return false
	}
}

// CloudDownloadImport is the one-shot import attached to a cloud task.
type CloudDownloadImport struct {
	ID             int64      `json:"id"`
	TaskID         int64      `json:"task_id"`
	Kind           string     `json:"kind"`
	Status         string     `json:"status"`
	SourcePath     string     `json:"source_path"`
	SourceSize     int64      `json:"source_size"`
	SHA256         string     `json:"sha256"`
	VideoID        *uuid.UUID `json:"video_id"`
	ArchiveBatchID *uuid.UUID `json:"archive_batch_id"`
	Attempt        int        `json:"attempt"`
	ErrorMessage   string     `json:"error_message"`
	CreatedBy      *uuid.UUID `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	StartedAt      *time.Time `json:"started_at"`
	ImportedAt     *time.Time `json:"imported_at"`
}

// CloudDownloadImportSummary is embedded in the cloud task list response.
type CloudDownloadImportSummary struct {
	ID             int64      `json:"id"`
	Kind           string     `json:"kind"`
	Status         string     `json:"status"`
	VideoID        *uuid.UUID `json:"video_id,omitempty"`
	ArchiveBatchID *uuid.UUID `json:"archive_batch_id,omitempty"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// CloudDownloadImportPreview describes the actual downloaded file.
type CloudDownloadImportPreview struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Mime string `json:"mime"`
}

// CloudDownloadImportInput is the explicit import form. It travels in the
// cloud:ingest Redis payload so the worker can create the target after the
// HTTP request returns; the archive password is intentionally not persisted
// in the database.
type CloudDownloadImportInput struct {
	Kind              string      `json:"kind"`
	Type              string      `json:"type,omitempty"`
	SiteCategory      string      `json:"site_category,omitempty"`
	Title             string      `json:"title,omitempty"`
	Description       string      `json:"description,omitempty"`
	Tags              []string    `json:"tags,omitempty"`
	ActorIDs          []uuid.UUID `json:"actor_ids,omitempty"`
	ActorNames        []string    `json:"actor_names,omitempty"`
	CollectionIDs     []uuid.UUID `json:"collection_ids,omitempty"`
	ImageCollectionID *uuid.UUID  `json:"image_collection_id,omitempty"`

	DefaultDescription        string      `json:"default_description,omitempty"`
	DefaultTags               []string    `json:"default_tags,omitempty"`
	DefaultVideoCollectionIDs []uuid.UUID `json:"default_video_collection_ids,omitempty"`
	DefaultImageCollectionIDs []uuid.UUID `json:"default_image_collection_ids,omitempty"`
	HasPassword               bool        `json:"has_password,omitempty"`
	Password                  string      `json:"password,omitempty"`
}
