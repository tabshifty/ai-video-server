package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

var (
	ErrTelegramAccountStateNotFound  = errors.New("Telegram 账号状态不存在")
	ErrTelegramHeartbeatNotFound     = errors.New("Telegram 采集器心跳不存在")
	ErrTelegramAuthorizationNotFound = errors.New("Telegram 授权记录不存在")
)

// GetTelegramAccountState returns the singleton account metadata.
func (r *VideoRepository) GetTelegramAccountState(ctx context.Context) (models.TelegramAccountState, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, telegram_user_id, COALESCE(username, ''), COALESCE(first_name, ''),
       COALESCE(last_name, ''), COALESCE(phone_masked, ''), status,
       COALESCE(last_error, ''), authorized_at, last_seen_at, created_at, updated_at
FROM telegram_account_state
WHERE id = 1`)
	item, err := scanTelegramAccountState(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.TelegramAccountState{}, ErrTelegramAccountStateNotFound
	}
	if err != nil {
		return models.TelegramAccountState{}, fmt.Errorf("读取 Telegram 账号状态: %w", err)
	}
	return item, nil
}

// UpsertTelegramAccountState replaces non-sensitive singleton account metadata.
func (r *VideoRepository) UpsertTelegramAccountState(ctx context.Context, item models.TelegramAccountState) error {
	item.ID = 1
	if item.Status == "" {
		item.Status = models.TelegramAccountStatusUnconfigured
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now().UTC()
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	_, err := r.pool.Exec(ctx, `
INSERT INTO telegram_account_state (
    id, telegram_user_id, username, first_name, last_name, phone_masked,
    status, last_error, authorized_at, last_seen_at, created_at, updated_at
)
VALUES (1, $1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, $9, $10, $11)
ON CONFLICT (id) DO UPDATE SET
    telegram_user_id = EXCLUDED.telegram_user_id,
    username = EXCLUDED.username,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    phone_masked = EXCLUDED.phone_masked,
    status = EXCLUDED.status,
    last_error = EXCLUDED.last_error,
    authorized_at = EXCLUDED.authorized_at,
    last_seen_at = EXCLUDED.last_seen_at,
    updated_at = EXCLUDED.updated_at`, item.TelegramUserID, strings.TrimSpace(item.Username),
		strings.TrimSpace(item.FirstName), strings.TrimSpace(item.LastName), strings.TrimSpace(item.PhoneMasked),
		item.Status, strings.TrimSpace(item.LastError), item.AuthorizedAt, item.LastSeenAt,
		item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("写入 Telegram 账号状态: %w", err)
	}
	return nil
}

// GetTelegramIngestorHeartbeat returns the singleton collector heartbeat.
func (r *VideoRepository) GetTelegramIngestorHeartbeat(ctx context.Context) (models.TelegramIngestorHeartbeat, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, status, account_status, COALESCE(version, ''), COALESCE(error_summary, ''),
       last_seen_at, created_at, updated_at
FROM telegram_ingestor_heartbeats
WHERE id = 1`)
	item, err := scanTelegramHeartbeat(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.TelegramIngestorHeartbeat{}, ErrTelegramHeartbeatNotFound
	}
	if err != nil {
		return models.TelegramIngestorHeartbeat{}, fmt.Errorf("读取 Telegram 采集器心跳: %w", err)
	}
	return item, nil
}

// UpsertTelegramIngestorHeartbeat records one collector liveness observation.
func (r *VideoRepository) UpsertTelegramIngestorHeartbeat(ctx context.Context, item models.TelegramIngestorHeartbeat) error {
	item.ID = 1
	if item.Status == "" {
		item.Status = models.TelegramIngestorStatusRunning
	}
	if item.AccountStatus == "" {
		item.AccountStatus = models.TelegramAccountStatusUnconfigured
	}
	if item.LastSeenAt.IsZero() {
		item.LastSeenAt = time.Now().UTC()
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = item.LastSeenAt
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.LastSeenAt
	}
	_, err := r.pool.Exec(ctx, `
INSERT INTO telegram_ingestor_heartbeats (
    id, status, account_status, version, error_summary, last_seen_at, created_at, updated_at
)
VALUES (1, $1, $2, $3, NULLIF($4, ''), $5, $6, $7)
ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status,
    account_status = EXCLUDED.account_status,
    version = EXCLUDED.version,
    error_summary = EXCLUDED.error_summary,
    last_seen_at = EXCLUDED.last_seen_at,
    updated_at = EXCLUDED.updated_at`, item.Status, item.AccountStatus, strings.TrimSpace(item.Version),
		strings.TrimSpace(item.ErrorSummary), item.LastSeenAt, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("写入 Telegram 采集器心跳: %w", err)
	}
	return nil
}

