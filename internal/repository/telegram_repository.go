package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"video-server/internal/models"
)

var (
	ErrTelegramSourceNotFound = errors.New("Telegram 来源不存在")
	ErrTelegramMediaNotFound  = errors.New("Telegram 媒体不存在")
)

// TelegramSourcePatch contains fields that may be changed without replacing a source.
type TelegramSourcePatch struct {
	ChatID                   *int64
	ClearChatID              bool
	Title                    *string
	Username                 *string
	Enabled                  *bool
	SyncStatus               *string
	HistoryCursorMessageID   *int64
	LastError                *string
	NextRetryAt              *time.Time
	ClearNextRetryAt         bool
	BackfillCompletedAt      *time.Time
	ClearBackfillCompletedAt bool
}

// TelegramMediaPatch contains fields that may be changed during ingestion.
type TelegramMediaPatch struct {
	DocumentID       *int64
	ClearDocumentID  bool
	DocumentDCID     *int
	ProcessingStatus *string
	TranscodeStatus  *string
	VideoID          *uuid.UUID
	ClearVideoID     bool
	SHA256           *string
	TempPath         *string
	Attempts         *int
	NextRetryAt      *time.Time
	ClearNextRetryAt bool
	LastError        *string
}

type telegramQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type telegramSourceRowScanner interface {
	Scan(dest ...any) error
}

type telegramMediaRowScanner interface {
	Scan(dest ...any) error
}

// CreateTelegramSource inserts an administrator-selected chat reference.
func (r *VideoRepository) CreateTelegramSource(ctx context.Context, source models.TelegramSource) error {
	return createTelegramSource(ctx, r.pool, source)
}

func createTelegramSource(ctx context.Context, db telegramQuerier, source models.TelegramSource) error {
	if source.ID == uuid.Nil {
		return fmt.Errorf("create Telegram source: missing id")
	}
	if strings.TrimSpace(source.ChatRef) == "" {
		return fmt.Errorf("create Telegram source: missing chat reference")
	}
	if source.HistoryCursorMessageID < 0 {
		return fmt.Errorf("create Telegram source: history cursor must be non-negative")
	}
	if source.SyncStatus == "" {
		source.SyncStatus = "pending"
	}
	if source.CreatedAt.IsZero() {
		source.CreatedAt = time.Now().UTC()
	}
	if source.UpdatedAt.IsZero() {
		source.UpdatedAt = source.CreatedAt
	}
	_, err := db.Exec(ctx, `
INSERT INTO telegram_sources (
    id, chat_id, chat_ref, title, username, enabled, sync_status,
    history_cursor_message_id, last_error, next_retry_at,
    backfill_completed_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, ''), $10, $11, $12, $13)
`, source.ID, nullableTelegramChatID(source.ChatID), strings.TrimSpace(source.ChatRef),
		strings.TrimSpace(source.Title), strings.TrimSpace(source.Username), source.Enabled,
		source.SyncStatus, source.HistoryCursorMessageID, source.LastError, source.NextRetryAt,
		source.BackfillCompletedAt, source.CreatedAt, source.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create Telegram source: %w", err)
	}
	return nil
}

