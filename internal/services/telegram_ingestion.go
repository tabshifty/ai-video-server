package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/telegram"
)

const (
	telegramHistoryPageSize        = 100
	telegramDownloadRecoveryWindow = 30 * time.Minute
)

// MessagePriority controls which download queue receives a discovered message.
type MessagePriority string

const (
	// MessagePriorityRealtime is used for Telegram update events.
	MessagePriorityRealtime MessagePriority = "realtime"
	// MessagePriorityBackfill is used for historical messages.
	MessagePriorityBackfill MessagePriority = "backfill"
)

// TelegramIngestion is the task-facing boundary used by the Asynq processor.
type TelegramIngestion interface {
	SyncSource(ctx context.Context, sourceID uuid.UUID) error
	ProcessMedia(ctx context.Context, mediaID uuid.UUID) error
	ReconcileDownloads(ctx context.Context, limit int) error
	ReconcileTranscodes(ctx context.Context, limit int) error
}

type telegramIngestionRepository interface {
	GetTelegramSource(ctx context.Context, sourceID uuid.UUID) (models.TelegramSource, error)
	ListTelegramSources(ctx context.Context, enabledOnly bool) ([]models.TelegramSource, error)
	UpdateTelegramSource(ctx context.Context, sourceID uuid.UUID, patch repository.TelegramSourcePatch) error
	InsertTelegramMedia(ctx context.Context, media models.TelegramMedia) (models.TelegramMedia, bool, error)
	InsertTelegramMediaAndAdvanceCursor(ctx context.Context, media models.TelegramMedia, cursor int64) (models.TelegramMedia, bool, error)
	AdvanceTelegramSourceCursor(ctx context.Context, sourceID uuid.UUID, cursor int64) error
	GetTelegramMedia(ctx context.Context, mediaID uuid.UUID) (models.TelegramMedia, error)
	FindTelegramMediaByDocumentID(ctx context.Context, documentID int64) (models.TelegramMedia, bool, error)
	ClaimTelegramMedia(ctx context.Context, mediaID uuid.UUID, allowed []string) (models.TelegramMedia, bool, error)
	UpdateTelegramMedia(ctx context.Context, mediaID uuid.UUID, patch repository.TelegramMediaPatch) error
	ListTelegramMediaNeedingTranscode(ctx context.Context, limit int) ([]models.TelegramMedia, error)
	ListTelegramMediaNeedingDownload(ctx context.Context, staleBefore time.Time, limit int) ([]models.TelegramMedia, error)
	FindVideoByHash(ctx context.Context, hash string, fileSize int64) (uuid.UUID, bool, error)
}

type telegramImportedFileSaver interface {
	SaveImportedFile(ctx context.Context, in LocalUploadInput, maxVideoSize int64) (UploadResult, error)
}

type telegramTaskScheduler interface {
	EnqueueRealtime(mediaID uuid.UUID) error
	EnqueueBackfill(mediaID uuid.UUID) error
}

type telegramTranscodeScheduler interface {
	EnqueueTranscode(videoID, inputPath, outputDir, targetFormat string, force bool) error
}

// TelegramIngestionService coordinates Telegram discovery, local import, and
// the existing video transcode queue.
type TelegramIngestionService struct {
	client        telegram.Client
	repo          telegramIngestionRepository
	upload        telegramImportedFileSaver
	tasks         telegramTaskScheduler
	transcode     telegramTranscodeScheduler
	uploadTempDir string
	storageRoot   string
	maxVideoSize  int64
	importUserID  uuid.UUID
}

// NewTelegramIngestionService constructs a Telegram ingestion coordinator.
func NewTelegramIngestionService(
	client telegram.Client,
	repo telegramIngestionRepository,
	upload telegramImportedFileSaver,
	tasks telegramTaskScheduler,
	transcode telegramTranscodeScheduler,
	uploadTempDir string,
	storageRoot string,
	maxVideoSize int64,
	importUserID uuid.UUID,
) *TelegramIngestionService {
	return &TelegramIngestionService{
		client:        client,
		repo:          repo,
		upload:        upload,
		tasks:         tasks,
		transcode:     transcode,
		uploadTempDir: filepath.Clean(strings.TrimSpace(uploadTempDir)),
		storageRoot:   filepath.Clean(strings.TrimSpace(storageRoot)),
		maxVideoSize:  maxVideoSize,
		importUserID:  importUserID,
	}
}

