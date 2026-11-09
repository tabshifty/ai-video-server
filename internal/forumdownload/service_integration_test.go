package forumdownload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"video-server/internal/repository"
	"video-server/internal/services"
)

type recordingQueue struct {
	mu   sync.Mutex
	ids  []int64
	fail bool
}

func (q *recordingQueue) EnqueueCloudDownload(id int64, _ int) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.ids = append(q.ids, id)
	if q.fail {
		return errors.New("queue temporarily down")
	}
	return nil
}

func integrationService(t *testing.T) (*Service, *recordingQueue) {
	t.Helper()
	dsn := os.Getenv("TELEGRAM_FORUM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("设置本地隔离 TELEGRAM_FORUM_TEST_DATABASE_URL 执行 PostgreSQL 集成验证")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || (parsed.Hostname() != "127.0.0.1" && parsed.Hostname() != "localhost") {
		t.Fatal("integration tests require a loopback database")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := "forum_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = admin.Exec(ctx, `CREATE SCHEMA `+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(ctx, `DROP SCHEMA `+schema+` CASCADE`); err != nil {
			t.Error(err)
		}
	})
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	config.MaxConns = 12
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	for _, name := range []string{"0037_collected_forum_posts.up.sql", "0042_cloud_download_tasks.up.sql", "0043_cloud_download_triggered.up.sql", "0044_cloud_download_source.up.sql", "0046_cloud_download_file_deletions.up.sql", "0048_telegram_forum_download.up.sql"} {
		b, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, string(b)); err != nil {
			t.Fatalf("migration %s: %v", name, err)
		}
	}
	q := &recordingQueue{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	downloads := services.NewCloudDownloadService(repository.NewCloudDownloadRepository(pool), nil, t.TempDir(), q, logger)
	return New(pool, downloads, logger), q
}

func seedPost(t *testing.T, s *Service, links []string) Post {
	t.Helper()
	id := uuid.New()
	ctx := context.Background()
	_, err := s.pool.Exec(ctx, `INSERT INTO collected_forum_posts(id,source,board_key,external_post_id,title,url,inspection_status,filter_decision,result_fingerprint,observed_at,inspected_at) VALUES($1,'sehuatang','95',$2,'测试帖子','https://example.invalid/post','inspected','included',$3,NOW(),NOW())`, id, fmt.Sprint(time.Now().UnixNano()), strings.Repeat("a", 64))
	if err != nil {
		t.Fatal(err)
	}
	for i, link := range links {
		if _, err := s.pool.Exec(ctx, `INSERT INTO collected_forum_post_resources(post_id,kind,value,position) VALUES($1,'ed2k',$2,$3)`, id, link, i); err != nil {
			t.Fatal(err)
		}
	}
	p, err := loadPost(ctx, s.pool, id, "")
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func testLink(n int) string { return fmt.Sprintf("ed2k://|file|测试%d.zip|123|%032x|/", n, n) }
func requestFor(p Post) Request {
	return Request{EventID: uuid.NewString(), ActorID: 123, ChatID: -1001, MessageID: 2, PostID: p.ID, Fingerprint: p.Fingerprint, All: true}
}
func count(t *testing.T, s *Service, table string) int {
	t.Helper()
	var n int
	if err := s.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM "+table).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestPostgresConcurrentReplayAndReceiptRecovery(t *testing.T) {
	s, q := integrationService(t)
	ctx := context.Background()
	p := seedPost(t, s, []string{testLink(1), "invalid", testLink(1), testLink(2)})
	req := requestFor(p)
	var wg sync.WaitGroup
	results := make(chan Result, 8)
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); result, err := s.submit(ctx, req); results <- result; errs <- err }()
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var first Result
	for result := range results {
		if first.ID == uuid.Nil {
			first = result
		}
		if result.ID != first.ID {
			t.Fatal("replay created new operation")
		}
	}
	if count(t, s, "cloud_download_tasks") != 2 || count(t, s, "telegram_forum_download_operations") != 1 {
		t.Fatal("duplicate durable state")
	}
	if first.Items[1].Outcome != "invalid" || first.Items[2].Outcome != "duplicate" {
		t.Fatalf("outcomes=%+v", first.Items)
	}
	q.mu.Lock()
	enqueues := len(q.ids)
	q.mu.Unlock()
	if enqueues != 2 {
		t.Fatalf("enqueue=%d", enqueues)
	}
	altered := req
	altered.MessageID++
	if _, err := s.submit(ctx, altered); !errors.Is(err, errConflict) {
		t.Fatalf("expected replay conflict: %v", err)
	}
	receipts, err := s.receipts(ctx)
	if err != nil || len(receipts) != 1 || receipts[0].Key != "accepted" {
		t.Fatalf("receipts=%+v err=%v", receipts, err)
	}
	if next, err := s.receipts(ctx); err != nil || len(next) != 0 {
		t.Fatalf("lease overlap: %+v %v", next, err)
	}
	if err := s.ack(ctx, receipts[0].ID, uuid.New()); !errors.Is(err, errConflict) {
		t.Fatal("wrong lease accepted")
	}
	if _, err := s.pool.Exec(ctx, `UPDATE telegram_forum_download_receipts SET lease_until=NOW()-INTERVAL '1 minute'`); err != nil {
		t.Fatal(err)
	}
	retried, err := s.receipts(ctx)
	if err != nil || len(retried) != 1 || retried[0].Token == receipts[0].Token {
		t.Fatal("lease did not recover")
	}
	if err := s.ack(ctx, retried[0].ID, retried[0].Token); err != nil {
		t.Fatal(err)
	}
	if _, err := s.pool.Exec(ctx, `UPDATE cloud_download_tasks SET status='completed'`); err != nil {
		t.Fatal(err)
	}
	terminals, err := s.receipts(ctx)
	if err != nil || len(terminals) != 2 {
		t.Fatalf("terminal=%+v %v", terminals, err)
	}
	for _, r := range terminals {
		var payload map[string]any
		if err := json.Unmarshal(r.Payload, &payload); err != nil {
			t.Fatal(err)
		}
		if payload["status"] != "completed" {
			t.Fatal(payload)
		}
		if err := s.ack(ctx, r.ID, r.Token); err != nil {
			t.Fatal(err)
		}
	}
	req.EventID = uuid.NewString()
	again, err := s.submit(ctx, req)
	if err != nil || again.Items[0].Outcome != "skipped_completed" {
		t.Fatalf("skip=%+v %v", again, err)
	}
}

