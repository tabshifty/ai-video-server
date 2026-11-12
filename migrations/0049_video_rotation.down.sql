-- 回滚前应先等待所有修正及清理任务终止，避免遗留未引用文件。
DROP TRIGGER IF EXISTS video_rotation_mutation_guard ON videos;
DROP FUNCTION IF EXISTS guard_video_rotation_mutation();
DROP TABLE IF EXISTS video_rotation_jobs;
