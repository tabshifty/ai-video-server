ALTER TABLE transcoding_jobs
    ADD COLUMN IF NOT EXISTS source_duration_seconds INT;

ALTER TABLE transcoding_jobs
    ADD COLUMN IF NOT EXISTS processed_seconds INT;

ALTER TABLE transcoding_jobs
    ADD COLUMN IF NOT EXISTS remaining_seconds INT;

ALTER TABLE transcoding_jobs
    ADD COLUMN IF NOT EXISTS progress_percent DOUBLE PRECISION;

ALTER TABLE transcoding_jobs
    ADD COLUMN IF NOT EXISTS progress_updated_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_transcoding_jobs_status_started_at ON transcoding_jobs(status, started_at DESC);
