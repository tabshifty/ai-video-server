ALTER TABLE videos
    ADD COLUMN IF NOT EXISTS views_count BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS likes_count BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS favorites_count BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_videos_status_created ON videos(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_videos_type_status_created ON videos(type, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_video_actions_user_video_action ON user_video_actions(user_id, video_id, action_type);
CREATE INDEX IF NOT EXISTS idx_user_video_actions_video_action ON user_video_actions(video_id, action_type);
