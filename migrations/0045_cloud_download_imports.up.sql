CREATE TABLE IF NOT EXISTS cloud_download_imports (
    id               BIGSERIAL PRIMARY KEY,
    task_id          BIGINT NOT NULL UNIQUE REFERENCES cloud_download_tasks(id) ON DELETE CASCADE,
    kind             TEXT NOT NULL CHECK (kind IN ('video', 'archive')),
    status           TEXT NOT NULL DEFAULT 'queued'
                     CHECK (status IN ('queued', 'hashing', 'copying', 'creating', 'imported', 'skipped_existing', 'failed')),
    source_path      TEXT,
    source_size      BIGINT,
    sha256           CHAR(64),
    video_id         UUID REFERENCES videos(id) ON DELETE SET NULL,
    archive_batch_id UUID REFERENCES archive_import_batches(id) ON DELETE SET NULL,
    attempt          INTEGER NOT NULL DEFAULT 0,
    error_message    TEXT,
    created_by       UUID,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at       TIMESTAMPTZ,
    imported_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS cloud_download_imports_status_idx
    ON cloud_download_imports (status);
