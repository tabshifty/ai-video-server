package services

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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

type archiveImportBatchPlannedFile struct {
	File models.ArchiveImportFileListItem
}

func planArchiveFilenameBatchUpdate(files []models.ArchiveImportFileListItem, in ArchiveImportBatchUpdateInput) ([]archiveImportBatchPlannedFile, error) {
	if in.TitleMode != ArchiveImportTitleModeFilename {
		return nil, &ArchiveImportBatchUpdateError{
			Reason: ArchiveImportBatchReasonInvalidPatch,
			Err:    fmt.Errorf("title_mode 仅支持 filename"),
		}
	}
	if len(in.Targets) == 0 {
		return nil, &ArchiveImportBatchUpdateError{
			Reason: ArchiveImportBatchReasonInvalidSelection,
			Err:    fmt.Errorf("至少选择一个文件"),
		}
	}

	targets := make(map[uuid.UUID]ArchiveImportBatchUpdateTarget, len(in.Targets))
	for _, target := range in.Targets {
		if target.ID == uuid.Nil || target.UpdatedAt.IsZero() {
			return nil, &ArchiveImportBatchUpdateError{
				Reason: ArchiveImportBatchReasonInvalidSelection,
				Err:    fmt.Errorf("目标 ID 和更新时间不能为空"),
			}
		}
		if _, exists := targets[target.ID]; exists {
			return nil, &ArchiveImportBatchUpdateError{
				Reason: ArchiveImportBatchReasonInvalidSelection,
				Err:    fmt.Errorf("文件 ID 重复"),
			}
		}
		targets[target.ID] = target
	}

	byID := make(map[uuid.UUID]models.ArchiveImportFileListItem, len(files))
	for _, file := range files {
		byID[file.ID] = file
	}
	if len(byID) != len(targets) {
		return nil, &ArchiveImportBatchUpdateError{
			Reason: ArchiveImportBatchReasonInvalidSelection,
			Err:    fmt.Errorf("部分文件不存在"),
		}
	}
	for id := range targets {
		if _, exists := byID[id]; !exists {
			return nil, &ArchiveImportBatchUpdateError{
				Reason: ArchiveImportBatchReasonInvalidSelection,
				Err:    fmt.Errorf("部分文件不存在"),
			}
		}
	}

	issues := make([]ArchiveImportBatchUpdateIssue, 0)
	batchID := byID[in.Targets[0].ID].BatchID
	for _, target := range in.Targets {
		file := byID[target.ID]
		if file.BatchID != batchID {
			issues = append(issues, batchUpdateIssueForFile(file, "所选文件不属于同一批次"))
		}
	}
	if len(issues) > 0 {
		return nil, batchUpdateError(ArchiveImportBatchReasonInvalidSelection, issues)
	}

	for _, target := range in.Targets {
		file := byID[target.ID]
		if !canReplaceArchiveFilenameTitle(file) {
			issues = append(issues, batchUpdateIssueForFile(file, "仅待处理或失败的视频可以替换标题"))
		}
	}
	if len(issues) > 0 {
		return nil, batchUpdateError(ArchiveImportBatchReasonIneligibleTarget, issues)
	}

	for _, target := range in.Targets {
		file := byID[target.ID]
		if !file.UpdatedAt.Equal(target.UpdatedAt) {
			issues = append(issues, batchUpdateIssueForFile(file, "文件信息已变化，请刷新后重试"))
		}
	}
	if len(issues) > 0 {
		return nil, batchUpdateError(ArchiveImportBatchReasonStaleTarget, issues)
	}

	for _, target := range in.Targets {
		file := byID[target.ID]
		title := deriveArchiveFilenameTitle(file.RelativePath)
		if title == "" || utf8.RuneCountInString(title) > 200 {
			issues = append(issues, batchUpdateIssueForFile(file, "文件名无法生成有效标题"))
		}
	}
	if len(issues) > 0 {
		return nil, batchUpdateError(ArchiveImportBatchReasonEmptyTitle, issues)
	}

	plans := make([]archiveImportBatchPlannedFile, 0, len(in.Targets))
	for _, target := range in.Targets {
		file := byID[target.ID]
		newTitle := deriveArchiveFilenameTitle(file.RelativePath)
		file.Description = mergeArchiveFilenameTitleDescription(file.Title, newTitle, file.Description)
		file.Title = newTitle
		if in.UpdateTags {
			file.Tags = normalizeArchiveTags(in.Tags)
		}
		if in.UpdateVideoType {
			file.VideoType = strings.ToLower(strings.TrimSpace(in.VideoType))
		}
		if in.UpdateVideoCollectionIDs {
			file.VideoCollectionIDs = dedupeArchiveUUIDs(in.VideoCollectionIDs)
		}
		if in.UpdateImageCollectionIDs {
			file.ImageCollectionIDs = dedupeArchiveUUIDs(in.ImageCollectionIDs)
		}
		plans = append(plans, archiveImportBatchPlannedFile{File: file})
	}
	return plans, nil
}

