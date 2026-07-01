package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/services"
	ffmpegpkg "video-server/pkg/ffmpeg"
)

const (
	// TypeVideoTranscode is the task type for transcoding uploads.
	TypeVideoTranscode = "video:transcode"
	TypeScrapeMovie    = "video:scrape:movie"
	TypeScrapeTV       = "video:scrape:tv"
	TypeScrapeAV       = "video:scrape:av"
	TypeScrapeRetag    = "video:scrape:retag"
	TypeOrphanFileScan = "system:orphan-files:scan"
	TypeEd2kDownload   = "download:ed2k"
	// TypeEd2kServerlistRefresh 是 server.met 定时刷新任务类型，由 asynq.Scheduler 按 cron 入队。
	TypeEd2kServerlistRefresh = "ed2k:serverlist:refresh"
)

// TranscodePayload carries identifiers for worker-side processing.
type TranscodePayload struct {
	VideoID      string `json:"video_id"`
	InputPath    string `json:"input_path"`
	OutputDir    string `json:"output_dir"`
	TargetFormat string `json:"target_format"`
	Force        bool   `json:"force,omitempty"`
}

// Ed2kDownloadPayload carries identifiers for ED2K download processing.
type Ed2kDownloadPayload struct {
	TaskID string `json:"task_id"`
}

// Ed2kServerlistRefreshPayload 为空 payload：刷新参数（URL/脚本路径）走 config + 执行器，
// 不随任务携带；scheduler 以空 payload 入队即可。
type Ed2kServerlistRefreshPayload struct{}

var ErrEd2kDownloadTaskInFlight = errors.New("ed2k download task already in flight")

const (
	ed2kDownloadPollMaxRetry = 20160
	ed2kDownloadPollDelay    = 30 * time.Second
	ed2kDownloadTaskTimeout  = time.Minute
)

// Enqueuer wraps asynq client operations.
type Enqueuer struct {
	client               *asynq.Client
	queue                string
	redisAddr            string
	redisPassword        string
	transcodeTaskTimeout time.Duration
}

func NewEnqueuer(redisAddr, redisPassword, queue string, transcodeTaskTimeout time.Duration) *Enqueuer {
	if transcodeTaskTimeout <= 0 {
		transcodeTaskTimeout = 6 * time.Hour
	}
	return &Enqueuer{
		client:               asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr, Password: redisPassword}),
		queue:                queue,
		redisAddr:            redisAddr,
		redisPassword:        redisPassword,
		transcodeTaskTimeout: transcodeTaskTimeout,
	}
}

func (e *Enqueuer) Close() error {
	return e.client.Close()
}

func (e *Enqueuer) EnqueueTranscode(payloadIn TranscodePayload) error {
	payload, err := json.Marshal(payloadIn)
	if err != nil {
		return fmt.Errorf("marshal transcode payload: %w", err)
	}
	_, err = e.client.Enqueue(
		asynq.NewTask(TypeVideoTranscode, payload),
		buildTranscodeTaskOptions(e.queue, e.transcodeTaskTimeout)...,
	)
	if err != nil {
		return fmt.Errorf("enqueue transcode task: %w", err)
	}
	return nil
}

func (e *Enqueuer) EnqueueEd2kDownload(payloadIn Ed2kDownloadPayload) error {
	payload, err := json.Marshal(payloadIn)
	if err != nil {
		return fmt.Errorf("marshal ed2k download payload: %w", err)
	}
	_, err = e.client.Enqueue(
		asynq.NewTask(TypeEd2kDownload, payload),
		buildEd2kDownloadTaskOptions(e.queue, payloadIn.TaskID)...,
	)
	if err != nil {
		return wrapEd2kDownloadEnqueueError(err)
	}
	return nil
}

func (e *Enqueuer) DeleteEd2kDownloadTask(taskID string) error {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil
	}
	inspector := e.newInspector()
	defer func() { _ = inspector.Close() }()
	if err := inspector.DeleteTask(e.queue, taskID); err != nil {
		if errors.Is(err, asynq.ErrTaskNotFound) {
			return nil
		}
		return fmt.Errorf("delete enqueued ed2k download task: %w", err)
	}
	return nil
}

