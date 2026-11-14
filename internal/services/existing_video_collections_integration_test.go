package services

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
)

// 真实 PostgreSQL 覆盖「已存在视频合集关系追加」的幂等与并发语义：
// 纯内存单测无法证明 ON CONFLICT 的实际行为。
func TestAttachExistingVideoCollectionsPostgres(t *testing.T) {
	pool := workspaceIntegrationDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	repo := repository.NewVideoRepository(pool)

	userID := uuid.New()
	suffix := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO users(id,username,email,password_hash) VALUES($1,$2,$3,'unused')`, userID, "attach-"+suffix, "attach-"+suffix+"@localhost"); err != nil {
		t.Fatal(err)
	}
	createCollection := func(name string) uuid.UUID {
		t.Helper()
		id := uuid.New()
		if _, err := pool.Exec(ctx, `INSERT INTO collections(id,name,normalized_name) VALUES($1,$2,$2)`, id, name); err != nil {
			t.Fatal(err)
		}
		return id
	}
	first := createCollection("合集A-" + suffix)
	second := createCollection("合集B-" + suffix)

	videoID := uuid.New()
	if err := repo.CreateVideo(ctx, models.Video{ID: videoID, UserID: &userID, Title: "已存在视频", Type: "short", Status: "ready"}); err != nil {
		t.Fatal(err)
	}
	video, err := repo.GetVideoByID(ctx, videoID)
	if err != nil {
		t.Fatal(err)
	}

	attachment, err := AttachExistingVideoCollections(ctx, repo, video, []uuid.UUID{first, second, first}, nil)
	if err != nil {
		t.Fatalf("首次追加失败: %v", err)
	}
	if attachment.AddedCollections != 2 {
		t.Fatalf("首次追加 = %d, want 2", attachment.AddedCollections)
	}

	attachment, err = AttachExistingVideoCollections(ctx, repo, video, []uuid.UUID{first, second}, nil)
	if err != nil {
		t.Fatalf("重复追加失败: %v", err)
	}
	if attachment.AddedCollections != 0 {
		t.Fatalf("重复追加 = %d, want 0", attachment.AddedCollections)
	}

	// 并发追加同一批关系时，总新增数必须等于关系数，不能重复计数。
	third := createCollection("合集C-" + suffix)
	var wg sync.WaitGroup
	var mu sync.Mutex
	total := 0
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, attachErr := AttachExistingVideoCollections(ctx, repo, video, []uuid.UUID{third}, nil)
			if attachErr != nil {
				t.Errorf("并发追加失败: %v", attachErr)
				return
			}
			mu.Lock()
			total += got.AddedCollections
			mu.Unlock()
		}()
	}
	wg.Wait()
	if total != 1 {
		t.Fatalf("并发新增合计 = %d, want 1", total)
	}

	if _, err := pool.Exec(ctx, `UPDATE videos SET status='pending_delete' WHERE id=$1`, videoID); err != nil {
		t.Fatal(err)
	}
	pending, err := repo.GetVideoByID(ctx, videoID)
	if err != nil {
		t.Fatal(err)
	}
	fourth := createCollection("合集D-" + suffix)
	attachment, err = AttachExistingVideoCollections(ctx, repo, pending, []uuid.UUID{fourth}, nil)
	if err != nil {
		t.Fatalf("待删除视频追加返回错误: %v", err)
	}
	if !attachment.VideoPendingDelete || attachment.AddedCollections != 0 {
		t.Fatalf("待删除视频不应写入关系: %+v", attachment)
	}
	var linked int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM video_collections WHERE video_id=$1 AND collection_id=$2`, videoID, fourth).Scan(&linked); err != nil {
		t.Fatal(err)
	}
	if linked != 0 {
		t.Fatalf("待删除视频被写入 %d 条关系", linked)
	}
}