// ListTelegramSources returns configured Telegram sources, optionally limited to enabled rows.
func (r *VideoRepository) ListTelegramSources(ctx context.Context, enabledOnly bool) ([]models.TelegramSource, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, COALESCE(chat_id, 0), chat_ref, COALESCE(title, ''), COALESCE(username, ''),
       enabled, sync_status, history_cursor_message_id, COALESCE(last_error, ''),
       next_retry_at, backfill_completed_at, created_at, updated_at
FROM telegram_sources
WHERE ($1 = FALSE OR enabled = TRUE)
ORDER BY created_at ASC, id ASC
`, enabledOnly)
	if err != nil {
		return nil, fmt.Errorf("list Telegram sources: %w", err)
	}
	defer rows.Close()

	items := make([]models.TelegramSource, 0)
	for rows.Next() {
		item, err := scanTelegramSource(rows)
		if err != nil {
			return nil, fmt.Errorf("scan Telegram source: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate Telegram sources: %w", err)
	}
	return items, nil
}

// GetTelegramSource returns one configured Telegram source.
func (r *VideoRepository) GetTelegramSource(ctx context.Context, sourceID uuid.UUID) (models.TelegramSource, error) {
	item, err := scanTelegramSource(r.pool.QueryRow(ctx, `
SELECT id, COALESCE(chat_id, 0), chat_ref, COALESCE(title, ''), COALESCE(username, ''),
       enabled, sync_status, history_cursor_message_id, COALESCE(last_error, ''),
       next_retry_at, backfill_completed_at, created_at, updated_at
FROM telegram_sources
WHERE id = $1
`, sourceID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.TelegramSource{}, fmt.Errorf("%w: %s", ErrTelegramSourceNotFound, sourceID)
	}
	if err != nil {
		return models.TelegramSource{}, fmt.Errorf("get Telegram source: %w", err)
	}
	return item, nil
}

// UpdateTelegramSource applies a partial source update and refreshes updated_at.
func (r *VideoRepository) UpdateTelegramSource(ctx context.Context, sourceID uuid.UUID, patch TelegramSourcePatch) error {
	return updateTelegramSource(ctx, r.pool, sourceID, patch)
}

func updateTelegramSource(ctx context.Context, db telegramQuerier, sourceID uuid.UUID, patch TelegramSourcePatch) error {
	sets := make([]string, 0, 12)
	args := []any{sourceID}
	addValue := func(expression string, value any) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", expression, len(args)))
	}

	if patch.ClearChatID {
		sets = append(sets, "chat_id = NULL")
	} else if patch.ChatID != nil {
		addValue("chat_id", nullableTelegramChatID(*patch.ChatID))
	}
	if patch.Title != nil {
		addValue("title", strings.TrimSpace(*patch.Title))
	}
	if patch.Username != nil {
		addValue("username", strings.TrimSpace(*patch.Username))
	}
	if patch.Enabled != nil {
		addValue("enabled", *patch.Enabled)
	}
	if patch.SyncStatus != nil {
		addValue("sync_status", strings.TrimSpace(*patch.SyncStatus))
	}
	if patch.HistoryCursorMessageID != nil {
		addValue("history_cursor_message_id", *patch.HistoryCursorMessageID)
	}
	if patch.LastError != nil {
		addValue("last_error", nullableTelegramString(*patch.LastError))
	}
	if patch.ClearNextRetryAt {
		sets = append(sets, "next_retry_at = NULL")
	} else if patch.NextRetryAt != nil {
		addValue("next_retry_at", patch.NextRetryAt)
	}
	if patch.ClearBackfillCompletedAt {
		sets = append(sets, "backfill_completed_at = NULL")
	} else if patch.BackfillCompletedAt != nil {
		addValue("backfill_completed_at", patch.BackfillCompletedAt)
	}
	if len(sets) == 0 {
		return nil
	}
	sets = append(sets, "updated_at = NOW()")
	result, err := db.Exec(ctx, "UPDATE telegram_sources SET "+strings.Join(sets, ", ")+" WHERE id = $1", args...)
	if err != nil {
		return fmt.Errorf("update Telegram source: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", ErrTelegramSourceNotFound, sourceID)
	}
	return nil
}

// InsertTelegramMedia records a message idempotently and returns whether it was newly inserted.
func (r *VideoRepository) InsertTelegramMedia(ctx context.Context, media models.TelegramMedia) (models.TelegramMedia, bool, error) {
	return insertTelegramMedia(ctx, r.pool, media)
}

func insertTelegramMedia(ctx context.Context, db telegramQuerier, media models.TelegramMedia) (models.TelegramMedia, bool, error) {
	if media.ID == uuid.Nil {
		return models.TelegramMedia{}, false, fmt.Errorf("insert Telegram media: missing id")
	}
	if media.SourceID == uuid.Nil {
		return models.TelegramMedia{}, false, fmt.Errorf("insert Telegram media: missing source id")
	}
	if media.MessageID < 0 {
		return models.TelegramMedia{}, false, fmt.Errorf("insert Telegram media: message id must be non-negative")
	}
	if media.FileSize < 0 {
		return models.TelegramMedia{}, false, fmt.Errorf("insert Telegram media: file size must be non-negative")
	}
	if media.ProcessingStatus == "" {
		media.ProcessingStatus = "discovered"
	}
	if media.TranscodeStatus == "" {
		media.TranscodeStatus = "not_required"
	}
	if media.MessageCreatedAt.IsZero() {
		media.MessageCreatedAt = time.Now().UTC()
	}
	if media.CreatedAt.IsZero() {
		media.CreatedAt = time.Now().UTC()
	}
	if media.UpdatedAt.IsZero() {
		media.UpdatedAt = media.CreatedAt
	}

	row := db.QueryRow(ctx, `
INSERT INTO telegram_media (
    id, source_id, chat_id, message_id, telegram_document_id, document_dc_id,
    filename, mime_type, file_size, caption, message_url, message_created_at,
    processing_status, transcode_status, video_id, sha256, temp_path, attempts,
    next_retry_at, last_error, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
        NULLIF($16, ''), NULLIF($17, ''), $18, $19, NULLIF($20, ''), $21, $22)
ON CONFLICT (source_id, message_id) DO NOTHING
RETURNING id, source_id, chat_id, message_id, telegram_document_id, document_dc_id,
          filename, mime_type, file_size, caption, message_url, message_created_at,
          processing_status, transcode_status, video_id, sha256, temp_path, attempts,
          next_retry_at, last_error, created_at, updated_at
`, media.ID, media.SourceID, media.ChatID, media.MessageID, media.DocumentID, media.DocumentDCID,
		strings.TrimSpace(media.Filename), strings.TrimSpace(media.MIMEType), media.FileSize,
		media.Caption, media.MessageURL, media.MessageCreatedAt, media.ProcessingStatus,
		media.TranscodeStatus, media.VideoID, media.SHA256, media.TempPath, media.Attempts,
		media.NextRetryAt, media.LastError, media.CreatedAt, media.UpdatedAt)
	stored, err := scanTelegramMedia(row)
	if err == nil {
		return stored, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return models.TelegramMedia{}, false, fmt.Errorf("insert Telegram media: %w", err)
	}

	stored, err = getTelegramMediaByMessage(ctx, db, media.SourceID, media.MessageID)
	if err != nil {
		return models.TelegramMedia{}, false, fmt.Errorf("get existing Telegram media: %w", err)
	}
	return stored, false, nil
}

// InsertTelegramMediaAndAdvanceCursor atomically registers one message and advances its source cursor.
func (r *VideoRepository) InsertTelegramMediaAndAdvanceCursor(ctx context.Context, media models.TelegramMedia, cursor int64) (models.TelegramMedia, bool, error) {
	return insertTelegramMediaAndAdvanceCursor(ctx, r.pool, media, cursor)
}

func insertTelegramMediaAndAdvanceCursor(ctx context.Context, pool telegramPoolBeginner, media models.TelegramMedia, cursor int64) (models.TelegramMedia, bool, error) {
	if cursor < 0 {
		return models.TelegramMedia{}, false, fmt.Errorf("advance Telegram cursor: cursor must be non-negative")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return models.TelegramMedia{}, false, fmt.Errorf("begin Telegram media transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	stored, inserted, err := insertTelegramMedia(ctx, tx, media)
	if err != nil {
		return models.TelegramMedia{}, false, err
	}
	if err := advanceTelegramSourceCursor(ctx, tx, media.SourceID, cursor); err != nil {
		return models.TelegramMedia{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.TelegramMedia{}, false, fmt.Errorf("commit Telegram media transaction: %w", err)
	}
	return stored, inserted, nil
}

// AdvanceTelegramSourceCursor advances a history cursor when a page contains
// no importable video messages.
func (r *VideoRepository) AdvanceTelegramSourceCursor(ctx context.Context, sourceID uuid.UUID, cursor int64) error {
	return advanceTelegramSourceCursor(ctx, r.pool, sourceID, cursor)
}

func advanceTelegramSourceCursor(ctx context.Context, db telegramQuerier, sourceID uuid.UUID, cursor int64) error {
	if sourceID == uuid.Nil {
		return fmt.Errorf("advance Telegram source cursor: missing source id")
	}
	if cursor < 0 {
		return fmt.Errorf("advance Telegram source cursor: cursor must be non-negative")
	}
	result, err := db.Exec(ctx, `
UPDATE telegram_sources
SET history_cursor_message_id = CASE
    WHEN history_cursor_message_id = 0 THEN $2
    ELSE LEAST(history_cursor_message_id, $2)
END, updated_at = NOW()
WHERE id = $1
`, sourceID, cursor)
	if err != nil {
		return fmt.Errorf("advance Telegram source cursor: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", ErrTelegramSourceNotFound, sourceID)
	}
	return nil
}

type telegramPoolBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// GetTelegramMedia returns one message record.
func (r *VideoRepository) GetTelegramMedia(ctx context.Context, mediaID uuid.UUID) (models.TelegramMedia, error) {
	item, err := scanTelegramMedia(r.pool.QueryRow(ctx, telegramMediaSelect+" WHERE id = $1", mediaID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.TelegramMedia{}, fmt.Errorf("%w: %s", ErrTelegramMediaNotFound, mediaID)
	}
	if err != nil {
		return models.TelegramMedia{}, fmt.Errorf("get Telegram media: %w", err)
	}
	return item, nil
}

// FindTelegramMediaByDocumentID finds a successfully imported media with the same Telegram document id.
func (r *VideoRepository) FindTelegramMediaByDocumentID(ctx context.Context, documentID int64) (models.TelegramMedia, bool, error) {
	if documentID <= 0 {
		return models.TelegramMedia{}, false, nil
	}
	item, err := scanTelegramMedia(r.pool.QueryRow(ctx, telegramMediaSelect+`
WHERE telegram_document_id = $1
  AND video_id IS NOT NULL
  AND processing_status IN ('imported', 'duplicate')
ORDER BY created_at ASC, id ASC
LIMIT 1`, documentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.TelegramMedia{}, false, nil
	}
	if err != nil {
		return models.TelegramMedia{}, false, fmt.Errorf("find Telegram media by document id: %w", err)
	}
	return item, true, nil
}

// ClaimTelegramMedia atomically claims a message whose processing status is allowed and whose retry time has arrived.
func (r *VideoRepository) ClaimTelegramMedia(ctx context.Context, mediaID uuid.UUID, allowed []string) (models.TelegramMedia, bool, error) {
	if len(allowed) == 0 {
		return models.TelegramMedia{}, false, nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.TelegramMedia{}, false, fmt.Errorf("begin Telegram media claim: %w", err)
	}
	defer tx.Rollback(ctx)

	var id uuid.UUID
	var currentAttempts int
	err = tx.QueryRow(ctx, `
SELECT id, attempts
FROM telegram_media
WHERE id = $1
  AND processing_status = ANY($2::varchar[])
  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
FOR UPDATE SKIP LOCKED
`, mediaID, allowed).Scan(&id, &currentAttempts)
	if errors.Is(err, pgx.ErrNoRows) {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return models.TelegramMedia{}, false, fmt.Errorf("commit empty Telegram media claim: %w", commitErr)
		}
		return models.TelegramMedia{}, false, nil
	}
	if err != nil {
		return models.TelegramMedia{}, false, fmt.Errorf("lock Telegram media: %w", err)
	}

	attempts := currentAttempts + 1
	item, err := scanTelegramMedia(tx.QueryRow(ctx, `
UPDATE telegram_media
SET processing_status = 'downloading', attempts = $2, last_error = NULL, updated_at = NOW()
WHERE id = $1
RETURNING `+telegramMediaColumns, id, attempts))
	if err != nil {
		return models.TelegramMedia{}, false, fmt.Errorf("claim Telegram media: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return models.TelegramMedia{}, false, fmt.Errorf("commit Telegram media claim: %w", err)
	}
	return item, true, nil
}

// UpdateTelegramMedia applies a partial media state update.
func (r *VideoRepository) UpdateTelegramMedia(ctx context.Context, mediaID uuid.UUID, patch TelegramMediaPatch) error {
	sets := make([]string, 0, 10)
	args := []any{mediaID}
	addValue := func(expression string, value any) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", expression, len(args)))
	}

	if patch.ClearDocumentID {
		sets = append(sets, "telegram_document_id = NULL")
	} else if patch.DocumentID != nil {
		addValue("telegram_document_id", *patch.DocumentID)
	}
	if patch.DocumentDCID != nil {
		addValue("document_dc_id", *patch.DocumentDCID)
	}
	if patch.ProcessingStatus != nil {
		addValue("processing_status", strings.TrimSpace(*patch.ProcessingStatus))
	}
	if patch.TranscodeStatus != nil {
		addValue("transcode_status", strings.TrimSpace(*patch.TranscodeStatus))
	}
	if patch.ClearVideoID {
		sets = append(sets, "video_id = NULL")
	} else if patch.VideoID != nil {
		addValue("video_id", *patch.VideoID)
	}
	if patch.SHA256 != nil {
		addValue("sha256", nullableTelegramString(*patch.SHA256))
	}
	if patch.TempPath != nil {
		addValue("temp_path", nullableTelegramString(*patch.TempPath))
	}
	if patch.Attempts != nil {
		addValue("attempts", *patch.Attempts)
	}
	if patch.ClearNextRetryAt {
		sets = append(sets, "next_retry_at = NULL")
	} else if patch.NextRetryAt != nil {
		addValue("next_retry_at", patch.NextRetryAt)
	}
	if patch.LastError != nil {
		addValue("last_error", nullableTelegramString(*patch.LastError))
	}
	if len(sets) == 0 {
		return nil
	}
	sets = append(sets, "updated_at = NOW()")
	result, err := r.pool.Exec(ctx, "UPDATE telegram_media SET "+strings.Join(sets, ", ")+" WHERE id = $1", args...)
	if err != nil {
		return fmt.Errorf("update Telegram media: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", ErrTelegramMediaNotFound, mediaID)
	}
	return nil
}

// ListTelegramMediaNeedingTranscode returns imported media whose transcode task still needs dispatch.
func (r *VideoRepository) ListTelegramMediaNeedingTranscode(ctx context.Context, limit int) ([]models.TelegramMedia, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, telegramMediaSelect+`
WHERE transcode_status = 'pending' AND video_id IS NOT NULL
ORDER BY updated_at ASC, id ASC
LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list Telegram media needing transcode: %w", err)
	}
	defer rows.Close()
	items := make([]models.TelegramMedia, 0, limit)
	for rows.Next() {
		item, scanErr := scanTelegramMedia(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan Telegram media needing transcode: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate Telegram media needing transcode: %w", err)
	}
	return items, nil
}

