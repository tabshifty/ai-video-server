DROP INDEX IF EXISTS idx_user_video_actions_video_action;
DROP INDEX IF EXISTS idx_user_video_actions_user_video_action;
DROP INDEX IF EXISTS idx_videos_type_status_created;
DROP INDEX IF EXISTS idx_videos_status_created;

ALTER TABLE videos
    DROP COLUMN IF EXISTS favorites_count,
    DROP COLUMN IF EXISTS likes_count,
    DROP COLUMN IF EXISTS views_count;
