package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"video-server/internal/models"
)

// ExistingVideoCollectionRepo 是「已存在视频合集关系追加」所需的最小持久化面。
// *repository.VideoRepository 实现该接口；测试可用假实现替代。
type ExistingVideoCollectionRepo interface {
	AddVideoCollectionsByIDsAnyTypeReturningAdded(ctx context.Context, videoID uuid.UUID, collectionIDs []uuid.UUID) (int, error)
	FillVideoImageCollectionIfEmpty(ctx context.Context, videoID uuid.UUID, imageCollectionID uuid.UUID) (bool, error)
}

// ExistingVideoCollectionAttachment 描述一次对已存在视频的合集关系追加结果。
type ExistingVideoCollectionAttachment struct {
	// AddedCollections 是本次实际新增的视频合集关系条数，不是本次选择的数量。
	AddedCollections int
	// AddedImageCollection 表示本次补齐了该视频此前缺失的图片合集。
	AddedImageCollection bool
	// VideoPendingDelete 表示目标视频正在等待删除，因此没有写入任何关系。
	VideoPendingDelete bool
}

// AttachExistingVideoCollections 把本次入库携带的合集关系追加到已存在视频上。
// 语义是幂等并集：只新增、不删除、不覆盖既有关系；视频处于待删除状态时不写入。
func AttachExistingVideoCollections(
	ctx context.Context,
	repo ExistingVideoCollectionRepo,
	video models.Video,
	collectionIDs []uuid.UUID,
	imageCollectionID *uuid.UUID,
) (ExistingVideoCollectionAttachment, error) {
	result := ExistingVideoCollectionAttachment{}
	if video.Status == "pending_delete" {
		result.VideoPendingDelete = true
		return result, nil
	}
	ids := dedupeExistingCollectionIDs(collectionIDs)
	if len(ids) > 0 {
		added, err := repo.AddVideoCollectionsByIDsAnyTypeReturningAdded(ctx, video.ID, ids)
		if err != nil {
			return result, fmt.Errorf("append existing video collections: %w", err)
		}
		result.AddedCollections = added
	}
	if imageCollectionID != nil && video.ImageCollectionID == nil && *imageCollectionID != uuid.Nil {
		filled, err := repo.FillVideoImageCollectionIfEmpty(ctx, video.ID, *imageCollectionID)
		if err != nil {
			return result, fmt.Errorf("fill existing video image collection: %w", err)
		}
		result.AddedImageCollection = filled
	}
	return result, nil
}

func dedupeExistingCollectionIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
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
