ALTER TABLE cloud_download_tasks
    DROP COLUMN IF EXISTS local_file_deleted_by,
    DROP COLUMN IF EXISTS local_file_deleted_at;
