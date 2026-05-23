UPDATE videos SET status = 'uploaded' WHERE status = 'av_scrape_pending';

ALTER TABLE videos DROP CONSTRAINT IF EXISTS videos_status_check;
ALTER TABLE videos ADD CONSTRAINT videos_status_check
    CHECK (status IN ('uploaded','scraping','tv_pending','processing','ready','failed'));

DROP INDEX IF EXISTS idx_videos_os_hash;
ALTER TABLE videos DROP COLUMN IF EXISTS os_hash;
