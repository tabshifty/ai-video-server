package services

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"video-server/internal/repository"
	"video-server/pkg/ffmpeg"
)

func TestRotationOwnedPath(t *testing.T) {
	root := t.TempDir()
	id := uuid.New()
	base := filepath.Join(root, "videos", id.String())
	if err := os.MkdirAll(base, 0755); err != nil {
		t.Fatal(err)
	}
	svc := NewVideoRotationService(nil, root)
	if err := svc.ownedPath(id, filepath.Join(base, "new.mp4")); err != nil {
		t.Fatal(err)
	}
	if err := svc.ownedPath(id, filepath.Join(root, "outside.mp4")); err == nil {
		t.Fatal("accepted out-of-scope file")
	}
	other := t.TempDir()
	linkID := uuid.New()
	if err := os.Symlink(other, filepath.Join(root, "videos", linkID.String())); err != nil {
		t.Fatal(err)
	}
	if err := svc.ownedPath(linkID, filepath.Join(root, "videos", linkID.String(), "x.mp4")); err == nil {
		t.Fatal("accepted linked video directory")
	}
	if err := checkRotationSpace(base, 1<<62); err == nil {
		t.Fatal("accepted insufficient space")
	}
}

func TestVideoRotationIntegration(t *testing.T) {
	pool := workspaceIntegrationDB(t)
	ctx := context.Background()
	repo := repository.NewVideoRepository(pool)
	root := t.TempDir()
	svc := NewVideoRotationService(repo, root)
	id, user := uuid.New(), uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users(id,username,email,password_hash,role) VALUES($1,'rotation-test','rotation@test.local','unused','admin')`, user); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "videos", id.String())
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "play.mp4")
	original := filepath.Join(dir, "source-original")
	thumb := filepath.Join(dir, "thumb.jpg")
	if out, err := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "testsrc2=size=160x96:rate=10", "-f", "lavfi", "-i", "sine=frequency=440", "-t", "1", "-c:v", "libx264", "-c:a", "aac", "-shortest", src).CombinedOutput(); err != nil {
		t.Fatalf("sample: %v %s", err, out)
	}
	bytes, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(original, bytes, 0600); err != nil {
		t.Fatal(err)
	}
	if err = ffmpeg.Thumbnail(ctx, src, thumb); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO videos(id,user_id,type,status,original_path,transcoded_path,thumbnail_path,metadata) VALUES($1,$2,'short','ready',$3,$4,$5,jsonb_build_object('playback_path',$4::text))`, id, user, original, src, thumb); err != nil {
		t.Fatal(err)
	}

	// Ordinary detail saves must still work before any rotation has occurred.
	if err = repo.AdminUpdateVideo(ctx, id, "普通视频", "", thumb, nil, map[string]any{"playback_path": src}, "", nil, nil, false, nil, false, nil, false, ""); err != nil {
		t.Fatal("ordinary detail save", err)
	}
	plan, err := svc.Preview(ctx, id, "auto")
	if err != nil {
		t.Fatal(err)
	}
	if plan.Mode != "lossless" {
		t.Fatalf("plan: %+v", plan)
	}
	request := uuid.New()
	job, err := svc.Submit(ctx, id, user, request, 90, plan.Mode, plan.Version, false)
	if err != nil {
		t.Fatal(err)
	}
	duplicate, err := svc.Submit(ctx, id, user, request, 90, plan.Mode, plan.Version, false)
	if err != nil || duplicate.ID != job.ID {
		t.Fatalf("replay %v", err)
	}
	if _, err = svc.Submit(ctx, id, user, uuid.New(), 90, plan.Mode, plan.Version, false); err == nil {
		t.Fatal("concurrent operation accepted")
	}
	if _, err = pool.Exec(ctx, `UPDATE videos SET status='uploaded' WHERE id=$1`, id); err == nil {
		t.Fatal("active job allowed retranscode")
	}
	if err = repo.DeleteVideoByID(ctx, id); err == nil {
		t.Fatal("active job allowed delete")
	}
	v, err := repo.GetVideoByID(ctx, id)
	if err != nil || v.Status != "ready" || v.TranscodedPath != src {
		t.Fatal("old playback not retained")
	}
	refs, err := repo.CollectReferencedFilePaths(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := refs[job.OutputPath]; !ok {
		t.Fatal("output not protected from orphan scan")
	}
	if err = svc.Process(ctx, job.ID); err != nil {
		t.Fatal(err)
	}
	saved, err := repo.RotationByID(ctx, job.ID)
	if err != nil || saved.Status != "succeeded" {
		t.Fatalf("job: %+v %v", saved, err)
	}
	for _, path := range []string{src, original, thumb} {
		if _, err = os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("old file retained %s %v", path, err)
		}
	}
	v, err = repo.GetVideoByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if v.Width != 96 || v.Height != 160 || v.Status != "ready" || v.OriginalPath != "" {
		t.Fatalf("video: %+v", v)
	}
	var metadata map[string]any
	if err = json.Unmarshal(v.Metadata, &metadata); err != nil {
		t.Fatal(err)
	}
	if metadata["playback_path"] != job.OutputPath {
		t.Fatal("stale playback metadata")
	}

	if err = repo.AdminUpdateVideo(ctx, id, "过期详情", "", thumb, nil, map[string]any{}, "", nil, nil, false, nil, false, nil, false, ""); err == nil {
		t.Fatal("stale details overwrote rotated file")
	}
	if err = repo.AdminUpdateVideo(ctx, id, "新详情", "", v.ThumbnailPath, nil, metadata, "", nil, nil, false, nil, false, nil, false, ""); err != nil {
		t.Fatal("current detail save", err)
	}
	if err = svc.Process(ctx, job.ID); err != nil {
		t.Fatal(err)
	}
	replay, err := svc.Submit(ctx, id, user, request, 90, plan.Mode, plan.Version, false)
	if err != nil || replay.ID != request {
		t.Fatal("post-cleanup replay failed", err)
	}
	// Mark the current auto-looking thumbnail as manually selected; it must survive.
	if err = repo.AdminUpdateVideoThumbnail(ctx, id, v.ThumbnailPath); err != nil {
		t.Fatal(err)
	}
	manual := v.ThumbnailPath
	plan, err = svc.Preview(ctx, id, "auto")
	if err != nil {
		t.Fatal(err)
	}
	next, err := svc.Submit(ctx, id, user, uuid.New(), 90, plan.Mode, plan.Version, false)
	if err != nil {
		t.Fatal(err)
	}
	if err = svc.Process(ctx, next.ID); err != nil {
		t.Fatal(err)
	}
	v, err = repo.GetVideoByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if v.Width != 160 || v.Height != 96 || v.ThumbnailPath != manual {
		t.Fatal("relative rotation or manual thumbnail failed")
	}
	if _, err = os.Stat(manual); err != nil {
		t.Fatal("manual thumbnail removed")
	}
	// A file replaced outside the application must never be published over.
	plan, err = svc.Preview(ctx, id, "auto")
	if err != nil {
		t.Fatal(err)
	}
	failed, err := svc.Submit(ctx, id, user, uuid.New(), 180, plan.Mode, plan.Version, false)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(v.TranscodedPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write([]byte("changed")); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if err = svc.Process(ctx, failed.ID); err == nil {
		t.Fatal("changed input accepted")
	}
	failed, err = repo.RotationByID(ctx, failed.ID)
	if err != nil || failed.Status != "failed" {
		t.Fatal("failure not persisted")
	}
	if _, err = os.Stat(v.TranscodedPath); err != nil {
		t.Fatal("failure removed old video")
	}

	// Cross-session mutations cannot overlap even when they use another repository instance.
	unlock, err := repo.LockVideoFiles(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if release, e := repository.NewVideoRepository(pool).LockVideoFiles(ctx, id); e == nil {
		release()
		t.Fatal("cross-session file lock missing")
	}
	unlock()
	// Simulate a crash immediately after publication and a shared old-file reference
	// that temporarily prevents cleanup. Recovery must only clean, never rotate again.
	plan, err = svc.Preview(ctx, id, "auto")
	if err != nil {
		t.Fatal(err)
	}
	recovery, err := svc.Submit(ctx, id, user, uuid.New(), 90, plan.Mode, plan.Version, false)
	if err != nil {
		t.Fatal(err)
	}
	p, err := ffmpeg.ProbeRotation(ctx, recovery.InputPath)
	if err != nil {
		t.Fatal(err)
	}
	if err = ffmpeg.RotateVideo(ctx, recovery.InputPath, recovery.OutputPath, 90, "lossless", p, nil); err != nil {
		t.Fatal(err)
	}
	if err = repo.SetRotationState(ctx, recovery.ID, "running", 95, ""); err != nil {
		t.Fatal(err)
	}
	out, err := ffmpeg.ProbeRotation(ctx, recovery.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	w, h := out.DisplaySize()
	size, err := RequiredRegularFileSize(recovery.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.PublishRotation(ctx, recovery, size, w, h, manual, map[string]any{"playback_path": recovery.OutputPath}); err != nil {
		t.Fatal(err)
	}
	referencedID := uuid.New()
	if _, err = pool.Exec(ctx, `INSERT INTO videos(id,type,status,transcoded_path) VALUES($1,'short','ready',$2)`, referencedID, recovery.InputPath); err != nil {
		t.Fatal(err)
	}
	if err = svc.Process(ctx, recovery.ID); err == nil {
		t.Fatal("referenced old file removed")
	}
	cleanupJob, err := repo.RotationByID(ctx, recovery.ID)
	if err != nil || cleanupJob.Status != "cleanup" || cleanupJob.ErrorMessage == "" {
		t.Fatal("cleanup failure not durable")
	}
	if _, err = pool.Exec(ctx, `DELETE FROM videos WHERE id=$1`, referencedID); err != nil {
		t.Fatal(err)
	}
	if err = svc.Process(ctx, recovery.ID); err != nil {
		t.Fatal(err)
	}
	after, err := ffmpeg.ProbeRotation(ctx, recovery.OutputPath)
	if err != nil || after.Rotation != out.Rotation {
		t.Fatal("recovery applied angle twice", err)
	}
	if _, err = os.Stat(recovery.InputPath); !os.IsNotExist(err) {
		t.Fatal("recovery did not remove old input")
	}
	// Migration can be reapplied and rolled back in this isolated fixture.
	for _, name := range []string{"0049_video_rotation.up.sql", "0049_video_rotation.down.sql", "0049_video_rotation.up.sql"} {
		raw, e := os.ReadFile(filepath.Join("../../migrations", name))
		if e != nil {
			t.Fatal(e)
		}
		if _, e = pool.Exec(ctx, string(raw)); e != nil {
			t.Fatalf("migration %s: %v", name, e)
		}
	}
	// Re-encoding requires consent even if the caller bypasses the UI.
	if _, err = svc.Submit(ctx, id, user, uuid.New(), 90, "reencode", "invalid", false); err == nil {
		t.Fatal("missing consent accepted")
	}
}

func TestRotationThumbnailClassification(t *testing.T) {
	for _, test := range []struct {
		path string
		want bool
	}{
		{"", true}, {"/managed/thumb.jpg", true}, {"/managed/rotation-" + uuid.NewString() + ".jpg", true},
		{"https://example.test/thumb.jpg", false}, {"/managed/poster.jpg", false}, {"/managed/rotation-family.jpg", false},
	} {
		if got := automaticRotationThumbnail(test.path); got != test.want {
			t.Errorf("%s: %v", test.path, got)
		}
	}
}
