CREATE TABLE IF NOT EXISTS video_subtitles (
    id UUID PRIMARY KEY,
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    source_type VARCHAR(20) NOT NULL CHECK (source_type IN ('uploaded', 'embedded')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('ready', 'failed')),
    language_code VARCHAR(32) NOT NULL DEFAULT '',
    label VARCHAR(255) NOT NULL DEFAULT '',
    format VARCHAR(32) NOT NULL DEFAULT '',
    mime_type VARCHAR(128) NOT NULL DEFAULT '',
    stored_path TEXT NOT NULL DEFAULT '',
    file_size BIGINT NOT NULL DEFAULT 0,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INT NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_video_subtitles_video_id
    ON video_subtitles(video_id, source_type, sort_order, created_at);

CREATE UNIQUE INDEX IF NOT EXISTS uq_video_subtitles_uploaded_default
    ON video_subtitles(video_id)
    WHERE is_default = TRUE AND source_type = 'uploaded';
