package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"video-server/internal/hashutil"
	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/pkg/ffmpeg"
)

var (
	ErrArchiveUnsupportedFormat = errors.New("unsupported archive format")
	ErrArchivePasswordRequired  = errors.New("archive password required")
	ErrArchiveNestedArchive     = errors.New("nested archive is not allowed")
	ErrArchiveBatchBusy         = errors.New("archive batch is processing")
)

type ArchiveImportService struct {
	db          *pgxpool.Pool
	uploadSvc   *UploadService
	imageSvc    *ImageService
	repo        *repository.VideoRepository
	enqueuer    archiveImportEnqueuer
	storageRoot string
	uploadTemp  string
	logger      *slog.Logger
}

type ArchiveImportUploadInput struct {
	UserID                    uuid.UUID
	Title                     string
	DefaultDescription        string
	DefaultTags               []string
	DefaultVideoCollectionIDs []uuid.UUID
	DefaultImageCollectionIDs []uuid.UUID
	HasPassword               bool
	Password                  string
}

type ArchiveImportFileUpdateInput struct {
	Title              string
	Description        string
	Tags               []string
	VideoType          string
	VideoCollectionIDs []uuid.UUID
	ImageCollectionIDs []uuid.UUID
}

type archiveImportEnqueuer interface {
	EnqueueTranscode(videoID, inputPath, outputDir, targetFormat string, force bool) error
	EnqueueScrapeMovie(videoID, filePath, filename string) error
	EnqueueScrapeTV(videoID, filePath, filename string) error
	EnqueueScrapeAV(videoID, filePath, filename string) error
}

type archiveImportFileMeta struct {
	RelativePath       string
	FilePath           string
	EntryType          string
	MediaKind          string
	VideoType          string
	FileSize           int64
	MIMEType           string
	Status             string
	Reason             string
	Title              string
	Description        string
	Tags               []string
	VideoCollectionIDs []uuid.UUID
	ImageCollectionIDs []uuid.UUID
	Metadata           map[string]any
}

type archiveImportFileRow struct {
	models.ArchiveImportFile
	TagsRaw               []byte
	VideoCollectionIDsRaw []byte
	ImageCollectionIDsRaw []byte
	MetadataRaw           []byte
}

func NewArchiveImportService(db *pgxpool.Pool, uploadSvc *UploadService, imageSvc *ImageService, repo *repository.VideoRepository, enqueuer archiveImportEnqueuer, storageRoot, uploadTemp string, logger *slog.Logger) *ArchiveImportService {
	return &ArchiveImportService{
		db:          db,
		uploadSvc:   uploadSvc,
		imageSvc:    imageSvc,
		repo:        repo,
		enqueuer:    enqueuer,
		storageRoot: storageRoot,
		uploadTemp:  uploadTemp,
		logger:      logger,
	}
}

func (s *ArchiveImportService) ListBatches(ctx context.Context, page, pageSize int) ([]models.ArchiveImportBatchListItem, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var total int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM archive_import_batches`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count archive batches: %w", err)
	}

	rows, err := s.db.Query(ctx, `
SELECT id, title, original_filename, archive_format, status, COALESCE(last_error, ''), total_entries, processable_entries, processed_entries, skipped_entries, failed_entries, created_at, updated_at, completed_at
FROM archive_import_batches
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list archive batches: %w", err)
	}
	defer rows.Close()

	items := make([]models.ArchiveImportBatchListItem, 0, pageSize)
	for rows.Next() {
		var item models.ArchiveImportBatchListItem
		if err := rows.Scan(&item.ID, &item.Title, &item.OriginalFilename, &item.ArchiveFormat, &item.Status, &item.LastError, &item.TotalEntries, &item.ProcessableEntries, &item.ProcessedEntries, &item.SkippedEntries, &item.FailedEntries, &item.CreatedAt, &item.UpdatedAt, &item.CompletedAt); err != nil {
			return nil, 0, fmt.Errorf("scan archive batch: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate archive batch: %w", err)
	}
	return items, total, nil
}

func (s *ArchiveImportService) GetBatch(ctx context.Context, batchID uuid.UUID) (models.ArchiveImportBatch, error) {
	var batch models.ArchiveImportBatch
	var defaultTagsRaw, defaultVideoCollectionsRaw, defaultImageCollectionsRaw []byte
	if err := s.db.QueryRow(ctx, `
SELECT
  id,
  user_id,
  COALESCE(title, ''),
  COALESCE(original_filename, ''),
  COALESCE(archive_format, ''),
  COALESCE(original_path, ''),
  COALESCE(extracted_dir, ''),
  COALESCE(status, ''),
  COALESCE(last_error, ''),
  COALESCE(total_entries, 0),
  COALESCE(processable_entries, 0),
  COALESCE(processed_entries, 0),
  COALESCE(skipped_entries, 0),
  COALESCE(failed_entries, 0),
  COALESCE(default_title_prefix, ''),
  COALESCE(default_description, ''),
  COALESCE(default_tags, '[]'::jsonb),
  COALESCE(default_video_collection_ids, '[]'::jsonb),
  COALESCE(default_image_collection_ids, '[]'::jsonb),
  created_at,
  updated_at,
  completed_at
FROM archive_import_batches
WHERE id = $1
`, batchID).Scan(
		&batch.ID,
		&batch.UserID,
		&batch.Title,
		&batch.OriginalFilename,
		&batch.ArchiveFormat,
		&batch.OriginalPath,
		&batch.ExtractedDir,
		&batch.Status,
		&batch.LastError,
		&batch.TotalEntries,
		&batch.ProcessableEntries,
		&batch.ProcessedEntries,
		&batch.SkippedEntries,
		&batch.FailedEntries,
		&batch.DefaultTitlePrefix,
		&batch.DefaultDescription,
		&defaultTagsRaw,
		&defaultVideoCollectionsRaw,
		&defaultImageCollectionsRaw,
		&batch.CreatedAt,
		&batch.UpdatedAt,
		&batch.CompletedAt,
	); err != nil {
		return models.ArchiveImportBatch{}, fmt.Errorf("get archive batch: %w", err)
	}
	if err := json.Unmarshal(defaultTagsRaw, &batch.DefaultTags); err != nil {
		batch.DefaultTags = nil
	}
	if err := json.Unmarshal(defaultVideoCollectionsRaw, &batch.DefaultVideoCollectionIDs); err != nil {
		batch.DefaultVideoCollectionIDs = nil
	}
	if err := json.Unmarshal(defaultImageCollectionsRaw, &batch.DefaultImageCollectionIDs); err != nil {
		batch.DefaultImageCollectionIDs = nil
	}
	return batch, nil
}

