ALTER TABLE videos
    ADD COLUMN IF NOT EXISTS transcoded_file_size BIGINT;

ALTER TABLE videos
    DROP CONSTRAINT IF EXISTS videos_transcoded_file_size_positive;

ALTER TABLE videos
    ADD CONSTRAINT videos_transcoded_file_size_positive
    CHECK (transcoded_file_size IS NULL OR transcoded_file_size > 0);