// HandleMessage records one Telegram video message and schedules its download.
func (s *TelegramIngestionService) HandleMessage(ctx context.Context, sourceID uuid.UUID, message telegram.Message, priority MessagePriority) error {
	if err := s.validateDependencies(); err != nil {
		return err
	}
	source, err := s.repo.GetTelegramSource(ctx, sourceID)
	if err != nil {
		return fmt.Errorf("get Telegram source: %w", err)
	}
	if !source.Enabled || source.SyncStatus == "paused" {
		return nil
	}
	if err := validateTelegramMessage(message); err != nil {
		return err
	}
	if source.ChatID != 0 && source.ChatID != message.ChatID {
		return fmt.Errorf("Telegram 消息 chat ID %d 与来源 chat ID %d 不一致", message.ChatID, source.ChatID)
	}

	media, inserted, err := s.repo.InsertTelegramMedia(ctx, telegramMediaFromMessage(sourceID, message))
	if err != nil {
		return fmt.Errorf("register Telegram message: %w", err)
	}
	if !inserted {
		return nil
	}
	if err := s.enqueueMedia(ctx, media.ID, priority); err != nil {
		return err
	}
	return nil
}

// HandleRealtimeMessage routes an update to the configured source matching its chat.
func (s *TelegramIngestionService) HandleRealtimeMessage(ctx context.Context, message telegram.Message) error {
	if err := s.validateDependencies(); err != nil {
		return err
	}
	sources, err := s.repo.ListTelegramSources(ctx, true)
	if err != nil {
		return fmt.Errorf("list Telegram sources for update: %w", err)
	}
	for _, source := range sources {
		if source.ChatID != 0 && source.ChatID == message.ChatID && source.Enabled && source.SyncStatus != "paused" {
			return s.HandleMessage(ctx, source.ID, message, MessagePriorityRealtime)
		}
	}
	return nil
}