func batchUpdateIssueForFile(file models.ArchiveImportFileListItem, message string) ArchiveImportBatchUpdateIssue {
	return ArchiveImportBatchUpdateIssue{
		ID:           file.ID,
		RelativePath: file.RelativePath,
		Message:      message,
	}
}

func batchUpdateError(reason string, issues []ArchiveImportBatchUpdateIssue) *ArchiveImportBatchUpdateError {
	return &ArchiveImportBatchUpdateError{Reason: reason, Issues: issues}
}

// BatchUpdateFiles replaces selected video titles and applies optional field patches atomically.
func (s *ArchiveImportService) BatchUpdateFiles(ctx context.Context, in ArchiveImportBatchUpdateInput) ([]models.ArchiveImportFileListItem, error) {
	if in.TitleMode != ArchiveImportTitleModeFilename {
		return nil, &ArchiveImportBatchUpdateError{
			Reason: ArchiveImportBatchReasonInvalidPatch,
			Err:    fmt.Errorf("title_mode 仅支持 filename"),
		}
	}
	if in.UpdateVideoType && !isArchiveImportVideoType(in.VideoType) {
		return nil, &ArchiveImportBatchUpdateError{
			Reason: ArchiveImportBatchReasonInvalidPatch,
			Err:    fmt.Errorf("视频类型无效"),
		}
	}

	var err error
	if in.UpdateVideoCollectionIDs {
		in.VideoCollectionIDs, err = s.resolveArchiveImportVideoCollectionIDs(ctx, in.VideoCollectionIDs)
		if err != nil {
			return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonInvalidPatch, Err: err}
		}
	}
	if in.UpdateImageCollectionIDs {
		in.ImageCollectionIDs, err = s.resolveArchiveImportVideoImageCollectionIDs(ctx, in.ImageCollectionIDs)
		if err != nil {
			return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonInvalidPatch, Err: err}
		}
	}

	ids := sortedArchiveImportBatchTargetIDs(in.Targets)
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonUpdateFailed, Err: err}
	}
	defer tx.Rollback(ctx)

	files, err := lockArchiveImportFilesTx(ctx, tx, ids)
	if err != nil {
		return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonUpdateFailed, Err: err}
	}
	plans, err := planArchiveFilenameBatchUpdate(files, in)
	if err != nil {
		return nil, err
	}
	batch, err := getArchiveImportBatchDefaultsTx(ctx, tx, plans[0].File.BatchID)
	if err != nil {
		return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonUpdateFailed, Err: err}
	}
	groups, err := getArchiveImportGroupsTx(ctx, tx, plans)
	if err != nil {
		return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonUpdateFailed, Err: err}
	}
	if err := applyArchiveImportBatchPlanTx(ctx, tx, plans, batch, groups, in); err != nil {
		return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonUpdateFailed, Err: err}
	}
	items, err := listArchiveFilesByIDsTx(ctx, tx, plans[0].File.BatchID, ids)
	if err != nil {
		return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonUpdateFailed, Err: err}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonUpdateFailed, Err: err}
	}
	return items, nil
}

func sortedArchiveImportBatchTargetIDs(targets []ArchiveImportBatchUpdateTarget) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(targets))
	for _, target := range targets {
		ids = append(ids, target.ID)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].String() < ids[j].String()
	})
	return ids
}

func lockArchiveImportFilesTx(ctx context.Context, tx pgx.Tx, ids []uuid.UUID) ([]models.ArchiveImportFileListItem, error) {
	rows, err := tx.Query(ctx, archiveImportFileSelectSQL(`
WHERE f.id = ANY($1)
ORDER BY f.id ASC
FOR UPDATE OF f
`), ids)
	if err != nil {
		return nil, fmt.Errorf("lock archive import files: %w", err)
	}
	defer rows.Close()

	files := make([]models.ArchiveImportFileListItem, 0, len(ids))
	for rows.Next() {
		file, scanErr := scanArchiveImportFileRecord(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan locked archive import file: %w", scanErr)
		}
		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate locked archive import files: %w", err)
	}
	return files, nil
}

