package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

const (
	archiveImportOverrideTitle              = "title"
	archiveImportOverrideDescription        = "description"
	archiveImportOverrideTags               = "tags"
	archiveImportOverrideVideoType          = "video_type"
	archiveImportOverrideVideoCollectionIDs = "video_collection_ids"
	archiveImportOverrideImageCollectionIDs = "image_collection_ids"
)

type ArchiveImportGroupCreateInput struct {
	Name               string
	Note               string
	FileIDs            []uuid.UUID
	Title              string
	Description        string
	Tags               []string
	VideoType          string
	VideoCollectionIDs []uuid.UUID
	ImageCollectionIDs []uuid.UUID
}

type ArchiveImportGroupUpdateInput struct {
	Name                     string
	Note                     string
	UpdateTitle              bool
	Title                    string
	UpdateDescription        bool
	Description              string
	UpdateTags               bool
	Tags                     []string
	UpdateVideoType          bool
	VideoType                string
	UpdateVideoCollectionIDs bool
	VideoCollectionIDs       []uuid.UUID
	UpdateImageCollectionIDs bool
	ImageCollectionIDs       []uuid.UUID
}

type archiveImportFieldOverrides struct {
	Title              bool `json:"title"`
	Description        bool `json:"description"`
	Tags               bool `json:"tags"`
	VideoType          bool `json:"video_type"`
	VideoCollectionIDs bool `json:"video_collection_ids"`
	ImageCollectionIDs bool `json:"image_collection_ids"`
}

type archiveImportResolvedDefaults struct {
	Title              string
	Description        string
	Tags               []string
	VideoType          string
	VideoCollectionIDs []uuid.UUID
	ImageCollectionIDs []uuid.UUID
}

type archiveImportScanner interface {
	Scan(dest ...any) error
}

func archiveImportFileSelectSQL(whereClause string) string {
	return `
SELECT
  f.id,
  f.batch_id,
  f.group_id,
  COALESCE(g.name, ''),
  f.relative_path,
  f.file_path,
  f.entry_type,
  f.media_kind,
  f.video_type,
  f.file_size,
  f.mime_type,
  f.status,
  COALESCE(f.reason, ''),
  COALESCE(f.title, ''),
  COALESCE(f.description, ''),
  COALESCE(f.tags, '[]'::jsonb),
  COALESCE(f.video_collection_ids, '[]'::jsonb),
  COALESCE(f.image_collection_ids, '[]'::jsonb),
  COALESCE(f.field_overrides, '{}'::jsonb),
  COALESCE(f.metadata, '{}'::jsonb),
  f.linked_video_id,
  f.linked_image_id,
  f.created_at,
  f.updated_at,
  f.processed_at
FROM archive_import_files f
LEFT JOIN archive_import_groups g ON g.id = f.group_id
` + whereClause
}

func scanArchiveImportFileRecord(scanner archiveImportScanner) (models.ArchiveImportFileListItem, error) {
	var item models.ArchiveImportFileListItem
	var tagsRaw, videoCollectionsRaw, imageCollectionsRaw, fieldOverridesRaw, metadataRaw []byte
	if err := scanner.Scan(
		&item.ID,
		&item.BatchID,
		&item.GroupID,
		&item.GroupName,
		&item.RelativePath,
		&item.FilePath,
		&item.EntryType,
		&item.MediaKind,
		&item.VideoType,
		&item.FileSize,
		&item.MIMEType,
		&item.Status,
		&item.Reason,
		&item.Title,
		&item.Description,
		&tagsRaw,
		&videoCollectionsRaw,
		&imageCollectionsRaw,
		&fieldOverridesRaw,
		&metadataRaw,
		&item.LinkedVideoID,
		&item.LinkedImageID,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.ProcessedAt,
	); err != nil {
		return models.ArchiveImportFileListItem{}, err
	}
	_ = json.Unmarshal(tagsRaw, &item.Tags)
	_ = json.Unmarshal(videoCollectionsRaw, &item.VideoCollectionIDs)
	_ = json.Unmarshal(imageCollectionsRaw, &item.ImageCollectionIDs)
	_ = json.Unmarshal(fieldOverridesRaw, &item.FieldOverrides)
	_ = json.Unmarshal(metadataRaw, &item.Metadata)
	if item.FieldOverrides == nil {
		item.FieldOverrides = map[string]bool{}
	}
	if item.Metadata == nil {
		item.Metadata = map[string]any{}
	}
	return item, nil
}

