CREATE TABLE IF NOT EXISTS iptv_playlists (
    id SMALLINT PRIMARY KEY DEFAULT 1,
    source_url TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ,
    skipped_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT iptv_playlists_singleton CHECK (id = 1),
    CONSTRAINT iptv_playlists_skipped_count_non_negative CHECK (skipped_count >= 0)
);

CREATE TABLE IF NOT EXISTS iptv_channels (
    id TEXT PRIMARY KEY,
    playlist_id SMALLINT NOT NULL REFERENCES iptv_playlists(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    group_title TEXT NOT NULL DEFAULT '',
    logo_url TEXT NOT NULL DEFAULT '',
    tvg_id TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT iptv_channels_sort_order_non_negative CHECK (sort_order >= 0),
    CONSTRAINT iptv_channels_http_url CHECK (url ~* '^https?://')
);

CREATE INDEX IF NOT EXISTS idx_iptv_channels_playlist_order
    ON iptv_channels(playlist_id, sort_order);
