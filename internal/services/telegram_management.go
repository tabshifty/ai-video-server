package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/telegram"
)

// telegramManagementSources is the durable source lifecycle boundary used by
// the API management service. AddConfirmed must derive storage identity from a
// canonical Telegram chat ID rather than a caller-supplied reference.
type telegramManagementSources interface {
	List(ctx context.Context) ([]models.TelegramSource, error)
	AddConfirmed(ctx context.Context, actorID uuid.UUID, preview telegram.ChatPreview) (models.TelegramSource, error)
	Pause(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error)
	Resume(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error)
	Recover(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error)
	StartBackfill(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error)
	GetProgress(ctx context.Context, sourceID uuid.UUID) (TelegramSourceProgress, error)
}

// telegramManagementAudits is the append-only record boundary for management
// actions that do not already commit their audit in the same source write.
type telegramManagementAudits interface {
	CreateTelegramAuditLog(ctx context.Context, item models.TelegramAuditLog) error
	ListTelegramAuditLogs(ctx context.Context, filter models.TelegramAuditFilter) ([]models.TelegramAuditLog, int, error)
}

// TelegramManagementServiceConfig wires the API-side Telegram management
// orchestration. The control service is normally a ControlClient and never
// exposes its private token to browser code.
type TelegramManagementServiceConfig struct {
	Control       telegram.ControlService
	Sources       telegramManagementSources
	Audits        telegramManagementAudits
	Confirmations *TelegramConfirmationService
	Clock         func() time.Time
}

// TelegramManagementService coordinates authenticated administrator actions
// across the API process, collector control process, and durable source store.
type TelegramManagementService struct {
	control       telegram.ControlService
	sources       telegramManagementSources
	audits        telegramManagementAudits
	confirmations *TelegramConfirmationService
	clock         func() time.Time
}

// NewTelegramManagementService constructs the API-side Telegram management
// orchestrator. Individual methods surface unavailable optional dependencies.
func NewTelegramManagementService(config TelegramManagementServiceConfig) *TelegramManagementService {
	clock := config.Clock
	if clock == nil {
		clock = time.Now
	}
	return &TelegramManagementService{
		control:       config.Control,
		sources:       config.Sources,
		audits:        config.Audits,
		confirmations: config.Confirmations,
		clock:         clock,
	}
}

// Status returns the collector's current account and runtime snapshot.
func (s *TelegramManagementService) Status(ctx context.Context, actorID uuid.UUID, requestID string) (telegram.ControlStatus, error) {
	if err := s.requireControl(ctx, actorID); err != nil {
		return telegram.ControlStatus{}, err
	}
	status, err := s.control.Status(ctx, strings.TrimSpace(requestID), actorID.String())
	if err != nil {
		return telegram.ControlStatus{}, fmt.Errorf("读取 Telegram 管理状态: %w", err)
	}
	return status, nil
}

// IssueConfirmation verifies the current administrator password and returns a
// short-lived one-time ticket for a high-risk Telegram operation.
func (s *TelegramManagementService) IssueConfirmation(ctx context.Context, actorID uuid.UUID, action, password string, confirmed bool) (TelegramConfirmationView, error) {
	if s == nil || s.confirmations == nil {
		return TelegramConfirmationView{}, errors.New("Telegram 确认服务不可用")
	}
	return s.confirmations.Issue(ctx, actorID, action, password, confirmed)
}

// StartPhone consumes an authorization ticket before forwarding the request to
// the collector. The collector persists no API credentials or browser state.
func (s *TelegramManagementService) StartPhone(ctx context.Context, actorID uuid.UUID, requestID, confirmationTicket string) (telegram.AuthorizationView, error) {
	if err := s.consumeConfirmation(actorID, TelegramConfirmationActionPhoneAuthorization, confirmationTicket); err != nil {
		return telegram.AuthorizationView{}, err
	}
	if err := s.requireControl(ctx, actorID); err != nil {
		return telegram.AuthorizationView{}, err
	}
	view, err := s.control.StartPhone(ctx, strings.TrimSpace(requestID), actorID.String())
	if err != nil {
		return telegram.AuthorizationView{}, fmt.Errorf("启动 Telegram 手机号授权: %w", err)
	}
	return view, nil
}

// StartQR consumes a reauthorization ticket before requesting a short-lived
// QR image from the collector.
func (s *TelegramManagementService) StartQR(ctx context.Context, actorID uuid.UUID, requestID, confirmationTicket string) (telegram.AuthorizationView, error) {
	if err := s.consumeConfirmation(actorID, TelegramConfirmationActionQRAuthorization, confirmationTicket); err != nil {
		return telegram.AuthorizationView{}, err
	}
	if err := s.requireControl(ctx, actorID); err != nil {
		return telegram.AuthorizationView{}, err
	}
	view, err := s.control.StartQR(ctx, strings.TrimSpace(requestID), actorID.String())
	if err != nil {
		return telegram.AuthorizationView{}, fmt.Errorf("启动 Telegram 二维码授权: %w", err)
	}
	return view, nil
}

