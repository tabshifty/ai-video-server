ALTER TABLE archive_import_batches
    DROP CONSTRAINT IF EXISTS archive_import_batches_status_check;

ALTER TABLE archive_import_batches
    ADD CONSTRAINT archive_import_batches_status_check
    CHECK (status IN ('uploaded','extracting','ready','processing','partial','completed','needs_password','needs_encoding','failed'));
