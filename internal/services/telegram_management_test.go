package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/telegram"
)

func TestTelegramManagementStartsPhoneOnlyAfterConsumingConfirmation(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	confirmation := newTelegramConfirmationTestService(t, actorID, time.Date(2026, time.September, 4, 8, 50, 0, 0, time.UTC))
	control := &telegramManagementTestControl{}
	service := NewTelegramManagementService(TelegramManagementServiceConfig{
		Control:       control,
		Confirmations: confirmation,
	})

	if _, err := service.StartPhone(context.Background(), actorID, "request-123", "missing-ticket"); !errors.Is(err, ErrTelegramConfirmationNotFound) {
		t.Fatalf("StartPhone() without ticket error = %v", err)
	}
	ticket, err := confirmation.Issue(context.Background(), actorID, TelegramConfirmationActionPhoneAuthorization, "admin-password", true)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	view, err := service.StartPhone(context.Background(), actorID, "request-123", ticket.Ticket)
	if err != nil {
		t.Fatalf("StartPhone() error = %v", err)
	}
	if view.Status != telegram.AuthorizationStatusAwaitingCode || control.startPhoneActor != actorID.String() || control.startPhoneRequestID != "request-123" {
		t.Fatalf("StartPhone() view=%+v control=%+v", view, control)
	}
	if _, err := service.StartPhone(context.Background(), actorID, "request-123", ticket.Ticket); !errors.Is(err, ErrTelegramConfirmationNotFound) {
		t.Fatalf("StartPhone() replay error = %v", err)
	}
}

func TestTelegramManagementRequiresPrivateTicketsAndAuditsWithoutInviteReference(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	confirmation := newTelegramConfirmationTestService(t, actorID, time.Date(2026, time.September, 4, 8, 51, 0, 0, time.UTC))
	control := &telegramManagementTestControl{
		preview: telegram.ChatPreview{
			PreviewID:    uuid.NewString(),
			ChatID:       -100321,
			Title:        "私密频道",
			ChatType:     telegram.TelegramChatTypeChannel,
			RequiresJoin: true,
			ExpiresAt:    "2026-09-04T09:00:00Z",
		},
		confirmed: telegram.ChatPreview{
			PreviewID: uuid.NewString(),
			ChatID:    -100321,
			Title:     "私密频道",
			ChatType:  telegram.TelegramChatTypeChannel,
		},
	}
	sources := &telegramManagementTestSources{}
	audits := &telegramManagementTestAudits{}
	service := NewTelegramManagementService(TelegramManagementServiceConfig{
		Control:       control,
		Sources:       sources,
		Audits:        audits,
		Confirmations: confirmation,
	})
	const invite = "https://t.me/+privateInviteTokenMustNotPersist"

	if _, err := service.PreviewSource(context.Background(), actorID, "request-preview", invite, ""); !errors.Is(err, ErrTelegramConfirmationNotFound) {
		t.Fatalf("PreviewSource() without ticket error = %v", err)
	}
	previewTicket, err := confirmation.Issue(context.Background(), actorID, TelegramConfirmationActionPrivateSourcePreview, "admin-password", true)
	if err != nil {
		t.Fatalf("preview Issue() error = %v", err)
	}
	preview, err := service.PreviewSource(context.Background(), actorID, "request-preview", invite, previewTicket.Ticket)
	if err != nil {
		t.Fatalf("PreviewSource() error = %v", err)
	}
	if preview.PreviewID == "" || control.previewRef != invite {
		t.Fatalf("preview=%+v control=%+v", preview, control)
	}
	if len(audits.items) != 1 {
		t.Fatalf("audit count = %d, want 1", len(audits.items))
	}
	if summary := string(audits.items[0].Summary); strings.Contains(summary, "privateInviteTokenMustNotPersist") {
		t.Fatalf("audit leaked invite token: %s", summary)
	}

	if _, err := service.ConfirmSource(context.Background(), actorID, "request-confirm", preview.PreviewID, invite, ""); !errors.Is(err, ErrTelegramConfirmationNotFound) {
		t.Fatalf("ConfirmSource() without ticket error = %v", err)
	}
	confirmTicket, err := confirmation.Issue(context.Background(), actorID, TelegramConfirmationActionPrivateSourceConfirm, "admin-password", true)
	if err != nil {
		t.Fatalf("confirm Issue() error = %v", err)
	}
	source, err := service.ConfirmSource(context.Background(), actorID, "request-confirm", preview.PreviewID, invite, confirmTicket.Ticket)
	if err != nil {
		t.Fatalf("ConfirmSource() error = %v", err)
	}
	if source.ChatID != -100321 || sources.addedActor != actorID || sources.added.ChatRef != "-100321" || control.confirmRef != invite {
		t.Fatalf("source=%+v sources=%+v control=%+v", source, sources, control)
	}
}

