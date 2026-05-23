ALTER TABLE videos ADD COLUMN os_hash CHAR(16);

CREATE INDEX IF NOT EXISTS idx_videos_os_hash
    ON videos(os_hash)
    WHERE os_hash IS NOT NULL;

ALTER TABLE videos DROP CONSTRAINT IF EXISTS videos_status_check;
ALTER TABLE videos ADD CONSTRAINT videos_status_check
    CHECK (status IN ('uploaded','scraping','tv_pending','av_scrape_pending','processing','ready','failed'));
