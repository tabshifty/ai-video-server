package repository

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func cleanupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("COLLECTION_CLEANUP_TEST_DSN")
	if dsn == "" {
		t.Skip("设置 COLLECTION_CLEANUP_TEST_DSN 运行隔离 PostgreSQL 集成测试")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ConnConfig.Host != "127.0.0.1" && cfg.ConnConfig.Host != "localhost" {
		t.Fatal("仅允许本机测试库")
	}
	ctx := context.Background()
	admin, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	name := "collection_cleanup_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	cfg.ConnConfig.Database = name
	cfg.MaxConns = 8
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		_, err := admin.Exec(context.Background(), "DROP DATABASE "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)")
		admin.Close()
		if err != nil {
			t.Error(err)
		}
	})
	paths, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil || len(paths) == 0 {
		t.Fatal("缺少迁移", err)
	}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, strings.TrimPrefix(string(raw), "\ufeff")); err != nil {
			t.Fatalf("迁移 %s: %v", path, err)
		}
	}
	return pool
}

func TestEmptyCollectionsPostgresProtectionAndSnapshot(t *testing.T) {
	pool := cleanupTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	repo := NewVideoRepository(pool)
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}
	for _, kind := range []string{"video", "image"} {
		t.Run(kind, func(t *testing.T) {
			table := "collections"
			field := "video_collection_ids"
			if kind == "image" {
				table = "collections_images"
				field = "image_collection_ids"
			}
			add := func(active bool) uuid.UUID {
				t.Helper()
				id := uuid.New()
				exec("INSERT INTO "+table+"(id,name,normalized_name,active) VALUES($1,$2,$2,$3)", id, id.String(), active)
				return id
			}
			free, inactive, member, draft, historical, cross := add(true), add(false), add(true), add(true), add(true), add(true)
			video, image := uuid.New(), uuid.New()
			exec("INSERT INTO videos(id,title,type,status) VALUES($1,'失败视频','short','failed')", video)
			exec("INSERT INTO images(id,title,status,active,original_path,stored_path) VALUES($1,'停用图片','failed',false,'fixture','fixture')", image)
			if kind == "video" {
				exec("INSERT INTO video_collections(video_id,collection_id) VALUES($1,$2)", video, member)
				ic := uuid.New()
				exec("INSERT INTO collections_images(id,name,normalized_name) VALUES($1,$2,$2)", ic, ic.String())
				exec("INSERT INTO video_collection_image_collections(video_collection_id,image_collection_id) VALUES($1,$2)", cross, ic)
			} else {
				exec("INSERT INTO image_collections(image_id,collection_id) VALUES($1,$2)", image, member)
				exec("UPDATE videos SET image_collection_id=$1 WHERE id=$2", cross, video)
			}
			batch := uuid.New()
			exec("INSERT INTO archive_import_batches(id,title,original_filename,archive_format,original_path,status,default_"+field+") VALUES($1,'历史批次','a.zip','zip','fixture','completed',jsonb_build_array($2::text))", batch, historical.String())
			exec("INSERT INTO archive_import_workspaces(batch_id,draft) VALUES($1,jsonb_build_object('files',jsonb_build_object('f',jsonb_build_object('config',jsonb_build_object($2::text,jsonb_build_array($3::text))))))", batch, field, draft.String())
			ids, err := repo.PreviewEmptyCollections(ctx, kind)
			if err != nil {
				t.Fatal(err)
			}
			for _, id := range []uuid.UUID{free, inactive} {
				if !slices.Contains(ids, id) {
					t.Fatalf("未包含空合集 %v", id)
				}
			}
			for _, id := range []uuid.UUID{member, draft, historical, cross} {
				if slices.Contains(ids, id) {
					t.Fatalf("误选有成员或引用合集 %v", id)
				}
			}
			// 预览后新增成员、草稿引用与新空合集，均不能误删。
			if kind == "video" {
				exec("INSERT INTO video_collections(video_id,collection_id) VALUES($1,$2)", video, free)
			} else {
				exec("INSERT INTO image_collections(image_id,collection_id) VALUES($1,$2)", image, free)
			}
			exec("UPDATE archive_import_workspaces SET draft=draft || jsonb_build_object('template',jsonb_build_object($2::text,jsonb_build_array($3::text))) WHERE batch_id=$1", batch, field, inactive.String())
			late := add(true)
			result, err := repo.CleanupEmptyCollections(ctx, kind, ids)
			if err != nil {
				t.Fatal(err)
			}
			if result.DeletedCount != 0 || result.SkippedCount != len(ids) {
				t.Fatalf("结果 %+v", result)
			}
			exec("DELETE FROM archive_import_workspaces WHERE batch_id=$1", batch)
			result, err = repo.CleanupEmptyCollections(ctx, kind, ids)
			if err != nil || result.DeletedCount != 1 || result.SkippedCount != len(ids)-1 {
				t.Fatalf("结果 %+v %v", result, err)
			}
			result, err = repo.CleanupEmptyCollections(ctx, kind, ids)
			if err != nil || result.DeletedCount != 0 || result.SkippedCount != len(ids) {
				t.Fatalf("重复执行 %+v %v", result, err)
			}
			var exists bool
			if err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM "+table+" WHERE id=$1)", late).Scan(&exists); err != nil || !exists {
				t.Fatalf("预览外合集被删: %v", err)
			}
			var media int
			if err = pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM videos WHERE id=$1)+(SELECT count(*) FROM images WHERE id=$2)", video, image).Scan(&media); err != nil || media != 2 {
				t.Fatal("媒体被删除", err)
			}
		})
	}
}

