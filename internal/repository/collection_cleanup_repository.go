package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CollectionCleanupResult counts only the confirmed preview candidates.
type CollectionCleanupResult struct {
	DeletedCount int `json:"deleted_count"`
	SkippedCount int `json:"skipped_count"`
}

// LockCollectionReferenceFiles coordinates cleanup with new disk upload sessions.
// A dedicated connection avoids holding a pooled query connection while scanning.
func (r *VideoRepository) LockCollectionReferenceFiles(ctx context.Context) (func(), error) {
	conn, err := pgx.ConnectConfig(ctx, r.pool.Config().ConnConfig.Copy())
	if err != nil {
		return nil, err
	}
	var locked bool
	if err = conn.QueryRow(ctx, "SELECT pg_try_advisory_lock(49350, 1)").Scan(&locked); err != nil || !locked {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = conn.Close(closeCtx)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("合集清理或上传初始化正在进行，请稍后重试")
	}
	return func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// 关闭专用连接即释放该会话的 advisory lock。
		_ = conn.Close(closeCtx)
	}, nil
}

func emptyCollectionQuery(kind string) (table, predicate string, err error) {
	switch kind {
	case "video":
		table = "collections"
		predicate = `NOT EXISTS (SELECT 1 FROM video_collections r WHERE r.collection_id=c.id)
   AND NOT EXISTS (SELECT 1 FROM video_collection_image_collections r WHERE r.video_collection_id=c.id)`
	case "image":
		table = "collections_images"
		predicate = `NOT EXISTS (SELECT 1 FROM image_collections r WHERE r.collection_id=c.id)
   AND NOT EXISTS (SELECT 1 FROM videos v WHERE v.image_collection_id=c.id)
   AND NOT EXISTS (SELECT 1 FROM video_image_collections r WHERE r.image_collection_id=c.id)
   AND NOT EXISTS (SELECT 1 FROM video_collection_image_collections r WHERE r.image_collection_id=c.id)`
	default:
		return "", "", fmt.Errorf("未知合集类型")
	}
	predicate += ` AND NOT EXISTS (SELECT 1 FROM config_refs r WHERE r.collection_id=c.id::text)`
	return table, predicate, nil
}

const cleanupConfigRefsSQL = `WITH config_refs AS MATERIALIZED (
 SELECT DISTINCT collection_id FROM collection_config_references WHERE kind=$1
) `

// PreviewEmptyCollections returns all eligible collections, including disabled ones.
func (r *VideoRepository) PreviewEmptyCollections(ctx context.Context, kind string) ([]uuid.UUID, error) {
	table, predicate, err := emptyCollectionQuery(kind)
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, cleanupConfigRefsSQL+"SELECT c.id FROM "+table+" c WHERE "+predicate+" ORDER BY c.id", kind)
	if err != nil {
		return nil, fmt.Errorf("预览空合集: %w", err)
	}
	defer rows.Close()
	ids, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		return nil, fmt.Errorf("读取空合集候选: %w", err)
	}
	return ids, nil
}

// CleanupEmptyCollections locks preview candidates before rechecking references.
// Foreign keys and JSON reference guards share-lock collection rows on new writes.
func (r *VideoRepository) CleanupEmptyCollections(ctx context.Context, kind string, candidates []uuid.UUID) (CollectionCleanupResult, error) {
	table, predicate, err := emptyCollectionQuery(kind)
	if err != nil {
		return CollectionCleanupResult{}, err
	}
	ids := dedupeCollectionIDs(candidates)
	result := CollectionCleanupResult{SkippedCount: len(ids)}
	if len(ids) == 0 {
		return result, nil
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return CollectionCleanupResult{}, fmt.Errorf("开始清理事务: %w", err)
	}
	defer tx.Rollback(ctx)
	// 锁等待有界，失败整笔回滚。按 ID 锁定使重复清理具有一致的锁顺序。
	if _, err = tx.Exec(ctx, `SET LOCAL lock_timeout = '5s'`); err != nil {
		return CollectionCleanupResult{}, err
	}
	rows, err := tx.Query(ctx, "SELECT id FROM "+table+" WHERE id=ANY($1::uuid[]) ORDER BY id FOR UPDATE", ids)
	if err != nil {
		return CollectionCleanupResult{}, fmt.Errorf("锁定空合集: %w", err)
	}
	_, err = pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		return CollectionCleanupResult{}, fmt.Errorf("读取锁定合集: %w", err)
	}
	// 必须另起语句：等待先到的引用写入提交后，以新快照重新检查。
	tag, err := tx.Exec(ctx, cleanupConfigRefsSQL+"DELETE FROM "+table+" c WHERE c.id=ANY($2::uuid[]) AND "+predicate, kind, ids)
	if err != nil {
		return CollectionCleanupResult{}, fmt.Errorf("清理空合集: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return CollectionCleanupResult{}, fmt.Errorf("提交空合集清理: %w", err)
	}
	result.DeletedCount = int(tag.RowsAffected())
	result.SkippedCount -= result.DeletedCount
	return result, nil
}
