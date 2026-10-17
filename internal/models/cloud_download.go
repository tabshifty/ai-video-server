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
	ID                int64      `json:"id"`
	PostID            *uuid.UUID `json:"post_id"`
	PostTitle         string     `json:"post_title"`
	Protocol          string     `json:"protocol"`
	InfoHash          string     `json:"info_hash"`
	SourceURL         string     `json:"source_url"`
	DeclaredName      string     `json:"declared_name"`
	DeclaredSize      *int64     `json:"declared_size"`
	Status            string     `json:"status"`
	BlockedReason     string     `json:"blocked_reason"`
	ErrorMessage      string     `json:"error_message"`
	RetryCount        int        `json:"retry_count"`
	PickCode          string     `json:"pick_code"`
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
}

// CloudDownloadStatus is the compact per-hash view used by the forum list.
type CloudDownloadStatus struct {
	TaskID        int64  `json:"task_id"`
	Status        string `json:"status"`
	BlockedReason string `json:"blocked_reason"`
}
