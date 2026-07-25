package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
	"video-server/internal/utils"
)

var ErrCollectionsOnlyForShort = errors.New("collections only support short videos")

func normalizeCollectionName(raw string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(raw)), " "))
}

func normalizeCollectionInput(in models.AdminCollectionInput) (models.AdminCollectionInput, error) {
	in.Name = strings.Join(strings.Fields(strings.TrimSpace(in.Name)), " ")
	if in.Name == "" {
		return in, fmt.Errorf("合集名称不能为空")
	}
	in.Description = strings.TrimSpace(in.Description)
	in.CoverURL = strings.TrimSpace(in.CoverURL)
	return in, nil
}

func scanAdminCollection(rows rowScanner) (models.AdminCollection, error) {
	var out models.AdminCollection
	if err := rows.Scan(
		&out.ID,
		&out.Name,
		&out.Description,
		&out.CoverURL,
		&out.SortOrder,
		&out.Active,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return models.AdminCollection{}, err
	}
	return out, nil
}

func (r *VideoRepository) ListCollections(ctx context.Context, q string, active *bool, page, pageSize int) ([]models.AdminCollection, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 6)
	next := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	keyword := strings.ToLower(strings.TrimSpace(q))
	if keyword != "" {
		pattern := "%" + keyword + "%"
		where = append(where, "(LOWER(name) LIKE "+next(pattern)+" OR LOWER(COALESCE(description,'')) LIKE "+next(pattern)+")")
	}
	if active != nil {
		where = append(where, "active = "+next(*active))
	}

	baseWhere := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM collections WHERE "+baseWhere, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count collections: %w", err)
	}

	args = append(args, pageSize, (page-1)*pageSize)
	sql := "SELECT id, name, COALESCE(description,''), COALESCE(cover_url,''), sort_order, active, created_at, updated_at FROM collections WHERE " +
		baseWhere + " ORDER BY sort_order DESC, updated_at DESC, name ASC LIMIT $" + fmt.Sprintf("%d", len(args)-1) +
		" OFFSET $" + fmt.Sprintf("%d", len(args))
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list collections: %w", err)
	}
	defer rows.Close()

	items := make([]models.AdminCollection, 0, pageSize)
	for rows.Next() {
		item, scanErr := scanAdminCollection(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("scan collection: %w", scanErr)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *VideoRepository) CreateCollection(ctx context.Context, input models.AdminCollectionInput) (models.AdminCollection, error) {
	input, err := normalizeCollectionInput(input)
	if err != nil {
		return models.AdminCollection{}, err
	}

	row := r.pool.QueryRow(ctx, `
INSERT INTO collections (
  id, name, normalized_name, description, cover_url, sort_order, active
)
VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING id, name, COALESCE(description,''), COALESCE(cover_url,''), sort_order, active, created_at, updated_at
`, uuid.New(), input.Name, normalizeCollectionName(input.Name), input.Description, input.CoverURL, input.SortOrder, input.Active)
	out, err := scanAdminCollection(row)
	if err != nil {
		return models.AdminCollection{}, fmt.Errorf("create collection: %w", err)
	}
	return out, nil
}

func (r *VideoRepository) UpdateCollection(ctx context.Context, collectionID uuid.UUID, input models.AdminCollectionInput) (models.AdminCollection, error) {
	input, err := normalizeCollectionInput(input)
	if err != nil {
		return models.AdminCollection{}, err
	}

	row := r.pool.QueryRow(ctx, `
UPDATE collections
SET
  name=$2,
  normalized_name=$3,
  description=$4,
  cover_url=$5,
  sort_order=$6,
  active=$7,
  updated_at=NOW()
WHERE id=$1
RETURNING id, name, COALESCE(description,''), COALESCE(cover_url,''), sort_order, active, created_at, updated_at
`, collectionID, input.Name, normalizeCollectionName(input.Name), input.Description, input.CoverURL, input.SortOrder, input.Active)
	out, err := scanAdminCollection(row)
	if err != nil {
		return models.AdminCollection{}, fmt.Errorf("update collection: %w", err)
	}
	return out, nil
}

func (r *VideoRepository) DeleteCollection(ctx context.Context, collectionID uuid.UUID) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx delete collection: %w", err)
	}
	defer tx.Rollback(ctx)

	var detached int64
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM video_collections WHERE collection_id=$1`, collectionID).Scan(&detached); err != nil {
		return 0, fmt.Errorf("count collection relations: %w", err)
	}

	tag, err := tx.Exec(ctx, `DELETE FROM collections WHERE id=$1`, collectionID)
	if err != nil {
		return 0, fmt.Errorf("delete collection: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return 0, pgx.ErrNoRows
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit delete collection: %w", err)
	}
	return detached, nil
}

func dedupeCollectionIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return nil
	}
	out := make([]uuid.UUID, 0, len(ids))
	seen := map[uuid.UUID]struct{}{}
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (r *VideoRepository) ResolveCollectionIDs(ctx context.Context, collectionIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(collectionIDs) == 0 {
		return nil, nil
	}
	deduped := dedupeCollectionIDs(collectionIDs)

	rows, err := r.pool.Query(ctx, `SELECT id FROM collections WHERE id = ANY($1)`, deduped)
	if err != nil {
		return nil, fmt.Errorf("query collections: %w", err)
	}
	defer rows.Close()

	exists := map[uuid.UUID]struct{}{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan collection id: %w", err)
		}
		exists[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]uuid.UUID, 0, len(deduped))
	for _, id := range deduped {
		if _, ok := exists[id]; !ok {
			return nil, fmt.Errorf("合集不存在: %s", id.String())
		}
		out = append(out, id)
	}
	return out, nil
}

func (r *VideoRepository) ReplaceVideoCollectionsByIDs(ctx context.Context, videoID uuid.UUID, videoType string, collectionIDs []uuid.UUID) error {
	normalizedType := strings.ToLower(strings.TrimSpace(videoType))
	if len(collectionIDs) > 0 && normalizedType != "short" {
		return ErrCollectionsOnlyForShort
	}
	return r.replaceVideoCollectionsByIDs(ctx, videoID, collectionIDs)
}

func (r *VideoRepository) ReplaceVideoCollectionsByIDsAnyType(ctx context.Context, videoID uuid.UUID, collectionIDs []uuid.UUID) error {
	return r.replaceVideoCollectionsByIDs(ctx, videoID, collectionIDs)
}

func (r *VideoRepository) AddVideoCollectionsByIDsAnyType(ctx context.Context, videoID uuid.UUID, collectionIDs []uuid.UUID) error {
	resolvedIDs, err := r.ResolveCollectionIDs(ctx, collectionIDs)
	if err != nil {
		return err
	}
	if len(resolvedIDs) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx add video collections: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, collectionID := range resolvedIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO video_collections(video_id, collection_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, videoID, collectionID); err != nil {
			return fmt.Errorf("insert video collection: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit add video collections: %w", err)
	}
	return nil
}

func (r *VideoRepository) AddVideoCollectionsByIDs(ctx context.Context, videoID uuid.UUID, videoType string, collectionIDs []uuid.UUID) error {
	normalizedType := strings.ToLower(strings.TrimSpace(videoType))
	if len(collectionIDs) > 0 && normalizedType != "short" {
		return ErrCollectionsOnlyForShort
	}
	return r.AddVideoCollectionsByIDsAnyType(ctx, videoID, collectionIDs)
}

func (r *VideoRepository) replaceVideoCollectionsByIDs(ctx context.Context, videoID uuid.UUID, collectionIDs []uuid.UUID) error {
	resolvedIDs, err := r.ResolveCollectionIDs(ctx, collectionIDs)
	if err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx replace collections: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM video_collections WHERE video_id=$1`, videoID); err != nil {
		return fmt.Errorf("clear video collections: %w", err)
	}
	for _, collectionID := range resolvedIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO video_collections(video_id, collection_id) VALUES ($1,$2)`, videoID, collectionID); err != nil {
			return fmt.Errorf("insert video collection: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit replace collections: %w", err)
	}
	return nil
}

func (r *VideoRepository) ListVideoCollections(ctx context.Context, videoID uuid.UUID) ([]models.VideoCollection, error) {
	rows, err := r.pool.Query(ctx, `
