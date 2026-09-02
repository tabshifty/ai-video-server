package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/telegram"
)

type telegramIngestionFakeRepository struct {
	sources            map[uuid.UUID]models.TelegramSource
	media              map[uuid.UUID]models.TelegramMedia
	videosByHash       map[string]uuid.UUID
	videos             map[uuid.UUID]models.Video
	insertCount        int
	advanceCursorCalls []int64
	mediaPatches       map[uuid.UUID][]repository.TelegramMediaPatch
	sourcePatches      map[uuid.UUID][]repository.TelegramSourcePatch
	claimAllowed       []string
	claimCalls         int
	downloadRecovery   []models.TelegramMedia
}

func newTelegramIngestionFakeRepository() *telegramIngestionFakeRepository {
	return &telegramIngestionFakeRepository{
		sources:       make(map[uuid.UUID]models.TelegramSource),
		media:         make(map[uuid.UUID]models.TelegramMedia),
		videosByHash:  make(map[string]uuid.UUID),
		videos:        make(map[uuid.UUID]models.Video),
		mediaPatches:  make(map[uuid.UUID][]repository.TelegramMediaPatch),
		sourcePatches: make(map[uuid.UUID][]repository.TelegramSourcePatch),
	}
}

func (r *telegramIngestionFakeRepository) GetTelegramSource(_ context.Context, sourceID uuid.UUID) (models.TelegramSource, error) {
	source, ok := r.sources[sourceID]
	if !ok {
		return models.TelegramSource{}, repository.ErrTelegramSourceNotFound
	}
	return source, nil
}

func (r *telegramIngestionFakeRepository) UpdateTelegramSource(_ context.Context, sourceID uuid.UUID, patch repository.TelegramSourcePatch) error {
	source, ok := r.sources[sourceID]
	if !ok {
		return repository.ErrTelegramSourceNotFound
	}
	r.sourcePatches[sourceID] = append(r.sourcePatches[sourceID], patch)
	if patch.ChatID != nil {
		source.ChatID = *patch.ChatID
	}
	if patch.Title != nil {
		source.Title = *patch.Title
	}
	if patch.Username != nil {
		source.Username = *patch.Username
	}
	if patch.Enabled != nil {
		source.Enabled = *patch.Enabled
	}
	if patch.SyncStatus != nil {
		source.SyncStatus = *patch.SyncStatus
	}
	if patch.HistoryCursorMessageID != nil {
		source.HistoryCursorMessageID = *patch.HistoryCursorMessageID
	}
	if patch.LastError != nil {
		source.LastError = *patch.LastError
	}
	if patch.ClearNextRetryAt {
		source.NextRetryAt = nil
	} else if patch.NextRetryAt != nil {
		source.NextRetryAt = patch.NextRetryAt
	}
	if patch.ClearBackfillCompletedAt {
		source.BackfillCompletedAt = nil
	} else if patch.BackfillCompletedAt != nil {
		source.BackfillCompletedAt = patch.BackfillCompletedAt
	}
	r.sources[sourceID] = source
	return nil
}

func (r *telegramIngestionFakeRepository) InsertTelegramMedia(_ context.Context, media models.TelegramMedia) (models.TelegramMedia, bool, error) {
	for _, stored := range r.media {
		if stored.SourceID == media.SourceID && stored.MessageID == media.MessageID {
			return stored, false, nil
		}
	}
	if media.ID == uuid.Nil {
		media.ID = uuid.New()
	}
	if media.ProcessingStatus == "" {
		media.ProcessingStatus = "discovered"
	}
	if media.TranscodeStatus == "" {
		media.TranscodeStatus = "not_required"
	}
	if media.CreatedAt.IsZero() {
		media.CreatedAt = time.Now().UTC()
	}
	if media.UpdatedAt.IsZero() {
		media.UpdatedAt = media.CreatedAt
	}
	r.media[media.ID] = media
	r.insertCount++
	return media, true, nil
}

func (r *telegramIngestionFakeRepository) InsertTelegramMediaAndAdvanceCursor(ctx context.Context, media models.TelegramMedia, cursor int64) (models.TelegramMedia, bool, error) {
	stored, inserted, err := r.InsertTelegramMedia(ctx, media)
	if err != nil {
		return models.TelegramMedia{}, false, err
	}
	if err := r.advanceCursor(media.SourceID, cursor); err != nil {
		return models.TelegramMedia{}, false, err
	}
	return stored, inserted, nil
}

