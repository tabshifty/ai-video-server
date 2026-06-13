package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

func (r *VideoRepository) FindImageByHash(ctx context.Context, hash string, fileSize int64) (uuid.UUID, bool, error) {
	var imageID uuid.UUID
	err := r.pool.QueryRow(ctx, `
SELECT image_id
FROM image_hashes
WHERE hash=$1 AND file_size=$2
`, hash, fileSize).Scan(&imageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, fmt.Errorf("find image by hash: %w", err)
	}
	return imageID, true, nil
}

func (r *VideoRepository) InsertImageHash(ctx context.Context, hash string, imageID uuid.UUID, fileSize int64) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO image_hashes(hash, file_size, image_id)
VALUES ($1,$2,$3)
`, hash, fileSize, imageID)
	if err != nil {
		return fmt.Errorf("insert image hash: %w", err)
	}
	return nil
}

func (r *VideoRepository) CreateImage(ctx context.Context, img models.Image) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO images (
  id, user_id, title, description, status, active,
  original_path, stored_path, original_mime, stored_mime, original_ext, stored_ext,
  file_size, width, height, metadata
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
`,
		img.ID,
		img.UserID,
		img.Title,
		img.Description,
		img.Status,
		img.Active,
		img.OriginalPath,
		img.StoredPath,
		img.OriginalMIME,
		img.StoredMIME,
		img.OriginalExt,
		img.StoredExt,
		img.FileSize,
		img.Width,
		img.Height,
		img.Metadata,
	)
	if err != nil {
		return fmt.Errorf("insert image: %w", err)
	}
	return nil
}

func (r *VideoRepository) GetImageByID(ctx context.Context, imageID uuid.UUID) (models.Image, error) {
	var out models.Image
	err := r.pool.QueryRow(ctx, `
SELECT
  id,
  user_id,
  COALESCE(title, ''),
  COALESCE(description, ''),
  status,
  active,
  COALESCE(original_path, ''),
  COALESCE(stored_path, ''),
  COALESCE(original_mime, ''),
  COALESCE(stored_mime, ''),
  COALESCE(original_ext, ''),
  COALESCE(stored_ext, ''),
  COALESCE(file_size, 0),
  COALESCE(width, 0),
  COALESCE(height, 0),
  COALESCE(metadata, '{}'::jsonb),
  created_at,
  updated_at
FROM images
WHERE id=$1
`, imageID).Scan(
		&out.ID,
		&out.UserID,
		&out.Title,
		&out.Description,
		&out.Status,
		&out.Active,
		&out.OriginalPath,
		&out.StoredPath,
		&out.OriginalMIME,
		&out.StoredMIME,
		&out.OriginalExt,
		&out.StoredExt,
		&out.FileSize,
		&out.Width,
		&out.Height,
		&out.Metadata,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if err != nil {
		return models.Image{}, fmt.Errorf("get image by id: %w", err)
	}
	return out, nil
}

func (r *VideoRepository) DeleteImageByID(ctx context.Context, imageID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM images WHERE id=$1`, imageID)
	if err != nil {
		return fmt.Errorf("delete image by id: %w", err)
	}
	return nil
}

func (r *VideoRepository) ListImageVariantPaths(ctx context.Context, imageID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT variant_path FROM image_variants WHERE image_id=$1`, imageID)
	if err != nil {
		return nil, fmt.Errorf("list image variant paths: %w", err)
	}
	defer rows.Close()
	out := make([]string, 0, 8)
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("scan image variant path: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *VideoRepository) GetImageVariantByKey(ctx context.Context, imageID uuid.UUID, key string) (path, mime string, width, height int, exists bool, err error) {
	err = r.pool.QueryRow(ctx, `
SELECT variant_path, COALESCE(mime, ''), COALESCE(width, 0), COALESCE(height, 0)
FROM image_variants
WHERE image_id=$1 AND transform_key=$2
`, imageID, key).Scan(&path, &mime, &width, &height)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", 0, 0, false, nil
		}
		return "", "", 0, 0, false, fmt.Errorf("get image variant by key: %w", err)
	}
	return path, mime, width, height, true, nil
}

