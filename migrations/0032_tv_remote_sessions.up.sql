ALTER TABLE tv_devices
    ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS tv_remote_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(128) NOT NULL,
    platform VARCHAR(32) NOT NULL DEFAULT 'android_tv',
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    items JSONB NOT NULL DEFAULT '[]'::jsonb,
    current_index INT NOT NULL DEFAULT 0,
    current_video_id UUID,
    ended_reason VARCHAR(32) NOT NULL DEFAULT '',
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT tv_remote_sessions_platform_check CHECK (platform IN ('android_tv')),
    CONSTRAINT tv_remote_sessions_status_check CHECK (status IN ('active','ended')),
    CONSTRAINT tv_remote_sessions_items_array_check CHECK (jsonb_typeof(items) = 'array'),
    CONSTRAINT tv_remote_sessions_current_index_check CHECK (current_index >= 0)
);

CREATE INDEX IF NOT EXISTS idx_tv_remote_sessions_user_updated_at
    ON tv_remote_sessions(user_id, updated_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tv_remote_sessions_device_active_unique
    ON tv_remote_sessions(device_id, platform)
    WHERE status = 'active';