// ListTelegramMediaNeedingDownload returns active media records that have
// remained unchanged long enough to be recovered after a worker interruption.
func (r *VideoRepository) ListTelegramMediaNeedingDownload(ctx context.Context, staleBefore time.Time, limit int) ([]models.TelegramMedia, error) {
	if staleBefore.IsZero() {
		staleBefore = time.Now().UTC().Add(-30 * time.Minute)
	}
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, telegramMediaSelect+`
WHERE processing_status IN ('queued', 'downloading', 'importing')
  AND updated_at <= $1
  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
ORDER BY updated_at ASC, id ASC
LIMIT $2`, staleBefore, limit)
	if err != nil {
		return nil, fmt.Errorf("list Telegram media needing download: %w", err)
	}
	defer rows.Close()
	items := make([]models.TelegramMedia, 0, limit)
	for rows.Next() {
		item, scanErr := scanTelegramMedia(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan Telegram media needing download: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate Telegram media needing download: %w", err)
	}
	return items, nil
}

// CountTelegramMediaByStatus counts media processing states for one source.
func (r *VideoRepository) CountTelegramMediaByStatus(ctx context.Context, sourceID uuid.UUID) (map[string]int, error) {
	rows, err := r.pool.Query(ctx, `
SELECT processing_status, COUNT(*)
FROM telegram_media
WHERE source_id = $1
GROUP BY processing_status
`, sourceID)
	if err != nil {
		return nil, fmt.Errorf("count Telegram media by status: %w", err)
	}
	defer rows.Close()
	counts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan Telegram media status count: %w", err)
		}
		counts[status] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate Telegram media status counts: %w", err)
	}
	return counts, nil
}

