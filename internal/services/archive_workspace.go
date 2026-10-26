package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"video-server/internal/models"
)

// ErrArchiveWorkspaceConflict means a newer saved draft must be loaded before editing.
var ErrArchiveWorkspaceConflict = errors.New("工作区已在其他窗口更新，请保留当前编辑并重新加载后再操作")

// ArchiveWorkspaceResponse combines saved editing state with current processing progress.
type ArchiveWorkspaceResponse struct {
	Batch     models.ArchiveImportBatch          `json:"batch"`
	Version   int64                              `json:"version"`
	Draft     ArchiveWorkspaceDraft              `json:"draft"`
	Files     []models.ArchiveImportFileListItem `json:"files"`
	Effective map[string]ArchiveWorkspaceValues  `json:"effective"`
	Jobs      []ArchiveWorkspaceJob              `json:"jobs"`
}

// ArchiveWorkspaceJob is one persistent import attempt and its immutable resolved values.
type ArchiveWorkspaceJob struct {
	ID        uuid.UUID              `json:"id"`
	FileID    uuid.UUID              `json:"file_id"`
	Snapshot  ArchiveWorkspaceValues `json:"snapshot"`
	State     string                 `json:"state"`
	LastError string                 `json:"last_error"`
}

// ArchiveWorkspaceSave carries the draft version used for optimistic concurrency control.
type ArchiveWorkspaceSave struct {
	Version int64                 `json:"version"`
	Draft   ArchiveWorkspaceDraft `json:"draft"`
}

// ArchiveWorkspaceSubmit captures an explicit selection from a saved draft.
type ArchiveWorkspaceSubmit struct {
	Version int64       `json:"version"`
	FileIDs []uuid.UUID `json:"file_ids"`
}

type archiveWorkspaceEnqueuer interface{ EnqueueArchiveWorkspace(uuid.UUID) error }

