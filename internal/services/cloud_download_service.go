package services

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/cloud"
	"video-server/internal/models"
	"video-server/internal/repository"
)

// CloudDownloadEnqueuer schedules the worker task for one task id.
type CloudDownloadEnqueuer interface {
	EnqueueCloudDownload(taskID int64, retryCount int) error
}

// cloudDownloadRepository is the persistence surface used by the service.
// *repository.CloudDownloadRepository implements it; tests provide a fake.
type cloudDownloadRepository interface {
	CreateOrGet(ctx context.Context, task models.CloudDownloadTask) (models.CloudDownloadTask, bool, error)
	GetByID(ctx context.Context, id int64) (models.CloudDownloadTask, error)
	List(ctx context.Context, status string, page, pageSize int) ([]models.CloudDownloadTask, int, error)
	StatusByHashes(ctx context.Context, hashes []string) (map[string]models.CloudDownloadStatus, error)
	SetStatus(ctx context.Context, id int64, status, blockedReason, errorMessage string) error
	SetProgress(ctx context.Context, id int64, percent float64, cloudStatus string) error
	SetReady(ctx context.Context, id int64, pickCode string) error
	SetLocalPath(ctx context.Context, id int64, localPath string) error
	BumpRetry(ctx context.Context, id int64) error
	ClaimDownloadTrigger(ctx context.Context, id int64) (bool, error)
	ClearDownloadTrigger(ctx context.Context, id int64) error
	ListUnfinished(ctx context.Context, limit int) ([]models.CloudDownloadTask, error)
}

// cloudSession is one attached 115 browser session. It exists so the state
// machine can be tested without a real browser.
type cloudSession interface {
	Close()
	Quota(ctx context.Context) (cloud.Quota, error)
	TargetDirID(ctx context.Context) (string, error)
	SubmitURLs(ctx context.Context, links []string, wpPathID string) (cloud.SubmitResult, error)
	FindTask(ctx context.Context, infoHash string) (*cloud.Task, error)
	WaitForTask(ctx context.Context, infoHash string, interval time.Duration, onProgress func(cloud.Task)) (*cloud.Task, error)
	CreateDownloadTask(ctx context.Context, name, pickCode string) error
}

type cloudSessionProvider func(ctx context.Context) (cloudSession, cloud.State, error)

// CloudDownloadService orchestrates parse -> 115 submit -> 115 wait -> download.
type CloudDownloadService struct {
	repo             cloudDownloadRepository
	client           *cloud.Client
	downloadDir      string
	enqueuer         CloudDownloadEnqueuer
	logger           *slog.Logger
	pollInterval     time.Duration
	filePollInterval time.Duration
	fileWaitTimeout  time.Duration
	ensureSession    cloudSessionProvider
}

// NewCloudDownloadService builds the service.
func NewCloudDownloadService(
	repo *repository.CloudDownloadRepository,
	client *cloud.Client,
	downloadDir string,
	enqueuer CloudDownloadEnqueuer,
	logger *slog.Logger,
) *CloudDownloadService {
	svc := &CloudDownloadService{
		repo:             repo,
		client:           client,
		downloadDir:      downloadDir,
		enqueuer:         enqueuer,
		logger:           logger,
		pollInterval:     15 * time.Second,
		filePollInterval: 5 * time.Second,
		fileWaitTimeout:  5 * time.Hour,
	}
	if client != nil {
		svc.ensureSession = func(ctx context.Context) (cloudSession, cloud.State, error) {
			session, state, err := client.EnsureSession(ctx)
			if err != nil {
				return nil, state, err
			}
			return liveCloudSession{session: session}, state, nil
		}
	}
	return svc
}

