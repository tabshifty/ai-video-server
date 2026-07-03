DROP INDEX IF EXISTS idx_videos_short_pending_delete_at;

UPDATE videos
SET status = 'ready',
    pending_delete_at = NULL,
    updated_at = NOW()
WHERE status = 'pending_delete';

ALTER TABLE videos DROP CONSTRAINT IF EXISTS videos_pending_delete_short_check;

ALTER TABLE videos DROP CONSTRAINT IF EXISTS videos_status_check;
ALTER TABLE videos ADD CONSTRAINT videos_status_check
    CHECK (status IN ('uploaded','scraping','tv_pending','av_scrape_pending','processing','ready','failed'));

ALTER TABLE videos DROP COLUMN IF EXISTS pending_delete_at;
