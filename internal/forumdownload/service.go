package forumdownload

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"video-server/internal/cloud"
	"video-server/internal/models"
	"video-server/internal/services"
)

// Service coordinates durable operations with the existing download state machine.
type Service struct {
	pool      *pgxpool.Pool
	downloads *services.CloudDownloadService
	logger    *slog.Logger
}

// New constructs the integration without starting a second Bot consumer.
func New(pool *pgxpool.Pool, downloads *services.CloudDownloadService, logger *slog.Logger) *Service {
	return &Service{pool, downloads, logger}
}

type querier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func loadPost(ctx context.Context, db querier, id uuid.UUID, tid string) (Post, error) {
	p := Post{Resources: []Resource{}}
	err := db.QueryRow(ctx, `SELECT id,title FROM collected_forum_posts WHERE (id=$1 OR ($2<>'' AND source='sehuatang' AND external_post_id=$2)) AND filter_decision='included' AND inspection_status IN ('inspected','restricted')`, id, tid).Scan(&p.ID, &p.Title)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, errNotFound
	}
	if err != nil {
		return p, err
	}
	rows, err := db.Query(ctx, `SELECT r.id,r.value FROM collected_forum_post_resources r WHERE r.post_id=$1 AND r.kind='ed2k' ORDER BY r.position,r.id FOR SHARE OF r`, p.ID)
	if err != nil {
		return p, err
	}
	hashes := []string{}
	byHash := map[string][]int{}
	for rows.Next() {
		var r Resource
		if err := rows.Scan(&r.ID, &r.Link); err != nil {
			rows.Close()
			return p, err
		}
		r = parseResource(r)
		if link, err := cloud.ParseLink(r.Link); err == nil && link.Kind == cloud.KindED2K {
			byHash[link.InfoHash] = append(byHash[link.InfoHash], len(p.Resources))
			hashes = append(hashes, link.InfoHash)
		}
		p.Resources = append(p.Resources, r)
	}
	rowErr := rows.Err()
	rows.Close()
	if rowErr != nil {
		return p, rowErr
	}
	if len(hashes) > 0 {
		states, err := db.Query(ctx, `SELECT id,info_hash,status,retry_count FROM cloud_download_tasks WHERE protocol='ed2k' AND info_hash=ANY($1)`, hashes)
		if err != nil {
			return p, err
		}
		defer states.Close()
		for states.Next() {
			var id int64
			var hash, status string
			var attempt int
			if err := states.Scan(&id, &hash, &status, &attempt); err != nil {
				return p, err
			}
			for _, index := range byHash[hash] {
				p.Resources[index].TaskID = &id
				p.Resources[index].Status = status
				p.Resources[index].RetryCount = attempt
			}
		}
		if err := states.Err(); err != nil {
			return p, err
		}
	}
	p.Fingerprint = fingerprint(p.Resources)
	return p, nil
}

// Request identifies a Telegram event and an explicit server-owned resource selection.
type Request struct {
	EventID       string         `json:"event_id"`
	ActorID       int64          `json:"actor_id"`
	ChatID        int64          `json:"chat_id"`
	MessageID     int64          `json:"message_id"`
	ThreadID      *int64         `json:"thread_id"`
	PostID        uuid.UUID      `json:"post_id"`
	Fingerprint   string         `json:"fingerprint"`
	ResourceIDs   []uuid.UUID    `json:"resource_ids"`
	All           bool           `json:"all"`
	RetryAttempts map[string]int `json:"retry_attempts"`
}

// Result reports the independently accepted, skipped, or failed items.
type Result struct {
	ID    uuid.UUID `json:"id"`
	Items []Item    `json:"items"`
}

// Item is one explicit selection outcome, rather than a batch-wide success flag.
type Item struct {
	ResourceID uuid.UUID `json:"resource_id"`
	Name       string    `json:"name"`
	TaskID     *int64    `json:"task_id"`
	Outcome    string    `json:"outcome"`
	Error      string    `json:"error,omitempty"`
}