// Submit parses a link, records the task (deduplicated by info hash) and
// schedules it. It returns the task plus whether this call created it.
func (s *CloudDownloadService) Submit(ctx context.Context, postID *uuid.UUID, rawLink string, createdBy *uuid.UUID, source string) (models.CloudDownloadTask, bool, error) {
	link, err := cloud.ParseLink(rawLink)
	if err != nil {
		return models.CloudDownloadTask{}, false, err
	}
	source = strings.TrimSpace(source)
	if source == "" {
		source = models.CloudDownloadSourceManual
		if postID != nil {
			source = models.CloudDownloadSourceForumPost
		}
	}
	switch source {
	case models.CloudDownloadSourceForumPost, models.CloudDownloadSourceToolbox, models.CloudDownloadSourceManual:
	default:
		return models.CloudDownloadTask{}, false, fmt.Errorf("invalid cloud download source %q", source)
	}
	task, created, err := s.repo.CreateOrGet(ctx, models.CloudDownloadTask{
		PostID:       postID,
		Source:       source,
		Protocol:     string(link.Kind),
		InfoHash:     link.InfoHash,
		SourceURL:    link.Raw,
		DeclaredName: link.Name,
		DeclaredSize: link.SizeBytes,
		Status:       models.CloudDownloadPending,
		CreatedBy:    createdBy,
	})
	if err != nil {
		return models.CloudDownloadTask{}, false, err
	}

	switch {
	case created || task.Status == models.CloudDownloadPending:
		// Fresh row or a pending row whose enqueue was lost; schedule it.
	case task.Status == models.CloudDownloadFailed || task.Status == models.CloudDownloadBlocked:
		// A new user action is also an explicit retry: reset the attempt so
		// the worker starts from submit and clears the old trigger marker.
		if err := s.repo.BumpRetry(ctx, task.ID); err != nil {
			return task, created, err
		}
		task, err = s.repo.GetByID(ctx, task.ID)
		if err != nil {
			return task, created, err
		}
	default:
		// submitting/queued/ready/downloading: the reconciler owns the row.
		return task, created, nil
	}

	if err := s.enqueuer.EnqueueCloudDownload(task.ID, task.RetryCount); err != nil {
		return task, created, fmt.Errorf("enqueue cloud download: %w", err)
	}
	return task, created, nil
}