func TestEmptyCollectionsPostgresAllConfigAndFormalReferences(t *testing.T) {
	pool := cleanupTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	repo := NewVideoRepository(pool)
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}
	user, video := uuid.New(), uuid.New()
	exec("INSERT INTO users(id,username,email,password_hash) VALUES($1,'cleanup','cleanup@localhost','unused')", user)
	exec("INSERT INTO videos(id,title,type,status,user_id) VALUES($1,'视频','short','ready',$2)", video, user)
	for _, kind := range []string{"video", "image"} {
		table, field := "collections", "video_collection_ids"
		if kind == "image" {
			table, field = "collections_images", "image_collection_ids"
		}
		for _, source := range []string{"batch", "file", "group", "workspace-template", "workspace-group", "workspace-file", "job", "formal"} {
			t.Run(kind+"/"+source, func(t *testing.T) {
				id, batch, file := uuid.New(), uuid.New(), uuid.New()
				exec("INSERT INTO "+table+"(id,name,normalized_name) VALUES($1,$2,$2)", id, id.String())
				exec("INSERT INTO archive_import_batches(id,title,original_filename,archive_format,original_path,status) VALUES($1,'历史批次','a.zip','zip','fixture','completed')", batch)
				exec("INSERT INTO archive_import_files(id,batch_id,relative_path,file_path,entry_type,media_kind,status) VALUES($1,$2,'a.mp4','fixture','file','video','ready')", file, batch)
				switch source {
				case "batch":
					exec("UPDATE archive_import_batches SET default_"+field+"=jsonb_build_array($2::text) WHERE id=$1", batch, id.String())
				case "file":
					exec("UPDATE archive_import_files SET "+field+"=jsonb_build_array($2::text) WHERE id=$1", file, id.String())
				case "group":
					exec("INSERT INTO archive_import_groups(batch_id,name,media_kind,"+field+") VALUES($1,'历史分组','video',jsonb_build_array($2::text))", batch, id.String())
				case "workspace-template", "workspace-group", "workspace-file":
					config := `{"` + field + `":["` + strings.ToUpper(id.String()) + `"]}`
					draft := `{"template":` + config + `}`
					if source == "workspace-group" {
						draft = `{"groups":[{"config":` + config + `}]}`
					}
					if source == "workspace-file" {
						draft = `{"files":{"id":{"config":` + config + `}}}`
					}
					exec("INSERT INTO archive_import_workspaces(batch_id,draft) VALUES($1,$2::jsonb)", batch, draft)
				case "job":
					exec("INSERT INTO archive_import_workspace_jobs(id,batch_id,file_id,snapshot,state) VALUES($1,$2,$3,jsonb_build_object($4::text,jsonb_build_array($5::text)),'completed')", uuid.New(), batch, file, field, id.String())
				case "formal":
					if kind == "image" {
						exec("INSERT INTO video_image_collections(video_id,image_collection_id) VALUES($1,$2)", video, id)
					} else {
						exec("INSERT INTO tv_remote_sessions(id,user_id,device_id,search_context,status) VALUES($1,$2,$3,jsonb_build_object('collection_id',$4::text),'ended')", uuid.New(), user, uuid.NewString(), id.String())
					}
				}
				ids, err := repo.PreviewEmptyCollections(ctx, kind)
				if err != nil {
					t.Fatal(err)
				}
				if slices.Contains(ids, id) {
					t.Fatal("有引用的合集被选中")
				}
				result, err := repo.CleanupEmptyCollections(ctx, kind, []uuid.UUID{id})
				if err != nil || result.DeletedCount != 0 || result.SkippedCount != 1 {
					t.Fatalf("有引用的合集被清理 %+v %v", result, err)
				}
			})
		}
	}
	// 两类合集之间的关系保护双方；无成员也不能删除。
	vc, ic := uuid.New(), uuid.New()
	exec("INSERT INTO collections(id,name,normalized_name) VALUES($1,$2,$2)", vc, vc.String())
	exec("INSERT INTO collections_images(id,name,normalized_name) VALUES($1,$2,$2)", ic, ic.String())
	exec("INSERT INTO video_collection_image_collections(video_collection_id,image_collection_id) VALUES($1,$2)", vc, ic)
	for kind, id := range map[string]uuid.UUID{"video": vc, "image": ic} {
		result, err := repo.CleanupEmptyCollections(ctx, kind, []uuid.UUID{id})
		if err != nil || result.DeletedCount != 0 {
			t.Fatalf("合集关联保护 %s: %+v %v", kind, result, err)
		}
	}
}

