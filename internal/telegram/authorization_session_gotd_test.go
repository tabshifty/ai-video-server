package telegram

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
)

func TestGotdAuthorizationSessionFactoryUsesSiblingTemporarySession(t *testing.T) {
	t.Parallel()

	sessionPath := filepath.Join(t.TempDir(), "collector.session")
	factory, err := NewGotdAuthorizationSessionFactory(GotdAuthorizationSessionFactoryConfig{
		APIID:       12345,
		APIHash:     "api-hash",
		Phone:       "+8613800000000",
		SessionPath: sessionPath,
	})
	if err != nil {
		t.Fatalf("NewGotdAuthorizationSessionFactory() error = %v", err)
	}
	authorizationID := uuid.New()
	session, err := factory.NewAuthorizationSession(authorizationID)
	if err != nil {
		t.Fatalf("NewAuthorizationSession() error = %v", err)
	}
	got, ok := session.(*gotdAuthorizationSession)
	if !ok {
		t.Fatalf("session type = %T, want *gotdAuthorizationSession", session)
	}
	if filepath.Dir(got.temporarySessionPath) != filepath.Dir(sessionPath) {
		t.Fatalf("temporary session must share final directory: temporary=%q final=%q", got.temporarySessionPath, sessionPath)
	}
	if !strings.Contains(filepath.Base(got.temporarySessionPath), authorizationID.String()) {
		t.Fatalf("temporary session path must include authorization ID: %q", got.temporarySessionPath)
	}
}

func TestPromoteAuthorizationSessionReplacesPersistentSession(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	persistentPath := filepath.Join(directory, "collector.session")
	temporaryPath := filepath.Join(directory, "collector.pending")
	if err := os.WriteFile(persistentPath, []byte("previous-session"), 0o600); err != nil {
		t.Fatalf("write persistent session: %v", err)
	}
	if err := os.WriteFile(temporaryPath, []byte("validated-session"), 0o600); err != nil {
		t.Fatalf("write temporary session: %v", err)
	}

	if err := promoteAuthorizationSession(temporaryPath, persistentPath); err != nil {
		t.Fatalf("promoteAuthorizationSession() error = %v", err)
	}
	data, err := os.ReadFile(persistentPath)
	if err != nil {
		t.Fatalf("read promoted session: %v", err)
	}
	if string(data) != "validated-session" {
		t.Fatalf("promoted session = %q, want validated session", data)
	}
	if _, err := os.Stat(temporaryPath); !os.IsNotExist(err) {
		t.Fatalf("temporary session remains after promotion: %v", err)
	}
}

func TestRenderTelegramQRDataURLDoesNotExposeRawToken(t *testing.T) {
	t.Parallel()

	const rawToken = "temporary-login-token"
	dataURL, err := renderTelegramQRDataURL(qrlogin.NewToken([]byte(rawToken), int(time.Now().Add(time.Minute).Unix())))
	if err != nil {
		t.Fatalf("renderTelegramQRDataURL() error = %v", err)
	}
	if !strings.HasPrefix(dataURL, "data:image/png;base64,") {
		t.Fatalf("QR image = %q, want PNG data URL", dataURL)
	}
	if strings.Contains(dataURL, rawToken) {
		t.Fatalf("QR data URL contains raw token")
	}
}

func TestTelegramIdentityFromAuthorization(t *testing.T) {
	t.Parallel()

	identity, err := telegramIdentityFromAuthorization(&tg.AuthAuthorization{User: &tg.User{
		ID:        701,
		Username:  "collector",
		FirstName: "采集",
		LastName:  "账号",
	}})
	if err != nil {
		t.Fatalf("telegramIdentityFromAuthorization() error = %v", err)
	}
	if identity != (TelegramIdentity{ID: 701, Username: "collector", FirstName: "采集", LastName: "账号"}) {
		t.Fatalf("identity = %+v", identity)
	}
}
