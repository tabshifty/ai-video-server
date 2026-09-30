DROP INDEX IF EXISTS idx_archive_import_file_images_image;
DROP TABLE IF EXISTS archive_import_file_images;

DROP INDEX IF EXISTS idx_video_images_image_id;
DROP TABLE IF EXISTS video_images;

DROP INDEX IF EXISTS idx_video_collection_image_collections_image;
DROP TABLE IF EXISTS video_collection_image_collections;

DROP INDEX IF EXISTS idx_video_image_collections_collection;
DROP TABLE IF EXISTS video_image_collections;

ALTER TABLE archive_import_groups
    DROP COLUMN IF EXISTS directory_path;

ALTER TABLE archive_import_groups
    DROP CONSTRAINT IF EXISTS archive_import_groups_media_kind_check;

UPDATE archive_import_groups
SET media_kind = CASE WHEN media_kind = 'mixed' THEN 'video' ELSE media_kind END
WHERE media_kind = 'mixed';

ALTER TABLE archive_import_groups
    ADD CONSTRAINT archive_import_groups_media_kind_check
    CHECK (media_kind IN ('video', 'image'));