// CreateTelegramAuthorization inserts a short-lived authorization record.
func (r *VideoRepository) CreateTelegramAuthorization(ctx context.Context, item models.TelegramAuthorization) error {
	if item.ID == uuid.Nil || item.ActorUserID == uuid.Nil {
		return errors.New("Telegram 授权缺少标识或操作者")
	}
	if item.ExpiresAt.IsZero() {
		return errors.New("Telegram 授权缺少失效时间")
	}
	if item.StartedAt.IsZero() {
		item.StartedAt = time.Now().UTC()
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = item.StartedAt
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.StartedAt
	}
	if item.Status == "" {
		item.Status = models.TelegramAuthorizationStatusPending
	}
	_, err := r.pool.Exec(ctx, `
INSERT INTO telegram_authorizations (
    id, kind, status, actor_user_id, telegram_user_id, error_summary,
    started_at, expires_at, finished_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, $9, $10, $11)`, item.ID,
		item.Kind, item.Status, item.ActorUserID, item.TelegramUserID, strings.TrimSpace(item.ErrorSummary),
		item.StartedAt, item.ExpiresAt, item.FinishedAt, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("创建 Telegram 授权记录: %w", err)
	}
	return nil
}

// GetTelegramAuthorization reads one public authorization lifecycle record.
func (r *VideoRepository) GetTelegramAuthorization(ctx context.Context, id uuid.UUID) (models.TelegramAuthorization, error) {
	row := r.pool.QueryRow(ctx, telegramAuthorizationSelect+` WHERE id = $1`, id)
	item, err := scanTelegramAuthorization(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.TelegramAuthorization{}, fmt.Errorf("%w: %s", ErrTelegramAuthorizationNotFound, id)
	}
	if err != nil {
		return models.TelegramAuthorization{}, fmt.Errorf("读取 Telegram 授权记录: %w", err)
	}
	return item, nil
}

// UpdateTelegramAuthorization updates only public lifecycle fields.
func (r *VideoRepository) UpdateTelegramAuthorization(ctx context.Context, id uuid.UUID, status, errorSummary string, telegramUserID *int64, finishedAt *time.Time) error {
	if id == uuid.Nil {
		return ErrTelegramAuthorizationNotFound
	}
	_, err := r.pool.Exec(ctx, `
UPDATE telegram_authorizations
SET status = $2, error_summary = NULLIF($3, ''), telegram_user_id = $4,
    finished_at = $5, updated_at = NOW()
WHERE id = $1`, id, strings.TrimSpace(status), strings.TrimSpace(errorSummary), telegramUserID, finishedAt)
	if err != nil {
		return fmt.Errorf("更新 Telegram 授权记录: %w", err)
	}
	return nil
}

// CreateTelegramAuditLog appends a sanitized administrative event.
func (r *VideoRepository) CreateTelegramAuditLog(ctx context.Context, item models.TelegramAuditLog) error {
	if item.ActorUserID == uuid.Nil || strings.TrimSpace(item.Action) == "" {
		return errors.New("Telegram 审计缺少操作者或动作")
	}
	if len(item.Summary) == 0 {
		item.Summary = json.RawMessage(`{}`)
	}
	var object map[string]any
	if err := json.Unmarshal(item.Summary, &object); err != nil || object == nil {
		return errors.New("Telegram 审计摘要必须是 JSON 对象")
	}
	if item.Result == "" {
		item.Result = "succeeded"
	}
	_, err := r.pool.Exec(ctx, `
INSERT INTO telegram_audit_logs (
    actor_user_id, action, target_type, target_id, result, summary, created_at
)
VALUES ($1, $2, $3, $4, $5, $6::jsonb, COALESCE($7, NOW()))`, item.ActorUserID,
		strings.TrimSpace(item.Action), strings.TrimSpace(item.TargetType), strings.TrimSpace(item.TargetID),
		strings.TrimSpace(item.Result), item.Summary, nullableTime(item.CreatedAt))
	if err != nil {
		return fmt.Errorf("写入 Telegram 审计: %w", err)
	}
	return nil
}

