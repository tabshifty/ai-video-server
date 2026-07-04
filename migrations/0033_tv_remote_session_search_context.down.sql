ALTER TABLE tv_remote_sessions
    DROP CONSTRAINT IF EXISTS tv_remote_sessions_search_context_object_check;

ALTER TABLE tv_remote_sessions
    DROP COLUMN IF EXISTS search_context;