func (e *Enqueuer) HasEd2kDownloadTask(taskID string) (bool, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return false, nil
	}
	inspector := e.newInspector()
	defer func() { _ = inspector.Close() }()
	_, err := inspector.GetTaskInfo(e.queue, taskID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, asynq.ErrTaskNotFound) {
		return false, nil
	}
	return false, fmt.Errorf("get enqueued ed2k download task info: %w", err)
}

func (e *Enqueuer) newInspector() *asynq.Inspector {
	return asynq.NewInspector(asynq.RedisClientOpt{Addr: e.redisAddr, Password: e.redisPassword})
}

func wrapEd2kDownloadEnqueueError(err error) error {
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		return fmt.Errorf("%w: %w", ErrEd2kDownloadTaskInFlight, err)
	}
	return fmt.Errorf("enqueue ed2k download task: %w", err)
}

func buildEd2kDownloadTaskOptions(queue, taskID string) []asynq.Option {
	opts := []asynq.Option{
		asynq.MaxRetry(ed2kDownloadPollMaxRetry),
		asynq.ProcessIn(2 * time.Second),
		asynq.Queue(queue),
		asynq.Timeout(ed2kDownloadTaskTimeout),
	}
	if strings.TrimSpace(taskID) != "" {
		opts = append(opts, asynq.TaskID(taskID))
	}
	return opts
}

func Ed2kDownloadRetryDelayFunc(fallback asynq.RetryDelayFunc) asynq.RetryDelayFunc {
	if fallback == nil {
		fallback = asynq.DefaultRetryDelayFunc
	}
	return func(n int, err error, task *asynq.Task) time.Duration {
		if task != nil && task.Type() == TypeEd2kDownload && errors.Is(err, ErrEd2kDownloadStillInProgress) {
			return ed2kDownloadPollDelay
		}
		return fallback(n, err, task)
	}
}

func IsEd2kDownloadFailure(err error) bool {
	return !errors.Is(err, ErrEd2kDownloadStillInProgress)
}

func buildTranscodeTaskOptions(queue string, timeout time.Duration) []asynq.Option {
	if timeout <= 0 {
		timeout = 6 * time.Hour
	}
	return []asynq.Option{
		asynq.MaxRetry(3),
		asynq.ProcessIn(2 * time.Second),
		asynq.Queue(queue),
		asynq.Timeout(timeout),
	}
}

// Processor handles task registration and processing logic.
type Processor struct {
	repo               *repository.VideoRepository
	trans              *services.TranscodeService
	scrape             *services.ScraperService
	subtitle           *services.SubtitleService
	enqueuer           *Enqueuer
	ed2kExecutor       Ed2kDownloadExecutor
	ed2kServerlistExec Ed2kServerlistRefreshExecutor
	logger             *slog.Logger
	storageRoot        string
	uploadGC           bool
}

func NewProcessor(repo *repository.VideoRepository, trans *services.TranscodeService, scrape *services.ScraperService, subtitle *services.SubtitleService, enqueuer *Enqueuer, ed2kExecutor Ed2kDownloadExecutor, ed2kServerlistExec Ed2kServerlistRefreshExecutor, logger *slog.Logger, storageRoot string) *Processor {
	return &Processor{repo: repo, trans: trans, scrape: scrape, subtitle: subtitle, enqueuer: enqueuer, ed2kExecutor: ed2kExecutor, ed2kServerlistExec: ed2kServerlistExec, logger: logger, storageRoot: storageRoot, uploadGC: true}
}

func (p *Processor) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeVideoTranscode, p.HandleTranscode)
	mux.HandleFunc(TypeScrapeMovie, p.HandleScrapeMovie)
	mux.HandleFunc(TypeScrapeTV, p.HandleScrapeTV)
	mux.HandleFunc(TypeScrapeAV, p.HandleScrapeAV)
	mux.HandleFunc(TypeScrapeRetag, p.HandleScrapeRetag)
	mux.HandleFunc(TypeOrphanFileScan, p.HandleOrphanFileScan)
	mux.HandleFunc(TypeEd2kDownload, p.HandleEd2kDownload)
	mux.HandleFunc(TypeEd2kServerlistRefresh, p.HandleEd2kServerlistRefresh)
}

