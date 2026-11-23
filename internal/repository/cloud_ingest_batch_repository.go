package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"video-server/internal/models"
)

// CloudIngestBatchRepository persists batch-level ingest presets. It stores no
// execution state: member tasks and their imports remain the source of truth.
type CloudIngestBatchRepository struct {
	pool *pgxpool.Pool
}

// NewCloudIngestBatchRepository builds the repository.
func NewCloudIngestBatchRepository(pool *pgxpool.Pool) *CloudIngestBatchRepository {
	return &CloudIngestBatchRepository{pool: pool}
}

// staleImportAfter is how long a non-terminal import may sit untouched before
// the repair scan re-enqueues it.
const staleImportAfter = "30 minutes"

const cloudIngestBatchColumns = `
    id,
    title_mode,
    COALESCE(uniform_title, ''),
    video_type,
    COALESCE(site_category, ''),
    COALESCE(description, ''),
    COALESCE(tags, '{}'),
    COALESCE(collection_ids, '{}'),
    COALESCE(actor_ids, '{}'),
    COALESCE(actor_names, '{}'),
    image_collection_id,
    forced_count,
    created_by,
    created_at`

func scanCloudIngestBatch(row interface{ Scan(...any) error }) (models.CloudIngestBatch, error) {
	var batch models.CloudIngestBatch
	err := row.Scan(
		&batch.ID,
		&batch.TitleMode,
		&batch.UniformTitle,
		&batch.VideoType,
		&batch.SiteCategory,
		&batch.Description,
		&batch.Tags,
		&batch.CollectionIDs,
		&batch.ActorIDs,
		&batch.ActorNames,
		&batch.ImageCollectionID,
		&batch.ForcedCount,
		&batch.CreatedBy,
		&batch.CreatedAt,
	)
	return batch, err
}

// Create inserts one batch preset.
func (r *CloudIngestBatchRepository) Create(ctx context.Context, batch models.CloudIngestBatch) (models.CloudIngestBatch, error) {
	row := r.pool.QueryRow(ctx, `
INSERT INTO cloud_ingest_batches
    (title_mode, uniform_title, video_type, site_category, description,
     tags, collection_ids, actor_ids, actor_names, image_collection_id,
     forced_count, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING`+cloudIngestBatchColumns,
		batch.TitleMode, batch.UniformTitle, batch.VideoType, batch.SiteCategory,
		batch.Description, batch.Tags, batch.CollectionIDs, batch.ActorIDs,
		batch.ActorNames, batch.ImageCollectionID, batch.ForcedCount, batch.CreatedBy)
	created, err := scanCloudIngestBatch(row)
	if err != nil {
		return models.CloudIngestBatch{}, fmt.Errorf("insert cloud ingest batch: %w", err)
	}
	return created, nil
}

// GetByID returns one batch preset.
func (r *CloudIngestBatchRepository) GetByID(ctx context.Context, id uuid.UUID) (models.CloudIngestBatch, error) {
	row := r.pool.QueryRow(ctx, `SELECT`+cloudIngestBatchColumns+` FROM cloud_ingest_batches WHERE id = $1`, id)
	batch, err := scanCloudIngestBatch(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.CloudIngestBatch{}, models.ErrCloudIngestBatchNotFound
	}
	if err != nil {
		return models.CloudIngestBatch{}, fmt.Errorf("load cloud ingest batch: %w", err)
	}
	return batch, nil
}

