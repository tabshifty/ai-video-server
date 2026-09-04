package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"video-server/internal/models"
	"video-server/internal/telegram"
)

func TestTelegramRuntimePausesBeforeSessionReplacement(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	first := newRuntimeTestClient()
	second := newRuntimeTestClient()
	factory := newRuntimeTestFactory(first, second)
	runtime := newTelegramRuntime(factory.New, runtimeTestRunner)
	runDone := make(chan error, 1)
	go func() { runDone <- runtime.Run(ctx) }()

	waitRuntimeSignal(t, first.runnerStarted, "first runtime start")
	if _, err := runtime.PreviewChat(context.Background(), "@before-pause"); err != nil {
		t.Fatalf("PreviewChat() before pause error = %v", err)
	}

	pauseCtx, pauseCancel := context.WithTimeout(context.Background(), time.Second)
	defer pauseCancel()
	if err := runtime.Pause(pauseCtx); err != nil {
		t.Fatalf("Pause() error = %v", err)
	}
	waitRuntimeSignal(t, first.runExited, "first runtime stop")
	if view := runtime.View(); view.Status != models.TelegramIngestorStatusAuthorizing {
		t.Fatalf("runtime view after pause = %+v, want authorizing", view)
	}
	if _, err := runtime.PreviewChat(context.Background(), "@while-paused"); !errors.Is(err, ErrTelegramRuntimeUnavailable) {
		t.Fatalf("PreviewChat() while paused error = %v, want ErrTelegramRuntimeUnavailable", err)
	}

	runtime.Resume()
	waitRuntimeSignal(t, second.runnerStarted, "second runtime start")
	if factory.Calls() != 2 {
		t.Fatalf("client factory calls = %d, want 2", factory.Calls())
	}

	cancel()
	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("runtime did not stop after context cancellation")
	}
}

func TestTelegramRuntimeWaitsForResumeAfterUnauthorizedSession(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	unauthorized := newRuntimeTestClient()
	unauthorized.runAuthorized = func(context.Context, func(context.Context) error) error {
		return telegram.ErrTelegramSessionUnauthorized
	}
	authorized := newRuntimeTestClient()
	factory := newRuntimeTestFactory(unauthorized, authorized)
	runtime := newTelegramRuntime(factory.New, runtimeTestRunner)
	runDone := make(chan error, 1)
	go func() { runDone <- runtime.Run(ctx) }()

	waitForRuntimeFactoryCalls(t, factory, 1)
	waitForRuntimeStatus(t, runtime, models.TelegramIngestorStatusStopped)
	if factory.Calls() != 1 {
		t.Fatalf("client factory calls before resume = %d, want 1", factory.Calls())
	}

	runtime.Resume()
	waitRuntimeSignal(t, authorized.runnerStarted, "authorized runtime start")

	cancel()
	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("runtime did not stop after context cancellation")
	}
}

type runtimeTestFactory struct {
	mu      sync.Mutex
	clients []*runtimeTestClient
	calls   int
}

func newRuntimeTestFactory(clients ...*runtimeTestClient) *runtimeTestFactory {
	return &runtimeTestFactory{clients: clients}
}

func (f *runtimeTestFactory) New() (telegramRuntimeClient, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.calls >= len(f.clients) {
		return nil, errors.New("unexpected Telegram runtime client request")
	}
	client := f.clients[f.calls]
	f.calls++
	return client, nil
}

func (f *runtimeTestFactory) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

type runtimeTestClient struct {
	runnerStarted chan struct{}
	runExited     chan struct{}
	runAuthorized func(context.Context, func(context.Context) error) error
}

func newRuntimeTestClient() *runtimeTestClient {
	client := &runtimeTestClient{
		runnerStarted: make(chan struct{}),
		runExited:     make(chan struct{}),
	}
	client.runAuthorized = func(ctx context.Context, run func(context.Context) error) error {
		err := run(ctx)
		close(client.runExited)
		return err
	}
	return client
}

func (c *runtimeTestClient) RunAuthorized(ctx context.Context, run func(context.Context) error) error {
	return c.runAuthorized(ctx, run)
}

func (c *runtimeTestClient) PreviewChat(context.Context, string) (telegram.ChatPreviewResult, error) {
	return telegram.ChatPreviewResult{ChatID: 1, Title: "预览来源"}, nil
}

func (c *runtimeTestClient) ConfirmChat(context.Context, string) (telegram.ChatPreviewResult, error) {
	return telegram.ChatPreviewResult{ChatID: 1, Title: "确认来源"}, nil
}

func runtimeTestRunner(ctx context.Context, client telegramRuntimeClient) error {
	testClient, ok := client.(*runtimeTestClient)
	if !ok {
		return errors.New("unexpected Telegram runtime client type")
	}
	close(testClient.runnerStarted)
	<-ctx.Done()
	return ctx.Err()
}

func waitRuntimeSignal(t *testing.T, signal <-chan struct{}, description string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", description)
	}
}

func waitForRuntimeStatus(t *testing.T, runtime *telegramRuntime, want string) {
	t.Helper()
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		if runtime.View().Status == want {
			return
		}
		select {
		case <-deadline.C:
			t.Fatalf("runtime status = %+v, want %q", runtime.View(), want)
		case <-ticker.C:
		}
	}
}

func waitForRuntimeFactoryCalls(t *testing.T, factory *runtimeTestFactory, want int) {
	t.Helper()
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		if factory.Calls() >= want {
			return
		}
		select {
		case <-deadline.C:
			t.Fatalf("client factory calls = %d, want at least %d", factory.Calls(), want)
		case <-ticker.C:
		}
	}
}
