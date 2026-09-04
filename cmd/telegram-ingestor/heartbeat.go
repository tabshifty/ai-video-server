package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"video-server/internal/models"
	"video-server/internal/repository"
)

const telegramHeartbeatInterval = 30 * time.Second

type telegramHeartbeatRepository interface {
	GetTelegramAccountState(ctx context.Context) (models.TelegramAccountState, error)
	UpsertTelegramIngestorHeartbeat(ctx context.Context, item models.TelegramIngestorHeartbeat) error
}

type telegramHeartbeatRuntime interface {
	View() telegramRuntimeView
}

// runTelegramHeartbeats writes immediately and then at a fixed interval until
// the collector process stops. Heartbeat write failures do not stop ingestion.
func runTelegramHeartbeats(ctx context.Context, repo telegramHeartbeatRepository, runtime telegramHeartbeatRuntime, version string, logger *slog.Logger) {
	if ctx == nil || repo == nil || runtime == nil {
		return
	}
	write := func() {
		if err := writeTelegramHeartbeat(ctx, repo, runtime, version, time.Now); err != nil && logger != nil {
			logger.Warn("写入 Telegram 采集器心跳失败", "error", err)
		}
	}
	write()
	ticker := time.NewTicker(telegramHeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			write()
		}
	}
}

func writeTelegramHeartbeat(
	ctx context.Context,
	repo telegramHeartbeatRepository,
	runtime telegramHeartbeatRuntime,
	version string,
	clock func() time.Time,
) error {
	if ctx == nil {
		return errors.New("Telegram 心跳 context 不能为空")
	}
	if repo == nil || runtime == nil {
		return errors.New("Telegram 心跳依赖不可用")
	}
	if clock == nil {
		clock = time.Now
	}

	accountStatus := models.TelegramAccountStatusUnconfigured
	account, err := repo.GetTelegramAccountState(ctx)
	if err == nil && strings.TrimSpace(account.Status) != "" {
		accountStatus = account.Status
	}
	if err != nil && !errors.Is(err, repository.ErrTelegramAccountStateNotFound) {
		accountStatus = models.TelegramAccountStatusError
	}
	view := runtime.View()
	if strings.TrimSpace(view.Status) == "" {
		view.Status = models.TelegramIngestorStatusStopped
	}
	errorSummary := strings.TrimSpace(view.Error)
	if err != nil && !errors.Is(err, repository.ErrTelegramAccountStateNotFound) {
		accountError := fmt.Sprintf("读取 Telegram 账号状态: %v", err)
		if errorSummary == "" {
			errorSummary = accountError
		} else {
			errorSummary += "; " + accountError
		}
	}
	now := clock().UTC()
	return repo.UpsertTelegramIngestorHeartbeat(ctx, models.TelegramIngestorHeartbeat{
		Status:        view.Status,
		AccountStatus: accountStatus,
		Version:       strings.TrimSpace(version),
		ErrorSummary:  errorSummary,
		LastSeenAt:    now,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
}
