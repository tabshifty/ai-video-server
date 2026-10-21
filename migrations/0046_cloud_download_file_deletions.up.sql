ALTER TABLE cloud_download_tasks
    ADD COLUMN IF NOT EXISTS cloud_file_id TEXT,
    ADD COLUMN IF NOT EXISTS cloud_parent_id TEXT,
    ADD COLUMN IF NOT EXISTS cloud_user_id TEXT,
    ADD COLUMN IF NOT EXISTS cloud_file_category INTEGER;

CREATE TABLE IF NOT EXISTS cloud_download_file_deletions (
    task_id        BIGINT PRIMARY KEY REFERENCES cloud_download_tasks(id) ON DELETE CASCADE,
    file_id        TEXT,
    parent_id      TEXT,
    status         TEXT NOT NULL DEFAULT 'deleting'
                   CHECK (status IN ('deleting', 'deleted', 'failed')),
    error_message  TEXT,
    note           TEXT,
    attempt        INTEGER NOT NULL DEFAULT 1,
    requested_by   UUID,
    requested_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at    TIMESTAMPTZ,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS cloud_download_file_deletions_status_idx
    ON cloud_download_file_deletions (status);
