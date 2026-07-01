UPDATE ed2k_download_tasks
SET status = 'running'
WHERE status = 'canceling';

UPDATE ed2k_download_tasks
SET status = 'deleted'
WHERE status = 'cancelled';

UPDATE ed2k_download_tasks
SET status = 'completed'
WHERE status = 'files_cleaned';

DROP INDEX IF EXISTS idx_ed2k_download_tasks_resource_hash_unique;

CREATE UNIQUE INDEX IF NOT EXISTS idx_ed2k_download_tasks_resource_hash_unique
    ON ed2k_download_tasks(resource_hash)
    WHERE status <> 'deleted';

ALTER TABLE ed2k_download_tasks
    DROP CONSTRAINT IF EXISTS ed2k_download_tasks_status_check;

ALTER TABLE ed2k_download_tasks
    ADD CONSTRAINT ed2k_download_tasks_status_check
    CHECK (status IN ('queued', 'running', 'completed', 'failed', 'deleted'));

ALTER TABLE ed2k_download_tasks
    DROP COLUMN IF EXISTS cleaned_at;
