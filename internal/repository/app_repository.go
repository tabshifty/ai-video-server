package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"video-server/internal/models"
)

func (r *VideoRepository) GetVideoDetail(ctx context.Context, videoID, userID uuid.UUID) (models.VideoDetail, error) {
	var detail models.VideoDetail
	err := r.pool.QueryRow(ctx, `
SELECT id, title, description, transcoded_path, thumbnail_path, duration_seconds, views_count, likes_count, favorites_count, metadata
FROM videos
WHERE id=$1 AND status='ready'
`, videoID).Scan(
		&detail.ID,
		&detail.Title,
		&detail.Description,
		&detail.TranscodedPath,
		&detail.ThumbnailPath,
		&detail.Duration,
		&detail.ViewsCount,
		&detail.LikesCount,
		&detail.FavoritesCount,
		&detail.Metadata,
	)
	if err != nil {
		return models.VideoDetail{}, fmt.Errorf("get video detail: %w", err)
	}

	tagsRows, err := r.pool.Query(ctx, `SELECT tag FROM video_tags WHERE video_id=$1 ORDER BY tag`, videoID)
	if err != nil {
		return models.VideoDetail{}, fmt.Errorf("query tags: %w", err)
	}
	defer tagsRows.Close()
	tags := make([]string, 0)
	for tagsRows.Next() {
		var tag string
		if err := tagsRows.Scan(&tag); err != nil {
			return models.VideoDetail{}, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	detail.Tags = tags

	rows, err := r.pool.Query(ctx, `
SELECT action_type, watch_seconds
FROM user_video_actions
WHERE user_id=$1 AND video_id=$2 AND action_type IN ('like','favorite','dislike','view','completed')
`, userID, videoID)
	if err != nil {
		return models.VideoDetail{}, fmt.Errorf("query user states: %w", err)
	}
	defer rows.Close()

	state := models.VideoUserState{}
	for rows.Next() {
		var action string
		var watch int
		if err := rows.Scan(&action, &watch); err != nil {
			return models.VideoDetail{}, fmt.Errorf("scan user state: %w", err)
		}
		switch action {
		case "like":
			state.IsLiked = true
		case "favorite":
			state.IsFavorited = true
		case "dislike":
			state.IsDisliked = true
		case "view":
			state.WatchSeconds = watch
		case "completed":
			state.IsCompleted = true
		}
	}
	detail.UserState = state

	if len(detail.Metadata) == 0 {
		detail.Metadata = []byte(`{}`)
	}
	return detail, nil
}

func (r *VideoRepository) UpsertViewHistory(ctx context.Context, userID, videoID uuid.UUID, watchSeconds int, completed bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var existed bool
	err = tx.QueryRow(ctx, `
SELECT EXISTS(
	SELECT 1 FROM user_video_actions
	WHERE user_id=$1 AND video_id=$2 AND action_type='view'
)`, userID, videoID).Scan(&existed)
	if err != nil {
		return fmt.Errorf("check existing view: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO user_video_actions(user_id, video_id, action_type, watch_seconds)
VALUES ($1,$2,'view',$3)
ON CONFLICT(user_id, video_id, action_type)
DO UPDATE SET watch_seconds=EXCLUDED.watch_seconds, updated_at=NOW()
`, userID, videoID, watchSeconds)
	if err != nil {
		return fmt.Errorf("upsert view history: %w", err)
	}

	if !existed {
		if _, err := tx.Exec(ctx, `UPDATE videos SET views_count = views_count + 1 WHERE id=$1`, videoID); err != nil {
			return fmt.Errorf("increment views_count: %w", err)
		}
	}

	if completed {
		if _, err := tx.Exec(ctx, `
INSERT INTO user_video_actions(user_id, video_id, action_type, watch_seconds)
VALUES ($1,$2,'completed',$3)
ON CONFLICT(user_id, video_id, action_type)
DO UPDATE SET watch_seconds=EXCLUDED.watch_seconds, updated_at=NOW()
`, userID, videoID, watchSeconds); err != nil {
			return fmt.Errorf("upsert completed: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit history tx: %w", err)
	}
	return nil
}

func (r *VideoRepository) ContinueWatching(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.HistoryItem, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM user_video_actions a
JOIN videos v ON v.id = a.video_id
WHERE a.user_id=$1
  AND a.action_type='view'
  AND a.watch_seconds > 0
  AND a.watch_seconds < GREATEST(v.duration_seconds - 5, 0)
  AND NOT EXISTS (
      SELECT 1 FROM user_video_actions c
      WHERE c.user_id = a.user_id
        AND c.video_id = a.video_id
        AND c.action_type='completed'
  )
`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count continue watching: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
SELECT v.id, v.title, v.thumbnail_path, v.duration_seconds, a.watch_seconds, a.updated_at
FROM user_video_actions a
JOIN videos v ON v.id = a.video_id
WHERE a.user_id=$1
  AND a.action_type='view'
  AND a.watch_seconds > 0
  AND a.watch_seconds < GREATEST(v.duration_seconds - 5, 0)
  AND NOT EXISTS (
      SELECT 1 FROM user_video_actions c
      WHERE c.user_id = a.user_id
        AND c.video_id = a.video_id
        AND c.action_type='completed'
  )
ORDER BY a.updated_at DESC
LIMIT $2 OFFSET $3
`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query continue watching: %w", err)
	}
	defer rows.Close()

	items := make([]models.HistoryItem, 0, limit)
	for rows.Next() {
		var item models.HistoryItem
		if err := rows.Scan(&item.VideoID, &item.Title, &item.ThumbnailPath, &item.Duration, &item.WatchSeconds, &item.LastWatchedAt); err != nil {
			return nil, 0, fmt.Errorf("scan continue item: %w", err)
		}
		if item.Duration > 0 {
			item.Progress = float64(item.WatchSeconds) / float64(item.Duration)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate continue rows: %w", err)
	}
	return items, total, nil
}

func (r *VideoRepository) DeleteHistory(ctx context.Context, userID, videoID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
DELETE FROM user_video_actions
WHERE user_id=$1 AND video_id=$2 AND action_type IN ('view','completed')
`, userID, videoID)
	if err != nil {
		return fmt.Errorf("delete history: %w", err)
	}
	return nil
}

func (r *VideoRepository) ToggleAction(ctx context.Context, userID, videoID uuid.UUID, action string) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var existed bool
	err = tx.QueryRow(ctx, `
SELECT EXISTS(
	SELECT 1 FROM user_video_actions
	WHERE user_id=$1 AND video_id=$2 AND action_type=$3
)`, userID, videoID, action).Scan(&existed)
	if err != nil {
		return false, fmt.Errorf("check action exists: %w", err)
	}

	if action == "dislike" {
		var removedLikes int64
		var removedFavorites int64
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM user_video_actions WHERE user_id=$1 AND video_id=$2 AND action_type='like'`, userID, videoID).Scan(&removedLikes); err != nil {
			return false, fmt.Errorf("count removed likes: %w", err)
		}
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM user_video_actions WHERE user_id=$1 AND video_id=$2 AND action_type='favorite'`, userID, videoID).Scan(&removedFavorites); err != nil {
			return false, fmt.Errorf("count removed favorites: %w", err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM user_video_actions WHERE user_id=$1 AND video_id=$2 AND action_type IN ('like','favorite')`, userID, videoID); err != nil {
			return false, fmt.Errorf("delete opposite actions: %w", err)
		}
		if removedLikes > 0 {
			if _, err := tx.Exec(ctx, `UPDATE videos SET likes_count = GREATEST(likes_count - $2, 0) WHERE id=$1`, videoID, removedLikes); err != nil {
				return false, fmt.Errorf("decrement likes_count by opposite delete: %w", err)
			}
		}
		if removedFavorites > 0 {
			if _, err := tx.Exec(ctx, `UPDATE videos SET favorites_count = GREATEST(favorites_count - $2, 0) WHERE id=$1`, videoID, removedFavorites); err != nil {
				return false, fmt.Errorf("decrement favorites_count by opposite delete: %w", err)
			}
		}
	}

	if existed {
		if _, err := tx.Exec(ctx, `DELETE FROM user_video_actions WHERE user_id=$1 AND video_id=$2 AND action_type=$3`, userID, videoID, action); err != nil {
			return false, fmt.Errorf("delete action: %w", err)
		}
		if action == "like" {
			_, err = tx.Exec(ctx, `UPDATE videos SET likes_count = GREATEST(likes_count - 1, 0) WHERE id=$1`, videoID)
		} else if action == "favorite" {
			_, err = tx.Exec(ctx, `UPDATE videos SET favorites_count = GREATEST(favorites_count - 1, 0) WHERE id=$1`, videoID)
		}
		if err != nil {
			return false, fmt.Errorf("decrement counters: %w", err)
		}
	} else {
		if _, err := tx.Exec(ctx, `
INSERT INTO user_video_actions(user_id, video_id, action_type, watch_seconds)
VALUES ($1,$2,$3,0)
ON CONFLICT(user_id, video_id, action_type)
DO UPDATE SET updated_at=NOW()
`, userID, videoID, action); err != nil {
			return false, fmt.Errorf("insert action: %w", err)
		}
		if action == "like" {
			_, err = tx.Exec(ctx, `UPDATE videos SET likes_count = likes_count + 1 WHERE id=$1`, videoID)
		} else if action == "favorite" {
			_, err = tx.Exec(ctx, `UPDATE videos SET favorites_count = favorites_count + 1 WHERE id=$1`, videoID)
		}
		if err != nil {
			return false, fmt.Errorf("increment counters: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit toggle tx: %w", err)
	}
	return !existed, nil
}

func (r *VideoRepository) SearchVideos(ctx context.Context, q, typ string, limit, offset int) ([]models.VideoListItem, int, error) {
	keyword := "%" + strings.ToLower(strings.TrimSpace(q)) + "%"
	if keyword == "%%" {
		keyword = "%"
	}

	var total int
	countSQL := `
SELECT COUNT(DISTINCT v.id)
FROM videos v
LEFT JOIN video_tags vt ON vt.video_id=v.id
WHERE v.status='ready'
  AND ($1 = 'all' OR v.type = $1)
  AND (
	LOWER(v.title) LIKE $2 OR
	LOWER(COALESCE(v.description,'')) LIKE $2 OR
	LOWER(COALESCE(vt.tag,'')) LIKE $2
  )`
	if err := r.pool.QueryRow(ctx, countSQL, typ, keyword).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count search videos: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
SELECT DISTINCT v.id, v.title, v.type, v.thumbnail_path, v.transcoded_path, v.duration_seconds, v.created_at
FROM videos v
LEFT JOIN video_tags vt ON vt.video_id=v.id
WHERE v.status='ready'
  AND ($1 = 'all' OR v.type = $1)
  AND (
	LOWER(v.title) LIKE $2 OR
	LOWER(COALESCE(v.description,'')) LIKE $2 OR
	LOWER(COALESCE(vt.tag,'')) LIKE $2
  )
ORDER BY v.created_at DESC
LIMIT $3 OFFSET $4
`, typ, keyword, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query search videos: %w", err)
	}
	defer rows.Close()

	items := make([]models.VideoListItem, 0, limit)
	for rows.Next() {
		var item models.VideoListItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Type, &item.ThumbnailPath, &item.TranscodedPath, &item.Duration, &item.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan search item: %w", err)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *VideoRepository) GetUploadedVideos(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.VideoListItem, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM videos WHERE user_id=$1`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count uploaded videos: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
SELECT id, title, type, thumbnail_path, transcoded_path, duration_seconds, created_at
FROM videos
WHERE user_id=$1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query uploaded videos: %w", err)
	}
	defer rows.Close()

	items := make([]models.VideoListItem, 0, limit)
	for rows.Next() {
		var item models.VideoListItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Type, &item.ThumbnailPath, &item.TranscodedPath, &item.Duration, &item.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan uploaded item: %w", err)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *VideoRepository) GetActionVideos(ctx context.Context, userID uuid.UUID, action string, limit, offset int) ([]models.VideoListItem, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM user_video_actions a
WHERE a.user_id=$1 AND a.action_type=$2
`, userID, action).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count action videos: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
SELECT v.id, v.title, v.type, v.thumbnail_path, v.transcoded_path, v.duration_seconds, a.updated_at
FROM user_video_actions a
JOIN videos v ON v.id = a.video_id
WHERE a.user_id=$1 AND a.action_type=$2
ORDER BY a.updated_at DESC
LIMIT $3 OFFSET $4
`, userID, action, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query action videos: %w", err)
	}
	defer rows.Close()

	items := make([]models.VideoListItem, 0, limit)
	for rows.Next() {
		var item models.VideoListItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Type, &item.ThumbnailPath, &item.TranscodedPath, &item.Duration, &item.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan action item: %w", err)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *VideoRepository) UpdateUserProfile(ctx context.Context, userID uuid.UUID, oldPassword, newEmail, newPassword string) error {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return fmt.Errorf("invalid old password")
	}

	nextEmail := strings.ToLower(strings.TrimSpace(user.Email))
	if strings.TrimSpace(newEmail) != "" {
		nextEmail = strings.ToLower(strings.TrimSpace(newEmail))
	}
	nextHash := user.PasswordHash
	if strings.TrimSpace(newPassword) != "" {
		if len(newPassword) < 6 {
			return fmt.Errorf("new password length must be >= 6")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash new password: %w", err)
		}
		nextHash = string(hash)
	}

	_, err = r.pool.Exec(ctx, `
UPDATE users
SET email=$2, password_hash=$3, updated_at=NOW()
WHERE id=$1
`, userID, nextEmail, nextHash)
	if err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}
	return nil
}
