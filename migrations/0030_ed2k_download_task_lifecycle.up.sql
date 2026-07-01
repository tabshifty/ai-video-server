ALTER TABLE ed2k_download_tasks
    DROP CONSTRAINT IF EXISTS ed2k_download_tasks_status_check;

ALTER TABLE ed2k_download_tasks
    ADD CONSTRAINT ed2k_download_tasks_status_check
    CHECK (status IN ('queued', 'running', 'canceling', 'cancelled', 'completed', 'failed', 'files_cleaned', 'deleted'));

ALTER TABLE ed2k_download_tasks
    ADD COLUMN IF NOT EXISTS cleaned_at TIMESTAMPTZ;

DROP INDEX IF EXISTS idx_ed2k_download_tasks_resource_hash_unique;

CREATE UNIQUE INDEX IF NOT EXISTS idx_ed2k_download_tasks_resource_hash_unique
    ON ed2k_download_tasks(resource_hash)
    WHERE status <> 'deleted' AND status <> 'cancelled';