// SubmitBatch explicitly submits up to CloudDownloadBatchLimit links. It
// creates no batch entity: every link becomes an independent task and gets
// its own outcome. Existing completed/active tasks are skipped; failed or
// blocked tasks are retried; one item's failure never aborts the batch.
func (s *CloudDownloadService) SubmitBatch(ctx context.Context, links []string, createdBy *uuid.UUID) (models.CloudDownloadBatchResult, error) {
	result := models.CloudDownloadBatchResult{Items: make([]models.CloudDownloadBatchItem, len(links))}
	if len(links) == 0 {
		return result, fmt.Errorf("links 不能为空")
	}
	if len(links) > models.CloudDownloadBatchLimit {
		return result, fmt.Errorf("单次最多 %d 条链接", models.CloudDownloadBatchLimit)
	}

	type candidate struct {
		index int
		raw   string
		link  cloud.Link
	}
	candidates := make([]candidate, 0, len(links))
	seen := make(map[string]int, len(links))
	for index, raw := range links {
		trimmed := strings.TrimSpace(raw)
		result.Items[index] = models.CloudDownloadBatchItem{Link: trimmed}
		if trimmed == "" {
			result.Items[index].Status = models.CloudBatchInvalid
			result.Items[index].Message = "链接为空"
			continue
		}
		parsed, err := cloud.ParseLink(trimmed)
		if err != nil {
			result.Items[index].Status = models.CloudBatchInvalid
			result.Items[index].Message = err.Error()
			continue
		}
		result.Items[index].InfoHash = parsed.InfoHash
		if firstIndex, ok := seen[parsed.InfoHash]; ok {
			result.Items[index].Status = models.CloudBatchDuplicate
			result.Items[index].Message = fmt.Sprintf("与第 %d 条重复", firstIndex+1)
			continue
		}
		seen[parsed.InfoHash] = index
		candidates = append(candidates, candidate{index: index, raw: trimmed, link: parsed})
	}

	hashes := make([]string, 0, len(candidates))
	for _, item := range candidates {
		hashes = append(hashes, item.link.InfoHash)
	}
	statuses := map[string]models.CloudDownloadStatus{}
	if len(hashes) > 0 {
		loaded, err := s.repo.StatusByHashes(ctx, hashes)
		if err != nil {
			return result, fmt.Errorf("load cloud download statuses: %w", err)
		}
		statuses = loaded
	}

	for _, item := range candidates {
		existing, exists := statuses[item.link.InfoHash]
		if exists {
			switch existing.Status {
			case models.CloudDownloadCompleted:
				result.Items[item.index].Status = models.CloudBatchSkippedCompleted
				id := existing.TaskID
				result.Items[item.index].TaskID = &id
				continue
			case models.CloudDownloadPending, models.CloudDownloadSubmitting,
				models.CloudDownloadQueued, models.CloudDownloadReady,
				models.CloudDownloadDownloading:
				result.Items[item.index].Status = models.CloudBatchSkippedActive
				id := existing.TaskID
				result.Items[item.index].TaskID = &id
				continue
			case models.CloudDownloadFailed, models.CloudDownloadBlocked:
				// Explicit user action: retry below.
			default:
				// Unknown future state: do not risk resetting it.
				result.Items[item.index].Status = models.CloudBatchSkippedActive
				id := existing.TaskID
				result.Items[item.index].TaskID = &id
				continue
			}
		}

		task, created, err := s.Submit(ctx, nil, item.raw, createdBy, models.CloudDownloadSourceToolbox)
		if err != nil {
			result.Items[item.index].Status = models.CloudBatchFailed
			result.Items[item.index].Message = err.Error()
			continue
		}
		id := task.ID
		result.Items[item.index].TaskID = &id
		switch {
		case created:
			result.Items[item.index].Status = models.CloudBatchSubmitted
		case task.Status == models.CloudDownloadCompleted:
			result.Items[item.index].Status = models.CloudBatchSkippedCompleted
		case exists && (existing.Status == models.CloudDownloadFailed || existing.Status == models.CloudDownloadBlocked):
			result.Items[item.index].Status = models.CloudBatchRetried
		case isCloudDownloadBatchActive(task.Status):
			// A concurrent submitter won the create race; do not report a new
			// submission for a task this request did not create.
			result.Items[item.index].Status = models.CloudBatchSkippedActive
		default:
			result.Items[item.index].Status = models.CloudBatchFailed
			result.Items[item.index].Message = "任务状态已变化，请刷新后重试"
		}
	}

	result.Summary = summarizeCloudDownloadBatch(result.Items)
	return result, nil
}

func isCloudDownloadBatchActive(status string) bool {
	switch status {
	case models.CloudDownloadPending, models.CloudDownloadSubmitting,
		models.CloudDownloadQueued, models.CloudDownloadReady,
		models.CloudDownloadDownloading:
		return true
	default:
		return false
	}
}

func summarizeCloudDownloadBatch(items []models.CloudDownloadBatchItem) models.CloudDownloadBatchSummary {
	summary := models.CloudDownloadBatchSummary{Total: len(items)}
	for _, item := range items {
		switch item.Status {
		case models.CloudBatchSubmitted:
			summary.Submitted++
		case models.CloudBatchRetried:
			summary.Retried++
		case models.CloudBatchSkippedCompleted:
			summary.SkippedCompleted++
		case models.CloudBatchSkippedActive:
			summary.SkippedActive++
		case models.CloudBatchDuplicate:
			summary.Duplicate++
		case models.CloudBatchInvalid:
			summary.Invalid++
		case models.CloudBatchFailed:
			summary.Failed++
		}
	}
	return summary
}

