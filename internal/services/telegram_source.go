package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/telegram"
)

var (
	ErrTelegramSourceAlreadyExists  = errors.New("Telegram 来源已存在")
	ErrTelegramSourceUnresolved     = errors.New("Telegram 来源尚未获得可保存的 chat ID")
	ErrTelegramSourceTypeInvalid    = errors.New("Telegram 来源类型不受支持")
	ErrTelegramSourceNotRecoverable = errors.New("Telegram 来源当前不处于可恢复失败状态")
)

type telegramSourceRepository interface {
	CreateTelegramSourceWithAudit(ctx context.Context, source models.TelegramSource, audit models.TelegramAuditLog) error
	UpdateTelegramSourceWithAudit(ctx context.Context, sourceID uuid.UUID, patch repository.TelegramSourcePatch, audit models.TelegramAuditLog) error
	ListTelegramSources(ctx context.Context, enabledOnly bool) ([]models.TelegramSource, error)
	GetTelegramSource(ctx context.Context, sourceID uuid.UUID) (models.TelegramSource, error)
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

// AddConfirmed creates a source from the collector's confirmed canonical chat
// identity. It intentionally derives ChatRef from ChatID so private invitation
// links and their tokens never enter the database.
func (s *TelegramSourceService) AddConfirmed(ctx context.Context, actorID uuid.UUID, preview telegram.ChatPreview) (models.TelegramSource, error) {
	if err := s.validateDependencies(); err != nil {
		return models.TelegramSource{}, err
	}
	if actorID == uuid.Nil {
		return models.TelegramSource{}, errors.New("Telegram 来源缺少管理员身份")
	}
	if err := validateConfirmedTelegramPreview(preview); err != nil {
		return models.TelegramSource{}, err
	}
	existing, err := s.repo.ListTelegramSources(ctx, false)
	if err != nil {
		return models.TelegramSource{}, fmt.Errorf("检查 Telegram 来源: %w", err)
	}
	canonicalRef := strconv.FormatInt(preview.ChatID, 10)
	for _, source := range existing {
		if source.ChatID == preview.ChatID || telegramChatRefKey(source.ChatRef) == canonicalRef {
			return models.TelegramSource{}, fmt.Errorf("%w: %s", ErrTelegramSourceAlreadyExists, source.ID)
		}
	}

	now := time.Now().UTC()
	source := models.TelegramSource{
		ID:         uuid.New(),
		ChatID:     preview.ChatID,
		ChatRef:    canonicalRef,
		Title:      strings.TrimSpace(preview.Title),
		Username:   strings.TrimSpace(preview.Username),
		Enabled:    true,
		SyncStatus: "pending",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	audit, err := telegramSourceCreatedAudit(actorID, source, preview.ChatType)
	if err != nil {
		return models.TelegramSource{}, err
	}
	if err := s.repo.CreateTelegramSourceWithAudit(ctx, source, audit); err != nil {
		if repository.IsUniqueViolation(err) {
			return models.TelegramSource{}, ErrTelegramSourceAlreadyExists
		}
		return models.TelegramSource{}, fmt.Errorf("新增 Telegram 已确认来源: %w", err)
	}
	if err := s.enqueueSource(source.ID); err != nil {
		return source, err
	}
	return source, nil
}

// Pause disables a source without changing its history cursor.
func (s *TelegramSourceService) Pause(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	return s.updateSourceWithAudit(ctx, actorID, sourceID, "source.paused", repository.TelegramSourcePatch{
		Enabled:    telegramBoolPointer(false),
		SyncStatus: telegramStringPointer("paused"),
	}, false)
}

// Resume enables a source at its existing history cursor and schedules sync.
func (s *TelegramSourceService) Resume(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	return s.updateSourceWithAudit(ctx, actorID, sourceID, "source.resumed", repository.TelegramSourcePatch{
		Enabled:          telegramBoolPointer(true),
		SyncStatus:       telegramStringPointer("pending"),
		LastError:        telegramStringPointer(""),
		ClearNextRetryAt: true,
	}, true)
}

// Recover clears a source-level failure while preserving its history cursor.
func (s *TelegramSourceService) Recover(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	if err := s.validateDependencies(); err != nil {
		return models.TelegramSource{}, err
	}
	if sourceID == uuid.Nil {
		return models.TelegramSource{}, repository.ErrTelegramSourceNotFound
	}
	source, err := s.repo.GetTelegramSource(ctx, sourceID)
	if err != nil {
		return models.TelegramSource{}, fmt.Errorf("读取 Telegram 来源: %w", err)
	}
	if source.SyncStatus != "error" {
		return models.TelegramSource{}, ErrTelegramSourceNotRecoverable
	}
	return s.updateKnownSourceWithAudit(ctx, actorID, source, "source.recovery_started", repository.TelegramSourcePatch{
		Enabled:          telegramBoolPointer(true),
		SyncStatus:       telegramStringPointer("pending"),
		LastError:        telegramStringPointer(""),
		ClearNextRetryAt: true,
	}, true)
}

// StartBackfill resets the history cursor and schedules a full history scan.
func (s *TelegramSourceService) StartBackfill(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	return s.updateSourceWithAudit(ctx, actorID, sourceID, "source.backfill_started", repository.TelegramSourcePatch{
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

func (s *TelegramSourceService) updateSourceWithAudit(ctx context.Context, actorID, sourceID uuid.UUID, action string, patch repository.TelegramSourcePatch, enqueue bool) (models.TelegramSource, error) {
	if err := s.validateDependencies(); err != nil {
		return models.TelegramSource{}, err
	}
	if sourceID == uuid.Nil {
		return models.TelegramSource{}, repository.ErrTelegramSourceNotFound
	}
	if actorID == uuid.Nil {
		return models.TelegramSource{}, errors.New("Telegram 来源缺少管理员身份")
	}
	source, err := s.repo.GetTelegramSource(ctx, sourceID)
	if err != nil {
		return models.TelegramSource{}, fmt.Errorf("读取 Telegram 来源: %w", err)
	}
	return s.updateKnownSourceWithAudit(ctx, actorID, source, action, patch, enqueue)
}

func (s *TelegramSourceService) updateKnownSourceWithAudit(ctx context.Context, actorID uuid.UUID, source models.TelegramSource, action string, patch repository.TelegramSourcePatch, enqueue bool) (models.TelegramSource, error) {
	audit, err := telegramSourceLifecycleAudit(actorID, action, source, patch)
	if err != nil {
		return models.TelegramSource{}, err
	}
	if err := s.repo.UpdateTelegramSourceWithAudit(ctx, source.ID, patch, audit); err != nil {
		return models.TelegramSource{}, fmt.Errorf("更新 Telegram 来源: %w", err)
	}
	updated, err := s.repo.GetTelegramSource(ctx, source.ID)
	if err != nil {
		return models.TelegramSource{}, fmt.Errorf("读取更新后的 Telegram 来源: %w", err)
	}
	if enqueue {
		if err := s.enqueueSource(source.ID); err != nil {
			return updated, fmt.Errorf("Telegram 来源状态已更新，但调度同步失败: %w", err)
		}
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

func validateConfirmedTelegramPreview(preview telegram.ChatPreview) error {
	if preview.ChatID == 0 || preview.RequiresJoin || preview.RequiresApproval {
		return ErrTelegramSourceUnresolved
	}
	if strings.TrimSpace(preview.Title) == "" {
		return ErrTelegramSourceUnresolved
	}
	switch preview.ChatType {
	case telegram.TelegramChatTypeGroup, telegram.TelegramChatTypeSupergroup, telegram.TelegramChatTypeChannel:
		return nil
	default:
		return ErrTelegramSourceTypeInvalid
	}
}

func telegramSourceCreatedAudit(actorID uuid.UUID, source models.TelegramSource, chatType string) (models.TelegramAuditLog, error) {
	summary, err := json.Marshal(map[string]any{
		"chat_id":   source.ChatID,
		"chat_type": strings.TrimSpace(chatType),
		"username":  strings.TrimSpace(source.Username),
	})
	if err != nil {
		return models.TelegramAuditLog{}, errors.New("编码 Telegram 来源审计失败")
	}
	return models.TelegramAuditLog{
		ActorUserID: actorID,
		Action:      "source.created",
		TargetType:  "source",
		TargetID:    source.ID.String(),
		Result:      "succeeded",
		Summary:     summary,
		CreatedAt:   source.CreatedAt,
	}, nil
}

func telegramSourceLifecycleAudit(actorID uuid.UUID, action string, source models.TelegramSource, patch repository.TelegramSourcePatch) (models.TelegramAuditLog, error) {
	if actorID == uuid.Nil || source.ID == uuid.Nil {
		return models.TelegramAuditLog{}, errors.New("Telegram 来源审计缺少操作者或来源")
	}
	enabled := source.Enabled
	if patch.Enabled != nil {
		enabled = *patch.Enabled
	}
	syncStatus := source.SyncStatus
	if patch.SyncStatus != nil {
		syncStatus = *patch.SyncStatus
	}
	summary, err := json.Marshal(map[string]any{
		"chat_id":     source.ChatID,
		"enabled":     enabled,
		"sync_status": strings.TrimSpace(syncStatus),
	})
	if err != nil {
		return models.TelegramAuditLog{}, errors.New("编码 Telegram 来源状态审计失败")
	}
	return models.TelegramAuditLog{
		ActorUserID: actorID,
		Action:      strings.TrimSpace(action),
		TargetType:  "source",
		TargetID:    source.ID.String(),
		Result:      "succeeded",
		Summary:     summary,
		CreatedAt:   time.Now().UTC(),
	}, nil
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
