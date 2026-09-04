package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"video-server/internal/models"
	"video-server/internal/telegram"
)

var (
	// ErrTelegramRuntimeUnavailable indicates that the collector has no active,
	// authorized MTProto connection for a control operation.
	ErrTelegramRuntimeUnavailable = errors.New("Telegram 采集器当前不可用")
)

type telegramRuntimeClient interface {
	RunAuthorized(ctx context.Context, fn func(context.Context) error) error
	PreviewChat(ctx context.Context, ref string) (telegram.ChatPreviewResult, error)
	ConfirmChat(ctx context.Context, ref string) (telegram.ChatPreviewResult, error)
}

type telegramRuntimeClientFactory func() (telegramRuntimeClient, error)

type telegramRuntimeRunner func(context.Context, telegramRuntimeClient) error

type telegramRuntimeView struct {
	Status string
	Error  string
}

// telegramRuntime owns the one long-lived collector connection. Its lifecycle
// is deliberately independent from the private control HTTP server so an
// unconfigured or reauthorizing account can still be managed from the page.
type telegramRuntime struct {
	newClient telegramRuntimeClientFactory
	run       telegramRuntimeRunner

	mu      sync.RWMutex
	started bool
	paused  bool
	client  telegramRuntimeClient
	cancel  context.CancelFunc
	done    chan struct{}
	view    telegramRuntimeView
	restart chan struct{}
}

func newTelegramRuntime(newClient telegramRuntimeClientFactory, run telegramRuntimeRunner) *telegramRuntime {
	return &telegramRuntime{
		newClient: newClient,
		run:       run,
		view: telegramRuntimeView{
			Status: models.TelegramIngestorStatusStopped,
		},
		restart: make(chan struct{}, 1),
	}
}

// Run starts the collector lifecycle and returns only when ctx is cancelled or
// the runtime itself is invalid. A session failure leaves the control plane
// alive and waits for Resume after a later authorization attempt.
func (r *telegramRuntime) Run(ctx context.Context) error {
	if r == nil {
		return errors.New("Telegram 采集运行时不可用")
	}
	if ctx == nil {
		return errors.New("Telegram 采集运行时 context 不能为空")
	}
	if r.newClient == nil {
		return errors.New("Telegram 采集客户端工厂不可用")
	}
	if r.run == nil {
		return errors.New("Telegram 采集运行回调不可用")
	}

	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return errors.New("Telegram 采集运行时已启动")
	}
	r.started = true
	r.mu.Unlock()
	defer func() {
		r.mu.Lock()
		r.started = false
		r.client = nil
		r.cancel = nil
		r.done = nil
		r.view = telegramRuntimeView{Status: models.TelegramIngestorStatusStopped}
		r.mu.Unlock()
	}()

	for {
		if !r.waitUntilRunnable(ctx) {
			return nil
		}

		client, err := r.newClient()
		if err != nil {
			r.setIdleError(err)
			if !r.waitForRestart(ctx) {
				return nil
			}
			continue
		}
		if client == nil {
			r.setIdleError(errors.New("Telegram 采集客户端工厂返回空值"))
			if !r.waitForRestart(ctx) {
				return nil
			}
			continue
		}

		runCtx, cancel := context.WithCancel(ctx)
		done := make(chan struct{})
		if !r.activate(client, cancel, done) {
			cancel()
			continue
		}
		err = client.RunAuthorized(runCtx, func(authorisedCtx context.Context) error {
			if !r.markRunning(done) {
				return context.Canceled
			}
			return r.run(authorisedCtx, client)
		})
		cancel()
		close(done)

		paused := r.deactivate(done, err)
		if ctx.Err() != nil {
			return nil
		}
		if paused {
			continue
		}
		if !r.waitForRestart(ctx) {
			return nil
		}
	}
}

// Pause stops the active collector connection and waits until it no longer
// owns the persistent Telegram session. It satisfies AuthorizationMaintenance.
func (r *telegramRuntime) Pause(ctx context.Context) error {
	if r == nil {
		return errors.New("Telegram 采集运行时不可用")
	}
	if ctx == nil {
		return errors.New("Telegram 暂停 context 不能为空")
	}

	r.mu.Lock()
	r.paused = true
	cancel := r.cancel
	done := r.done
	if done == nil {
		r.view = telegramRuntimeView{Status: models.TelegramIngestorStatusAuthorizing}
	} else {
		r.view = telegramRuntimeView{Status: models.TelegramIngestorStatusDraining}
	}
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done == nil {
		return nil
	}
	select {
	case <-done:
		r.mu.Lock()
		if r.paused {
			r.view = telegramRuntimeView{Status: models.TelegramIngestorStatusAuthorizing}
		}
		r.mu.Unlock()
		return nil
	case <-ctx.Done():
		r.abortPause(done)
		return fmt.Errorf("等待 Telegram 采集连接停止: %w", ctx.Err())
	}
}

