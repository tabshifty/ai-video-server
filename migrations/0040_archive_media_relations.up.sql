CREATE TABLE IF NOT EXISTS video_image_collections (
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    image_collection_id UUID NOT NULL REFERENCES collections_images(id) ON DELETE CASCADE,
    source VARCHAR(24) NOT NULL DEFAULT 'direct' CHECK (source IN ('direct', 'video_collection', 'archive_import')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (video_id, image_collection_id)
);

CREATE INDEX IF NOT EXISTS idx_video_image_collections_collection
    ON video_image_collections(image_collection_id);

CREATE TABLE IF NOT EXISTS video_collection_image_collections (
    video_collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    image_collection_id UUID NOT NULL REFERENCES collections_images(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (video_collection_id, image_collection_id)
);

CREATE INDEX IF NOT EXISTS idx_video_collection_image_collections_image
    ON video_collection_image_collections(image_collection_id);

CREATE TABLE IF NOT EXISTS video_images (
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    source VARCHAR(24) NOT NULL DEFAULT 'archive_import' CHECK (source IN ('archive_import', 'manual', 'video_collection')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (video_id, image_id)
);

CREATE INDEX IF NOT EXISTS idx_video_images_image_id ON video_images(image_id);

CREATE TABLE IF NOT EXISTS archive_import_file_images (
    video_file_id UUID NOT NULL REFERENCES archive_import_files(id) ON DELETE CASCADE,
    image_file_id UUID NOT NULL REFERENCES archive_import_files(id) ON DELETE CASCADE,
    source VARCHAR(24) NOT NULL DEFAULT 'directory' CHECK (source IN ('directory', 'manual')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (video_file_id, image_file_id),
    CHECK (video_file_id <> image_file_id)
);

CREATE INDEX IF NOT EXISTS idx_archive_import_file_images_image
    ON archive_import_file_images(image_file_id);

ALTER TABLE archive_import_groups
    ADD COLUMN IF NOT EXISTS directory_path TEXT NOT NULL DEFAULT '';

ALTER TABLE archive_import_groups
    DROP CONSTRAINT IF EXISTS archive_import_groups_media_kind_check;

ALTER TABLE archive_import_groups
    ADD CONSTRAINT archive_import_groups_media_kind_check
    CHECK (media_kind IN ('video', 'image', 'mixed'));

INSERT INTO video_image_collections (video_id, image_collection_id, source)
SELECT id, image_collection_id, 'direct'
FROM videos
WHERE image_collection_id IS NOT NULL
ON CONFLICT (video_id, image_collection_id) DO NOTHING;