func (s *Service) submit(ctx context.Context, req Request) (Result, error) {
	var result Result
	if req.EventID == "" || len(req.EventID) > 200 || req.ActorID <= 0 || req.ChatID == 0 || req.MessageID <= 0 || req.PostID == uuid.Nil || req.Fingerprint == "" || (req.ThreadID != nil && *req.ThreadID <= 0) {
		return result, errInvalid
	}
	encoded, err := json.Marshal(req)
	if err != nil {
		return result, err
	}
	sum := sha256.Sum256(encoded)
	requestHash := hex.EncodeToString(sum[:])
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)
	// Serialize replays before checking persisted outcomes. Hash collisions only serialize unrelated requests.
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, req.EventID); err != nil {
		return result, err
	}
	var savedHash string
	var saved []byte
	err = tx.QueryRow(ctx, `SELECT request_hash,result FROM telegram_forum_download_operations WHERE event_id=$1`, req.EventID).Scan(&savedHash, &saved)
	if err == nil {
		if savedHash != requestHash {
			return result, errConflict
		}
		err = json.Unmarshal(saved, &result)
		return result, err
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return result, err
	}
	// Lock parent while selecting; a concurrent delete cannot turn a valid selection into an arbitrary URL.
	var locked uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT id FROM collected_forum_posts WHERE id=$1 FOR SHARE`, req.PostID).Scan(&locked); errors.Is(err, pgx.ErrNoRows) {
		return result, errNotFound
	} else if err != nil {
		return result, err
	}
	post, err := loadPost(ctx, tx, req.PostID, "")
	if err != nil {
		return result, err
	}
	if post.Fingerprint != req.Fingerprint {
		return result, errConflict
	}
	selected, err := selectResources(post.Resources, req.ResourceIDs, req.All)
	if err != nil {
		return result, err
	}
	// A stable cross-post order prevents reversed multi-resource requests deadlocking.
	hashes := map[string]bool{}
	for _, r := range selected {
		if link, err := cloud.ParseLink(r.Link); err == nil && link.Kind == cloud.KindED2K {
			hashes[link.InfoHash] = true
		}
	}
	keys := make([]string, 0, len(hashes))
	for hash := range hashes {
		keys = append(keys, hash)
	}
	sort.Strings(keys)
	for _, hash := range keys {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, "telegram-download:"+hash); err != nil {
			return result, err
		}
	}
	result = Result{ID: uuid.New(), Items: []Item{}}
	_, err = tx.Exec(ctx, `INSERT INTO telegram_forum_download_operations(id,event_id,request_hash,actor_id,chat_id,message_id,thread_id,post_id,result) VALUES($1,$2,$3,$4,$5,$6,$7,$8,'{}')`, result.ID, req.EventID, requestHash, req.ActorID, req.ChatID, req.MessageID, req.ThreadID, req.PostID)
	if err != nil {
		return result, err
	}
	tasks := []models.CloudDownloadTask{}
	seen := map[int64]bool{}
	for _, r := range selected {
		item := Item{ResourceID: r.ID, Name: r.Name, Outcome: models.CloudBatchInvalid}
		attempt := r.RetryCount
		if r.Status != "invalid" {
			// Savepoint contains malformed or failed items without aborting the accepted siblings.
			nested, beginErr := tx.Begin(ctx)
			if beginErr != nil {
				return result, beginErr
			}
			confirmedRetry := -1
			if retry, ok := req.RetryAttempts[r.ID.String()]; ok {
				confirmedRetry = retry
			}
			task, created, submitErr := s.downloads.SubmitInTransaction(ctx, nested, post.ID, r.Link, confirmedRetry)
			if submitErr != nil {
				if rollbackErr := nested.Rollback(ctx); rollbackErr != nil {
					return result, rollbackErr
				}
				item.Outcome = models.CloudBatchFailed
				item.Error = "提交失败，请稍后重试"
				if errors.Is(submitErr, models.ErrCloudDownloadNotRetryable) {
					item.Error = "任务已变化，请刷新并确认重试"
				}
				s.logger.Warn("Telegram resource submit failed", "resource_id", r.ID, "error", submitErr)
			} else {
				if err := nested.Commit(ctx); err != nil {
					return result, err
				}
				item.TaskID = &task.ID
				attempt = task.RetryCount
				switch {
				case seen[task.ID]:
					item.Outcome = models.CloudBatchDuplicate
				case task.Status == models.CloudDownloadCompleted:
					item.Outcome = models.CloudBatchSkippedCompleted
				case created:
					item.Outcome = models.CloudBatchSubmitted
				case task.RetryCount > r.RetryCount:
					item.Outcome = models.CloudBatchRetried
				case task.Status == models.CloudDownloadPending:
					item.Outcome = models.CloudBatchSubmitted
				default:
					item.Outcome = models.CloudBatchSkippedActive
				}
				seen[task.ID] = true
				if task.Status == models.CloudDownloadPending && item.Outcome != models.CloudBatchDuplicate {
					tasks = append(tasks, task)
				}
			}
		} else {
			item.Error = "无效 ED2K 链接"
		}
		if _, err = tx.Exec(ctx, `INSERT INTO telegram_forum_download_items(operation_id,resource_id,task_id,attempt,outcome) VALUES($1,$2,$3,$4,$5)`, result.ID, r.ID, item.TaskID, attempt, item.Outcome); err != nil {
			return result, err
		}
		result.Items = append(result.Items, item)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return result, err
	}
	if _, err = tx.Exec(ctx, `UPDATE telegram_forum_download_operations SET result=$2 WHERE id=$1`, result.ID, payload); err != nil {
		return result, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO telegram_forum_download_receipts(operation_id,receipt_key,payload) VALUES($1,'accepted',$2)`, result.ID, payload); err != nil {
		return result, err
	}
	if err = tx.Commit(ctx); err != nil {
		return result, err
	}
	for _, task := range tasks {
		if err := s.downloads.EnqueueCommittedTask(task); err != nil {
			s.logger.Warn("Telegram task wakeup deferred to reconciler", "task_id", task.ID, "error", err)
		}
	}
	return result, nil
}

