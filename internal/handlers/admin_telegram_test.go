package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/services"
	"video-server/internal/telegram"
	"video-server/internal/utils"
)

type adminTelegramManagementFakeService struct {
	status        telegram.ControlStatus
	confirmation  services.TelegramConfirmationView
	authorization telegram.AuthorizationView
	preview       telegram.ChatPreview
	source        models.TelegramSource
	sources       []models.TelegramSource
	progress      services.TelegramSourceProgress
	audits        []models.TelegramAuditLog
	err           error

	actorID              uuid.UUID
	confirmationAction   string
	confirmationPassword string
	confirmationAccepted bool
	authorizationKind    string
	authorizationID      string
	confirmationTicket   string
	code                 string
	password             string
	chatRef              string
	previewID            string
	sourceAction         string
	sourceID             uuid.UUID
}

func (s *adminTelegramManagementFakeService) Status(_ context.Context, actorID uuid.UUID, _ string) (telegram.ControlStatus, error) {
	s.actorID = actorID
	return s.status, s.err
}

func (s *adminTelegramManagementFakeService) IssueConfirmation(_ context.Context, actorID uuid.UUID, action, password string, confirmed bool) (services.TelegramConfirmationView, error) {
	s.actorID = actorID
	s.confirmationAction = action
	s.confirmationPassword = password
	s.confirmationAccepted = confirmed
	return s.confirmation, s.err
}

func (s *adminTelegramManagementFakeService) StartPhone(_ context.Context, actorID uuid.UUID, _ string, ticket string) (telegram.AuthorizationView, error) {
	s.actorID = actorID
	s.authorizationKind = telegram.AuthorizationKindPhone
	s.confirmationTicket = ticket
	return s.authorization, s.err
}

func (s *adminTelegramManagementFakeService) StartQR(_ context.Context, actorID uuid.UUID, _ string, ticket string) (telegram.AuthorizationView, error) {
	s.actorID = actorID
	s.authorizationKind = telegram.AuthorizationKindQR
	s.confirmationTicket = ticket
	return s.authorization, s.err
}

func (s *adminTelegramManagementFakeService) GetAuthorization(_ context.Context, actorID uuid.UUID, authorizationID, _ string) (telegram.AuthorizationView, error) {
	s.actorID = actorID
	s.authorizationID = authorizationID
	return s.authorization, s.err
}

func (s *adminTelegramManagementFakeService) SubmitCode(_ context.Context, actorID uuid.UUID, authorizationID, _ string, code string) (telegram.AuthorizationView, error) {
	s.actorID = actorID
	s.authorizationID = authorizationID
	s.code = code
	return s.authorization, s.err
}

func (s *adminTelegramManagementFakeService) SubmitPassword(_ context.Context, actorID uuid.UUID, authorizationID, _ string, password string) (telegram.AuthorizationView, error) {
	s.actorID = actorID
	s.authorizationID = authorizationID
	s.password = password
	return s.authorization, s.err
}

func (s *adminTelegramManagementFakeService) CancelAuthorization(_ context.Context, actorID uuid.UUID, authorizationID, _ string) error {
	s.actorID = actorID
	s.authorizationID = authorizationID
	return s.err
}

func (s *adminTelegramManagementFakeService) PreviewSource(_ context.Context, actorID uuid.UUID, _ string, chatRef, ticket string) (telegram.ChatPreview, error) {
	s.actorID = actorID
	s.chatRef = chatRef
	s.confirmationTicket = ticket
	return s.preview, s.err
}

func (s *adminTelegramManagementFakeService) ConfirmSource(_ context.Context, actorID uuid.UUID, _ string, previewID, chatRef, ticket string) (models.TelegramSource, error) {
	s.actorID = actorID
	s.previewID = previewID
	s.chatRef = chatRef
	s.confirmationTicket = ticket
	return s.source, s.err
}

func (s *adminTelegramManagementFakeService) ListSources(context.Context) ([]models.TelegramSource, error) {
	return s.sources, s.err
}