func TestEmptyCollectionsPostgresMigrationCompatibility(t *testing.T) {
	pool := cleanupTestDB(t)
	ctx := context.Background()
	apply := func(name string) {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join("../../migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, string(raw)); err != nil {
			t.Fatal(err)
		}
	}
	apply("0050_empty_collection_cleanup.down.sql")
	batch := uuid.New()
	// 升级前已有失效引用不得阻止迁移或无关配置更新。
	if _, err := pool.Exec(ctx, `INSERT INTO archive_import_batches(id,title,original_filename,archive_format,original_path,default_video_collection_ids) VALUES($1,'旧批次','a.zip','zip','fixture','["not-a-uuid"]')`, batch); err != nil {
		t.Fatal(err)
	}
	apply("0050_empty_collection_cleanup.up.sql")
	if _, err := pool.Exec(ctx, `UPDATE archive_import_batches SET title='标题更新', default_video_collection_ids=default_video_collection_ids WHERE id=$1`, batch); err != nil {
		t.Fatal("历史引用兼容", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE archive_import_batches SET default_image_collection_ids=jsonb_build_array($2::text) WHERE id=$1`, batch, uuid.NewString()); err == nil {
		t.Fatal("接受了新加入的失效引用")
	}
	if _, err := NewVideoRepository(pool).PreviewEmptyCollections(ctx, "video"); err != nil {
		t.Fatal(err)
	}
	apply("0050_empty_collection_cleanup.down.sql")
	apply("0050_empty_collection_cleanup.up.sql")
}

func waitForCleanupLock(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.NewTimer(3 * time.Second)
	defer timeout.Stop()
	for {
		var waiting bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND pid<>pg_backend_pid() AND wait_event_type='Lock')`).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			return
		}
		select {
		case <-ticker.C:
		case <-timeout.C:
			t.Fatal("未观察到并发引用锁等待")
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
}

func TestEmptyCollectionsPostgresConcurrentConfigWrites(t *testing.T) {
	pool := cleanupTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	repo := NewVideoRepository(pool)
	for _, kind := range []string{"video", "image"} {
		t.Run(kind, func(t *testing.T) {
			table, field := "collections", "default_video_collection_ids"
			if kind == "image" {
				table, field = "collections_images", "default_image_collection_ids"
			}
			id, batch := uuid.New(), uuid.New()
			if _, err := pool.Exec(ctx, "INSERT INTO "+table+"(id,name,normalized_name) VALUES($1,$2,$2)", id, id.String()); err != nil {
				t.Fatal(err)
			}
			if _, err := pool.Exec(ctx, "INSERT INTO archive_import_batches(id,title,original_filename,archive_format,original_path) VALUES($1,'批次','a.zip','zip','fixture')", batch); err != nil {
				t.Fatal(err)
			}
			ids, err := repo.PreviewEmptyCollections(ctx, kind)
			if err != nil {
				t.Fatal(err)
			}
			tx, err := pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			if _, err = tx.Exec(ctx, "UPDATE archive_import_batches SET "+field+"=jsonb_build_array($2::text) WHERE id=$1", batch, id.String()); err != nil {
				t.Fatal(err)
			}
			type outcome struct {
				result CollectionCleanupResult
				err    error
			}
			done := make(chan outcome, 1)
			go func() { result, err := repo.CleanupEmptyCollections(ctx, kind, ids); done <- outcome{result, err} }()
			waitForCleanupLock(t, pool, ctx)
			if err = tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			result := <-done
			if result.err != nil || result.result.DeletedCount != 0 || result.result.SkippedCount != len(ids) {
				t.Fatalf("并发新增引用被误删 %+v", result)
			}
			// 反方向：清理先取得合集锁，等待的配置写入在删除提交后必须失败。
			if _, err = pool.Exec(ctx, "UPDATE archive_import_batches SET "+field+"='[]' WHERE id=$1", batch); err != nil {
				t.Fatal(err)
			}
			tx, err = pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			if _, err = tx.Exec(ctx, "SELECT id FROM "+table+" WHERE id=$1 FOR UPDATE", id); err != nil {
				t.Fatal(err)
			}
			writer := make(chan error, 1)
			go func() {
				_, err := pool.Exec(ctx, "UPDATE archive_import_batches SET "+field+"=jsonb_build_array($2::text) WHERE id=$1", batch, id.String())
				writer <- err
			}()
			waitForCleanupLock(t, pool, ctx)
			if _, err = tx.Exec(ctx, "DELETE FROM "+table+" WHERE id=$1", id); err != nil {
				t.Fatal(err)
			}
			if err = tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			err = <-writer
			if err == nil {
				t.Fatal("清理后并发写入了失效配置引用")
			}
		})
	}
}