// SyncSource resolves a source and walks its history from the persisted cursor.
func (s *TelegramIngestionService) SyncSource(ctx context.Context, sourceID uuid.UUID) error {
	if err := s.validateDependencies(); err != nil {
		return err
	}
	source, err := s.repo.GetTelegramSource(ctx, sourceID)
	if err != nil {
		return fmt.Errorf("get Telegram source: %w", err)
	}
	if !source.Enabled || source.SyncStatus == "paused" {
		return nil
	}

	chat, err := s.client.ResolveChat(ctx, source.ChatRef)
	if err != nil {
		return s.recordSourceFailure(ctx, sourceID, fmt.Errorf("resolve Telegram source: %w", err))
	}
	if chat.ID == 0 {
		return s.recordSourceFailure(ctx, sourceID, errors.New("resolve Telegram source: empty chat ID"))
	}
	if err := s.repo.UpdateTelegramSource(ctx, sourceID, repository.TelegramSourcePatch{
		ChatID:                   int64Pointer(chat.ID),
		Title:                    stringPointer(chat.Title),
		Username:                 stringPointer(chat.Username),
		SyncStatus:               stringPointer("backfilling"),
		LastError:                stringPointer(""),
		ClearNextRetryAt:         true,
		ClearBackfillCompletedAt: true,
	}); err != nil {
		return fmt.Errorf("mark Telegram source backfilling: %w", err)
	}

	offsetID := source.HistoryCursorMessageID
	for {
		page, err := s.readHistoryPage(ctx, chat.ID, offsetID)
		if err != nil {
			return s.recordSourceFailure(ctx, sourceID, fmt.Errorf("read Telegram history: %w", err))
		}
		if len(page) == 0 {
			break
		}

		pageCursor := smallestTelegramMessageID(page)
		if pageCursor <= 0 {
			break
		}
		accepted := 0
		for _, message := range page {
			if message.ChatID == 0 {
				message.ChatID = chat.ID
			}
			if message.ChatID != chat.ID || !isTelegramVideoMessage(message) {
				continue
			}
			media, inserted, insertErr := s.repo.InsertTelegramMediaAndAdvanceCursor(ctx, telegramMediaFromMessage(sourceID, message), pageCursor)
			if insertErr != nil {
				return s.recordSourceFailure(ctx, sourceID, fmt.Errorf("register Telegram history message %d: %w", message.MessageID, insertErr))
			}
			accepted++
			if inserted {
				if enqueueErr := s.enqueueMedia(ctx, media.ID, MessagePriorityBackfill); enqueueErr != nil {
					return s.recordSourceFailure(ctx, sourceID, enqueueErr)
				}
			}
		}
		if accepted == 0 {
			if err := s.repo.AdvanceTelegramSourceCursor(ctx, sourceID, pageCursor); err != nil {
				return s.recordSourceFailure(ctx, sourceID, fmt.Errorf("advance Telegram history cursor: %w", err))
			}
		}
		if offsetID > 0 && pageCursor >= offsetID {
			break
		}
		offsetID = pageCursor
	}

	completedAt := time.Now().UTC()
	if err := s.repo.UpdateTelegramSource(ctx, sourceID, repository.TelegramSourcePatch{
		SyncStatus:          stringPointer("live"),
		LastError:           stringPointer(""),
		ClearNextRetryAt:    true,
		BackfillCompletedAt: &completedAt,
	}); err != nil {
		return fmt.Errorf("mark Telegram source live: %w", err)
	}
	return nil
}

func (s *TelegramIngestionService) readHistoryPage(ctx context.Context, chatID, offsetID int64) ([]telegram.Message, error) {
	page := make([]telegram.Message, 0, telegramHistoryPageSize)
	if err := s.client.History(ctx, chatID, offsetID, telegramHistoryPageSize, func(message telegram.Message) error {
		page = append(page, message)
		return nil
	}); err != nil {
		return nil, err
	}
	return page, nil
}