func (s *ArchiveImportService) ListFiles(ctx context.Context, batchID uuid.UUID) ([]models.ArchiveImportFileListItem, error) {
	rows, err := s.db.Query(ctx, archiveImportFileSelectSQL(`
WHERE f.batch_id = $1
ORDER BY f.relative_path ASC
`), batchID)
	if err != nil {
		return nil, fmt.Errorf("list archive files: %w", err)
	}
	defer rows.Close()

	items := make([]models.ArchiveImportFileListItem, 0, 64)
	for rows.Next() {
		item, scanErr := scanArchiveImportFileRecord(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan archive file: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate archive files: %w", err)
	}
	return items, nil
}

func (s *ArchiveImportService) UploadArchive(ctx context.Context, in ArchiveImportUploadInput, fileHeader *multipart.FileHeader) (models.ArchiveImportBatch, error) {
	if fileHeader == nil {
		return models.ArchiveImportBatch{}, fmt.Errorf("archive file is required")
	}
	format, err := detectArchiveFormat(fileHeader.Filename)
	if err != nil {
		return models.ArchiveImportBatch{}, err
	}
	if in.HasPassword && strings.TrimSpace(in.Password) == "" {
		return models.ArchiveImportBatch{}, fmt.Errorf("archive password is required")
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = strings.TrimSuffix(strings.TrimSpace(fileHeader.Filename), filepath.Ext(fileHeader.Filename))
		if title == "" {
			title = "untitled"
		}
	}
	batchID := uuid.New()
	batchRoot := filepath.Join(s.storageRoot, "archive-imports", batchID.String())
	originalDir := filepath.Join(batchRoot, "original")
	extractedDir := filepath.Join(batchRoot, "extracted")
	tempDir := filepath.Join(s.uploadTemp, "archive-imports", batchID.String())
	if err := os.MkdirAll(originalDir, 0o755); err != nil {
		return models.ArchiveImportBatch{}, fmt.Errorf("create archive original dir: %w", err)
	}
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return models.ArchiveImportBatch{}, fmt.Errorf("create archive temp dir: %w", err)
	}

	originalFilename := filepath.Base(strings.TrimSpace(fileHeader.Filename))
	if originalFilename == "." || originalFilename == string(filepath.Separator) || originalFilename == "" {
		originalFilename = batchID.String() + "." + format
	}
	tempPath := filepath.Join(tempDir, originalFilename)
	src, err := fileHeader.Open()
	if err != nil {
		return models.ArchiveImportBatch{}, fmt.Errorf("open archive upload: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(tempPath)
	if err != nil {
		return models.ArchiveImportBatch{}, fmt.Errorf("create archive temp file: %w", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		_ = os.Remove(tempPath)
		return models.ArchiveImportBatch{}, fmt.Errorf("write archive temp file: %w", err)
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(tempPath)
		return models.ArchiveImportBatch{}, fmt.Errorf("close archive temp file: %w", err)
	}

	originalPath := filepath.Join(originalDir, originalFilename)
	if err := ffmpeg.CopyFile(originalPath, tempPath); err != nil {
		_ = os.Remove(tempPath)
		return models.ArchiveImportBatch{}, fmt.Errorf("copy archive to storage: %w", err)
	}
	_ = os.Remove(tempPath)

	batch := models.ArchiveImportBatch{
		ID:                        batchID,
		UserID:                    &in.UserID,
		Title:                     title,
		OriginalFilename:          fileHeader.Filename,
		ArchiveFormat:             format,
		OriginalPath:              originalPath,
		ExtractedDir:              extractedDir,
		Status:                    "uploaded",
		LastError:                 "",
		DefaultTitlePrefix:        title,
		DefaultDescription:        strings.TrimSpace(in.DefaultDescription),
		DefaultTags:               normalizeArchiveTags(in.DefaultTags),
		DefaultVideoCollectionIDs: dedupeArchiveUUIDs(in.DefaultVideoCollectionIDs),
		DefaultImageCollectionIDs: dedupeArchiveUUIDs(in.DefaultImageCollectionIDs),
	}
	if err := s.insertArchiveBatch(ctx, batch); err != nil {
		_ = os.RemoveAll(batchRoot)
		return models.ArchiveImportBatch{}, err
	}

	if err := s.extractArchive(ctx, format, originalPath, extractedDir, passwordForInput(in)); err != nil {
		status := "failed"
		if isArchivePasswordRelatedError(err) {
			status = "needs_password"
		}
		if updateErr := s.updateArchiveBatchStatus(ctx, batchID, status, err.Error()); updateErr != nil {
			return models.ArchiveImportBatch{}, updateErr
		}
		batch.Status = status
		batch.LastError = sanitizeArchiveError(err)
		return batch, err
	}

	if err := s.recordArchiveEntries(ctx, batchID, batch); err != nil {
		_ = s.updateArchiveBatchStatus(ctx, batchID, "failed", err.Error())
		return models.ArchiveImportBatch{}, err
	}
	if err := s.refreshArchiveBatchStats(ctx, batchID); err != nil {
		_ = s.updateArchiveBatchStatus(ctx, batchID, "failed", err.Error())
		return models.ArchiveImportBatch{}, err
	}

	refreshed, err := s.GetBatch(ctx, batchID)
	if err != nil {
		return models.ArchiveImportBatch{}, err
	}
	return refreshed, nil
}

func (s *ArchiveImportService) GetBatchWithFiles(ctx context.Context, batchID uuid.UUID) (models.ArchiveImportBatch, []models.ArchiveImportFileListItem, error) {
	batch, err := s.GetBatch(ctx, batchID)
	if err != nil {
		return models.ArchiveImportBatch{}, nil, err
	}
	files, err := s.ListFiles(ctx, batchID)
	if err != nil {
		return models.ArchiveImportBatch{}, nil, err
	}
	return batch, files, nil
}

func (s *ArchiveImportService) GetFile(ctx context.Context, fileID uuid.UUID) (models.ArchiveImportFileListItem, error) {
	return s.getArchiveFile(ctx, fileID)
}

func (s *ArchiveImportService) DeleteBatch(ctx context.Context, batchID uuid.UUID) error {
	batch, err := s.GetBatch(ctx, batchID)
	if err != nil {
		return err
	}
	if err := validateArchiveBatchDeletion(batch); err != nil {
		return err
	}

	if _, err := s.db.Exec(ctx, `DELETE FROM archive_import_batches WHERE id=$1`, batchID); err != nil {
		return fmt.Errorf("delete archive batch: %w", err)
	}

	cleanupPaths := archiveBatchCleanupPaths(batch)
	for _, path := range cleanupPaths {
		if path == "" {
			continue
		}
		if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove archive batch path %s: %w", path, err)
		}
	}
	return nil
}

func (s *ArchiveImportService) UpdateFile(ctx context.Context, fileID uuid.UUID, in ArchiveImportFileUpdateInput) (models.ArchiveImportFileListItem, error) {
	file, err := s.getArchiveFile(ctx, fileID)
	if err != nil {
		return models.ArchiveImportFileListItem{}, err
	}
	batch, err := s.GetBatch(ctx, file.BatchID)
	if err != nil {
		return models.ArchiveImportFileListItem{}, err
	}
	var group *models.ArchiveImportGroup
	if file.GroupID != nil {
		groupValue, getErr := s.GetGroup(ctx, *file.GroupID)
		if getErr != nil {
			return models.ArchiveImportFileListItem{}, getErr
		}
		group = &groupValue
	}
	defaults := resolveArchiveImportFileDefaults(file, batch, group)
	overrides := archiveImportFieldOverridesForFile(file, batch, group)

	title := strings.TrimSpace(in.Title)
	if title == "" || title == strings.TrimSpace(defaults.Title) {
		file.Title = defaults.Title
		overrides.Title = false
	} else {
		file.Title = title
		overrides.Title = true
	}

	description := strings.TrimSpace(in.Description)
	if description == strings.TrimSpace(defaults.Description) {
		file.Description = defaults.Description
		overrides.Description = false
	} else {
		file.Description = description
		overrides.Description = true
	}

	tags := normalizeArchiveTags(in.Tags)
	if sameArchiveStringSet(tags, defaults.Tags) {
		file.Tags = append([]string{}, defaults.Tags...)
		overrides.Tags = false
	} else {
		file.Tags = append([]string{}, tags...)
		overrides.Tags = true
	}

	videoType := strings.ToLower(strings.TrimSpace(in.VideoType))
	if file.MediaKind != "video" {
		videoType = strings.ToLower(strings.TrimSpace(file.VideoType))
	}
	if videoType == "" {
		videoType = "short"
	}
	if file.MediaKind == "video" && strings.EqualFold(videoType, defaults.VideoType) {
		file.VideoType = defaults.VideoType
		overrides.VideoType = false
	} else {
		file.VideoType = videoType
		overrides.VideoType = file.MediaKind == "video"
	}

	videoCollections := dedupeArchiveUUIDs(in.VideoCollectionIDs)
	if file.MediaKind != "video" {
		videoCollections = nil
	}
	if sameArchiveUUIDSet(videoCollections, defaults.VideoCollectionIDs) {
		file.VideoCollectionIDs = append([]uuid.UUID{}, defaults.VideoCollectionIDs...)
		overrides.VideoCollectionIDs = false
	} else {
		file.VideoCollectionIDs = append([]uuid.UUID{}, videoCollections...)
		overrides.VideoCollectionIDs = file.MediaKind == "video"
	}

	imageCollections := dedupeArchiveUUIDs(in.ImageCollectionIDs)
	if file.MediaKind == "video" {
		imageCollections, err = normalizeArchiveVideoImageCollectionIDs(imageCollections)
		if err != nil {
			return models.ArchiveImportFileListItem{}, err
		}
	} else if file.MediaKind != "image" {
		imageCollections = nil
	}
	if sameArchiveUUIDSet(imageCollections, defaults.ImageCollectionIDs) {
		file.ImageCollectionIDs = append([]uuid.UUID{}, defaults.ImageCollectionIDs...)
		overrides.ImageCollectionIDs = false
	} else {
		file.ImageCollectionIDs = append([]uuid.UUID{}, imageCollections...)
		overrides.ImageCollectionIDs = file.MediaKind == "video" || file.MediaKind == "image"
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.ArchiveImportFileListItem{}, fmt.Errorf("begin archive file update tx: %w", err)
	}
	defer tx.Rollback(ctx)
	if err := updateArchiveImportFileStateTx(ctx, tx, file, overrides); err != nil {
		return models.ArchiveImportFileListItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.ArchiveImportFileListItem{}, fmt.Errorf("commit archive file update tx: %w", err)
	}
	return s.getArchiveFile(ctx, fileID)
}

func (s *ArchiveImportService) ProcessFile(ctx context.Context, fileID uuid.UUID) (models.ArchiveImportFileListItem, error) {
	file, err := s.getArchiveFile(ctx, fileID)
	if err != nil {
		return models.ArchiveImportFileListItem{}, err
	}
	batch, err := s.GetBatch(ctx, file.BatchID)
	if err != nil {
		return models.ArchiveImportFileListItem{}, err
	}
	if file.Status == "ready" || file.Status == "existing" || file.Status == "skipped" {
		return file, nil
	}
	if file.EntryType != "file" || (file.MediaKind != "video" && file.MediaKind != "image") {
		return models.ArchiveImportFileListItem{}, fmt.Errorf("file is not processable")
	}
	if err := s.markArchiveFileProcessing(ctx, fileID); err != nil {
		return models.ArchiveImportFileListItem{}, err
	}

	workPath, err := s.copyArchiveWorkFile(ctx, file)
	if err != nil {
		_ = s.updateArchiveFileFailure(ctx, fileID, err.Error())
		return models.ArchiveImportFileListItem{}, err
	}
	defer func() { _ = os.Remove(workPath) }()

	title := archiveFileTitleForProcessing(file, batch)
	description := strings.TrimSpace(file.Description)
	tags := append([]string{}, file.Tags...)
	if len(tags) == 0 {
		tags = []string{}
	}

	switch file.MediaKind {
	case "video":
		videoType := strings.TrimSpace(file.VideoType)
		if videoType == "" {
			videoType = "short"
		}
		archiveVideoCollections := dedupeArchiveUUIDs(file.VideoCollectionIDs)
		videoCollections := archiveVideoCollections
		if videoType != "short" {
			videoCollections = nil
		}
		imageCollectionIDs, err := archiveVideoImageCollectionsForProcessing(file, batch)
		if err != nil {
			_ = s.updateArchiveFileFailure(ctx, fileID, err.Error())
			return models.ArchiveImportFileListItem{}, err
		}
		var videoImageCollectionID *uuid.UUID
		if len(imageCollectionIDs) == 1 {
			videoImageCollectionID = &imageCollectionIDs[0]
		}
		hash, err := hashutil.SHA256(workPath)
		if err != nil {
			_ = s.updateArchiveFileFailure(ctx, fileID, err.Error())
			return models.ArchiveImportFileListItem{}, fmt.Errorf("hash archive video work file: %w", err)
		}
		userID := uuid.Nil
		if batch.UserID != nil {
			userID = *batch.UserID
		}
		archiveFilename := filepath.Base(file.RelativePath)
		resultPreview, err := s.uploadSvc.PreviewUploadedFile(ctx, LocalUploadInput{
			UserID:            userID,
			FilePath:          workPath,
			Filename:          archiveFilename,
			FileSize:          file.FileSize,
			Title:             title,
			Desc:              description,
			Type:              videoType,
			Tags:              tags,
			CollectionIDs:     videoCollections,
			ImageCollectionID: videoImageCollectionID,
			Hash:              hash,
		})
		if err != nil {
			_ = s.updateArchiveFileFailure(ctx, fileID, err.Error())
			return models.ArchiveImportFileListItem{}, err
		}
		sourcePath, err := s.copyArchiveVideoSourceFile(resultPreview.VideoID, file, workPath)
		if err != nil {
			_ = s.updateArchiveFileFailure(ctx, fileID, err.Error())
			return models.ArchiveImportFileListItem{}, err
		}
		result, err := s.uploadSvc.SaveUploadedFile(ctx, LocalUploadInput{
			UserID:            userID,
			FilePath:          sourcePath,
			Filename:          archiveFilename,
			FileSize:          file.FileSize,
			Title:             title,
			Desc:              description,
			Type:              videoType,
			Tags:              tags,
			CollectionIDs:     videoCollections,
			ImageCollectionID: videoImageCollectionID,
			Hash:              hash,
		}, 0)
		if err != nil {
			_ = s.updateArchiveFileFailure(ctx, fileID, err.Error())
			return models.ArchiveImportFileListItem{}, err
		}
		if len(archiveVideoCollections) > 0 && videoType != "short" {
			if s.repo == nil {
				_ = s.updateArchiveFileFailure(ctx, fileID, "video repository unavailable")
				return models.ArchiveImportFileListItem{}, fmt.Errorf("video repository unavailable")
			}
			if err := s.repo.AddVideoCollectionsByIDsAnyType(ctx, result.VideoID, archiveVideoCollections); err != nil {
				_ = s.updateArchiveFileFailure(ctx, fileID, err.Error())
				return models.ArchiveImportFileListItem{}, err
			}
		}
		if err := s.markArchiveVideoFileProcessed(ctx, fileID, result.VideoID, result.AlreadyExists, result.Status); err != nil {
			return models.ArchiveImportFileListItem{}, err
		}
		if !result.AlreadyExists && s.enqueuer != nil {
			if videoType == "short" {
				if err := s.enqueuer.EnqueueTranscode(result.VideoID.String(), result.InputPath, result.OutputDir, result.TargetFormat, false); err != nil {
					_ = s.updateArchiveFileFailure(ctx, fileID, err.Error())
					return models.ArchiveImportFileListItem{}, err
				}
			} else {
				var enqueueErr error
				switch videoType {
				case "movie":
					enqueueErr = s.enqueuer.EnqueueScrapeMovie(result.VideoID.String(), result.InputPath, filepath.Base(file.RelativePath))
				case "episode":
					enqueueErr = s.enqueuer.EnqueueScrapeTV(result.VideoID.String(), result.InputPath, filepath.Base(file.RelativePath))
				case "av":
					enqueueErr = s.enqueuer.EnqueueScrapeAV(result.VideoID.String(), result.InputPath, filepath.Base(file.RelativePath))
				}
				if enqueueErr != nil {
					_ = s.updateArchiveFileFailure(ctx, fileID, enqueueErr.Error())
					return models.ArchiveImportFileListItem{}, enqueueErr
				}
			}
		}
	case "image":
		userID := uuid.Nil
		if batch.UserID != nil {
			userID = *batch.UserID
		}
		result, err := s.imageSvc.saveFromLocalPath(ctx, SaveImageInput{
			UserID:        userID,
			Title:         title,
			Description:   description,
			CollectionIDs: file.ImageCollectionIDs,
		}, workPath, filepath.Base(file.RelativePath), 0)
		if err != nil {
			_ = s.updateArchiveFileFailure(ctx, fileID, err.Error())
			return models.ArchiveImportFileListItem{}, err
		}
		if result.AlreadyExists && len(file.ImageCollectionIDs) > 0 && s.repo != nil {
			if err := s.repo.AddImageCollectionsByIDs(ctx, result.ImageID, file.ImageCollectionIDs); err != nil {
				_ = s.updateArchiveFileFailure(ctx, fileID, err.Error())
				return models.ArchiveImportFileListItem{}, err
			}
		}
		if err := s.markArchiveImageFileProcessed(ctx, fileID, result.ImageID, result.AlreadyExists, result.Status); err != nil {
			return models.ArchiveImportFileListItem{}, err
		}
	}

	if err := s.refreshArchiveBatchStats(ctx, file.BatchID); err != nil {
		return models.ArchiveImportFileListItem{}, err
	}
	return s.getArchiveFile(ctx, fileID)
}

func (s *ArchiveImportService) ProcessAllFiles(ctx context.Context, batchID uuid.UUID) ([]models.ArchiveImportFileListItem, error) {
	files, err := s.ListFiles(ctx, batchID)
	if err != nil {
		return nil, err
	}
	var firstErr error
	for _, item := range files {
		if shouldProcessArchiveFileInBatch(item) {
			if _, err := s.ProcessFile(ctx, item.ID); err != nil {
				if firstErr == nil {
					firstErr = err
				}
			}
		}
	}
	updated, listErr := s.ListFiles(ctx, batchID)
	if listErr != nil {
		return nil, listErr
	}
	if firstErr != nil && s.logger != nil {
		s.logger.Warn("archive batch partial processing completed", "batch_id", batchID.String(), "error", firstErr)
	}
	return updated, nil
}

func shouldProcessArchiveFileInBatch(item models.ArchiveImportFileListItem) bool {
	if item.EntryType != "file" {
		return false
	}
	if item.MediaKind != "video" && item.MediaKind != "image" {
		return false
	}
	switch item.Status {
	case "pending", "failed", "processing":
		return true
	default:
		return false
	}
}

func normalizeArchiveVideoImageCollectionIDs(ids []uuid.UUID) ([]uuid.UUID, error) {
	out := dedupeArchiveUUIDs(ids)
	if len(out) > 1 {
		return nil, fmt.Errorf("视频仅支持关联一个图片图集")
	}
	return out, nil
}

func archiveVideoImageCollectionsForProcessing(file models.ArchiveImportFileListItem, batch models.ArchiveImportBatch) ([]uuid.UUID, error) {
	fileIDs := dedupeArchiveUUIDs(file.ImageCollectionIDs)
	if len(fileIDs) == 0 {
		return nil, nil
	}
	defaultIDs := dedupeArchiveUUIDs(batch.DefaultImageCollectionIDs)
	if sameArchiveUUIDSet(fileIDs, defaultIDs) {
		return nil, nil
	}
	return normalizeArchiveVideoImageCollectionIDs(fileIDs)
}

func sameArchiveUUIDSet(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(left))
	for _, id := range left {
		seen[id] = struct{}{}
	}
	for _, id := range right {
		if _, ok := seen[id]; !ok {
			return false
		}
	}
	return true
}

func (s *ArchiveImportService) RetryExtract(ctx context.Context, batchID uuid.UUID, password string) (models.ArchiveImportBatch, error) {
	batch, err := s.GetBatch(ctx, batchID)
	if err != nil {
		return models.ArchiveImportBatch{}, err
	}
	if batch.OriginalPath == "" || batch.ExtractedDir == "" {
		return models.ArchiveImportBatch{}, fmt.Errorf("batch paths missing")
	}
	if err := s.clearArchiveBatchFiles(ctx, batchID); err != nil {
		return models.ArchiveImportBatch{}, err
	}
	if err := os.RemoveAll(batch.ExtractedDir); err != nil && !os.IsNotExist(err) {
		return models.ArchiveImportBatch{}, fmt.Errorf("clear archive extracted dir: %w", err)
	}
	if err := s.extractArchive(ctx, batch.ArchiveFormat, batch.OriginalPath, batch.ExtractedDir, strings.TrimSpace(password)); err != nil {
		status := "failed"
		if isArchivePasswordRelatedError(err) {
			status = "needs_password"
		}
		if updateErr := s.updateArchiveBatchStatus(ctx, batchID, status, err.Error()); updateErr != nil {
			return models.ArchiveImportBatch{}, updateErr
		}
		return models.ArchiveImportBatch{}, err
	}
	if err := s.recordArchiveEntries(ctx, batchID, batch); err != nil {
		_ = s.updateArchiveBatchStatus(ctx, batchID, "failed", err.Error())
		return models.ArchiveImportBatch{}, err
	}
	if err := s.refreshArchiveBatchStats(ctx, batchID); err != nil {
		_ = s.updateArchiveBatchStatus(ctx, batchID, "failed", err.Error())
		return models.ArchiveImportBatch{}, err
	}
	return s.GetBatch(ctx, batchID)
}

func (s *ArchiveImportService) insertArchiveBatch(ctx context.Context, batch models.ArchiveImportBatch) error {
	defaultTagsRaw, err := json.Marshal(batch.DefaultTags)
	if err != nil {
		return fmt.Errorf("marshal default tags: %w", err)
	}
	defaultVideoCollectionsRaw, err := json.Marshal(batch.DefaultVideoCollectionIDs)
	if err != nil {
		return fmt.Errorf("marshal default video collections: %w", err)
	}
	defaultImageCollectionsRaw, err := json.Marshal(batch.DefaultImageCollectionIDs)
	if err != nil {
		return fmt.Errorf("marshal default image collections: %w", err)
	}
	_, err = s.db.Exec(ctx, `
INSERT INTO archive_import_batches (
  id, user_id, title, original_filename, archive_format, original_path, extracted_dir, status, last_error,
  default_title_prefix, default_description, default_tags, default_video_collection_ids, default_image_collection_ids
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
`, batch.ID, batch.UserID, batch.Title, batch.OriginalFilename, batch.ArchiveFormat, batch.OriginalPath, batch.ExtractedDir, batch.Status, batch.LastError, batch.DefaultTitlePrefix, batch.DefaultDescription, defaultTagsRaw, defaultVideoCollectionsRaw, defaultImageCollectionsRaw)
	if err != nil {
		return fmt.Errorf("insert archive batch: %w", err)
	}
	return nil
}

func (s *ArchiveImportService) updateArchiveBatchStatus(ctx context.Context, batchID uuid.UUID, status, lastError string) error {
	_, err := s.db.Exec(ctx, `
UPDATE archive_import_batches
SET status=$2, last_error=$3, updated_at=NOW()
WHERE id=$1
`, batchID, status, sanitizeArchiveText(lastError))
	if err != nil {
		return fmt.Errorf("update archive batch status: %w", err)
	}
	return nil
}

func (s *ArchiveImportService) clearArchiveBatchFiles(ctx context.Context, batchID uuid.UUID) error {
	if _, err := s.db.Exec(ctx, `DELETE FROM archive_import_files WHERE batch_id=$1`, batchID); err != nil {
		return fmt.Errorf("clear archive files: %w", err)
	}
	return nil
}

func (s *ArchiveImportService) refreshArchiveBatchStats(ctx context.Context, batchID uuid.UUID) error {
	var total, processable, processed, skipped, failed int
	if err := s.db.QueryRow(ctx, `
SELECT
  COUNT(*)::INT AS total,
  COUNT(*) FILTER (WHERE media_kind IN ('video', 'image'))::INT AS processable,
  COUNT(*) FILTER (WHERE status IN ('ready', 'existing'))::INT AS processed,
  COUNT(*) FILTER (WHERE status = 'skipped')::INT AS skipped,
  COUNT(*) FILTER (WHERE status = 'failed')::INT AS failed
FROM archive_import_files
WHERE batch_id = $1
`, batchID).Scan(&total, &processable, &processed, &skipped, &failed); err != nil {
		return fmt.Errorf("count archive file stats: %w", err)
	}
	status := "ready"
	if failed > 0 {
		status = "partial"
	}
	if skipped > 0 && processed == 0 && failed == 0 {
		status = "ready"
	}
	if processable > 0 && processed >= processable && failed == 0 {
		status = "completed"
	}
	if processable == 0 && total > 0 {
		status = "ready"
	}
	var completedAt any
	if status == "completed" {
		completedAt = time.Now().UTC()
	}
	_, err := s.db.Exec(ctx, `
UPDATE archive_import_batches
SET
  total_entries=$2,
  processable_entries=$3,
  processed_entries=$4,
  skipped_entries=$5,
  failed_entries=$6,
  status=$7,
  completed_at=$8,
  updated_at=NOW()
WHERE id=$1
`, batchID, total, processable, processed, skipped, failed, status, completedAt)
	if err != nil {
		return fmt.Errorf("update archive batch stats: %w", err)
	}
	return nil
}

func (s *ArchiveImportService) recordArchiveEntries(ctx context.Context, batchID uuid.UUID, batch models.ArchiveImportBatch) error {
	entries, err := scanArchiveEntries(batch.ExtractedDir, batch)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return s.refreshArchiveBatchStats(ctx, batchID)
	}
	values := make([]any, 0, len(entries)*19)
	placeholders := make([]string, 0, len(entries))
	for idx, entry := range entries {
		tagsRaw, err := json.Marshal(entry.Tags)
		if err != nil {
			return fmt.Errorf("marshal archive file tags: %w", err)
		}
		videoCollectionsRaw, err := json.Marshal(entry.VideoCollectionIDs)
		if err != nil {
			return fmt.Errorf("marshal archive file video collections: %w", err)
		}
		imageCollectionsRaw, err := json.Marshal(entry.ImageCollectionIDs)
		if err != nil {
			return fmt.Errorf("marshal archive file image collections: %w", err)
		}
		metadataRaw, err := json.Marshal(entry.Metadata)
		if err != nil {
			return fmt.Errorf("marshal archive file metadata: %w", err)
		}
		base := idx*19 + 1
		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)", base, base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10, base+11, base+12, base+13, base+14, base+15, base+16, base+17, base+18))
		values = append(values,
			uuid.New(),
			batchID,
			entry.RelativePath,
			entry.FilePath,
			entry.EntryType,
			entry.MediaKind,
			entry.VideoType,
			entry.FileSize,
			entry.MIMEType,
			entry.Status,
			entry.Reason,
			entry.Title,
			entry.Description,
			tagsRaw,
			videoCollectionsRaw,
			imageCollectionsRaw,
			nil,
			nil,
			metadataRaw,
		)
	}
	sql := `
INSERT INTO archive_import_files (
  id, batch_id, relative_path, file_path, entry_type, media_kind, video_type, file_size, mime_type, status,
  reason, title, description, tags, video_collection_ids, image_collection_ids, linked_video_id, linked_image_id, metadata
)
VALUES ` + strings.Join(placeholders, ",")
	if _, err := s.db.Exec(ctx, sql, values...); err != nil {
		return fmt.Errorf("insert archive file rows: %w", err)
	}
	return s.refreshArchiveBatchStats(ctx, batchID)
}