// Retry resets a failed/blocked task and schedules it again.
func (s *CloudDownloadService) Retry(ctx context.Context, id int64) (models.CloudDownloadTask, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.CloudDownloadTask{}, fmt.Errorf("%w: %d", models.ErrCloudDownloadNotFound, id)
		}
		return models.CloudDownloadTask{}, err
	}
	if task.Status != models.CloudDownloadFailed && task.Status != models.CloudDownloadBlocked {
		return task, fmt.Errorf("%w: %s", models.ErrCloudDownloadNotRetryable, task.Status)
	}
	if err := s.repo.BumpRetry(ctx, id); err != nil {
		return models.CloudDownloadTask{}, err
	}
	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.CloudDownloadTask{}, err
	}
	if err := s.enqueuer.EnqueueCloudDownload(updated.ID, updated.RetryCount); err != nil {
		// Keep the row recoverable instead of leaving it pending with no job.
		_ = s.fail(ctx, id, fmt.Sprintf("重新入队失败: %v", err))
		return models.CloudDownloadTask{}, fmt.Errorf("enqueue cloud download: %w", err)
	}
	return updated, nil
}

// Get returns one task.
func (s *CloudDownloadService) Get(ctx context.Context, id int64) (models.CloudDownloadTask, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns a filtered page.
func (s *CloudDownloadService) List(ctx context.Context, status string, page, pageSize int) ([]models.CloudDownloadTask, int, error) {
	return s.repo.List(ctx, status, page, pageSize)
}

// StatusByHashes returns the compact status map for the forum list.
func (s *CloudDownloadService) StatusByHashes(ctx context.Context, hashes []string) (map[string]models.CloudDownloadStatus, error) {
	return s.repo.StatusByHashes(ctx, hashes)
}

// BrowserState reports whether the browser is reachable and logged in.
func (s *CloudDownloadService) BrowserState(ctx context.Context) cloud.State {
	if s.client == nil {
		return cloud.State{}
	}
	return s.client.Inspect(ctx)
}

// Reconcile re-enqueues every non-terminal task. asynq task-ID uniqueness
// makes this safe to run while another worker is processing the same row.
func (s *CloudDownloadService) Reconcile(ctx context.Context) error {
	tasks, err := s.repo.ListUnfinished(ctx, 500)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return nil
	}
	enqueued := 0
	for _, task := range tasks {
		if err := s.enqueuer.EnqueueCloudDownload(task.ID, task.RetryCount); err != nil {
			s.logWarn("reconcile cloud download: enqueue failed", "task_id", task.ID, "error", err)
			continue
		}
		enqueued++
	}
	s.logInfo("cloud download reconciled", "unfinished", len(tasks), "enqueued", enqueued)
	return nil
}

// RunReconciler runs one reconciliation immediately and then periodically
// until ctx is canceled. It is started by the worker process.
func (s *CloudDownloadService) RunReconciler(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 2 * time.Minute
	}
	s.reconcileOnce(ctx)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.reconcileOnce(ctx)
		}
	}
}

func (s *CloudDownloadService) reconcileOnce(ctx context.Context) {
	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := s.Reconcile(runCtx); err != nil {
		s.logWarn("reconcile cloud download failed", "error", err)
	}
}

// MarkFailed records a terminal failure after asynq exhausted its retries.
func (s *CloudDownloadService) MarkFailed(ctx context.Context, id int64, message string) error {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if models.CloudDownloadTerminal(task.Status) || task.Status == models.CloudDownloadBlocked {
		return nil
	}
	return s.fail(ctx, id, truncateMessage(message, 800))
}