// Resume permits the runtime to build a new client from the persistent session
// after authorization succeeds or an attempted authorization is rolled back.
func (r *telegramRuntime) Resume() {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.paused = false
	if r.done == nil {
		r.view = telegramRuntimeView{Status: models.TelegramIngestorStatusStopped}
	}
	r.mu.Unlock()
	r.signalRestart()
}

// PreviewChat delegates a non-mutating source lookup only while the collector
// has an active authorized connection.
func (r *telegramRuntime) PreviewChat(ctx context.Context, ref string) (telegram.ChatPreviewResult, error) {
	client, err := r.activeClient()
	if err != nil {
		return telegram.ChatPreviewResult{}, err
	}
	return client.PreviewChat(ctx, ref)
}

// ConfirmChat delegates an explicitly confirmed source action only while the
// collector has an active authorized connection.
func (r *telegramRuntime) ConfirmChat(ctx context.Context, ref string) (telegram.ChatPreviewResult, error) {
	client, err := r.activeClient()
	if err != nil {
		return telegram.ChatPreviewResult{}, err
	}
	return client.ConfirmChat(ctx, ref)
}

// View returns a stable snapshot for heartbeats and control status responses.
func (r *telegramRuntime) View() telegramRuntimeView {
	if r == nil {
		return telegramRuntimeView{Status: models.TelegramIngestorStatusStopped}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.view
}

func (r *telegramRuntime) activate(client telegramRuntimeClient, cancel context.CancelFunc, done chan struct{}) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.paused {
		return false
	}
	r.client = client
	r.cancel = cancel
	r.done = done
	r.view = telegramRuntimeView{Status: models.TelegramIngestorStatusStopped}
	return true
}

func (r *telegramRuntime) markRunning(done chan struct{}) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.paused || r.done != done {
		return false
	}
	r.view = telegramRuntimeView{Status: models.TelegramIngestorStatusRunning}
	return true
}

func (r *telegramRuntime) deactivate(done chan struct{}, err error) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done != done {
		return r.paused
	}
	r.client = nil
	r.cancel = nil
	r.done = nil
	if r.paused {
		r.view = telegramRuntimeView{Status: models.TelegramIngestorStatusAuthorizing}
		return true
	}
	if errors.Is(err, telegram.ErrTelegramSessionUnauthorized) {
		r.view = telegramRuntimeView{
			Status: models.TelegramIngestorStatusStopped,
			Error:  "Telegram session 尚未授权",
		}
		return false
	}
	if err == nil {
		r.view = telegramRuntimeView{
			Status: models.TelegramIngestorStatusError,
			Error:  "Telegram 采集连接意外停止",
		}
		return false
	}
	r.view = telegramRuntimeView{
		Status: models.TelegramIngestorStatusError,
		Error:  telegramRuntimeError(err),
	}
	return false
}

func (r *telegramRuntime) abortPause(done chan struct{}) {
	r.mu.Lock()
	if r.paused && (r.done == done || r.done == nil) {
		r.paused = false
		r.view = telegramRuntimeView{Status: models.TelegramIngestorStatusStopped}
	}
	r.mu.Unlock()
	r.signalRestart()
}

func (r *telegramRuntime) activeClient() (telegramRuntimeClient, error) {
	if r == nil {
		return nil, ErrTelegramRuntimeUnavailable
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.paused || r.view.Status != models.TelegramIngestorStatusRunning || r.client == nil {
		return nil, ErrTelegramRuntimeUnavailable
	}
	return r.client, nil
}

func (r *telegramRuntime) setIdleError(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.paused {
		r.view = telegramRuntimeView{Status: models.TelegramIngestorStatusAuthorizing}
		return
	}
	r.view = telegramRuntimeView{
		Status: models.TelegramIngestorStatusError,
		Error:  telegramRuntimeError(err),
	}
}

func (r *telegramRuntime) waitUntilRunnable(ctx context.Context) bool {
	for {
		r.mu.RLock()
		paused := r.paused
		r.mu.RUnlock()
		if !paused {
			return true
		}
		if !r.waitForRestart(ctx) {
			return false
		}
	}
}

func (r *telegramRuntime) waitForRestart(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-r.restart:
		return true
	}
}

func (r *telegramRuntime) signalRestart() {
	select {
	case r.restart <- struct{}{}:
	default:
	}
}

func telegramRuntimeError(err error) string {
	if err == nil {
		return ""
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "Telegram 采集运行失败"
	}
	const maxLength = 500
	if len(message) > maxLength {
		return message[:maxLength]
	}
	return message
}
