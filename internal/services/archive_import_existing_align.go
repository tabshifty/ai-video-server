package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

const (
	archiveExistingVideoPendingDeleteReason = "复用已有视频（待删除），未关联合集"
	archiveRelationFailurePrefix            = "补齐合集关系失败"
)

// archiveExistingVideoReason 把实际新增的合集数量转换成压缩包文件行可见的说明。
// 计数口径是本次真正新写入的关系条数，不是本次选择的数量。
func archiveExistingVideoReason(added int) string {
	if added > 0 {
		return fmt.Sprintf("复用已有视频，新增 %d 个视频合集", added)
	}
	return "复用已有视频，合集无变化"
}

// applySnapshotCollections 让关系对齐使用本次作业快照里的合集配置，
// 而不是文件行里可能已经过期的旧配置。
func applySnapshotCollections(file *models.ArchiveImportFileListItem, snapshot *ArchiveWorkspaceValues) {
	if snapshot == nil {
		return
	}
	file.VideoCollectionIDs = append([]uuid.UUID{}, snapshot.VideoCollectionIDs...)
	file.ImageCollectionIDs = append([]uuid.UUID{}, snapshot.ImageCollectionIDs...)
}

// alignArchiveFileRelations 对已入库文件做一次幂等的关系对齐：把文件/分组当前配置的
// 合集关系补写到已关联的正式记录上。它不重新解包、不重建媒体、不改元数据。
func (s *ArchiveImportService) alignArchiveFileRelations(ctx context.Context, file models.ArchiveImportFileListItem, batch models.ArchiveImportBatch) error {
	if s.repo == nil {
		return nil
	}
	if file.LinkedVideoID != nil {
		video, err := s.repo.GetVideoByID(ctx, *file.LinkedVideoID)
		if errors.Is(err, pgx.ErrNoRows) {
			// 视频行可能已被清理（linked_video_id 没有外键），视为无需对齐。
			return nil
		}
		if err != nil {
			return fmt.Errorf("读取已关联视频失败: %w", err)
		}
		// 与首次导入保持同一口径：等于批次默认值的图片合集不写进视频关系。
		imageCollectionIDs, filterErr := archiveVideoImageCollectionsForProcessing(file, batch)
		if filterErr != nil {
			return filterErr
		}
		attachment, err := AttachExistingVideoCollections(ctx, s.repo, video, file.VideoCollectionIDs, firstArchiveImageCollectionID(imageCollectionIDs))
		if err != nil {
			return err
		}
		if !attachment.VideoPendingDelete && len(imageCollectionIDs) > 0 {
			if err := s.repo.AddVideoImageCollections(ctx, video.ID, imageCollectionIDs, "archive_import"); err != nil {
				return err
			}
		}
		if file.Status == "existing" {
			reason := archiveExistingVideoReason(attachment.AddedCollections)
			if attachment.VideoPendingDelete {
				reason = archiveExistingVideoPendingDeleteReason
			}
			return s.setArchiveFileReason(ctx, file.ID, reason)
		}
		if file.Status == "ready" && strings.HasPrefix(file.Reason, archiveRelationFailurePrefix) {
			// 上一次仅关系对齐失败的原因不能长期挂在已入库文件上。
			return s.setArchiveFileReason(ctx, file.ID, "")
		}
		return nil
	}
	if file.LinkedImageID != nil && len(file.ImageCollectionIDs) > 0 {
		if err := s.repo.AddImageCollectionsByIDs(ctx, *file.LinkedImageID, dedupeArchiveUUIDs(file.ImageCollectionIDs)); err != nil {
			return err
		}
	}
	return nil
}

func (s *ArchiveImportService) setArchiveFileReason(ctx context.Context, fileID uuid.UUID, reason string) error {
	_, err := s.db.Exec(ctx, `UPDATE archive_import_files SET reason=$2, updated_at=NOW() WHERE id=$1`, fileID, sanitizeArchiveText(reason))
	if err != nil {
		return fmt.Errorf("update archive file reason: %w", err)
	}
	return nil
}

func firstArchiveImageCollectionID(ids []uuid.UUID) *uuid.UUID {
	deduped := dedupeArchiveUUIDs(ids)
	if len(deduped) == 0 {
		return nil
	}
	first := deduped[0]
	return &first
}
