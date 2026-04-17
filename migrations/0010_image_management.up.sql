CREATE TABLE IF NOT EXISTS images (
    id UUID PRIMARY KEY,
    user_id UUID,
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'ready' CHECK (status IN ('ready','failed')),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    original_path TEXT NOT NULL,
    stored_path TEXT NOT NULL,
    original_mime VARCHAR(120) NOT NULL DEFAULT '',
    stored_mime VARCHAR(120) NOT NULL DEFAULT '',
    original_ext VARCHAR(20) NOT NULL DEFAULT '',
    stored_ext VARCHAR(20) NOT NULL DEFAULT '',
    file_size BIGINT NOT NULL DEFAULT 0,
    width INT NOT NULL DEFAULT 0,
    height INT NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_images_status_created ON images(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_images_active_created ON images(active, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_images_user_created ON images(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS image_hashes (
    hash VARCHAR(64) NOT NULL,
    file_size BIGINT NOT NULL,
    image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (hash, file_size)
);

CREATE INDEX IF NOT EXISTS idx_image_hashes_image_id ON image_hashes(image_id);

CREATE TABLE IF NOT EXISTS image_actors (
    image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL REFERENCES actors(id),
    source VARCHAR(40) NOT NULL DEFAULT 'manual',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (image_id, actor_id)
);

CREATE INDEX IF NOT EXISTS idx_image_actors_actor_id ON image_actors(actor_id);

CREATE TABLE IF NOT EXISTS collections_images (
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

CREATE UNIQUE INDEX IF NOT EXISTS idx_collections_images_normalized_name ON collections_images(normalized_name);
CREATE INDEX IF NOT EXISTS idx_collections_images_active_sort ON collections_images(active, sort_order DESC, updated_at DESC);

CREATE TABLE IF NOT EXISTS image_collections (
    image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    collection_id UUID NOT NULL REFERENCES collections_images(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (image_id, collection_id)
);

CREATE INDEX IF NOT EXISTS idx_image_collections_collection_id ON image_collections(collection_id);

CREATE TABLE IF NOT EXISTS image_variants (
    id BIGSERIAL PRIMARY KEY,
    image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    transform_key VARCHAR(160) NOT NULL,
    variant_path TEXT NOT NULL,
    width INT NOT NULL DEFAULT 0,
    height INT NOT NULL DEFAULT 0,
    mime VARCHAR(120) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (image_id, transform_key)
);

CREATE INDEX IF NOT EXISTS idx_image_variants_image_id ON image_variants(image_id);