const telegramMediaColumns = `
id, source_id, chat_id, message_id, telegram_document_id, document_dc_id,
filename, mime_type, file_size, caption, message_url, message_created_at,
processing_status, transcode_status, video_id, sha256, temp_path, attempts,
next_retry_at, last_error, created_at, updated_at`

const telegramMediaSelect = `SELECT ` + telegramMediaColumns + ` FROM telegram_media`

func getTelegramMediaByMessage(ctx context.Context, db telegramQuerier, sourceID uuid.UUID, messageID int64) (models.TelegramMedia, error) {
	return scanTelegramMedia(db.QueryRow(ctx, telegramMediaSelect+`
WHERE source_id = $1 AND message_id = $2`, sourceID, messageID))
}

func scanTelegramSource(row telegramSourceRowScanner) (models.TelegramSource, error) {
	var item models.TelegramSource
	var nextRetryAt sql.NullTime
	var backfillCompletedAt sql.NullTime
	err := row.Scan(
		&item.ID,
		&item.ChatID,
		&item.ChatRef,
		&item.Title,
		&item.Username,
		&item.Enabled,
		&item.SyncStatus,
		&item.HistoryCursorMessageID,
		&item.LastError,
		&nextRetryAt,
		&backfillCompletedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return models.TelegramSource{}, err
	}
	item.NextRetryAt = nullTimePtr(nextRetryAt)
	item.BackfillCompletedAt = nullTimePtr(backfillCompletedAt)
	return item, nil
}

