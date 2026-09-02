CREATE TABLE IF NOT EXISTS telegram_sources (
    id UUID PRIMARY KEY,
    chat_id BIGINT UNIQUE,
    chat_ref TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    username TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sync_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    history_cursor_message_id BIGINT NOT NULL DEFAULT 0,
    last_error TEXT,
    next_retry_at TIMESTAMPTZ,
    backfill_completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT telegram_sources_sync_status_check CHECK (
        sync_status IN ('pending', 'backfilling', 'live', 'paused', 'error')
    ),
    CONSTRAINT telegram_sources_history_cursor_non_negative CHECK (
        history_cursor_message_id >= 0
    )
);

CREATE INDEX IF NOT EXISTS idx_telegram_sources_enabled_status
    ON telegram_sources(enabled, sync_status);

CREATE TABLE IF NOT EXISTS telegram_media (
    id UUID PRIMARY KEY,
    source_id UUID NOT NULL REFERENCES telegram_sources(id) ON DELETE CASCADE,
    chat_id BIGINT NOT NULL,
    message_id BIGINT NOT NULL,
    telegram_document_id BIGINT,
    document_dc_id INT,
    filename TEXT NOT NULL DEFAULT '',
    mime_type VARCHAR(255) NOT NULL DEFAULT '',
    file_size BIGINT NOT NULL DEFAULT 0,
    caption TEXT NOT NULL DEFAULT '',
    message_url TEXT NOT NULL DEFAULT '',
    message_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processing_status VARCHAR(20) NOT NULL DEFAULT 'discovered',
    transcode_status VARCHAR(20) NOT NULL DEFAULT 'not_required',
    video_id UUID REFERENCES videos(id) ON DELETE SET NULL,
    sha256 VARCHAR(64),
    temp_path TEXT,
    attempts INT NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT telegram_media_message_non_negative CHECK (message_id >= 0),
    CONSTRAINT telegram_media_file_size_non_negative CHECK (file_size >= 0),
    CONSTRAINT telegram_media_attempts_non_negative CHECK (attempts >= 0),
    CONSTRAINT telegram_media_sha256_length CHECK (
        sha256 IS NULL OR length(sha256) = 64
    ),
    CONSTRAINT telegram_media_processing_status_check CHECK (
        processing_status IN (
            'discovered', 'queued', 'downloading', 'importing',
            'imported', 'duplicate', 'failed', 'skipped'
        )
    ),
    CONSTRAINT telegram_media_transcode_status_check CHECK (
        transcode_status IN ('not_required', 'pending', 'enqueued', 'ready', 'failed')
    ),
    CONSTRAINT telegram_media_source_message_unique UNIQUE (source_id, message_id)
);

CREATE INDEX IF NOT EXISTS idx_telegram_media_document_id
    ON telegram_media(telegram_document_id);
CREATE INDEX IF NOT EXISTS idx_telegram_media_processing_status
    ON telegram_media(processing_status);
CREATE INDEX IF NOT EXISTS idx_telegram_media_next_retry_at
    ON telegram_media(next_retry_at);
CREATE INDEX IF NOT EXISTS idx_telegram_media_transcode_status
    ON telegram_media(transcode_status);
CREATE INDEX IF NOT EXISTS idx_telegram_media_source_status
    ON telegram_media(source_id, processing_status, updated_at DESC);
