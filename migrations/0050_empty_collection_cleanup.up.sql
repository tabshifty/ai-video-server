-- 仅增加视图、函数和新引用检查；不回填或清除历史配置，旧程序可继续写合法引用。
CREATE OR REPLACE FUNCTION collection_config_refs(doc JSONB)
RETURNS TABLE(kind TEXT, collection_id TEXT)
LANGUAGE sql IMMUTABLE PARALLEL SAFE AS $$
 SELECT DISTINCT kind, lower(value #>> '{}') FROM (
  SELECT 'video'::text AS kind, value FROM jsonb_path_query(doc, '$.**.video_collection_ids[*]') AS paths(value)
  UNION ALL SELECT 'video', value FROM jsonb_path_query(doc, '$.**.default_video_collection_ids[*]') AS paths(value)
  UNION ALL SELECT 'video', value FROM jsonb_path_query(doc, '$.**.collection_id') AS paths(value)
  UNION ALL SELECT 'image', value FROM jsonb_path_query(doc, '$.**.image_collection_ids[*]') AS paths(value)
  UNION ALL SELECT 'image', value FROM jsonb_path_query(doc, '$.**.default_image_collection_ids[*]') AS paths(value)
  UNION ALL SELECT 'image', value FROM jsonb_path_query(doc, '$.**.image_collection_id') AS paths(value)
 ) refs WHERE jsonb_typeof(value) = 'string'
$$;

-- 保留的历史配置、已完成作业快照也属于业务引用，不按状态排除。
CREATE OR REPLACE VIEW collection_config_references AS
 SELECT refs.kind, refs.collection_id FROM (
  SELECT jsonb_build_object('default_video_collection_ids', default_video_collection_ids,
   'default_image_collection_ids', default_image_collection_ids) AS doc FROM archive_import_batches
  UNION ALL SELECT jsonb_build_object('video_collection_ids', video_collection_ids,
   'image_collection_ids', image_collection_ids) FROM archive_import_files
  UNION ALL SELECT jsonb_build_object('video_collection_ids', video_collection_ids,
   'image_collection_ids', image_collection_ids) FROM archive_import_groups
  UNION ALL SELECT draft FROM archive_import_workspaces
  UNION ALL SELECT snapshot FROM archive_import_workspace_jobs
  UNION ALL SELECT search_context FROM tv_remote_sessions
 ) configs CROSS JOIN LATERAL collection_config_refs(doc) refs;

CREATE OR REPLACE FUNCTION guard_collection_config_refs() RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE previous JSONB := '{}'::jsonb; current_doc JSONB; ref RECORD;
BEGIN
 IF TG_OP = 'UPDATE' THEN previous := to_jsonb(OLD); END IF;
 -- 与引用视图保持一致：只检查配置列，不扫描媒体 metadata 等无关 JSON。
 SELECT jsonb_object_agg(key, value) INTO previous FROM jsonb_each(previous)
  WHERE key IN ('default_video_collection_ids','default_image_collection_ids',
   'video_collection_ids','image_collection_ids','draft','snapshot','search_context');
 SELECT jsonb_object_agg(key, value) INTO current_doc FROM jsonb_each(to_jsonb(NEW))
  WHERE key IN ('default_video_collection_ids','default_image_collection_ids',
   'video_collection_ids','image_collection_ids','draft','snapshot','search_context');
 -- 只校验新加入的引用，允许旧配置中已有的失效 ID 在无关更新中原样保留。
 FOR ref IN
  SELECT * FROM collection_config_refs(current_doc)
  EXCEPT SELECT * FROM collection_config_refs(previous)
  ORDER BY 1, 2
 LOOP
  IF ref.kind = 'video' THEN
   PERFORM id FROM collections WHERE id = ref.collection_id::uuid FOR KEY SHARE;
  ELSE
   PERFORM id FROM collections_images WHERE id = ref.collection_id::uuid FOR KEY SHARE;
  END IF;
  IF NOT FOUND THEN
   RAISE EXCEPTION '合集不存在或已清理，请刷新后重新选择' USING ERRCODE = '23503';
  END IF;
 END LOOP;
 RETURN NEW;
END $$;

CREATE TRIGGER guard_collection_refs BEFORE INSERT OR UPDATE OF default_video_collection_ids, default_image_collection_ids ON archive_import_batches
 FOR EACH ROW EXECUTE FUNCTION guard_collection_config_refs();
CREATE TRIGGER guard_collection_refs BEFORE INSERT OR UPDATE OF video_collection_ids, image_collection_ids ON archive_import_files
 FOR EACH ROW EXECUTE FUNCTION guard_collection_config_refs();
CREATE TRIGGER guard_collection_refs BEFORE INSERT OR UPDATE OF video_collection_ids, image_collection_ids ON archive_import_groups
 FOR EACH ROW EXECUTE FUNCTION guard_collection_config_refs();
CREATE TRIGGER guard_collection_refs BEFORE INSERT OR UPDATE OF draft ON archive_import_workspaces
 FOR EACH ROW EXECUTE FUNCTION guard_collection_config_refs();
CREATE TRIGGER guard_collection_refs BEFORE INSERT OR UPDATE OF snapshot ON archive_import_workspace_jobs
 FOR EACH ROW EXECUTE FUNCTION guard_collection_config_refs();
CREATE TRIGGER guard_collection_refs BEFORE INSERT OR UPDATE OF search_context ON tv_remote_sessions
 FOR EACH ROW EXECUTE FUNCTION guard_collection_config_refs();
