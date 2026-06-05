CREATE TABLE IF NOT EXISTS orphan_file_scans (
    id BIGINT PRIMARY KEY CHECK (id = 1),
    status VARCHAR(20) NOT NULL DEFAULT 'idle' CHECK (status IN ('idle', 'pending', 'running', 'completed', 'failed', 'deleted')),
    total_files BIGINT NOT NULL DEFAULT 0,
    referenced_files BIGINT NOT NULL DEFAULT 0,
    orphan_files BIGINT NOT NULL DEFAULT 0,
    deleted_files BIGINT NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO orphan_file_scans (id, status)
VALUES (1, 'idle')
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS orphan_file_scan_items (
    id BIGSERIAL PRIMARY KEY,
    scan_id BIGINT NOT NULL REFERENCES orphan_file_scans(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    relative_path TEXT NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    mod_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orphan_file_scan_items_scan_id
    ON orphan_file_scan_items(scan_id, id);

CREATE INDEX IF NOT EXISTS idx_orphan_file_scan_items_relative_path
    ON orphan_file_scan_items(relative_path);
