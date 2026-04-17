CREATE TABLE IF NOT EXISTS collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(120) NOT NULL,
    normalized_name VARCHAR(160) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    cover_url TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_collections_normalized_name ON collections(normalized_name);
CREATE INDEX IF NOT EXISTS idx_collections_active_sort ON collections(active, sort_order DESC, updated_at DESC);

CREATE TABLE IF NOT EXISTS video_collections (
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (video_id, collection_id)
);

CREATE INDEX IF NOT EXISTS idx_video_collections_collection_id ON video_collections(collection_id);
