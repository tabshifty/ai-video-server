package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/telegram"
)

type ingestorControlRepository interface {
	GetTelegramAccountState(ctx context.Context) (models.TelegramAccountState, error)
	GetTelegramIngestorHeartbeat(ctx context.Context) (models.TelegramIngestorHeartbeat, error)
}

type ingestorControlRuntime interface {
	View() telegramRuntimeView
}

type ingestorControlAuthorizations interface {
	ActiveAuthorization(ctx context.Context, actorID string) (*telegram.AuthorizationView, error)
	StartPhone(ctx context.Context, actorID string) (telegram.AuthorizationView, error)
	StartQR(ctx context.Context, actorID string) (telegram.AuthorizationView, error)
	GetAuthorization(ctx context.Context, authorizationID, actorID string) (telegram.AuthorizationView, error)
	SubmitCode(ctx context.Context, authorizationID, actorID, code string) (telegram.AuthorizationView, error)
	SubmitPassword(ctx context.Context, authorizationID, actorID, password string) (telegram.AuthorizationView, error)
	Cancel(ctx context.Context, authorizationID, actorID string) error
}

type ingestorControlPreviews interface {
	Preview(ctx context.Context, actorID, chatRef string) (telegram.ChatPreview, error)
	Confirm(ctx context.Context, actorID, previewID, chatRef string) (telegram.ChatPreview, error)
}

// ingestorControlService is the process-local implementation of Telegram's
// private control protocol. It never accepts or persists a session file.
type ingestorControlService struct {
	repo           ingestorControlRepository
	runtime        ingestorControlRuntime
	authorizations ingestorControlAuthorizations
	previews       ingestorControlPreviews
}

func newIngestorControlService(
	repo ingestorControlRepository,
	runtime ingestorControlRuntime,
	authorizations ingestorControlAuthorizations,
	previews ingestorControlPreviews,
) *ingestorControlService {
	return &ingestorControlService{
		repo:           repo,
		runtime:        runtime,
		authorizations: authorizations,
		previews:       previews,
	}
}

// Status returns durable account data together with the current in-process
// runtime view. A missing singleton row is a normal unconfigured state.
func (s *ingestorControlService) Status(ctx context.Context, requestID, actorID string) (telegram.ControlStatus, error) {
	if s == nil || s.repo == nil || s.runtime == nil {
		return telegram.ControlStatus{}, errors.New("Telegram 控制服务不可用")
	}
	if ctx == nil {
		return telegram.ControlStatus{}, errors.New("Telegram 控制状态 context 不能为空")
	}
	var authorization *telegram.AuthorizationView
	if s.authorizations != nil {
		view, err := s.authorizations.ActiveAuthorization(ctx, actorID)
		if err != nil {
			return telegram.ControlStatus{}, fmt.Errorf("读取 Telegram 当前授权: %w", err)
		}
		authorization = view
	}
	account, err := s.repo.GetTelegramAccountState(ctx)
	if err != nil && !errors.Is(err, repository.ErrTelegramAccountStateNotFound) {
		return telegram.ControlStatus{}, fmt.Errorf("读取 Telegram 账号状态: %w", err)
	}
	if errors.Is(err, repository.ErrTelegramAccountStateNotFound) {
		account.Status = models.TelegramAccountStatusUnconfigured
	}
	if strings.TrimSpace(account.Status) == "" {
		account.Status = models.TelegramAccountStatusUnconfigured
	}

	view := s.runtime.View()
	if strings.TrimSpace(view.Status) == "" {
		view.Status = models.TelegramIngestorStatusStopped
	}
	status := telegram.ControlStatus{
		RequestID:      strings.TrimSpace(requestID),
		IngestorStatus: view.Status,
		AccountStatus:  account.Status,
		Account:        controlAccountMap(account),
		Authorization:  authorization,
		Error:          strings.TrimSpace(view.Error),
	}
	heartbeat, err := s.repo.GetTelegramIngestorHeartbeat(ctx)
	if err != nil && !errors.Is(err, repository.ErrTelegramHeartbeatNotFound) {
		return telegram.ControlStatus{}, fmt.Errorf("读取 Telegram 采集器心跳: %w", err)
	}
	if err == nil {
		status.Heartbeat = controlHeartbeatMap(heartbeat)
	}
	return status, nil
}