// GetAuthorization returns the collector's sanitized authorization progress.
func (s *TelegramManagementService) GetAuthorization(ctx context.Context, actorID uuid.UUID, authorizationID, requestID string) (telegram.AuthorizationView, error) {
	if err := s.requireControl(ctx, actorID); err != nil {
		return telegram.AuthorizationView{}, err
	}
	view, err := s.control.GetAuthorization(ctx, strings.TrimSpace(authorizationID), strings.TrimSpace(requestID), actorID.String())
	if err != nil {
		return telegram.AuthorizationView{}, fmt.Errorf("读取 Telegram 授权状态: %w", err)
	}
	return view, nil
}

// SubmitCode forwards a transient Telegram phone code directly to the
// collector. The value is not stored by this service.
func (s *TelegramManagementService) SubmitCode(ctx context.Context, actorID uuid.UUID, authorizationID, requestID, code string) (telegram.AuthorizationView, error) {
	if err := s.requireControl(ctx, actorID); err != nil {
		return telegram.AuthorizationView{}, err
	}
	view, err := s.control.SubmitCode(ctx, strings.TrimSpace(authorizationID), strings.TrimSpace(requestID), actorID.String(), code)
	if err != nil {
		return telegram.AuthorizationView{}, fmt.Errorf("提交 Telegram 验证码: %w", err)
	}
	return view, nil
}

// SubmitPassword forwards a transient Telegram 2FA password directly to the
// collector. Whitespace is preserved because it can be part of the password.
func (s *TelegramManagementService) SubmitPassword(ctx context.Context, actorID uuid.UUID, authorizationID, requestID, password string) (telegram.AuthorizationView, error) {
	if err := s.requireControl(ctx, actorID); err != nil {
		return telegram.AuthorizationView{}, err
	}
	view, err := s.control.SubmitPassword(ctx, strings.TrimSpace(authorizationID), strings.TrimSpace(requestID), actorID.String(), password)
	if err != nil {
		return telegram.AuthorizationView{}, fmt.Errorf("提交 Telegram 二次验证密码: %w", err)
	}
	return view, nil
}

// CancelAuthorization stops an authorization owned by the submitting admin.
func (s *TelegramManagementService) CancelAuthorization(ctx context.Context, actorID uuid.UUID, authorizationID, requestID string) error {
	if err := s.requireControl(ctx, actorID); err != nil {
		return err
	}
	if err := s.control.CancelAuthorization(ctx, strings.TrimSpace(authorizationID), strings.TrimSpace(requestID), actorID.String()); err != nil {
		return fmt.Errorf("取消 Telegram 授权: %w", err)
	}
	return nil
}

// PreviewSource resolves a source candidate. Private invitations require a
// one-time ticket before the raw reference crosses the API-to-collector link.
func (s *TelegramManagementService) PreviewSource(ctx context.Context, actorID uuid.UUID, requestID, chatRef, confirmationTicket string) (telegram.ChatPreview, error) {
	if err := s.requireControl(ctx, actorID); err != nil {
		return telegram.ChatPreview{}, err
	}
	private := telegram.IsPrivateInviteReference(chatRef)
	if private {
		if err := s.consumeConfirmation(actorID, TelegramConfirmationActionPrivateSourcePreview, confirmationTicket); err != nil {
			return telegram.ChatPreview{}, err
		}
	}
	preview, err := s.control.PreviewChat(ctx, strings.TrimSpace(requestID), actorID.String(), chatRef)
	if err != nil {
		return telegram.ChatPreview{}, fmt.Errorf("预解析 Telegram 来源: %w", err)
	}
	if private {
		if err := s.auditPreview(ctx, actorID, preview); err != nil {
			return telegram.ChatPreview{}, err
		}
	}
	return preview, nil
}

// ConfirmSource rechecks the collector preview and creates a durable source
// from its canonical chat identity. Private confirmation has a separate
// ticket because it can join an invitation-only chat.
func (s *TelegramManagementService) ConfirmSource(ctx context.Context, actorID uuid.UUID, requestID, previewID, chatRef, confirmationTicket string) (models.TelegramSource, error) {
	if err := s.requireControl(ctx, actorID); err != nil {
		return models.TelegramSource{}, err
	}
	if s.sources == nil {
		return models.TelegramSource{}, errors.New("Telegram 来源服务不可用")
	}
	if telegram.IsPrivateInviteReference(chatRef) {
		if err := s.consumeConfirmation(actorID, TelegramConfirmationActionPrivateSourceConfirm, confirmationTicket); err != nil {
			return models.TelegramSource{}, err
		}
	}
	preview, err := s.control.ConfirmChat(ctx, strings.TrimSpace(requestID), actorID.String(), strings.TrimSpace(previewID), chatRef)
	if err != nil {
		return models.TelegramSource{}, fmt.Errorf("确认 Telegram 来源: %w", err)
	}
	source, err := s.sources.AddConfirmed(ctx, actorID, preview)
	if err != nil {
		return models.TelegramSource{}, fmt.Errorf("创建 Telegram 来源: %w", err)
	}
	return source, nil
}