func (p *Processor) HandleTranscode(ctx context.Context, task *asynq.Task) error {
	var payload TranscodePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}
	videoID, err := uuid.Parse(payload.VideoID)
	if err != nil {
		return fmt.Errorf("invalid video id: %w", err)
	}

	video, err := p.repo.GetVideoByID(ctx, videoID)
	if err != nil {
		return err
	}
	if video.Status == "ready" || (video.Status == "processing" && !payload.Force) {
		p.logger.Info("skip duplicate transcode task", "video_id", videoID, "status", video.Status)
		return nil
	}

	inputPath := payload.InputPath
	if inputPath == "" {
		inputPath = video.OriginalPath
	}
	if inputPath == "" {
		return fmt.Errorf("empty input path for video %s", videoID)
	}

	sourcePlaybackProbe, sourcePlaybackProbeErr := services.ProbePlaybackCompatibility(ctx, inputPath)
	if sourcePlaybackProbeErr != nil {
		p.logger.Warn("source playback compatibility probe failed", "video_id", videoID, "input_path", inputPath, "error", sourcePlaybackProbeErr)
	}

	jobID, err := p.repo.InsertTranscodingJob(ctx, videoID, video.UserID, "running")
	if err != nil {
		return err
	}
	_ = p.repo.UpdateVideoStatus(ctx, videoID, "processing")

	var (
		lastProgressUpdated time.Time
		lastProcessed       = -1
	)
	progressHandler := func(progress services.TranscodeProgress) {
		now := time.Now()
		if progress.ProcessedSeconds < 0 {
			progress.ProcessedSeconds = 0
		}
		if progress.SourceDurationSeconds > 0 && progress.ProcessedSeconds > progress.SourceDurationSeconds {
			progress.ProcessedSeconds = progress.SourceDurationSeconds
		}
		if progress.RemainingSeconds < 0 {
			progress.RemainingSeconds = 0
		}
		shouldPersist := lastProcessed < 0 ||
			progress.ProcessedSeconds-lastProcessed >= 1 ||
			now.Sub(lastProgressUpdated) >= time.Second ||
			progress.RemainingSeconds == 0 ||
			progress.ProgressPercent >= 100
		if !shouldPersist {
			return
		}

		progressPercent := roundTo2(progress.ProgressPercent)
		sourceDuration := intPtrIfPositive(progress.SourceDurationSeconds)
		processed := intPtr(progress.ProcessedSeconds)
		remaining := intPtr(progress.RemainingSeconds)
		if err := p.repo.UpdateTranscodingJobProgress(ctx, jobID, sourceDuration, processed, remaining, &progressPercent); err != nil {
			p.logger.Warn("update transcode progress failed", "video_id", videoID, "job_id", jobID, "error", err)
			return
		}
		lastProcessed = progress.ProcessedSeconds
		lastProgressUpdated = now
	}

	if shouldPreserveEpisodeDolbyVisionSource(video.Type, sourcePlaybackProbe, sourcePlaybackProbeErr) {
		result, directCopyErr := p.processEpisodeDolbyVisionDirectCopy(ctx, video, inputPath, sourcePlaybackProbe)
		if directCopyErr != nil {
			p.finalizeTranscodeFailure(ctx, videoID, jobID, directCopyErr.Error())
			p.logger.Error("dv direct copy failed", "video_id", videoID, "error", directCopyErr)
			return directCopyErr
		}
		if err := p.repo.UpdateTranscodeResult(ctx, videoID, result.TranscodedPath, result.ThumbnailPath, result.Duration, result.Width, result.Height, result.Metadata); err != nil {
			p.finalizeTranscodeFailure(ctx, videoID, jobID, err.Error())
			return err
		}
		if p.subtitle != nil {
			if _, err := p.subtitle.SyncEmbeddedSubtitles(ctx, videoID, result.TranscodedPath); err != nil {
				p.logger.Warn("sync embedded subtitles failed", "video_id", videoID, "input_path", result.TranscodedPath, "error", err)
			}
		}
		if result.Duration > 0 {
			sourceDuration := intPtr(result.Duration)
			processed := intPtr(result.Duration)
			remaining := intPtr(0)
			progressPercent := 100.0
			if err := p.repo.UpdateTranscodingJobProgress(ctx, jobID, sourceDuration, processed, remaining, &progressPercent); err != nil {
				p.logger.Warn("update dv direct copy progress failed", "video_id", videoID, "job_id", jobID, "error", err)
			}
		}
		if err := p.repo.FinishTranscodingJob(ctx, jobID, "success", ""); err != nil {
			return err
		}
		p.logger.Info("dv direct copy completed", "video_id", videoID, "output", result.TranscodedPath)
		return nil
	}

	result, transcodeErr := p.trans.Process(ctx, video.ID, inputPath, video.Type, progressHandler)
	if transcodeErr != nil {
		p.finalizeTranscodeFailure(ctx, videoID, jobID, transcodeErr.Error())
		p.logger.Error("transcode failed", "video_id", videoID, "error", transcodeErr)
		return transcodeErr
	}

	thumbPath, metadata := resolveTranscodePersistence(video, result)
	outputPlaybackProbe, outputPlaybackProbeErr := services.ProbePlaybackCompatibility(ctx, result.TranscodedPath)
	if outputPlaybackProbeErr != nil {
		p.logger.Warn("output playback compatibility probe failed", "video_id", videoID, "output_path", result.TranscodedPath, "error", outputPlaybackProbeErr)
	}
	metadata[services.PlaybackCompatibilityMetadataKey] = services.BuildPlaybackCompatibilityMetadata(
		sourcePlaybackProbe,
		sourcePlaybackProbeErr,
		outputPlaybackProbe,
		outputPlaybackProbeErr,
	)
	preservedSourcePath, preserveErr := p.preserveEpisodeDolbyVisionSource(video, inputPath, sourcePlaybackProbe, sourcePlaybackProbeErr)
	if preserveErr != nil {
		p.logger.Warn("preserve dv source failed, fallback to regular transcode output", "video_id", videoID, "input_path", inputPath, "error", preserveErr)
		preservedSourcePath = ""
	}
	if preservedSourcePath != "" {
		metadata = services.MergePlaybackCompatibilityMetadata(metadata, map[string]any{
			"source_playback_path": preservedSourcePath,
		})
	}
	if err := p.repo.UpdateTranscodeResult(ctx, videoID, result.TranscodedPath, thumbPath, result.Duration, result.Width, result.Height, metadata); err != nil {
		p.finalizeTranscodeFailure(ctx, videoID, jobID, err.Error())
		return err
	}
	if p.subtitle != nil {
		subtitleInputPath := inputPath
		if preservedSourcePath != "" {
			subtitleInputPath = preservedSourcePath
		}
		if _, err := p.subtitle.SyncEmbeddedSubtitles(ctx, videoID, subtitleInputPath); err != nil {
			p.logger.Warn("sync embedded subtitles failed", "video_id", videoID, "input_path", subtitleInputPath, "error", err)
		}
	}
	if err := p.repo.FinishTranscodingJob(ctx, jobID, "success", ""); err != nil {
		return err
	}
	if p.uploadGC && !shouldPreserveEpisodeDolbyVisionSource(video.Type, sourcePlaybackProbe, sourcePlaybackProbeErr) {
		_ = os.Remove(video.OriginalPath)
	}

	p.logger.Info("transcode completed", "video_id", videoID, "output", result.TranscodedPath)
	return nil
}

