CREATE TABLE IF NOT EXISTS video_rotation_jobs (
 id UUID PRIMARY KEY,
 video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
 requested_by UUID NOT NULL,
 angle INT NOT NULL CHECK(angle IN (90,180,270)),
 mode TEXT NOT NULL CHECK(mode IN ('lossless','reencode')),
 status TEXT NOT NULL DEFAULT 'queued' CHECK(status IN ('queued','running','cleanup','succeeded','failed')),
 input_path TEXT NOT NULL,
 input_size BIGINT NOT NULL,
 input_mtime BIGINT NOT NULL,
 original_path TEXT NOT NULL DEFAULT '',
 old_thumbnail TEXT NOT NULL DEFAULT '',
 output_path TEXT NOT NULL,
 thumbnail_path TEXT NOT NULL,
 progress DOUBLE PRECISION NOT NULL DEFAULT 0,
 error_message TEXT NOT NULL DEFAULT '',
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS video_rotation_active ON video_rotation_jobs(video_id) WHERE status IN ('queued','running','cleanup');
CREATE INDEX IF NOT EXISTS video_rotation_recent ON video_rotation_jobs(video_id,created_at DESC);

-- 同一视频的所有物理变更统一受数据库约束，旧版程序仍可操作没有修正任务的视频。
CREATE OR REPLACE FUNCTION guard_video_rotation_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_OP = 'UPDATE' THEN
  IF (NEW.status,NEW.type,NEW.original_path,NEW.transcoded_path,NEW.thumbnail_path)
    IS NOT DISTINCT FROM (OLD.status,OLD.type,OLD.original_path,OLD.transcoded_path,OLD.thumbnail_path) THEN
   RETURN NEW;
  END IF;
 END IF;
 IF EXISTS(SELECT 1 FROM video_rotation_jobs WHERE video_id=OLD.id AND status IN ('queued','running','cleanup')
   AND id::text IS DISTINCT FROM current_setting('app.video_rotation_job',true)) THEN
  RAISE EXCEPTION '视频方向修正或旧文件清理进行中，请稍后操作' USING ERRCODE='55000';
 END IF;
 IF TG_OP='DELETE' THEN RETURN OLD; END IF;
 RETURN NEW;
END $$;
DROP TRIGGER IF EXISTS video_rotation_mutation_guard ON videos;
CREATE TRIGGER video_rotation_mutation_guard BEFORE UPDATE OR DELETE ON videos FOR EACH ROW EXECUTE FUNCTION guard_video_rotation_mutation();