// ListSources returns persisted source lifecycle state.
func (s *TelegramManagementService) ListSources(ctx context.Context) ([]models.TelegramSource, error) {
	if s == nil || s.sources == nil {
		return nil, errors.New("Telegram 来源服务不可用")
	}
	items, err := s.sources.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取 Telegram 来源: %w", err)
	}
	return items, nil
}

// PauseSource disables a source while retaining its persisted history cursor.
func (s *TelegramManagementService) PauseSource(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	return s.updateSource(ctx, actorID, sourceID, func() (models.TelegramSource, error) {
		return s.sources.Pause(ctx, actorID, sourceID)
	})
}

// ResumeSource reenables a source at its existing history cursor.
func (s *TelegramManagementService) ResumeSource(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	return s.updateSource(ctx, actorID, sourceID, func() (models.TelegramSource, error) {
		return s.sources.Resume(ctx, actorID, sourceID)
	})
}

// RecoverSource restarts a source explicitly left in an error state while
// retaining its history cursor.
func (s *TelegramManagementService) RecoverSource(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	return s.updateSource(ctx, actorID, sourceID, func() (models.TelegramSource, error) {
		return s.sources.Recover(ctx, actorID, sourceID)
	})
}

// StartBackfill resets a source cursor and schedules a full history scan.
func (s *TelegramManagementService) StartBackfill(ctx context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	return s.updateSource(ctx, actorID, sourceID, func() (models.TelegramSource, error) {
		return s.sources.StartBackfill(ctx, actorID, sourceID)
	})
}

// GetProgress returns persisted media processing counts for one source.
func (s *TelegramManagementService) GetProgress(ctx context.Context, sourceID uuid.UUID) (TelegramSourceProgress, error) {
	if s == nil || s.sources == nil {
		return TelegramSourceProgress{}, errors.New("Telegram 来源服务不可用")
	}
	progress, err := s.sources.GetProgress(ctx, sourceID)
	if err != nil {
		return TelegramSourceProgress{}, fmt.Errorf("读取 Telegram 来源进度: %w", err)
	}
	return progress, nil
}

// ListAudits exposes the administrator-visible, sanitized audit history.
func (s *TelegramManagementService) ListAudits(ctx context.Context, filter models.TelegramAuditFilter) ([]models.TelegramAuditLog, int, error) {
	if s == nil || s.audits == nil {
		return nil, 0, errors.New("Telegram 审计服务不可用")
	}
	items, total, err := s.audits.ListTelegramAuditLogs(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("读取 Telegram 审计: %w", err)
	}
	return items, total, nil
}

func (s *TelegramManagementService) updateSource(ctx context.Context, actorID, sourceID uuid.UUID, update func() (models.TelegramSource, error)) (models.TelegramSource, error) {
	if s == nil || s.sources == nil {
		return models.TelegramSource{}, errors.New("Telegram 来源服务不可用")
	}
	if actorID == uuid.Nil {
		return models.TelegramSource{}, errors.New("Telegram 来源缺少管理员身份")
	}
	if sourceID == uuid.Nil {
		return models.TelegramSource{}, errors.New("Telegram 来源 ID 格式错误")
	}
	source, err := update()
	if err != nil {
		return models.TelegramSource{}, err
	}
	return source, nil
}

func (s *TelegramManagementService) requireControl(ctx context.Context, actorID uuid.UUID) error {
	if s == nil || s.control == nil {
		return errors.New("Telegram 控制服务不可用")
	}
	if ctx == nil {
		return errors.New("Telegram 管理 context 不能为空")
	}
	if actorID == uuid.Nil {
		return errors.New("Telegram 管理缺少管理员身份")
	}
	return nil
}

func (s *TelegramManagementService) consumeConfirmation(actorID uuid.UUID, action, ticket string) error {
	if s == nil || s.confirmations == nil {
		return errors.New("Telegram 确认服务不可用")
	}
	return s.confirmations.Consume(actorID, action, ticket)
}

func (s *TelegramManagementService) auditPreview(ctx context.Context, actorID uuid.UUID, preview telegram.ChatPreview) error {
	if s == nil || s.audits == nil {
		return errors.New("Telegram 审计服务不可用")
	}
	summary, err := telegramManagementAuditSummary(map[string]any{
		"chat_id":           preview.ChatID,
		"chat_type":         strings.TrimSpace(preview.ChatType),
		"requires_join":     preview.RequiresJoin,
		"requires_approval": preview.RequiresApproval,
	})
	if err != nil {
		return err
	}
	return s.audits.CreateTelegramAuditLog(ctx, models.TelegramAuditLog{
		ActorUserID: actorID,
		Action:      "source.private_previewed",
		TargetType:  "preview",
		TargetID:    strings.TrimSpace(preview.PreviewID),
		Result:      "succeeded",
		Summary:     summary,
		CreatedAt:   s.now(),
	})
}

func telegramManagementAuditSummary(value map[string]any) (json.RawMessage, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, errors.New("编码 Telegram 审计摘要失败")
	}
	return json.RawMessage(encoded), nil
}

func (s *TelegramManagementService) now() time.Time {
	if s == nil || s.clock == nil {
		return time.Now().UTC()
	}
	return s.clock().UTC()
}