func scanArchiveEntries(extractedDir string, batch models.ArchiveImportBatch) ([]archiveImportFileMeta, error) {
	extractedDir = strings.TrimSpace(extractedDir)
	if extractedDir == "" {
		return nil, nil
	}
	items := make([]archiveImportFileMeta, 0, 64)
	err := filepath.WalkDir(extractedDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(extractedDir, path)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		if d.IsDir() {
			items = append(items, archiveImportFileMeta{
				RelativePath: rel,
				FilePath:     path,
				EntryType:    "directory",
				MediaKind:    "directory",
				Status:       "skipped",
				Reason:       "directory",
				Title:        deriveArchiveFileTitle(rel),
			})
			return nil
		}
		mediaKind, mimeType, videoType, reason := classifyArchiveFile(path)
		status := "pending"
		if mediaKind == "archive" {
			status = "skipped"
			reason = "nested_archive_not_allowed"
		} else if mediaKind == "other" {
			status = "skipped"
			if reason == "" {
				reason = "unsupported_file_type"
			}
		}
		item := archiveImportFileMeta{
			RelativePath:       rel,
			FilePath:           path,
			EntryType:          "file",
			MediaKind:          mediaKind,
			VideoType:          videoType,
			FileSize:           info.Size(),
			MIMEType:           mimeType,
			Status:             status,
			Reason:             reason,
			Title:              archiveFileTitleForScannedFile(mediaKind, rel, batch),
			Description:        batch.DefaultDescription,
			Tags:               append([]string{}, batch.DefaultTags...),
			VideoCollectionIDs: append([]uuid.UUID{}, batch.DefaultVideoCollectionIDs...),
			ImageCollectionIDs: archiveImageCollectionsForScannedFile(mediaKind, batch.DefaultImageCollectionIDs),
			Metadata: map[string]any{
				"original_filename": filepath.Base(rel),
				"relative_path":     rel,
			},
		}
		if mediaKind == "video" {
			item.MediaKind = "video"
		}
		if mediaKind == "image" {
			item.MediaKind = "image"
		}
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan extracted archive entries: %w", err)
	}
	return items, nil
}

