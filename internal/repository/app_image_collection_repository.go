package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/utils"
)

const (
	appImageCollectionCoverWidth   = 480
	appImageCollectionCoverHeight  = 640
	appImageCollectionCoverQuality = 78
	appImageThumbnailSize          = 160
	appImageThumbnailQuality       = 72
)

func resolveAppImageCollectionCoverURL(imageID *uuid.UUID, fallbackCoverURL string) string {
	if imageID != nil {
		return utils.AppImageViewURL(*imageID, appImageCollectionCoverWidth, appImageCollectionCoverHeight, "cover", appImageCollectionCoverQuality)
	}
	return strings.TrimSpace(fallbackCoverURL)
}

func scanAppVideoImageCollection(rows rowScanner) (*models.VideoImageCollection, error) {
	var out models.VideoImageCollection
	var coverImageID string
	var createdAt, updatedAt any
	if err := rows.Scan(
		&out.ID,
		&out.Name,
		&out.CoverURL,
		&coverImageID,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, err
	}
	out.CoverURL = resolveAppImageCollectionCoverURL(parseNullableUUIDText(coverImageID), out.CoverURL)
	return &out, nil
}

func scanAppImageCollectionListItem(rows rowScanner) (models.ImageCollectionListItem, error) {
	var out models.ImageCollectionListItem
	var coverImageID string
	if err := rows.Scan(
		&out.ID,
		&out.Name,
		&out.Description,
		&out.CoverURL,
		&coverImageID,
		&out.ImageCount,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return models.ImageCollectionListItem{}, err
	}
	out.CoverURL = resolveAppImageCollectionCoverURL(parseNullableUUIDText(coverImageID), out.CoverURL)
	return out, nil
}

func scanAppImageCollectionImage(rows rowScanner) (models.ImageCollectionImage, error) {
	var out models.ImageCollectionImage
	if err := rows.Scan(
		&out.ID,
		&out.Title,
		&out.Description,
		&out.Width,
		&out.Height,
	); err != nil {
		return models.ImageCollectionImage{}, err
	}
	out.ThumbnailURL = utils.AppImageViewURL(out.ID, appImageThumbnailSize, appImageThumbnailSize, "cover", appImageThumbnailQuality)
	out.ViewURL = utils.AppImageViewURL(out.ID, 0, 0, "", 0)
	return out, nil
}

func (r *VideoRepository) ListAppImageCollections(ctx context.Context, limit, offset int) ([]models.ImageCollectionListItem, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM collections_images c
WHERE c.active = TRUE
  AND EXISTS (
    SELECT 1
    FROM image_collections ic
    JOIN images i ON i.id = ic.image_id
    WHERE ic.collection_id = c.id
      AND i.active = TRUE
      AND i.status = 'ready'
  )
`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count app image collections: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
SELECT
  c.id,
  c.name,
  COALESCE(c.description, ''),
  COALESCE(c.cover_url, ''),
  COALESCE(preview.cover_image_id::text, ''),
  preview.image_count,
  c.created_at,
  c.updated_at
FROM collections_images c
JOIN LATERAL (
  SELECT
    COUNT(*)::INT AS image_count,
    (
      ARRAY_AGG(
        ic.image_id
        ORDER BY
          CASE WHEN c.cover_image_id IS NOT NULL AND ic.image_id = c.cover_image_id THEN 0 ELSE 1 END,
          ic.created_at ASC,
          ic.image_id ASC
      )
    )[1] AS cover_image_id
  FROM image_collections ic
  JOIN images i ON i.id = ic.image_id
  WHERE ic.collection_id = c.id
    AND i.active = TRUE
    AND i.status = 'ready'
) preview ON preview.image_count > 0
WHERE c.active = TRUE
ORDER BY c.sort_order DESC, c.updated_at DESC, c.name ASC
LIMIT $1 OFFSET $2
`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list app image collections: %w", err)
	}
	defer rows.Close()

	items := make([]models.ImageCollectionListItem, 0, limit)
	for rows.Next() {
		item, scanErr := scanAppImageCollectionListItem(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("scan app image collection: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate app image collections: %w", err)
	}
	return items, total, nil
}