func (r *telegramIngestionFakeRepository) advanceCursor(sourceID uuid.UUID, cursor int64) error {
	source, ok := r.sources[sourceID]
	if !ok {
		return repository.ErrTelegramSourceNotFound
	}
	r.advanceCursorCalls = append(r.advanceCursorCalls, cursor)
	if source.HistoryCursorMessageID == 0 || cursor < source.HistoryCursorMessageID {
		source.HistoryCursorMessageID = cursor
	}
	r.sources[sourceID] = source
	return nil
}

func (r *telegramIngestionFakeRepository) AdvanceTelegramSourceCursor(ctx context.Context, sourceID uuid.UUID, cursor int64) error {
	return r.advanceCursor(sourceID, cursor)
}

func (r *telegramIngestionFakeRepository) ListTelegramSources(_ context.Context, enabledOnly bool) ([]models.TelegramSource, error) {
	items := make([]models.TelegramSource, 0, len(r.sources))
	for _, source := range r.sources {
		if enabledOnly && !source.Enabled {
			continue
		}
		items = append(items, source)
	}
	return items, nil
}

func (r *telegramIngestionFakeRepository) GetTelegramMedia(_ context.Context, mediaID uuid.UUID) (models.TelegramMedia, error) {
	media, ok := r.media[mediaID]
	if !ok {
		return models.TelegramMedia{}, repository.ErrTelegramMediaNotFound
	}
	return media, nil
}

func (r *telegramIngestionFakeRepository) FindTelegramMediaByDocumentID(_ context.Context, documentID int64) (models.TelegramMedia, bool, error) {
	for _, media := range r.media {
		if media.DocumentID != nil && *media.DocumentID == documentID && media.VideoID != nil && (media.ProcessingStatus == "imported" || media.ProcessingStatus == "duplicate") {
			return media, true, nil
		}
	}
	return models.TelegramMedia{}, false, nil
}

func (r *telegramIngestionFakeRepository) ClaimTelegramMedia(_ context.Context, mediaID uuid.UUID, allowed []string) (models.TelegramMedia, bool, error) {
	r.claimCalls++
	media, ok := r.media[mediaID]
	if !ok {
		return models.TelegramMedia{}, false, repository.ErrTelegramMediaNotFound
	}
	for _, status := range allowed {
		if status == media.ProcessingStatus {
			media.Attempts++
			media.ProcessingStatus = "downloading"
			r.media[mediaID] = media
			r.claimAllowed = append([]string(nil), allowed...)
			return media, true, nil
		}
	}
	return models.TelegramMedia{}, false, nil
}

func (r *telegramIngestionFakeRepository) UpdateTelegramMedia(_ context.Context, mediaID uuid.UUID, patch repository.TelegramMediaPatch) error {
	media, ok := r.media[mediaID]
	if !ok {
		return repository.ErrTelegramMediaNotFound
	}
	r.mediaPatches[mediaID] = append(r.mediaPatches[mediaID], patch)
	if patch.ClearDocumentID {
		media.DocumentID = nil
	} else if patch.DocumentID != nil {
		media.DocumentID = patch.DocumentID
	}
	if patch.DocumentDCID != nil {
		media.DocumentDCID = patch.DocumentDCID
	}
	if patch.ProcessingStatus != nil {
		media.ProcessingStatus = *patch.ProcessingStatus
	}
	if patch.TranscodeStatus != nil {
		media.TranscodeStatus = *patch.TranscodeStatus
	}
	if patch.VideoID != nil {
		media.VideoID = patch.VideoID
	}
	if patch.ClearVideoID {
		media.VideoID = nil
	}
	if patch.SHA256 != nil {
		media.SHA256 = *patch.SHA256
	}
	if patch.TempPath != nil {
		media.TempPath = *patch.TempPath
	}
	if patch.Attempts != nil {
		media.Attempts = *patch.Attempts
	}
	if patch.NextRetryAt != nil {
		media.NextRetryAt = patch.NextRetryAt
	}
	if patch.ClearNextRetryAt {
		media.NextRetryAt = nil
	}
	if patch.LastError != nil {
		media.LastError = *patch.LastError
	}
	r.media[mediaID] = media
	return nil
}

func (r *telegramIngestionFakeRepository) FindVideoByHash(_ context.Context, hash string, size int64) (uuid.UUID, bool, error) {
	videoID, ok := r.videosByHash[fmt.Sprintf("%s:%d", hash, size)]
	return videoID, ok, nil
}