func archiveImageCollectionsForScannedFile(mediaKind string, defaultImageCollectionIDs []uuid.UUID) []uuid.UUID {
	if strings.TrimSpace(mediaKind) != "image" {
		return nil
	}
	return append([]uuid.UUID{}, defaultImageCollectionIDs...)
}

func archiveFileTitleForScannedFile(mediaKind, relativePath string, batch models.ArchiveImportBatch) string {
	if strings.TrimSpace(mediaKind) == "video" {
		return archiveVideoDefaultTitle(batch, relativePath)
	}
	return deriveArchiveFileTitle(relativePath)
}

func archiveFileTitleForProcessing(file models.ArchiveImportFileListItem, batch models.ArchiveImportBatch) string {
	title := strings.TrimSpace(file.Title)
	if title != "" {
		return title
	}
	if strings.TrimSpace(file.MediaKind) == "video" {
		return archiveVideoDefaultTitle(batch, file.RelativePath)
	}
	return deriveArchiveFileTitle(file.RelativePath)
}

func archiveFileTitleForUpdate(inputTitle string, file models.ArchiveImportFileListItem, batchTitle string) string {
	title := strings.TrimSpace(inputTitle)
	if title != "" {
		return title
	}
	if strings.TrimSpace(file.MediaKind) == "video" {
		if title = strings.TrimSpace(batchTitle); title != "" {
			return title
		}
		return deriveArchiveFileTitle(file.RelativePath)
	}
	return file.Title
}