func archiveImportGroupSelectSQL(whereClause string) string {
	return `
SELECT
  id,
  batch_id,
  name,
  COALESCE(note, ''),
  media_kind,
  title,
  description,
  tags,
  video_type,
  video_collection_ids,
  image_collection_ids,
  created_at,
  updated_at
FROM archive_import_groups
` + whereClause
}

func scanArchiveImportGroupRecord(scanner archiveImportScanner) (models.ArchiveImportGroup, error) {
	var item models.ArchiveImportGroup
	var title, description, videoType sql.NullString
	var tagsRaw, videoCollectionsRaw, imageCollectionsRaw []byte
	if err := scanner.Scan(
		&item.ID,
		&item.BatchID,
		&item.Name,
		&item.Note,
		&item.MediaKind,
		&title,
		&description,
		&tagsRaw,
		&videoType,
		&videoCollectionsRaw,
		&imageCollectionsRaw,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return models.ArchiveImportGroup{}, err
	}
	if title.Valid {
		value := strings.TrimSpace(title.String)
		item.Title = &value
	}
	if description.Valid {
		value := strings.TrimSpace(description.String)
		item.Description = &value
	}
	if videoType.Valid {
		value := strings.ToLower(strings.TrimSpace(videoType.String))
		item.VideoType = &value
	}
	if len(tagsRaw) > 0 && string(tagsRaw) != "null" {
		_ = json.Unmarshal(tagsRaw, &item.Tags)
	}
	if len(videoCollectionsRaw) > 0 && string(videoCollectionsRaw) != "null" {
		_ = json.Unmarshal(videoCollectionsRaw, &item.VideoCollectionIDs)
	}
	if len(imageCollectionsRaw) > 0 && string(imageCollectionsRaw) != "null" {
		_ = json.Unmarshal(imageCollectionsRaw, &item.ImageCollectionIDs)
	}
	return item, nil
}

