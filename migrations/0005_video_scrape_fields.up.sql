ALTER TABLE videos
    ADD COLUMN IF NOT EXISTS tmdb_id INT;

ALTER TABLE videos
    DROP CONSTRAINT IF EXISTS videos_status_check;

ALTER TABLE videos
    ADD CONSTRAINT videos_status_check
        CHECK (status IN ('uploaded','scraping','processing','ready','failed'));

CREATE UNIQUE INDEX IF NOT EXISTS idx_videos_type_tmdb_unique
    ON videos(type, tmdb_id)
    WHERE tmdb_id IS NOT NULL AND type IN ('movie','episode');