// ProcessMedia claims and imports one persisted Telegram message.
func (s *TelegramIngestionService) ProcessMedia(ctx context.Context, mediaID uuid.UUID) error {
	if err := s.validateDependencies(); err != nil {
		return err
	}
	media, err := s.repo.GetTelegramMedia(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("get Telegram media before claim: %w", err)
	}
	source, err := s.repo.GetTelegramSource(ctx, media.SourceID)
	if err != nil {
		return s.recordMediaFailure(ctx, media, fmt.Errorf("get Telegram source before claim: %w", err))
	}
	if !source.Enabled || source.SyncStatus == "paused" {
		return nil
	}
	media, claimed, err := s.repo.ClaimTelegramMedia(ctx, mediaID, []string{"discovered", "queued", "failed"})
	if err != nil {
		return fmt.Errorf("claim Telegram media: %w", err)
	}
	if !claimed {
		return nil
	}

	if media.DocumentID != nil && *media.DocumentID > 0 {
		if existing, found, findErr := s.repo.FindTelegramMediaByDocumentID(ctx, *media.DocumentID); findErr != nil {
			return s.recordMediaFailure(ctx, media, fmt.Errorf("find Telegram document duplicate: %w", findErr))
		} else if found && existing.VideoID != nil {
			return s.finishTelegramDuplicate(ctx, media, *existing.VideoID, media.SHA256, media.TempPath)
		}
	}

	source, err = s.repo.GetTelegramSource(ctx, media.SourceID)
	if err != nil {
		return s.recordMediaFailure(ctx, media, fmt.Errorf("get Telegram source for media: %w", err))
	}
	if !source.Enabled || source.SyncStatus == "paused" {
		return s.restorePausedMedia(ctx, media)
	}
	message, err := s.client.RefreshMessage(ctx, media.ChatID, media.MessageID)
	if err != nil {
		return s.recordMediaFailure(ctx, media, fmt.Errorf("refresh Telegram message: %w", err))
	}
	if message.ChatID == 0 {
		message.ChatID = media.ChatID
	}
	if message.ChatID != media.ChatID {
		return s.recordMediaFailure(ctx, media, fmt.Errorf("refreshed Telegram message chat ID %d differs from %d", message.ChatID, media.ChatID))
	}
	if err := validateTelegramMessage(message); err != nil {
		return s.recordMediaFailure(ctx, media, err)
	}
	if message.DocumentID != 0 && (media.DocumentID == nil || *media.DocumentID != message.DocumentID) {
		media.DocumentID = int64Pointer(message.DocumentID)
	}
	if message.DocumentDCID > 0 {
		media.DocumentDCID = intPointer(message.DocumentDCID)
	}
	if err := s.persistRefreshedDocument(ctx, media); err != nil {
		return s.recordMediaFailure(ctx, media, fmt.Errorf("persist refreshed Telegram document: %w", err))
	}
	if message.DocumentID > 0 {
		if existing, found, findErr := s.repo.FindTelegramMediaByDocumentID(ctx, message.DocumentID); findErr != nil {
			return s.recordMediaFailure(ctx, media, fmt.Errorf("find refreshed Telegram document duplicate: %w", findErr))
		} else if found && existing.VideoID != nil && existing.ID != media.ID {
			return s.finishTelegramDuplicate(ctx, media, *existing.VideoID, media.SHA256, media.TempPath)
		}
	}

	tempPath, fileHash, fileSize, err := s.obtainTelegramFile(ctx, media, message)
	if err != nil {
		return s.recordMediaFailure(ctx, media, err)
	}
	if err := s.repo.UpdateTelegramMedia(ctx, media.ID, repository.TelegramMediaPatch{
		ProcessingStatus: stringPointer("importing"),
		SHA256:           stringPointer(fileHash),
		TempPath:         stringPointer(tempPath),
		ClearNextRetryAt: true,
		LastError:        stringPointer(""),
	}); err != nil {
		return fmt.Errorf("mark Telegram media importing: %w", err)
	}
	media.TempPath = tempPath
	media.SHA256 = fileHash
	media.ProcessingStatus = "importing"

	videoID, found, err := s.repo.FindVideoByHash(ctx, fileHash, fileSize)
	if err != nil {
		return s.recordMediaFailure(ctx, media, fmt.Errorf("find imported video by hash: %w", err))
	}
	if found {
		return s.finishTelegramDuplicate(ctx, media, videoID, fileHash, tempPath)
	}

	title, description, metadata := BuildTelegramVideoMetadata(message, source.Title, time.Now().UTC())
	if s.upload == nil {
		return s.recordMediaFailure(ctx, media, errors.New("Telegram import service unavailable"))
	}
	result, err := s.upload.SaveImportedFile(ctx, LocalUploadInput{
		UserID:   s.importUserID,
		FilePath: tempPath,
		Filename: message.Filename,
		FileSize: fileSize,
		Title:    title,
		Desc:     description,
		Type:     "short",
		Hash:     fileHash,
		Metadata: metadata,
	}, s.maxVideoSize)
	if err != nil {
		return s.recordMediaFailure(ctx, media, fmt.Errorf("import Telegram video: %w", err))
	}
	if result.VideoID == uuid.Nil {
		return s.recordMediaFailure(ctx, media, errors.New("Telegram import returned empty video ID"))
	}
	if result.AlreadyExists {
		if err := s.finishTelegramDuplicate(ctx, media, result.VideoID, fileHash, tempPath); err != nil {
			return err
		}
		return nil
	}

	transcodeStatus := "not_required"
	if result.Enqueue {
		transcodeStatus = "pending"
	}
	if err := s.repo.UpdateTelegramMedia(ctx, media.ID, repository.TelegramMediaPatch{
		ProcessingStatus: stringPointer("imported"),
		TranscodeStatus:  stringPointer(transcodeStatus),
		VideoID:          uuidPointer(result.VideoID),
		SHA256:           stringPointer(fileHash),
		TempPath:         stringPointer(tempPath),
		ClearNextRetryAt: true,
		LastError:        stringPointer(""),
	}); err != nil {
		return fmt.Errorf("mark imported Telegram media: %w", err)
	}
	media.VideoID = &result.VideoID
	media.ProcessingStatus = "imported"
	media.TranscodeStatus = transcodeStatus

	if !result.Enqueue {
		return nil
	}
	if s.transcode == nil {
		return s.recordTranscodeFailure(ctx, media, errors.New("transcode queue unavailable"))
	}
	inputPath := result.InputPath
	if strings.TrimSpace(inputPath) == "" {
		inputPath = tempPath
	}
	if err := s.transcode.EnqueueTranscode(result.VideoID.String(), inputPath, result.OutputDir, result.TargetFormat, false); err != nil {
		return s.recordTranscodeFailure(ctx, media, fmt.Errorf("enqueue Telegram transcode: %w", err))
	}
	if err := s.repo.UpdateTelegramMedia(ctx, media.ID, repository.TelegramMediaPatch{
		TranscodeStatus:  stringPointer("enqueued"),
		ClearNextRetryAt: true,
		LastError:        stringPointer(""),
	}); err != nil {
		return fmt.Errorf("mark Telegram transcode enqueued: %w", err)
	}
	return nil
}

