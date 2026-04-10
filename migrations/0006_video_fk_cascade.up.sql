ALTER TABLE transcoding_jobs
    DROP CONSTRAINT IF EXISTS transcoding_jobs_video_id_fkey;

ALTER TABLE transcoding_jobs
    ADD CONSTRAINT transcoding_jobs_video_id_fkey
        FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE;

ALTER TABLE user_video_actions
    DROP CONSTRAINT IF EXISTS user_video_actions_video_id_fkey;

ALTER TABLE user_video_actions
    ADD CONSTRAINT user_video_actions_video_id_fkey
        FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE;