func (r *VideoRepository) UpsertImageVariant(ctx context.Context, imageID uuid.UUID, key, path, mime string, width, height int) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO image_variants(image_id, transform_key, variant_path, mime, width, height)
VALUES ($1,$2,$3,$4,$5,$6)
ON CONFLICT(image_id, transform_key)
DO UPDATE SET
  variant_path=EXCLUDED.variant_path,
  mime=EXCLUDED.mime,
  width=EXCLUDED.width,
  height=EXCLUDED.height,
  updated_at=NOW()
`, imageID, key, path, mime, width, height)
	if err != nil {
		return fmt.Errorf("upsert image variant: %w", err)
	}
	return nil
}

func (r *VideoRepository) AdminListImages(ctx context.Context, f models.AdminImageFilter) ([]models.AdminImageListItem, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 12)
	next := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if q := strings.TrimSpace(strings.ToLower(f.Keyword)); q != "" {
		like := "%" + q + "%"
		p := next(like)
		where = append(where, "(LOWER(i.title) LIKE "+p+" OR LOWER(COALESCE(i.description,'')) LIKE "+p+")")
	}
	if s := strings.TrimSpace(strings.ToLower(f.Status)); s != "" {
		where = append(where, "i.status="+next(s))
	}
	if f.Active != nil {
		where = append(where, "i.active="+next(*f.Active))
	}
	if len(f.StoredMIMEs) > 0 {
		mimeArgs := make([]string, 0, len(f.StoredMIMEs))
		for _, mime := range f.StoredMIMEs {
			normalized := strings.TrimSpace(strings.ToLower(mime))
			if normalized == "" {
				continue
			}
			mimeArgs = append(mimeArgs, next(normalized))
		}
		if len(mimeArgs) > 0 {
			where = append(where, "LOWER(COALESCE(i.stored_mime,'')) IN ("+strings.Join(mimeArgs, ",")+")")
		}
	}
	if f.ActorID != nil {
		where = append(where, "EXISTS (SELECT 1 FROM image_actors ia WHERE ia.image_id=i.id AND ia.actor_id="+next(*f.ActorID)+")")
	}
	if f.CollectionID != nil {
		where = append(where, "EXISTS (SELECT 1 FROM image_collections ic WHERE ic.image_id=i.id AND ic.collection_id="+next(*f.CollectionID)+")")
	}

	baseWhere := strings.Join(where, " AND ")
	var total int
	countSQL := "SELECT COUNT(*) FROM images i LEFT JOIN users u ON u.id=i.user_id WHERE " + baseWhere
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("admin count images: %w", err)
	}

	args = append(args, f.PageSize, (f.Page-1)*f.PageSize)
	listSQL := "SELECT i.id, i.title, COALESCE(i.description,''), i.status, i.active, COALESCE(i.stored_path,''), COALESCE(i.stored_mime,''), COALESCE(i.file_size,0), COALESCE(i.width,0), COALESCE(i.height,0), COALESCE(u.username,''), i.user_id, i.created_at, i.updated_at FROM images i LEFT JOIN users u ON u.id=i.user_id WHERE " + baseWhere + " ORDER BY i.created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)-1) + " OFFSET $" + fmt.Sprintf("%d", len(args))
	rows, err := r.pool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("admin list images: %w", err)
	}
	defer rows.Close()

	items := make([]models.AdminImageListItem, 0, f.PageSize)
	for rows.Next() {
		var item models.AdminImageListItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Status, &item.Active, &item.StoredPath, &item.StoredMIME, &item.FileSize, &item.Width, &item.Height, &item.UploadUser, &item.UploadUserID, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan admin image: %w", err)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *VideoRepository) ListImageActors(ctx context.Context, imageID uuid.UUID) ([]models.AdminImageActor, error) {
	rows, err := r.pool.Query(ctx, `
