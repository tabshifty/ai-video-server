ALTER TABLE cloud_download_tasks
    ADD COLUMN IF NOT EXISTS local_file_deleted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS local_file_deleted_by UUID;
