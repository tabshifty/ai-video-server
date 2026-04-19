package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

func normalizeImageCollectionName(raw string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(raw)), " "))
}

func normalizeImageCollectionInput(in models.AdminImageCollectionInput) (models.AdminImageCollectionInput, error) {
	in.Name = strings.Join(strings.Fields(strings.TrimSpace(in.Name)), " ")
	if in.Name == "" {
		return in, fmt.Errorf("合集名称不能为空")
	}
	in.Description = strings.TrimSpace(in.Description)
	in.CoverURL = strings.TrimSpace(in.CoverURL)
	return in, nil
}

func scanAdminImageCollection(rows rowScanner) (models.AdminImageCollection, error) {
	var out models.AdminImageCollection
	if err := rows.Scan(
		&out.ID,
		&out.Name,
		&out.Description,
		&out.CoverURL,
		&out.SortOrder,
		&out.Active,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return models.AdminImageCollection{}, err
	}
	return out, nil
}

func (r *VideoRepository) ListImageCollections(ctx context.Context, q string, active *bool, page, pageSize int) ([]models.AdminImageCollection, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 6)
	next := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	keyword := strings.ToLower(strings.TrimSpace(q))
	if keyword != "" {
		pattern := "%" + keyword + "%"
		where = append(where, "(LOWER(name) LIKE "+next(pattern)+" OR LOWER(COALESCE(description,'')) LIKE "+next(pattern)+")")
	}
	if active != nil {
		where = append(where, "active = "+next(*active))
	}

	baseWhere := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM collections_images WHERE "+baseWhere, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count image collections: %w", err)
	}

	args = append(args, pageSize, (page-1)*pageSize)
	sql := "SELECT id, name, COALESCE(description,''), COALESCE(cover_url,''), sort_order, active, created_at, updated_at FROM collections_images WHERE " +
		baseWhere + " ORDER BY sort_order DESC, updated_at DESC, name ASC LIMIT $" + fmt.Sprintf("%d", len(args)-1) +
		" OFFSET $" + fmt.Sprintf("%d", len(args))
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list image collections: %w", err)
	}
	defer rows.Close()

	items := make([]models.AdminImageCollection, 0, pageSize)
	for rows.Next() {
		item, scanErr := scanAdminImageCollection(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("scan image collection: %w", scanErr)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *VideoRepository) CreateImageCollection(ctx context.Context, input models.AdminImageCollectionInput) (models.AdminImageCollection, error) {
	input, err := normalizeImageCollectionInput(input)
	if err != nil {
		return models.AdminImageCollection{}, err
	}

	row := r.pool.QueryRow(ctx, `
INSERT INTO collections_images (
  id, name, normalized_name, description, cover_url, sort_order, active
)
VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING id, name, COALESCE(description,''), COALESCE(cover_url,''), sort_order, active, created_at, updated_at
`, uuid.New(), input.Name, normalizeImageCollectionName(input.Name), input.Description, input.CoverURL, input.SortOrder, input.Active)
	out, err := scanAdminImageCollection(row)
	if err != nil {
		return models.AdminImageCollection{}, fmt.Errorf("create image collection: %w", err)
	}
	return out, nil
}

func (r *VideoRepository) UpdateImageCollection(ctx context.Context, collectionID uuid.UUID, input models.AdminImageCollectionInput) (models.AdminImageCollection, error) {
	input, err := normalizeImageCollectionInput(input)
	if err != nil {
		return models.AdminImageCollection{}, err
	}

	row := r.pool.QueryRow(ctx, `
UPDATE collections_images
SET
  name=$2,
  normalized_name=$3,
  description=$4,
  cover_url=$5,
  sort_order=$6,
  active=$7,
  updated_at=NOW()
WHERE id=$1
RETURNING id, name, COALESCE(description,''), COALESCE(cover_url,''), sort_order, active, created_at, updated_at
`, collectionID, input.Name, normalizeImageCollectionName(input.Name), input.Description, input.CoverURL, input.SortOrder, input.Active)
	out, err := scanAdminImageCollection(row)
	if err != nil {
		return models.AdminImageCollection{}, fmt.Errorf("update image collection: %w", err)
	}
	return out, nil
}

func (r *VideoRepository) DeleteImageCollection(ctx context.Context, collectionID uuid.UUID) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx delete image collection: %w", err)
	}
	defer tx.Rollback(ctx)

	var detached int64
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM image_collections WHERE collection_id=$1`, collectionID).Scan(&detached); err != nil {
		return 0, fmt.Errorf("count image collection relations: %w", err)
	}

	tag, err := tx.Exec(ctx, `DELETE FROM collections_images WHERE id=$1`, collectionID)
	if err != nil {
		return 0, fmt.Errorf("delete image collection: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return 0, pgx.ErrNoRows
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit delete image collection: %w", err)
	}
	return detached, nil
}

func dedupeImageCollectionIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return nil
	}
	out := make([]uuid.UUID, 0, len(ids))
	seen := map[uuid.UUID]struct{}{}
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func normalizeSingleImageCollectionID(ids []uuid.UUID) (*uuid.UUID, error) {
	deduped := dedupeImageCollectionIDs(ids)
	if len(deduped) == 0 {
		return nil, nil
	}
	if len(deduped) > 1 {
		return nil, fmt.Errorf("视频仅支持关联一个图片合集")
	}
	id := deduped[0]
	return &id, nil
}

func (r *VideoRepository) ResolveVideoImageCollectionID(ctx context.Context, collectionIDs []uuid.UUID) (*uuid.UUID, error) {
	imageCollectionID, err := normalizeSingleImageCollectionID(collectionIDs)
	if err != nil {
		return nil, err
	}
	if imageCollectionID == nil {
		return nil, nil
	}

	var existingID uuid.UUID
	if err := r.pool.QueryRow(ctx, `SELECT id FROM collections_images WHERE id=$1`, *imageCollectionID).Scan(&existingID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("图片合集不存在: %s", imageCollectionID.String())
		}
		return nil, fmt.Errorf("query image collection by id: %w", err)
	}
	return &existingID, nil
}

func (r *VideoRepository) ResolveImageCollectionIDs(ctx context.Context, collectionIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(collectionIDs) == 0 {
		return nil, nil
	}
	deduped := dedupeImageCollectionIDs(collectionIDs)

	rows, err := r.pool.Query(ctx, `SELECT id FROM collections_images WHERE id = ANY($1)`, deduped)
	if err != nil {
		return nil, fmt.Errorf("query image collections: %w", err)
	}
	defer rows.Close()

	exists := map[uuid.UUID]struct{}{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan image collection id: %w", err)
		}
		exists[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]uuid.UUID, 0, len(deduped))
	for _, id := range deduped {
		if _, ok := exists[id]; !ok {
			return nil, fmt.Errorf("图片合集不存在: %s", id.String())
		}
		out = append(out, id)
	}
	return out, nil
}

func (r *VideoRepository) ReplaceImageCollectionsByIDs(ctx context.Context, imageID uuid.UUID, collectionIDs []uuid.UUID) error {
	resolvedIDs, err := r.ResolveImageCollectionIDs(ctx, collectionIDs)
	if err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx replace image collections: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM image_collections WHERE image_id=$1`, imageID); err != nil {
		return fmt.Errorf("clear image collections: %w", err)
	}
	for _, collectionID := range resolvedIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO image_collections(image_id, collection_id) VALUES ($1,$2)`, imageID, collectionID); err != nil {
			return fmt.Errorf("insert image collection: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit replace image collections: %w", err)
	}
	return nil
}

func (r *VideoRepository) ListImageCollectionsForAdmin(ctx context.Context, imageID uuid.UUID) ([]models.AdminImageCollection, error) {
	rows, err := r.pool.Query(ctx, `
SELECT c.id, c.name, COALESCE(c.description,''), COALESCE(c.cover_url,''), c.sort_order, c.active, c.created_at, c.updated_at
FROM image_collections ic
JOIN collections_images c ON c.id = ic.collection_id
WHERE ic.image_id=$1
ORDER BY c.sort_order DESC, c.updated_at DESC, c.name ASC
`, imageID)
	if err != nil {
		return nil, fmt.Errorf("list image collections for admin: %w", err)
	}
	defer rows.Close()

	out := make([]models.AdminImageCollection, 0, 8)
	for rows.Next() {
		item, err := scanAdminImageCollection(rows)
		if err != nil {
			return nil, fmt.Errorf("scan image collection for admin: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
