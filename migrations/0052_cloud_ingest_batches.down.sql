DROP INDEX IF EXISTS cloud_download_tasks_awaiting_ingest_idx;
DROP INDEX IF EXISTS cloud_download_tasks_ingest_batch_idx;

ALTER TABLE cloud_download_tasks
    DROP COLUMN IF EXISTS ingest_title,
    DROP COLUMN IF EXISTS ingest_batch_id;

DROP INDEX IF EXISTS cloud_ingest_batches_created_at_idx;
DROP TABLE IF EXISTS cloud_ingest_batches;