// Receipt is a leased outgoing message; acknowledge only after Telegram confirms delivery.
type Receipt struct {
	ID        uuid.UUID       `json:"id"`
	Token     uuid.UUID       `json:"token"`
	Key       string          `json:"key"`
	ActorID   int64           `json:"actor_id"`
	ChatID    int64           `json:"chat_id"`
	MessageID int64           `json:"message_id"`
	ThreadID  *int64          `json:"thread_id"`
	Payload   json.RawMessage `json:"payload"`
}

func (s *Service) receipts(ctx context.Context) ([]Receipt, error) {
	// Materialize terminal notifications durably. Replayed/retried tasks are distinguished by attempt.
	_, err := s.pool.Exec(ctx, `INSERT INTO telegram_forum_download_receipts(operation_id,receipt_key,payload)
 SELECT i.operation_id,'terminal:'||i.task_id, jsonb_build_object('task_id',t.id,'name',t.declared_name,'status',CASE WHEN t.retry_count<>i.attempt THEN 'superseded' ELSE t.status END,'blocked_reason',t.blocked_reason)
 FROM telegram_forum_download_items i JOIN cloud_download_tasks t ON t.id=i.task_id
 WHERE i.outcome IN ('submitted','retried','skipped_active') AND (t.status IN ('completed','failed','blocked') OR t.retry_count<>i.attempt)
 ON CONFLICT(operation_id,receipt_key) DO NOTHING`)
	if err != nil {
		return nil, err
	}
	token := uuid.New()
	rows, err := s.pool.Query(ctx, `WITH pending AS (SELECT r.id FROM telegram_forum_download_receipts r WHERE r.delivered_at IS NULL AND (r.lease_until IS NULL OR r.lease_until<NOW()) AND (r.receipt_key='accepted' OR EXISTS(SELECT 1 FROM telegram_forum_download_receipts a WHERE a.operation_id=r.operation_id AND a.receipt_key='accepted' AND a.delivered_at IS NOT NULL)) ORDER BY r.created_at,r.id LIMIT 5 FOR UPDATE SKIP LOCKED), leased AS (UPDATE telegram_forum_download_receipts r SET lease_token=$1,lease_until=NOW()+INTERVAL '5 minutes' FROM pending p WHERE r.id=p.id RETURNING r.*) SELECT r.id,r.lease_token,r.receipt_key,o.actor_id,o.chat_id,o.message_id,o.thread_id,r.payload FROM leased r JOIN telegram_forum_download_operations o ON o.id=r.operation_id ORDER BY r.created_at,r.id`, token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Receipt{}
	for rows.Next() {
		var r Receipt
		if err := rows.Scan(&r.ID, &r.Token, &r.Key, &r.ActorID, &r.ChatID, &r.MessageID, &r.ThreadID, &r.Payload); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Service) ack(ctx context.Context, id, token uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `UPDATE telegram_forum_download_receipts SET delivered_at=NOW(),lease_until=NULL WHERE id=$1 AND lease_token=$2 AND delivered_at IS NULL`, id, token)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errConflict
	}
	return nil
}

func parseActor(raw string, owner int64) error {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id != owner || owner <= 0 {
		return fmt.Errorf("无权操作")
	}
	return nil
}

// RequestTimeout bounds an integration request, including database lock contention.
const RequestTimeout = 30 * time.Second
