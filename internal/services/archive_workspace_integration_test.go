package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"video-server/internal/repository"
)

// 仅使用显式指定的本机临时数据库；绝不读取项目 .env 或默认业务 DSN。
func workspaceIntegrationDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("ARCHIVE_WORKSPACE_TEST_DSN")
	if dsn == "" {
		t.Skip("设置 ARCHIVE_WORKSPACE_TEST_DSN 可运行本机 PostgreSQL 集成测试")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ConnConfig.Host != "127.0.0.1" && cfg.ConnConfig.Host != "localhost" && cfg.ConnConfig.Host != "::1" {
		t.Fatal("集成测试只允许本机 PostgreSQL")
	}
	ctx := context.Background()
	admin, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	dbName := "archive_workspace_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{dbName}.Sanitize()); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	cfg.ConnConfig.Database = dbName
	cfg.MaxConns = 12
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		_, _ = admin.Exec(context.Background(), "DROP DATABASE "+pgx.Identifier{dbName}.Sanitize()+" WITH (FORCE)")
		admin.Close()
	})
	paths, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil || len(paths) == 0 {
		t.Fatalf("迁移文件: %v", err)
	}
	for _, p := range paths {
		raw, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = pool.Exec(ctx, strings.TrimPrefix(string(raw), "\ufeff")); e != nil {
			t.Fatalf("迁移 %s: %v", p, e)
		}
	}
	// 新迁移可重复执行，并能在隔离库中完整撤销、重新应用。
	for _, name := range []string{"0047_archive_workspace.up.sql", "0047_archive_workspace.down.sql", "0047_archive_workspace.up.sql"} {
		raw, e := os.ReadFile(filepath.Join("../../migrations", name))
		if e != nil {
			t.Fatal(e)
		}
		if _, e = pool.Exec(ctx, string(raw)); e != nil {
			t.Fatalf("迁移回滚/重放 %s: %v", name, e)
		}
	}
	return pool
}

type workspaceIntegrationQueue struct {
	sync.Mutex
	transcodes    int
	failTranscode bool
	wake          chan uuid.UUID
	failWake      bool
}

func (q *workspaceIntegrationQueue) EnqueueArchiveWorkspace(id uuid.UUID) error {
	q.Lock()
	defer q.Unlock()
	if q.wake != nil {
		select {
		case q.wake <- id:
		default:
		}
	}
	if q.failWake {
		return errors.New("模拟 Redis 暂不可用")
	}
	return nil
}
func (q *workspaceIntegrationQueue) EnqueueTranscode(string, string, string, string, bool) error {
	q.Lock()
	defer q.Unlock()
	q.transcodes++
	if q.failTranscode {
		return errors.New("模拟转码入队失败")
	}
	return nil
}
func (*workspaceIntegrationQueue) EnqueueArchiveExtract(uuid.UUID, string) error   { return nil }
func (*workspaceIntegrationQueue) EnqueueScrapeMovie(string, string, string) error { return nil }
func (*workspaceIntegrationQueue) EnqueueScrapeTV(string, string, string) error    { return nil }
func (*workspaceIntegrationQueue) EnqueueScrapeAV(string, string, string) error    { return nil }