func (r *telegramIngestionFakeRepository) GetVideoByID(_ context.Context, videoID uuid.UUID) (models.Video, error) {
	video, ok := r.videos[videoID]
	if !ok {
		return models.Video{}, errors.New("video not found")
	}
	return video, nil
}

func (r *telegramIngestionFakeRepository) ListTelegramMediaNeedingTranscode(_ context.Context, _ int) ([]models.TelegramMedia, error) {
	items := make([]models.TelegramMedia, 0)
	for _, media := range r.media {
		if media.TranscodeStatus == "pending" && media.VideoID != nil {
			items = append(items, media)
		}
	}
	return items, nil
}

func (r *telegramIngestionFakeRepository) ListTelegramMediaNeedingDownload(_ context.Context, _ time.Time, _ int) ([]models.TelegramMedia, error) {
	return append([]models.TelegramMedia(nil), r.downloadRecovery...), nil
}

type telegramIngestionFakeClient struct {
	chat          telegram.Chat
	refreshed     map[string]telegram.Message
	historyPages  map[int64][]telegram.Message
	historyCalls  []int64
	historyHook   func()
	downloads     int
	downloadData  []byte
	refreshErrors error
	downloadError error
}

func (c *telegramIngestionFakeClient) ResolveChat(context.Context, string) (telegram.Chat, error) {
	return c.chat, nil
}

func (c *telegramIngestionFakeClient) History(_ context.Context, _, offsetID int64, _ int, fn func(telegram.Message) error) error {
	c.historyCalls = append(c.historyCalls, offsetID)
	for _, message := range c.historyPages[offsetID] {
		if err := fn(message); err != nil {
			return err
		}
	}
	if c.historyHook != nil {
		c.historyHook()
	}
	return nil
}

func (c *telegramIngestionFakeClient) Subscribe(context.Context, func(telegram.Message) error) error {
	return nil
}

func (c *telegramIngestionFakeClient) RefreshMessage(_ context.Context, chatID, messageID int64) (telegram.Message, error) {
	if c.refreshErrors != nil {
		return telegram.Message{}, c.refreshErrors
	}
	if message, ok := c.refreshed[telegramMessageKey(chatID, messageID)]; ok {
		return message, nil
	}
	return telegram.Message{ChatID: chatID, MessageID: messageID}, nil
}

func (c *telegramIngestionFakeClient) Download(_ context.Context, _ telegram.Message, dst io.Writer) error {
	c.downloads++
	if c.downloadError != nil {
		return c.downloadError
	}
	_, err := io.Copy(dst, bytes.NewReader(c.downloadData))
	return err
}

type telegramIngestionFakeUpload struct {
	inputs []LocalUploadInput
	result UploadResult
	err    error
}

func (u *telegramIngestionFakeUpload) SaveImportedFile(_ context.Context, input LocalUploadInput, _ int64) (UploadResult, error) {
	u.inputs = append(u.inputs, input)
	return u.result, u.err
}

type telegramIngestionFakeTasks struct {
	realtime []uuid.UUID
	backfill []uuid.UUID
	err      error
}

func (q *telegramIngestionFakeTasks) EnqueueRealtime(mediaID uuid.UUID) error {
	if q.err != nil {
		return q.err
	}
	q.realtime = append(q.realtime, mediaID)
	return nil
}

func (q *telegramIngestionFakeTasks) EnqueueBackfill(mediaID uuid.UUID) error {
	if q.err != nil {
		return q.err
	}
	q.backfill = append(q.backfill, mediaID)
	return nil
}

type telegramIngestionFakeTranscode struct {
	payloads []telegramTranscodeRequest
	err      error
}

type telegramTranscodeRequest struct {
	videoID      string
	inputPath    string
	outputDir    string
	targetFormat string
}

func (q *telegramIngestionFakeTranscode) EnqueueTranscode(videoID, inputPath, outputDir, targetFormat string, _ bool) error {
	q.payloads = append(q.payloads, telegramTranscodeRequest{
		videoID:      videoID,
		inputPath:    inputPath,
		outputDir:    outputDir,
		targetFormat: targetFormat,
	})
	return q.err
}

func telegramMessageKey(chatID, messageID int64) string {
	return fmt.Sprintf("%d:%d", chatID, messageID)
}