func (s *TelegramIngestionService) obtainTelegramFile(ctx context.Context, media models.TelegramMedia, message telegram.Message) (string, string, int64, error) {
	if strings.TrimSpace(media.TempPath) != "" && strings.TrimSpace(media.SHA256) != "" {
		if info, statErr := os.Stat(media.TempPath); statErr == nil && info.Mode().IsRegular() {
			fileHash, hashErr := hashTelegramFile(media.TempPath)
			if hashErr == nil && strings.EqualFold(fileHash, media.SHA256) && telegramFileSizeMatches(message.Size, info.Size()) && s.fileSizeAllowed(info.Size()) {
				return media.TempPath, fileHash, info.Size(), nil
			}
		}
	}
	if err := s.ensureUploadTempDir(); err != nil {
		return "", "", 0, err
	}
	file, err := os.CreateTemp(s.uploadTempDir, "telegram-*.part")
	if err != nil {
		return "", "", 0, fmt.Errorf("create Telegram temporary file: %w", err)
	}
	tempPath := file.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	hasher := sha256.New()
	writer := io.MultiWriter(file, hasher)
	if err := s.client.Download(ctx, message, writer); err != nil {
		if closeErr := file.Close(); closeErr != nil {
			return "", "", 0, fmt.Errorf("download Telegram video: %w; close temporary file: %v", err, closeErr)
		}
		return "", "", 0, fmt.Errorf("download Telegram video: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", "", 0, fmt.Errorf("close Telegram temporary file: %w", err)
	}
	info, err := os.Stat(tempPath)
	if err != nil {
		return "", "", 0, fmt.Errorf("stat Telegram temporary file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", "", 0, errors.New("Telegram temporary download is not a regular file")
	}
	if !s.fileSizeAllowed(info.Size()) {
		return "", "", 0, ErrFileTooLarge
	}
	if !telegramFileSizeMatches(message.Size, info.Size()) {
		return "", "", 0, fmt.Errorf("Telegram file size mismatch: expected %d, got %d", message.Size, info.Size())
	}
	removeTemp = false
	return tempPath, hex.EncodeToString(hasher.Sum(nil)), info.Size(), nil
}

func (s *TelegramIngestionService) persistRefreshedDocument(ctx context.Context, media models.TelegramMedia) error {
	patch := repository.TelegramMediaPatch{}
	if media.DocumentID != nil {
		patch.DocumentID = media.DocumentID
	}
	if media.DocumentDCID != nil {
		patch.DocumentDCID = media.DocumentDCID
	}
	if patch.DocumentID == nil && patch.DocumentDCID == nil {
		return nil
	}
	return s.repo.UpdateTelegramMedia(ctx, media.ID, patch)
}

func (s *TelegramIngestionService) finishTelegramDuplicate(ctx context.Context, media models.TelegramMedia, videoID uuid.UUID, fileHash, tempPath string) error {
	if videoID == uuid.Nil {
		return errors.New("duplicate Telegram media has empty video ID")
	}
	if err := s.repo.UpdateTelegramMedia(ctx, media.ID, repository.TelegramMediaPatch{
		ProcessingStatus: stringPointer("duplicate"),
		TranscodeStatus:  stringPointer("not_required"),
		VideoID:          uuidPointer(videoID),
		SHA256:           stringPointer(fileHash),
		TempPath:         stringPointer(""),
		ClearNextRetryAt: true,
		LastError:        stringPointer(""),
	}); err != nil {
		return fmt.Errorf("mark duplicate Telegram media: %w", err)
	}
	if strings.TrimSpace(tempPath) != "" {
		if err := os.Remove(tempPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove duplicate Telegram temporary file: %w", err)
		}
	}
	return nil
}

func (s *TelegramIngestionService) restorePausedMedia(ctx context.Context, media models.TelegramMedia) error {
	attempts := media.Attempts
	if attempts > 0 {
		attempts--
	}
	if err := s.repo.UpdateTelegramMedia(ctx, media.ID, repository.TelegramMediaPatch{
		ProcessingStatus: stringPointer("queued"),
		Attempts:         &attempts,
		ClearNextRetryAt: true,
		LastError:        stringPointer(""),
	}); err != nil {
		return fmt.Errorf("restore paused Telegram media: %w", err)
	}
	return nil
}

// ReconcileDownloads requeues media records left in an active state by a
// process restart or a worker timeout.
func (s *TelegramIngestionService) ReconcileDownloads(ctx context.Context, limit int) error {
	if err := s.validateDependencies(); err != nil {
		return err
	}
	staleBefore := time.Now().UTC().Add(-telegramDownloadRecoveryWindow)
	items, err := s.repo.ListTelegramMediaNeedingDownload(ctx, staleBefore, limit)
	if err != nil {
		return fmt.Errorf("list Telegram download recovery: %w", err)
	}
	var firstErr error
	for _, media := range items {
		source, sourceErr := s.repo.GetTelegramSource(ctx, media.SourceID)
		if sourceErr != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("get Telegram source for recovery: %w", sourceErr)
			}
			continue
		}
		if !source.Enabled || source.SyncStatus == "paused" {
			if updateErr := s.restorePausedMedia(ctx, media); updateErr != nil && firstErr == nil {
				firstErr = updateErr
			}
			continue
		}
		if enqueueErr := s.enqueueMedia(ctx, media.ID, MessagePriorityBackfill); enqueueErr != nil && firstErr == nil {
			firstErr = enqueueErr
		}
	}
	return firstErr
}

