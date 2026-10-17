ALTER TABLE cloud_download_tasks
    ADD COLUMN IF NOT EXISTS download_triggered_at TIMESTAMPTZ;
