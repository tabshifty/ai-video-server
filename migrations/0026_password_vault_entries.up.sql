CREATE TABLE IF NOT EXISTS password_vault_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    account TEXT NOT NULL DEFAULT '',
    password_ciphertext TEXT NOT NULL,
    url TEXT NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_vault_entries_updated ON password_vault_entries(updated_at DESC);
