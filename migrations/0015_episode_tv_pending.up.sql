ALTER TABLE videos
    DROP CONSTRAINT IF EXISTS videos_status_check;

ALTER TABLE videos
    ADD CONSTRAINT videos_status_check
        CHECK (status IN ('uploaded','scraping','tv_pending','processing','ready','failed'));

DROP INDEX IF EXISTS idx_videos_type_tmdb_unique;

CREATE UNIQUE INDEX IF NOT EXISTS idx_videos_movie_tmdb_unique
    ON videos(type, tmdb_id)
    WHERE tmdb_id IS NOT NULL AND type = 'movie';
