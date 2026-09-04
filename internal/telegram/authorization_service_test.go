package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
)

func TestAuthorizationServicePhoneBindsAccountWithoutPersistingSecret(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	repo := newAuthorizationServiceFakeRepository()
	session := &authorizationServiceFakeSession{}
	maintenance := &authorizationServiceFakeMaintenance{}
	service := NewAuthorizationService(AuthorizationServiceConfig{
		Phone:       "+8613800000000",
		Repository:  repo,
		Sessions:    authorizationServiceFakeFactory{session: session},
		Maintenance: maintenance,
		State:       NewAuthorizationStateMachine(nil, time.Minute),
	})

	view, err := service.StartPhone(context.Background(), actorID.String())
	if err != nil {
		t.Fatalf("StartPhone() error = %v", err)
	}
	if view.Status != AuthorizationStatusAwaitingCode {
		t.Fatalf("StartPhone() status = %q, want %q", view.Status, AuthorizationStatusAwaitingCode)
	}
	if repo.account.Status != models.TelegramAccountStatusAuthorizing || repo.account.PhoneMasked != "+861****0000" {
		t.Fatalf("account after start = %+v", repo.account)
	}
	if maintenance.pauseCalls != 1 || maintenance.resumeCalls != 0 {
		t.Fatalf("maintenance calls after start = pause:%d resume:%d", maintenance.pauseCalls, maintenance.resumeCalls)
	}

	const code = "246810"
	if _, err := service.SubmitCode(context.Background(), view.ID.String(), actorID.String(), code); err != nil {
		t.Fatalf("SubmitCode() error = %v", err)
	}
	if session.code != code {
		t.Fatalf("session code = %q, want %q", session.code, code)
	}
	if encoded, err := json.Marshal(repo); err != nil {
		t.Fatalf("marshal fake repository: %v", err)
	} else if strings.Contains(string(encoded), code) {
		t.Fatalf("authorization code leaked into persisted state: %s", encoded)
	}

	identity := TelegramIdentity{ID: 701, Username: "collector", FirstName: "采集", LastName: "账号"}
	session.emit(AuthorizationSessionUpdate{Identity: &identity})

	got, err := service.GetAuthorization(context.Background(), view.ID.String(), uuid.New().String())
	if err != nil {
		t.Fatalf("GetAuthorization() error = %v", err)
	}
	if got.Status != AuthorizationStatusSucceeded || got.TelegramUserID == nil || *got.TelegramUserID != identity.ID {
		t.Fatalf("authorization result = %+v", got)
	}
	if !session.promoted || session.discarded {
		t.Fatalf("session promotion state = promoted:%v discarded:%v", session.promoted, session.discarded)
	}
	if repo.account.Status != models.TelegramAccountStatusAuthorized || repo.account.TelegramUserID == nil || *repo.account.TelegramUserID != identity.ID {
		t.Fatalf("authorized account = %+v", repo.account)
	}
	if maintenance.resumeCalls != 1 {
		t.Fatalf("maintenance resume calls = %d, want 1", maintenance.resumeCalls)
	}
	if len(repo.audits) < 2 {
		t.Fatalf("audit count = %d, want start and success records", len(repo.audits))
	}
}

func TestAuthorizationServiceRejectsQRBeforePhoneBinding(t *testing.T) {
	t.Parallel()

	service := NewAuthorizationService(AuthorizationServiceConfig{
		Repository: newAuthorizationServiceFakeRepository(),
		Sessions:   authorizationServiceFakeFactory{session: &authorizationServiceFakeSession{}},
		State:      NewAuthorizationStateMachine(nil, time.Minute),
	})

	_, err := service.StartQR(context.Background(), uuid.New().String())
	if !errors.Is(err, ErrAuthorizationPhoneBindingRequired) {
		t.Fatalf("StartQR() error = %v, want ErrAuthorizationPhoneBindingRequired", err)
	}
}

func TestAuthorizationServiceRejectsMismatchedReauthorizationAndKeepsBinding(t *testing.T) {
	t.Parallel()

	boundID := int64(1001)
	repo := newAuthorizationServiceFakeRepository()
	repo.account = models.TelegramAccountState{
		ID:             1,
		TelegramUserID: &boundID,
		Username:       "bound",
		Status:         models.TelegramAccountStatusAuthorized,
	}
	session := &authorizationServiceFakeSession{}
	maintenance := &authorizationServiceFakeMaintenance{}
	service := NewAuthorizationService(AuthorizationServiceConfig{
		Repository:  repo,
		Sessions:    authorizationServiceFakeFactory{session: session},
		Maintenance: maintenance,
		State:       NewAuthorizationStateMachine(nil, time.Minute),
	})

	view, err := service.StartQR(context.Background(), uuid.New().String())
	if err != nil {
		t.Fatalf("StartQR() error = %v", err)
	}
	mismatch := TelegramIdentity{ID: 1002, Username: "different"}
	session.emit(AuthorizationSessionUpdate{Identity: &mismatch})

	got, err := service.GetAuthorization(context.Background(), view.ID.String(), uuid.New().String())
	if err != nil {
		t.Fatalf("GetAuthorization() error = %v", err)
	}
	if got.Status != AuthorizationStatusFailed || !strings.Contains(got.Error, "不一致") {
		t.Fatalf("authorization result = %+v", got)
	}
	if session.promoted || !session.discarded {
		t.Fatalf("mismatched session state = promoted:%v discarded:%v", session.promoted, session.discarded)
	}
	if repo.account.TelegramUserID == nil || *repo.account.TelegramUserID != boundID || repo.account.Status != models.TelegramAccountStatusAuthorized {
		t.Fatalf("existing binding changed: %+v", repo.account)
	}
	if maintenance.resumeCalls != 1 {
		t.Fatalf("maintenance resume calls = %d, want 1", maintenance.resumeCalls)
	}
}

