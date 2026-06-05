package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"video-server/internal/models"
)

var ErrOrphanFileScanRunning = errors.New("orphan file scan running")

type orphanFileScanDB interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func (r *VideoRepository) GetOrphanFileScan(ctx context.Context) (models.AdminOrphanFileScan, error) {
	scan, err := scanOrphanFileScanState(ctx, r.pool)
	if err != nil {
		return models.AdminOrphanFileScan{}, err
	}
	items, err := listOrphanFileScanItems(ctx, r.pool)
	if err != nil {
		return models.AdminOrphanFileScan{}, err
	}
	scan.Items = items
	return scan, nil
}

func (r *VideoRepository) BeginOrphanFileScan(ctx context.Context) (models.AdminOrphanFileScan, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.AdminOrphanFileScan{}, fmt.Errorf("begin orphan file scan tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `INSERT INTO orphan_file_scans (id, status) VALUES (1, 'idle') ON CONFLICT (id) DO NOTHING`); err != nil {
		return models.AdminOrphanFileScan{}, fmt.Errorf("seed orphan file scan row: %w", err)
	}

	prev, err := scanOrphanFileScanState(ctx, tx)
	if err != nil {
		return models.AdminOrphanFileScan{}, err
	}
	if prev.Status == "pending" || prev.Status == "running" {
		return prev, ErrOrphanFileScanRunning
	}

	if _, err := tx.Exec(ctx, `
UPDATE orphan_file_scans
SET
  status='pending',
  total_files=0,
  referenced_files=0,
  orphan_files=0,
  deleted_files=0,
  error_message='',
  started_at=NOW(),
  finished_at=NULL,
  updated_at=NOW()
WHERE id=1
`); err != nil {
		return models.AdminOrphanFileScan{}, fmt.Errorf("mark orphan file scan pending: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return models.AdminOrphanFileScan{}, fmt.Errorf("commit orphan file scan prepare: %w", err)
	}
	return prev, nil
}

func (r *VideoRepository) RestoreOrphanFileScan(ctx context.Context, prev models.AdminOrphanFileScan) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin orphan file scan restore tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
UPDATE orphan_file_scans
SET
  status=$2,
  total_files=$3,
  referenced_files=$4,
  orphan_files=$5,
  deleted_files=$6,
  error_message=$7,
  started_at=$8,
  finished_at=$9,
  updated_at=NOW()
WHERE id=$1
`, prev.ID, prev.Status, prev.TotalFiles, prev.ReferencedFiles, prev.OrphanFiles, prev.DeletedFiles, prev.Error, timePtrOrNil(prev.StartedAt), timePtrOrNil(prev.FinishedAt)); err != nil {
		return fmt.Errorf("restore orphan file scan: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit orphan file scan restore: %w", err)
	}
	return nil
}

func (r *VideoRepository) MarkOrphanFileScanRunning(ctx context.Context) error {
	return setOrphanFileScanRunning(ctx, r.pool)
}

func (r *VideoRepository) CompleteOrphanFileScan(ctx context.Context, totalFiles, referencedFiles, orphanFiles int64, items []models.AdminOrphanFileScanItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin orphan file scan complete tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM orphan_file_scan_items WHERE scan_id = 1`); err != nil {
		return fmt.Errorf("clear orphan file scan items: %w", err)
	}

	if len(items) > 0 {
		batch := &pgx.Batch{}
		for _, item := range items {
			batch.Queue(`
INSERT INTO orphan_file_scan_items (scan_id, file_path, relative_path, size_bytes, mod_time)
VALUES (1, $1, $2, $3, $4)
`, item.FilePath, item.RelativePath, item.SizeBytes, item.ModTime)
		}
		br := tx.SendBatch(ctx, batch)
		for i := 0; i < batch.Len(); i++ {
			if _, err := br.Exec(); err != nil {
				_ = br.Close()
				return fmt.Errorf("insert orphan file scan item: %w", err)
			}
		}
		if err := br.Close(); err != nil {
			return fmt.Errorf("close orphan file scan item batch: %w", err)
		}
	}

	if _, err := tx.Exec(ctx, `
UPDATE orphan_file_scans
SET
  status='completed',
  total_files=$2,
  referenced_files=$3,
  orphan_files=$4,
  deleted_files=0,
  error_message='',
  finished_at=NOW(),
  updated_at=NOW()
WHERE id=$1
`, 1, totalFiles, referencedFiles, orphanFiles); err != nil {
		return fmt.Errorf("update orphan file scan complete state: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit orphan file scan complete: %w", err)
	}
	return nil
}