// TestAlignArchiveFileRelationsPostgres 覆盖压缩包 existing 文件的关系对齐：
// 追加、幂等、以及待删除视频不写入。
func TestAlignArchiveFileRelationsPostgres(t *testing.T) {
	pool := workspaceIntegrationDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	repo := repository.NewVideoRepository(pool)
	root := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewArchiveImportService(pool, NewUploadService(repo, root, root, logger), NewImageService(repo, root, root, logger), repo, &workspaceIntegrationQueue{}, root, root, logger)

	suffix := uuid.NewString()
	userID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users(id,username,email,password_hash) VALUES($1,$2,$3,'unused')`, userID, "align-"+suffix, "align-"+suffix+"@localhost"); err != nil {
		t.Fatal(err)
	}
	collectionID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO collections(id,name,normalized_name) VALUES($1,$2,$2)`, collectionID, "对齐合集-"+suffix); err != nil {
		t.Fatal(err)
	}
	videoID := uuid.New()
	if err := repo.CreateVideo(ctx, models.Video{ID: videoID, UserID: &userID, Title: "已存在视频", Type: "short", Status: "ready"}); err != nil {
		t.Fatal(err)
	}

	batchID := uuid.New()
	dir := filepath.Join(root, batchID.String())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO archive_import_batches(id,title,original_filename,archive_format,original_path,extracted_dir,status) VALUES($1,'批次','素材.zip','zip',$2,$3,'ready')`, batchID, filepath.Join(dir, "素材.zip"), dir); err != nil {
		t.Fatal(err)
	}
	fileID := uuid.New()
	collectionIDsJSON, err := json.Marshal([]uuid.UUID{collectionID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO archive_import_files(id,batch_id,relative_path,file_path,entry_type,media_kind,file_size,title,status,linked_video_id,video_collection_ids) VALUES($1,$2,'a.mp4',$3,'file','video',10,'a.mp4','existing',$4,$5)`, fileID, batchID, filepath.Join(dir, "a.mp4"), videoID, collectionIDsJSON); err != nil {
		t.Fatal(err)
	}

	file, err := svc.getArchiveFile(ctx, fileID)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := svc.GetBatch(ctx, batchID)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.alignArchiveFileRelations(ctx, file, batch); err != nil {
		t.Fatalf("首次关系对齐失败: %v", err)
	}
	if file, err = svc.getArchiveFile(ctx, fileID); err != nil {
		t.Fatal(err)
	}
	if file.Reason != archiveExistingVideoReason(1) {
		t.Fatalf("首次对齐原因 = %q, want %q", file.Reason, archiveExistingVideoReason(1))
	}

	if err := svc.alignArchiveFileRelations(ctx, file, batch); err != nil {
		t.Fatalf("重复关系对齐失败: %v", err)
	}
	if file, err = svc.getArchiveFile(ctx, fileID); err != nil {
		t.Fatal(err)
	}
	if file.Reason != archiveExistingVideoReason(0) {
		t.Fatalf("重复对齐原因 = %q, want %q", file.Reason, archiveExistingVideoReason(0))
	}
	var linked int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM video_collections WHERE video_id=$1 AND collection_id=$2`, videoID, collectionID).Scan(&linked); err != nil {
		t.Fatal(err)
	}
	if linked != 1 {
		t.Fatalf("video_collections 行数 = %d, want 1（幂等）", linked)
	}

	if _, err := pool.Exec(ctx, `UPDATE videos SET status='pending_delete' WHERE id=$1`, videoID); err != nil {
		t.Fatal(err)
	}
	secondCollection := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO collections(id,name,normalized_name) VALUES($1,$2,$2)`, secondCollection, "待删除合集-"+suffix); err != nil {
		t.Fatal(err)
	}
	updatedJSON, err := json.Marshal([]uuid.UUID{secondCollection})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE archive_import_files SET video_collection_ids=$2 WHERE id=$1`, fileID, updatedJSON); err != nil {
		t.Fatal(err)
	}
	if file, err = svc.getArchiveFile(ctx, fileID); err != nil {
		t.Fatal(err)
	}
	if err := svc.alignArchiveFileRelations(ctx, file, batch); err != nil {
		t.Fatalf("待删除视频对齐返回错误: %v", err)
	}
	if file, err = svc.getArchiveFile(ctx, fileID); err != nil {
		t.Fatal(err)
	}
	if file.Reason != archiveExistingVideoPendingDeleteReason {
		t.Fatalf("待删除原因 = %q, want %q", file.Reason, archiveExistingVideoPendingDeleteReason)
	}
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM video_collections WHERE video_id=$1 AND collection_id=$2`, videoID, secondCollection).Scan(&linked); err != nil {
		t.Fatal(err)
	}
	if linked != 0 {
		t.Fatalf("待删除视频被写入 %d 条关系", linked)
	}
}
