-- 草稿与源文件物理位置分离；新列有默认值，旧二进制仍可写入批次。
ALTER TABLE archive_import_batches ADD COLUMN IF NOT EXISTS workspace_generation BOOLEAN NOT NULL DEFAULT FALSE;
CREATE TABLE IF NOT EXISTS archive_import_workspaces (
 batch_id UUID PRIMARY KEY REFERENCES archive_import_batches(id) ON DELETE CASCADE,
 version BIGINT NOT NULL DEFAULT 1,
 draft JSONB NOT NULL,
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS archive_import_workspace_jobs (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 batch_id UUID NOT NULL REFERENCES archive_import_batches(id) ON DELETE CASCADE,
 file_id UUID NOT NULL REFERENCES archive_import_files(id) ON DELETE CASCADE,
 snapshot JSONB NOT NULL,
 state TEXT NOT NULL DEFAULT 'queued' CHECK (state IN ('queued','running','completed','failed')),
 last_error TEXT NOT NULL DEFAULT '',
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_archive_workspace_active_file ON archive_import_workspace_jobs(file_id) WHERE state IN ('queued','running');
CREATE INDEX IF NOT EXISTS idx_archive_workspace_jobs_pending ON archive_import_workspace_jobs(batch_id,state,created_at);
