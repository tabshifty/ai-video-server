ALTER TABLE tv_remote_sessions
    ADD COLUMN IF NOT EXISTS autoplay_next_enabled BOOLEAN NOT NULL DEFAULT TRUE;
