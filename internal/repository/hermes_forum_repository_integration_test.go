package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestCleanupExpiredForumPostsSQLAgainstPostgres(t *testing.T) {
	dsn := os.Getenv("HERMES_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set HERMES_TEST_DATABASE_URL to run the PostgreSQL retention integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to PostgreSQL test database: %v", err)
	}
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin PostgreSQL retention test: %v", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
CREATE TEMP TABLE collected_forum_posts (
    id INTEGER PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE TEMP TABLE collected_forum_post_resources (
    post_id INTEGER NOT NULL,
    kind TEXT NOT NULL
);`)
	if err != nil {
		t.Fatalf("create PostgreSQL retention test tables: %v", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO collected_forum_posts (id, created_at) VALUES
    (1, NOW() - INTERVAL '29 days'),
    (2, NOW() - INTERVAL '31 days'),
    (3, NOW() - INTERVAL '31 days'),
    (4, NOW() - INTERVAL '31 days'),
    (5, NOW() - INTERVAL '30 days');
INSERT INTO collected_forum_post_resources (post_id, kind) VALUES
    (3, 'attachment'),
    (4, 'ed2k');`)
	if err != nil {
		t.Fatalf("seed PostgreSQL retention test data: %v", err)
	}

	if _, err := tx.Exec(ctx, cleanupExpiredForumPostsSQL); err != nil {
		t.Fatalf("run PostgreSQL retention cleanup: %v", err)
	}

	rows, err := tx.Query(ctx, `SELECT id FROM collected_forum_posts ORDER BY id`)
	if err != nil {
		t.Fatalf("query PostgreSQL retention test data: %v", err)
	}
	defer rows.Close()

	var retained []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan PostgreSQL retention test data: %v", err)
		}
		retained = append(retained, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate PostgreSQL retention test data: %v", err)
	}

	want := []int{1, 3, 4, 5}
	if len(retained) != len(want) {
		t.Fatalf("retained IDs=%v, want %v", retained, want)
	}
	for i := range want {
		if retained[i] != want[i] {
			t.Fatalf("retained IDs=%v, want %v", retained, want)
		}
	}
}
