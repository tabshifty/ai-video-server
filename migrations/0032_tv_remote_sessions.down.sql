DROP INDEX IF EXISTS idx_tv_remote_sessions_device_active_unique;
DROP INDEX IF EXISTS idx_tv_remote_sessions_user_updated_at;
DROP TABLE IF EXISTS tv_remote_sessions;

ALTER TABLE tv_devices
    DROP COLUMN IF EXISTS last_seen_at;