func (s *adminTelegramManagementFakeService) PauseSource(_ context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	s.actorID = actorID
	s.sourceAction = "pause"
	s.sourceID = sourceID
	return s.source, s.err
}

func (s *adminTelegramManagementFakeService) ResumeSource(_ context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	s.actorID = actorID
	s.sourceAction = "resume"
	s.sourceID = sourceID
	return s.source, s.err
}

func (s *adminTelegramManagementFakeService) StartBackfill(_ context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	s.actorID = actorID
	s.sourceAction = "backfill"
	s.sourceID = sourceID
	return s.source, s.err
}

func (s *adminTelegramManagementFakeService) RecoverSource(_ context.Context, actorID, sourceID uuid.UUID) (models.TelegramSource, error) {
	s.actorID = actorID
	s.sourceAction = "recover"
	s.sourceID = sourceID
	return s.source, s.err
}

func (s *adminTelegramManagementFakeService) GetProgress(context.Context, uuid.UUID) (services.TelegramSourceProgress, error) {
	return s.progress, s.err
}

func (s *adminTelegramManagementFakeService) ListAudits(context.Context, models.TelegramAuditFilter) ([]models.TelegramAuditLog, int, error) {
	return s.audits, len(s.audits), s.err
}

func newAdminTelegramManagementRouter(service telegramManagementService) (*gin.Engine, string) {
	gin.SetMode(gin.TestMode)
	const secret = "telegram-route-test-secret"
	api := &API{telegramManagementSvc: service, jwtSecret: secret}
	router := gin.New()
	api.Register(router)
	return router, secret
}

func newAdminTelegramRequest(method, target, token string, body []byte) *http.Request {
	request := httptest.NewRequest(method, target, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	return request
}

func decodeAdminTelegramEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) apiEnvelope {
	t.Helper()
	var envelope apiEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, recorder.Body.String())
	}
	return envelope
}

func TestAdminTelegramManagementRoutesRequireAdminAndRejectDirectSourceCreate(t *testing.T) {
	service := &adminTelegramManagementFakeService{status: telegram.ControlStatus{IngestorStatus: "running"}}
	router, secret := newAdminTelegramManagementRouter(service)

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, newAdminTelegramRequest(http.MethodGet, "/api/v1/admin/telegram/status", "", nil))
	if envelope := decodeAdminTelegramEnvelope(t, unauthorized); envelope.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized envelope = %+v", envelope)
	}

	userToken := signAdminTelegramTestToken(t, secret, uuid.New(), "user")
	forbidden := httptest.NewRecorder()
	router.ServeHTTP(forbidden, newAdminTelegramRequest(http.MethodGet, "/api/v1/admin/telegram/status", userToken, nil))
	if envelope := decodeAdminTelegramEnvelope(t, forbidden); envelope.Code != http.StatusForbidden {
		t.Fatalf("forbidden envelope = %+v", envelope)
	}

	actorID := uuid.New()
	adminToken := signAdminTelegramTestToken(t, secret, actorID, "admin")
	allowed := httptest.NewRecorder()
	router.ServeHTTP(allowed, newAdminTelegramRequest(http.MethodGet, "/api/v1/admin/telegram/status", adminToken, nil))
	if envelope := decodeAdminTelegramEnvelope(t, allowed); envelope.Code != 0 || service.actorID != actorID {
		t.Fatalf("allowed envelope = %+v actor=%s", envelope, service.actorID)
	}

	directCreate := httptest.NewRecorder()
	router.ServeHTTP(directCreate, newAdminTelegramRequest(http.MethodPost, "/api/v1/admin/telegram/sources", adminToken, []byte(`{"chat_ref":"@not-allowed"}`)))
	if directCreate.Code != http.StatusNotFound {
		t.Fatalf("direct source create status = %d, want 404", directCreate.Code)
	}
}

