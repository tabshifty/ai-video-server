package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

var (
	ErrHermesForumPostNotFound       = errors.New("Hermes forum post not found")
	ErrHermesForumInspectionConflict = errors.New("Hermes forum inspection conflict")
)

type HermesForumRepository struct {
	videoRepo *VideoRepository
}

func NewHermesForumRepository(videoRepo *VideoRepository) *HermesForumRepository {
	return &HermesForumRepository{videoRepo: videoRepo}
}

func buildAdminForumPostListSQL(page, pageSize int) (string, string, []any) {
	const where = `
WHERE p.inspection_status = 'inspected'
  AND p.filter_decision = 'included'
  AND p.created_at >= NOW() - INTERVAL '30 days'`
	countSQL := `SELECT COUNT(*) FROM collected_forum_posts p` + where
	listSQL := `
SELECT p.id,
       p.title,
       p.url,
       COALESCE(
           ARRAY_AGG(r.value ORDER BY r.position) FILTER (WHERE r.kind = 'attachment'),
           ARRAY[]::TEXT[]
       ) AS attachments,
       COALESCE(
           ARRAY_AGG(r.value ORDER BY r.position) FILTER (WHERE r.kind = 'ed2k'),
           ARRAY[]::TEXT[]
       ) AS ed2k_links,
       p.observed_at
FROM collected_forum_posts p
LEFT JOIN collected_forum_post_resources r ON r.post_id = p.id` + where + `
GROUP BY p.id, p.title, p.url, p.observed_at
ORDER BY p.observed_at DESC, p.id DESC
LIMIT $1 OFFSET $2`
	return countSQL, listSQL, []any{pageSize, (page - 1) * pageSize}
}

// ListAdminForumPosts returns the recent, included forum resource projection for the admin UI.
func (r *HermesForumRepository) ListAdminForumPosts(ctx context.Context, page, pageSize int) ([]models.AdminForumPostListItem, int, error) {
	countSQL, listSQL, listArgs := buildAdminForumPostListSQL(page, pageSize)
	var total int
	if err := r.videoRepo.pool.QueryRow(ctx, countSQL).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin forum posts: %w", err)
	}

	rows, err := r.videoRepo.pool.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin forum posts: %w", err)
	}
	defer rows.Close()

	items := make([]models.AdminForumPostListItem, 0, pageSize)
	for rows.Next() {
		var item models.AdminForumPostListItem
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.URL,
			&item.Attachments,
			&item.ED2KLinks,
			&item.ObservedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan admin forum post: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate admin forum posts: %w", err)
	}
	return items, total, nil
}

func (r *HermesForumRepository) DiscoverForumPosts(ctx context.Context, input models.ForumPostDiscoverInput) ([]models.ForumPostDiscoverResult, error) {
	tx, err := r.videoRepo.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin forum discover transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM collected_forum_posts WHERE created_at < NOW() - INTERVAL '30 days'`); err != nil {
		return nil, fmt.Errorf("clean expired forum posts: %w", err)
	}

	status := models.ForumPostStatusDedupeOnly
	if input.Mode == models.ForumPostDiscoverModeNormal {
		status = models.ForumPostStatusPending
	}
	results := make([]models.ForumPostDiscoverResult, 0, len(input.Posts))
	for _, post := range input.Posts {
		var id uuid.UUID
		var storedStatus string
		err := tx.QueryRow(ctx, `
INSERT INTO collected_forum_posts
    (source, board_key, external_post_id, title, url, inspection_status, observed_at)
VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, $7)
ON CONFLICT (source, external_post_id) DO NOTHING
RETURNING id, inspection_status`, input.Source, input.BoardKey, post.TID, post.Title, post.URL, status, input.ObservedAt).Scan(&id, &storedStatus)

		created := err == nil
		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.QueryRow(ctx, `
SELECT id, inspection_status
FROM collected_forum_posts
WHERE source = $1 AND external_post_id = $2`, input.Source, post.TID).Scan(&id, &storedStatus)
		}
		if err != nil {
			return nil, fmt.Errorf("store forum post %q: %w", post.TID, err)
		}

		disposition := models.ForumPostDispositionDuplicate
		if created {
			disposition = models.ForumPostDispositionCreated
		} else if storedStatus == models.ForumPostStatusPending {
			disposition = models.ForumPostDispositionPending
		}
		results = append(results, models.ForumPostDiscoverResult{
			TID: post.TID, ID: id, Disposition: disposition, InspectionStatus: storedStatus,
		})
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit forum discover transaction: %w", err)
	}
	return results, nil
}

func (r *HermesForumRepository) CompleteForumPostInspection(ctx context.Context, postID uuid.UUID, input models.ForumPostInspectionInput, fingerprint string) (models.ForumPostInspectionResult, error) {
	tx, err := r.videoRepo.pool.Begin(ctx)
	if err != nil {
		return models.ForumPostInspectionResult{}, fmt.Errorf("begin forum inspection transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var status string
	var storedFingerprint *string
	var inspectedAt *time.Time
	err = tx.QueryRow(ctx, `
SELECT inspection_status, result_fingerprint, inspected_at
FROM collected_forum_posts
WHERE id = $1
FOR UPDATE`, postID).Scan(&status, &storedFingerprint, &inspectedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ForumPostInspectionResult{}, ErrHermesForumPostNotFound
	}
	if err != nil {
		return models.ForumPostInspectionResult{}, fmt.Errorf("lock forum post: %w", err)
	}
	if status != models.ForumPostStatusPending {
		if storedFingerprint != nil && *storedFingerprint == fingerprint && inspectedAt != nil {
			return models.ForumPostInspectionResult{Status: status, Idempotent: true, InspectedAt: *inspectedAt}, nil
		}
		return models.ForumPostInspectionResult{}, ErrHermesForumInspectionConflict
	}

	reasonsJSON, err := json.Marshal(input.FilterReasons)
	if err != nil {
		return models.ForumPostInspectionResult{}, fmt.Errorf("marshal forum filter reasons: %w", err)
	}
	var completedAt time.Time
	err = tx.QueryRow(ctx, `
UPDATE collected_forum_posts
SET inspection_status = $2,
    filter_decision = $3,
    filter_reasons = $4::jsonb,
    fetch_method = NULLIF($5, ''),
    error_summary = NULLIF($6, ''),
    result_fingerprint = $7,
    inspected_at = NOW()
WHERE id = $1
RETURNING inspected_at`, postID, input.Status, input.FilterDecision, string(reasonsJSON), input.FetchMethod, input.ErrorSummary, fingerprint).Scan(&completedAt)
	if err != nil {
		return models.ForumPostInspectionResult{}, fmt.Errorf("update forum inspection: %w", err)
	}
	for i, value := range input.Attachments {
		if _, err := tx.Exec(ctx, `INSERT INTO collected_forum_post_resources (post_id, kind, value, position) VALUES ($1, 'attachment', $2, $3)`, postID, value, i); err != nil {
			return models.ForumPostInspectionResult{}, fmt.Errorf("insert forum attachment: %w", err)
		}
	}
	for i, value := range input.ED2KLinks {
		if _, err := tx.Exec(ctx, `INSERT INTO collected_forum_post_resources (post_id, kind, value, position) VALUES ($1, 'ed2k', $2, $3)`, postID, value, i); err != nil {
			return models.ForumPostInspectionResult{}, fmt.Errorf("insert forum ed2k link: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return models.ForumPostInspectionResult{}, fmt.Errorf("commit forum inspection transaction: %w", err)
	}
	return models.ForumPostInspectionResult{Status: input.Status, InspectedAt: completedAt}, nil
}
