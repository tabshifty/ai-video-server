package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// ScrapePayload carries metadata discovery input for movie/episode uploads.
type ScrapePayload struct {
	VideoID  string `json:"video_id"`
	FilePath string `json:"file_path"`
	Filename string `json:"filename"`
	Type     string `json:"type"` // movie | episode
}

func (e *Enqueuer) EnqueueScrapeMovie(payloadIn ScrapePayload) error {
	return e.enqueueScrapeTask(TypeScrapeMovie, payloadIn)
}

func (e *Enqueuer) EnqueueScrapeTV(payloadIn ScrapePayload) error {
	return e.enqueueScrapeTask(TypeScrapeTV, payloadIn)
}

func (e *Enqueuer) enqueueScrapeTask(taskType string, payloadIn ScrapePayload) error {
	payload, err := json.Marshal(payloadIn)
	if err != nil {
		return fmt.Errorf("marshal scrape payload: %w", err)
	}
	_, err = e.client.Enqueue(
		asynq.NewTask(taskType, payload),
		asynq.MaxRetry(3),
		asynq.Queue(e.queue),
	)
	if err != nil {
		return fmt.Errorf("enqueue scrape task: %w", err)
	}
	return nil
}

func (p *Processor) HandleScrapeMovie(ctx context.Context, task *asynq.Task) error {
	return p.handleScrape(ctx, task, "movie")
}

func (p *Processor) HandleScrapeTV(ctx context.Context, task *asynq.Task) error {
	return p.handleScrape(ctx, task, "episode")
}

func (p *Processor) handleScrape(ctx context.Context, task *asynq.Task, expectedType string) error {
	if p.scrape == nil || p.enqueuer == nil {
		return fmt.Errorf("scrape processor not configured")
	}

	var payload ScrapePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal scrape payload: %w", err)
	}
	videoID, err := uuid.Parse(payload.VideoID)
	if err != nil {
		return fmt.Errorf("invalid video id: %w", err)
	}
	video, err := p.repo.GetVideoByID(ctx, videoID)
	if err != nil {
		return err
	}
	if video.Type != expectedType {
		return fmt.Errorf("video type mismatch: want=%s got=%s", expectedType, video.Type)
	}
	if video.Status == "ready" || video.Status == "processing" {
		p.logger.Info("skip scrape task on finalized video", "video_id", videoID, "status", video.Status)
		return nil
	}

	var scrapeErr error
	if expectedType == "movie" {
		_, scrapeErr = p.scrape.ScrapeMovieUpload(ctx, videoID, payload.FilePath, payload.Filename)
	} else {
		_, scrapeErr = p.scrape.ScrapeEpisodeUpload(ctx, videoID, payload.FilePath, payload.Filename)
	}
	if scrapeErr != nil {
		_ = p.repo.AppendVideoMetadata(ctx, videoID, "scrape_error", scrapeErr.Error())
		_ = p.repo.UpdateVideoStatus(ctx, videoID, "uploaded")
		p.logger.Error("scrape failed, continue to transcode", "video_id", videoID, "error", scrapeErr)
	}

	if err := p.enqueuer.EnqueueTranscode(TranscodePayload{
		VideoID:      videoID.String(),
		InputPath:    video.OriginalPath,
		TargetFormat: "mp4",
	}); err != nil {
		return err
	}
	return nil
}