func scanTelegramMedia(row telegramMediaRowScanner) (models.TelegramMedia, error) {
	var item models.TelegramMedia
	var documentID sql.NullInt64
	var documentDCID sql.NullInt32
	var videoID sql.NullString
	var sha256 sql.NullString
	var tempPath sql.NullString
	var nextRetryAt sql.NullTime
	var lastError sql.NullString
	err := row.Scan(
		&item.ID,
		&item.SourceID,
		&item.ChatID,
		&item.MessageID,
		&documentID,
		&documentDCID,
		&item.Filename,
		&item.MIMEType,
		&item.FileSize,
		&item.Caption,
		&item.MessageURL,
		&item.MessageCreatedAt,
		&item.ProcessingStatus,
		&item.TranscodeStatus,
		&videoID,
		&sha256,
		&tempPath,
		&item.Attempts,
		&nextRetryAt,
		&lastError,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return models.TelegramMedia{}, err
	}
	if documentID.Valid {
		value := documentID.Int64
		item.DocumentID = &value
	}
	if documentDCID.Valid {
		value := int(documentDCID.Int32)
		item.DocumentDCID = &value
	}
	if videoID.Valid && strings.TrimSpace(videoID.String) != "" {
		value, err := uuid.Parse(videoID.String)
		if err != nil {
			return models.TelegramMedia{}, fmt.Errorf("parse Telegram media video id: %w", err)
		}
		item.VideoID = &value
	}
	if sha256.Valid {
		item.SHA256 = sha256.String
	}
	if tempPath.Valid {
		item.TempPath = tempPath.String
	}
	item.NextRetryAt = nullTimePtr(nextRetryAt)
	if lastError.Valid {
		item.LastError = lastError.String
	}
	return item, nil
}

func nullableTelegramChatID(chatID int64) any {
	if chatID == 0 {
		return nil
	}
	return chatID
}

func nullableTelegramString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
