DROP TRIGGER IF EXISTS guard_collection_refs ON tv_remote_sessions;
DROP TRIGGER IF EXISTS guard_collection_refs ON archive_import_workspace_jobs;
DROP TRIGGER IF EXISTS guard_collection_refs ON archive_import_workspaces;
DROP TRIGGER IF EXISTS guard_collection_refs ON archive_import_groups;
DROP TRIGGER IF EXISTS guard_collection_refs ON archive_import_files;
DROP TRIGGER IF EXISTS guard_collection_refs ON archive_import_batches;
DROP FUNCTION IF EXISTS guard_collection_config_refs();
DROP VIEW IF EXISTS collection_config_references;
DROP FUNCTION IF EXISTS collection_config_refs(JSONB);