func (s *ingestorControlService) StartPhone(ctx context.Context, _ string, actorID string) (telegram.AuthorizationView, error) {
	if s == nil || s.authorizations == nil {
		return telegram.AuthorizationView{}, errors.New("Telegram 授权服务不可用")
	}
	return s.authorizations.StartPhone(ctx, actorID)
}

func (s *ingestorControlService) StartQR(ctx context.Context, _ string, actorID string) (telegram.AuthorizationView, error) {
	if s == nil || s.authorizations == nil {
		return telegram.AuthorizationView{}, errors.New("Telegram 授权服务不可用")
	}
	return s.authorizations.StartQR(ctx, actorID)
}

func (s *ingestorControlService) GetAuthorization(ctx context.Context, authorizationID, _ string, actorID string) (telegram.AuthorizationView, error) {
	if s == nil || s.authorizations == nil {
		return telegram.AuthorizationView{}, errors.New("Telegram 授权服务不可用")
	}
	return s.authorizations.GetAuthorization(ctx, authorizationID, actorID)
}

func (s *ingestorControlService) SubmitCode(ctx context.Context, authorizationID, _ string, actorID, code string) (telegram.AuthorizationView, error) {
	if s == nil || s.authorizations == nil {
		return telegram.AuthorizationView{}, errors.New("Telegram 授权服务不可用")
	}
	return s.authorizations.SubmitCode(ctx, authorizationID, actorID, code)
}

func (s *ingestorControlService) SubmitPassword(ctx context.Context, authorizationID, _ string, actorID, password string) (telegram.AuthorizationView, error) {
	if s == nil || s.authorizations == nil {
		return telegram.AuthorizationView{}, errors.New("Telegram 授权服务不可用")
	}
	return s.authorizations.SubmitPassword(ctx, authorizationID, actorID, password)
}

func (s *ingestorControlService) CancelAuthorization(ctx context.Context, authorizationID, _ string, actorID string) error {
	if s == nil || s.authorizations == nil {
		return errors.New("Telegram 授权服务不可用")
	}
	return s.authorizations.Cancel(ctx, authorizationID, actorID)
}

func (s *ingestorControlService) PreviewChat(ctx context.Context, _ string, actorID, chatRef string) (telegram.ChatPreview, error) {
	if s == nil || s.previews == nil {
		return telegram.ChatPreview{}, errors.New("Telegram 来源预览服务不可用")
	}
	return s.previews.Preview(ctx, actorID, chatRef)
}

func (s *ingestorControlService) ConfirmChat(ctx context.Context, _ string, actorID, previewID, chatRef string) (telegram.ChatPreview, error) {
	if s == nil || s.previews == nil {
		return telegram.ChatPreview{}, errors.New("Telegram 来源预览服务不可用")
	}
	return s.previews.Confirm(ctx, actorID, previewID, chatRef)
}

func controlAccountMap(account models.TelegramAccountState) map[string]any {
	return map[string]any{
		"telegram_user_id": account.TelegramUserID,
		"username":         strings.TrimSpace(account.Username),
		"first_name":       strings.TrimSpace(account.FirstName),
		"last_name":        strings.TrimSpace(account.LastName),
		"phone_masked":     strings.TrimSpace(account.PhoneMasked),
		"status":           strings.TrimSpace(account.Status),
		"last_error":       strings.TrimSpace(account.LastError),
		"authorized_at":    account.AuthorizedAt,
		"last_seen_at":     account.LastSeenAt,
	}
}

func controlHeartbeatMap(heartbeat models.TelegramIngestorHeartbeat) map[string]any {
	return map[string]any{
		"status":         strings.TrimSpace(heartbeat.Status),
		"account_status": strings.TrimSpace(heartbeat.AccountStatus),
		"version":        strings.TrimSpace(heartbeat.Version),
		"error_summary":  strings.TrimSpace(heartbeat.ErrorSummary),
		"last_seen_at":   heartbeat.LastSeenAt,
	}
}

var _ telegram.ControlService = (*ingestorControlService)(nil)