// Process advances one task. It is idempotent: every step tolerates the task
// having already been submitted, finished or downloaded before a restart.
func (s *CloudDownloadService) Process(ctx context.Context, id int64) error {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // stale queue job for a deleted task
		}
		return err
	}
	if models.CloudDownloadTerminal(task.Status) || task.Status == models.CloudDownloadBlocked {
		// blocked needs an explicit user retry, which resets the row to pending.
		return nil
	}
	if s.ensureSession == nil {
		return s.fail(ctx, id, "云下载浏览器客户端未配置")
	}

	session, _, err := s.ensureSession(ctx)
	if err != nil {
		return s.handleSessionError(ctx, id, err)
	}
	defer session.Close()

	quota, quotaErr := session.Quota(ctx)
	if quotaErr == nil && quota.Known && quota.SurplusKnown && quota.Surplus <= 0 &&
		(task.Status == models.CloudDownloadPending || task.Status == models.CloudDownloadSubmitting) {
		return s.block(ctx, id, models.CloudBlockedQuotaExhausted, "115 云下载配额已用尽")
	}

	// submitting means an earlier attempt was interrupted somewhere between
	// SetStatus and the 115 response. A retried pending row can also still be
	// present on 115 after a manual retry. Ask 115 first: if the task is
	// already there (finished included), continue with waiting; otherwise
	// submit. A fresh pending row skips the lookup and relies on 10008.
	submitNeeded := task.Status == models.CloudDownloadPending
	if task.Status == models.CloudDownloadSubmitting ||
		(task.Status == models.CloudDownloadPending && task.RetryCount > 0) {
		existing, findErr := session.FindTask(ctx, task.InfoHash)
		if findErr != nil {
			return findErr // transient: retry the lookup on the next attempt
		}
		submitNeeded = existing == nil
	}
	if submitNeeded {
		if err := s.repo.SetStatus(ctx, id, models.CloudDownloadSubmitting, "", ""); err != nil {
			return err
		}
		dirID, err := session.TargetDirID(ctx)
		if err != nil {
			return err // transient; the next attempt re-submits
		}
		result, err := session.SubmitURLs(ctx, []string{task.SourceURL}, dirID)
		if err != nil {
			return err // transient: CDP/network; the next attempt re-submits
		}
		if !result.State && result.ErrCode != cloud.ErrCodeTaskExists {
			return s.fail(ctx, id, fmt.Sprintf("115 拒绝提交（errcode %d）: %s", result.ErrCode, result.ErrorMsg))
		}
		if err := s.repo.SetStatus(ctx, id, models.CloudDownloadQueued, "", ""); err != nil {
			return err
		}
	}

	cloudTask, err := session.WaitForTask(ctx, task.InfoHash, s.pollInterval, func(progress cloud.Task) {
		if progress.Percent > 0 {
			_ = s.repo.SetProgress(ctx, id, progress.Percent, progress.StatusText)
		}
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return s.block(ctx, id, models.CloudBlockedNeedsAttention, "等待 115 云下载超时")
		}
		return s.fail(ctx, id, fmt.Sprintf("115 云下载失败: %v", err))
	}
	if cloudTask == nil || cloudTask.PickCode == "" {
		return s.fail(ctx, id, "115 未返回 pick_code")
	}
	if err := s.repo.SetReady(ctx, id, cloudTask.PickCode); err != nil {
		return err
	}

	name := cloudTask.Name
	if name == "" {
		name = task.DeclaredName
	}
	if name == "" {
		return s.fail(ctx, id, "115 未返回文件名")
	}

	// Claim the one-shot trigger. A redelivered job or a second worker sees
	// the marker and only waits for the already-started download.
	claimed, err := s.repo.ClaimDownloadTrigger(ctx, id)
	if err != nil {
		return err
	}
	if claimed {
		if err := s.repo.SetStatus(ctx, id, models.CloudDownloadDownloading, "", ""); err != nil {
			return err
		}
		if err := session.CreateDownloadTask(ctx, name, cloudTask.PickCode); err != nil {
			clearCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			_ = s.repo.ClearDownloadTrigger(clearCtx, id)
			cancel()
			return s.fail(ctx, id, fmt.Sprintf("触发本地下载失败: %v", err))
		}
	} else if err := s.repo.SetStatus(ctx, id, models.CloudDownloadDownloading, "", ""); err != nil {
		return err
	}

	expectedSize := expectedDownloadSize(cloudTask, task.DeclaredSize)
	path, err := s.waitForFile(ctx, name, expectedSize)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return s.block(ctx, id, models.CloudBlockedNeedsAttention, "等待本地文件落盘超时")
		}
		return s.fail(ctx, id, fmt.Sprintf("等待文件落盘失败: %v", err))
	}
	completeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return s.repo.SetLocalPath(completeCtx, id, path)
}

