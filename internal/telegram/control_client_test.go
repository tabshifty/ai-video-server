package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestControlClientForwardsAuthorizationRequestWithPrivateHeaders(t *testing.T) {
	t.Parallel()

	actorID := uuid.NewString()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != ControlPhoneAuthorizationPath {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get(ControlTokenHeader); got != "control-token" {
			t.Fatalf("control token = %q", got)
		}
		if got := r.Header.Get(RequestIDHeader); got != "request-123" {
			t.Fatalf("request ID = %q", got)
		}
		if got := r.Header.Get(ControlActorHeader); got != actorID {
			t.Fatalf("actor ID = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
			"data": AuthorizationView{
				ID:        uuid.New(),
				Kind:      AuthorizationKindPhone,
				Status:    AuthorizationStatusAwaitingCode,
				ExpiresAt: time.Date(2026, time.September, 4, 9, 0, 0, 0, time.UTC),
			},
		})
	}))
	defer server.Close()

	client, err := NewControlClient(ControlClientConfig{
		BaseURL:    server.URL,
		Token:      "control-token",
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("NewControlClient() error = %v", err)
	}
	view, err := client.StartPhone(context.Background(), "request-123", actorID)
	if err != nil {
		t.Fatalf("StartPhone() error = %v", err)
	}
	if view.Kind != AuthorizationKindPhone || view.Status != AuthorizationStatusAwaitingCode || view.ID == uuid.Nil {
		t.Fatalf("authorization view = %+v", view)
	}
}

func TestControlClientReturnsSanitizedRemoteError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != ControlBasePath+"/authorizations/test/password" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": "Telegram 二次验证密码错误",
		})
	}))
	defer server.Close()

	client, err := NewControlClient(ControlClientConfig{BaseURL: server.URL, Token: "control-token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("NewControlClient() error = %v", err)
	}
	_, err = client.SubmitPassword(context.Background(), "test", "request-123", uuid.NewString(), "actual-secret")
	if err == nil {
		t.Fatal("SubmitPassword() error = nil")
	}
	var controlErr *ControlRequestError
	if !AsControlRequestError(err, &controlErr) || controlErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("SubmitPassword() error = %T %v", err, err)
	}
	if strings.Contains(err.Error(), "actual-secret") {
		t.Fatalf("error leaked password: %v", err)
	}
}

func TestNewControlClientRejectsIncompleteConfiguration(t *testing.T) {
	t.Parallel()

	for _, config := range []ControlClientConfig{
		{Token: "control-token"},
		{BaseURL: "http://telegram-ingestor:8091"},
		{BaseURL: "ftp://telegram-ingestor:8091", Token: "control-token"},
	} {
		if _, err := NewControlClient(config); err == nil {
			t.Fatalf("NewControlClient(%+v) error = nil", config)
		}
	}
}