func TestAuthorizationServiceRestrictsSecretSubmissionToOwnerAndCancels(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	repo := newAuthorizationServiceFakeRepository()
	session := &authorizationServiceFakeSession{}
	service := NewAuthorizationService(AuthorizationServiceConfig{
		Repository: repo,
		Sessions:   authorizationServiceFakeFactory{session: session},
		State:      NewAuthorizationStateMachine(nil, time.Minute),
	})
	view, err := service.StartPhone(context.Background(), actorID.String())
	if err != nil {
		t.Fatalf("StartPhone() error = %v", err)
	}
	if _, err := service.SubmitCode(context.Background(), view.ID.String(), uuid.New().String(), "12345"); !errors.Is(err, ErrAuthorizationNotOwner) {
		t.Fatalf("non-owner SubmitCode() error = %v, want ErrAuthorizationNotOwner", err)
	}
	if err := service.Cancel(context.Background(), view.ID.String(), actorID.String()); err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	if !session.closed || !session.discarded {
		t.Fatalf("cancelled session state = closed:%v discarded:%v", session.closed, session.discarded)
	}
	got, err := service.GetAuthorization(context.Background(), view.ID.String(), actorID.String())
	if err != nil {
		t.Fatalf("GetAuthorization() error = %v", err)
	}
	if got.Status != AuthorizationStatusCancelled || got.QRImageDataURL != "" {
		t.Fatalf("cancelled authorization = %+v", got)
	}
}

type authorizationServiceFakeRepository struct {
	mu             sync.Mutex
	account        models.TelegramAccountState
	heartbeat      models.TelegramIngestorHeartbeat
	authorizations map[uuid.UUID]models.TelegramAuthorization
	audits         []models.TelegramAuditLog
}

func newAuthorizationServiceFakeRepository() *authorizationServiceFakeRepository {
	return &authorizationServiceFakeRepository{authorizations: map[uuid.UUID]models.TelegramAuthorization{}}
}

func (r *authorizationServiceFakeRepository) GetTelegramAccountState(context.Context) (models.TelegramAccountState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.account, nil
}

func (r *authorizationServiceFakeRepository) UpsertTelegramAccountState(_ context.Context, account models.TelegramAccountState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.account = account
	return nil
}

func (r *authorizationServiceFakeRepository) GetTelegramIngestorHeartbeat(context.Context) (models.TelegramIngestorHeartbeat, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.heartbeat, nil
}

func (r *authorizationServiceFakeRepository) CreateTelegramAuthorization(_ context.Context, item models.TelegramAuthorization) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.authorizations[item.ID] = item
	return nil
}

func (r *authorizationServiceFakeRepository) UpdateTelegramAuthorization(_ context.Context, id uuid.UUID, status, errorSummary string, telegramUserID *int64, finishedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	item := r.authorizations[id]
	item.Status = status
	item.ErrorSummary = errorSummary
	item.TelegramUserID = telegramUserID
	item.FinishedAt = finishedAt
	r.authorizations[id] = item
	return nil
}

func (r *authorizationServiceFakeRepository) CreateTelegramAuditLog(_ context.Context, item models.TelegramAuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.audits = append(r.audits, item)
	return nil
}

type authorizationServiceFakeFactory struct {
	session *authorizationServiceFakeSession
}

func (f authorizationServiceFakeFactory) NewAuthorizationSession(uuid.UUID) (AuthorizationSession, error) {
	if f.session == nil {
		return nil, errors.New("fake session unavailable")
	}
	return f.session, nil
}

type authorizationServiceFakeSession struct {
	mu        sync.Mutex
	notify    func(AuthorizationSessionUpdate)
	code      string
	password  string
	promoted  bool
	discarded bool
	closed    bool
}

func (s *authorizationServiceFakeSession) Start(_ context.Context, _ string, notify func(AuthorizationSessionUpdate)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notify = notify
	return nil
}

func (s *authorizationServiceFakeSession) SubmitCode(_ context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.code = code
	return nil
}

func (s *authorizationServiceFakeSession) SubmitPassword(_ context.Context, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.password = password
	return nil
}

func (s *authorizationServiceFakeSession) Promote(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.promoted = true
	return nil
}

func (s *authorizationServiceFakeSession) Discard() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.discarded = true
	return nil
}

func (s *authorizationServiceFakeSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
}

func (s *authorizationServiceFakeSession) emit(update AuthorizationSessionUpdate) {
	s.mu.Lock()
	notify := s.notify
	s.mu.Unlock()
	if notify != nil {
		notify(update)
	}
}

type authorizationServiceFakeMaintenance struct {
	pauseCalls  int
	resumeCalls int
}

func (m *authorizationServiceFakeMaintenance) Pause(context.Context) error {
	m.pauseCalls++
	return nil
}

func (m *authorizationServiceFakeMaintenance) Resume() {
	m.resumeCalls++
}
