CREATE TABLE IF NOT EXISTS collected_forum_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source VARCHAR(64) NOT NULL,
    board_key VARCHAR(128) NOT NULL,
    external_post_id VARCHAR(128) NOT NULL,
    title TEXT,
    url TEXT,
    inspection_status VARCHAR(24) NOT NULL,
    filter_decision VARCHAR(16),
    filter_reasons JSONB NOT NULL DEFAULT '[]'::jsonb,
    fetch_method VARCHAR(64),
    error_summary TEXT,
    result_fingerprint CHAR(64),
    observed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    inspected_at TIMESTAMPTZ,
    UNIQUE (source, external_post_id),
    CONSTRAINT collected_forum_posts_status_check
        CHECK (inspection_status IN ('dedupe_only', 'pending', 'inspected', 'restricted', 'failed')),
    CONSTRAINT collected_forum_posts_filter_decision_check
        CHECK (filter_decision IS NULL OR filter_decision IN ('included', 'excluded')),
    CONSTRAINT collected_forum_posts_filter_reasons_array_check
        CHECK (jsonb_typeof(filter_reasons) = 'array'),
    CONSTRAINT collected_forum_posts_state_check CHECK (
        (inspection_status = 'dedupe_only'
            AND title IS NULL AND url IS NULL AND observed_at IS NULL
            AND filter_decision IS NULL AND filter_reasons = '[]'::jsonb
            AND fetch_method IS NULL AND error_summary IS NULL
            AND result_fingerprint IS NULL AND inspected_at IS NULL)
        OR
        (inspection_status = 'pending'
            AND title IS NOT NULL AND title <> '' AND url IS NOT NULL AND url <> ''
            AND observed_at IS NOT NULL
            AND filter_decision IS NULL AND filter_reasons = '[]'::jsonb
            AND fetch_method IS NULL AND error_summary IS NULL
            AND result_fingerprint IS NULL AND inspected_at IS NULL)
        OR
        (inspection_status IN ('inspected', 'restricted', 'failed')
            AND title IS NOT NULL AND title <> '' AND url IS NOT NULL AND url <> ''
            AND observed_at IS NOT NULL
            AND filter_decision IS NOT NULL AND result_fingerprint IS NOT NULL AND inspected_at IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_collected_forum_posts_created_at
    ON collected_forum_posts (created_at);

CREATE TABLE IF NOT EXISTS collected_forum_post_resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES collected_forum_posts(id) ON DELETE CASCADE,
    kind VARCHAR(16) NOT NULL CHECK (kind IN ('attachment', 'ed2k')),
    value TEXT NOT NULL,
    position INTEGER NOT NULL CHECK (position >= 0),
    UNIQUE (post_id, kind, position)
);
