ALTER TABLE cloud_download_tasks
    ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'forum_post';

ALTER TABLE cloud_download_tasks
    DROP CONSTRAINT IF EXISTS cloud_download_tasks_source_check;

ALTER TABLE cloud_download_tasks
    ADD CONSTRAINT cloud_download_tasks_source_check
    CHECK (source IN ('forum_post', 'toolbox', 'manual'));
