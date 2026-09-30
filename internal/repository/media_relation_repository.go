package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *VideoRepository) ReplaceVideoImageCollections(ctx context.Context, videoID uuid.UUID, collectionIDs []uuid.UUID, source string) error {
	collectionIDs, err := r.ResolveImageCollectionIDs(ctx, collectionIDs)
	if err != nil {
		return err
	}
	if source == "" {
		source = "direct"
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin replace video image collections: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM video_image_collections WHERE video_id=$1`, videoID); err != nil {
		return fmt.Errorf("clear video image collections: %w", err)
	}
	for _, collectionID := range collectionIDs {
		if _, err := tx.Exec(ctx, `
INSERT INTO video_image_collections(video_id, image_collection_id, source)
VALUES ($1,$2,$3)
ON CONFLICT (video_id, image_collection_id) DO UPDATE SET source=EXCLUDED.source
`, videoID, collectionID, source); err != nil {
			return fmt.Errorf("insert video image collection: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit video image collections: %w", err)
	}
	return nil
}

func (r *VideoRepository) ListVideoImageCollectionIDs(ctx context.Context, videoID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
SELECT image_collection_id
FROM video_image_collections
WHERE video_id=$1
ORDER BY created_at ASC, image_collection_id ASC
`, videoID)
	if err != nil {
		return nil, fmt.Errorf("list video image collections: %w", err)
	}
	defer rows.Close()
	ids := make([]uuid.UUID, 0, 4)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan video image collection: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *VideoRepository) ReplaceVideoCollectionImageCollections(ctx context.Context, videoCollectionID uuid.UUID, imageCollectionIDs []uuid.UUID) error {
	imageCollectionIDs, err := r.ResolveImageCollectionIDs(ctx, imageCollectionIDs)
	if err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin replace collection image collections: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM video_collection_image_collections WHERE video_collection_id=$1`, videoCollectionID); err != nil {
		return fmt.Errorf("clear collection image collections: %w", err)
	}
	for _, imageCollectionID := range imageCollectionIDs {
		if _, err := tx.Exec(ctx, `
INSERT INTO video_collection_image_collections(video_collection_id, image_collection_id)
VALUES ($1,$2)
ON CONFLICT DO NOTHING
`, videoCollectionID, imageCollectionID); err != nil {
			return fmt.Errorf("insert collection image collection: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit collection image collections: %w", err)
	}
	return nil
}

func (r *VideoRepository) ListInheritedImageCollectionIDs(ctx context.Context, videoID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
SELECT DISTINCT vci.image_collection_id
FROM video_collections vc
JOIN video_collection_image_collections vci ON vci.video_collection_id=vc.collection_id
WHERE vc.video_id=$1
ORDER BY vci.image_collection_id
`, videoID)
	if err != nil {
		return nil, fmt.Errorf("list inherited image collections: %w", err)
	}
	defer rows.Close()
	ids := make([]uuid.UUID, 0, 4)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan inherited image collection: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *VideoRepository) ReplaceVideoImages(ctx context.Context, videoID uuid.UUID, imageIDs []uuid.UUID, source string) error {
	imageIDs = dedupeMediaUUIDs(imageIDs)
	if source == "" {
		source = "manual"
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin replace video images: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM video_images WHERE video_id=$1`, videoID); err != nil {
		return fmt.Errorf("clear video images: %w", err)
	}
	for _, imageID := range imageIDs {
		if _, err := tx.Exec(ctx, `
INSERT INTO video_images(video_id, image_id, source)
VALUES ($1,$2,$3)
ON CONFLICT (video_id, image_id) DO UPDATE SET source=EXCLUDED.source
`, videoID, imageID, source); err != nil {
			return fmt.Errorf("insert video image: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit video images: %w", err)
	}
	return nil
}

func dedupeMediaUUIDs(ids []uuid.UUID) []uuid.UUID {
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