// ListTelegramAuditLogs returns a paginated, read-only audit slice.
func (r *VideoRepository) ListTelegramAuditLogs(ctx context.Context, filter models.TelegramAuditFilter) ([]models.TelegramAuditLog, int, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	where := []string{"TRUE"}
	args := make([]any, 0, 6)
	add := func(expression string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(expression, len(args)))
	}
	if action := strings.TrimSpace(filter.Action); action != "" {
		add("action = $%d", action)
	}
	if result := strings.TrimSpace(filter.Result); result != "" {
		add("result = $%d", result)
	}
	if filter.StartTime != nil {
		add("created_at >= $%d", filter.StartTime)
	}
	if filter.EndTime != nil {
		add("created_at < $%d", filter.EndTime)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM telegram_audit_logs WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("统计 Telegram 审计: %w", err)
	}
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	rows, err := r.pool.Query(ctx, `
SELECT id, actor_user_id, action, target_type, target_id, result, summary, created_at
FROM telegram_audit_logs
WHERE `+whereSQL+` ORDER BY created_at DESC, id DESC LIMIT $`+fmt.Sprint(len(args)-1)+` OFFSET $`+fmt.Sprint(len(args)), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询 Telegram 审计: %w", err)
	}
	defer rows.Close()
	items := make([]models.TelegramAuditLog, 0, pageSize)
	for rows.Next() {
		var item models.TelegramAuditLog
		if err := rows.Scan(&item.ID, &item.ActorUserID, &item.Action, &item.TargetType, &item.TargetID, &item.Result, &item.Summary, &item.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("读取 Telegram 审计: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("遍历 Telegram 审计: %w", err)
	}
	return items, total, nil
}

const telegramAuthorizationSelect = `SELECT id, kind, status, actor_user_id, telegram_user_id,
       COALESCE(error_summary, ''), started_at, expires_at, finished_at, created_at, updated_at
FROM telegram_authorizations`

func scanTelegramAccountState(row interface{ Scan(...any) error }) (models.TelegramAccountState, error) {
	var item models.TelegramAccountState
	var userID sql.NullInt64
	var authorizedAt, lastSeenAt sql.NullTime
	err := row.Scan(&item.ID, &userID, &item.Username, &item.FirstName, &item.LastName, &item.PhoneMasked,
		&item.Status, &item.LastError, &authorizedAt, &lastSeenAt, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return models.TelegramAccountState{}, err
	}
	if userID.Valid {
		value := userID.Int64
		item.TelegramUserID = &value
	}
	item.AuthorizedAt = nullTimePtr(authorizedAt)
	item.LastSeenAt = nullTimePtr(lastSeenAt)
	return item, nil
}

func scanTelegramHeartbeat(row interface{ Scan(...any) error }) (models.TelegramIngestorHeartbeat, error) {
	var item models.TelegramIngestorHeartbeat
	if err := row.Scan(&item.ID, &item.Status, &item.AccountStatus, &item.Version, &item.ErrorSummary,
		&item.LastSeenAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return models.TelegramIngestorHeartbeat{}, err
	}
	return item, nil
}

func scanTelegramAuthorization(row interface{ Scan(...any) error }) (models.TelegramAuthorization, error) {
	var item models.TelegramAuthorization
	var telegramUserID sql.NullInt64
	var errorSummary sql.NullString
	var finishedAt sql.NullTime
	if err := row.Scan(&item.ID, &item.Kind, &item.Status, &item.ActorUserID, &telegramUserID, &errorSummary,
		&item.StartedAt, &item.ExpiresAt, &finishedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return models.TelegramAuthorization{}, err
	}
	if telegramUserID.Valid {
		value := telegramUserID.Int64
		item.TelegramUserID = &value
	}
	if errorSummary.Valid {
		item.ErrorSummary = errorSummary.String
	}
	item.FinishedAt = nullTimePtr(finishedAt)
	return item, nil
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}
