ALTER TABLE cloud_download_tasks
    DROP CONSTRAINT IF EXISTS cloud_download_tasks_source_check;

ALTER TABLE cloud_download_tasks
    DROP COLUMN IF EXISTS source;
