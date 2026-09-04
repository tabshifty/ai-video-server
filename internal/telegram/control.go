package telegram

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/google/uuid"
)

const (
	// ControlBasePath is the private HTTP namespace exposed only to the API container.
	ControlBasePath = "/internal/telegram"
	// ControlStatusPath returns collector and account status.
	ControlStatusPath = ControlBasePath + "/status"
	// ControlPhoneAuthorizationPath starts a phone authorization flow.
	ControlPhoneAuthorizationPath = ControlBasePath + "/authorizations/phone"
	// ControlQRAuthorizationPath starts a QR authorization flow.
	ControlQRAuthorizationPath = ControlBasePath + "/authorizations/qr"
	// ControlChatPreviewPath resolves a chat before it is added.
	ControlChatPreviewPath = ControlBasePath + "/chats/preview"
	// ControlChatConfirmPath confirms a previously previewed chat.
	ControlChatConfirmPath = ControlBasePath + "/chats/confirm"

	// ControlTokenHeader authenticates API-to-ingestor requests.
	ControlTokenHeader = "X-Telegram-Control-Token"
	// RequestIDHeader carries the caller's correlation identifier.
	RequestIDHeader = "X-Request-ID"
	// ControlActorHeader identifies the administrator who owns a flow.
	ControlActorHeader = "X-Telegram-Actor-ID"
)

// ControlStatus is the structured collector status returned to the management API.
type ControlStatus struct {
	RequestID      string             `json:"request_id,omitempty"`
	IngestorStatus string             `json:"ingestor_status"`
	AccountStatus  string             `json:"account_status"`
	Account        map[string]any     `json:"account,omitempty"`
	Heartbeat      map[string]any     `json:"heartbeat,omitempty"`
	Authorization  *AuthorizationView `json:"authorization,omitempty"`
	Sources        []map[string]any   `json:"sources,omitempty"`
	Error          string             `json:"error,omitempty"`
}