func (r *VideoRepository) GetAppImageCollectionDetail(ctx context.Context, collectionID uuid.UUID) (models.ImageCollectionDetail, error) {
	var detail models.ImageCollectionDetail
	var coverImageID string
	if err := r.pool.QueryRow(ctx, `
SELECT
  c.id,
  c.name,
  COALESCE(c.description, ''),
  COALESCE(c.cover_url, ''),
  COALESCE(preview.cover_image_id::text, ''),
  preview.image_count,
  c.created_at,
  c.updated_at
FROM collections_images c
JOIN LATERAL (
  SELECT
    COUNT(*)::INT AS image_count,
    (
      ARRAY_AGG(
        ic.image_id
        ORDER BY
          CASE WHEN c.cover_image_id IS NOT NULL AND ic.image_id = c.cover_image_id THEN 0 ELSE 1 END,
          ic.created_at ASC,
          ic.image_id ASC
      )
    )[1] AS cover_image_id
  FROM image_collections ic
  JOIN images i ON i.id = ic.image_id
  WHERE ic.collection_id = c.id
    AND i.active = TRUE
    AND i.status = 'ready'
) preview ON preview.image_count > 0
WHERE c.id = $1
  AND c.active = TRUE
`, collectionID).Scan(
		&detail.ID,
		&detail.Name,
		&detail.Description,
		&detail.CoverURL,
		&coverImageID,
		&detail.ImageCount,
		&detail.CreatedAt,
		&detail.UpdatedAt,
	); err != nil {
		return models.ImageCollectionDetail{}, fmt.Errorf("get app image collection detail: %w", err)
	}
	detail.CoverURL = resolveAppImageCollectionCoverURL(parseNullableUUIDText(coverImageID), detail.CoverURL)

	rows, err := r.pool.Query(ctx, `
SELECT
  i.id,
  COALESCE(i.title, ''),
  COALESCE(i.description, ''),
  COALESCE(i.width, 0),
  COALESCE(i.height, 0)
FROM image_collections ic
JOIN images i ON i.id = ic.image_id
JOIN collections_images c ON c.id = ic.collection_id
WHERE ic.collection_id = $1
  AND c.active = TRUE
  AND i.active = TRUE
  AND i.status = 'ready'
ORDER BY
  CASE WHEN c.cover_image_id IS NOT NULL AND i.id = c.cover_image_id THEN 0 ELSE 1 END,
  ic.created_at ASC,
  i.id ASC
`, collectionID)
	if err != nil {
		return models.ImageCollectionDetail{}, fmt.Errorf("list app image collection images: %w", err)
	}
	defer rows.Close()

	images := make([]models.ImageCollectionImage, 0, detail.ImageCount)
	for rows.Next() {
		item, scanErr := scanAppImageCollectionImage(rows)
		if scanErr != nil {
			return models.ImageCollectionDetail{}, fmt.Errorf("scan app image collection image: %w", scanErr)
		}
		images = append(images, item)
	}
	if err := rows.Err(); err != nil {
		return models.ImageCollectionDetail{}, fmt.Errorf("iterate app image collection images: %w", err)
	}
	detail.Images = images
	return detail, nil
}

func (r *VideoRepository) GetAppVideoImageCollection(ctx context.Context, collectionID uuid.UUID) (*models.VideoImageCollection, error) {
	row := r.pool.QueryRow(ctx, `
SELECT
  c.id,
  c.name,
  COALESCE(c.cover_url, ''),
  COALESCE(preview.cover_image_id::text, ''),
  c.created_at,
  c.updated_at
FROM collections_images c
JOIN LATERAL (
  SELECT
    COUNT(*)::INT AS image_count,
    (
      ARRAY_AGG(
        ic.image_id
        ORDER BY
          CASE WHEN c.cover_image_id IS NOT NULL AND ic.image_id = c.cover_image_id THEN 0 ELSE 1 END,
          ic.created_at ASC,
          ic.image_id ASC
      )
    )[1] AS cover_image_id
  FROM image_collections ic
  JOIN images i ON i.id = ic.image_id
  WHERE ic.collection_id = c.id
    AND i.active = TRUE
    AND i.status = 'ready'
) preview ON preview.image_count > 0
WHERE c.id = $1
  AND c.active = TRUE
`, collectionID)
	item, err := scanAppVideoImageCollection(row)
	if err != nil {
		return nil, fmt.Errorf("get app video image collection: %w", err)
	}
	return item, nil
}