func (p *Processor) processEpisodeDolbyVisionDirectCopy(
	ctx context.Context,
	video models.Video,
	inputPath string,
	sourceProbe ffmpegpkg.PlaybackCompatibilityProbe,
) (services.TranscodeResult, error) {
	sourcePath, err := p.preserveEpisodeDolbyVisionSource(video, inputPath, sourceProbe, nil)
	if err != nil {
		return services.TranscodeResult{}, err
	}
	if strings.TrimSpace(sourcePath) == "" {
		return services.TranscodeResult{}, fmt.Errorf("dv direct copy: source path not preserved")
	}
	return buildDolbyVisionDirectCopyResult(ctx, sourcePath, sourceProbe, ffmpegpkg.Probe, ffmpegpkg.Thumbnail)
}

type videoProbeFunc func(context.Context, string) (ffmpegpkg.VideoProbe, error)
type thumbnailFunc func(context.Context, string, string) error

func buildDolbyVisionDirectCopyResult(
	ctx context.Context,
	sourcePath string,
	sourceProbe ffmpegpkg.PlaybackCompatibilityProbe,
	probe videoProbeFunc,
	thumbnail thumbnailFunc,
) (services.TranscodeResult, error) {
	sourcePath = strings.TrimSpace(sourcePath)
	if sourcePath == "" {
		return services.TranscodeResult{}, fmt.Errorf("dv direct copy: empty source path")
	}
	thumbPath := filepath.Join(filepath.Dir(sourcePath), "thumb.jpg")
	if err := thumbnail(ctx, sourcePath, thumbPath); err != nil {
		return services.TranscodeResult{}, fmt.Errorf("create dv direct copy thumbnail: %w", err)
	}

	videoProbe, probeErr := probe(ctx, sourcePath)
	duration := 0
	width := 0
	height := 0
	metadata := map[string]any{
		"codec":             firstNonEmptyString(strings.TrimSpace(videoProbe.Codec), strings.TrimSpace(sourceProbe.Codec), "unknown"),
		"transcode_mode":    "direct_copy",
		"transcode_profile": "dv_source_copy",
		"playback_path":     sourcePath,
		"playback_codec":    firstNonEmptyString(strings.TrimSpace(sourceProbe.Codec), strings.TrimSpace(videoProbe.Codec), "source"),
	}
	if probeErr != nil {
		metadata["probe_error"] = probeErr.Error()
	} else {
		if videoProbe.Duration > 0 {
			duration = int(math.Round(videoProbe.Duration))
		}
		if videoProbe.Width > 0 {
			width = videoProbe.Width
		}
		if videoProbe.Height > 0 {
			height = videoProbe.Height
		}
	}
	metadata[services.PlaybackCompatibilityMetadataKey] = services.BuildPlaybackCompatibilityMetadata(
		sourceProbe,
		nil,
		sourceProbe,
		nil,
	)
	metadata = services.MergePlaybackCompatibilityMetadata(metadata, map[string]any{
		"source_playback_path": sourcePath,
	})

	return services.TranscodeResult{
		TranscodedPath: sourcePath,
		ThumbnailPath:  thumbPath,
		Duration:       duration,
		Width:          width,
		Height:         height,
		Metadata:       metadata,
	}, nil
}

