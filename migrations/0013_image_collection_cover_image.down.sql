DROP INDEX IF EXISTS idx_collections_images_cover_image_id;

ALTER TABLE collections_images
    DROP COLUMN IF EXISTS cover_image_id;
