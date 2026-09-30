package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"video-server/internal/models"
)

// ListArchiveImportImageFileIDs returns the currently selected image files for
// a pending video entry.
func (s *ArchiveImportService) ListArchiveImportImageFileIDs(ctx context.Context, videoFileID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := s.db.Query(ctx, `
SELECT image_file_id
FROM archive_import_file_images
WHERE video_file_id=$1
ORDER BY created_at ASC, image_file_id ASC
`, videoFileID)
	if err != nil {
		return nil, fmt.Errorf("list archive import image links: %w", err)
	}
	defer rows.Close()
	ids := make([]uuid.UUID, 0, 8)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan archive import image link: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate archive import image links: %w", err)
	}
	return ids, nil
}

// UpdateArchiveImportImageFileIDs replaces the image candidates for one video
// entry. Both sides must be pending files in the same archive batch.
func (s *ArchiveImportService) UpdateArchiveImportImageFileIDs(ctx context.Context, videoFileID uuid.UUID, imageFileIDs []uuid.UUID) ([]uuid.UUID, error) {
	imageFileIDs = dedupeArchiveUUIDs(imageFileIDs)
	video, err := s.getArchiveFile(ctx, videoFileID)
	if err != nil {
		return nil, err
	}
	if video.MediaKind != "video" || video.EntryType != "file" {
		return nil, fmt.Errorf("archive relation source must be a video file")
	}
	if isArchiveImportFrozenStatus(video.Status) {
		return nil, fmt.Errorf("processed archive files can no longer change image links")
	}

	if len(imageFileIDs) > 0 {
		rows, queryErr := s.db.Query(ctx, `
SELECT id, batch_id, media_kind, entry_type, status
FROM archive_import_files
WHERE batch_id=$1 AND id = ANY($2)
`, video.BatchID, imageFileIDs)
		if queryErr != nil {
			return nil, fmt.Errorf("validate archive image links: %w", queryErr)
		}
		defer rows.Close()
		seen := make(map[uuid.UUID]struct{}, len(imageFileIDs))
		for rows.Next() {
			var id, batchID uuid.UUID
			var mediaKind, entryType, status string
			if scanErr := rows.Scan(&id, &batchID, &mediaKind, &entryType, &status); scanErr != nil {
				return nil, fmt.Errorf("scan archive image link target: %w", scanErr)
			}
			if batchID != video.BatchID || mediaKind != "image" || entryType != "file" {
				return nil, fmt.Errorf("archive image link target must be an image in the same batch")
			}
			if isArchiveImportFrozenStatus(status) {
				return nil, fmt.Errorf("processed archive image files can no longer change image links")
			}
			seen[id] = struct{}{}
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate archive image link targets: %w", err)
		}
		if len(seen) != len(imageFileIDs) {
			return nil, fmt.Errorf("some archive image link targets were not found")
		}
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin archive image link update: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM archive_import_file_images WHERE video_file_id=$1`, videoFileID); err != nil {
		return nil, fmt.Errorf("clear archive image links: %w", err)
	}
	for _, imageFileID := range imageFileIDs {
		if _, err := tx.Exec(ctx, `
INSERT INTO archive_import_file_images(video_file_id, image_file_id, source)
VALUES ($1,$2,'manual')
ON CONFLICT (video_file_id, image_file_id) DO UPDATE SET source='manual'
`, videoFileID, imageFileID); err != nil {
			return nil, fmt.Errorf("save archive image link: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit archive image link update: %w", err)
	}
	return imageFileIDs, nil
}

func (s *ArchiveImportService) syncArchiveVideoImageLinks(ctx context.Context, batchID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
INSERT INTO video_images(video_id, image_id, source)
SELECT vf.linked_video_id, im.linked_image_id, 'archive_import'
FROM archive_import_file_images rel
JOIN archive_import_files vf ON vf.id=rel.video_file_id AND vf.batch_id=$1
JOIN archive_import_files im ON im.id=rel.image_file_id AND im.batch_id=$1
WHERE vf.linked_video_id IS NOT NULL AND im.linked_image_id IS NOT NULL
ON CONFLICT (video_id, image_id) DO NOTHING
`, batchID)
	if err != nil {
		return fmt.Errorf("sync archive video image links: %w", err)
	}
	return nil
}

func archiveImportFileWithImageLinks(file models.ArchiveImportFileListItem, imageFileIDs []uuid.UUID) models.ArchiveImportFileListItem {
	file.RelatedImageFileIDs = append([]uuid.UUID{}, imageFileIDs...)
	return file
}
