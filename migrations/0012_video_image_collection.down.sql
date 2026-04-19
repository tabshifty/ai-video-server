DROP INDEX IF EXISTS idx_videos_image_collection_id;

ALTER TABLE videos
    DROP COLUMN IF EXISTS image_collection_id;
