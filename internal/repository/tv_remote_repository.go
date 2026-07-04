package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

func (r *VideoRepository) ListUserTVDevices(ctx context.Context, userID uuid.UUID, platform string) ([]models.TvDeviceRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT
    id,
    device_id,
    device_name,
    platform,
    user_id,
    last_authorized_at,
    last_seen_at,
    created_at,
    updated_at
FROM tv_devices
WHERE user_id = $1
  AND platform = $2
ORDER BY last_authorized_at DESC NULLS LAST, updated_at DESC, device_name ASC
`, userID, strings.TrimSpace(platform))
	if err != nil {
		return nil, fmt.Errorf("list user tv devices: %w", err)
	}
	defer rows.Close()

	items := make([]models.TvDeviceRecord, 0)
	for rows.Next() {
		var item models.TvDeviceRecord
		if err := rows.Scan(
			&item.ID,
			&item.DeviceID,
			&item.DeviceName,
			&item.Platform,
			&item.UserID,
			&item.LastAuthorizedAt,
			&item.LastSeenAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user tv device: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user tv devices: %w", err)
	}
	return items, nil
}

func (r *VideoRepository) GetTVDeviceByDeviceIDAndUser(ctx context.Context, userID uuid.UUID, deviceID string, platform string) (models.TvDeviceRecord, error) {
	var item models.TvDeviceRecord
	err := r.pool.QueryRow(ctx, `
SELECT
    id,
    device_id,
    device_name,
    platform,
    user_id,
    last_authorized_at,
    last_seen_at,
    created_at,
    updated_at
FROM tv_devices
WHERE user_id = $1
  AND device_id = $2
  AND platform = $3
LIMIT 1
`, userID, strings.TrimSpace(deviceID), strings.TrimSpace(platform)).Scan(
		&item.ID,
		&item.DeviceID,
		&item.DeviceName,
		&item.Platform,
		&item.UserID,
		&item.LastAuthorizedAt,
		&item.LastSeenAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return models.TvDeviceRecord{}, fmt.Errorf("get tv device by device id and user: %w", err)
	}
	return item, nil
}

func (r *VideoRepository) TouchTVDeviceSeen(ctx context.Context, userID uuid.UUID, deviceID string, platform string, seenAt time.Time) error {
	tag, err := r.pool.Exec(ctx, `
UPDATE tv_devices
SET last_seen_at = $4,
    updated_at = NOW()
WHERE user_id = $1
  AND device_id = $2
  AND platform = $3
`, userID, strings.TrimSpace(deviceID), strings.TrimSpace(platform), seenAt)
	if err != nil {
		return fmt.Errorf("touch tv device seen: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *VideoRepository) CreateOrReplaceTVRemoteSession(ctx context.Context, session models.TvRemoteSession, now time.Time) error {
	itemsJSON, err := json.Marshal(session.Items)
	if err != nil {
		return fmt.Errorf("marshal tv remote session items: %w", err)
	}
	var searchContextJSON []byte
	if session.SearchContext != nil {
		searchContextJSON, err = json.Marshal(session.SearchContext)
		if err != nil {
			return fmt.Errorf("marshal tv remote session search context: %w", err)
		}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin create tv remote session tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var deviceRowID uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT id
FROM tv_devices
WHERE user_id = $1
  AND device_id = $2
  AND platform = $3
FOR UPDATE
`, session.UserID, session.DeviceID, session.Platform).Scan(&deviceRowID); err != nil {
		return fmt.Errorf("lock tv device for remote session: %w", err)
	}

	if _, err := tx.Exec(ctx, `
UPDATE tv_remote_sessions
SET status = 'ended',
    ended_reason = 'replaced',
    ended_at = $1,
    updated_at = $1
WHERE device_id = $2
  AND platform = $3
  AND status = 'active'
`, now, session.DeviceID, session.Platform); err != nil {
		return fmt.Errorf("replace active tv remote session: %w", err)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO tv_remote_sessions (
    id, user_id, device_id, platform, status, items, search_context, current_index, current_video_id, ended_reason, ended_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8, $9, $10, $11, $12, $12)
`, session.ID, session.UserID, session.DeviceID, session.Platform, session.Status, itemsJSON, searchContextJSON, session.CurrentIndex, session.CurrentVideoID, session.EndedReason, session.EndedAt, now); err != nil {
		return fmt.Errorf("insert tv remote session: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit create tv remote session tx: %w", err)
	}
	return nil
}

func (r *VideoRepository) GetTVRemoteSessionByID(ctx context.Context, sessionID uuid.UUID) (models.TvRemoteSession, error) {
	return scanTVRemoteSessionRow(r.pool.QueryRow(ctx, `
SELECT
    s.id,
    s.user_id,
    s.device_id,
    COALESCE(d.device_name, ''),
	s.platform,
	s.status,
	s.items,
	s.search_context,
	s.current_index,
	s.current_video_id,
	s.ended_reason,
    s.ended_at,
    s.created_at,
    s.updated_at
FROM tv_remote_sessions s
LEFT JOIN tv_devices d
  ON d.device_id = s.device_id
 AND d.platform = s.platform
WHERE s.id = $1
LIMIT 1
`, sessionID))
}

func (r *VideoRepository) GetActiveTVRemoteSessionForDevice(ctx context.Context, userID uuid.UUID, deviceID string, platform string) (models.TvRemoteSession, error) {
	return scanTVRemoteSessionRow(r.pool.QueryRow(ctx, `
SELECT
    s.id,
    s.user_id,
    s.device_id,
    COALESCE(d.device_name, ''),
	s.platform,
	s.status,
	s.items,
	s.search_context,
	s.current_index,
	s.current_video_id,
	s.ended_reason,
    s.ended_at,
    s.created_at,
    s.updated_at
FROM tv_remote_sessions s
LEFT JOIN tv_devices d
  ON d.device_id = s.device_id
 AND d.platform = s.platform
WHERE s.user_id = $1
  AND s.device_id = $2
  AND s.platform = $3
  AND s.status = 'active'
ORDER BY s.updated_at DESC
LIMIT 1
`, userID, strings.TrimSpace(deviceID), strings.TrimSpace(platform)))
}

func (r *VideoRepository) UpdateTVRemoteSessionState(
	ctx context.Context,
	sessionID uuid.UUID,
	userID uuid.UUID,
	items []models.TvRemoteSessionItem,
	currentIndex int,
	currentVideoID uuid.UUID,
	searchContext *models.TvRemoteSearchContext,
	now time.Time,
) error {
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("marshal tv remote session items: %w", err)
	}
	var searchContextJSON []byte
	if searchContext != nil {
		searchContextJSON, err = json.Marshal(searchContext)
		if err != nil {
			return fmt.Errorf("marshal tv remote session search context: %w", err)
		}
	}
	tag, err := r.pool.Exec(ctx, `
UPDATE tv_remote_sessions
SET items = $3::jsonb,
    search_context = $4::jsonb,
    current_index = $5,
    current_video_id = $6,
    updated_at = $7
WHERE id = $1
  AND user_id = $2
  AND status = 'active'
`, sessionID, userID, itemsJSON, searchContextJSON, currentIndex, currentVideoID, now)
	if err != nil {
		return fmt.Errorf("update tv remote session state: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *VideoRepository) EndTVRemoteSession(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, endedReason string, endedAt time.Time) error {
	tag, err := r.pool.Exec(ctx, `
UPDATE tv_remote_sessions
SET status = 'ended',
    ended_reason = $3,
    ended_at = $4,
    updated_at = $4
WHERE id = $1
  AND user_id = $2
  AND status = 'active'
`, sessionID, userID, strings.TrimSpace(endedReason), endedAt)
	if err != nil {
		return fmt.Errorf("end tv remote session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

type tvRemoteSessionScanner interface {
	Scan(dest ...any) error
}

func scanTVRemoteSessionRow(row tvRemoteSessionScanner) (models.TvRemoteSession, error) {
	var session models.TvRemoteSession
	var rawItems []byte
	var rawSearchContext []byte
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.DeviceID,
		&session.DeviceName,
		&session.Platform,
		&session.Status,
		&rawItems,
		&rawSearchContext,
		&session.CurrentIndex,
		&session.CurrentVideoID,
		&session.EndedReason,
		&session.EndedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return models.TvRemoteSession{}, fmt.Errorf("scan tv remote session: %w", err)
	}
	if len(rawItems) == 0 {
		session.Items = []models.TvRemoteSessionItem{}
		return session, nil
	}
	if err := json.Unmarshal(rawItems, &session.Items); err != nil {
		return models.TvRemoteSession{}, fmt.Errorf("unmarshal tv remote session items: %w", err)
	}
	if len(rawSearchContext) > 0 && string(rawSearchContext) != "null" {
		var searchContext models.TvRemoteSearchContext
		if err := json.Unmarshal(rawSearchContext, &searchContext); err != nil {
			return models.TvRemoteSession{}, fmt.Errorf("unmarshal tv remote session search context: %w", err)
		}
		session.SearchContext = &searchContext
	}
	return session, nil
}
