DROP INDEX IF EXISTS idx_videos_type_tmdb_unique;

ALTER TABLE videos
    DROP CONSTRAINT IF EXISTS videos_status_check;

ALTER TABLE videos
    ADD CONSTRAINT videos_status_check
        CHECK (status IN ('uploaded','processing','ready','failed'));

ALTER TABLE videos
    DROP COLUMN IF EXISTS tmdb_id;

