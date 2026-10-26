-- 降级前先排空工作台入库任务。优先回滚二进制并保留新增表，避免丢失草稿。
DROP TABLE IF EXISTS archive_import_workspace_jobs;
DROP TABLE IF EXISTS archive_import_workspaces;
ALTER TABLE archive_import_batches DROP COLUMN IF EXISTS workspace_generation;
