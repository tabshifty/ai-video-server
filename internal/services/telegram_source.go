package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
)

var (
	ErrTelegramChatRefRequired     = errors.New("Telegram 群组引用不能为空")
	ErrTelegramSourceAlreadyExists = errors.New("Telegram 来源已存在")
)

type telegramSourceRepository interface {
	CreateTelegramSource(ctx context.Context, source models.TelegramSource) error
	ListTelegramSources(ctx context.Context, enabledOnly bool) ([]models.TelegramSource, error)
	GetTelegramSource(ctx context.Context, sourceID uuid.UUID) (models.TelegramSource, error)
	UpdateTelegramSource(ctx context.Context, sourceID uuid.UUID, patch repository.TelegramSourcePatch) error
	CountTelegramMediaByStatus(ctx context.Context, sourceID uuid.UUID) (map[string]int, error)
}

type telegramSourceTaskScheduler interface {
	EnqueueSourceSync(sourceID uuid.UUID) error
}

// TelegramSourceProgress combines source state with persisted media counts.
type TelegramSourceProgress struct {
	Source models.TelegramSource `json:"source"`
	Counts map[string]int        `json:"counts"`
}

// TelegramSourceService manages administrator-selected Telegram sources.
type TelegramSourceService struct {
	repo  telegramSourceRepository
	tasks telegramSourceTaskScheduler
}

// NewTelegramSourceService constructs the source management service.
func NewTelegramSourceService(repo telegramSourceRepository, tasks telegramSourceTaskScheduler) *TelegramSourceService {
	return &TelegramSourceService{repo: repo, tasks: tasks}
}

// List returns all configured Telegram sources, including paused sources.
func (s *TelegramSourceService) List(ctx context.Context) ([]models.TelegramSource, error) {
	if err := s.validateDependencies(); err != nil {
		return nil, err
	}
	items, err := s.repo.ListTelegramSources(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("读取 Telegram 来源: %w", err)
	}
	return items, nil
}

