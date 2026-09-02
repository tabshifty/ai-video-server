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
	"video-server/internal/utils"
)

type adminTelegramFakeService struct {
	sources   []models.TelegramSource
	progress  services.TelegramSourceProgress
	addRef    string
	addSource models.TelegramSource
	err       error
}

func (s *adminTelegramFakeService) List(context.Context) ([]models.TelegramSource, error) {
	return s.sources, s.err
}

func (s *adminTelegramFakeService) Add(_ context.Context, chatRef string) (models.TelegramSource, error) {
	s.addRef = chatRef
	return s.addSource, s.err
}

func (s *adminTelegramFakeService) Pause(context.Context, uuid.UUID) (models.TelegramSource, error) {
	return s.addSource, s.err
}

func (s *adminTelegramFakeService) Resume(context.Context, uuid.UUID) (models.TelegramSource, error) {
	return s.addSource, s.err
}

func (s *adminTelegramFakeService) StartBackfill(context.Context, uuid.UUID) (models.TelegramSource, error) {
	return s.addSource, s.err
}

func (s *adminTelegramFakeService) GetProgress(context.Context, uuid.UUID) (services.TelegramSourceProgress, error) {
	return s.progress, s.err
}

func newAdminTelegramContext(method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(method, path, bytes.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	return context, recorder
}

func decodeAdminTelegramEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) apiEnvelope {
	t.Helper()
	var envelope apiEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, recorder.Body.String())
	}
	return envelope
}

func TestAdminTelegramAddSourceAcceptsChatReference(t *testing.T) {
	fake := &adminTelegramFakeService{addSource: models.TelegramSource{ID: uuid.New(), ChatRef: "@group"}}
	api := &API{telegramSourceSvc: fake}
	context, recorder := newAdminTelegramContext(http.MethodPost, "/api/v1/admin/telegram/sources", []byte(`{"chat_ref":" @group "}`))

	api.AdminTelegramAddSource(context)
	envelope := decodeAdminTelegramEnvelope(t, recorder)
	if recorder.Code != http.StatusCreated || envelope.Code != 0 || fake.addRef != " @group " {
		t.Fatalf("envelope = %+v, add ref = %q", envelope, fake.addRef)
	}
}

func TestAdminTelegramListSourcesReturnsServiceErrorEnvelope(t *testing.T) {
	fake := &adminTelegramFakeService{err: errors.New("database unavailable")}
	api := &API{telegramSourceSvc: fake}
	context, recorder := newAdminTelegramContext(http.MethodGet, "/api/v1/admin/telegram/sources", nil)

	api.AdminTelegramSources(context)
	envelope := decodeAdminTelegramEnvelope(t, recorder)
	if envelope.Code == 0 || envelope.Msg != "database unavailable" {
		t.Fatalf("error envelope = %+v", envelope)
	}
}

func TestAdminTelegramProgressReturnsUnifiedEnvelope(t *testing.T) {
	sourceID := uuid.New()
	fake := &adminTelegramFakeService{progress: services.TelegramSourceProgress{
		Source: models.TelegramSource{ID: sourceID},
		Counts: map[string]int{"imported": 2},
	}}
	api := &API{telegramSourceSvc: fake}
	context, recorder := newAdminTelegramContext(http.MethodGet, "/api/v1/admin/telegram/sources/"+sourceID.String()+"/progress", nil)
	context.Params = gin.Params{{Key: "id", Value: sourceID.String()}}

	api.AdminTelegramProgress(context)
	envelope := decodeAdminTelegramEnvelope(t, recorder)
	if envelope.Code != 0 {
		t.Fatalf("progress envelope = %+v", envelope)
	}
}

func TestAdminTelegramRoutesRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const secret = "telegram-route-test-secret"
	api := &API{
		telegramSourceSvc: &adminTelegramFakeService{},
		jwtSecret:         secret,
	}
	router := gin.New()
	api.Register(router)

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/admin/telegram/sources", nil))
	if envelope := decodeAdminTelegramEnvelope(t, unauthorized); envelope.Code != 401 {
		t.Fatalf("unauthorized envelope = %+v", envelope)
	}

	userToken := signAdminTelegramTestToken(t, secret, "user")
	forbiddenRequest := httptest.NewRequest(http.MethodGet, "/api/v1/admin/telegram/sources", nil)
	forbiddenRequest.Header.Set("Authorization", "Bearer "+userToken)
	forbidden := httptest.NewRecorder()
	router.ServeHTTP(forbidden, forbiddenRequest)
	if envelope := decodeAdminTelegramEnvelope(t, forbidden); envelope.Code != 403 {
		t.Fatalf("forbidden envelope = %+v", envelope)
	}

	adminToken := signAdminTelegramTestToken(t, secret, "admin")
	allowedRequest := httptest.NewRequest(http.MethodGet, "/api/v1/admin/telegram/sources", nil)
	allowedRequest.Header.Set("Authorization", "Bearer "+adminToken)
	allowed := httptest.NewRecorder()
	router.ServeHTTP(allowed, allowedRequest)
	if envelope := decodeAdminTelegramEnvelope(t, allowed); envelope.Code != 0 {
		t.Fatalf("allowed envelope = %+v", envelope)
	}
}

func signAdminTelegramTestToken(t *testing.T, secret, role string) string {
	t.Helper()
	now := time.Now().UTC()
	claims := utils.JWTClaims{
		UserID:    uuid.NewString(),
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
