CREATE TABLE IF NOT EXISTS telegram_account_state (
    id SMALLINT PRIMARY KEY DEFAULT 1,
    telegram_user_id BIGINT,
    username TEXT NOT NULL DEFAULT '',
    first_name TEXT NOT NULL DEFAULT '',
    last_name TEXT NOT NULL DEFAULT '',
    phone_masked VARCHAR(32) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'unconfigured',
    last_error TEXT,
    authorized_at TIMESTAMPTZ,
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT telegram_account_state_singleton CHECK (id = 1),
    CONSTRAINT telegram_account_state_status_check CHECK (
        status IN ('unconfigured', 'authorizing', 'authorized', 'reauthorizing', 'error')
    )
);

CREATE TABLE IF NOT EXISTS telegram_ingestor_heartbeats (
    id SMALLINT PRIMARY KEY DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'stopped',
    account_status VARCHAR(20) NOT NULL DEFAULT 'unconfigured',
    version VARCHAR(128) NOT NULL DEFAULT '',
    error_summary TEXT,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT telegram_ingestor_heartbeats_singleton CHECK (id = 1),
    CONSTRAINT telegram_ingestor_heartbeats_status_check CHECK (
        status IN ('running', 'authorizing', 'draining', 'error', 'stopped')
    )
);

CREATE TABLE IF NOT EXISTS telegram_authorizations (
    id UUID PRIMARY KEY,
    kind VARCHAR(12) NOT NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'pending',
    actor_user_id UUID NOT NULL REFERENCES users(id),
    telegram_user_id BIGINT,
    error_summary TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT telegram_authorizations_kind_check CHECK (kind IN ('phone', 'qr')),
    CONSTRAINT telegram_authorizations_status_check CHECK (
        status IN (
            'pending', 'awaiting_code', 'awaiting_password', 'scanning',
            'succeeded', 'failed', 'cancelled', 'expired'
        )
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_telegram_authorizations_active
    ON telegram_authorizations ((1))
    WHERE status IN ('pending', 'awaiting_code', 'awaiting_password', 'scanning');

CREATE INDEX IF NOT EXISTS idx_telegram_authorizations_actor_created
    ON telegram_authorizations(actor_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_telegram_authorizations_expires_at
    ON telegram_authorizations(expires_at);

CREATE TABLE IF NOT EXISTS telegram_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    actor_user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(64) NOT NULL,
    target_type VARCHAR(32) NOT NULL DEFAULT '',
    target_id VARCHAR(128) NOT NULL DEFAULT '',
    result VARCHAR(20) NOT NULL DEFAULT 'succeeded',
    summary JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT telegram_audit_logs_result_check CHECK (
        result IN ('succeeded', 'failed', 'cancelled')
    ),
    CONSTRAINT telegram_audit_logs_summary_object CHECK (
        jsonb_typeof(summary) = 'object'
    )
);

CREATE INDEX IF NOT EXISTS idx_telegram_audit_logs_created_at
    ON telegram_audit_logs(created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_telegram_audit_logs_action_result
    ON telegram_audit_logs(action, result, created_at DESC);