// ReconcileTranscodes retries transcode tasks whose database state is pending.
func (s *TelegramIngestionService) ReconcileTranscodes(ctx context.Context, limit int) error {
	if err := s.validateDependencies(); err != nil {
		return err
	}
	items, err := s.repo.ListTelegramMediaNeedingTranscode(ctx, limit)
	if err != nil {
		return fmt.Errorf("list Telegram transcode compensation: %w", err)
	}
	var firstErr error
	for _, media := range items {
		if media.VideoID == nil || *media.VideoID == uuid.Nil {
			continue
		}
		if s.transcode == nil {
			if firstErr == nil {
				firstErr = errors.New("transcode queue unavailable")
			}
			continue
		}
		if err := s.transcode.EnqueueTranscode(
			media.VideoID.String(),
			media.TempPath,
			filepath.Join(s.storageRoot, "videos"),
			"mp4",
			false,
		); err != nil {
			persistErr := s.repo.UpdateTelegramMedia(ctx, media.ID, repository.TelegramMediaPatch{
				TranscodeStatus: stringPointer("pending"),
				LastError:       stringPointer(err.Error()),
			})
			if persistErr != nil && firstErr == nil {
				firstErr = fmt.Errorf("reconcile Telegram transcode: %w; persist failure: %v", err, persistErr)
			} else if firstErr == nil {
				firstErr = fmt.Errorf("reconcile Telegram transcode: %w", err)
			}
			continue
		}
		if err := s.repo.UpdateTelegramMedia(ctx, media.ID, repository.TelegramMediaPatch{
			TranscodeStatus:  stringPointer("enqueued"),
			ClearNextRetryAt: true,
			LastError:        stringPointer(""),
		}); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("mark reconciled Telegram transcode: %w", err)
		}
	}
	return firstErr
}