func telegramTestMessage(chatID, messageID, documentID int64, caption string) telegram.Message {
	return telegram.Message{
		ChatID:     chatID,
		MessageID:  messageID,
		DocumentID: documentID,
		Filename:   "clip.mp4",
		MIMEType:   "video/mp4",
		Size:       5,
		Caption:    caption,
		MessageURL: fmt.Sprintf("https://t.me/c/123/%d", messageID),
		SentAt:     time.Date(2026, 9, 2, 8, 0, int(messageID), 0, time.UTC),
	}
}

func telegramTestService(t *testing.T, repo *telegramIngestionFakeRepository, client *telegramIngestionFakeClient, upload *telegramIngestionFakeUpload, tasks *telegramIngestionFakeTasks, transcode *telegramIngestionFakeTranscode) *TelegramIngestionService {
	t.Helper()
	return NewTelegramIngestionService(
		client,
		repo,
		upload,
		tasks,
		transcode,
		t.TempDir(),
		t.TempDir(),
		1024,
		uuid.New(),
	)
}

func telegramTimePtr(value time.Time) *time.Time {
	return &value
}

func seedTelegramSource(repo *telegramIngestionFakeRepository) uuid.UUID {
	sourceID := uuid.New()
	repo.sources[sourceID] = models.TelegramSource{
		ID:         sourceID,
		ChatID:     -100123,
		ChatRef:    "@test-group",
		Title:      "测试群组",
		Enabled:    true,
		SyncStatus: "pending",
	}
	return sourceID
}

func TestTelegramIngestionDuplicateMessageDoesNotEnqueueTwice(t *testing.T) {
	repo := newTelegramIngestionFakeRepository()
	sourceID := seedTelegramSource(repo)
	client := &telegramIngestionFakeClient{}
	tasks := &telegramIngestionFakeTasks{}
	service := telegramTestService(t, repo, client, &telegramIngestionFakeUpload{}, tasks, &telegramIngestionFakeTranscode{})
	message := telegramTestMessage(-100123, 42, 777, "caption")

	if err := service.HandleMessage(context.Background(), sourceID, message, MessagePriorityRealtime); err != nil {
		t.Fatalf("first HandleMessage() error = %v", err)
	}
	if err := service.HandleMessage(context.Background(), sourceID, message, MessagePriorityRealtime); err != nil {
		t.Fatalf("duplicate HandleMessage() error = %v", err)
	}
	if repo.insertCount != 1 {
		t.Fatalf("insert count = %d, want 1", repo.insertCount)
	}
	if len(tasks.realtime) != 1 {
		t.Fatalf("realtime task count = %d, want 1", len(tasks.realtime))
	}
}

func TestTelegramIngestionDocumentIDDuplicateSkipsDownload(t *testing.T) {
	repo := newTelegramIngestionFakeRepository()
	sourceID := seedTelegramSource(repo)
	videoID := uuid.New()
	first := telegramTestMessage(-100123, 1, 888, "first")
	firstMedia := models.TelegramMedia{ID: uuid.New(), SourceID: sourceID, ChatID: first.ChatID, MessageID: first.MessageID, DocumentID: &first.DocumentID, ProcessingStatus: "imported", VideoID: &videoID}
	repo.media[firstMedia.ID] = firstMedia
	second := telegramTestMessage(-100123, 2, 888, "second")
	secondMedia := models.TelegramMedia{ID: uuid.New(), SourceID: sourceID, ChatID: second.ChatID, MessageID: second.MessageID, DocumentID: &second.DocumentID, ProcessingStatus: "queued"}
	repo.media[secondMedia.ID] = secondMedia
	client := &telegramIngestionFakeClient{refreshed: map[string]telegram.Message{telegramMessageKey(second.ChatID, second.MessageID): second}}
	service := telegramTestService(t, repo, client, &telegramIngestionFakeUpload{}, &telegramIngestionFakeTasks{}, &telegramIngestionFakeTranscode{})

	if err := service.ProcessMedia(context.Background(), secondMedia.ID); err != nil {
		t.Fatalf("ProcessMedia() error = %v", err)
	}
	stored := repo.media[secondMedia.ID]
	if client.downloads != 0 {
		t.Fatalf("download count = %d, want 0", client.downloads)
	}
	if stored.ProcessingStatus != "duplicate" || stored.VideoID == nil || *stored.VideoID != videoID {
		t.Fatalf("stored duplicate = %+v", stored)
	}
}

