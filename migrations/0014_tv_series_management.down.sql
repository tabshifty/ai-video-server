ALTER TABLE episodes DROP CONSTRAINT IF EXISTS episodes_video_id_fkey;

ALTER TABLE episodes
ADD CONSTRAINT episodes_video_id_fkey
FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE;

ALTER TABLE series DROP COLUMN IF EXISTS active;
