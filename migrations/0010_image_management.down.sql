DROP INDEX IF EXISTS idx_image_variants_image_id;
DROP TABLE IF EXISTS image_variants;

DROP INDEX IF EXISTS idx_image_collections_collection_id;
DROP TABLE IF EXISTS image_collections;

DROP INDEX IF EXISTS idx_collections_images_active_sort;
DROP INDEX IF EXISTS idx_collections_images_normalized_name;
DROP TABLE IF EXISTS collections_images;

DROP INDEX IF EXISTS idx_image_actors_actor_id;
DROP TABLE IF EXISTS image_actors;

DROP INDEX IF EXISTS idx_image_hashes_image_id;
DROP TABLE IF EXISTS image_hashes;

DROP INDEX IF EXISTS idx_images_user_created;
DROP INDEX IF EXISTS idx_images_active_created;
DROP INDEX IF EXISTS idx_images_status_created;
DROP TABLE IF EXISTS images;
