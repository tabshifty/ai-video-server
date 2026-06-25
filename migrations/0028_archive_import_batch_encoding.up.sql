ALTER TABLE archive_import_batches
    ADD COLUMN IF NOT EXISTS encoding_mode VARCHAR(16) NOT NULL DEFAULT '';

ALTER TABLE archive_import_batches
    ADD COLUMN IF NOT EXISTS encoding_requested_mode VARCHAR(16) NOT NULL DEFAULT 'auto';

ALTER TABLE archive_import_batches
    DROP CONSTRAINT IF EXISTS archive_import_batches_status_check;

ALTER TABLE archive_import_batches
    ADD CONSTRAINT archive_import_batches_status_check
    CHECK (status IN ('uploaded','ready','processing','partial','completed','needs_password','needs_encoding','failed'));

ALTER TABLE archive_import_batches
    DROP CONSTRAINT IF EXISTS archive_import_batches_encoding_mode_check;

ALTER TABLE archive_import_batches
    ADD CONSTRAINT archive_import_batches_encoding_mode_check
    CHECK (encoding_mode IN ('','utf8','gbk'));

ALTER TABLE archive_import_batches
    DROP CONSTRAINT IF EXISTS archive_import_batches_encoding_requested_mode_check;

ALTER TABLE archive_import_batches
    ADD CONSTRAINT archive_import_batches_encoding_requested_mode_check
    CHECK (encoding_requested_mode IN ('auto','utf8','gbk'));
