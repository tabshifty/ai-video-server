package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/services"
)

const (
	// TypeVideoTranscode is the task type for transcoding uploads.
	TypeVideoTranscode = "video:transcode"
	TypeScrapeMovie    = "video:scrape:movie"
	TypeScrapeTV       = "video:scrape:tv"
	TypeScrapeAV       = "video:scrape:av"
	TypeScrapeRetag    = "video:scrape:retag"
	TypeOrphanFileScan = "system:orphan-files:scan"
)

// TranscodePayload carries identifiers for worker-side processing.
type TranscodePayload struct {
	VideoID      string `json:"video_id"`
	InputPath    string `json:"input_path"`
	OutputDir    string `json:"output_dir"`
	TargetFormat string `json:"target_format"`
	Force        bool   `json:"force,omitempty"`
}

// Enqueuer wraps asynq client operations.
type Enqueuer struct {
	client               *asynq.Client
	queue                string
	transcodeTaskTimeout time.Duration
}

func NewEnqueuer(redisAddr, redisPassword, queue string, transcodeTaskTimeout time.Duration) *Enqueuer {
	if transcodeTaskTimeout <= 0 {
		transcodeTaskTimeout = 6 * time.Hour
	}
	return &Enqueuer{
		client:               asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr, Password: redisPassword}),
		queue:                queue,
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
	repo        *repository.VideoRepository
	trans       *services.TranscodeService
	scrape      *services.ScraperService
	subtitle    *services.SubtitleService
	enqueuer    *Enqueuer
	logger      *slog.Logger
	storageRoot string
	uploadGC    bool
}

func NewProcessor(repo *repository.VideoRepository, trans *services.TranscodeService, scrape *services.ScraperService, subtitle *services.SubtitleService, enqueuer *Enqueuer, logger *slog.Logger, storageRoot string) *Processor {
	return &Processor{repo: repo, trans: trans, scrape: scrape, subtitle: subtitle, enqueuer: enqueuer, logger: logger, storageRoot: storageRoot, uploadGC: true}
}

func (p *Processor) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeVideoTranscode, p.HandleTranscode)
	mux.HandleFunc(TypeScrapeMovie, p.HandleScrapeMovie)
	mux.HandleFunc(TypeScrapeTV, p.HandleScrapeTV)
	mux.HandleFunc(TypeScrapeAV, p.HandleScrapeAV)
	mux.HandleFunc(TypeScrapeRetag, p.HandleScrapeRetag)
	mux.HandleFunc(TypeOrphanFileScan, p.HandleOrphanFileScan)
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
	if err := p.repo.UpdateTranscodeResult(ctx, videoID, result.TranscodedPath, thumbPath, result.Duration, result.Width, result.Height, metadata); err != nil {
		p.finalizeTranscodeFailure(ctx, videoID, jobID, err.Error())
		return err
	}
	if p.subtitle != nil {
		if _, err := p.subtitle.SyncEmbeddedSubtitles(ctx, videoID, inputPath); err != nil {
			p.logger.Warn("sync embedded subtitles failed", "video_id", videoID, "input_path", inputPath, "error", err)
		}
	}
	if err := p.repo.FinishTranscodingJob(ctx, jobID, "success", ""); err != nil {
		return err
	}
	if p.uploadGC {
		_ = os.Remove(video.OriginalPath)
	}

	p.logger.Info("transcode completed", "video_id", videoID, "output", result.TranscodedPath)
	return nil
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