func TestAdminTelegramManagementAuthorizationRoutesForwardTransientValues(t *testing.T) {
	authorizationID := uuid.New()
	service := &adminTelegramManagementFakeService{
		confirmation:  services.TelegramConfirmationView{Ticket: "confirmation-ticket", Action: services.TelegramConfirmationActionPhoneAuthorization},
		authorization: telegram.AuthorizationView{ID: authorizationID, Kind: telegram.AuthorizationKindPhone, Status: telegram.AuthorizationStatusAwaitingCode},
	}
	router, secret := newAdminTelegramManagementRouter(service)
	actorID := uuid.New()
	token := signAdminTelegramTestToken(t, secret, actorID, "admin")

	confirmation := httptest.NewRecorder()
	router.ServeHTTP(confirmation, newAdminTelegramRequest(http.MethodPost, "/api/v1/admin/telegram/confirmations", token, []byte(`{"action":"authorization_phone","current_password":"admin-password","confirmed":true}`)))
	if envelope := decodeAdminTelegramEnvelope(t, confirmation); envelope.Code != 0 || service.confirmationAction != services.TelegramConfirmationActionPhoneAuthorization || service.confirmationPassword != "admin-password" || !service.confirmationAccepted {
		t.Fatalf("confirmation envelope=%+v service=%+v", envelope, service)
	}

	start := httptest.NewRecorder()
	router.ServeHTTP(start, newAdminTelegramRequest(http.MethodPost, "/api/v1/admin/telegram/authorizations/phone", token, []byte(`{"confirmation_ticket":"confirmation-ticket"}`)))
	if envelope := decodeAdminTelegramEnvelope(t, start); envelope.Code != 0 || service.authorizationKind != telegram.AuthorizationKindPhone || service.confirmationTicket != "confirmation-ticket" || service.actorID != actorID {
		t.Fatalf("start envelope=%+v service=%+v", envelope, service)
	}

	code := httptest.NewRecorder()
	router.ServeHTTP(code, newAdminTelegramRequest(http.MethodPost, "/api/v1/admin/telegram/authorizations/"+authorizationID.String()+"/code", token, []byte(`{"code":"12345"}`)))
	if envelope := decodeAdminTelegramEnvelope(t, code); envelope.Code != 0 || service.authorizationID != authorizationID.String() || service.code != "12345" {
		t.Fatalf("code envelope=%+v service=%+v", envelope, service)
	}

	password := httptest.NewRecorder()
	router.ServeHTTP(password, newAdminTelegramRequest(http.MethodPost, "/api/v1/admin/telegram/authorizations/"+authorizationID.String()+"/password", token, []byte(`{"password":"telegram-2fa"}`)))
	if envelope := decodeAdminTelegramEnvelope(t, password); envelope.Code != 0 || service.password != "telegram-2fa" {
		t.Fatalf("password envelope=%+v service=%+v", envelope, service)
	}

	cancel := httptest.NewRecorder()
	router.ServeHTTP(cancel, newAdminTelegramRequest(http.MethodPost, "/api/v1/admin/telegram/authorizations/"+authorizationID.String()+"/cancel", token, nil))
	if envelope := decodeAdminTelegramEnvelope(t, cancel); envelope.Code != 0 || service.authorizationID != authorizationID.String() {
		t.Fatalf("cancel envelope=%+v service=%+v", envelope, service)
	}
}