// AttachTask binds a task to a batch unless another batch already claimed it.
// The bool reports whether this call wrote the binding.
func (r *CloudIngestBatchRepository) AttachTask(ctx context.Context, taskID int64, batchID uuid.UUID, title string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE cloud_download_tasks
SET ingest_batch_id = $2,
    ingest_title = NULLIF($3, ''),
    updated_at = now()
WHERE id = $1 AND ingest_batch_id IS NULL`, taskID, batchID, title)
	if err != nil {
		return false, fmt.Errorf("attach cloud ingest batch: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// Delete removes one batch preset that never received a member. The NOT EXISTS
// guard turns "errors only happen before any attach" into a SQL invariant, so a
// future code path cannot silently detach members by cleaning up a batch.
func (r *CloudIngestBatchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.pool.Exec(ctx, `
DELETE FROM cloud_ingest_batches b
WHERE b.id = $1
  AND NOT EXISTS (SELECT 1 FROM cloud_download_tasks t WHERE t.ingest_batch_id = b.id)`, id); err != nil {
		return fmt.Errorf("delete cloud ingest batch: %w", err)
	}
	return nil
}

// ListSummaries returns the most recent batches with derived member counts.
func (r *CloudIngestBatchRepository) ListSummaries(ctx context.Context, limit int) ([]models.CloudIngestBatchSummary, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
SELECT`+cloudIngestBatchColumns+`,
    COUNT(t.id) AS members,
    COUNT(*) FILTER (WHERE i.status = 'imported') AS imported,
    COUNT(*) FILTER (WHERE i.status = 'skipped_existing') AS skipped,
    COUNT(*) FILTER (WHERE i.status = 'failed') AS failed,
    COUNT(*) FILTER (WHERE i.status IN ('queued', 'hashing', 'copying', 'creating')) AS importing,
    COUNT(*) FILTER (WHERE t.id IS NOT NULL AND t.status <> 'completed') AS not_ready,
    COUNT(*) FILTER (
        WHERE t.id IS NOT NULL AND t.status = 'completed'
          AND t.local_file_deleted_at IS NULL AND i.id IS NULL
    ) AS awaiting_import
FROM cloud_ingest_batches b
LEFT JOIN cloud_download_tasks t ON t.ingest_batch_id = b.id
LEFT JOIN cloud_download_imports i ON i.task_id = t.id
GROUP BY b.id
ORDER BY b.created_at DESC
LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list cloud ingest batches: %w", err)
	}
	defer rows.Close()

	summaries := make([]models.CloudIngestBatchSummary, 0, limit)
	for rows.Next() {
		var item models.CloudIngestBatchSummary
		if err := rows.Scan(
			&item.ID,
			&item.TitleMode,
			&item.UniformTitle,
			&item.VideoType,
			&item.SiteCategory,
			&item.Description,
			&item.Tags,
			&item.CollectionIDs,
			&item.ActorIDs,
			&item.ActorNames,
			&item.ImageCollectionID,
			&item.ForcedCount,
			&item.CreatedBy,
			&item.CreatedAt,
			&item.Counts.Members,
			&item.Counts.Imported,
			&item.Counts.Skipped,
			&item.Counts.Failed,
			&item.Counts.Importing,
			&item.Counts.NotReady,
			&item.Counts.AwaitingRun,
		); err != nil {
			return nil, fmt.Errorf("scan cloud ingest batch: %w", err)
		}
		summaries = append(summaries, item)
	}
	return summaries, rows.Err()
}

// ListAwaitingAutoImport returns completed batch members whose import is
// missing or stuck, so a lost wakeup can be repaired. "Stuck" means a
// non-terminal import whose only timestamp is older than staleImportAfter;
// without it a record left in queued by a crash would never be woken again.
func (r *CloudIngestBatchRepository) ListAwaitingAutoImport(ctx context.Context, limit int) ([]int64, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
SELECT t.id
FROM cloud_download_tasks t
LEFT JOIN cloud_download_imports i ON i.task_id = t.id
WHERE t.ingest_batch_id IS NOT NULL
  AND t.status = 'completed'
  AND COALESCE(t.local_path, '') <> ''
  AND t.local_file_deleted_at IS NULL
  AND (
        i.id IS NULL
     OR (i.status IN ('queued', 'hashing', 'copying', 'creating')
         AND COALESCE(i.started_at, i.created_at) < now() - $2::interval)
  )
ORDER BY t.id
LIMIT $1`, limit, staleImportAfter)
	if err != nil {
		return nil, fmt.Errorf("list awaiting auto import: %w", err)
	}
	defer rows.Close()

	ids := make([]int64, 0, limit)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan awaiting auto import: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
