ALTER TABLE videos
    ADD COLUMN IF NOT EXISTS image_collection_id UUID
    REFERENCES collections_images(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_videos_image_collection_id ON videos(image_collection_id);