func (s *TelegramIngestionService) enqueueMedia(ctx context.Context, mediaID uuid.UUID, priority MessagePriority) error {
	if err := s.repo.UpdateTelegramMedia(ctx, mediaID, repository.TelegramMediaPatch{
		ProcessingStatus: stringPointer("queued"),
		ClearNextRetryAt: true,
		LastError:        stringPointer(""),
	}); err != nil {
		return fmt.Errorf("queue Telegram media state: %w", err)
	}
	if s.tasks == nil {
		return s.recordMediaFailureByID(ctx, mediaID, errors.New("Telegram task queue unavailable"))
	}
	var err error
	if priority == MessagePriorityRealtime {
		err = s.tasks.EnqueueRealtime(mediaID)
	} else {
		err = s.tasks.EnqueueBackfill(mediaID)
	}
	if err == nil {
		return nil
	}
	return s.recordMediaFailureByID(ctx, mediaID, fmt.Errorf("enqueue Telegram media: %w", err))
}

func (s *TelegramIngestionService) recordMediaFailureByID(ctx context.Context, mediaID uuid.UUID, err error) error {
	media, getErr := s.repo.GetTelegramMedia(ctx, mediaID)
	if getErr != nil {
		return fmt.Errorf("%w; get media for failure state: %v", err, getErr)
	}
	return s.recordMediaFailure(ctx, media, err)
}

func (s *TelegramIngestionService) recordMediaFailure(ctx context.Context, media models.TelegramMedia, err error) error {
	if err == nil {
		return nil
	}
	nextRetryAt := time.Now().UTC().Add(telegramRetryDelay(media.Attempts, err))
	persistErr := s.repo.UpdateTelegramMedia(ctx, media.ID, repository.TelegramMediaPatch{
		ProcessingStatus: stringPointer("failed"),
		NextRetryAt:      &nextRetryAt,
		LastError:        stringPointer(err.Error()),
	})
	if persistErr != nil {
		return fmt.Errorf("%w; persist Telegram media failure: %v", err, persistErr)
	}
	return err
}

func (s *TelegramIngestionService) recordTranscodeFailure(ctx context.Context, media models.TelegramMedia, err error) error {
	if err == nil {
		return nil
	}
	persistErr := s.repo.UpdateTelegramMedia(ctx, media.ID, repository.TelegramMediaPatch{
		TranscodeStatus: stringPointer("pending"),
		LastError:       stringPointer(err.Error()),
	})
	if persistErr != nil {
		return fmt.Errorf("%w; persist Telegram transcode failure: %v", err, persistErr)
	}
	return err
}

func (s *TelegramIngestionService) recordSourceFailure(ctx context.Context, sourceID uuid.UUID, err error) error {
	if err == nil {
		return nil
	}
	nextRetryAt := time.Now().UTC().Add(telegramRetryDelay(1, err))
	persistErr := s.repo.UpdateTelegramSource(ctx, sourceID, repository.TelegramSourcePatch{
		SyncStatus:  stringPointer("error"),
		NextRetryAt: &nextRetryAt,
		LastError:   stringPointer(err.Error()),
	})
	if persistErr != nil {
		return fmt.Errorf("%w; persist Telegram source failure: %v", err, persistErr)
	}
	return err
}

