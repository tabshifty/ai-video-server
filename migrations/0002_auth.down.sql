DROP INDEX IF EXISTS idx_transcoding_jobs_user_id;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;

ALTER TABLE transcoding_jobs DROP CONSTRAINT IF EXISTS fk_transcoding_jobs_user_id;
ALTER TABLE transcoding_jobs DROP COLUMN IF EXISTS user_id;

ALTER TABLE user_video_actions DROP CONSTRAINT IF EXISTS fk_user_video_actions_user_id;
ALTER TABLE videos DROP CONSTRAINT IF EXISTS fk_videos_user_id;

DROP TABLE IF EXISTS users;