func TestAdminTelegramManagementSourceRoutesUsePreviewConfirmFlow(t *testing.T) {
	sourceID := uuid.New()
	service := &adminTelegramManagementFakeService{
		preview:  telegram.ChatPreview{PreviewID: "preview-1", ChatID: -100123, Title: "私密频道", ChatType: telegram.TelegramChatTypeChannel},
		source:   models.TelegramSource{ID: sourceID, ChatID: -100123, ChatRef: "-100123", Title: "私密频道"},
		sources:  []models.TelegramSource{{ID: sourceID, ChatID: -100123, ChatRef: "-100123"}},
		progress: services.TelegramSourceProgress{Source: models.TelegramSource{ID: sourceID}, Counts: map[string]int{"imported": 2}},
		audits:   []models.TelegramAuditLog{{Action: "source.created"}},
	}
	router, secret := newAdminTelegramManagementRouter(service)
	token := signAdminTelegramTestToken(t, secret, uuid.New(), "admin")
	const invite = "https://t.me/+privateInviteToken"

	preview := httptest.NewRecorder()
	router.ServeHTTP(preview, newAdminTelegramRequest(http.MethodPost, "/api/v1/admin/telegram/sources/preview", token, []byte(`{"chat_ref":"`+invite+`","confirmation_ticket":"preview-ticket"}`)))
	if envelope := decodeAdminTelegramEnvelope(t, preview); envelope.Code != 0 || service.chatRef != invite || service.confirmationTicket != "preview-ticket" {
		t.Fatalf("preview envelope=%+v service=%+v", envelope, service)
	}

	confirm := httptest.NewRecorder()
	router.ServeHTTP(confirm, newAdminTelegramRequest(http.MethodPost, "/api/v1/admin/telegram/sources/confirm", token, []byte(`{"preview_id":"preview-1","chat_ref":"`+invite+`","confirmation_ticket":"confirm-ticket"}`)))
	if envelope := decodeAdminTelegramEnvelope(t, confirm); confirm.Code != http.StatusCreated || envelope.Code != 0 || service.previewID != "preview-1" || service.chatRef != invite || service.confirmationTicket != "confirm-ticket" {
		t.Fatalf("confirm envelope=%+v service=%+v", envelope, service)
	}

	list := httptest.NewRecorder()
	router.ServeHTTP(list, newAdminTelegramRequest(http.MethodGet, "/api/v1/admin/telegram/sources", token, nil))
	if envelope := decodeAdminTelegramEnvelope(t, list); envelope.Code != 0 {
		t.Fatalf("list envelope=%+v", envelope)
	}

	for _, operation := range []struct {
		path   string
		action string
	}{
		{path: "/pause", action: "pause"},
		{path: "/resume", action: "resume"},
		{path: "/backfill", action: "backfill"},
		{path: "/recover", action: "recover"},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, newAdminTelegramRequest(http.MethodPost, "/api/v1/admin/telegram/sources/"+sourceID.String()+operation.path, token, nil))
		if envelope := decodeAdminTelegramEnvelope(t, response); envelope.Code != 0 || service.sourceAction != operation.action || service.sourceID != sourceID {
			t.Fatalf("%s envelope=%+v service=%+v", operation.action, envelope, service)
		}
	}

	progress := httptest.NewRecorder()
	router.ServeHTTP(progress, newAdminTelegramRequest(http.MethodGet, "/api/v1/admin/telegram/sources/"+sourceID.String()+"/progress", token, nil))
	if envelope := decodeAdminTelegramEnvelope(t, progress); envelope.Code != 0 {
		t.Fatalf("progress envelope=%+v", envelope)
	}

	audits := httptest.NewRecorder()
	router.ServeHTTP(audits, newAdminTelegramRequest(http.MethodGet, "/api/v1/admin/telegram/audits?page=1&page_size=20", token, nil))
	if envelope := decodeAdminTelegramEnvelope(t, audits); envelope.Code != 0 {
		t.Fatalf("audits envelope=%+v", envelope)
	}
}

func TestAdminTelegramManagementUsesUnifiedErrors(t *testing.T) {
	service := &adminTelegramManagementFakeService{err: errors.New("collector unavailable")}
	router, secret := newAdminTelegramManagementRouter(service)
	token := signAdminTelegramTestToken(t, secret, uuid.New(), "admin")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminTelegramRequest(http.MethodGet, "/api/v1/admin/telegram/status", token, nil))
	if envelope := decodeAdminTelegramEnvelope(t, response); envelope.Code != telegramAdminErrorCode || envelope.Msg != "collector unavailable" {
		t.Fatalf("error envelope=%+v", envelope)
	}
}

func signAdminTelegramTestToken(t *testing.T, secret string, userID uuid.UUID, role string) string {
	t.Helper()
	now := time.Now().UTC()
	claims := utils.JWTClaims{
		UserID:    userID.String(),
		Role:      role,
		TokenType: utils.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}
