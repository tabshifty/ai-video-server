package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const controlResponseMaxBytes = 1 << 20

var (
	// ErrControlUnavailable indicates that the API process could not reach a
	// valid response from the session-owning collector control service.
	ErrControlUnavailable = errors.New("Telegram 控制通道不可用")
)

// ControlClientConfig configures the API-side client for the collector's
// private control server. BaseURL must point to the internal service address.
type ControlClientConfig struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// ControlRequestError preserves only a sanitized remote error message and its
// HTTP status. It deliberately does not retain request bodies or headers.
type ControlRequestError struct {
	StatusCode int
	Message    string
}

func (e *ControlRequestError) Error() string {
	if e == nil {
		return "Telegram 控制操作失败"
	}
	message := strings.TrimSpace(e.Message)
	if message == "" {
		message = "Telegram 控制操作失败"
	}
	if e.StatusCode <= 0 {
		return message
	}
	return fmt.Sprintf("Telegram 控制操作失败（%d）：%s", e.StatusCode, message)
}

// AsControlRequestError extracts an HTTP error returned by the collector.
func AsControlRequestError(err error, target **ControlRequestError) bool {
	return errors.As(err, target)
}

// ControlClient forwards API management requests to the in-cluster collector.
// It implements ControlService so the API layer can remain independent from
// the transport used by the session-owning process.
type ControlClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

var _ ControlService = (*ControlClient)(nil)

// NewControlClient creates a private-control client with a bounded default
// timeout. Callers may inject an HTTP client for tests or custom transports.
func NewControlClient(config ControlClientConfig) (*ControlClient, error) {
	baseURL, err := normalizeControlBaseURL(config.BaseURL)
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(config.Token)
	if token == "" {
		return nil, errors.New("Telegram 控制令牌不能为空")
	}
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 75 * time.Second}
	}
	return &ControlClient{baseURL: baseURL, token: token, httpClient: client}, nil
}

// Status returns the collector's current control status.
func (c *ControlClient) Status(ctx context.Context, requestID, actorID string) (ControlStatus, error) {
	var status ControlStatus
	if err := c.do(ctx, http.MethodGet, ControlStatusPath, requestID, actorID, nil, &status); err != nil {
		return ControlStatus{}, err
	}
	return status, nil
}

// StartPhone begins a phone-code authorization in the collector process.
func (c *ControlClient) StartPhone(ctx context.Context, requestID, actorID string) (AuthorizationView, error) {
	return c.authorizationRequest(ctx, http.MethodPost, ControlPhoneAuthorizationPath, requestID, actorID, nil)
}

// StartQR begins a QR authorization in the collector process.
func (c *ControlClient) StartQR(ctx context.Context, requestID, actorID string) (AuthorizationView, error) {
	return c.authorizationRequest(ctx, http.MethodPost, ControlQRAuthorizationPath, requestID, actorID, nil)
}

// GetAuthorization reads the public state of an authorization flow.
func (c *ControlClient) GetAuthorization(ctx context.Context, authorizationID, requestID, actorID string) (AuthorizationView, error) {
	return c.authorizationRequest(ctx, http.MethodGet, controlAuthorizationPath(authorizationID), requestID, actorID, nil)
}

// SubmitCode passes a one-time code directly to the collector without storing
// it in API state or error values.
func (c *ControlClient) SubmitCode(ctx context.Context, authorizationID, requestID, actorID, code string) (AuthorizationView, error) {
	return c.authorizationRequest(ctx, http.MethodPost, controlAuthorizationPath(authorizationID, "code"), requestID, actorID, struct {
		Code string `json:"code"`
	}{Code: code})
}

// SubmitPassword passes a Telegram 2FA password directly to the collector.
func (c *ControlClient) SubmitPassword(ctx context.Context, authorizationID, requestID, actorID, password string) (AuthorizationView, error) {
	return c.authorizationRequest(ctx, http.MethodPost, controlAuthorizationPath(authorizationID, "password"), requestID, actorID, struct {
		Password string `json:"password"`
	}{Password: password})
}