func (s *ArchiveImportService) lockWorkspaceBatch(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,47))`, id.String()); err != nil {
		return err
	}
	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM archive_import_batches WHERE id=$1`, id).Scan(&status); err != nil {
		return err
	}
	switch status {
	case "uploaded", "extracting", "needs_password", "needs_encoding":
		return fmt.Errorf("请先完成压缩包解压")
	case "failed":
		var count int
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM archive_import_files WHERE batch_id=$1`, id).Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("请先重试解压")
		}
	}
	return nil
}

func readWorkspace(ctx context.Context, tx pgx.Tx, id uuid.UUID) (ArchiveWorkspaceDraft, int64, error) {
	var raw []byte
	var version int64
	err := tx.QueryRow(ctx, `SELECT draft,version FROM archive_import_workspaces WHERE batch_id=$1`, id).Scan(&raw, &version)
	if err != nil {
		return ArchiveWorkspaceDraft{}, 0, err
	}
	var draft ArchiveWorkspaceDraft
	err = json.Unmarshal(raw, &draft)
	return draft, version, err
}

// GetArchiveWorkspace lazily adapts old batches without recreating their groups or overrides.
func (s *ArchiveImportService) GetArchiveWorkspace(ctx context.Context, id uuid.UUID) (ArchiveWorkspaceResponse, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	defer tx.Rollback(ctx)
	if err = s.lockWorkspaceBatch(ctx, tx, id); err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	draft, version, err := readWorkspace(ctx, tx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		batch, e := s.GetBatch(ctx, id)
		if e != nil {
			return ArchiveWorkspaceResponse{}, e
		}
		files, e := s.ListFiles(ctx, id)
		if e != nil {
			return ArchiveWorkspaceResponse{}, e
		}
		groups, e := s.ListGroups(ctx, id)
		if e != nil {
			return ArchiveWorkspaceResponse{}, e
		}
		var fresh bool
		if e = tx.QueryRow(ctx, `SELECT workspace_generation FROM archive_import_batches WHERE id=$1`, id).Scan(&fresh); e != nil {
			return ArchiveWorkspaceResponse{}, e
		}
		draft = newArchiveWorkspace(batch, files, groups, fresh)
		raw, e := json.Marshal(draft)
		if e != nil {
			return ArchiveWorkspaceResponse{}, e
		}
		if _, e = tx.Exec(ctx, `INSERT INTO archive_import_workspaces(batch_id,draft) VALUES($1,$2)`, id, raw); e != nil {
			return ArchiveWorkspaceResponse{}, e
		}
		version = 1
	} else if err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	return s.workspaceResponse(ctx, id, draft, version)
}

func (s *ArchiveImportService) workspaceResponse(ctx context.Context, id uuid.UUID, draft ArchiveWorkspaceDraft, version int64) (ArchiveWorkspaceResponse, error) {
	r := ArchiveWorkspaceResponse{Draft: draft, Version: version, Effective: map[string]ArchiveWorkspaceValues{}, Jobs: []ArchiveWorkspaceJob{}}
	var err error
	r.Batch, err = s.GetBatch(ctx, id)
	if err != nil {
		return r, err
	}
	r.Files, err = s.ListFiles(ctx, id)
	if err != nil {
		return r, err
	}
	rows, err := s.db.Query(ctx, `SELECT DISTINCT ON (file_id) id,file_id,snapshot,state,last_error FROM archive_import_workspace_jobs WHERE batch_id=$1 ORDER BY file_id,created_at DESC,id DESC`, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	active := map[string]ArchiveWorkspaceValues{}
	for rows.Next() {
		var j ArchiveWorkspaceJob
		var raw []byte
		if err = rows.Scan(&j.ID, &j.FileID, &raw, &j.State, &j.LastError); err != nil {
			return r, err
		}
		if err = json.Unmarshal(raw, &j.Snapshot); err != nil {
			return r, err
		}
		r.Jobs = append(r.Jobs, j)
		if j.State == "queued" || j.State == "running" || j.State == "completed" {
			active[j.FileID.String()] = j.Snapshot
		}
	}
	if err = rows.Err(); err != nil {
		return r, err
	}
	for _, f := range r.Files {
		if f.EntryType == "directory" {
			continue
		}
		v, e := resolveWorkspaceFile(draft, f.ID.String(), f.MediaKind)
		if e != nil {
			return r, e
		}
		if saved, ok := active[f.ID.String()]; ok {
			v = saved
		} else if workspaceSuccess(f) {
			v = ArchiveWorkspaceValues{Title: f.Title, TitleMode: "uniform", Description: f.Description, Tags: f.Tags, VideoType: f.VideoType, VideoCollectionIDs: f.VideoCollectionIDs, ImageCollectionIDs: f.ImageCollectionIDs}
		}
		r.Effective[f.ID.String()] = v
	}
	return r, nil
}

func workspaceActiveJobs(ctx context.Context, tx pgx.Tx, id uuid.UUID) (map[string]bool, error) {
	rows, err := tx.Query(ctx, `SELECT file_id FROM archive_import_workspace_jobs WHERE batch_id=$1 AND state IN ('queued','running')`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	active := map[string]bool{}
	for rows.Next() {
		var fileID uuid.UUID
		if err = rows.Scan(&fileID); err != nil {
			return nil, err
		}
		active[fileID.String()] = true
	}
	return active, rows.Err()
}

// SaveArchiveWorkspace rejects stale writes and validates protected membership on the server.
func (s *ArchiveImportService) SaveArchiveWorkspace(ctx context.Context, id uuid.UUID, in ArchiveWorkspaceSave) (ArchiveWorkspaceResponse, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	defer tx.Rollback(ctx)
	if err = s.lockWorkspaceBatch(ctx, tx, id); err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	old, version, err := readWorkspace(ctx, tx, id)
	if err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	if version != in.Version {
		return ArchiveWorkspaceResponse{}, ErrArchiveWorkspaceConflict
	}
	if err = lockWorkspaceFiles(ctx, tx, id); err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	files, err := s.ListFiles(ctx, id)
	if err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	active, err := workspaceActiveJobs(ctx, tx, id)
	if err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	if err = validateWorkspace(in.Draft, old, files, active); err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	if err = s.validateWorkspaceCollections(ctx, in.Draft); err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	raw, err := json.Marshal(in.Draft)
	if err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE archive_import_workspaces SET draft=$2,version=version+1,updated_at=NOW() WHERE batch_id=$1`, id, raw); err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	return s.workspaceResponse(ctx, id, in.Draft, version+1)
}