func (s *ArchiveImportService) ListGroups(ctx context.Context, batchID uuid.UUID) ([]models.ArchiveImportGroup, error) {
	rows, err := s.db.Query(ctx, archiveImportGroupSelectSQL(`
WHERE batch_id = $1
ORDER BY created_at DESC, name ASC
`), batchID)
	if err != nil {
		return nil, fmt.Errorf("list archive groups: %w", err)
	}
	defer rows.Close()

	items := make([]models.ArchiveImportGroup, 0, 16)
	for rows.Next() {
		item, scanErr := scanArchiveImportGroupRecord(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan archive group: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate archive groups: %w", err)
	}
	return items, nil
}

func (s *ArchiveImportService) GetGroup(ctx context.Context, groupID uuid.UUID) (models.ArchiveImportGroup, error) {
	row := s.db.QueryRow(ctx, archiveImportGroupSelectSQL(`WHERE id = $1`), groupID)
	item, err := scanArchiveImportGroupRecord(row)
	if err != nil {
		return models.ArchiveImportGroup{}, fmt.Errorf("get archive group: %w", err)
	}
	return item, nil
}

func (s *ArchiveImportService) CreateGroup(ctx context.Context, batchID uuid.UUID, in ArchiveImportGroupCreateInput) (models.ArchiveImportGroup, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return models.ArchiveImportGroup{}, fmt.Errorf("group name is required")
	}
	fileIDs := dedupeArchiveUUIDs(in.FileIDs)
	if len(fileIDs) == 0 {
		return models.ArchiveImportGroup{}, fmt.Errorf("at least one file is required")
	}
	batch, err := s.GetBatch(ctx, batchID)
	if err != nil {
		return models.ArchiveImportGroup{}, err
	}
	if err := s.ensureArchiveImportGroupNameUnique(ctx, batchID, name, uuid.Nil); err != nil {
		return models.ArchiveImportGroup{}, err
	}
	files, err := s.listArchiveFilesByIDs(ctx, batchID, fileIDs)
	if err != nil {
		return models.ArchiveImportGroup{}, err
	}
	if len(files) != len(fileIDs) {
		return models.ArchiveImportGroup{}, fmt.Errorf("some files were not found")
	}
	mediaKind, err := validateArchiveImportGroupSelection(files)
	if err != nil {
		return models.ArchiveImportGroup{}, err
	}
	groups, err := s.ListGroups(ctx, batchID)
	if err != nil {
		return models.ArchiveImportGroup{}, err
	}
	groupMap := indexArchiveImportGroups(groups)
	groupID := uuid.New()
	group := models.ArchiveImportGroup{
		ID:        groupID,
		BatchID:   batchID,
		Name:      name,
		Note:      strings.TrimSpace(in.Note),
		MediaKind: mediaKind,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if title := strings.TrimSpace(in.Title); title != "" {
		group.Title = &title
	}
	if description := strings.TrimSpace(in.Description); description != "" {
		group.Description = &description
	}
	if mediaKind == "video" {
		if tags := normalizeArchiveTags(in.Tags); len(tags) > 0 {
			group.Tags = tags
		}
		videoType := strings.ToLower(strings.TrimSpace(in.VideoType))
		if videoType != "" {
			if !isArchiveImportVideoType(videoType) {
				return models.ArchiveImportGroup{}, fmt.Errorf("invalid video type")
			}
			group.VideoType = &videoType
		}
		group.VideoCollectionIDs, err = s.resolveArchiveImportVideoCollectionIDs(ctx, in.VideoCollectionIDs)
		if err != nil {
			return models.ArchiveImportGroup{}, err
		}
		group.ImageCollectionIDs, err = s.resolveArchiveImportVideoImageCollectionIDs(ctx, in.ImageCollectionIDs)
		if err != nil {
			return models.ArchiveImportGroup{}, err
		}
	} else {
		group.ImageCollectionIDs, err = s.resolveArchiveImportImageCollectionIDs(ctx, in.ImageCollectionIDs)
		if err != nil {
			return models.ArchiveImportGroup{}, err
		}
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.ArchiveImportGroup{}, fmt.Errorf("begin archive group create tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
INSERT INTO archive_import_groups (
  id, batch_id, name, note, media_kind
)
VALUES ($1,$2,$3,$4,$5)
`, groupID, batchID, group.Name, group.Note, group.MediaKind); err != nil {
		return models.ArchiveImportGroup{}, fmt.Errorf("insert archive group: %w", err)
	}
	if err := updateArchiveImportGroupTx(ctx, tx, group); err != nil {
		return models.ArchiveImportGroup{}, err
	}
	for _, file := range files {
		currentGroup := archiveImportGroupPtrByID(groupMap, file.GroupID)
		overrides := archiveImportFieldOverridesForFile(file, batch, currentGroup)
		archiveImportApplyGroupToFile(&file, batch, &group, overrides)
		file.GroupID = &groupID
		if err := updateArchiveImportFileStateTx(ctx, tx, file, overrides); err != nil {
			return models.ArchiveImportGroup{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return models.ArchiveImportGroup{}, fmt.Errorf("commit archive group create tx: %w", err)
	}
	return s.GetGroup(ctx, groupID)
}

func (s *ArchiveImportService) UpdateGroup(ctx context.Context, groupID uuid.UUID, in ArchiveImportGroupUpdateInput) (models.ArchiveImportGroup, error) {
	current, err := s.GetGroup(ctx, groupID)
	if err != nil {
		return models.ArchiveImportGroup{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return models.ArchiveImportGroup{}, fmt.Errorf("group name is required")
	}
	if err := s.ensureArchiveImportGroupNameUnique(ctx, current.BatchID, name, current.ID); err != nil {
		return models.ArchiveImportGroup{}, err
	}
	batch, err := s.GetBatch(ctx, current.BatchID)
	if err != nil {
		return models.ArchiveImportGroup{}, err
	}
	next := current
	next.Name = name
	next.Note = strings.TrimSpace(in.Note)
	if current.MediaKind == "video" {
		if in.UpdateTitle {
			title := strings.TrimSpace(in.Title)
			if title == "" {
				next.Title = nil
			} else {
				next.Title = &title
			}
		}
		if in.UpdateDescription {
			description := strings.TrimSpace(in.Description)
			next.Description = &description
		}
		if in.UpdateTags {
			next.Tags = normalizeArchiveTags(in.Tags)
			if next.Tags == nil {
				next.Tags = []string{}
			}
		}
		if in.UpdateVideoType {
			videoType := strings.ToLower(strings.TrimSpace(in.VideoType))
			if videoType == "" {
				next.VideoType = nil
			} else {
				if !isArchiveImportVideoType(videoType) {
					return models.ArchiveImportGroup{}, fmt.Errorf("invalid video type")
				}
				next.VideoType = &videoType
			}
		}
		if in.UpdateVideoCollectionIDs {
			next.VideoCollectionIDs, err = s.resolveArchiveImportVideoCollectionIDs(ctx, in.VideoCollectionIDs)
			if err != nil {
				return models.ArchiveImportGroup{}, err
			}
			if next.VideoCollectionIDs == nil {
				next.VideoCollectionIDs = []uuid.UUID{}
			}
		}
		if in.UpdateImageCollectionIDs {
			next.ImageCollectionIDs, err = s.resolveArchiveImportVideoImageCollectionIDs(ctx, in.ImageCollectionIDs)
			if err != nil {
				return models.ArchiveImportGroup{}, err
			}
			if next.ImageCollectionIDs == nil {
				next.ImageCollectionIDs = []uuid.UUID{}
			}
		}
	} else {
		if in.UpdateTitle {
			title := strings.TrimSpace(in.Title)
			if title == "" {
				next.Title = nil
			} else {
				next.Title = &title
			}
		}
		if in.UpdateDescription {
			description := strings.TrimSpace(in.Description)
			next.Description = &description
		}
		if in.UpdateImageCollectionIDs {
			next.ImageCollectionIDs, err = s.resolveArchiveImportImageCollectionIDs(ctx, in.ImageCollectionIDs)
			if err != nil {
				return models.ArchiveImportGroup{}, err
			}
			if next.ImageCollectionIDs == nil {
				next.ImageCollectionIDs = []uuid.UUID{}
			}
		}
	}

	files, err := s.listArchiveFilesByGroup(ctx, groupID)
	if err != nil {
		return models.ArchiveImportGroup{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.ArchiveImportGroup{}, fmt.Errorf("begin archive group update tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := updateArchiveImportGroupTx(ctx, tx, next); err != nil {
		return models.ArchiveImportGroup{}, err
	}
	for _, file := range files {
		if isArchiveImportFrozenStatus(file.Status) {
			continue
		}
		overrides := archiveImportFieldOverridesForFile(file, batch, &current)
		archiveImportApplyGroupToFile(&file, batch, &next, overrides)
		if err := updateArchiveImportFileStateTx(ctx, tx, file, overrides); err != nil {
			return models.ArchiveImportGroup{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return models.ArchiveImportGroup{}, fmt.Errorf("commit archive group update tx: %w", err)
	}
	return s.GetGroup(ctx, groupID)
}

func (s *ArchiveImportService) DeleteGroup(ctx context.Context, groupID uuid.UUID) error {
	group, err := s.GetGroup(ctx, groupID)
	if err != nil {
		return err
	}
	batch, err := s.GetBatch(ctx, group.BatchID)
	if err != nil {
		return err
	}
	files, err := s.listArchiveFilesByGroup(ctx, groupID)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin archive group delete tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, file := range files {
		overrides := archiveImportFieldOverridesForFile(file, batch, &group)
		file.GroupID = nil
		file.GroupName = ""
		if !isArchiveImportFrozenStatus(file.Status) {
			archiveImportApplyGroupToFile(&file, batch, nil, overrides)
		}
		if err := updateArchiveImportFileStateTx(ctx, tx, file, overrides); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM archive_import_groups WHERE id=$1`, groupID); err != nil {
		return fmt.Errorf("delete archive group: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit archive group delete tx: %w", err)
	}
	return nil
}

func (s *ArchiveImportService) AssignFilesToGroup(ctx context.Context, groupID uuid.UUID, fileIDs []uuid.UUID) error {
	group, err := s.GetGroup(ctx, groupID)
	if err != nil {
		return err
	}
	fileIDs = dedupeArchiveUUIDs(fileIDs)
	if len(fileIDs) == 0 {
		return fmt.Errorf("at least one file is required")
	}
	batch, err := s.GetBatch(ctx, group.BatchID)
	if err != nil {
		return err
	}
	files, err := s.listArchiveFilesByIDs(ctx, group.BatchID, fileIDs)
	if err != nil {
		return err
	}
	if len(files) != len(fileIDs) {
		return fmt.Errorf("some files were not found")
	}
	if _, err := validateArchiveImportGroupSelection(files); err != nil {
		return err
	}
	for _, file := range files {
		if file.MediaKind != group.MediaKind {
			return fmt.Errorf("group media kind mismatch")
		}
	}
	groups, err := s.ListGroups(ctx, group.BatchID)
	if err != nil {
		return err
	}
	groupMap := indexArchiveImportGroups(groups)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin archive group assign tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, file := range files {
		currentGroup := archiveImportGroupPtrByID(groupMap, file.GroupID)
		overrides := archiveImportFieldOverridesForFile(file, batch, currentGroup)
		file.GroupID = &group.ID
		file.GroupName = group.Name
		archiveImportApplyGroupToFile(&file, batch, &group, overrides)
		if err := updateArchiveImportFileStateTx(ctx, tx, file, overrides); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit archive group assign tx: %w", err)
	}
	return nil
}

func (s *ArchiveImportService) RemoveFilesFromGroup(ctx context.Context, batchID uuid.UUID, fileIDs []uuid.UUID) error {
	fileIDs = dedupeArchiveUUIDs(fileIDs)
	if len(fileIDs) == 0 {
		return fmt.Errorf("at least one file is required")
	}
	batch, err := s.GetBatch(ctx, batchID)
	if err != nil {
		return err
	}
	files, err := s.listArchiveFilesByIDs(ctx, batchID, fileIDs)
	if err != nil {
		return err
	}
	if len(files) != len(fileIDs) {
		return fmt.Errorf("some files were not found")
	}
	for _, file := range files {
		if isArchiveImportFrozenStatus(file.Status) {
			return fmt.Errorf("processed files can no longer move between groups")
		}
	}
	groups, err := s.ListGroups(ctx, batchID)
	if err != nil {
		return err
	}
	groupMap := indexArchiveImportGroups(groups)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin archive group remove tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, file := range files {
		currentGroup := archiveImportGroupPtrByID(groupMap, file.GroupID)
		overrides := archiveImportFieldOverridesForFile(file, batch, currentGroup)
		file.GroupID = nil
		file.GroupName = ""
		archiveImportApplyGroupToFile(&file, batch, nil, overrides)
		if err := updateArchiveImportFileStateTx(ctx, tx, file, overrides); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit archive group remove tx: %w", err)
	}
	return nil
}

func (s *ArchiveImportService) ProcessGroup(ctx context.Context, groupID uuid.UUID) ([]models.ArchiveImportFileListItem, error) {
	group, err := s.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	files, err := s.listArchiveFilesByGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	var firstErr error
	for _, item := range files {
		if shouldProcessArchiveFileInBatch(item) {
			if _, err := s.ProcessFile(ctx, item.ID); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	updated, err := s.listArchiveFilesByGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if firstErr != nil && s.logger != nil {
		s.logger.Warn("archive group partial processing completed", "group_id", groupID.String(), "batch_id", group.BatchID.String(), "error", firstErr)
	}
	return updated, nil
}

func (s *ArchiveImportService) listArchiveFilesByIDs(ctx context.Context, batchID uuid.UUID, fileIDs []uuid.UUID) ([]models.ArchiveImportFileListItem, error) {
	rows, err := s.db.Query(ctx, archiveImportFileSelectSQL(`
WHERE f.batch_id = $1
  AND f.id = ANY($2)
ORDER BY f.relative_path ASC
`), batchID, fileIDs)
	if err != nil {
		return nil, fmt.Errorf("list archive files by ids: %w", err)
	}
	defer rows.Close()

	items := make([]models.ArchiveImportFileListItem, 0, len(fileIDs))
	for rows.Next() {
		item, scanErr := scanArchiveImportFileRecord(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan archive file by ids: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate archive files by ids: %w", err)
	}
	return items, nil
}

func (s *ArchiveImportService) listArchiveFilesByGroup(ctx context.Context, groupID uuid.UUID) ([]models.ArchiveImportFileListItem, error) {
	rows, err := s.db.Query(ctx, archiveImportFileSelectSQL(`
WHERE f.group_id = $1
ORDER BY f.relative_path ASC
`), groupID)
	if err != nil {
		return nil, fmt.Errorf("list archive files by group: %w", err)
	}
	defer rows.Close()

	items := make([]models.ArchiveImportFileListItem, 0, 64)
	for rows.Next() {
		item, scanErr := scanArchiveImportFileRecord(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan archive file by group: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate archive files by group: %w", err)
	}
	return items, nil
}

func (s *ArchiveImportService) ensureArchiveImportGroupNameUnique(ctx context.Context, batchID uuid.UUID, name string, excludeID uuid.UUID) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("group name is required")
	}
	var exists bool
	if excludeID == uuid.Nil {
		if err := s.db.QueryRow(ctx, `
SELECT EXISTS(
  SELECT 1
  FROM archive_import_groups
  WHERE batch_id = $1
    AND lower(name) = lower($2)
)
`, batchID, name).Scan(&exists); err != nil {
			return fmt.Errorf("check archive group name: %w", err)
		}
	} else {
		if err := s.db.QueryRow(ctx, `
SELECT EXISTS(
  SELECT 1
  FROM archive_import_groups
  WHERE batch_id = $1
    AND lower(name) = lower($2)
    AND id <> $3
)
`, batchID, name, excludeID).Scan(&exists); err != nil {
			return fmt.Errorf("check archive group name: %w", err)
		}
	}
	if exists {
		return fmt.Errorf("group name already exists in this batch")
	}
	return nil
}

func updateArchiveImportGroupTx(ctx context.Context, tx pgx.Tx, group models.ArchiveImportGroup) error {
	tagsRaw, err := marshalNullableArchiveJSON(group.Tags)
	if err != nil {
		return fmt.Errorf("marshal archive group tags: %w", err)
	}
	videoCollectionsRaw, err := marshalNullableArchiveJSON(group.VideoCollectionIDs)
	if err != nil {
		return fmt.Errorf("marshal archive group video collections: %w", err)
	}
	imageCollectionsRaw, err := marshalNullableArchiveJSON(group.ImageCollectionIDs)
	if err != nil {
		return fmt.Errorf("marshal archive group image collections: %w", err)
	}
	var title any
	if group.Title != nil {
		title = strings.TrimSpace(*group.Title)
	}
	var description any
	if group.Description != nil {
		description = strings.TrimSpace(*group.Description)
	}
	var videoType any
	if group.VideoType != nil {
		videoType = strings.ToLower(strings.TrimSpace(*group.VideoType))
	}
	if _, err := tx.Exec(ctx, `
UPDATE archive_import_groups
SET
  name=$2,
  note=$3,
  title=$4,
  description=$5,
  tags=$6,
  video_type=$7,
  video_collection_ids=$8,
  image_collection_ids=$9,
  updated_at=NOW()
WHERE id=$1
`, group.ID, group.Name, group.Note, title, description, tagsRaw, videoType, videoCollectionsRaw, imageCollectionsRaw); err != nil {
		return fmt.Errorf("update archive group: %w", err)
	}
	return nil
}

func updateArchiveImportFileStateTx(ctx context.Context, tx pgx.Tx, file models.ArchiveImportFileListItem, overrides archiveImportFieldOverrides) error {
	tagsRaw, err := json.Marshal(normalizeArchiveTags(file.Tags))
	if err != nil {
		return fmt.Errorf("marshal archive file tags: %w", err)
	}
	videoCollectionsRaw, err := json.Marshal(dedupeArchiveUUIDs(file.VideoCollectionIDs))
	if err != nil {
		return fmt.Errorf("marshal archive file video collections: %w", err)
	}
	imageCollectionsRaw, err := json.Marshal(dedupeArchiveUUIDs(file.ImageCollectionIDs))
	if err != nil {
		return fmt.Errorf("marshal archive file image collections: %w", err)
	}
	fieldOverridesRaw, err := json.Marshal(overrides.toMap())
	if err != nil {
		return fmt.Errorf("marshal archive file field overrides: %w", err)
	}
	videoType := strings.ToLower(strings.TrimSpace(file.VideoType))
	if videoType == "" {
		videoType = "short"
	}
	if file.MediaKind == "video" {
		file.ImageCollectionIDs, err = normalizeArchiveVideoImageCollectionIDs(file.ImageCollectionIDs)
		if err != nil {
			return err
		}
		imageCollectionsRaw, err = json.Marshal(file.ImageCollectionIDs)
		if err != nil {
			return fmt.Errorf("marshal archive file image collections: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `
UPDATE archive_import_files
SET
  group_id=$2,
  title=$3,
  description=$4,
  tags=$5,
  video_type=$6,
  video_collection_ids=$7,
  image_collection_ids=$8,
  field_overrides=$9,
  updated_at=NOW()
WHERE id=$1
`, file.ID, file.GroupID, strings.TrimSpace(file.Title), strings.TrimSpace(file.Description), tagsRaw, videoType, videoCollectionsRaw, imageCollectionsRaw, fieldOverridesRaw); err != nil {
		return fmt.Errorf("update archive file state: %w", err)
	}
	return nil
}

func archiveImportFieldOverridesForFile(file models.ArchiveImportFileListItem, batch models.ArchiveImportBatch, group *models.ArchiveImportGroup) archiveImportFieldOverrides {
	if len(file.FieldOverrides) > 0 {
		return archiveImportFieldOverridesFromMap(file.FieldOverrides)
	}
	defaults := resolveArchiveImportFileDefaults(file, batch, group)
	return archiveImportFieldOverrides{
		Title:              strings.TrimSpace(file.Title) != strings.TrimSpace(defaults.Title),
		Description:        strings.TrimSpace(file.Description) != strings.TrimSpace(defaults.Description),
		Tags:               !sameArchiveStringSet(file.Tags, defaults.Tags),
		VideoType:          strings.ToLower(strings.TrimSpace(file.VideoType)) != strings.ToLower(strings.TrimSpace(defaults.VideoType)),
		VideoCollectionIDs: !sameArchiveUUIDSet(dedupeArchiveUUIDs(file.VideoCollectionIDs), dedupeArchiveUUIDs(defaults.VideoCollectionIDs)),
		ImageCollectionIDs: !sameArchiveUUIDSet(dedupeArchiveUUIDs(file.ImageCollectionIDs), dedupeArchiveUUIDs(defaults.ImageCollectionIDs)),
	}
}

func archiveImportFieldOverridesFromMap(values map[string]bool) archiveImportFieldOverrides {
	return archiveImportFieldOverrides{
		Title:              values[archiveImportOverrideTitle],
		Description:        values[archiveImportOverrideDescription],
		Tags:               values[archiveImportOverrideTags],
		VideoType:          values[archiveImportOverrideVideoType],
		VideoCollectionIDs: values[archiveImportOverrideVideoCollectionIDs],
		ImageCollectionIDs: values[archiveImportOverrideImageCollectionIDs],
	}
}

func (o archiveImportFieldOverrides) toMap() map[string]bool {
	return map[string]bool{
		archiveImportOverrideTitle:              o.Title,
		archiveImportOverrideDescription:        o.Description,
		archiveImportOverrideTags:               o.Tags,
		archiveImportOverrideVideoType:          o.VideoType,
		archiveImportOverrideVideoCollectionIDs: o.VideoCollectionIDs,
		archiveImportOverrideImageCollectionIDs: o.ImageCollectionIDs,
	}
}

func resolveArchiveImportFileDefaults(file models.ArchiveImportFileListItem, batch models.ArchiveImportBatch, group *models.ArchiveImportGroup) archiveImportResolvedDefaults {
	defaults := archiveImportResolvedDefaults{
		Title:              deriveArchiveFileTitle(file.RelativePath),
		Description:        strings.TrimSpace(batch.DefaultDescription),
		Tags:               nil,
		VideoType:          "short",
		VideoCollectionIDs: nil,
		ImageCollectionIDs: nil,
	}
	switch strings.TrimSpace(file.MediaKind) {
	case "video":
		defaults.Title = archiveVideoDefaultTitle(batch, file.RelativePath)
		defaults.Tags = append([]string{}, normalizeArchiveTags(batch.DefaultTags)...)
		defaults.VideoCollectionIDs = append([]uuid.UUID{}, dedupeArchiveUUIDs(batch.DefaultVideoCollectionIDs)...)
	case "image":
		defaults.ImageCollectionIDs = append([]uuid.UUID{}, dedupeArchiveUUIDs(batch.DefaultImageCollectionIDs)...)
	}
	if group == nil {
		return defaults
	}
	if group.Title != nil {
		if value := strings.TrimSpace(*group.Title); value != "" {
			defaults.Title = value
		}
	}
	if group.Description != nil {
		defaults.Description = strings.TrimSpace(*group.Description)
	}
	if strings.TrimSpace(file.MediaKind) == "video" {
		if group.Tags != nil {
			defaults.Tags = append([]string{}, normalizeArchiveTags(group.Tags)...)
		}
		if group.VideoType != nil {
			if value := strings.ToLower(strings.TrimSpace(*group.VideoType)); value != "" {
				defaults.VideoType = value
			}
		}
		if group.VideoCollectionIDs != nil {
			defaults.VideoCollectionIDs = append([]uuid.UUID{}, dedupeArchiveUUIDs(group.VideoCollectionIDs)...)
		}
		if group.ImageCollectionIDs != nil {
			imageIDs, err := normalizeArchiveVideoImageCollectionIDs(group.ImageCollectionIDs)
			if err == nil {
				defaults.ImageCollectionIDs = append([]uuid.UUID{}, imageIDs...)
			}
		}
		return defaults
	}
	if strings.TrimSpace(file.MediaKind) == "image" && group.ImageCollectionIDs != nil {
		defaults.ImageCollectionIDs = append([]uuid.UUID{}, dedupeArchiveUUIDs(group.ImageCollectionIDs)...)
	}
	return defaults
}

func archiveImportApplyGroupToFile(file *models.ArchiveImportFileListItem, batch models.ArchiveImportBatch, group *models.ArchiveImportGroup, overrides archiveImportFieldOverrides) {
	defaults := resolveArchiveImportFileDefaults(*file, batch, group)
	if !overrides.Title {
		file.Title = defaults.Title
	}
	if !overrides.Description {
		file.Description = defaults.Description
	}
	if strings.TrimSpace(file.MediaKind) == "video" {
		if !overrides.Tags {
			file.Tags = append([]string{}, defaults.Tags...)
		}
		if !overrides.VideoType {
			file.VideoType = defaults.VideoType
		}
		if !overrides.VideoCollectionIDs {
			file.VideoCollectionIDs = append([]uuid.UUID{}, defaults.VideoCollectionIDs...)
		}
		if !overrides.ImageCollectionIDs {
			file.ImageCollectionIDs = append([]uuid.UUID{}, defaults.ImageCollectionIDs...)
		}
		return
	}
	if strings.TrimSpace(file.MediaKind) == "image" && !overrides.ImageCollectionIDs {
		file.ImageCollectionIDs = append([]uuid.UUID{}, defaults.ImageCollectionIDs...)
	}
}

func indexArchiveImportGroups(groups []models.ArchiveImportGroup) map[uuid.UUID]models.ArchiveImportGroup {
	out := make(map[uuid.UUID]models.ArchiveImportGroup, len(groups))
	for _, group := range groups {
		out[group.ID] = group
	}
	return out
}

func archiveImportGroupPtrByID(index map[uuid.UUID]models.ArchiveImportGroup, id *uuid.UUID) *models.ArchiveImportGroup {
	if id == nil {
		return nil
	}
	group, ok := index[*id]
	if !ok {
		return nil
	}
	copyGroup := group
	return &copyGroup
}

func validateArchiveImportGroupSelection(files []models.ArchiveImportFileListItem) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("at least one file is required")
	}
	mediaKind := strings.TrimSpace(files[0].MediaKind)
	if mediaKind != "video" && mediaKind != "image" {
		return "", fmt.Errorf("only video or image files can join a group")
	}
	for _, file := range files {
		if strings.TrimSpace(file.MediaKind) != mediaKind {
			return "", fmt.Errorf("group selection cannot mix video and image files")
		}
		if isArchiveImportFrozenStatus(file.Status) {
			return "", fmt.Errorf("processed files can no longer move between groups")
		}
	}
	return mediaKind, nil
}

func sameArchiveStringSet(left, right []string) bool {
	normalizedLeft := normalizeArchiveTags(left)
	normalizedRight := normalizeArchiveTags(right)
	if len(normalizedLeft) != len(normalizedRight) {
		return false
	}
	seen := make(map[string]struct{}, len(normalizedLeft))
	for _, item := range normalizedLeft {
		seen[item] = struct{}{}
	}
	for _, item := range normalizedRight {
		if _, ok := seen[item]; !ok {
			return false
		}
	}
	return true
}

func isArchiveImportFrozenStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "ready", "existing":
		return true
	default:
		return false
	}
}

func isArchiveImportVideoType(videoType string) bool {
	switch strings.ToLower(strings.TrimSpace(videoType)) {
	case "short", "movie", "episode", "av":
		return true
	default:
		return false
	}
}

func marshalNullableArchiveJSON(value any) ([]byte, error) {
	switch typed := value.(type) {
	case nil:
		return nil, nil
	case []string:
		if typed == nil {
			return nil, nil
		}
	case []uuid.UUID:
		if typed == nil {
			return nil, nil
		}
	}
	return json.Marshal(value)
}

func (s *ArchiveImportService) resolveArchiveImportVideoCollectionIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	resolved := dedupeArchiveUUIDs(ids)
	if len(resolved) == 0 || s.repo == nil {
		return resolved, nil
	}
	return s.repo.ResolveCollectionIDs(ctx, resolved)
}

func (s *ArchiveImportService) resolveArchiveImportImageCollectionIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	resolved := dedupeArchiveUUIDs(ids)
	if len(resolved) == 0 || s.repo == nil {
		return resolved, nil
	}
	return s.repo.ResolveImageCollectionIDs(ctx, resolved)
}

func (s *ArchiveImportService) resolveArchiveImportVideoImageCollectionIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	resolved, err := normalizeArchiveVideoImageCollectionIDs(ids)
	if err != nil {
		return nil, err
	}
	if len(resolved) == 0 || s.repo == nil {
		return resolved, nil
	}
	groupID, err := s.repo.ResolveVideoImageCollectionID(ctx, resolved)
	if err != nil {
		return nil, err
	}
	if groupID == nil {
		return nil, nil
	}
	return []uuid.UUID{*groupID}, nil
}
