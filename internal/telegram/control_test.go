package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestControlHandlerRequiresInternalTokenAndPreservesRequestID(t *testing.T) {
	t.Parallel()

	service := &fakeControlService{}
	handler := NewControlHandler(service, "control-secret")
	server := httptest.NewServer(handler)
	defer server.Close()

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+ControlStatusPath, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	request.Header.Set(ControlTokenHeader, "wrong")
	request.Header.Set(RequestIDHeader, "request-123")
	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatalf("request with wrong token: %v", err)
	}
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong token status = %d, want 401", response.StatusCode)
	}
	_ = response.Body.Close()

	request, err = http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+ControlStatusPath, nil)
	if err != nil {
		t.Fatalf("new authorized request: %v", err)
	}
	request.Header.Set(ControlTokenHeader, "control-secret")
	request.Header.Set(RequestIDHeader, "request-123")
	response, err = server.Client().Do(request)
	if err != nil {
		t.Fatalf("authorized request: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || service.lastRequestID != "request-123" {
		t.Fatalf("status=%d request_id=%q", response.StatusCode, service.lastRequestID)
	}
}

func TestControlHandlerDoesNotEchoAuthorizationSecrets(t *testing.T) {
	t.Parallel()

	handler := NewControlHandler(&fakeControlService{}, "control-secret")
	server := httptest.NewServer(handler)
	defer server.Close()
	body := `{"code":"12345","password":"private-secret"}`
	request, err := http.NewRequest(http.MethodPost, server.URL+"/internal/telegram/authorizations/id/code", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	request.Header.Set(ControlTokenHeader, "control-secret")
	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatalf("secret request: %v", err)
	}
	defer response.Body.Close()
	var raw map[string]any
	if err := json.NewDecoder(response.Body).Decode(&raw); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	encoded, _ := json.Marshal(raw)
	if strings.Contains(string(encoded), "private-secret") || strings.Contains(string(encoded), "12345") {
		t.Fatalf("response echoed authorization secret: %s", encoded)
	}
}

type fakeControlService struct {
	lastRequestID string
}

func (f *fakeControlService) Status(_ context.Context, requestID, _ string) (ControlStatus, error) {
	f.lastRequestID = requestID
	return ControlStatus{}, nil
}

func (f *fakeControlService) StartPhone(context.Context, string, string) (AuthorizationView, error) {
	return AuthorizationView{}, nil
}

func (f *fakeControlService) StartQR(context.Context, string, string) (AuthorizationView, error) {
	return AuthorizationView{}, nil
}

func (f *fakeControlService) GetAuthorization(context.Context, string, string, string) (AuthorizationView, error) {
	return AuthorizationView{}, nil
}

func (f *fakeControlService) SubmitCode(context.Context, string, string, string, string) (AuthorizationView, error) {
	return AuthorizationView{}, nil
}

func (f *fakeControlService) SubmitPassword(context.Context, string, string, string, string) (AuthorizationView, error) {
	return AuthorizationView{}, nil
}

func (f *fakeControlService) CancelAuthorization(context.Context, string, string, string) error {
	return nil
}

func (f *fakeControlService) PreviewChat(context.Context, string, string, string) (ChatPreview, error) {
	return ChatPreview{}, nil
}

func (f *fakeControlService) ConfirmChat(context.Context, string, string, string, string) (ChatPreview, error) {
	return ChatPreview{}, nil
}
