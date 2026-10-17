package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

// Browser is a minimal Chrome DevTools Protocol client bound to one endpoint.
// It only performs the three operations this feature needs: list targets,
// evaluate an expression in a page, and capture a screenshot.
type Browser struct {
	baseURL string
	http    *http.Client
}

// Target describes one CDP page target.
type Target struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"`
	Title                string `json:"title"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// BrowserVersion is the /json/version payload subset we care about.
type BrowserVersion struct {
	Browser         string `json:"Browser"`
	UserAgent       string `json:"User-Agent"`
	ProtocolVersion string `json:"Protocol-Version"`
}

// NewBrowser creates a client for a base URL such as http://127.0.0.1:9333.
func NewBrowser(baseURL string) *Browser {
	return &Browser{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// BaseURL reports the configured endpoint.
func (b *Browser) BaseURL() string { return b.baseURL }

// Version returns /json/version, which also proves the endpoint is alive.
func (b *Browser) Version(ctx context.Context) (BrowserVersion, error) {
	var out BrowserVersion
	if err := b.getJSON(ctx, "/json/version", &out); err != nil {
		return BrowserVersion{}, err
	}
	return out, nil
}

// Targets returns every page target.
func (b *Browser) Targets(ctx context.Context) ([]Target, error) {
	var out []Target
	if err := b.getJSON(ctx, "/json/list", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// FindTarget returns the first page target whose URL matches want.
func (b *Browser) FindTarget(ctx context.Context, want func(Target) bool) (Target, bool, error) {
	targets, err := b.Targets(ctx)
	if err != nil {
		return Target{}, false, err
	}
	for _, t := range targets {
		if t.Type == "page" && want(t) {
			return t, true, nil
		}
	}
	return Target{}, false, nil
}

// OpenTab creates a new tab and returns its target description.
func (b *Browser) OpenTab(ctx context.Context, rawURL string) (Target, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, b.baseURL+"/json/new?"+url.QueryEscape(rawURL), nil)
	if err != nil {
		return Target{}, err
	}
	resp, err := b.http.Do(req)
	if err != nil {
		return Target{}, fmt.Errorf("open tab: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Target{}, fmt.Errorf("open tab: status %d", resp.StatusCode)
	}
	var out Target
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Target{}, fmt.Errorf("open tab: %w", err)
	}
	return out, nil
}

func (b *Browser) getJSON(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := b.http.Do(req)
	if err != nil {
		return fmt.Errorf("cdp %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cdp %s: status %d", path, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("cdp %s: %w", path, err)
	}
	return nil
}

// Session is one attached page connection.
type Session struct {
	conn *websocket.Conn
	mu   sync.Mutex
	seq  int64
}

// Attach connects to a page target's debugger websocket.
func (b *Browser) Attach(ctx context.Context, target Target) (*Session, error) {
	if target.WebSocketDebuggerURL == "" {
		return nil, fmt.Errorf("target %s has no websocket url", target.ID)
	}
	conn, _, err := websocket.Dial(ctx, target.WebSocketDebuggerURL, nil)
	if err != nil {
		return nil, fmt.Errorf("attach %s: %w", target.ID, err)
	}
	conn.SetReadLimit(64 << 20)
	return &Session{conn: conn}, nil
}

// Close releases the session.
func (s *Session) Close() {
	if s.conn != nil {
		_ = s.conn.Close(websocket.StatusNormalClosure, "")
	}
}

// Evaluate runs a JavaScript expression in the page and decodes its JSON value
// into out. The expression may return a promise when awaitPromise is useful.
func (s *Session) Evaluate(ctx context.Context, expression string, out any) error {
	raw, err := s.call(ctx, "Runtime.evaluate", map[string]any{
		"expression":    expression,
		"returnByValue": true,
		"awaitPromise":  true,
		"userGesture":   true,
	})
	if err != nil {
		return err
	}
	var result struct {
		Result struct {
			Value json.RawMessage `json:"value"`
		} `json:"result"`
		ExceptionDetails *struct {
			Text      string `json:"text"`
			Exception *struct {
				Description string `json:"description"`
			} `json:"exception"`
		} `json:"exceptionDetails"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return fmt.Errorf("evaluate: decode result: %w", err)
	}
	if result.ExceptionDetails != nil {
		detail := result.ExceptionDetails.Text
		if result.ExceptionDetails.Exception != nil && result.ExceptionDetails.Exception.Description != "" {
			detail = result.ExceptionDetails.Exception.Description
		}
		return fmt.Errorf("evaluate: %s", detail)
	}
	if out == nil || len(result.Result.Value) == 0 {
		return nil
	}
	if err := json.Unmarshal(result.Result.Value, out); err != nil {
		return fmt.Errorf("evaluate: decode value: %w", err)
	}
	return nil
}

// Reload reloads the attached page, which is used to refresh an existing login
// tab before its QR code is captured.
func (s *Session) Reload(ctx context.Context) error {
	_, err := s.call(ctx, "Page.reload", map[string]any{"ignoreCache": true})
	return err
}

// Screenshot captures the visible viewport as PNG bytes.
func (s *Session) Screenshot(ctx context.Context) ([]byte, error) {
	raw, err := s.call(ctx, "Page.captureScreenshot", map[string]any{"format": "png"})
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []byte `json:"data"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("screenshot: %w", err)
	}
	return result.Data, nil
}

func (s *Session) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.seq++
	id := s.seq
	payload, err := json.Marshal(map[string]any{"id": id, "method": method, "params": params})
	if err != nil {
		return nil, err
	}
	writeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := s.conn.Write(writeCtx, websocket.MessageText, payload); err != nil {
		return nil, fmt.Errorf("cdp %s: write: %w", method, err)
	}

	for {
		readCtx, cancelRead := context.WithTimeout(ctx, 30*time.Second)
		_, data, err := s.conn.Read(readCtx)
		cancelRead()
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("cdp %s: connection closed", method)
			}
			return nil, fmt.Errorf("cdp %s: read: %w", method, err)
		}
		var envelope struct {
			ID     int64           `json:"id"`
			Result json.RawMessage `json:"result"`
			Error  *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(data, &envelope); err != nil {
			continue
		}
		if envelope.ID != id {
			continue // event or a response to another call
		}
		if envelope.Error != nil {
			return nil, fmt.Errorf("cdp %s: %s", method, envelope.Error.Message)
		}
		return envelope.Result, nil
	}
}