func TestPostgresSavepointAndCommittedQueueRecovery(t *testing.T) {
	s, q := integrationService(t)
	ctx := context.Background()
	p := seedPost(t, s, []string{testLink(1), testLink(2), testLink(3)})
	_, err := s.pool.Exec(ctx, `CREATE FUNCTION reject_test_task() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.info_hash='00000000000000000000000000000002' THEN RAISE EXCEPTION 'test resource failure'; END IF; RETURN NEW; END $$; CREATE TRIGGER reject_task BEFORE INSERT ON cloud_download_tasks FOR EACH ROW EXECUTE FUNCTION reject_test_task()`)
	if err != nil {
		t.Fatal(err)
	}
	q.fail = true
	req := requestFor(p)
	result, err := s.submit(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if result.Items[1].Outcome != "failed" || count(t, s, "cloud_download_tasks") != 2 {
		t.Fatalf("partial=%+v", result)
	}
	if _, err := s.pool.Exec(ctx, `DROP TRIGGER reject_task ON cloud_download_tasks`); err != nil {
		t.Fatal(err)
	}
	q.mu.Lock()
	q.fail = false
	q.ids = nil
	q.mu.Unlock()
	if err := s.downloads.Reconcile(ctx); err != nil {
		t.Fatal(err)
	}
	q.mu.Lock()
	n := len(q.ids)
	q.mu.Unlock()
	if n != 2 {
		t.Fatalf("committed tasks not recovered: %d", n)
	}
	task := result.Items[0].TaskID
	if _, err := s.pool.Exec(ctx, `UPDATE cloud_download_tasks SET status='failed' WHERE id=$1`, task); err != nil {
		t.Fatal(err)
	}
	req.EventID = uuid.NewString()
	req.All = false
	req.ResourceIDs = []uuid.UUID{p.Resources[0].ID}
	req.RetryAttempts = map[string]int{p.Resources[0].ID.String(): 0}
	retry, err := s.submit(ctx, req)
	if err != nil || retry.Items[0].Outcome != "retried" {
		t.Fatalf("retry=%+v %v", retry, err)
	}
	var attempt int
	if err := s.pool.QueryRow(ctx, `SELECT retry_count FROM cloud_download_tasks WHERE id=$1`, task).Scan(&attempt); err != nil || attempt != 1 {
		t.Fatalf("attempt=%d %v", attempt, err)
	}
	// A replay after the task fails again must return its original result without another retry.
	if _, err := s.pool.Exec(ctx, `UPDATE cloud_download_tasks SET status='failed' WHERE id=$1`, task); err != nil {
		t.Fatal(err)
	}
	if _, err := s.submit(ctx, req); err != nil {
		t.Fatal(err)
	}
	if err := s.pool.QueryRow(ctx, `SELECT retry_count FROM cloud_download_tasks WHERE id=$1`, task).Scan(&attempt); err != nil || attempt != 1 {
		t.Fatal("replay retried again")
	}
	// Down migration removes only integration state, retaining reusable download tasks.
	down, err := os.ReadFile("../../migrations/0048_telegram_forum_download.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.pool.Exec(ctx, string(down)); err != nil {
		t.Fatal(err)
	}
	if count(t, s, "cloud_download_tasks") != 2 {
		t.Fatal("rollback removed downloads")
	}
	up, err := os.ReadFile("../../migrations/0048_telegram_forum_download.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.pool.Exec(ctx, string(up)); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresResourceSnapshotAndRetryConfirmation(t *testing.T) {
	s, _ := integrationService(t)
	ctx := context.Background()
	variant := strings.Replace(testLink(1), "ed2k://|", "ed2k://", 1)
	p := seedPost(t, s, []string{testLink(1), variant})
	req := requestFor(p)
	req.All = false
	req.ResourceIDs = []uuid.UUID{p.Resources[0].ID}
	first, err := s.submit(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	taskID := first.Items[0].TaskID
	if _, err := s.pool.Exec(ctx, `UPDATE cloud_download_tasks SET status='failed' WHERE id=$1`, taskID); err != nil {
		t.Fatal(err)
	}
	p, err = loadPost(ctx, s.pool, p.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	if p.Resources[1].Status != "failed" || p.Resources[1].TaskID == nil || *p.Resources[1].TaskID != *taskID {
		t.Fatalf("variant status=%+v", p.Resources[1])
	}
	// Original click did not confirm retry: a newly failed task must not be reset.
	req.EventID = uuid.NewString()
	req.ResourceIDs = []uuid.UUID{p.Resources[1].ID}
	refused, err := s.submit(ctx, req)
	if err != nil || refused.Items[0].Outcome != "failed" {
		t.Fatalf("refusal=%+v %v", refused, err)
	}
	var attempt int
	if err := s.pool.QueryRow(ctx, `SELECT retry_count FROM cloud_download_tasks WHERE id=$1`, taskID).Scan(&attempt); err != nil || attempt != 0 {
		t.Fatal("unconfirmed retry occurred")
	}
	req.EventID = uuid.NewString()
	req.RetryAttempts = map[string]int{p.Resources[1].ID.String(): 0}
	retry, err := s.submit(ctx, req)
	if err != nil || retry.Items[0].Outcome != "retried" {
		t.Fatalf("retry=%+v %v", retry, err)
	}
	if _, err := s.pool.Exec(ctx, `UPDATE collected_forum_post_resources SET value=$2 WHERE id=$1`, p.Resources[1].ID, testLink(2)); err != nil {
		t.Fatal(err)
	}
	req.EventID = uuid.NewString()
	if _, err := s.submit(ctx, req); !errors.Is(err, errConflict) {
		t.Fatalf("old snapshot accepted: %v", err)
	}
}

func TestPostgresTransactionalLookupLocksRetryAttempt(t *testing.T) {
	s, _ := integrationService(t)
	ctx := context.Background()
	p := seedPost(t, s, []string{testLink(1)})
	first, err := s.submit(ctx, requestFor(p))
	if err != nil {
		t.Fatal(err)
	}
	taskID := first.Items[0].TaskID
	if _, err := s.pool.Exec(ctx, `UPDATE cloud_download_tasks SET status='failed' WHERE id=$1`, taskID); err != nil {
		t.Fatal(err)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	locked := repository.NewCloudDownloadRepositoryWithTx(tx)
	if _, err := locked.GetByInfoHash(ctx, fmt.Sprintf("%032x", 1)); err != nil {
		t.Fatal(err)
	}
	other, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Rollback(ctx)
	if _, err := other.Exec(ctx, `SET LOCAL lock_timeout='100ms'`); err != nil {
		t.Fatal(err)
	}
	if _, err := other.Exec(ctx, `UPDATE cloud_download_tasks SET retry_count=retry_count+1 WHERE id=$1`, taskID); err == nil {
		t.Fatal("concurrent entry changed the confirmed attempt before its transactional retry")
	}
}