// ChatPreview is the sanitized result of resolving a Telegram chat reference.
type ChatPreview struct {
	PreviewID string `json:"preview_id,omitempty"`
	ChatID    int64  `json:"chat_id"`
	Title     string `json:"title"`
	Username  string `json:"username,omitempty"`
	ChatType  string `json:"chat_type"`
	ChatRef   string `json:"chat_ref,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// ControlService is implemented by the session-owning ingestor process.
// Secrets are accepted only as arguments and must never be persisted.
type ControlService interface {
	Status(ctx context.Context, requestID, actorID string) (ControlStatus, error)
	StartPhone(ctx context.Context, requestID, actorID string) (AuthorizationView, error)
	StartQR(ctx context.Context, requestID, actorID string) (AuthorizationView, error)
	GetAuthorization(ctx context.Context, authorizationID, requestID, actorID string) (AuthorizationView, error)
	SubmitCode(ctx context.Context, authorizationID, requestID, actorID, code string) (AuthorizationView, error)
	SubmitPassword(ctx context.Context, authorizationID, requestID, actorID, password string) (AuthorizationView, error)
	CancelAuthorization(ctx context.Context, authorizationID, requestID, actorID string) error
	PreviewChat(ctx context.Context, requestID, actorID, chatRef string) (ChatPreview, error)
	ConfirmChat(ctx context.Context, requestID, actorID, previewID, chatRef string) (ChatPreview, error)
}

// NewControlHandler creates the private ingestor control HTTP handler.
func NewControlHandler(service ControlService, token string) http.Handler {
	h := &controlHandler{service: service, token: strings.TrimSpace(token), sequence: new(uint64)}
	return h
}

type controlHandler struct {
	service  ControlService
	token    string
	sequence *uint64
}

func (h *controlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.service == nil {
		writeControlError(w, http.StatusServiceUnavailable, "Telegram 控制服务不可用")
		return
	}
	if !constantTimeTokenEqual(r.Header.Get(ControlTokenHeader), h.token) {
		writeControlError(w, http.StatusUnauthorized, "Telegram 控制通道认证失败")
		return
	}
	requestID := strings.TrimSpace(r.Header.Get(RequestIDHeader))
	if requestID == "" {
		requestID = h.nextRequestID()
	}
	w.Header().Set(RequestIDHeader, requestID)
	actorID := strings.TrimSpace(r.Header.Get(ControlActorHeader))
	path := strings.TrimRight(r.URL.Path, "/")

	switch {
	case r.Method == http.MethodGet && path == ControlStatusPath:
		status, err := h.service.Status(r.Context(), requestID, actorID)
		h.writeResult(w, status, err)
	case r.Method == http.MethodPost && path == ControlPhoneAuthorizationPath:
		view, err := h.service.StartPhone(r.Context(), requestID, actorID)
		h.writeResult(w, view, err)
	case r.Method == http.MethodPost && path == ControlQRAuthorizationPath:
		view, err := h.service.StartQR(r.Context(), requestID, actorID)
		h.writeResult(w, view, err)
	case r.Method == http.MethodPost && path == ControlChatPreviewPath:
		var input chatPreviewRequest
		if !h.decode(w, r, &input) {
			return
		}
		preview, err := h.service.PreviewChat(r.Context(), requestID, actorID, input.ChatRef)
		h.writeResult(w, preview, err)
	case r.Method == http.MethodPost && path == ControlChatConfirmPath:
		var input chatConfirmRequest
		if !h.decode(w, r, &input) {
			return
		}
		preview, err := h.service.ConfirmChat(r.Context(), requestID, actorID, input.PreviewID, input.ChatRef)
		h.writeResult(w, preview, err)
	case strings.HasPrefix(path, ControlBasePath+"/authorizations/"):
		h.handleAuthorization(w, r, path, requestID, actorID)
	default:
		writeControlError(w, http.StatusNotFound, "Telegram 控制接口不存在")
	}
}

type codeRequest struct {
	Code string `json:"code"`
}

type passwordRequest struct {
	Password string `json:"password"`
}

type chatPreviewRequest struct {
	ChatRef string `json:"chat_ref"`
}

type chatConfirmRequest struct {
	PreviewID string `json:"preview_id"`
	ChatRef   string `json:"chat_ref"`
}

func (h *controlHandler) handleAuthorization(w http.ResponseWriter, r *http.Request, path, requestID, actorID string) {
	parts := strings.Split(strings.TrimPrefix(path, ControlBasePath+"/authorizations/"), "/")
	if len(parts) < 1 || strings.TrimSpace(parts[0]) == "" {
		writeControlError(w, http.StatusNotFound, "Telegram 授权接口不存在")
		return
	}
	authorizationID := strings.TrimSpace(parts[0])
	switch {
	case r.Method == http.MethodGet && len(parts) == 1:
		view, err := h.service.GetAuthorization(r.Context(), authorizationID, requestID, actorID)
		h.writeResult(w, view, err)
	case r.Method == http.MethodPost && len(parts) == 2 && parts[1] == "code":
		var input codeRequest
		if !h.decode(w, r, &input) {
			return
		}
		if strings.TrimSpace(input.Code) == "" {
			writeControlError(w, http.StatusBadRequest, "Telegram 验证码不能为空")
			return
		}
		view, err := h.service.SubmitCode(r.Context(), authorizationID, requestID, actorID, input.Code)
		h.writeResult(w, view, err)
	case r.Method == http.MethodPost && len(parts) == 2 && parts[1] == "password":
		var input passwordRequest
		if !h.decode(w, r, &input) {
			return
		}
		if strings.TrimSpace(input.Password) == "" {
			writeControlError(w, http.StatusBadRequest, "Telegram 二次验证密码不能为空")
			return
		}
		view, err := h.service.SubmitPassword(r.Context(), authorizationID, requestID, actorID, input.Password)
		h.writeResult(w, view, err)
	case r.Method == http.MethodPost && len(parts) == 2 && parts[1] == "cancel":
		err := h.service.CancelAuthorization(r.Context(), authorizationID, requestID, actorID)
		h.writeResult(w, map[string]bool{"cancelled": err == nil}, err)
	default:
		writeControlError(w, http.StatusNotFound, "Telegram 授权接口不存在")
	}
}

func (h *controlHandler) decode(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeControlError(w, http.StatusBadRequest, "Telegram 控制请求格式错误")
		return false
	}
	return true
}

func (h *controlHandler) writeResult(w http.ResponseWriter, value any, err error) {
	if err != nil {
		writeControlError(w, controlStatusForError(err), sanitizeControlError(err))
		return
	}
	writeControlJSON(w, http.StatusOK, map[string]any{"ok": true, "data": value})
}

func (h *controlHandler) nextRequestID() string {
	value := atomic.AddUint64(h.sequence, 1)
	return fmt.Sprintf("telegram-control-%d", value)
}

func constantTimeTokenEqual(provided, expected string) bool {
	provided = strings.TrimSpace(provided)
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return false
	}
	providedBytes := []byte(provided)
	expectedBytes := []byte(expected)
	if len(providedBytes) != len(expectedBytes) {
		// Compare against a same-length buffer so the result still does not expose
		// the expected token through an early byte comparison.
		padded := make([]byte, len(expectedBytes))
		copy(padded, providedBytes)
		_ = subtle.ConstantTimeCompare(padded, expectedBytes)
		return false
	}
	return subtle.ConstantTimeCompare(providedBytes, expectedBytes) == 1
}

func writeControlJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}

func writeControlError(w http.ResponseWriter, status int, message string) {
	writeControlJSON(w, status, map[string]any{
		"ok":    false,
		"error": strings.TrimSpace(message),
	})
}

func controlStatusForError(err error) int {
	switch {
	case errors.Is(err, ErrAuthorizationInProgress):
		return http.StatusConflict
	case errors.Is(err, ErrAuthorizationNotOwner):
		return http.StatusForbidden
	case errors.Is(err, ErrAuthorizationNotFound):
		return http.StatusNotFound
	default:
		return http.StatusBadRequest
	}
}

func sanitizeControlError(err error) string {
	if err == nil {
		return ""
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "Telegram 控制操作失败"
	}
	// Service implementations should return sanitized errors. This final layer
	// avoids accidentally reflecting common secret field values in a response.
	for _, marker := range []string{"code=", "password=", "phone_code=", "phone_password="} {
		if index := strings.Index(strings.ToLower(message), marker); index >= 0 {
			message = strings.TrimSpace(message[:index])
		}
	}
	if message == "" {
		return "Telegram 控制操作失败"
	}
	return message
}

// ParseControlAuthorizationID validates UUID IDs at service boundaries that use UUIDs.
func ParseControlAuthorizationID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, errors.New("Telegram 授权 ID 格式错误")
	}
	return id, nil
}
