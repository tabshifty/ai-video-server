package repository

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestTVScreensaverImagesPostgresVisibilityAndCursor(t *testing.T) {
	dsn := os.Getenv("TV_SCREENSAVER_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("设置 TV_SCREENSAVER_TEST_DATABASE_URL 后验证真实 PostgreSQL 候选查询")
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
	// 单连接临时表独立于业务 schema；不需要合集表，确保查询包含未归合集图片。
	_, err = pool.Exec(ctx, `CREATE TEMP TABLE images (
id UUID PRIMARY KEY, active BOOLEAN, status TEXT, stored_path TEXT, updated_at TIMESTAMPTZ DEFAULT now()
);
INSERT INTO images (id, active, status, stored_path) VALUES
('00000000-0000-0000-0000-000000000001', TRUE, 'ready', '/fixture/unassigned'),
('00000000-0000-0000-0000-000000000002', FALSE, 'ready', '/fixture/inactive'),
('00000000-0000-0000-0000-000000000003', TRUE, 'ready', '/fixture/second'),
('00000000-0000-0000-0000-000000000004', TRUE, 'failed', '/fixture/failed'),
('00000000-0000-0000-0000-000000000005', TRUE, 'ready', ''),
('00000000-0000-0000-0000-000000000006', TRUE, 'processing', '/fixture/pending'),
('00000000-0000-0000-0000-000000000007', TRUE, 'ready', '/fixture/last');`)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewVideoRepository(pool)
	first, err := repo.ListTVScreensaverImages(ctx, uuid.Nil, 1)
	if err != nil || len(first.Items) != 1 || first.Items[0].ID.String() != "00000000-0000-0000-0000-000000000001" || first.NextCursor != first.Items[0].ID.String() {
		t.Fatalf("first page=%+v err=%v", first, err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM images WHERE id = $1`, first.Items[0].ID); err != nil {
		t.Fatal(err)
	}
	second, err := repo.ListTVScreensaverImages(ctx, uuid.MustParse(first.NextCursor), 1)
	if err != nil || len(second.Items) != 1 || second.Items[0].ID.String() != "00000000-0000-0000-0000-000000000003" {
		t.Fatalf("cursor after deletion skipped image: page=%+v err=%v", second, err)
	}
	last, err := repo.ListTVScreensaverImages(ctx, uuid.MustParse(second.NextCursor), 1)
	if err != nil || len(last.Items) != 1 || last.Items[0].ID.String() != "00000000-0000-0000-0000-000000000007" || last.NextCursor != "" {
		t.Fatalf("terminal page=%+v err=%v", last, err)
	}
	if last.Items[0].ViewURL != "/api/v1/images/00000000-0000-0000-0000-000000000007/view" {
		t.Fatalf("unexpected view URL: %s", last.Items[0].ViewURL)
	}
	encoded, err := json.Marshal(last)
	if err != nil || strings.Contains(string(encoded), "fixture") || strings.Contains(string(encoded), "stored_path") {
		t.Fatalf("candidate must not expose storage paths: %s err=%v", encoded, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE images SET active = FALSE`); err != nil {
		t.Fatal(err)
	}
	empty, err := repo.ListTVScreensaverImages(ctx, uuid.Nil, 200)
	if err != nil || empty.Items == nil || len(empty.Items) != 0 || empty.NextCursor != "" {
		t.Fatalf("empty library=%+v err=%v", empty, err)
	}
}