func archiveVideoDefaultTitle(batch models.ArchiveImportBatch, relativePath string) string {
	if title := strings.TrimSpace(batch.DefaultTitlePrefix); title != "" {
		return title
	}
	if title := strings.TrimSpace(batch.Title); title != "" {
		return title
	}
	return deriveArchiveFileTitle(relativePath)
}

func archiveBatchCleanupPaths(batch models.ArchiveImportBatch) []string {
	seen := map[string]struct{}{}
	paths := make([]string, 0, 4)
	appendPath := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}

	appendPath(filepath.Dir(strings.TrimSpace(batch.OriginalPath)))
	appendPath(strings.TrimSpace(batch.ExtractedDir))
	return paths
}

func validateArchiveBatchDeletion(batch models.ArchiveImportBatch) error {
	if strings.EqualFold(strings.TrimSpace(batch.Status), "processing") {
		return ErrArchiveBatchBusy
	}
	return nil
}

func classifyArchiveFile(path string) (mediaKind, mimeType, videoType, reason string) {
	videoType = "short"
	if strings.TrimSpace(path) == "" {
		return "other", "", videoType, "empty_path"
	}
	ext := strings.ToLower(filepath.Ext(path))
	if isAllowedArchiveVideoExt(path) {
		return "video", detectArchiveMIME(path), videoType, ""
	}
	detected, err := mimetype.DetectFile(path)
	if err == nil {
		mimeType = strings.ToLower(strings.TrimSpace(detected.String()))
	}
	if mimeType == "" {
		mimeType = detectArchiveMIME(path)
	}
	if isAllowedImageMIME(mimeType) {
		return "image", mimeType, videoType, ""
	}
	if isArchiveExt(ext) || strings.Contains(mimeType, "zip") || strings.Contains(mimeType, "rar") || strings.Contains(mimeType, "7z") {
		return "archive", mimeType, videoType, ""
	}
	return "other", mimeType, videoType, ""
}

func isAllowedArchiveVideoExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(name)))
	switch ext {
	case ".mp4", ".mov", ".mkv", ".avi", ".wmv", ".flv", ".webm", ".m4v":
		return true
	default:
		return false
	}
}

func detectArchiveMIME(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".zip":
		return "application/zip"
	case ".rar":
		return "application/x-rar-compressed"
	case ".7z":
		return "application/x-7z-compressed"
	default:
		return ""
	}
}

func isArchiveExt(ext string) bool {
	switch strings.ToLower(strings.TrimSpace(ext)) {
	case ".zip", ".rar", ".7z":
		return true
	default:
		return false
	}
}

func deriveArchiveFileTitle(relativePath string) string {
	relativePath = filepath.ToSlash(strings.TrimSpace(relativePath))
	if relativePath == "" {
		return "untitled"
	}
	base := filepath.Base(relativePath)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.Join(strings.Fields(base), " ")
	if base == "" {
		return "untitled"
	}
	return base
}

func normalizeArchiveTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	out := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		value := strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(tag)), " "))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func dedupeArchiveUUIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return nil
	}
	out := make([]uuid.UUID, 0, len(ids))
	seen := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func passwordForInput(in ArchiveImportUploadInput) string {
	if !in.HasPassword {
		return ""
	}
	return strings.TrimSpace(in.Password)
}