func TestTelegramIngestionPausedSourceLeavesMediaQueued(t *testing.T) {
	repo := newTelegramIngestionFakeRepository()
	sourceID := seedTelegramSource(repo)
	source := repo.sources[sourceID]
	source.SyncStatus = "paused"
	repo.sources[sourceID] = source
	mediaID := uuid.New()
	repo.media[mediaID] = models.TelegramMedia{
		ID:               mediaID,
		SourceID:         sourceID,
		ChatID:           source.ChatID,
		MessageID:        12,
		ProcessingStatus: "queued",
		TranscodeStatus:  "not_required",
	}
	service := telegramTestService(t, repo, &telegramIngestionFakeClient{}, &telegramIngestionFakeUpload{}, &telegramIngestionFakeTasks{}, &telegramIngestionFakeTranscode{})

	if err := service.ProcessMedia(context.Background(), mediaID); err != nil {
		t.Fatalf("ProcessMedia() error = %v", err)
	}
	if repo.claimCalls != 0 {
		t.Fatalf("claim calls = %d, want 0 for paused source", repo.claimCalls)
	}
	if got := repo.media[mediaID].ProcessingStatus; got != "queued" {
		t.Fatalf("processing status = %q, want queued", got)
	}
}

func TestTelegramIngestionReconcileDownloadsRequeuesStaleMedia(t *testing.T) {
	repo := newTelegramIngestionFakeRepository()
	sourceID := seedTelegramSource(repo)
	mediaID := uuid.New()
	media := models.TelegramMedia{
		ID:               mediaID,
		SourceID:         sourceID,
		ChatID:           -100123,
		MessageID:        13,
		ProcessingStatus: "downloading",
		TranscodeStatus:  "not_required",
		NextRetryAt:      telegramTimePtr(time.Now().Add(-time.Minute)),
	}
	repo.media[mediaID] = media
	repo.downloadRecovery = []models.TelegramMedia{media}
	tasks := &telegramIngestionFakeTasks{}
	service := telegramTestService(t, repo, &telegramIngestionFakeClient{}, &telegramIngestionFakeUpload{}, tasks, &telegramIngestionFakeTranscode{})

	if err := service.ReconcileDownloads(context.Background(), 10); err != nil {
		t.Fatalf("ReconcileDownloads() error = %v", err)
	}
	if len(tasks.backfill) != 1 || tasks.backfill[0] != mediaID {
		t.Fatalf("backfill tasks = %v, want [%s]", tasks.backfill, mediaID)
	}
	stored := repo.media[mediaID]
	if stored.ProcessingStatus != "queued" || stored.NextRetryAt != nil {
		t.Fatalf("requeued media = %+v", stored)
	}
}

func TestTelegramIngestionHashDuplicateAssociatesExistingVideo(t *testing.T) {
	repo := newTelegramIngestionFakeRepository()
	sourceID := seedTelegramSource(repo)
	videoID := uuid.New()
	content := []byte("video")
	message := telegramTestMessage(-100123, 3, 999, "hash duplicate")
	media := models.TelegramMedia{ID: uuid.New(), SourceID: sourceID, ChatID: message.ChatID, MessageID: message.MessageID, DocumentID: &message.DocumentID, ProcessingStatus: "queued"}
	repo.media[media.ID] = media
	repo.videosByHash[fmt.Sprintf("%s:%d", "0cab1c9617404faf2b24e221e189ca5945813e14d3f766345b09ca13bbe28ffc", len(content))] = videoID
	client := &telegramIngestionFakeClient{downloadData: content, refreshed: map[string]telegram.Message{telegramMessageKey(message.ChatID, message.MessageID): message}}
	service := telegramTestService(t, repo, client, &telegramIngestionFakeUpload{}, &telegramIngestionFakeTasks{}, &telegramIngestionFakeTranscode{})

	if err := service.ProcessMedia(context.Background(), media.ID); err != nil {
		t.Fatalf("ProcessMedia() error = %v", err)
	}
	stored := repo.media[media.ID]
	if client.downloads != 1 {
		t.Fatalf("download count = %d, want 1", client.downloads)
	}
	if stored.ProcessingStatus != "duplicate" || stored.VideoID == nil || *stored.VideoID != videoID {
		t.Fatalf("stored hash duplicate = %+v", stored)
	}
}

