CREATE TABLE IF NOT EXISTS ed2k_download_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_link TEXT NOT NULL,
    resource_hash VARCHAR(64) NOT NULL,
    title VARCHAR(255) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    declared_size BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(24) NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'running', 'completed', 'failed', 'deleted')),
    progress_text TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    output_dir TEXT NOT NULL DEFAULT '',
    downloaded_path TEXT NOT NULL DEFAULT '',
    retry_count INT NOT NULL DEFAULT 0,
    files JSONB NOT NULL DEFAULT '[]'::jsonb,
    history JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ed2k_download_tasks_resource_hash_unique
    ON ed2k_download_tasks(resource_hash)
    WHERE status <> 'deleted';
CREATE INDEX IF NOT EXISTS idx_ed2k_download_tasks_status_updated
    ON ed2k_download_tasks(status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_ed2k_download_tasks_created
    ON ed2k_download_tasks(created_at DESC);
