CREATE TABLE IF NOT EXISTS cloud_ingest_batches (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title_mode          TEXT NOT NULL DEFAULT 'uniform'
                        CHECK (title_mode IN ('uniform', 'filename')),
    uniform_title       TEXT NOT NULL DEFAULT '',
    video_type          TEXT NOT NULL DEFAULT 'short'
                        CHECK (video_type IN ('short', 'movie', 'episode', 'av')),
    site_category       TEXT NOT NULL DEFAULT '',
    description         TEXT NOT NULL DEFAULT '',
    tags                TEXT[] NOT NULL DEFAULT '{}',
    collection_ids      UUID[] NOT NULL DEFAULT '{}',
    actor_ids           UUID[] NOT NULL DEFAULT '{}',
    actor_names         TEXT[] NOT NULL DEFAULT '{}',
    image_collection_id UUID REFERENCES collections_images(id) ON DELETE SET NULL,
    forced_count        INTEGER NOT NULL DEFAULT 0,
    created_by          UUID,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- 视频合集只属于短视频：服务层已拦，这里做最后一道兜底。
    CONSTRAINT cloud_ingest_batches_collections_only_for_short
        CHECK (video_type = 'short' OR cardinality(collection_ids) = 0)
);

CREATE INDEX IF NOT EXISTS cloud_ingest_batches_created_at_idx
    ON cloud_ingest_batches (created_at DESC);

ALTER TABLE cloud_download_tasks
    ADD COLUMN IF NOT EXISTS ingest_batch_id UUID
        REFERENCES cloud_ingest_batches(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS ingest_title TEXT;

CREATE INDEX IF NOT EXISTS cloud_download_tasks_ingest_batch_idx
    ON cloud_download_tasks (ingest_batch_id, created_at DESC);

-- 自动入库修复扫描：只看已完成且挂了批的任务。
CREATE INDEX IF NOT EXISTS cloud_download_tasks_awaiting_ingest_idx
    ON cloud_download_tasks (id)
    WHERE ingest_batch_id IS NOT NULL AND status = 'completed';
