ALTER TABLE videos ADD COLUMN IF NOT EXISTS pending_delete_at TIMESTAMPTZ;

ALTER TABLE videos DROP CONSTRAINT IF EXISTS videos_status_check;
ALTER TABLE videos ADD CONSTRAINT videos_status_check
    CHECK (status IN ('uploaded','scraping','tv_pending','av_scrape_pending','processing','ready','pending_delete','failed'));

ALTER TABLE videos DROP CONSTRAINT IF EXISTS videos_pending_delete_short_check;
ALTER TABLE videos ADD CONSTRAINT videos_pending_delete_short_check
    CHECK (status <> 'pending_delete' OR type = 'short');

CREATE INDEX IF NOT EXISTS idx_videos_short_pending_delete_at
    ON videos(pending_delete_at DESC)
    WHERE type = 'short' AND status = 'pending_delete';