func (s *ArchiveImportService) copyArchiveWorkFile(ctx context.Context, file models.ArchiveImportFileListItem) (string, error) {
	workDir := filepath.Join(s.uploadTemp, "archive-imports", file.BatchID.String(), "work")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return "", fmt.Errorf("create archive work dir: %w", err)
	}
	target := filepath.Join(workDir, filepath.FromSlash(file.RelativePath))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", fmt.Errorf("create archive work parent dir: %w", err)
	}
	if err := ffmpeg.CopyFile(target, file.FilePath); err != nil {
		return "", fmt.Errorf("copy archive work file: %w", err)
	}
	return target, nil
}

func (s *ArchiveImportService) copyArchiveVideoSourceFile(videoID uuid.UUID, file models.ArchiveImportFileListItem, workPath string) (string, error) {
	sourceDir := filepath.Join(s.storageRoot, "videos", videoID.String())
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		return "", fmt.Errorf("create archive video source dir: %w", err)
	}
	target := filepath.Join(sourceDir, archiveVideoSourceFilename(file))
	if err := ffmpeg.CopyFile(target, workPath); err != nil {
		return "", fmt.Errorf("copy archive video source file: %w", err)
	}
	return target, nil
}

func archiveVideoSourceFilename(file models.ArchiveImportFileListItem) string {
	ext := strings.TrimSpace(filepath.Ext(strings.TrimSpace(file.RelativePath)))
	if ext == "" {
		ext = strings.TrimSpace(filepath.Ext(strings.TrimSpace(file.FilePath)))
	}
	return "source-original" + ext
}