SELECT c.id, c.name, COALESCE(c.cover_url,'')
FROM video_collections vc
JOIN collections c ON c.id = vc.collection_id
WHERE vc.video_id=$1
ORDER BY c.sort_order DESC, c.updated_at DESC, c.name ASC
`, videoID)
	if err != nil {
		return nil, fmt.Errorf("list video collections: %w", err)
	}
	defer rows.Close()

	out := make([]models.VideoCollection, 0, 8)
	for rows.Next() {
		var item models.VideoCollection
		if err := rows.Scan(&item.ID, &item.Name, &item.CoverURL); err != nil {
			return nil, fmt.Errorf("scan video collection: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *VideoRepository) ListVideoCollectionsForAdmin(ctx context.Context, videoID uuid.UUID) ([]models.AdminVideoCollection, error) {
	rows, err := r.pool.Query(ctx, `
SELECT c.id, c.name, COALESCE(c.cover_url,''), c.sort_order, c.active
FROM video_collections vc
JOIN collections c ON c.id = vc.collection_id
WHERE vc.video_id=$1
ORDER BY c.sort_order DESC, c.updated_at DESC, c.name ASC
`, videoID)
	if err != nil {
		return nil, fmt.Errorf("list video collections for admin: %w", err)
	}
	defer rows.Close()

	out := make([]models.AdminVideoCollection, 0, 8)
	for rows.Next() {
		var item models.AdminVideoCollection
		if err := rows.Scan(&item.ID, &item.Name, &item.CoverURL, &item.SortOrder, &item.Active); err != nil {
			return nil, fmt.Errorf("scan admin video collection: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *VideoRepository) ListCollectionsByVideoIDs(ctx context.Context, videoIDs []uuid.UUID) (map[uuid.UUID][]models.VideoCollection, error) {
	result := make(map[uuid.UUID][]models.VideoCollection, len(videoIDs))
	if len(videoIDs) == 0 {
		return result, nil
	}
	ids := dedupeCollectionVideoIDs(videoIDs)
	rows, err := r.pool.Query(ctx, `
