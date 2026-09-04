package main

import (
	"context"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/telegram"
)

func TestIngestorControlServiceProjectsStatusAndDelegatesOperations(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	userID := int64(42)
	now := time.Date(2026, time.September, 4, 8, 30, 0, 0, time.UTC)
	repo := &ingestorControlTestRepository{
		account: models.TelegramAccountState{
			TelegramUserID: &userID,
			Username:       "collector",
			PhoneMasked:    "+861****0000",
			Status:         models.TelegramAccountStatusAuthorized,
		},
		heartbeat: models.TelegramIngestorHeartbeat{
			Status:        models.TelegramIngestorStatusRunning,
			AccountStatus: models.TelegramAccountStatusAuthorized,
			LastSeenAt:    now,
		},
	}
	authorizationID := uuid.New()
	authorizations := &ingestorControlTestAuthorization{active: &telegram.AuthorizationView{
		ID:      authorizationID,
		Kind:    telegram.AuthorizationKindPhone,
		Status:  telegram.AuthorizationStatusAwaitingCode,
		OwnerID: actorID,
	}}
	previews := &ingestorControlTestPreview{}
	service := newIngestorControlService(repo, ingestorControlTestRuntime{view: telegramRuntimeView{
		Status: models.TelegramIngestorStatusRunning,
	}}, authorizations, previews)

	status, err := service.Status(context.Background(), "request-123", actorID.String())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.RequestID != "request-123" || status.IngestorStatus != models.TelegramIngestorStatusRunning || status.AccountStatus != models.TelegramAccountStatusAuthorized {
		t.Fatalf("Status() = %+v", status)
	}
	if got, _ := status.Account["username"].(string); got != "collector" {
		t.Fatalf("status account = %+v", status.Account)
	}
	if got, _ := status.Heartbeat["status"].(string); got != models.TelegramIngestorStatusRunning {
		t.Fatalf("status heartbeat = %+v", status.Heartbeat)
	}
	if status.Authorization == nil || status.Authorization.ID != authorizationID || authorizations.activeActor != actorID.String() {
		t.Fatalf("status authorization = %+v, fake=%+v", status.Authorization, authorizations)
	}

	if _, err := service.StartPhone(context.Background(), "request-123", actorID.String()); err != nil {
		t.Fatalf("StartPhone() error = %v", err)
	}
	if authorizations.startPhoneActor != actorID.String() {
		t.Fatalf("StartPhone() actor = %q", authorizations.startPhoneActor)
	}
	preview, err := service.PreviewChat(context.Background(), "request-123", actorID.String(), "@channel")
	if err != nil {
		t.Fatalf("PreviewChat() error = %v", err)
	}
	if preview.PreviewID != "preview-1" || previews.previewActor != actorID.String() || previews.previewRef != "@channel" {
		t.Fatalf("PreviewChat() = %+v, fake=%+v", preview, previews)
	}
	if _, err := service.ConfirmChat(context.Background(), "request-123", actorID.String(), "preview-1", "@channel"); err != nil {
		t.Fatalf("ConfirmChat() error = %v", err)
	}
	if previews.confirmActor != actorID.String() || previews.confirmID != "preview-1" || previews.confirmRef != "@channel" {
		t.Fatalf("ConfirmChat fake = %+v", previews)
	}
}

func TestWriteTelegramHeartbeatUsesRuntimeAndAccountState(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.September, 4, 8, 31, 0, 0, time.UTC)
	repo := &ingestorControlTestRepository{
		account: models.TelegramAccountState{Status: models.TelegramAccountStatusAuthorized},
	}
	runtime := ingestorControlTestRuntime{view: telegramRuntimeView{
		Status: models.TelegramIngestorStatusError,
		Error:  "Telegram 网络连接失败",
	}}
	if err := writeTelegramHeartbeat(context.Background(), repo, runtime, "test-build", func() time.Time { return now }); err != nil {
		t.Fatalf("writeTelegramHeartbeat() error = %v", err)
	}
	got := repo.writtenHeartbeat
	if got.Status != models.TelegramIngestorStatusError || got.AccountStatus != models.TelegramAccountStatusAuthorized || got.ErrorSummary != "Telegram 网络连接失败" || got.Version != "test-build" || !got.LastSeenAt.Equal(now) {
		t.Fatalf("written heartbeat = %+v", got)
	}
}

func TestRunTelegramControlPlaneServesAndWritesHeartbeatBeforeAuthorization(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	userID := int64(42)
	repo := &controlPlaneTestRepository{account: models.TelegramAccountState{
		TelegramUserID: &userID,
		Status:         models.TelegramAccountStatusAuthorized,
	}}
	runtime := &controlPlaneTestRuntime{
		started: make(chan struct{}, 1),
		view:    telegramRuntimeView{Status: models.TelegramIngestorStatusStopped},
	}
	control := newIngestorControlService(repo, runtime, nil, nil)
	done := make(chan error, 1)
	go func() {
		done <- runTelegramControlPlaneOnListener(ctx, listener, "control-token", control, repo, runtime, "test-build", nil)
	}()

	waitRuntimeSignal(t, runtime.started, "control plane runtime start")
	response := getControlStatus(t, listener.Addr().String(), "control-token")
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("control status response = %d, want %d", response.StatusCode, http.StatusOK)
	}
	waitForControlPlaneHeartbeats(t, repo, 1)

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("runTelegramControlPlaneOnListener() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("control plane did not stop after context cancellation")
	}
}

