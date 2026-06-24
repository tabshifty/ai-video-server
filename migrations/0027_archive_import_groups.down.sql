DROP INDEX IF EXISTS idx_archive_import_files_batch_group_status;
DROP INDEX IF EXISTS idx_archive_import_files_group_id;

ALTER TABLE archive_import_files
    DROP CONSTRAINT IF EXISTS archive_import_files_group_id_fkey;

ALTER TABLE archive_import_files
    DROP COLUMN IF EXISTS field_overrides,
    DROP COLUMN IF EXISTS group_id;

DROP INDEX IF EXISTS idx_archive_import_groups_batch_created;
DROP INDEX IF EXISTS idx_archive_import_groups_batch_name_unique;
DROP TABLE IF EXISTS archive_import_groups;
