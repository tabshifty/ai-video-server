ALTER TABLE cloud_download_tasks
    ADD COLUMN IF NOT EXISTS ingest_forced BOOLEAN NOT NULL DEFAULT false;

COMMENT ON COLUMN cloud_download_tasks.ingest_forced IS
    '该条在创建入库批时被管理员强制放行（后缀无法识别为视频）；只表示照常下载，不承诺能入库。';
