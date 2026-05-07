CREATE TABLE IF NOT EXISTS tv_devices (
    id UUID PRIMARY KEY,
    device_id VARCHAR(128) NOT NULL,
    device_name VARCHAR(255) NOT NULL,
    platform VARCHAR(32) NOT NULL DEFAULT 'android_tv',
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    last_authorized_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT tv_devices_platform_check CHECK (platform IN ('android_tv'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tv_devices_device_platform
    ON tv_devices(device_id, platform);

CREATE TABLE IF NOT EXISTS tv_auth_sessions (
    id UUID PRIMARY KEY,
    pair_code VARCHAR(12) NOT NULL UNIQUE,
    device_id VARCHAR(128) NOT NULL,
    device_name VARCHAR(255) NOT NULL,
    platform VARCHAR(32) NOT NULL DEFAULT 'android_tv',
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    access_token TEXT NOT NULL DEFAULT '',
    refresh_token TEXT NOT NULL DEFAULT '',
    approved_username VARCHAR(50) NOT NULL DEFAULT '',
    approved_role VARCHAR(20) NOT NULL DEFAULT '',
    approved_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT tv_auth_sessions_status_check CHECK (status IN ('pending','approved','expired','denied')),
    CONSTRAINT tv_auth_sessions_platform_check CHECK (platform IN ('android_tv'))
);

CREATE INDEX IF NOT EXISTS idx_tv_auth_sessions_device_id
    ON tv_auth_sessions(device_id);

CREATE INDEX IF NOT EXISTS idx_tv_auth_sessions_status_expires_at
    ON tv_auth_sessions(status, expires_at);
