ALTER TABLE archive_import_batches
    DROP CONSTRAINT IF EXISTS archive_import_batches_encoding_requested_mode_check;

ALTER TABLE archive_import_batches
    DROP CONSTRAINT IF EXISTS archive_import_batches_encoding_mode_check;

ALTER TABLE archive_import_batches
    DROP CONSTRAINT IF EXISTS archive_import_batches_status_check;

UPDATE archive_import_batches
SET status = 'failed'
WHERE status = 'needs_encoding';

ALTER TABLE archive_import_batches
    ADD CONSTRAINT archive_import_batches_status_check
    CHECK (status IN ('uploaded','ready','processing','partial','completed','needs_password','failed'));

ALTER TABLE archive_import_batches
    DROP COLUMN IF EXISTS encoding_requested_mode;

ALTER TABLE archive_import_batches
    DROP COLUMN IF EXISTS encoding_mode;