// CancelAuthorization asks the collector to stop an owned authorization flow.
func (c *ControlClient) CancelAuthorization(ctx context.Context, authorizationID, requestID, actorID string) error {
	return c.do(ctx, http.MethodPost, controlAuthorizationPath(authorizationID, "cancel"), requestID, actorID, nil, nil)
}

// PreviewChat resolves a source reference without adding it to storage.
func (c *ControlClient) PreviewChat(ctx context.Context, requestID, actorID, chatRef string) (ChatPreview, error) {
	var preview ChatPreview
	err := c.do(ctx, http.MethodPost, ControlChatPreviewPath, requestID, actorID, struct {
		ChatRef string `json:"chat_ref"`
	}{ChatRef: chatRef}, &preview)
	if err != nil {
		return ChatPreview{}, err
	}
	return preview, nil
}

// ConfirmChat completes an administrator-confirmed source lookup.
func (c *ControlClient) ConfirmChat(ctx context.Context, requestID, actorID, previewID, chatRef string) (ChatPreview, error) {
	var preview ChatPreview
	err := c.do(ctx, http.MethodPost, ControlChatConfirmPath, requestID, actorID, struct {
		PreviewID string `json:"preview_id"`
		ChatRef   string `json:"chat_ref"`
	}{PreviewID: previewID, ChatRef: chatRef}, &preview)
	if err != nil {
		return ChatPreview{}, err
	}
	return preview, nil
}

func (c *ControlClient) authorizationRequest(ctx context.Context, method, path, requestID, actorID string, input any) (AuthorizationView, error) {
	var view AuthorizationView
	if err := c.do(ctx, method, path, requestID, actorID, input, &view); err != nil {
		return AuthorizationView{}, err
	}
	return view, nil
}

func (c *ControlClient) do(ctx context.Context, method, path, requestID, actorID string, input, output any) error {
	if c == nil || c.httpClient == nil || strings.TrimSpace(c.baseURL) == "" || strings.TrimSpace(c.token) == "" {
		return ErrControlUnavailable
	}
	if ctx == nil {
		return errors.New("Telegram 控制请求 context 不能为空")
	}
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return errors.New("编码 Telegram 控制请求失败")
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("%w: 创建请求失败", ErrControlUnavailable)
	}
	req.Header.Set(ControlTokenHeader, c.token)
	if requestID = strings.TrimSpace(requestID); requestID != "" {
		req.Header.Set(RequestIDHeader, requestID)
	}
	if actorID = strings.TrimSpace(actorID); actorID != "" {
		req.Header.Set(ControlActorHeader, actorID)
	}
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: 请求失败", ErrControlUnavailable)
	}
	defer response.Body.Close()

	var envelope controlClientEnvelope
	decoder := json.NewDecoder(io.LimitReader(response.Body, controlResponseMaxBytes))
	if err := decoder.Decode(&envelope); err != nil {
		return fmt.Errorf("%w: 响应格式无效", ErrControlUnavailable)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !envelope.OK {
		return &ControlRequestError{StatusCode: response.StatusCode, Message: envelope.Error}
	}
	if output == nil {
		return nil
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return fmt.Errorf("%w: 响应缺少数据", ErrControlUnavailable)
	}
	if err := json.Unmarshal(envelope.Data, output); err != nil {
		return fmt.Errorf("%w: 响应数据无效", ErrControlUnavailable)
	}
	return nil
}

type controlClientEnvelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
}

func normalizeControlBaseURL(value string) (string, error) {
	value = strings.TrimSuffix(strings.TrimSpace(value), "/")
	if value == "" {
		return "", errors.New("Telegram 控制地址不能为空")
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", errors.New("Telegram 控制地址必须是有效的 HTTP 地址")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("Telegram 控制地址不能包含查询参数或片段")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func controlAuthorizationPath(authorizationID string, suffix ...string) string {
	path := ControlBasePath + "/authorizations/" + url.PathEscape(strings.TrimSpace(authorizationID))
	for _, part := range suffix {
		path += "/" + url.PathEscape(strings.TrimSpace(part))
	}
	return path
}
