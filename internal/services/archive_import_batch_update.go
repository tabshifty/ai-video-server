package services

import (
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
)

var archiveFilenameAdPattern = regexp.MustCompile(`(?i)www\.98T\.la@`)

// ArchiveImportTitleModeFilename replaces a title with one derived from its archive filename.
const ArchiveImportTitleModeFilename = "filename"

const (
	// ArchiveImportBatchReasonInvalidSelection indicates that no valid targets were selected.
	ArchiveImportBatchReasonInvalidSelection = "invalid_selection"
	// ArchiveImportBatchReasonIneligibleTarget indicates that a selected target cannot be updated.
	ArchiveImportBatchReasonIneligibleTarget = "ineligible_target"
	// ArchiveImportBatchReasonStaleTarget indicates that a selected target changed after it was loaded.
	ArchiveImportBatchReasonStaleTarget = "stale_target"
	// ArchiveImportBatchReasonEmptyTitle indicates that filename cleanup produced an empty title.
	ArchiveImportBatchReasonEmptyTitle = "empty_derived_title"
	// ArchiveImportBatchReasonInvalidPatch indicates that the requested field patch is invalid.
	ArchiveImportBatchReasonInvalidPatch = "invalid_patch"
	// ArchiveImportBatchReasonUpdateFailed indicates that persistence of the batch update failed.
	ArchiveImportBatchReasonUpdateFailed = "update_failed"
)

// ArchiveImportBatchUpdateTarget identifies a file and the version the caller observed.
type ArchiveImportBatchUpdateTarget struct {
	ID        uuid.UUID
	UpdatedAt time.Time
}

// ArchiveImportBatchUpdateInput describes a batch patch for selected archive files.
type ArchiveImportBatchUpdateInput struct {
	Targets                  []ArchiveImportBatchUpdateTarget
	TitleMode                string
	UpdateTags               bool
	Tags                     []string
	UpdateVideoType          bool
	VideoType                string
	UpdateVideoCollectionIDs bool
	VideoCollectionIDs       []uuid.UUID
	UpdateImageCollectionIDs bool
	ImageCollectionIDs       []uuid.UUID
}

// ArchiveImportBatchUpdateIssue describes one target that prevented a batch update.
type ArchiveImportBatchUpdateIssue struct {
	ID           uuid.UUID `json:"id"`
	RelativePath string    `json:"relative_path"`
	Message      string    `json:"message"`
}

// ArchiveImportBatchUpdateError reports a stable failure reason and affected targets.
type ArchiveImportBatchUpdateError struct {
	Reason string                          `json:"reason"`
	Issues []ArchiveImportBatchUpdateIssue `json:"issues"`
	Err    error                           `json:"-"`
}

// Error returns the wrapped error message when available, otherwise the stable reason.
func (e *ArchiveImportBatchUpdateError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Reason
}

// Unwrap exposes the underlying update error.
func (e *ArchiveImportBatchUpdateError) Unwrap() error {
	return e.Err
}

func deriveArchiveFilenameTitle(relativePath string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(relativePath), `\`, "/")
	if normalized == "" {
		return ""
	}
	base := path.Base(normalized)
	if base == "." || base == "/" {
		return ""
	}
	base = strings.TrimSuffix(base, path.Ext(base))
	base = archiveFilenameAdPattern.ReplaceAllString(base, "")
	return strings.Join(strings.Fields(base), " ")
}

func mergeArchiveFilenameTitleDescription(oldTitle, newTitle, description string) string {
	oldTitle = strings.TrimSpace(oldTitle)
	newTitle = strings.TrimSpace(newTitle)
	if oldTitle == "" || oldTitle == newTitle {
		return description
	}
	if description == "" {
		return oldTitle
	}
	return oldTitle + "\n" + description
}

func canReplaceArchiveFilenameTitle(file models.ArchiveImportFileListItem) bool {
	if file.EntryType != "file" || file.MediaKind != "video" {
		return false
	}
	return file.Status == "pending" || file.Status == "failed"
}
