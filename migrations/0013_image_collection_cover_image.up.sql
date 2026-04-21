ALTER TABLE collections_images
    ADD COLUMN IF NOT EXISTS cover_image_id UUID
    REFERENCES images(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_collections_images_cover_image_id
    ON collections_images(cover_image_id);