func TestTelegramIngestionNewVideoUsesShortMetadataAndReconcilesTranscode(t *testing.T) {
	repo := newTelegramIngestionFakeRepository()
	sourceID := seedTelegramSource(repo)
	message := telegramTestMessage(-100123, 4, 1000, "视频说明")
	media := models.TelegramMedia{ID: uuid.New(), SourceID: sourceID, ChatID: message.ChatID, MessageID: message.MessageID, DocumentID: &message.DocumentID, ProcessingStatus: "queued"}
	repo.media[media.ID] = media
	videoID := uuid.New()
	upload := &telegramIngestionFakeUpload{result: UploadResult{VideoID: videoID, Status: "uploaded", Enqueue: true, InputPath: "/tmp/import.mp4", OutputDir: "/storage/videos", TargetFormat: "mp4"}}
	transcode := &telegramIngestionFakeTranscode{err: errors.New("redis unavailable")}
	client := &telegramIngestionFakeClient{downloadData: []byte("video"), refreshed: map[string]telegram.Message{telegramMessageKey(message.ChatID, message.MessageID): message}}
	service := telegramTestService(t, repo, client, upload, &telegramIngestionFakeTasks{}, transcode)

	if err := service.ProcessMedia(context.Background(), media.ID); err == nil {
		t.Fatal("ProcessMedia() error = nil, want transcode enqueue error")
	}
	if len(upload.inputs) != 1 {
		t.Fatalf("SaveImportedFile call count = %d, want 1", len(upload.inputs))
	}
	input := upload.inputs[0]
	if input.Type != "short" || input.UserID == uuid.Nil {
		t.Fatalf("unexpected import input = %+v", input)
	}
	if input.Title != "视频说明" || !strings.Contains(input.Desc, "视频说明") {
		t.Fatalf("metadata title/description missing caption: %+v", input)
	}
	if input.Metadata["document_id"] != int64(1000) || input.Metadata["chat_id"] != int64(-100123) {
		t.Fatalf("unexpected Telegram metadata: %#v", input.Metadata)
	}
	stored := repo.media[media.ID]
	if stored.TranscodeStatus != "pending" || stored.VideoID == nil || *stored.VideoID != videoID {
		t.Fatalf("transcode failure state = %+v", stored)
	}

	transcode.err = nil
	if err := service.ReconcileTranscodes(context.Background(), 10); err != nil {
		t.Fatalf("ReconcileTranscodes() error = %v", err)
	}
	if len(transcode.payloads) != 2 {
		t.Fatalf("transcode enqueue count = %d, want failed attempt plus reconciliation", len(transcode.payloads))
	}
	if repo.media[media.ID].TranscodeStatus != "enqueued" {
		t.Fatalf("transcode status = %q, want enqueued", repo.media[media.ID].TranscodeStatus)
	}
}

func TestTelegramIngestionFloodWaitPersistsRetryTime(t *testing.T) {
	repo := newTelegramIngestionFakeRepository()
	sourceID := seedTelegramSource(repo)
	message := telegramTestMessage(-100123, 5, 1001, "flood")
	media := models.TelegramMedia{ID: uuid.New(), SourceID: sourceID, ChatID: message.ChatID, MessageID: message.MessageID, ProcessingStatus: "queued"}
	repo.media[media.ID] = media
	client := &telegramIngestionFakeClient{
		refreshed:     map[string]telegram.Message{telegramMessageKey(message.ChatID, message.MessageID): message},
		downloadError: &telegram.FloodWaitError{Seconds: 37, Err: errors.New("FLOOD_WAIT_37")},
	}
	service := telegramTestService(t, repo, client, &telegramIngestionFakeUpload{}, &telegramIngestionFakeTasks{}, &telegramIngestionFakeTranscode{})

	if err := service.ProcessMedia(context.Background(), media.ID); err == nil {
		t.Fatal("ProcessMedia() error = nil, want FloodWait error")
	}
	stored := repo.media[media.ID]
	if stored.NextRetryAt == nil || stored.ProcessingStatus != "failed" {
		t.Fatalf("FloodWait state = %+v", stored)
	}
	if delay := time.Until(*stored.NextRetryAt); delay < 30*time.Second || delay > 40*time.Second {
		t.Fatalf("FloodWait retry delay = %v, want about 37s", delay)
	}
}

