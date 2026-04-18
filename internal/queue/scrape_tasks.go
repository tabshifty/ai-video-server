package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"video-server/internal/services"
)

// ScrapePayload carries metadata discovery input for movie/episode/av uploads.
type ScrapePayload struct {
	VideoID  string `json:"video_id"`
	FilePath string `json:"file_path"`
	Filename string `json:"filename"`
	Type     string `json:"type"` // movie | episode | av
}

// RetagScrapePayload is used when admins change a short video to movie/episode/av and need async auto-scrape.
type RetagScrapePayload struct {
	VideoID       string `json:"video_id"`
	TargetType    string `json:"target_type"` // movie | episode | av
	SeasonNumber  int    `json:"season_number,omitempty"`
	EpisodeNumber int    `json:"episode_number,omitempty"`
}

func (e *Enqueuer) EnqueueScrapeMovie(payloadIn ScrapePayload) error {
	return e.enqueueScrapeTask(TypeScrapeMovie, payloadIn)
}

func (e *Enqueuer) EnqueueScrapeTV(payloadIn ScrapePayload) error {
	return e.enqueueScrapeTask(TypeScrapeTV, payloadIn)
}

func (e *Enqueuer) EnqueueScrapeAV(payloadIn ScrapePayload) error {
	return e.enqueueScrapeTask(TypeScrapeAV, payloadIn)
}

func (e *Enqueuer) EnqueueScrapeRetag(payloadIn RetagScrapePayload) error {
	payload, err := json.Marshal(payloadIn)
	if err != nil {
		return fmt.Errorf("marshal retag scrape payload: %w", err)
	}
	_, err = e.client.Enqueue(
		asynq.NewTask(TypeScrapeRetag, payload),
		asynq.MaxRetry(3),
		asynq.Queue(e.queue),
	)
	if err != nil {
		return fmt.Errorf("enqueue retag scrape task: %w", err)
	}
	return nil
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

func (p *Processor) HandleScrapeAV(ctx context.Context, task *asynq.Task) error {
	return p.handleScrape(ctx, task, "av")
}

func (p *Processor) HandleScrapeRetag(ctx context.Context, task *asynq.Task) error {
	if p.scrape == nil {
		return fmt.Errorf("scrape processor not configured")
	}
	var payload RetagScrapePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal retag scrape payload: %w", err)
	}
	videoID, err := uuid.Parse(payload.VideoID)
	if err != nil {
		return fmt.Errorf("invalid video id: %w", err)
	}
	targetType := strings.ToLower(strings.TrimSpace(payload.TargetType))
	if targetType != "movie" && targetType != "episode" && targetType != "av" {
		return fmt.Errorf("invalid target_type: %s", payload.TargetType)
	}
	video, err := p.repo.GetVideoByID(ctx, videoID)
	if err != nil {
		return err
	}
	if video.Type != targetType {
		return fmt.Errorf("video type mismatch for retag scrape: want=%s got=%s", targetType, video.Type)
	}

	searchTitle := strings.TrimSpace(video.Title)
	if searchTitle == "" {
		name := filepath.Base(strings.TrimSpace(video.OriginalPath))
		searchTitle = strings.TrimSpace(strings.TrimSuffix(name, filepath.Ext(name)))
	}
	if searchTitle == "" {
		searchTitle = video.ID.String()
	}

	var scrapeErr error
	switch targetType {
	case "movie":
		scrapeErr = p.autoScrapeMovie(ctx, videoID, searchTitle)
	case "episode":
		if payload.SeasonNumber <= 0 || payload.EpisodeNumber <= 0 {
			scrapeErr = fmt.Errorf("season_number and episode_number are required for episode retag scrape")
			break
		}
		scrapeErr = p.autoScrapeEpisode(ctx, videoID, searchTitle, payload.SeasonNumber, payload.EpisodeNumber)
	case "av":
		scrapeErr = p.autoScrapeAV(ctx, videoID, searchTitle)
	}
	if scrapeErr != nil {
		_ = p.repo.AppendVideoMetadata(ctx, videoID, "scrape_error", scrapeErr.Error())
		_ = p.repo.UpdateVideoStatus(ctx, videoID, "uploaded")
		p.logger.Error("retag scrape failed", "video_id", videoID, "target_type", targetType, "error", scrapeErr)
		return scrapeErr
	}
	if err := p.repo.UpdateVideoStatus(ctx, videoID, "ready"); err != nil {
		return err
	}
	p.logger.Info("retag scrape completed", "video_id", videoID, "target_type", targetType)
	return nil
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
	} else if expectedType == "av" {
		_, scrapeErr = p.scrape.ScrapeAVUpload(ctx, videoID, payload.FilePath, payload.Filename)
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

func (p *Processor) autoScrapeMovie(ctx context.Context, videoID uuid.UUID, title string) error {
	candidates, err := p.scrape.PreviewMovie(ctx, title, 0)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return fmt.Errorf("no movie candidate for %q", title)
	}
	first := candidates[0]
	input := services.ConfirmScrapeInput{
		VideoID:     videoID,
		TMDBID:      anyInt(first["tmdb_id"]),
		Title:       anyString(first["title"]),
		Overview:    anyString(first["overview"]),
		PosterURL:   firstNonEmpty(anyString(first["poster_url"]), anyString(first["poster_path"])),
		ReleaseDate: anyString(first["release_date"]),
		Metadata:    anyMap(first["metadata"]),
	}
	if input.TMDBID <= 0 {
		return fmt.Errorf("invalid movie tmdb_id from first candidate")
	}
	return p.scrape.ConfirmMovie(ctx, input)
}

