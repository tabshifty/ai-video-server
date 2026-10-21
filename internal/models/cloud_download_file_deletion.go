package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Cloud download file deletion statuses. The record is one per task and
// tracks moving the 115 cloud source file to the 115 recycle bin. A deleted
// record is terminal; the download task itself remains completed.
const (
	CloudFileDeleting = "deleting"
	CloudFileDeleted  = "deleted"
	CloudFileFailed   = "failed"
)

var (
	// ErrCloudFileDeleteNotFound means the download task id does not exist.
	ErrCloudFileDeleteNotFound = errors.New("cloud: download task not found for file deletion")
	// ErrCloudFileDeleteNotCompleted means the source file has not finished
	// downloading.
	ErrCloudFileDeleteNotCompleted = errors.New("cloud: downloaded file is not completed")
	// ErrCloudFileDeleteLocalMissing means the local copy is gone; deleting the
	// 115 source would remove the only remaining copy.
	ErrCloudFileDeleteLocalMissing = errors.New("cloud: downloaded local file is missing")
	// ErrCloudFileDeleteUnavailable means the historical 115 task cannot be
	// resolved to an actual delete_file_id/wp_path_id.
	ErrCloudFileDeleteUnavailable = errors.New("cloud: cannot locate the 115 cloud file")
	// ErrCloudFileDeleteInProgress means another request currently owns the
	// deletion attempt.
	ErrCloudFileDeleteInProgress = errors.New("cloud: 115 cloud file deletion is already in progress")
	// ErrCloudFileDeleteBrowserDown means the 115 browser is not reachable.
	ErrCloudFileDeleteBrowserDown = errors.New("cloud: 115 browser is unavailable")
	// ErrCloudFileDeleteTimeout means the cleanup did not finish in its budget.
	ErrCloudFileDeleteTimeout = errors.New("cloud: 115 cloud file deletion timed out")
	// ErrCloudFileDeleteLoginRequired means the 115 browser session is gone.
	ErrCloudFileDeleteLoginRequired = errors.New("cloud: 115 login is required")
	// ErrCloudFileDeleteFailed means 115 acknowledged the request but reported a
	// non-recoverable deletion failure.
	ErrCloudFileDeleteFailed = errors.New("cloud: 115 cloud file deletion failed")
)

// CloudDownloadFileDeletion is the audit record for one task's cleanup.
type CloudDownloadFileDeletion struct {
	TaskID       int64      `json:"task_id"`
	FileID       string     `json:"file_id"`
	ParentID     string     `json:"parent_id"`
	Status       string     `json:"status"`
	ErrorMessage string     `json:"error_message"`
	Note         string     `json:"note"`
	Attempt      int        `json:"attempt"`
	RequestedBy  *uuid.UUID `json:"requested_by"`
	RequestedAt  time.Time  `json:"requested_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CloudDownloadFileDeletionSummary is embedded in the cloud task list.
type CloudDownloadFileDeletionSummary struct {
	Status       string     `json:"status"`
	ErrorMessage string     `json:"error_message,omitempty"`
	Note         string     `json:"note,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CloudFileDeletionTerminal reports whether the cleanup is finished.
func CloudFileDeletionTerminal(status string) bool {
	return status == CloudFileDeleted || status == CloudFileFailed
}

// ErrCloudFileDeletionNotFound means no cleanup record exists for the task.
var ErrCloudFileDeletionNotFound = errors.New("cloud: file deletion record not found")
