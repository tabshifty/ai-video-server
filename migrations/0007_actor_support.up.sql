CREATE TABLE IF NOT EXISTS actors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(120) NOT NULL,
    normalized_name VARCHAR(160) NOT NULL,
    aliases TEXT[] NOT NULL DEFAULT '{}'::text[],
    gender VARCHAR(20) NOT NULL DEFAULT '',
    country VARCHAR(80) NOT NULL DEFAULT '',
    birth_date DATE,
    avatar_url TEXT NOT NULL DEFAULT '',
    source VARCHAR(40) NOT NULL DEFAULT 'manual',
    external_id VARCHAR(120) NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_actors_normalized_name ON actors(normalized_name);
CREATE INDEX IF NOT EXISTS idx_actors_active_name ON actors(active, name);

CREATE TABLE IF NOT EXISTS video_actors (
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL REFERENCES actors(id),
    source VARCHAR(40) NOT NULL DEFAULT 'manual',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (video_id, actor_id)
);

CREATE INDEX IF NOT EXISTS idx_video_actors_actor_id ON video_actors(actor_id);