func TestArchiveWorkspacePostgresLifecycle(t *testing.T) {
	pool := workspaceIntegrationDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	root := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := repository.NewVideoRepository(pool)
	queue := &workspaceIntegrationQueue{failWake: true, wake: make(chan uuid.UUID, 20)}
	svc := NewArchiveImportService(pool, NewUploadService(repo, root, root, logger), NewImageService(repo, root, root, logger), repo, queue, root, root, logger)
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO users(id,username,email,password_hash) VALUES($1,'archive-test','archive-test@localhost','unused')`, uuid.Nil)
	fixture := func(fresh bool) (uuid.UUID, []uuid.UUID) {
		t.Helper()
		batch := uuid.New()
		dir := filepath.Join(root, batch.String())
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		exec(`INSERT INTO archive_import_batches(id,title,original_filename,archive_format,original_path,extracted_dir,status,default_title_prefix,workspace_generation) VALUES($1,'上传标题','素材.zip','zip',$2,$3,'ready','上传标题',$4)`, batch, filepath.Join(dir, "素材.zip"), dir, fresh)
		ids := []uuid.UUID{uuid.New(), uuid.New()}
		for i, id := range ids {
			name := []string{"第一.mp4", "第二.mp4"}[i]
			p := filepath.Join(dir, name)
			if err := os.WriteFile(p, []byte("media-bytes-"+name), 0600); err != nil {
				t.Fatal(err)
			}
			exec(`INSERT INTO archive_import_files(id,batch_id,relative_path,file_path,entry_type,media_kind,file_size,title) VALUES($1,$2,$3,$4,'file','video',$5,'旧文件标题')`, id, batch, "父/子/"+name, p, len("media-bytes-"+name))
		}
		return batch, ids
	}
	clone := func(d ArchiveWorkspaceDraft) ArchiveWorkspaceDraft {
		t.Helper()
		raw, _ := json.Marshal(d)
		var result ArchiveWorkspaceDraft
		if err := json.Unmarshal(raw, &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	batch, ids := fixture(true)
	r, err := svc.GetArchiveWorkspace(ctx, batch)
	if err != nil {
		t.Fatal(err)
	}
	if r.Version != 1 || len(r.Draft.Groups) != 3 {
		t.Fatalf("新工作台初始化: %+v", r)
	}
	if _, err = svc.AcquireLegacyWorkspaceGuard(ctx, "batch", batch); err == nil {
		t.Fatal("新工作台仍可被旧接口修改")
	}
	title := "快照标题"
	r.Draft.Groups[0].Config.Title = &title
	r, err = svc.SaveArchiveWorkspace(ctx, batch, ArchiveWorkspaceSave{Version: r.Version, Draft: r.Draft})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.SaveArchiveWorkspace(ctx, batch, ArchiveWorkspaceSave{Version: 1, Draft: r.Draft}); !errors.Is(err, ErrArchiveWorkspaceConflict) {
		t.Fatalf("版本冲突: %v", err)
	}
	r, err = svc.SubmitArchiveWorkspace(ctx, batch, ArchiveWorkspaceSubmit{Version: r.Version, FileIDs: []uuid.UUID{ids[0], ids[0]}})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Jobs) != 1 || r.Jobs[0].Snapshot.Title != "快照标题" {
		t.Fatalf("重复选择/持久快照: %+v", r.Jobs)
	}
	if err = svc.DeleteBatch(ctx, batch); err == nil {
		t.Fatal("存在活动任务时仍能删除批次")
	}
	bad := clone(r.Draft)
	f := bad.Files[ids[0].String()]
	f.Excluded = true
	bad.Files[ids[0].String()] = f
	if _, err = svc.SaveArchiveWorkspace(ctx, batch, ArchiveWorkspaceSave{Version: r.Version, Draft: bad}); err == nil {
		t.Fatal("活动文件可被排除")
	}
	// 同批次其他文件与父组默认值仍可保存；快照不随之变化。
	changed := clone(r.Draft)
	title = "后续默认标题"
	changed.Groups[0].Config.Title = &title
	f = changed.Files[ids[1].String()]
	f.Name = "新的虚拟名称.mp4"
	changed.Files[ids[1].String()] = f
	r, err = svc.SaveArchiveWorkspace(ctx, batch, ArchiveWorkspaceSave{Version: r.Version, Draft: changed})
	if err != nil {
		t.Fatal(err)
	}
	if r.Effective[ids[0].String()].Title != "快照标题" || r.Effective[ids[1].String()].Title != "后续默认标题" {
		t.Fatal("快照与后续编辑混淆")
	}
	// Redis 唤醒丢失后，重启协调器会再次发现持久任务。
	for len(queue.wake) > 0 {
		<-queue.wake
	}
	reconcileCtx, stop := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { svc.RunWorkspaceReconciler(reconcileCtx); close(done) }()
	select {
	case got := <-queue.wake:
		if got != batch {
			t.Fatal(got)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	stop()
	<-done
	queue.Lock()
	queue.failTranscode = true
	queue.Unlock()
	if err = svc.ProcessArchiveWorkspace(ctx, batch); err != nil {
		t.Fatal(err)
	}
	r, err = svc.GetArchiveWorkspace(ctx, batch)
	if err != nil {
		t.Fatal(err)
	}
	if r.Jobs[0].State != "failed" || r.Files[0].Status == "processing" {
		t.Fatalf("部分失败未落库: %+v", r.Jobs)
	}
	if !strings.Contains(r.Jobs[0].LastError, "模拟转码入队失败") {
		t.Fatalf("非预期处理失败: %+v", r.Jobs)
	}
	var videoID uuid.UUID
	var original, titleBefore string
	if err = pool.QueryRow(ctx, `SELECT linked_video_id FROM archive_import_files WHERE id=$1`, ids[0]).Scan(&videoID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT original_path,title FROM videos WHERE id=$1`, videoID).Scan(&original, &titleBefore); err != nil {
		t.Fatal(err)
	}
	if titleBefore != "快照标题" {
		t.Fatalf("未使用快照: %s", titleBefore)
	}
	queue.Lock()
	queue.failTranscode = false
	queue.failWake = false
	queue.Unlock()
	r, err = svc.SubmitArchiveWorkspace(ctx, batch, ArchiveWorkspaceSubmit{Version: r.Version, FileIDs: []uuid.UUID{ids[0]}})
	if err != nil {
		t.Fatal(err)
	}
	if err = svc.ProcessArchiveWorkspace(ctx, batch); err != nil {
		t.Fatal(err)
	}
	var count int
	var pathAfter, titleAfter string
	if err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM videos`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT original_path,title FROM videos WHERE id=$1`, videoID).Scan(&pathAfter, &titleAfter); err != nil {
		t.Fatal(err)
	}
	if count != 1 || pathAfter != original || titleAfter != titleBefore {
		t.Fatalf("重试覆盖源/标题或产生重复: %d %s %s", count, pathAfter, titleAfter)
	}
	if _, err = os.Stat(original); err != nil {
		t.Fatalf("源路径已失效: %v", err)
	}
	// 模拟源已成功但作业完成事务未提交的进程退出，重启后仍补发转码。
	exec(`UPDATE archive_import_workspace_jobs SET state='running' WHERE id=(SELECT id FROM archive_import_workspace_jobs WHERE file_id=$1 ORDER BY created_at DESC,id DESC LIMIT 1)`, ids[0])
	exec(`UPDATE archive_import_batches SET processed_entries=0 WHERE id=$1`, batch)
	queue.Lock()
	before := queue.transcodes
	queue.Unlock()
	if err = svc.ProcessArchiveWorkspace(ctx, batch); err != nil {
		t.Fatal(err)
	}
	queue.Lock()
	after := queue.transcodes
	queue.Unlock()
	if after != before+1 {
		t.Fatal("未恢复记录已创建但后续任务未完成的作业")
	}
	if err = pool.QueryRow(ctx, `SELECT processed_entries FROM archive_import_batches WHERE id=$1`, batch).Scan(&count); err != nil || count != 1 {
		t.Fatalf("恢复后批次统计未刷新: %d %v", count, err)
	}
	// 重复视频追加两个合集关系，并保持已有元数据和物理源。
	vc1, vc2, ic1, ic2 := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	exec(`INSERT INTO collections(id,name,normalized_name) VALUES($1,'视频一','视频一'),($2,'视频二','视频二')`, vc1, vc2)
	exec(`INSERT INTO collections_images(id,name,normalized_name) VALUES($1,'图片一','图片一'),($2,'图片二','图片二')`, ic1, ic2)
	if err = repo.AddVideoCollectionsByIDsAnyType(ctx, videoID, []uuid.UUID{vc1}); err != nil {
		t.Fatal(err)
	}
	if err = repo.AddVideoImageCollections(ctx, videoID, []uuid.UUID{ic1}, "direct"); err != nil {
		t.Fatal(err)
	}
	dupBatch, dupIDs := fixture(true)
	dup, err := svc.GetArchiveWorkspace(ctx, dupBatch)
	if err != nil {
		t.Fatal(err)
	}
	vcs, ics := []uuid.UUID{vc2}, []uuid.UUID{ic2}
	dup.Draft.Groups[0].Config.VideoCollectionIDs = &vcs
	dup.Draft.Groups[0].Config.ImageCollectionIDs = &ics
	dup, err = svc.SaveArchiveWorkspace(ctx, dupBatch, ArchiveWorkspaceSave{Version: dup.Version, Draft: dup.Draft})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.SubmitArchiveWorkspace(ctx, dupBatch, ArchiveWorkspaceSubmit{Version: dup.Version, FileIDs: dupIDs[:1]}); err != nil {
		t.Fatal(err)
	}
	if err = svc.ProcessArchiveWorkspace(ctx, dupBatch); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"video_collections", "video_image_collections"} {
		if err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+table+" WHERE video_id=$1", videoID).Scan(&count); err != nil || count != 2 {
			t.Fatalf("%s 未追加保留关系: %d %v", table, count, err)
		}
	}
	if err = pool.QueryRow(ctx, `SELECT original_path,title FROM videos WHERE id=$1`, videoID).Scan(&pathAfter, &titleAfter); err != nil {
		t.Fatal(err)
	}
	if pathAfter != original || titleAfter != titleBefore {
		t.Fatal("重复视频覆盖原内容")
	}
	// 图片加入全部图片合集，重复入库追加关系并保留旧标题；不生成逐图片视频关联。
	gifBytes, err := base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7")
	if err != nil {
		t.Fatal(err)
	}
	imageBatch, imageIDs := fixture(true)
	imagePath := filepath.Join(root, "测试图片.gif")
	if err = os.WriteFile(imagePath, gifBytes, 0600); err != nil {
		t.Fatal(err)
	}
	exec(`UPDATE archive_import_files SET media_kind='image',relative_path='www.98T.la@原图.gif',file_path=$2,file_size=$3 WHERE id=$1`, imageIDs[0], imagePath, len(gifBytes))
	imageDraft, err := svc.GetArchiveWorkspace(ctx, imageBatch)
	if err != nil {
		t.Fatal(err)
	}
	imageCollections := []uuid.UUID{ic1, ic2}
	imageDraft.Draft.Groups[0].Config.ImageCollectionIDs = &imageCollections
	imageDraft, err = svc.SaveArchiveWorkspace(ctx, imageBatch, ArchiveWorkspaceSave{Version: imageDraft.Version, Draft: imageDraft.Draft})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.SubmitArchiveWorkspace(ctx, imageBatch, ArchiveWorkspaceSubmit{Version: imageDraft.Version, FileIDs: imageIDs[:1]}); err != nil {
		t.Fatal(err)
	}
	if err = svc.ProcessArchiveWorkspace(ctx, imageBatch); err != nil {
		t.Fatal(err)
	}
	var imageID uuid.UUID
	if err = pool.QueryRow(ctx, `SELECT linked_image_id FROM archive_import_files WHERE id=$1`, imageIDs[0]).Scan(&imageID); err != nil || imageID == uuid.Nil {
		imageResult, _ := svc.GetArchiveWorkspace(ctx, imageBatch)
		t.Fatalf("图片未入库: %v %+v", err, imageResult.Jobs)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM image_collections WHERE image_id=$1`, imageID).Scan(&count); err != nil || count != 2 {
		t.Fatalf("图片遗漏多合集: %d %v", count, err)
	}
	if err = pool.QueryRow(ctx, `SELECT title FROM images WHERE id=$1`, imageID).Scan(&titleAfter); err != nil || titleAfter != "原图" {
		t.Fatalf("图片默认标题: %s %v", titleAfter, err)
	}
	// 同批次第二个源条目复用同图，已入库文件仍不能排除/换组。
	exec(`UPDATE archive_import_files SET media_kind='image',relative_path='另一张.gif',file_path=$2,file_size=$3 WHERE id=$1`, imageIDs[1], imagePath, len(gifBytes))
	imageDraft, err = svc.GetArchiveWorkspace(ctx, imageBatch)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.SubmitArchiveWorkspace(ctx, imageBatch, ArchiveWorkspaceSubmit{Version: imageDraft.Version, FileIDs: imageIDs[1:]}); err != nil {
		t.Fatal(err)
	}
	if err = svc.ProcessArchiveWorkspace(ctx, imageBatch); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM images`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("图片重复创建: %d %v", count, err)
	}
	if err = pool.QueryRow(ctx, `SELECT title FROM images WHERE id=$1`, imageID).Scan(&titleAfter); err != nil || titleAfter != "原图" {
		t.Fatal("重复图片覆盖标题")
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM video_images`).Scan(&count); err != nil || count != 0 {
		t.Fatal("工作台自动创建了逐图片关联")
	}
	imageDraft, err = svc.GetArchiveWorkspace(ctx, imageBatch)
	if err != nil {
		t.Fatal(err)
	}
	frozen := clone(imageDraft.Draft)
	frozenFile := frozen.Files[imageIDs[0].String()]
	frozenFile.Excluded = true
	frozen.Files[imageIDs[0].String()] = frozenFile
	if _, err = svc.SaveArchiveWorkspace(ctx, imageBatch, ArchiveWorkspaceSave{Version: imageDraft.Version, Draft: frozen}); err == nil {
		t.Fatal("成功图片可被排除")
	}
	if err = svc.DeleteBatch(ctx, imageBatch); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM images WHERE id=$1`, imageID).Scan(&count); err != nil || count != 1 {
		t.Fatal("删除批次误删已入库媒体")
	}
	// 删除解压目录失败也必须回到可重试状态，避免永久停在 extracting。
	retryBatch, _ := fixture(true)
	obstruction := filepath.Join(root, "不是目录")
	if err = os.WriteFile(obstruction, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	exec(`UPDATE archive_import_batches SET extracted_dir=$2 WHERE id=$1`, retryBatch, filepath.Join(obstruction, "子目录"))
	if _, err = svc.RetryExtract(ctx, retryBatch, "", "auto"); err == nil {
		t.Fatal("预期目录清理失败")
	}
	var retryStatus string
	if err = pool.QueryRow(ctx, `SELECT status FROM archive_import_batches WHERE id=$1`, retryBatch).Scan(&retryStatus); err != nil || retryStatus != "failed" {
		t.Fatalf("清理失败后状态: %s %v", retryStatus, err)
	}
	// 旧批次路径式组名、图片自定义标题在首次保存和后续继承时保留。
	oldBatch, oldIDs := fixture(false)
	gid := uuid.New()
	exec(`INSERT INTO archive_import_groups(id,batch_id,name,media_kind,title) VALUES($1,$2,'父/子','image','旧组标题')`, gid, oldBatch)
	exec(`UPDATE archive_import_files SET group_id=$2,media_kind='image',title='旧图片标题' WHERE id=$1`, oldIDs[0], gid)
	release, err := svc.AcquireLegacyWorkspaceGuard(ctx, "batch", oldBatch)
	if err != nil {
		t.Fatal(err)
	}
	release()
	old, err := svc.GetArchiveWorkspace(ctx, oldBatch)
	if err != nil {
		t.Fatal(err)
	}
	if old.Effective[oldIDs[0].String()].Title != "旧图片标题" || old.Draft.Files[oldIDs[0].String()].GroupID != gid.String() {
		t.Fatal("旧图片标题/归属丢失")
	}
	old, err = svc.SaveArchiveWorkspace(ctx, oldBatch, ArchiveWorkspaceSave{Version: old.Version, Draft: old.Draft})
	if err != nil {
		t.Fatalf("旧嵌套组首次保存: %v", err)
	}
	old, err = svc.SaveArchiveWorkspace(ctx, oldBatch, ArchiveWorkspaceSave{Version: old.Version, Draft: clone(old.Draft)})
	if err != nil {
		t.Fatalf("旧数据往返保存: %v", err)
	}
	if old.Effective[oldIDs[0].String()].Title != "旧图片标题" {
		t.Fatal("往返后标题丢失")
	}
}
