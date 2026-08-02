ALTER TABLE videos
    DROP CONSTRAINT IF EXISTS videos_transcoded_file_size_positive;

ALTER TABLE videos
    DROP COLUMN IF EXISTS transcoded_file_size;
