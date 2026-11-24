package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"video-server/internal/models"
)

// 真实 PostgreSQL 回归：批量入库最常见的请求（没有标签、没有演员、没有合集）
// 曾经把 nil 切片编码成 SQL NULL，直接撞 NOT NULL 约束导致 500。
// 只有真库能拦住这类问题，fake repository 永远发现不了。
func TestCloudIngestBatchRepositoryPersistsEmptyArrays(t *testing.T) {
	dsn := os.Getenv("CLOUD_INGEST_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("设置 CLOUD_INGEST_TEST_DATABASE_URL 后验证真实 PostgreSQL 数组列写入")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	// 单连接保证临时表对后续语句可见。
	cfg.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	// 与 migrations/0052 的列定义和约束保持一致。
	if _, err := pool.Exec(ctx, `
CREATE TEMP TABLE cloud_ingest_batches (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title_mode          TEXT NOT NULL DEFAULT 'uniform'
                        CHECK (title_mode IN ('uniform', 'filename')),
    uniform_title       TEXT NOT NULL DEFAULT '',
    video_type          TEXT NOT NULL DEFAULT 'short'
                        CHECK (video_type IN ('short', 'movie', 'episode', 'av')),
    site_category       TEXT NOT NULL DEFAULT '',
    description         TEXT NOT NULL DEFAULT '',
    tags                TEXT[] NOT NULL DEFAULT '{}',
    collection_ids      UUID[] NOT NULL DEFAULT '{}',
    actor_ids           UUID[] NOT NULL DEFAULT '{}',
    actor_names         TEXT[] NOT NULL DEFAULT '{}',
    image_collection_id UUID,
    forced_count        INTEGER NOT NULL DEFAULT 0,
    created_by          UUID,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT cloud_ingest_batches_collections_only_for_short
        CHECK (video_type = 'short' OR cardinality(collection_ids) = 0)
)`); err != nil {
		t.Fatal(err)
	}

	repo := NewCloudIngestBatchRepository(pool)

	// 显式 nil：这就是 normalizeCloudIngestBatchInput 修好之前会传进来的形状。
	created, err := repo.Create(ctx, models.CloudIngestBatch{
		TitleMode:     models.CloudIngestTitleFilename,
		VideoType:     "short",
		Tags:          nil,
		ActorNames:    nil,
		ActorIDs:      nil,
		CollectionIDs: nil,
	})
	if err != nil {
		t.Fatalf("create batch with empty arrays: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Fatal("batch id was not generated")
	}
	if created.Tags == nil || created.ActorNames == nil || created.ActorIDs == nil || created.CollectionIDs == nil {
		t.Fatalf("empty arrays must round-trip as empty slices, got %+v", created)
	}

	// 非短视频带合集必须被数据库约束拦下。
	if _, err := repo.Create(ctx, models.CloudIngestBatch{
		TitleMode:     models.CloudIngestTitleUniform,
		UniformTitle:  "合辑",
		VideoType:     "movie",
		CollectionIDs: []uuid.UUID{uuid.New()},
	}); err == nil {
		t.Fatal("database must reject collections on a non-short batch")
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete batch: %v", err)
	}
	if _, err := repo.GetByID(ctx, created.ID); err == nil {
		t.Fatal("deleted batch should not be readable")
	}
}

// 生产事故回归（真库）：失败写入必须是条件更新，不能覆盖已成功的导入。
func TestMarkFailedIfActiveDoesNotClobberImported(t *testing.T) {
	dsn := os.Getenv("CLOUD_INGEST_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("设置 CLOUD_INGEST_TEST_DATABASE_URL 后验证真实 PostgreSQL 条件更新")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, `
CREATE TEMP TABLE cloud_download_imports (
    id            BIGSERIAL PRIMARY KEY,
    task_id       BIGINT NOT NULL UNIQUE,
    kind          TEXT NOT NULL DEFAULT 'video',
    status        TEXT NOT NULL,
    error_message TEXT
);
INSERT INTO cloud_download_imports (task_id, status) VALUES (1, 'imported'), (2, 'copying');`); err != nil {
		t.Fatal(err)
	}

	repo := NewCloudDownloadImportRepository(pool)

	changed, err := repo.MarkFailedIfActive(ctx, 1, "文件缺失")
	if err != nil {
		t.Fatalf("mark imported row: %v", err)
	}
	if changed {
		t.Fatal("must not flip an imported import to failed")
	}
	var status string
	if err := pool.QueryRow(ctx, `SELECT status FROM cloud_download_imports WHERE task_id = 1`).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "imported" {
		t.Fatalf("status = %q, want imported", status)
	}

	changed, err = repo.MarkFailedIfActive(ctx, 2, "文件缺失")
	if err != nil {
		t.Fatalf("mark active row: %v", err)
	}
	if !changed {
		t.Fatal("an active import must be marked failed")
	}
}