func listArchiveFilesByIDsTx(ctx context.Context, tx pgx.Tx, batchID uuid.UUID, ids []uuid.UUID) ([]models.ArchiveImportFileListItem, error) {
	rows, err := tx.Query(ctx, archiveImportFileSelectSQL(`
WHERE f.batch_id = $1
  AND f.id = ANY($2)
ORDER BY f.relative_path ASC
`), batchID, ids)
	if err != nil {
		return nil, fmt.Errorf("list archive import files in transaction: %w", err)
	}
	defer rows.Close()

	items := make([]models.ArchiveImportFileListItem, 0, len(ids))
	for rows.Next() {
		item, scanErr := scanArchiveImportFileRecord(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan archive import file in transaction: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate archive import files in transaction: %w", err)
	}
	if len(items) != len(ids) {
		return nil, fmt.Errorf("archive import batch result incomplete: got %d files, want %d", len(items), len(ids))
	}
	return items, nil
}

func getArchiveImportBatchDefaultsTx(ctx context.Context, tx pgx.Tx, batchID uuid.UUID) (models.ArchiveImportBatch, error) {
	var batch models.ArchiveImportBatch
	var tagsRaw, videoCollectionsRaw, imageCollectionsRaw []byte
	if err := tx.QueryRow(ctx, `
SELECT
  id,
  COALESCE(title, ''),
  COALESCE(default_title_prefix, ''),
  COALESCE(default_description, ''),
  COALESCE(default_tags, '[]'::jsonb),
  COALESCE(default_video_collection_ids, '[]'::jsonb),
  COALESCE(default_image_collection_ids, '[]'::jsonb)
FROM archive_import_batches
WHERE id = $1
`, batchID).Scan(
		&batch.ID,
		&batch.Title,
		&batch.DefaultTitlePrefix,
		&batch.DefaultDescription,
		&tagsRaw,
		&videoCollectionsRaw,
		&imageCollectionsRaw,
	); err != nil {
		return models.ArchiveImportBatch{}, fmt.Errorf("get archive import batch defaults: %w", err)
	}
	if err := json.Unmarshal(tagsRaw, &batch.DefaultTags); err != nil {
		return models.ArchiveImportBatch{}, fmt.Errorf("decode archive import batch tags: %w", err)
	}
	if err := json.Unmarshal(videoCollectionsRaw, &batch.DefaultVideoCollectionIDs); err != nil {
		return models.ArchiveImportBatch{}, fmt.Errorf("decode archive import batch video collections: %w", err)
	}
	if err := json.Unmarshal(imageCollectionsRaw, &batch.DefaultImageCollectionIDs); err != nil {
		return models.ArchiveImportBatch{}, fmt.Errorf("decode archive import batch image collections: %w", err)
	}
	return batch, nil
}

func getArchiveImportGroupsTx(ctx context.Context, tx pgx.Tx, plans []archiveImportBatchPlannedFile) (map[uuid.UUID]models.ArchiveImportGroup, error) {
	groupIDs := make([]uuid.UUID, 0, len(plans))
	seen := make(map[uuid.UUID]struct{}, len(plans))
	for _, plan := range plans {
		if plan.File.GroupID == nil {
			continue
		}
		if _, exists := seen[*plan.File.GroupID]; exists {
			continue
		}
		seen[*plan.File.GroupID] = struct{}{}
		groupIDs = append(groupIDs, *plan.File.GroupID)
	}
	if len(groupIDs) == 0 {
		return map[uuid.UUID]models.ArchiveImportGroup{}, nil
	}
	sort.Slice(groupIDs, func(i, j int) bool {
		return groupIDs[i].String() < groupIDs[j].String()
	})

	rows, err := tx.Query(ctx, archiveImportGroupSelectSQL(`
WHERE id = ANY($1)
ORDER BY id ASC
`), groupIDs)
	if err != nil {
		return nil, fmt.Errorf("get archive import groups: %w", err)
	}
	defer rows.Close()

	groups := make(map[uuid.UUID]models.ArchiveImportGroup, len(groupIDs))
	for rows.Next() {
		group, scanErr := scanArchiveImportGroupRecord(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan archive import group: %w", scanErr)
		}
		groups[group.ID] = group
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate archive import groups: %w", err)
	}
	return groups, nil
}

func applyArchiveImportBatchPlanTx(
	ctx context.Context,
	tx pgx.Tx,
	plans []archiveImportBatchPlannedFile,
	batch models.ArchiveImportBatch,
	groups map[uuid.UUID]models.ArchiveImportGroup,
	in ArchiveImportBatchUpdateInput,
) error {
	for _, plan := range plans {
		file := plan.File
		group := archiveImportGroupPtrByID(groups, file.GroupID)
		defaults := resolveArchiveImportFileDefaults(file, batch, group)
		overrides := archiveImportFieldOverridesFromMap(file.FieldOverrides)
		overrides.Title = strings.TrimSpace(file.Title) != strings.TrimSpace(defaults.Title)
		overrides.Description = strings.TrimSpace(file.Description) != strings.TrimSpace(defaults.Description)
		if in.UpdateTags {
			overrides.Tags = !sameArchiveStringSet(file.Tags, defaults.Tags)
		}
		if in.UpdateVideoType {
			overrides.VideoType = !strings.EqualFold(strings.TrimSpace(file.VideoType), strings.TrimSpace(defaults.VideoType))
		}
		if in.UpdateVideoCollectionIDs {
			overrides.VideoCollectionIDs = !sameArchiveUUIDSet(file.VideoCollectionIDs, defaults.VideoCollectionIDs)
		}
		if in.UpdateImageCollectionIDs {
			overrides.ImageCollectionIDs = !sameArchiveUUIDSet(file.ImageCollectionIDs, defaults.ImageCollectionIDs)
		}
		if err := updateArchiveImportFileStateTx(ctx, tx, file, overrides); err != nil {
			return err
		}
	}
	return nil
}
