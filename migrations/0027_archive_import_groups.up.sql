CREATE TABLE IF NOT EXISTS archive_import_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id UUID NOT NULL REFERENCES archive_import_batches(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    media_kind VARCHAR(20) NOT NULL CHECK (media_kind IN ('video','image')),
    title VARCHAR(200),
    description TEXT,
    tags JSONB,
    video_type VARCHAR(10) CHECK (video_type IN ('short','movie','episode','av')),
    video_collection_ids JSONB,
    image_collection_ids JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_archive_import_groups_batch_name_unique
    ON archive_import_groups(batch_id, lower(name));
CREATE INDEX IF NOT EXISTS idx_archive_import_groups_batch_created
    ON archive_import_groups(batch_id, created_at DESC);

ALTER TABLE archive_import_files
    ADD COLUMN IF NOT EXISTS group_id UUID,
    ADD COLUMN IF NOT EXISTS field_overrides JSONB NOT NULL DEFAULT '{}'::jsonb;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'archive_import_files_group_id_fkey'
    ) THEN
        ALTER TABLE archive_import_files
            ADD CONSTRAINT archive_import_files_group_id_fkey
            FOREIGN KEY (group_id) REFERENCES archive_import_groups(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_archive_import_files_group_id
    ON archive_import_files(group_id);
CREATE INDEX IF NOT EXISTS idx_archive_import_files_batch_group_status
    ON archive_import_files(batch_id, group_id, status, media_kind);