func getControlStatus(t *testing.T, address, token string) *http.Response {
	t.Helper()
	client := &http.Client{Timeout: 100 * time.Millisecond}
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	for {
		request, err := http.NewRequest(http.MethodGet, "http://"+address+telegram.ControlStatusPath, nil)
		if err != nil {
			t.Fatalf("http.NewRequest() error = %v", err)
		}
		request.Header.Set(telegram.ControlTokenHeader, token)
		response, err := client.Do(request)
		if err == nil {
			return response
		}
		select {
		case <-deadline.C:
			t.Fatalf("request control status: %v", err)
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func waitForControlPlaneHeartbeats(t *testing.T, repo *controlPlaneTestRepository, want int) {
	t.Helper()
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		if repo.HeartbeatWrites() >= want {
			return
		}
		select {
		case <-deadline.C:
			t.Fatalf("heartbeat writes = %d, want at least %d", repo.HeartbeatWrites(), want)
		case <-ticker.C:
		}
	}
}

type ingestorControlTestRuntime struct {
	view telegramRuntimeView
}

func (r ingestorControlTestRuntime) View() telegramRuntimeView {
	return r.view
}

type ingestorControlTestRepository struct {
	account          models.TelegramAccountState
	heartbeat        models.TelegramIngestorHeartbeat
	writtenHeartbeat models.TelegramIngestorHeartbeat
}

func (r *ingestorControlTestRepository) GetTelegramAccountState(context.Context) (models.TelegramAccountState, error) {
	return r.account, nil
}

func (r *ingestorControlTestRepository) GetTelegramIngestorHeartbeat(context.Context) (models.TelegramIngestorHeartbeat, error) {
	return r.heartbeat, nil
}

func (r *ingestorControlTestRepository) UpsertTelegramIngestorHeartbeat(_ context.Context, heartbeat models.TelegramIngestorHeartbeat) error {
	r.writtenHeartbeat = heartbeat
	return nil
}

type ingestorControlTestAuthorization struct {
	startPhoneActor string
	activeActor     string
	active          *telegram.AuthorizationView
}

func (s *ingestorControlTestAuthorization) ActiveAuthorization(_ context.Context, actorID string) (*telegram.AuthorizationView, error) {
	s.activeActor = actorID
	return s.active, nil
}

func (s *ingestorControlTestAuthorization) StartPhone(_ context.Context, actorID string) (telegram.AuthorizationView, error) {
	s.startPhoneActor = actorID
	return telegram.AuthorizationView{}, nil
}

func (s *ingestorControlTestAuthorization) StartQR(context.Context, string) (telegram.AuthorizationView, error) {
	return telegram.AuthorizationView{}, nil
}

func (s *ingestorControlTestAuthorization) GetAuthorization(context.Context, string, string) (telegram.AuthorizationView, error) {
	return telegram.AuthorizationView{}, nil
}

func (s *ingestorControlTestAuthorization) SubmitCode(context.Context, string, string, string) (telegram.AuthorizationView, error) {
	return telegram.AuthorizationView{}, nil
}

func (s *ingestorControlTestAuthorization) SubmitPassword(context.Context, string, string, string) (telegram.AuthorizationView, error) {
	return telegram.AuthorizationView{}, nil
}

func (s *ingestorControlTestAuthorization) Cancel(context.Context, string, string) error {
	return nil
}

type ingestorControlTestPreview struct {
	previewActor string
	previewRef   string
	confirmActor string
	confirmID    string
	confirmRef   string
}

func (s *ingestorControlTestPreview) Preview(_ context.Context, actorID, chatRef string) (telegram.ChatPreview, error) {
	s.previewActor = actorID
	s.previewRef = chatRef
	return telegram.ChatPreview{PreviewID: "preview-1", ChatID: -1001, Title: "频道", ChatType: telegram.TelegramChatTypeChannel}, nil
}

func (s *ingestorControlTestPreview) Confirm(_ context.Context, actorID, previewID, chatRef string) (telegram.ChatPreview, error) {
	s.confirmActor = actorID
	s.confirmID = previewID
	s.confirmRef = chatRef
	return telegram.ChatPreview{PreviewID: previewID, ChatID: -1001, Title: "频道", ChatType: telegram.TelegramChatTypeChannel}, nil
}

type controlPlaneTestRuntime struct {
	started chan struct{}
	view    telegramRuntimeView
}

func (r *controlPlaneTestRuntime) Run(ctx context.Context) error {
	select {
	case r.started <- struct{}{}:
	default:
	}
	<-ctx.Done()
	return nil
}

func (r *controlPlaneTestRuntime) View() telegramRuntimeView {
	return r.view
}

type controlPlaneTestRepository struct {
	mu              sync.Mutex
	account         models.TelegramAccountState
	heartbeat       models.TelegramIngestorHeartbeat
	heartbeatWrites int
}

func (r *controlPlaneTestRepository) GetTelegramAccountState(context.Context) (models.TelegramAccountState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.account, nil
}

func (r *controlPlaneTestRepository) GetTelegramIngestorHeartbeat(context.Context) (models.TelegramIngestorHeartbeat, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.heartbeat, nil
}

func (r *controlPlaneTestRepository) UpsertTelegramIngestorHeartbeat(_ context.Context, heartbeat models.TelegramIngestorHeartbeat) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.heartbeat = heartbeat
	r.heartbeatWrites++
	return nil
}

func (r *controlPlaneTestRepository) HeartbeatWrites() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.heartbeatWrites
}
