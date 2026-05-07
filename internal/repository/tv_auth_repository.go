package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

func (r *VideoRepository) CreateTVAuthSession(ctx context.Context, session models.TvAuthSession) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO tv_auth_sessions (
    id, pair_code, device_id, device_name, platform, status, expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`,
		session.ID,
		session.PairCode,
		session.DeviceID,
		session.DeviceName,
		session.Platform,
		session.Status,
		session.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert tv auth session: %w", err)
	}
	return nil
}

func (r *VideoRepository) GetTVAuthSession(ctx context.Context, sessionID uuid.UUID) (models.TvAuthSession, error) {
	var session models.TvAuthSession
	err := r.pool.QueryRow(ctx, `
SELECT
    id, pair_code, device_id, device_name, platform, status, user_id,
    access_token, refresh_token, approved_username, approved_role,
    approved_at, expires_at, created_at, updated_at
FROM tv_auth_sessions
WHERE id = $1
`, sessionID).Scan(
		&session.ID,
		&session.PairCode,
		&session.DeviceID,
		&session.DeviceName,
		&session.Platform,
		&session.Status,
		&session.UserID,
		&session.AccessToken,
		&session.RefreshToken,
		&session.ApprovedUsername,
		&session.ApprovedRole,
		&session.ApprovedAt,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return models.TvAuthSession{}, fmt.Errorf("get tv auth session: %w", err)
	}
	return session, nil
}

func (r *VideoRepository) UpdateTVAuthSessionExpired(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
UPDATE tv_auth_sessions
SET status = 'expired', updated_at = NOW()
WHERE id = $1 AND status = 'pending'
`, sessionID)
	if err != nil {
		return fmt.Errorf("expire tv auth session: %w", err)
	}
	return nil
}

func (r *VideoRepository) ApproveTVAuthSession(
	ctx context.Context,
	sessionID uuid.UUID,
	user models.User,
	tokens models.AuthTokens,
	approvedAt time.Time,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin approve tv auth session tx: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
UPDATE tv_auth_sessions
SET status = 'approved',
    user_id = $2,
    access_token = $3,
    refresh_token = $4,
    approved_username = $5,
    approved_role = $6,
    approved_at = $7,
    updated_at = NOW()
WHERE id = $1
  AND status = 'pending'
  AND expires_at > NOW()
`, sessionID, user.ID, tokens.AccessToken, tokens.RefreshToken, user.Username, user.Role, approvedAt)
	if err != nil {
		return fmt.Errorf("update approved tv auth session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO tv_devices (id, device_id, device_name, platform, user_id, last_authorized_at)
SELECT gen_random_uuid(), device_id, device_name, platform, $2, $3
FROM tv_auth_sessions
WHERE id = $1
ON CONFLICT (device_id, platform)
DO UPDATE SET
    device_name = EXCLUDED.device_name,
    user_id = EXCLUDED.user_id,
    last_authorized_at = EXCLUDED.last_authorized_at,
    updated_at = NOW()
`, sessionID, user.ID, approvedAt); err != nil {
		return fmt.Errorf("upsert tv device: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit approve tv auth session: %w", err)
	}
	return nil
}

func (r *VideoRepository) DenyTVAuthSession(ctx context.Context, sessionID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
UPDATE tv_auth_sessions
SET status = 'denied', updated_at = NOW()
WHERE id = $1
  AND status = 'pending'
  AND expires_at > NOW()
`, sessionID)
	if err != nil {
		return fmt.Errorf("deny tv auth session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