func (r *VideoRepository) FailOrphanFileScan(ctx context.Context, errMsg string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin orphan file scan fail tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM orphan_file_scan_items WHERE scan_id = 1`); err != nil {
		return fmt.Errorf("clear orphan file scan items: %w", err)
	}
	if _, err := tx.Exec(ctx, `
UPDATE orphan_file_scans
SET
  status='failed',
  total_files=0,
  referenced_files=0,
  orphan_files=0,
  deleted_files=0,
  error_message=$2,
  finished_at=NOW(),
  updated_at=NOW()
WHERE id=$1
`, 1, strings.TrimSpace(errMsg)); err != nil {
		return fmt.Errorf("update orphan file scan failed state: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit orphan file scan fail: %w", err)
	}
	return nil
}

func (r *VideoRepository) MarkOrphanFileScanDeleted(ctx context.Context, deletedCount int64) error {
	_, err := r.pool.Exec(ctx, `
UPDATE orphan_file_scans
SET
  status='deleted',
  deleted_files=$2,
  updated_at=NOW()
WHERE id=$1
`, 1, deletedCount)
	if err != nil {
		return fmt.Errorf("mark orphan file scan deleted: %w", err)
	}
	return nil
}

func (r *VideoRepository) CollectReferencedFilePaths(ctx context.Context, storageRoot string) (map[string]struct{}, error) {
	refs := map[string]struct{}{}
	add := func(value string) {
		if path := normalizePotentialStoragePath(value); path != "" {
			refs[path] = struct{}{}
		}
	}
	addJSON := func(raw []byte) error {
		if len(raw) == 0 {
			return nil
		}
		var parsed any
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return err
		}
		collectLocalStoragePaths(parsed, add)
		return nil
	}

	rows, err := r.pool.Query(ctx, `
SELECT COALESCE(original_path, ''), COALESCE(transcoded_path, ''), COALESCE(thumbnail_path, ''), COALESCE(metadata, '{}'::jsonb)
FROM videos
`)
	if err != nil {
		return nil, fmt.Errorf("query video file references: %w", err)
	}
	for rows.Next() {
		var originalPath, transcodedPath, thumbnailPath string
		var metadata []byte
		if err := rows.Scan(&originalPath, &transcodedPath, &thumbnailPath, &metadata); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan video file references: %w", err)
		}
		add(originalPath)
		add(transcodedPath)
		add(thumbnailPath)
		if err := addJSON(metadata); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan video metadata references: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate video file references: %w", err)
	}
	rows.Close()

	rows, err = r.pool.Query(ctx, `
SELECT COALESCE(original_path, ''), COALESCE(stored_path, ''), COALESCE(metadata, '{}'::jsonb)
FROM images
`)
	if err != nil {
		return nil, fmt.Errorf("query image file references: %w", err)
	}
	for rows.Next() {
		var originalPath, storedPath string
		var metadata []byte
		if err := rows.Scan(&originalPath, &storedPath, &metadata); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan image file references: %w", err)
		}
		add(originalPath)
		add(storedPath)
		if err := addJSON(metadata); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan image metadata references: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate image file references: %w", err)
	}
	rows.Close()

	rows, err = r.pool.Query(ctx, `SELECT COALESCE(variant_path, '') FROM image_variants`)
	if err != nil {
		return nil, fmt.Errorf("query image variant references: %w", err)
	}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan image variant reference: %w", err)
		}
		add(path)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate image variant references: %w", err)
	}
	rows.Close()

	rows, err = r.pool.Query(ctx, `SELECT COALESCE(stored_path, '') FROM video_subtitles`)
	if err != nil {
		return nil, fmt.Errorf("query subtitle file references: %w", err)
	}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan subtitle file reference: %w", err)
		}
		add(path)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate subtitle file references: %w", err)
	}
	rows.Close()

	rows, err = r.pool.Query(ctx, `
SELECT COALESCE(poster_path, ''), COALESCE(backdrop_path, '')
FROM series
`)
	if err != nil {
		return nil, fmt.Errorf("query tv series references: %w", err)
	}
	for rows.Next() {
		var posterPath, backdropPath string
		if err := rows.Scan(&posterPath, &backdropPath); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan tv series reference: %w", err)
		}
		add(posterPath)
		add(backdropPath)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate tv series references: %w", err)
	}
	rows.Close()

	rows, err = r.pool.Query(ctx, `SELECT COALESCE(poster_path, '') FROM seasons`)
	if err != nil {
		return nil, fmt.Errorf("query tv season references: %w", err)
	}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan tv season reference: %w", err)
		}
		add(path)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate tv season references: %w", err)
	}
	rows.Close()

	rows, err = r.pool.Query(ctx, `SELECT COALESCE(still_path, '') FROM episodes`)
	if err != nil {
		return nil, fmt.Errorf("query tv episode references: %w", err)
	}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan tv episode reference: %w", err)
		}
		add(path)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate tv episode references: %w", err)
	}
	rows.Close()

	rows, err = r.pool.Query(ctx, `
SELECT COALESCE(cover_url, '')
FROM collections
`)
	if err != nil {
		return nil, fmt.Errorf("query collection references: %w", err)
	}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan collection reference: %w", err)
		}
		add(path)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate collection references: %w", err)
	}
	rows.Close()

	rows, err = r.pool.Query(ctx, `
SELECT COALESCE(cover_url, '')
FROM collections_images
`)
	if err != nil {
		return nil, fmt.Errorf("query image collection references: %w", err)
	}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan image collection reference: %w", err)
		}
		add(path)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate image collection references: %w", err)
	}
	rows.Close()

	rows, err = r.pool.Query(ctx, `
SELECT COALESCE(avatar_url, '')
FROM actors
`)
	if err != nil {
		return nil, fmt.Errorf("query actor references: %w", err)
	}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan actor reference: %w", err)
		}
		add(path)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate actor references: %w", err)
	}
	rows.Close()

	rows, err = r.pool.Query(ctx, `SELECT id FROM actors`)
	if err != nil {
		return nil, fmt.Errorf("query actor avatar references: %w", err)
	}
	for rows.Next() {
		var actorID uuid.UUID
		if err := rows.Scan(&actorID); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan actor avatar id: %w", err)
		}
		if storageRoot == "" || actorID == uuid.Nil {
			continue
		}
		dir := filepath.Join(storageRoot, "actors", actorID.String())
		for _, candidate := range []string{
			filepath.Join(dir, "avatar.jpg"),
			filepath.Join(dir, "avatar.png"),
			filepath.Join(dir, "avatar.webp"),
			filepath.Join(dir, "avatar.gif"),
		} {
			add(candidate)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate actor avatar references: %w", err)
	}
	rows.Close()

	return refs, nil
}

func scanOrphanFileScanState(ctx context.Context, db orphanFileScanDB) (models.AdminOrphanFileScan, error) {
	var scan models.AdminOrphanFileScan
	var startedAt sql.NullTime
	var finishedAt sql.NullTime
	err := db.QueryRow(ctx, `
SELECT id, status, total_files, referenced_files, orphan_files, deleted_files, COALESCE(error_message, ''), started_at, finished_at, created_at, updated_at
FROM orphan_file_scans
WHERE id = 1
`).Scan(
		&scan.ID,
		&scan.Status,
		&scan.TotalFiles,
		&scan.ReferencedFiles,
		&scan.OrphanFiles,
		&scan.DeletedFiles,
		&scan.Error,
		&startedAt,
		&finishedAt,
		&scan.CreatedAt,
		&scan.UpdatedAt,
	)
	if err != nil {
		return models.AdminOrphanFileScan{}, fmt.Errorf("load orphan file scan state: %w", err)
	}
	scan.Error = strings.TrimSpace(scan.Error)
	scan.StartedAt = nullTimePtr(startedAt)
	scan.FinishedAt = nullTimePtr(finishedAt)
	return scan, nil
}

func listOrphanFileScanItems(ctx context.Context, db orphanFileScanDB) ([]models.AdminOrphanFileScanItem, error) {
	rows, err := db.Query(ctx, `
SELECT id, file_path, relative_path, size_bytes, mod_time, created_at
FROM orphan_file_scan_items
WHERE scan_id = 1
ORDER BY id ASC
`)
	if err != nil {
		return nil, fmt.Errorf("load orphan file scan items: %w", err)
	}
	defer rows.Close()

	items := make([]models.AdminOrphanFileScanItem, 0, 128)
	for rows.Next() {
		var item models.AdminOrphanFileScanItem
		if err := rows.Scan(&item.ID, &item.FilePath, &item.RelativePath, &item.SizeBytes, &item.ModTime, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan orphan file scan item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orphan file scan items: %w", err)
	}
	return items, nil
}

func setOrphanFileScanRunning(ctx context.Context, db orphanFileScanDB) error {
	_, err := db.Exec(ctx, `
UPDATE orphan_file_scans
SET
  status='running',
  total_files=0,
  referenced_files=0,
  orphan_files=0,
  deleted_files=0,
  error_message='',
  finished_at=NULL,
  updated_at=NOW()
WHERE id=1
`)
	if err != nil {
		return fmt.Errorf("mark orphan file scan running: %w", err)
	}
	return nil
}

func normalizePotentialStoragePath(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.Contains(lower, "://") {
		return ""
	}
	if strings.HasPrefix(value, "/api/") {
		return ""
	}
	if filepath.Ext(value) == "" {
		return ""
	}
	return filepath.Clean(value)
}

func collectLocalStoragePaths(value any, add func(string)) {
	switch v := value.(type) {
	case map[string]any:
		for _, child := range v {
			collectLocalStoragePaths(child, add)
		}
	case []any:
		for _, child := range v {
			collectLocalStoragePaths(child, add)
		}
	case string:
		add(v)
	}
}

func timePtrOrNil(v *time.Time) any {
	if v == nil {
		return nil
	}
	return *v
}
