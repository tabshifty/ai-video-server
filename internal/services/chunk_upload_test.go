package services

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestChunkUploadCompleteFileSurvivesAbort(t *testing.T) {
	t.Parallel()

	svc := NewChunkUploadService(t.TempDir())
	userID := uuid.New()

	session, err := svc.Init(context.Background(), userID, "demo.mp4", 6, 3, 2, "hash", "short", "t", "d", nil, nil, nil, "", "", nil)
	if err != nil {
		t.Fatalf("init session: %v", err)
	}
	// 兼容旧版仅创建会话根目录、尚未接收首个分片的会话。
	if err := os.Remove(filepath.Join(svc.sessionDir(session.ID), "chunks")); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.SaveChunk(context.Background(), session.ID, 0, bytes.NewReader([]byte("abc"))); err != nil {
		t.Fatalf("save chunk 0: %v", err)
	}
	if _, err := svc.SaveChunk(context.Background(), session.ID, 1, bytes.NewReader([]byte("def"))); err != nil {
		t.Fatalf("save chunk 1: %v", err)
	}

	_, mergedPath, err := svc.Complete(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("complete session: %v", err)
	}

	if err := svc.Abort(session.ID); err != nil {
		t.Fatalf("abort session: %v", err)
	}

	raw, err := os.ReadFile(mergedPath)
	if err != nil {
		t.Fatalf("read merged file after abort: %v", err)
	}
	if string(raw) != "abcdef" {
		t.Fatalf("merged content = %q, want %q", string(raw), "abcdef")
	}
}

func TestChunkUploadCollectionReferences(t *testing.T) {
	t.Parallel()
	svc := NewChunkUploadService(t.TempDir())
	ctx := context.Background()
	vc, ic := uuid.New(), uuid.New()
	session, err := svc.Init(ctx, uuid.New(), "demo.mp4", 6, 3, 2, "hash", "short", "t", "d", nil, nil, nil, ic.String(), "", []string{vc.String()})
	if err != nil {
		t.Fatal(err)
	}
	for kind, id := range map[string]uuid.UUID{"video": vc, "image": ic} {
		refs, err := svc.CollectionReferences(ctx, kind)
		if err != nil || !refs[id] {
			t.Fatalf("%s 上传引用未保护 %v %v", kind, refs, err)
		}
	}
	// 多次更新元数据后仍可读取完整引用。
	if _, err = svc.SaveChunk(ctx, session.ID, 0, bytes.NewBufferString("abc")); err != nil {
		t.Fatal(err)
	}
	if refs, err := svc.CollectionReferences(ctx, "video"); err != nil || !refs[vc] {
		t.Fatalf("保存分片后引用丢失 %v", err)
	}
	writer := make(chan error, 1)
	go func() {
		for i := 0; i < 50; i++ {
			if _, err := svc.SaveChunk(ctx, session.ID, 0, bytes.NewBufferString("abc")); err != nil {
				writer <- err
				return
			}
		}
		writer <- nil
	}()
	for i := 0; i < 50; i++ {
		if refs, err := svc.CollectionReferences(ctx, "image"); err != nil || !refs[ic] {
			t.Fatalf("并发更新时引用读取失败 %v", err)
		}
	}
	if err := <-writer; err != nil {
		t.Fatal(err)
	}
	if err = svc.Abort(session.ID); err != nil {
		t.Fatal(err)
	}
	if refs, err := svc.CollectionReferences(ctx, "video"); err != nil || len(refs) != 0 {
		t.Fatalf("取消后保留引用 %v %v", refs, err)
	}
	// 损坏数据必须阻止清理，不能将未知引用视为不存在。
	if err = os.MkdirAll(svc.sessionDir(session.ID), 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(svc.metaPath(session.ID), []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.CollectionReferences(ctx, "image"); err == nil {
		t.Fatal("忽略损坏的上传元数据")
	}
}

type pausedChunkReader struct {
	started chan struct{}
	proceed chan struct{}
	read    bool
}

func (r *pausedChunkReader) Read(p []byte) (int, error) {
	if r.read {
		return 0, io.EOF
	}
	r.read = true
	close(r.started)
	<-r.proceed
	return copy(p, []byte("abc")), nil
}

func TestChunkUploadAbortCannotReviveCollectionReferences(t *testing.T) {
	t.Parallel()
	svc := NewChunkUploadService(t.TempDir())
	id := uuid.New()
	session, err := svc.Init(context.Background(), uuid.New(), "demo.mp4", 3, 3, 1, "hash", "short", "t", "d", nil, nil, nil, "", "", []string{id.String()})
	if err != nil {
		t.Fatal(err)
	}
	reader := &pausedChunkReader{started: make(chan struct{}), proceed: make(chan struct{})}
	done := make(chan error, 1)
	go func() { _, err := svc.SaveChunk(context.Background(), session.ID, 0, reader); done <- err }()
	<-reader.started
	if err = svc.Abort(session.ID); err != nil {
		t.Fatal(err)
	}
	if refs, err := svc.CollectionReferences(context.Background(), "video"); err != nil || len(refs) != 0 {
		t.Fatalf("取消后仍有引用 %v %v", refs, err)
	}
	close(reader.proceed)
	if err = <-done; err == nil {
		t.Fatal("取消后的分片更新仍成功")
	}
	if _, err = os.Stat(svc.sessionDir(session.ID)); !os.IsNotExist(err) {
		t.Fatalf("取消后会话被重建 %v", err)
	}
	if refs, err := svc.CollectionReferences(context.Background(), "video"); err != nil || len(refs) != 0 {
		t.Fatalf("引用复活 %v %v", refs, err)
	}
}