func (s *ArchiveImportService) markArchiveFileProcessing(ctx context.Context, fileID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
UPDATE archive_import_files
SET status='processing', reason='', updated_at=NOW()
WHERE id=$1
`, fileID)
	if err != nil {
		return fmt.Errorf("update archive file processing: %w", err)
	}
	return nil
}

func (s *ArchiveImportService) updateArchiveFileFailure(ctx context.Context, fileID uuid.UUID, reason string) error {
	_, err := s.db.Exec(ctx, `
UPDATE archive_import_files
SET status='failed', reason=$2, processed_at=NOW(), updated_at=NOW()
WHERE id=$1
`, fileID, sanitizeArchiveText(reason))
	if err != nil {
		return fmt.Errorf("update archive file failure: %w", err)
	}
	return s.refreshArchiveBatchStats(ctx, s.mustArchiveFileBatchID(ctx, fileID))
}

func (s *ArchiveImportService) markArchiveVideoFileProcessed(ctx context.Context, fileID, videoID uuid.UUID, alreadyExists bool, status string) error {
	fileStatus := "ready"
	if alreadyExists {
		fileStatus = "existing"
	}
	if strings.TrimSpace(status) == "failed" {
		fileStatus = "failed"
	}
	var linked any = videoID
	_, err := s.db.Exec(ctx, `
UPDATE archive_import_files
SET status=$2, reason=$3, linked_video_id=$4, processed_at=NOW(), updated_at=NOW()
WHERE id=$1
`, fileID, fileStatus, "", linked)
	if err != nil {
		return fmt.Errorf("update archive video file: %w", err)
	}
	return nil
}

func (s *ArchiveImportService) markArchiveImageFileProcessed(ctx context.Context, fileID, imageID uuid.UUID, alreadyExists bool, status string) error {
	fileStatus := "ready"
	if alreadyExists {
		fileStatus = "existing"
	}
	if strings.TrimSpace(status) == "failed" {
		fileStatus = "failed"
	}
	var linked any = imageID
	_, err := s.db.Exec(ctx, `
UPDATE archive_import_files
SET status=$2, reason=$3, linked_image_id=$4, processed_at=NOW(), updated_at=NOW()
WHERE id=$1
`, fileID, fileStatus, "", linked)
	if err != nil {
		return fmt.Errorf("update archive image file: %w", err)
	}
	return nil
}

func (s *ArchiveImportService) mustArchiveFileBatchID(ctx context.Context, fileID uuid.UUID) uuid.UUID {
	var batchID uuid.UUID
	_ = s.db.QueryRow(ctx, `SELECT batch_id FROM archive_import_files WHERE id=$1`, fileID).Scan(&batchID)
	return batchID
}

func (s *ArchiveImportService) getArchiveFile(ctx context.Context, fileID uuid.UUID) (models.ArchiveImportFileListItem, error) {
	rows, err := s.db.Query(ctx, archiveImportFileSelectSQL(`WHERE f.id = $1`), fileID)
	if err != nil {
		return models.ArchiveImportFileListItem{}, fmt.Errorf("query archive file: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return models.ArchiveImportFileListItem{}, fmt.Errorf("archive file not found")
	}
	item, err := scanArchiveImportFileRecord(rows)
	if err != nil {
		return models.ArchiveImportFileListItem{}, fmt.Errorf("scan archive file: %w", err)
	}
	return item, rows.Err()
}

func isArchivePasswordRelatedError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "password") || strings.Contains(msg, "passphrase") || strings.Contains(msg, "encrypted")
}

func sanitizeArchiveText(value string) string {
	return strings.TrimSpace(strings.ToValidUTF8(value, "\uFFFD"))
}

func sanitizeArchiveError(err error) string {
	if err == nil {
		return ""
	}
	return sanitizeArchiveText(err.Error())
}

func detectArchiveFormat(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	switch ext {
	case ".zip":
		return "zip", nil
	case ".rar":
		return "rar", nil
	case ".7z":
		return "7z", nil
	default:
		return "", ErrArchiveUnsupportedFormat
	}
}

func (s *ArchiveImportService) extractArchive(ctx context.Context, format, archivePath, extractedDir, password string) error {
	if err := os.RemoveAll(extractedDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear archive extraction dir: %w", err)
	}
	if err := os.MkdirAll(extractedDir, 0o755); err != nil {
		return fmt.Errorf("create archive extraction dir: %w", err)
	}

	switch strings.ToLower(strings.TrimSpace(format)) {
	case "zip":
		if strings.TrimSpace(password) != "" {
			return s.extractWithUnzip(ctx, archivePath, extractedDir, password)
		}
		return extractZipArchive(archivePath, extractedDir)
	case "rar", "7z":
		if strings.TrimSpace(password) != "" {
			if err := s.extractWithSevenZip(ctx, archivePath, extractedDir, password); err == nil {
				return nil
			} else if !errors.Is(err, exec.ErrNotFound) {
				return err
			}
		}
		return s.extractWithBsdtar(ctx, archivePath, extractedDir)
	default:
		return ErrArchiveUnsupportedFormat
	}
}

func extractZipArchive(archivePath, extractedDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			targetDir, err := safeArchiveTargetPath(extractedDir, file.Name)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(targetDir, 0o755); err != nil {
				return fmt.Errorf("create zip dir: %w", err)
			}
			continue
		}
		targetPath, err := safeArchiveTargetPath(extractedDir, file.Name)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return fmt.Errorf("create zip parent dir: %w", err)
		}
		rc, err := file.Open()
		if err != nil {
			if isArchivePasswordRelatedError(err) {
				return ErrArchivePasswordRequired
			}
			return fmt.Errorf("open zip entry: %w", err)
		}
		out, err := os.Create(targetPath)
		if err != nil {
			rc.Close()
			return fmt.Errorf("create zip entry file: %w", err)
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return fmt.Errorf("extract zip entry: %w", err)
		}
		if err := out.Close(); err != nil {
			rc.Close()
			return fmt.Errorf("close zip entry file: %w", err)
		}
		if err := rc.Close(); err != nil {
			return fmt.Errorf("close zip entry: %w", err)
		}
	}
	return nil
}

func safeArchiveTargetPath(baseDir, name string) (string, error) {
	cleaned := filepath.Clean(strings.TrimSpace(name))
	if cleaned == "." || cleaned == string(filepath.Separator) || strings.HasPrefix(cleaned, "..") {
		return "", ErrArchiveNestedArchive
	}
	target := filepath.Join(baseDir, cleaned)
	rel, err := filepath.Rel(baseDir, target)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") {
		return "", ErrArchiveNestedArchive
	}
	return target, nil
}

func (s *ArchiveImportService) extractWithBsdtar(ctx context.Context, archivePath, extractedDir string) error {
	cmd := exec.CommandContext(ctx, "bsdtar", "-xf", archivePath, "-C", extractedDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		output := sanitizeArchiveText(stderr.String())
		if output != "" {
			err = fmt.Errorf("%w: %s", err, output)
		}
		if isArchivePasswordRelatedError(err) {
			return ErrArchivePasswordRequired
		}
		return fmt.Errorf("extract archive with bsdtar: %w", err)
	}
	return nil
}

func (s *ArchiveImportService) extractWithUnzip(ctx context.Context, archivePath, extractedDir, password string) error {
	args := []string{"-o", "-q", "-P", password, archivePath, "-d", extractedDir}
	cmd := exec.CommandContext(ctx, "unzip", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		output := sanitizeArchiveText(stderr.String())
		if output != "" {
			err = fmt.Errorf("%w: %s", err, output)
		}
		if isArchivePasswordRelatedError(err) {
			return ErrArchivePasswordRequired
		}
		return fmt.Errorf("extract archive with unzip: %w", err)
	}
	return nil
}

func (s *ArchiveImportService) extractWithSevenZip(ctx context.Context, archivePath, extractedDir, password string) error {
	exe := findArchiveExtractorCommand("7zz", "7z")
	if exe == "" {
		return exec.ErrNotFound
	}
	args := []string{"x", "-y", "-o" + extractedDir}
	if strings.TrimSpace(password) != "" {
		args = append(args, "-p"+password)
	}
	args = append(args, archivePath)
	cmd := exec.CommandContext(ctx, exe, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		output := sanitizeArchiveText(stderr.String())
		if output != "" {
			err = fmt.Errorf("%w: %s", err, output)
		}
		if isArchivePasswordRelatedError(err) {
			return ErrArchivePasswordRequired
		}
		return fmt.Errorf("extract archive with %s: %w", exe, err)
	}
	return nil
}

func findArchiveExtractorCommand(names ...string) string {
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil && strings.TrimSpace(path) != "" {
			return path
		}
	}
	return ""
}