SELECT vc.video_id, c.id, c.name, COALESCE(c.cover_url,'')
FROM video_collections vc
JOIN collections c ON c.id = vc.collection_id
WHERE vc.video_id = ANY($1)
ORDER BY c.sort_order DESC, c.updated_at DESC, c.name ASC
`, ids)
	if err != nil {
		return nil, fmt.Errorf("list collections by video ids: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var videoID uuid.UUID
		var item models.VideoCollection
		if err := rows.Scan(&videoID, &item.ID, &item.Name, &item.CoverURL); err != nil {
			return nil, fmt.Errorf("scan collections by video ids: %w", err)
		}
		result[videoID] = append(result[videoID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func dedupeCollectionVideoIDs(ids []uuid.UUID) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(ids))
	seen := map[uuid.UUID]struct{}{}
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// shortVideosByCollectionVisibilityClause 限定合集可见性：合集启用且至少有一条
// 可播放短视频（type='short' 且 status='ready'）。它与可见短视频合集定义一致，见 ADR-0017。
const shortVideosByCollectionVisibilityClause = `c.active = TRUE
  AND EXISTS (
    SELECT 1
    FROM video_collections vc
    JOIN videos v ON v.id = vc.video_id
    WHERE vc.collection_id = c.id
      AND v.type = 'short'
      AND v.status = 'ready'
  )`

// IsVisibleAppShortCollection 校验合集当前是否允许从手机端目录发起新投屏会话。
// 已建立会话的后续补页不调用此方法，因此合集下线不会中断正在播放的会话。
func (r *VideoRepository) IsVisibleAppShortCollection(ctx context.Context, collectionID uuid.UUID) (bool, error) {
	var visible bool
	if err := r.pool.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1
  FROM collections c
  WHERE c.id = $1
    AND `+shortVideosByCollectionVisibilityClause+`
)`, collectionID).Scan(&visible); err != nil {
		return false, fmt.Errorf("check app short collection visibility: %w", err)
	}
	return visible, nil
}

// ListAppShortCollections 返回手机端可见的短视频合集目录，按后台 sort_order DESC、
// updated_at DESC 排序。每条返回最终封面 URL 与可播放短视频数量；封面优先使用合集
// cover_url，未设置时回退到合集内最新可播放短视频的缩略图（见短视频合集封面回退）。
func (r *VideoRepository) ListAppShortCollections(ctx context.Context, limit, offset int) ([]models.ShortCollectionListItem, int, error) {
	const where = shortVideosByCollectionVisibilityClause

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM collections c WHERE `+where).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count app short collections: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
SELECT
  c.id,
  c.name,
  COALESCE(c.description, ''),
  COALESCE(c.cover_url, ''),
  preview.cover_video_id,
  preview.playable_count,
  c.updated_at
FROM collections c
JOIN LATERAL (
  SELECT
    COUNT(*)::INT AS playable_count,
    (
      SELECT v.id
      FROM video_collections vc
      JOIN videos v ON v.id = vc.video_id
      WHERE vc.collection_id = c.id
        AND v.type = 'short'
        AND v.status = 'ready'
      ORDER BY v.created_at DESC
      LIMIT 1
    ) AS cover_video_id
  FROM video_collections vc
  JOIN videos v ON v.id = vc.video_id
  WHERE vc.collection_id = c.id
    AND v.type = 'short'
    AND v.status = 'ready'
) preview ON preview.playable_count > 0
WHERE `+where+`
ORDER BY c.sort_order DESC, c.updated_at DESC, c.name ASC
LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list app short collections: %w", err)
	}
	defer rows.Close()

	items := make([]models.ShortCollectionListItem, 0, limit)
	for rows.Next() {
		var (
			item         models.ShortCollectionListItem
			coverURL     string
			coverVideoID uuid.UUID
		)
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &coverURL, &coverVideoID, &item.PlayableCount, &item.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan app short collection: %w", err)
		}
		item.CoverURL = resolveAppShortCollectionCoverURL(coverURL, coverVideoID)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// resolveAppShortCollectionCoverURL 解析合集最终封面：管理员设置的 cover_url 非空时直接使用；
// 否则回退到合集内最新可播放短视频的缩略图 URL；两者都不可用时返回空字符串（由前端占位图兜底）。
func resolveAppShortCollectionCoverURL(coverURL string, coverVideoID uuid.UUID) string {
	if trimmed := strings.TrimSpace(coverURL); trimmed != "" {
		return trimmed
	}
	if coverVideoID == uuid.Nil {
		return ""
	}
	return utils.VideoThumbnailURL(coverVideoID)
}