func (p *Processor) preserveEpisodeDolbyVisionSource(
	video models.Video,
	inputPath string,
	sourceProbe ffmpegpkg.PlaybackCompatibilityProbe,
	sourceProbeErr error,
) (string, error) {
	if !shouldPreserveEpisodeDolbyVisionSource(video.Type, sourceProbe, sourceProbeErr) {
		return "", nil
	}
	inputPath = strings.TrimSpace(inputPath)
	if inputPath == "" {
		return "", fmt.Errorf("preserve dv source: empty input path")
	}
	targetPath, err := p.episodeDolbyVisionSourceTargetPath(video, inputPath)
	if err != nil {
		return "", err
	}
	if services.IsSameFilePath(inputPath, targetPath) {
		return targetPath, nil
	}
	if _, err := os.Stat(inputPath); err != nil && os.IsNotExist(err) {
		info, targetErr := os.Stat(targetPath)
		if targetErr == nil && !info.IsDir() {
			return targetPath, nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return "", fmt.Errorf("preserve dv source: create target dir: %w", err)
	}
	if err := moveFilePreservingSource(inputPath, targetPath); err != nil {
		return "", fmt.Errorf("preserve dv source: move source file: %w", err)
	}
	return targetPath, nil
}

func (p *Processor) episodeDolbyVisionSourceTargetPath(video models.Video, inputPath string) (string, error) {
	inputPath = strings.TrimSpace(inputPath)
	if inputPath == "" {
		return "", fmt.Errorf("preserve dv source: empty input path")
	}
	ext := strings.TrimSpace(filepath.Ext(inputPath))
	if ext == "" {
		return "", fmt.Errorf("preserve dv source: source extension missing")
	}
	return filepath.Join(p.storageRoot, "videos", video.ID.String(), "source-dv"+ext), nil
}

func intPtr(v int) *int {
	return &v
}

func intPtrIfPositive(v int) *int {
	if v <= 0 {
		return nil
	}
	return &v
}

func roundTo2(v float64) float64 {
	return math.Round(v*100) / 100
}

func resolveTranscodePersistence(video models.Video, result services.TranscodeResult) (string, map[string]any) {
	if video.Type != "av" {
		return result.ThumbnailPath, cloneMetadata(result.Metadata)
	}

	existingMetadata := decodeMetadataMap(video.Metadata)
	metadata := mergeMetadata(existingMetadata, result.Metadata)
	if !hasAVScrapedPosterMetadata(existingMetadata) {
		return result.ThumbnailPath, metadata
	}

	for key, value := range existingMetadata {
		if isAVPosterMetadataKey(key) {
			metadata[key] = value
		}
	}
	return chooseAVPosterThumbnailPath(existingMetadata, video.ThumbnailPath), metadata
}

func decodeMetadataMap(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var metadata map[string]any
	if err := json.Unmarshal(raw, &metadata); err != nil || metadata == nil {
		return map[string]any{}
	}
	return metadata
}

func mergeMetadata(base, patch map[string]any) map[string]any {
	merged := cloneMetadata(base)
	for key, value := range patch {
		merged[key] = value
	}
	return merged
}

func cloneMetadata(metadata map[string]any) map[string]any {
	if metadata == nil {
		return map[string]any{}
	}
	clone := make(map[string]any, len(metadata))
	for key, value := range metadata {
		clone[key] = value
	}
	return clone
}

func hasAVScrapedPosterMetadata(metadata map[string]any) bool {
	for _, key := range []string{
		"poster_original_file_path",
		"poster_cropped_file_path",
		"poster_thumb_file_path",
		"poster_original_path",
		"poster_cropped_path",
		"poster_thumb_path",
	} {
		if strings.TrimSpace(stringFromAny(metadata[key])) != "" {
			return true
		}
	}
	return false
}

func isAVPosterMetadataKey(key string) bool {
	key = strings.TrimSpace(key)
	return strings.HasPrefix(key, "poster_") || key == "poster" || key == "thumb" || key == "thumb_url"
}

func chooseAVPosterThumbnailPath(metadata map[string]any, fallback string) string {
	fallback = strings.TrimSpace(fallback)
	if fallback != "" {
		return fallback
	}

	originalFilePath := strings.TrimSpace(stringFromAny(metadata["poster_original_file_path"]))
	thumbFilePath := strings.TrimSpace(stringFromAny(metadata["poster_thumb_file_path"]))
	croppedFilePath := strings.TrimSpace(stringFromAny(metadata["poster_cropped_file_path"]))
	switch strings.ToLower(strings.TrimSpace(stringFromAny(metadata["poster_variant"]))) {
	case "thumb":
		return firstNonEmptyString(thumbFilePath, croppedFilePath, originalFilePath)
	case "original":
		return firstNonEmptyString(originalFilePath, thumbFilePath, croppedFilePath)
	default:
		return firstNonEmptyString(croppedFilePath, thumbFilePath, originalFilePath)
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func stringFromAny(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func shouldPreserveEpisodeDolbyVisionSource(videoType string, sourceProbe ffmpegpkg.PlaybackCompatibilityProbe, sourceProbeErr error) bool {
	return strings.EqualFold(strings.TrimSpace(videoType), "episode") && sourceProbeErr == nil && sourceProbe.VideoStreamFound && sourceProbe.DolbyVision
}

func moveFilePreservingSource(srcPath, dstPath string) error {
	if err := os.Rename(srcPath, dstPath); err == nil {
		return nil
	}
	if err := ffmpegpkg.CopyFile(dstPath, srcPath); err != nil {
		return err
	}
	if err := os.Remove(srcPath); err != nil {
		_ = os.Remove(dstPath)
		return err
	}
	return nil
}

func (p *Processor) finalizeTranscodeFailure(ctx context.Context, videoID uuid.UUID, jobID int64, reason string) {
	finalizeCtx := ctx
	cancel := func() {}
	if ctx == nil || ctx.Err() != nil {
		finalizeCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	}
	defer cancel()

	if err := p.repo.MarkVideoFailed(finalizeCtx, videoID, reason); err != nil {
		p.logger.Error("mark video failed state failed", "video_id", videoID, "job_id", jobID, "error", err)
	}
	if err := p.repo.FinishTranscodingJob(finalizeCtx, jobID, "failed", reason); err != nil {
		p.logger.Error("finish transcoding job failed state failed", "video_id", videoID, "job_id", jobID, "error", err)
	}
}
