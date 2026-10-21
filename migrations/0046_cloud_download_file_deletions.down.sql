DROP TABLE IF EXISTS cloud_download_file_deletions;

ALTER TABLE cloud_download_tasks
    DROP COLUMN IF EXISTS cloud_file_category,
    DROP COLUMN IF EXISTS cloud_user_id,
    DROP COLUMN IF EXISTS cloud_parent_id,
    DROP COLUMN IF EXISTS cloud_file_id;
