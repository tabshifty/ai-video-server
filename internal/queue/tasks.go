package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"video-server/internal/repository"
	"video-server/internal/services"
)

const (
	// TypeVideoTranscode is the task type for transcoding uploads.
	TypeVideoTranscode = "video:transcode"
	TypeScrapeMovie    = "video:scrape:movie"
	TypeScrapeTV       = "video:scrape:tv"
	TypeScrapeAV       = "video:scrape:av"
)

// TranscodePayload carries identifiers for worker-side processing.
type TranscodePayload struct {
	VideoID      string `json:"video_id"`
	InputPath    string `json:"input_path"`
	OutputDir    string `json:"output_dir"`
	TargetFormat string `json:"target_format"`
}

// Enqueuer wraps asynq client operations.
type Enqueuer struct {
	client *asynq.Client
	queue  string
}

func NewEnqueuer(redisAddr, redisPassword, queue string) *Enqueuer {
	return &Enqueuer{
		client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr, Password: redisPassword}),
		queue:  queue,
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
		asynq.MaxRetry(3),
		asynq.ProcessIn(2*time.Second),
		asynq.Queue(e.queue),
	)
	if err != nil {
		return fmt.Errorf("enqueue transcode task: %w", err)
	}
	return nil
}

// Processor handles task registration and processing logic.
type Processor struct {
	repo     *repository.VideoRepository
	trans    *services.TranscodeService
	scrape   *services.ScraperService
	enqueuer *Enqueuer
	logger   *slog.Logger
	uploadGC bool
}

func NewProcessor(repo *repository.VideoRepository, trans *services.TranscodeService, scrape *services.ScraperService, enqueuer *Enqueuer, logger *slog.Logger) *Processor {
	return &Processor{repo: repo, trans: trans, scrape: scrape, enqueuer: enqueuer, logger: logger, uploadGC: true}
}

func (p *Processor) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeVideoTranscode, p.HandleTranscode)
	mux.HandleFunc(TypeScrapeMovie, p.HandleScrapeMovie)
	mux.HandleFunc(TypeScrapeTV, p.HandleScrapeTV)
	mux.HandleFunc(TypeScrapeAV, p.HandleScrapeAV)
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
	if video.Status == "ready" || video.Status == "processing" {
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

	jobID, err := p.repo.InsertTranscodingJob(ctx, videoID, video.UserID, "running")
	if err != nil {
		return err
	}
	_ = p.repo.UpdateVideoStatus(ctx, videoID, "processing")

	result, transcodeErr := p.trans.Process(ctx, video.ID, inputPath, video.Type)
	if transcodeErr != nil {
		_ = p.repo.MarkVideoFailed(ctx, videoID, transcodeErr.Error())
		_ = p.repo.FinishTranscodingJob(ctx, jobID, "failed", transcodeErr.Error())
		p.logger.Error("transcode failed", "video_id", videoID, "error", transcodeErr)
		return transcodeErr
	}

	if err := p.repo.UpdateTranscodeResult(ctx, videoID, result.TranscodedPath, result.ThumbnailPath, result.Duration, result.Width, result.Height, result.Metadata); err != nil {
		_ = p.repo.MarkVideoFailed(ctx, videoID, err.Error())
		_ = p.repo.FinishTranscodingJob(ctx, jobID, "failed", err.Error())
		return err
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