func (s *TelegramIngestionService) validateDependencies() error {
	if s == nil {
		return errors.New("Telegram ingestion service unavailable")
	}
	if s.client == nil {
		return errors.New("Telegram client unavailable")
	}
	if s.repo == nil {
		return errors.New("Telegram repository unavailable")
	}
	return nil
}

func (s *TelegramIngestionService) ensureUploadTempDir() error {
	if strings.TrimSpace(s.uploadTempDir) == "" || s.uploadTempDir == "." {
		return errors.New("Telegram upload temporary directory is empty")
	}
	if err := os.MkdirAll(s.uploadTempDir, 0o700); err != nil {
		return fmt.Errorf("create Telegram upload temporary directory: %w", err)
	}
	return nil
}

func (s *TelegramIngestionService) fileSizeAllowed(size int64) bool {
	return size >= 0 && (s.maxVideoSize <= 0 || size <= s.maxVideoSize)
}

func telegramMediaFromMessage(sourceID uuid.UUID, message telegram.Message) models.TelegramMedia {
	media := models.TelegramMedia{
		ID:               uuid.New(),
		SourceID:         sourceID,
		ChatID:           message.ChatID,
		MessageID:        message.MessageID,
		Filename:         message.Filename,
		MIMEType:         message.MIMEType,
		FileSize:         message.Size,
		Caption:          message.Caption,
		MessageURL:       message.MessageURL,
		MessageCreatedAt: message.SentAt,
		ProcessingStatus: "discovered",
		TranscodeStatus:  "not_required",
	}
	if message.DocumentID > 0 {
		media.DocumentID = int64Pointer(message.DocumentID)
	}
	if message.DocumentDCID > 0 {
		media.DocumentDCID = intPointer(message.DocumentDCID)
	}
	return media
}

func validateTelegramMessage(message telegram.Message) error {
	if message.ChatID == 0 {
		return errors.New("Telegram message chat ID is empty")
	}
	if message.MessageID <= 0 {
		return errors.New("Telegram message ID must be positive")
	}
	if message.Size < 0 {
		return errors.New("Telegram video size cannot be negative")
	}
	if !isTelegramVideoMessage(message) {
		return fmt.Errorf("Telegram message %d is not a video document", message.MessageID)
	}
	return nil
}

func isTelegramVideoMessage(message telegram.Message) bool {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(message.MIMEType)), "video/") {
		return true
	}
	switch strings.ToLower(filepath.Ext(strings.TrimSpace(message.Filename))) {
	case ".avi", ".m4v", ".mkv", ".mov", ".mp4", ".mpeg", ".mpg", ".ts", ".webm", ".wmv":
		return true
	default:
		return false
	}
}

func smallestTelegramMessageID(messages []telegram.Message) int64 {
	var smallest int64
	for _, message := range messages {
		if message.MessageID <= 0 {
			continue
		}
		if smallest == 0 || message.MessageID < smallest {
			smallest = message.MessageID
		}
	}
	return smallest
}

func telegramFileSizeMatches(expected, actual int64) bool {
	return expected <= 0 || expected == actual
}

func hashTelegramFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open Telegram temporary file: %w", err)
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("hash Telegram temporary file: %w", err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func telegramRetryDelay(attempts int, err error) time.Duration {
	var floodWait *telegram.FloodWaitError
	if errors.As(err, &floodWait) && floodWait != nil && floodWait.Seconds > 0 {
		return time.Duration(floodWait.Seconds) * time.Second
	}
	if attempts < 1 {
		attempts = 1
	}
	delay := 2 * time.Second
	for i := 1; i < attempts && delay < time.Hour; i++ {
		delay *= 2
	}
	if delay > time.Hour {
		return time.Hour
	}
	return delay
}

func int64Pointer(value int64) *int64 {
	return &value
}

func intPointer(value int) *int {
	return &value
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	return &value
}

func stringPointer(value string) *string {
	return &value
}

var _ TelegramIngestion = (*TelegramIngestionService)(nil)