SELECT a.id, a.name, COALESCE(a.avatar_url,''), a.active, COALESCE(ia.source,'')
FROM image_actors ia
JOIN actors a ON a.id = ia.actor_id
WHERE ia.image_id=$1
ORDER BY ia.created_at ASC, a.name ASC
`, imageID)
	if err != nil {
		return nil, fmt.Errorf("list image actors: %w", err)
	}
	defer rows.Close()

	out := make([]models.AdminImageActor, 0, 8)
	for rows.Next() {
		var item models.AdminImageActor
		if err := rows.Scan(&item.ID, &item.Name, &item.AvatarURL, &item.Active, &item.BindSource); err != nil {
			return nil, fmt.Errorf("scan image actor: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *VideoRepository) AdminImageDetail(ctx context.Context, imageID uuid.UUID) (models.AdminImageDetail, error) {
	var out models.AdminImageDetail
	var metadata []byte
	err := r.pool.QueryRow(ctx, `
SELECT
  id, user_id, COALESCE(title,''), COALESCE(description,''), status, active,
  COALESCE(original_path,''), COALESCE(stored_path,''), COALESCE(original_mime,''), COALESCE(stored_mime,''),
  COALESCE(original_ext,''), COALESCE(stored_ext,''), COALESCE(file_size,0), COALESCE(width,0), COALESCE(height,0),
  COALESCE(metadata, '{}'::jsonb), created_at, updated_at
FROM images WHERE id=$1
`, imageID).Scan(
		&out.ID,
		&out.UserID,
		&out.Title,
		&out.Description,
		&out.Status,
		&out.Active,
		&out.OriginalPath,
		&out.StoredPath,
		&out.OriginalMIME,
		&out.StoredMIME,
		&out.OriginalExt,
		&out.StoredExt,
		&out.FileSize,
		&out.Width,
		&out.Height,
		&metadata,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if err != nil {
		return models.AdminImageDetail{}, fmt.Errorf("admin image detail: %w", err)
	}
	if len(metadata) == 0 {
		metadata = []byte(`{}`)
	}
	out.Metadata = metadata

	actors, err := r.ListImageActors(ctx, imageID)
	if err != nil {
		return models.AdminImageDetail{}, fmt.Errorf("admin image actors: %w", err)
	}
	out.Actors = actors
	collections, err := r.ListImageCollectionsForAdmin(ctx, imageID)
	if err != nil {
		return models.AdminImageDetail{}, fmt.Errorf("admin image collections: %w", err)
	}
	out.Collections = collections
	return out, nil
}

func (r *VideoRepository) AdminUpdateImage(ctx context.Context, imageID uuid.UUID, title, description string, active bool, metadata map[string]any, actorIDs []uuid.UUID, actorNames []string, updateActors bool, collectionIDs []uuid.UUID, updateCollections bool) error {
	raw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal admin image metadata: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
UPDATE images
SET title=$2, description=$3, active=$4, metadata=$5, updated_at=NOW()
WHERE id=$1
`, imageID, title, description, active, raw)
	if err != nil {
		return fmt.Errorf("admin update image: %w", err)
	}

	if updateActors {
		if err := r.ReplaceImageActorsByInput(ctx, imageID, actorIDs, actorNames, "admin_edit"); err != nil {
			return fmt.Errorf("replace image actors: %w", err)
		}
	}
	if updateCollections {
		if err := r.ReplaceImageCollectionsByIDs(ctx, imageID, collectionIDs); err != nil {
			return fmt.Errorf("replace image collections: %w", err)
		}
	}
	return nil
}

func (r *VideoRepository) DeleteImageByIDCascade(ctx context.Context, imageID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM images WHERE id=$1`, imageID)
	if err != nil {
		return fmt.Errorf("delete image by id cascade: %w", err)
	}
	return nil
}

func boolPointerFromQuery(raw string) (*bool, error) {
	v := strings.TrimSpace(strings.ToLower(raw))
	if v == "" {
		return nil, nil
	}
	if v == "1" || v == "true" || v == "yes" {
		b := true
		return &b, nil
	}
	if v == "0" || v == "false" || v == "no" {
		b := false
		return &b, nil
	}
	return nil, fmt.Errorf("invalid bool value")
}

func normalizeImageTitle(filename, title string) string {
	v := strings.TrimSpace(title)
	if v != "" {
		return v
	}
	fallback := strings.TrimSuffix(strings.TrimSpace(filename), filepathExt(filename))
	if fallback == "" {
		return "untitled"
	}
	return fallback
}

func filepathExt(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[i:]
		}
	}
	return ""
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
