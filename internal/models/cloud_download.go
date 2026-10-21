package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Cloud download task statuses. The state machine is:
// pending -> submitting -> queued -> ready -> downloading -> completed,
// with failed (retryable terminal) and blocked (waiting for a human).
const (
	CloudDownloadPending     = "pending"
	CloudDownloadSubmitting  = "submitting"
	CloudDownloadQueued      = "queued"
	CloudDownloadReady       = "ready"
	CloudDownloadDownloading = "downloading"
	CloudDownloadCompleted   = "completed"
	CloudDownloadFailed      = "failed"
	CloudDownloadBlocked     = "blocked"
)

// Cloud download blocked reasons.
const (
	CloudBlockedLoginRequired  = "login_required"
	CloudBlockedQuotaExhausted = "quota_exhausted"
	CloudBlockedBrowserDown    = "browser_down"
	CloudBlockedNeedsAttention = "needs_attention"
)

// Cloud download task provenance. The first explicit action keeps its source;
// later submissions of the same info hash do not rewrite it.
const (
	CloudDownloadSourceForumPost = "forum_post"
	CloudDownloadSourceToolbox   = "toolbox"
	CloudDownloadSourceManual    = "manual"
)

// CloudDownloadBatchLimit is the maximum number of links one explicit batch
// submit may contain. It is enforced by both the handler and the service.
const CloudDownloadBatchLimit = 50

// Per-link outcomes for an explicit batch submit.
const (
	CloudBatchSubmitted        = "submitted"
	CloudBatchRetried          = "retried"
	CloudBatchSkippedCompleted = "skipped_completed"
	CloudBatchSkippedActive    = "skipped_active"
	CloudBatchDuplicate        = "duplicate"
	CloudBatchInvalid          = "invalid"
	CloudBatchFailed           = "failed"
)

var (
	// ErrCloudDownloadNotFound means the task id does not exist.
	ErrCloudDownloadNotFound = errors.New("cloud: download task not found")
	// ErrCloudDownloadNotRetryable means the task is not in a manual retry state.
	ErrCloudDownloadNotRetryable = errors.New("cloud: download task is not retryable")
)

// CloudDownloadTerminal reports whether a status needs no further work.
func CloudDownloadTerminal(status string) bool {
	return status == CloudDownloadCompleted || status == CloudDownloadFailed
}

// CloudDownloadTask is one ed2k/magnet link tracked through 115 cloud download.
type CloudDownloadTask struct {
	ID            int64      `json:"id"`
	PostID        *uuid.UUID `json:"post_id"`
	PostTitle     string     `json:"post_title"`
	Source        string     `json:"source"`
	Protocol      string     `json:"protocol"`
	InfoHash      string     `json:"info_hash"`
	SourceURL     string     `json:"source_url"`
	DeclaredName  string     `json:"declared_name"`
	DeclaredSize  *int64     `json:"declared_size"`
	Status        string     `json:"status"`
	BlockedReason string     `json:"blocked_reason"`
	ErrorMessage  string     `json:"error_message"`
	RetryCount    int        `json:"retry_count"`
	PickCode      string     `json:"pick_code"`
	// CloudFileID is 115's delete_file_id (the actual cloud file); CloudParentID
	// is its wp_path_id. They are deliberately separate from the 115 task's
	// file_id, which can be the parent cloud-download directory.
	CloudFileID       string     `json:"-"`
	CloudParentID     string     `json:"-"`
	CloudUserID       string     `json:"-"`
	CloudFileCategory int        `json:"-"`
	CloudStatus       string     `json:"cloud_status"`
	ProgressPercent   *float64   `json:"progress_percent"`
	ProgressUpdatedAt *time.Time `json:"progress_updated_at"`
	LocalPath         string     `json:"local_path"`
	// DownloadTriggeredAt is set exactly once before the native download is
	// triggered, so a redelivered job does not start a second local copy.
	DownloadTriggeredAt *time.Time `json:"download_triggered_at"`
	CreatedBy           *uuid.UUID `json:"created_by"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	StartedAt           *time.Time `json:"started_at"`
	FinishedAt          *time.Time `json:"finished_at"`
	// Import is the explicit media-library import summary, when one exists.
	Import *CloudDownloadImportSummary `json:"import,omitempty"`
	// FileDeletion is the optional 115 recycle-bin cleanup summary.
	FileDeletion *CloudDownloadFileDeletionSummary `json:"file_deletion,omitempty"`
}

// CloudDownloadStatus is the compact per-hash view used by the forum list.
type CloudDownloadStatus struct {
	TaskID        int64  `json:"task_id"`
	Status        string `json:"status"`
	BlockedReason string `json:"blocked_reason"`
}

// CloudDownloadBatchItem is one link's outcome in an explicit batch submit.
type CloudDownloadBatchItem struct {
	Link     string `json:"link"`
	InfoHash string `json:"info_hash,omitempty"`
	Status   string `json:"status"`
	TaskID   *int64 `json:"task_id,omitempty"`
	Message  string `json:"message,omitempty"`
}

// CloudDownloadBatchSummary counts item outcomes for the batch response.
type CloudDownloadBatchSummary struct {
	Total            int `json:"total"`
	Submitted        int `json:"submitted"`
	Retried          int `json:"retried"`
	SkippedCompleted int `json:"skipped_completed"`
	SkippedActive    int `json:"skipped_active"`
	Duplicate        int `json:"duplicate"`
	Invalid          int `json:"invalid"`
	Failed           int `json:"failed"`
}

// CloudDownloadBatchResult is returned by the explicit batch endpoint.
type CloudDownloadBatchResult struct {
	Items   []CloudDownloadBatchItem  `json:"items"`
	Summary CloudDownloadBatchSummary `json:"summary"`
}