// Add stores a chat reference and schedules asynchronous resolution.
func (s *TelegramSourceService) Add(ctx context.Context, chatRef string) (models.TelegramSource, error) {
	if err := s.validateDependencies(); err != nil {
		return models.TelegramSource{}, err
	}
	chatRef = normalizeTelegramChatRef(chatRef)
	if chatRef == "" {
		return models.TelegramSource{}, ErrTelegramChatRefRequired
	}
	existing, err := s.repo.ListTelegramSources(ctx, false)
	if err != nil {
		return models.TelegramSource{}, fmt.Errorf("检查 Telegram 来源: %w", err)
	}
	wantedKey := telegramChatRefKey(chatRef)
	for _, source := range existing {
		if telegramChatRefKey(source.ChatRef) == wantedKey {
			return models.TelegramSource{}, fmt.Errorf("%w: %s", ErrTelegramSourceAlreadyExists, source.ID)
		}
	}

	now := time.Now().UTC()
	source := models.TelegramSource{
		ID:         uuid.New(),
		ChatRef:    chatRef,
		Enabled:    true,
		SyncStatus: "pending",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.CreateTelegramSource(ctx, source); err != nil {
		if repository.IsUniqueViolation(err) {
			return models.TelegramSource{}, ErrTelegramSourceAlreadyExists
		}
		return models.TelegramSource{}, fmt.Errorf("新增 Telegram 来源: %w", err)
	}
	if err := s.enqueueSource(source.ID); err != nil {
		return source, err
	}
	return source, nil
}

// Pause disables a source without changing its history cursor.
func (s *TelegramSourceService) Pause(ctx context.Context, sourceID uuid.UUID) (models.TelegramSource, error) {
	return s.updateSource(ctx, sourceID, repository.TelegramSourcePatch{
		Enabled:    telegramBoolPointer(false),
		SyncStatus: telegramStringPointer("paused"),
	}, false)
}

// Resume enables a source at its existing history cursor and schedules sync.
func (s *TelegramSourceService) Resume(ctx context.Context, sourceID uuid.UUID) (models.TelegramSource, error) {
	return s.updateSource(ctx, sourceID, repository.TelegramSourcePatch{
		Enabled:          telegramBoolPointer(true),
		SyncStatus:       telegramStringPointer("pending"),
		LastError:        telegramStringPointer(""),
		ClearNextRetryAt: true,
	}, true)
}

// StartBackfill resets the history cursor and schedules a full history scan.
func (s *TelegramSourceService) StartBackfill(ctx context.Context, sourceID uuid.UUID) (models.TelegramSource, error) {
	return s.updateSource(ctx, sourceID, repository.TelegramSourcePatch{
		Enabled:                  telegramBoolPointer(true),
		SyncStatus:               telegramStringPointer("pending"),
		HistoryCursorMessageID:   telegramInt64Pointer(0),
		LastError:                telegramStringPointer(""),
		ClearNextRetryAt:         true,
		ClearBackfillCompletedAt: true,
	}, true)
}

// GetProgress returns source state and media counts grouped by processing status.
func (s *TelegramSourceService) GetProgress(ctx context.Context, sourceID uuid.UUID) (TelegramSourceProgress, error) {
	if err := s.validateDependencies(); err != nil {
		return TelegramSourceProgress{}, err
	}
	source, err := s.repo.GetTelegramSource(ctx, sourceID)
	if err != nil {
		return TelegramSourceProgress{}, fmt.Errorf("读取 Telegram 来源: %w", err)
	}
	counts, err := s.repo.CountTelegramMediaByStatus(ctx, sourceID)
	if err != nil {
		return TelegramSourceProgress{}, fmt.Errorf("统计 Telegram 来源进度: %w", err)
	}
	if counts == nil {
		counts = map[string]int{}
	}
	return TelegramSourceProgress{Source: source, Counts: counts}, nil
}

func (s *TelegramSourceService) updateSource(ctx context.Context, sourceID uuid.UUID, patch repository.TelegramSourcePatch, enqueue bool) (models.TelegramSource, error) {
	if err := s.validateDependencies(); err != nil {
		return models.TelegramSource{}, err
	}
	if sourceID == uuid.Nil {
		return models.TelegramSource{}, repository.ErrTelegramSourceNotFound
	}
	if _, err := s.repo.GetTelegramSource(ctx, sourceID); err != nil {
		return models.TelegramSource{}, fmt.Errorf("读取 Telegram 来源: %w", err)
	}
	if err := s.repo.UpdateTelegramSource(ctx, sourceID, patch); err != nil {
		return models.TelegramSource{}, fmt.Errorf("更新 Telegram 来源: %w", err)
	}
	if enqueue {
		if err := s.enqueueSource(sourceID); err != nil {
			return models.TelegramSource{}, err
		}
	}
	updated, err := s.repo.GetTelegramSource(ctx, sourceID)
	if err != nil {
		return models.TelegramSource{}, fmt.Errorf("读取更新后的 Telegram 来源: %w", err)
	}
	return updated, nil
}

func (s *TelegramSourceService) enqueueSource(sourceID uuid.UUID) error {
	if s.tasks == nil {
		return errors.New("Telegram 控制队列不可用")
	}
	if err := s.tasks.EnqueueSourceSync(sourceID); err != nil {
		return fmt.Errorf("调度 Telegram 来源同步: %w", err)
	}
	return nil
}

func (s *TelegramSourceService) validateDependencies() error {
	if s == nil || s.repo == nil {
		return errors.New("Telegram 来源服务不可用")
	}
	return nil
}

func normalizeTelegramChatRef(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}

func telegramChatRefKey(value string) string {
	return strings.ToLower(normalizeTelegramChatRef(value))
}

func telegramBoolPointer(value bool) *bool       { return &value }
func telegramStringPointer(value string) *string { return &value }
func telegramInt64Pointer(value int64) *int64    { return &value }
