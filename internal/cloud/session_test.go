package cloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

// Run the actual browser expression through CDP and a JS runtime: an old page
// still has USER_ID even though the authenticated endpoint rejects its cookie.
func TestProbeVerifiesServerSession(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is required to execute the browser probe")
	}
	cases := []struct {
		name      string
		response  any
		status    int
		failure   string
		html      bool
		userID    string
		loggedIn  bool
		wantError bool
	}{
		{name: "expired with stale USER_ID", response: map[string]any{"state": false, "error": "请先登录,后操作！"}, status: 200, userID: "123"},
		{name: "valid session", response: map[string]any{"state": true}, status: 200, userID: "123", loggedIn: true},
		{name: "numeric success", response: map[string]any{"state": 1}, status: 200, userID: "123", loggedIn: true},
		{name: "string success", response: map[string]any{"state": "1"}, status: 200, userID: "123", loggedIn: true},
		{name: "page not ready", response: map[string]any{"state": true}, status: 200},
		{name: "HTML is unknown not logout", status: 200, html: true, userID: "123", wantError: true},
		{name: "timeout is unknown not logout", failure: "TimeoutError", userID: "123", wantError: true},
		{name: "server error is unknown not logout", status: 503, userID: "123", wantError: true},
		{name: "unexpected JSON", response: map[string]any{"state": false, "error": "系统繁忙"}, status: 200, userID: "123", wantError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := websocket.Accept(w, r, nil)
				if err != nil {
					return
				}
				defer conn.Close(websocket.StatusNormalClosure, "")
				ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
				defer cancel()
				_, data, err := conn.Read(ctx)
				if err != nil {
					return
				}
				var request struct {
					ID     int `json:"id"`
					Params struct {
						Expression string `json:"expression"`
					} `json:"params"`
				}
				if err := json.Unmarshal(data, &request); err != nil {
					return
				}
				input, _ := json.Marshal(map[string]any{"expression": request.Params.Expression, "response": tc.response, "status": tc.status, "failure": tc.failure, "html": tc.html, "userId": tc.userID})
				cmd := exec.CommandContext(ctx, node, "testdata/session_probe.cjs")
				cmd.Stdin = strings.NewReader(string(input))
				result, err := cmd.Output()
				if err != nil {
					t.Errorf("execute probe: %v", err)
					return
				}
				response, _ := json.Marshal(map[string]any{"id": request.ID, "result": map[string]any{"result": map[string]any{"value": json.RawMessage(result)}}})
				_ = conn.Write(ctx, websocket.MessageText, response)
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			session, err := NewBrowser(server.URL).Attach(ctx, Target{WebSocketDebuggerURL: "ws" + strings.TrimPrefix(server.URL, "http")})
			if err != nil {
				t.Fatal(err)
			}
			defer session.Close()
			state, err := Probe(ctx, session)
			if (err != nil) != tc.wantError {
				t.Fatalf("Probe error = %v, wantError = %v", err, tc.wantError)
			}
			if !state.Up || state.LoggedIn != tc.loggedIn {
				t.Fatalf("Probe = %+v, want loggedIn = %v", state, tc.loggedIn)
			}
		})
	}
}

func TestInspectReportsSessionValidationError(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json/version":
			_, _ = w.Write([]byte(`{"Browser":"115Browser"}`))
		case "/json/list":
			_ = json.NewEncoder(w).Encode([]Target{{Type: "page", URL: PanPageURL, WebSocketDebuggerURL: "ws" + strings.TrimPrefix(server.URL, "http") + "/page"}})
		case "/page":
			conn, err := websocket.Accept(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close(websocket.StatusNormalClosure, "")
			ctx, cancel := context.WithTimeout(r.Context(), time.Second)
			defer cancel()
			_, data, err := conn.Read(ctx)
			if err != nil {
				return
			}
			var request struct {
				ID int `json:"id"`
			}
			_ = json.Unmarshal(data, &request)
			reply, _ := json.Marshal(map[string]any{"id": request.ID, "result": map[string]any{"result": map[string]any{"value": map[string]any{"url": PanPageURL, "userId": "123", "lixianKey": true, "download": true, "error": "115 登录状态验证失败，请稍后刷新"}}}})
			_ = conn.Write(ctx, websocket.MessageText, reply)
		}
	}))
	defer server.Close()
	state := NewClient(server.URL, "").Inspect(context.Background())
	if !state.Up || state.LoggedIn || state.SessionError == "" {
		t.Fatalf("Inspect must distinguish validation failure from expired login: %+v", state)
	}
}