func (p *Processor) autoScrapeEpisode(ctx context.Context, videoID uuid.UUID, title string, seasonNumber, episodeNumber int) error {
	candidates, err := p.scrape.PreviewTV(ctx, title, 0)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return fmt.Errorf("no tv candidate for %q", title)
	}
	first := candidates[0]
	input := services.ConfirmScrapeInput{
		VideoID:       videoID,
		TMDBID:        anyInt(first["tmdb_id"]),
		Title:         anyString(first["title"]),
		Overview:      anyString(first["overview"]),
		PosterURL:     firstNonEmpty(anyString(first["poster_url"]), anyString(first["poster_path"])),
		ReleaseDate:   anyString(first["release_date"]),
		Metadata:      anyMap(first["metadata"]),
		SeasonNumber:  seasonNumber,
		EpisodeNumber: episodeNumber,
	}
	if input.TMDBID <= 0 {
		return fmt.Errorf("invalid tv tmdb_id from first candidate")
	}
	return p.scrape.ConfirmEpisode(ctx, input)
}

func (p *Processor) autoScrapeAV(ctx context.Context, videoID uuid.UUID, title string) error {
	candidates, err := p.scrape.PreviewAV(ctx, title)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return fmt.Errorf("no av candidate for %q", title)
	}
	first := candidates[0]
	meta := anyMap(first["metadata"])
	if meta == nil {
		meta = map[string]any{}
	}
	if anyString(meta["detail_url"]) == "" {
		meta["detail_url"] = anyString(first["detail_url"])
	}
	if anyString(meta["scrape_source"]) == "" {
		meta["scrape_source"] = anyString(first["scrape_source"])
	}
	input := services.ConfirmScrapeInput{
		VideoID:     videoID,
		ExternalID:  anyString(first["external_id"]),
		Title:       anyString(first["title"]),
		Overview:    anyString(first["overview"]),
		PosterURL:   firstNonEmpty(anyString(first["poster_url"]), anyString(first["poster_path"])),
		ReleaseDate: anyString(first["release_date"]),
		Metadata:    meta,
	}
	if input.ExternalID == "" {
		return fmt.Errorf("invalid av external_id from first candidate")
	}
	return p.scrape.ConfirmAV(ctx, input)
}

func anyString(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func anyInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int8:
		return int(n)
	case int16:
		return int(n)
	case int32:
		return int(n)
	case int64:
		return int(n)
	case float32:
		return int(n)
	case float64:
		return int(n)
	case json.Number:
		i64, _ := n.Int64()
		return int(i64)
	default:
		return 0
	}
}

func anyMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