func TestTelegramManagementListsStatusAndSourceOperationsWithAudit(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	sourceID := uuid.New()
	sources := &telegramManagementTestSources{items: []models.TelegramSource{{ID: sourceID, ChatID: -1001, ChatRef: "-1001", Enabled: true}}}
	audits := &telegramManagementTestAudits{}
	control := &telegramManagementTestControl{status: telegram.ControlStatus{
		IngestorStatus: "running",
		AccountStatus:  "authorized",
	}}
	service := NewTelegramManagementService(TelegramManagementServiceConfig{Control: control, Sources: sources, Audits: audits})

	status, err := service.Status(context.Background(), actorID, "request-status")
	if err != nil || status.IngestorStatus != "running" || control.statusActor != actorID.String() {
		t.Fatalf("Status() status=%+v err=%v control=%+v", status, err, control)
	}
	if _, err := service.PauseSource(context.Background(), actorID, sourceID); err != nil {
		t.Fatalf("PauseSource() error = %v", err)
	}
	if sources.lastAction != "pause" || sources.lastActor != actorID {
		t.Fatalf("source operation=%q actor=%s", sources.lastAction, sources.lastActor)
	}
	if _, err := service.RecoverSource(context.Background(), actorID, sourceID); err != nil {
		t.Fatalf("RecoverSource() error = %v", err)
	}
	if sources.lastAction != "recover" || sources.lastActor != actorID {
		t.Fatalf("source operation=%q actor=%s", sources.lastAction, sources.lastActor)
	}
}

type telegramManagementTestControl struct {
	status              telegram.ControlStatus
	preview             telegram.ChatPreview
	confirmed           telegram.ChatPreview
	statusActor         string
	startPhoneActor     string
	startPhoneRequestID string
	previewRef          string
	confirmRef          string
}

func (s *telegramManagementTestControl) Status(_ context.Context, _ string, actorID string) (telegram.ControlStatus, error) {
	s.statusActor = actorID
	return s.status, nil
}

func (s *telegramManagementTestControl) StartPhone(_ context.Context, requestID, actorID string) (telegram.AuthorizationView, error) {
	s.startPhoneActor = actorID
	s.startPhoneRequestID = requestID
	return telegram.AuthorizationView{ID: uuid.New(), Kind: telegram.AuthorizationKindPhone, Status: telegram.AuthorizationStatusAwaitingCode}, nil
}

func (s *telegramManagementTestControl) StartQR(context.Context, string, string) (telegram.AuthorizationView, error) {
	return telegram.AuthorizationView{}, nil
}

func (s *telegramManagementTestControl) GetAuthorization(context.Context, string, string, string) (telegram.AuthorizationView, error) {
	return telegram.AuthorizationView{}, nil
}

func (s *telegramManagementTestControl) SubmitCode(context.Context, string, string, string, string) (telegram.AuthorizationView, error) {
	return telegram.AuthorizationView{}, nil
}

func (s *telegramManagementTestControl) SubmitPassword(context.Context, string, string, string, string) (telegram.AuthorizationView, error) {
	return telegram.AuthorizationView{}, nil
}

func (s *telegramManagementTestControl) CancelAuthorization(context.Context, string, string, string) error {
	return nil
}

func (s *telegramManagementTestControl) PreviewChat(_ context.Context, _ string, _ string, chatRef string) (telegram.ChatPreview, error) {
	s.previewRef = chatRef
	return s.preview, nil
}

func (s *telegramManagementTestControl) ConfirmChat(_ context.Context, _ string, _ string, _ string, chatRef string) (telegram.ChatPreview, error) {
	s.confirmRef = chatRef
	return s.confirmed, nil
}

type telegramManagementTestSources struct {
	items      []models.TelegramSource
	added      models.TelegramSource
	addedActor uuid.UUID
	lastAction string
	lastActor  uuid.UUID
}

func (s *telegramManagementTestSources) List(context.Context) ([]models.TelegramSource, error) {
	return s.items, nil
}

func (s *telegramManagementTestSources) AddConfirmed(_ context.Context, actorID uuid.UUID, preview telegram.ChatPreview) (models.TelegramSource, error) {
	s.addedActor = actorID
	s.added = models.TelegramSource{ID: uuid.New(), ChatID: preview.ChatID, ChatRef: "-100321", Title: preview.Title, Enabled: true}
	return s.added, nil
}

func (s *telegramManagementTestSources) Pause(_ context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	s.lastAction = "pause"
	s.lastActor = actorID
	return models.TelegramSource{ID: sourceID, ChatID: -1001, ChatRef: "-1001", Enabled: false}, nil
}

func (s *telegramManagementTestSources) Resume(_ context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	s.lastAction = "resume"
	s.lastActor = actorID
	return models.TelegramSource{ID: sourceID}, nil
}

func (s *telegramManagementTestSources) Recover(_ context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	s.lastAction = "recover"
	s.lastActor = actorID
	return models.TelegramSource{ID: sourceID}, nil
}

func (s *telegramManagementTestSources) StartBackfill(_ context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	s.lastAction = "backfill"
	s.lastActor = actorID
	return models.TelegramSource{ID: sourceID}, nil
}

func (s *telegramManagementTestSources) GetProgress(context.Context, uuid.UUID) (TelegramSourceProgress, error) {
	return TelegramSourceProgress{}, nil
}

type telegramManagementTestAudits struct {
	items []models.TelegramAuditLog
	err   error
}

func (s *telegramManagementTestAudits) CreateTelegramAuditLog(_ context.Context, item models.TelegramAuditLog) error {
	if s.err != nil {
		return s.err
	}
	s.items = append(s.items, item)
	return nil
}

func (s *telegramManagementTestAudits) ListTelegramAuditLogs(context.Context, models.TelegramAuditFilter) ([]models.TelegramAuditLog, int, error) {
	return s.items, len(s.items), s.err
}

func TestTelegramManagementAuditSummaryIsJSONObject(t *testing.T) {
	t.Parallel()

	value, err := telegramManagementAuditSummary(map[string]any{"chat_id": int64(-1001)})
	if err != nil {
		t.Fatalf("telegramManagementAuditSummary() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(value, &decoded); err != nil || decoded["chat_id"] == nil {
		t.Fatalf("summary=%s err=%v", value, err)
	}
}