func TestTelegramIngestionHistoryCursorAllowsRepeatedScanWithoutMissingMessages(t *testing.T) {
	repo := newTelegramIngestionFakeRepository()
	sourceID := seedTelegramSource(repo)
	m1 := telegramTestMessage(-100123, 1, 2001, "one")
	m2 := telegramTestMessage(-100123, 2, 2002, "two")
	m3 := telegramTestMessage(-100123, 3, 2003, "three")
	client := &telegramIngestionFakeClient{
		chat: telegram.Chat{ID: -100123, Title: "测试群组"},
		historyPages: map[int64][]telegram.Message{
			0: {m3, m2},
			2: {m1},
			1: nil,
		},
	}
	service := telegramTestService(t, repo, client, &telegramIngestionFakeUpload{}, &telegramIngestionFakeTasks{}, &telegramIngestionFakeTranscode{})

	if err := service.SyncSource(context.Background(), sourceID); err != nil {
		t.Fatalf("first SyncSource() error = %v", err)
	}
	if repo.sources[sourceID].HistoryCursorMessageID != 1 {
		t.Fatalf("history cursor = %d, want 1", repo.sources[sourceID].HistoryCursorMessageID)
	}
	if err := service.SyncSource(context.Background(), sourceID); err != nil {
		t.Fatalf("repeated SyncSource() error = %v", err)
	}
	if repo.insertCount != 3 {
		t.Fatalf("insert count after repeated scan = %d, want 3", repo.insertCount)
	}
	if repo.sources[sourceID].SyncStatus != "live" {
		t.Fatalf("sync status = %q, want live", repo.sources[sourceID].SyncStatus)
	}
}

func TestTelegramIngestionStopsBackfillWhenSourceIsPaused(t *testing.T) {
	repo := newTelegramIngestionFakeRepository()
	sourceID := seedTelegramSource(repo)
	client := &telegramIngestionFakeClient{
		chat: telegram.Chat{ID: -100123, Title: "测试群组"},
		historyPages: map[int64][]telegram.Message{
			0: {telegramTestMessage(-100123, 2, 4002, "two")},
			2: {telegramTestMessage(-100123, 1, 4001, "one")},
		},
	}
	client.historyHook = func() {
		source := repo.sources[sourceID]
		source.Enabled = false
		source.SyncStatus = "paused"
		repo.sources[sourceID] = source
		client.historyHook = nil
	}
	service := telegramTestService(t, repo, client, &telegramIngestionFakeUpload{}, &telegramIngestionFakeTasks{}, &telegramIngestionFakeTranscode{})

	if err := service.SyncSource(context.Background(), sourceID); err != nil {
		t.Fatalf("SyncSource() error = %v", err)
	}
	if got := repo.sources[sourceID].SyncStatus; got != "paused" {
		t.Fatalf("sync status = %q, want paused", got)
	}
	if len(client.historyCalls) != 1 {
		t.Fatalf("history calls = %v, want one page before pause", client.historyCalls)
	}
}

func TestTelegramIngestionKeepsDownloadedFileWhenImportFails(t *testing.T) {
	repo := newTelegramIngestionFakeRepository()
	sourceID := seedTelegramSource(repo)
	message := telegramTestMessage(-100123, 6, 3001, "retry import")
	media := models.TelegramMedia{ID: uuid.New(), SourceID: sourceID, ChatID: message.ChatID, MessageID: message.MessageID, ProcessingStatus: "queued"}
	repo.media[media.ID] = media
	client := &telegramIngestionFakeClient{downloadData: []byte("video"), refreshed: map[string]telegram.Message{telegramMessageKey(message.ChatID, message.MessageID): message}}
	upload := &telegramIngestionFakeUpload{err: errors.New("import failed")}
	uploadDir := t.TempDir()
	service := NewTelegramIngestionService(client, repo, upload, &telegramIngestionFakeTasks{}, &telegramIngestionFakeTranscode{}, uploadDir, t.TempDir(), 1024, uuid.New())

	if err := service.ProcessMedia(context.Background(), media.ID); err == nil {
		t.Fatal("ProcessMedia() error = nil, want import error")
	}
	stored := repo.media[media.ID]
	if stored.TempPath == "" || stored.SHA256 == "" || stored.ProcessingStatus != "failed" {
		t.Fatalf("import failure state = %+v", stored)
	}
	if _, err := os.Stat(stored.TempPath); err != nil {
		t.Fatalf("downloaded temp file should be retained: %v", err)
	}
	if filepath.Dir(stored.TempPath) != filepath.Clean(uploadDir) {
		t.Fatalf("temp path = %q, want under %q", stored.TempPath, uploadDir)
	}
}