func (s *CloudDownloadService) handleSessionError(ctx context.Context, id int64, err error) error {
	switch {
	case errors.Is(err, cloud.ErrBrowserDown):
		return s.block(ctx, id, models.CloudBlockedBrowserDown, "115 浏览器未运行或调试端口不可达")
	case errors.Is(err, cloud.ErrLoginRequired):
		return s.block(ctx, id, models.CloudBlockedLoginRequired, "115 登录已失效，请在管理端扫码登录")
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return s.block(ctx, id, models.CloudBlockedNeedsAttention, "任务超时中断")
	default:
		return err // transient: let the queue retry
	}
}

// setTerminalStatus writes a final state with a context detached from the
// asynq task. On timeout the task context is already canceled, and using it
// here would leave the row stuck in a non-terminal status forever.
func (s *CloudDownloadService) setTerminalStatus(ctx context.Context, id int64, status, reason, message string) error {
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return s.repo.SetStatus(writeCtx, id, status, reason, message)
}

func (s *CloudDownloadService) block(ctx context.Context, id int64, reason, message string) error {
	return s.setTerminalStatus(ctx, id, models.CloudDownloadBlocked, reason, message)
}

func (s *CloudDownloadService) fail(ctx context.Context, id int64, message string) error {
	return s.setTerminalStatus(ctx, id, models.CloudDownloadFailed, "", message)
}

// waitForFile polls the download directory until a stable, complete file
// appears. The file must keep the same size and mtime for one poll interval so
// a downloader that writes directly to the final name cannot be mistaken for a
// finished file.
func (s *CloudDownloadService) waitForFile(ctx context.Context, name string, wantSize *int64) (string, error) {
	if err := os.MkdirAll(s.downloadDir, 0o755); err != nil {
		return "", fmt.Errorf("create download dir: %w", err)
	}
	stem := strings.TrimSuffix(name, filepath.Ext(name))
	interval := s.filePollInterval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	waitTimeout := s.fileWaitTimeout
	if waitTimeout <= 0 {
		waitTimeout = 5 * time.Hour
	}
	deadline := time.Now().Add(waitTimeout)

	type observation struct {
		size    int64
		modTime time.Time
	}
	seen := make(map[string]observation)

	for {
		if path, info, ok := findDownloadedFile(s.downloadDir, name, stem, wantSize); ok {
			current := observation{size: info.Size(), modTime: info.ModTime()}
			previous, exists := seen[path]
			seen[path] = current
			if exists && previous.size == current.size && previous.modTime.Equal(current.modTime) {
				return path, nil
			}
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("timed out waiting for %q in %s", name, s.downloadDir)
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(interval):
		}
	}
}

// findDownloadedFile returns the exact declared name when present, otherwise
// the first "name (n).ext" duplicate the native downloader creates when the
// target already exists. Partial files are skipped.
func findDownloadedFile(dir, name, stem string, wantSize *int64) (string, os.FileInfo, bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", nil, false
	}
	extension := filepath.Ext(name)
	var duplicatePath string
	var duplicateInfo os.FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		if isPartialDownloadName(fileName) {
			continue
		}
		exact := fileName == name
		duplicate := strings.HasPrefix(fileName, stem+" (") && filepath.Ext(fileName) == extension
		if !exact && !duplicate {
			continue
		}
		info, err := entry.Info()
		if err != nil || info.Size() == 0 {
			continue
		}
		if wantSize != nil && *wantSize > 0 && info.Size() != *wantSize {
			continue
		}
		if exact {
			return filepath.Join(dir, fileName), info, true
		}
		if duplicatePath == "" {
			duplicatePath = filepath.Join(dir, fileName)
			duplicateInfo = info
		}
	}
	if duplicatePath != "" {
		return duplicatePath, duplicateInfo, true
	}
	return "", nil, false
}

func isPartialDownloadName(name string) bool {
	lower := strings.ToLower(name)
	for _, suffix := range []string{".crdownload", ".download", ".part", ".tmp", ".aria2", ".!qb", ".bc!"} {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}
	return false
}

