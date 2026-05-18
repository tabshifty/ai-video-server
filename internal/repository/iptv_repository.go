package repository

import (
	"context"
	"database/sql"
	"fmt"

	"video-server/internal/models"
)

func (r *VideoRepository) GetIPTVPlaylist(ctx context.Context) (models.IPTVPlaylistMeta, []models.IPTVChannel, error) {
	meta, err := r.getIPTVPlaylistMeta(ctx)
	if err != nil {
		return models.IPTVPlaylistMeta{}, nil, err
	}
	channels, err := r.listIPTVChannels(ctx)
	if err != nil {
		return models.IPTVPlaylistMeta{}, nil, err
	}
	return meta, channels, nil
}

func (r *VideoRepository) SaveIPTVSourceURL(ctx context.Context, sourceURL string) (models.IPTVPlaylistMeta, []models.IPTVChannel, error) {
	_, err := r.pool.Exec(ctx, `
INSERT INTO iptv_playlists (id, source_url)
VALUES (1, $1)
ON CONFLICT (id) DO UPDATE SET source_url = EXCLUDED.source_url
`, sourceURL)
	if err != nil {
		return models.IPTVPlaylistMeta{}, nil, fmt.Errorf("save iptv source url: %w", err)
	}
	return r.GetIPTVPlaylist(ctx)
}

func (r *VideoRepository) ReplaceIPTVPlaylist(ctx context.Context, meta models.IPTVPlaylistMeta, channels []models.IPTVChannel) (models.IPTVPlaylistMeta, []models.IPTVChannel, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.IPTVPlaylistMeta{}, nil, fmt.Errorf("begin replace iptv playlist: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
INSERT INTO iptv_playlists (id, source_url, updated_at, skipped_count)
VALUES (1, $1, COALESCE($2, NOW()), $3)
ON CONFLICT (id) DO UPDATE
SET source_url = COALESCE(NULLIF(EXCLUDED.source_url, ''), iptv_playlists.source_url),
    updated_at = EXCLUDED.updated_at,
    skipped_count = EXCLUDED.skipped_count
`, meta.SourceURL, meta.UpdatedAt, meta.SkippedCount)
	if err != nil {
		return models.IPTVPlaylistMeta{}, nil, fmt.Errorf("upsert iptv playlist: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM iptv_channels WHERE playlist_id = 1`); err != nil {
		return models.IPTVPlaylistMeta{}, nil, fmt.Errorf("delete iptv channels: %w", err)
	}
	for _, channel := range channels {
		_, err := tx.Exec(ctx, `
INSERT INTO iptv_channels (id, playlist_id, name, url, group_title, logo_url, tvg_id, sort_order)
VALUES ($1, 1, $2, $3, $4, $5, $6, $7)
`, channel.ID, channel.Name, channel.URL, channel.Group, channel.LogoURL, channel.TVGID, channel.SortOrder)
		if err != nil {
			return models.IPTVPlaylistMeta{}, nil, fmt.Errorf("insert iptv channel: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return models.IPTVPlaylistMeta{}, nil, fmt.Errorf("commit replace iptv playlist: %w", err)
	}
	return r.GetIPTVPlaylist(ctx)
}

func (r *VideoRepository) getIPTVPlaylistMeta(ctx context.Context) (models.IPTVPlaylistMeta, error) {
	var meta models.IPTVPlaylistMeta
	var updatedAt sql.NullTime
	err := r.pool.QueryRow(ctx, `
SELECT COALESCE(source_url, ''), updated_at, skipped_count
FROM iptv_playlists
WHERE id = 1
`).Scan(&meta.SourceURL, &updatedAt, &meta.SkippedCount)
	if err != nil {
		if IsNotFound(err) {
			return models.IPTVPlaylistMeta{}, nil
		}
		return models.IPTVPlaylistMeta{}, fmt.Errorf("get iptv playlist meta: %w", err)
	}
	if updatedAt.Valid {
		meta.UpdatedAt = &updatedAt.Time
	}
	return meta, nil
}

func (r *VideoRepository) listIPTVChannels(ctx context.Context) ([]models.IPTVChannel, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, name, url, COALESCE(group_title, ''), COALESCE(logo_url, ''), COALESCE(tvg_id, ''), sort_order
FROM iptv_channels
WHERE playlist_id = 1
ORDER BY sort_order ASC, created_at ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list iptv channels: %w", err)
	}
	defer rows.Close()

	channels := make([]models.IPTVChannel, 0)
	for rows.Next() {
		var channel models.IPTVChannel
		if err := rows.Scan(&channel.ID, &channel.Name, &channel.URL, &channel.Group, &channel.LogoURL, &channel.TVGID, &channel.SortOrder); err != nil {
			return nil, fmt.Errorf("scan iptv channel: %w", err)
		}
		channels = append(channels, channel)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return channels, nil
}
