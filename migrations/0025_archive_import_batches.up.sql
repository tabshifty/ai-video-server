CREATE TABLE IF NOT EXISTS archive_import_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    title VARCHAR(200) NOT NULL,
    original_filename TEXT NOT NULL,
    archive_format VARCHAR(10) NOT NULL CHECK (archive_format IN ('zip','rar','7z')),
    original_path TEXT NOT NULL,
    extracted_dir TEXT NOT NULL DEFAULT '',
    status VARCHAR(24) NOT NULL DEFAULT 'uploaded' CHECK (status IN ('uploaded','ready','processing','partial','completed','needs_password','failed')),
    last_error TEXT NOT NULL DEFAULT '',
    total_entries INT NOT NULL DEFAULT 0,
    processable_entries INT NOT NULL DEFAULT 0,
    processed_entries INT NOT NULL DEFAULT 0,
    skipped_entries INT NOT NULL DEFAULT 0,
    failed_entries INT NOT NULL DEFAULT 0,
    default_title_prefix VARCHAR(200) NOT NULL DEFAULT '',
    default_description TEXT NOT NULL DEFAULT '',
    default_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_video_collection_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_image_collection_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_archive_import_batches_status_created ON archive_import_batches(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_archive_import_batches_created ON archive_import_batches(created_at DESC);

CREATE TABLE IF NOT EXISTS archive_import_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id UUID NOT NULL REFERENCES archive_import_batches(id) ON DELETE CASCADE,
    relative_path TEXT NOT NULL,
    file_path TEXT NOT NULL,
    entry_type VARCHAR(20) NOT NULL CHECK (entry_type IN ('file','directory')),
    media_kind VARCHAR(20) NOT NULL CHECK (media_kind IN ('video','image','archive','other','directory')),
    video_type VARCHAR(10) NOT NULL DEFAULT 'short' CHECK (video_type IN ('short','movie','episode','av')),
    file_size BIGINT NOT NULL DEFAULT 0,
    mime_type VARCHAR(120) NOT NULL DEFAULT '',
    status VARCHAR(24) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','processing','ready','existing','skipped','failed')),
    reason TEXT NOT NULL DEFAULT '',
    title VARCHAR(200) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    video_collection_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    image_collection_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    linked_video_id UUID,
    linked_image_id UUID,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    UNIQUE (batch_id, relative_path)
);

CREATE INDEX IF NOT EXISTS idx_archive_import_files_batch_id ON archive_import_files(batch_id);
CREATE INDEX IF NOT EXISTS idx_archive_import_files_batch_status ON archive_import_files(batch_id, status, media_kind);
CREATE INDEX IF NOT EXISTS idx_archive_import_files_linked_video_id ON archive_import_files(linked_video_id);
CREATE INDEX IF NOT EXISTS idx_archive_import_files_linked_image_id ON archive_import_files(linked_image_id);