// expectedDownloadSize prefers the size 115 reports for the actual task and
// falls back to the size embedded in the link when 115 does not report it.
func expectedDownloadSize(task *cloud.Task, declared *int64) *int64 {
	if task != nil && task.Size > 0 {
		size := task.Size
		return &size
	}
	return declared
}

// LoginQR opens (or reuses) the 115 login page and returns a data URL for the
// QR code, falling back to a viewport screenshot when the QR cannot be read
// directly. The tab intentionally stays open: the login page polls 115 for the
// scan result and writes the session cookie, so closing it after capture would
// break the login. Reusing the tab avoids one new tab per refresh.
func (s *CloudDownloadService) LoginQR(ctx context.Context) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("cloud: browser client is not configured")
	}
	tab, found, err := s.client.Browser.FindTarget(ctx, func(t cloud.Target) bool {
		return is115LoginURL(t.URL)
	})
	if err != nil {
		return "", err
	}
	if !found {
		tab, err = s.client.Browser.OpenTab(ctx, "https://115.com/login")
		if err != nil {
			return "", err
		}
	}
	session, err := s.client.Browser.Attach(ctx, tab)
	if err != nil {
		return "", err
	}
	defer session.Close()

	// A reused tab may already show an expired QR; reload it so the page
	// fetches a fresh one before we capture.
	if found {
		_ = session.Reload(ctx)
	}

	const qrExpr = `(() => {
		const canvases = Array.from(document.querySelectorAll('canvas'));
		for (const c of canvases) {
			try {
				const url = c.toDataURL('image/png');
				if (url && url.length > 200) return url;
			} catch (e) {}
		}
		const imgs = Array.from(document.querySelectorAll('img'));
		for (const img of imgs) {
			if (img.src && img.src.startsWith('data:image') && img.src.length > 200) return img.src;
		}
		return '';
	})()`

	deadline := time.Now().Add(10 * time.Second)
	for {
		var dataURL string
		_ = session.Evaluate(ctx, qrExpr, &dataURL)
		if strings.HasPrefix(dataURL, "data:image") {
			return dataURL, nil
		}
		if time.Now().After(deadline) {
			break
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(400 * time.Millisecond):
		}
	}

	png, err := session.Screenshot(ctx)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

func (s *CloudDownloadService) logInfo(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Info(msg, args...)
	}
}

func (s *CloudDownloadService) logWarn(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Warn(msg, args...)
	}
}

func is115LoginURL(raw string) bool {
	if !strings.Contains(raw, "115.com") {
		return false
	}
	return strings.Contains(raw, "/login") || strings.Contains(raw, "ac=login")
}

func truncateMessage(message string, maxRunes int) string {
	runes := []rune(message)
	if len(runes) <= maxRunes {
		return message
	}
	return string(runes[:maxRunes]) + "…"
}

// liveCloudSession adapts the CDP session to the cloudSession interface.
type liveCloudSession struct {
	session *cloud.Session
}

func (l liveCloudSession) Close() { l.session.Close() }

func (l liveCloudSession) Quota(ctx context.Context) (cloud.Quota, error) {
	return cloud.QuotaInfo(ctx, l.session)
}

func (l liveCloudSession) TargetDirID(ctx context.Context) (string, error) {
	return cloud.TargetDirID(ctx, l.session)
}

func (l liveCloudSession) SubmitURLs(ctx context.Context, links []string, wpPathID string) (cloud.SubmitResult, error) {
	return cloud.SubmitURLs(ctx, l.session, links, wpPathID)
}

func (l liveCloudSession) FindTask(ctx context.Context, infoHash string) (*cloud.Task, error) {
	return cloud.FindTask(ctx, l.session, infoHash)
}

func (l liveCloudSession) WaitForTask(ctx context.Context, infoHash string, interval time.Duration, onProgress func(cloud.Task)) (*cloud.Task, error) {
	return cloud.WaitForTask(ctx, l.session, infoHash, interval, onProgress)
}

func (l liveCloudSession) CreateDownloadTask(ctx context.Context, name, pickCode string) error {
	return cloud.CreateDownloadTask(ctx, l.session, name, pickCode)
}
