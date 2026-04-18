DROP INDEX IF EXISTS idx_transcoding_jobs_status_started_at;

ALTER TABLE transcoding_jobs
    DROP COLUMN IF EXISTS progress_updated_at;

ALTER TABLE transcoding_jobs
    DROP COLUMN IF EXISTS progress_percent;

ALTER TABLE transcoding_jobs
    DROP COLUMN IF EXISTS remaining_seconds;

ALTER TABLE transcoding_jobs
    DROP COLUMN IF EXISTS processed_seconds;

ALTER TABLE transcoding_jobs
    DROP COLUMN IF EXISTS source_duration_seconds;
