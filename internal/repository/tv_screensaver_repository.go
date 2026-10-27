package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/utils"
)

// ListTVScreensaverImages 读取可用正式图片，沿用客户端图库的 active/ready 可见性。
// 图片是全库共享资产，不按上传者或合集过滤；直接查询 images 避免跨合集重复和漏掉未归合集图片。
func (r *VideoRepository) ListTVScreensaverImages(ctx context.Context, after uuid.UUID, limit int) (models.TVScreensaverImages, error) {
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, updated_at
FROM images
WHERE active = TRUE AND status = 'ready'
  AND COALESCE(stored_path, '') <> ''
  AND id > $1
ORDER BY id ASC
LIMIT $2`, after, limit+1)
	if err != nil {
		return models.TVScreensaverImages{}, fmt.Errorf("list TV screensaver images: %w", err)
	}
	defer rows.Close()

	result := models.TVScreensaverImages{Items: make([]models.TVScreensaverImage, 0, limit)}
	for rows.Next() {
		var item models.TVScreensaverImage
		if err := rows.Scan(&item.ID, &item.UpdatedAt); err != nil {
			return models.TVScreensaverImages{}, fmt.Errorf("scan TV screensaver image: %w", err)
		}
		if len(result.Items) == limit {
			result.NextCursor = result.Items[limit-1].ID.String()
			break
		}
		item.ViewURL = utils.AppImageViewURL(item.ID, 0, 0, "", 0)
		result.Items = append(result.Items, item)
	}
	if err := rows.Err(); err != nil {
		return models.TVScreensaverImages{}, fmt.Errorf("iterate TV screensaver images: %w", err)
	}
	return result, nil
}
