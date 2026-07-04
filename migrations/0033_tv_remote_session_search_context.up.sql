ALTER TABLE tv_remote_sessions
    ADD COLUMN IF NOT EXISTS search_context JSONB;

ALTER TABLE tv_remote_sessions
    DROP CONSTRAINT IF EXISTS tv_remote_sessions_search_context_object_check;

ALTER TABLE tv_remote_sessions
    ADD CONSTRAINT tv_remote_sessions_search_context_object_check
    CHECK (search_context IS NULL OR jsonb_typeof(search_context) = 'object');
