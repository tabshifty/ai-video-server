CREATE TABLE IF NOT EXISTS cloud_download_tasks (
    id                  BIGSERIAL PRIMARY KEY,
    post_id             UUID REFERENCES collected_forum_posts(id) ON DELETE SET NULL,
    protocol            TEXT NOT NULL CHECK (protocol IN ('ed2k', 'magnet')),
    info_hash           TEXT NOT NULL,
    source_url          TEXT NOT NULL,
    declared_name       TEXT,
    declared_size       BIGINT,
    status              TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'submitting', 'queued', 'ready', 'downloading', 'completed', 'failed', 'blocked')),
    blocked_reason      TEXT,
    error_message       TEXT,
    retry_count         INTEGER NOT NULL DEFAULT 0,
    pick_code           TEXT,
    cloud_status        TEXT,
    progress_percent    DOUBLE PRECISION,
    progress_updated_at TIMESTAMPTZ,
    local_path          TEXT,
    created_by          UUID,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at          TIMESTAMPTZ,
    finished_at         TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS cloud_download_tasks_info_hash_key
    ON cloud_download_tasks (info_hash);

CREATE INDEX IF NOT EXISTS cloud_download_tasks_status_created_idx
    ON cloud_download_tasks (status, created_at DESC);

CREATE INDEX IF NOT EXISTS cloud_download_tasks_post_idx
    ON cloud_download_tasks (post_id);