func (s *ArchiveImportService) validateWorkspaceCollections(ctx context.Context, d ArchiveWorkspaceDraft) error {
	if s.repo == nil {
		return nil
	}
	video, image := []uuid.UUID{}, []uuid.UUID{}
	add := func(c ArchiveWorkspaceConfig) {
		if c.VideoCollectionIDs != nil {
			video = append(video, (*c.VideoCollectionIDs)...)
		}
		if c.ImageCollectionIDs != nil {
			image = append(image, (*c.ImageCollectionIDs)...)
		}
	}
	for _, g := range d.Groups {
		add(g.Config)
	}
	for _, f := range d.Files {
		add(f.Config)
	}
	if len(video) > 0 {
		if _, err := s.repo.ResolveCollectionIDs(ctx, dedupeArchiveUUIDs(video)); err != nil {
			return err
		}
	}
	if len(image) > 0 {
		if _, err := s.repo.ResolveImageCollectionIDs(ctx, dedupeArchiveUUIDs(image)); err != nil {
			return err
		}
	}
	return nil
}

// SubmitArchiveWorkspace persists immutable work before trying to wake the queue.
func (s *ArchiveImportService) SubmitArchiveWorkspace(ctx context.Context, id uuid.UUID, in ArchiveWorkspaceSubmit) (ArchiveWorkspaceResponse, error) {
	if _, ok := s.enqueuer.(archiveWorkspaceEnqueuer); !ok {
		return ArchiveWorkspaceResponse{}, fmt.Errorf("入库队列不可用")
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	defer tx.Rollback(ctx)
	if err = s.lockWorkspaceBatch(ctx, tx, id); err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	draft, version, err := readWorkspace(ctx, tx, id)
	if err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	if in.Version != version {
		return ArchiveWorkspaceResponse{}, ErrArchiveWorkspaceConflict
	}
	if err = lockWorkspaceFiles(ctx, tx, id); err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	files, err := s.ListFiles(ctx, id)
	if err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	active, err := workspaceActiveJobs(ctx, tx, id)
	if err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	byID := map[uuid.UUID]models.ArchiveImportFileListItem{}
	for _, f := range files {
		byID[f.ID] = f
	}
	ids := dedupeArchiveUUIDs(in.FileIDs)
	if len(ids) == 0 {
		return ArchiveWorkspaceResponse{}, fmt.Errorf("请先选择待入库文件")
	}
	if err = s.validateWorkspaceCollections(ctx, draft); err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	for _, fileID := range ids {
		f, ok := byID[fileID]
		if !ok {
			return ArchiveWorkspaceResponse{}, fmt.Errorf("文件不属于本批次")
		}
		node, ok := draft.Files[fileID.String()]
		if !ok || !workspaceMedia(f) {
			return ArchiveWorkspaceResponse{}, fmt.Errorf("选择包含不支持入库的文件")
		}
		if node.Excluded || workspaceSuccess(f) || active[fileID.String()] {
			continue
		}
		v, e := resolveWorkspaceFile(draft, fileID.String(), f.MediaKind)
		if e != nil {
			return ArchiveWorkspaceResponse{}, e
		}
		if v.Title == "" {
			return ArchiveWorkspaceResponse{}, fmt.Errorf("%s 的入库标题为空，请补填或切换标题模式", node.Name)
		}
		raw, e := json.Marshal(v)
		if e != nil {
			return ArchiveWorkspaceResponse{}, e
		}
		if _, err = tx.Exec(ctx, `INSERT INTO archive_import_workspace_jobs(batch_id,file_id,snapshot) VALUES($1,$2,$3)`, id, fileID, raw); err != nil {
			return ArchiveWorkspaceResponse{}, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return ArchiveWorkspaceResponse{}, err
	}
	// The database is authoritative. A failed wakeup is recovered by the worker reconciler.
	if err = s.enqueuer.(archiveWorkspaceEnqueuer).EnqueueArchiveWorkspace(id); err != nil {
		s.logger.Warn("archive workspace wakeup deferred", "batch_id", id, "error", err)
	}
	return s.workspaceResponse(ctx, id, draft, version)
}

// GenerateArchiveWorkspaceGroups adds missing directory-based groups without moving files.
func (s *ArchiveImportService) GenerateArchiveWorkspaceGroups(ctx context.Context, id uuid.UUID, version int64) (ArchiveWorkspaceResponse, error) {
	r, err := s.GetArchiveWorkspace(ctx, id)
	if err != nil {
		return r, err
	}
	if r.Version != version {
		return r, ErrArchiveWorkspaceConflict
	}
	generateWorkspaceDirectoryGroups(&r.Draft, r.Files, false)
	return s.SaveArchiveWorkspace(ctx, id, ArchiveWorkspaceSave{Version: version, Draft: r.Draft})
}

// ProcessArchiveWorkspace drains persisted jobs, using an advisory lock to prevent concurrent execution.
func (s *ArchiveImportService) ProcessArchiveWorkspace(ctx context.Context, id uuid.UUID) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		tx, err := s.db.Begin(ctx)
		if err != nil {
			return err
		}
		var job ArchiveWorkspaceJob
		var raw []byte
		err = tx.QueryRow(ctx, `SELECT id,file_id,snapshot FROM archive_import_workspace_jobs WHERE batch_id=$1 AND state IN ('queued','running') ORDER BY created_at,id LIMIT 1`, id).Scan(&job.ID, &job.FileID, &raw)
		if errors.Is(err, pgx.ErrNoRows) {
			_ = tx.Rollback(ctx)
			return nil
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err = json.Unmarshal(raw, &job.Snapshot); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		// A transaction advisory lock serializes attempts without hiding their running status.
		// If the worker is killed, PostgreSQL releases this lock and the reconciler wakes it again.
		var claimed bool
		if err = tx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock(hashtextextended($1,48))`, job.ID.String()).Scan(&claimed); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if !claimed {
			_ = tx.Rollback(ctx)
			return nil
		}
		var stateNow string
		if err = s.db.QueryRow(ctx, `SELECT state FROM archive_import_workspace_jobs WHERE id=$1`, job.ID).Scan(&stateNow); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if stateNow == "completed" || stateNow == "failed" {
			_ = tx.Rollback(ctx)
			continue
		}
		_, err = s.db.Exec(ctx, `UPDATE archive_import_workspace_jobs SET state='running',updated_at=NOW() WHERE id=$1`, job.ID)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		_, workErr := s.processArchiveFile(ctx, job.FileID, &job.Snapshot)
		if workErr == nil {
			workErr = s.finishWorkspaceVideo(ctx, job.FileID)
		}
		if workErr == nil {
			workErr = s.refreshArchiveBatchStats(ctx, id)
		}
		state, message := "completed", ""
		if workErr != nil {
			if ctx.Err() != nil {
				_ = tx.Rollback(context.WithoutCancel(ctx))
				return ctx.Err()
			}
			state = "failed"
			if e := s.updateArchiveFileFailure(ctx, job.FileID, workErr.Error()); e != nil {
				_ = tx.Rollback(ctx)
				return e
			}
			message = sanitizeArchiveText(workErr.Error())
		}
		_, err = tx.Exec(ctx, `UPDATE archive_import_workspace_jobs SET state=$2,last_error=$3,updated_at=NOW() WHERE id=$1`, job.ID, state, message)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err = tx.Commit(ctx); err != nil {
			return err
		}
	}
}

// RunWorkspaceReconciler restores pending work after a lost queue message or a hard worker restart.
func (s *ArchiveImportService) RunWorkspaceReconciler(ctx context.Context) {
	reconcile := func() {
		enq, ok := s.enqueuer.(archiveWorkspaceEnqueuer)
		if !ok {
			return
		}
		rows, err := s.db.Query(ctx, `SELECT DISTINCT batch_id FROM archive_import_workspace_jobs WHERE state IN ('queued','running')`)
		if err != nil {
			s.logger.Error("archive workspace reconcile", "error", err)
			return
		}
		ids := []uuid.UUID{}
		for rows.Next() {
			var id uuid.UUID
			if err = rows.Scan(&id); err != nil {
				break
			}
			ids = append(ids, id)
		}
		if err == nil {
			err = rows.Err()
		}
		rows.Close()
		if err != nil {
			s.logger.Error("archive workspace reconcile", "error", err)
			return
		}
		for _, id := range ids {
			if err = enq.EnqueueArchiveWorkspace(id); err != nil {
				s.logger.Warn("archive workspace enqueue", "error", err)
			}
		}
	}
	reconcile()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			reconcile()
		}
	}
}

// finishWorkspaceVideo repairs the record-before-enqueue gap on every persisted attempt.
func (s *ArchiveImportService) finishWorkspaceVideo(ctx context.Context, fileID uuid.UUID) error {
	file, err := s.getArchiveFile(ctx, fileID)
	if err != nil {
		return err
	}
	if file.LinkedVideoID == nil {
		return nil
	}
	if s.repo == nil || s.enqueuer == nil {
		return fmt.Errorf("视频后续处理服务不可用")
	}
	video, err := s.repo.GetVideoByID(ctx, *file.LinkedVideoID)
	if err != nil {
		return err
	}
	switch video.Status {
	case "uploaded":
		return s.enqueuer.EnqueueTranscode(video.ID.String(), video.OriginalPath, s.storageRoot+"/videos", "mp4", false)
	case "scraping", "tv_pending", "av_scrape_pending":
		switch video.Type {
		case "movie":
			return s.enqueuer.EnqueueScrapeMovie(video.ID.String(), video.OriginalPath, file.RelativePath)
		case "episode":
			return s.enqueuer.EnqueueScrapeTV(video.ID.String(), video.OriginalPath, file.RelativePath)
		case "av":
			return s.enqueuer.EnqueueScrapeAV(video.ID.String(), video.OriginalPath, file.RelativePath)
		}
	}
	return nil
}

// AcquireLegacyWorkspaceGuard serializes legacy mutations with workspace initialization.
func (s *ArchiveImportService) AcquireLegacyWorkspaceGuard(ctx context.Context, kind string, id uuid.UUID) (func(), error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	fail := func(e error) (func(), error) { _ = tx.Rollback(ctx); return nil, e }
	batchID := id
	switch kind {
	case "file":
		err = tx.QueryRow(ctx, `SELECT batch_id FROM archive_import_files WHERE id=$1`, id).Scan(&batchID)
	case "group":
		err = tx.QueryRow(ctx, `SELECT batch_id FROM archive_import_groups WHERE id=$1`, id).Scan(&batchID)
	}
	if err != nil {
		return fail(err)
	}
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,47))`, batchID.String()); err != nil {
		return fail(err)
	}
	var found uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT id FROM archive_import_batches WHERE id=$1`, batchID).Scan(&found); err != nil {
		return fail(err)
	}
	var exists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM archive_import_workspaces WHERE batch_id=$1) OR (SELECT workspace_generation FROM archive_import_batches WHERE id=$1)`, batchID).Scan(&exists); err != nil {
		return fail(err)
	}
	if exists {
		return fail(fmt.Errorf("该批次已使用全屏工作台，请在工作台内编辑或提交"))
	}
	return func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }, nil
}

func lockWorkspaceFiles(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	rows, err := tx.Query(ctx, `SELECT id FROM archive_import_files WHERE batch_id=$1 ORDER BY id FOR UPDATE`, id)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			return err
		}
	}
	return rows.Err()
}
