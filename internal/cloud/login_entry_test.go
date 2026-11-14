package cloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

type fakeLoginCDP struct {
	mu            sync.Mutex
	loginVerdict  string
	enterPan      bool
	navigatedURLs []string
	navigated     bool
}

func (f *fakeLoginCDP) markNavigated(rawURL string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.navigatedURLs = append(f.navigatedURLs, rawURL)
	f.navigated = true
}

func (f *fakeLoginCDP) navigations() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.navigatedURLs...)
}

func (f *fakeLoginCDP) onPanPage() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.navigated
}

// newFakeLoginCDP serves a 115 tab parked on the login landing page. Probes
// report a missing bridge until a Page.navigate reaches the pan view.
func newFakeLoginCDP(t *testing.T, loginVerdict string) (*httptest.Server, *fakeLoginCDP) {
	t.Helper()
	return newFakeLoginCDPWithHint(t, loginVerdict, loginVerdict == "logged_in")
}

func newFakeLoginCDPWithHint(t *testing.T, loginVerdict string, enterPan bool) (*httptest.Server, *fakeLoginCDP) {
	t.Helper()
	state := &fakeLoginCDP{loginVerdict: loginVerdict, enterPan: enterPan}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json/version":
			_, _ = w.Write([]byte(`{"Browser":"115Browser"}`))
		case "/json/list":
			_ = json.NewEncoder(w).Encode([]Target{{
				Type:                 "page",
				URL:                  "https://115.com/",
				WebSocketDebuggerURL: "ws" + strings.TrimPrefix(server.URL, "http") + "/page",
			}})
		case "/page":
			conn, err := websocket.Accept(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close(websocket.StatusNormalClosure, "")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			for {
				_, data, readErr := conn.Read(ctx)
				if readErr != nil {
					return
				}
				var request struct {
					ID     int `json:"id"`
					Method string
					Params struct {
						Expression string `json:"expression"`
						URL        string `json:"url"`
					} `json:"params"`
				}
				if err := json.Unmarshal(data, &request); err != nil {
					return
				}
				var value any
				switch request.Method {
				case "Page.navigate":
					state.markNavigated(request.Params.URL)
					value = map[string]any{}
				case "Runtime.evaluate":
					switch {
					case strings.Contains(request.Params.Expression, "get_quota_package_info"):
						value = map[string]any{}
					case strings.Contains(request.Params.Expression, "enter_pan"):
						value = map[string]any{"state": state.loginVerdict, "enter_pan": state.enterPan}
					case state.onPanPage():
						value = map[string]any{"url": PanPageURL, "lixianKey": true, "download": true, "userId": "123", "loggedIn": true}
					default:
						value = map[string]any{"url": "https://115.com/", "lixianKey": false, "download": false, "userId": "", "loggedIn": false}
					}
				default:
					return
				}
				reply, _ := json.Marshal(map[string]any{
					"id":     request.ID,
					"result": map[string]any{"result": map[string]any{"value": value}},
				})
				if writeErr := conn.Write(ctx, websocket.MessageText, reply); writeErr != nil {
					return
				}
			}
		}
	}))
	t.Cleanup(server.Close)
	return server, state
}

func TestInspectEntersPanPageAfterScannedLogin(t *testing.T) {
	server, fake := newFakeLoginCDP(t, "logged_in")

	state := NewClient(server.URL, "").Inspect(context.Background())

	if !state.LoggedIn {
		t.Fatalf("Inspect() must report the completed scan as logged in: %+v", state)
	}
	if state.PageURL != PanPageURL {
		t.Fatalf("page url = %q, want %q", state.PageURL, PanPageURL)
	}
	navigations := fake.navigations()
	if len(navigations) != 1 || navigations[0] != PanPageURL {
		t.Fatalf("navigations = %#v, want one entry to the pan page", navigations)
	}
}

func TestInspectKeepsLandingPageWhenSessionExpired(t *testing.T) {
	server, fake := newFakeLoginCDP(t, "logged_out")

	state := NewClient(server.URL, "").Inspect(context.Background())

	if !state.Up || state.LoggedIn {
		t.Fatalf("expired session state = %+v", state)
	}
	if state.SessionError != "" {
		t.Fatalf("expired login must not be reported as a validation error: %q", state.SessionError)
	}
	if navigations := fake.navigations(); len(navigations) != 0 {
		t.Fatalf("logged-out landing page must not be navigated: %#v", navigations)
	}
}

func TestInspectDoesNotNavigateHomePageWithoutEnterPanHint(t *testing.T) {
	server, fake := newFakeLoginCDPWithHint(t, "logged_in", false)

	state := NewClient(server.URL, "").Inspect(context.Background())

	if state.LoggedIn {
		t.Fatalf("a home page without the enter-pan hint must not be reported as a usable session: %+v", state)
	}
	if navigations := fake.navigations(); len(navigations) != 0 {
		t.Fatalf("home page must not be hijacked without the enter-pan hint: %#v", navigations)
	}
}

func TestIs115LoginLandingPage(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"https://115.com/", true},
		{"https://www.115.com/", true},
		{"https://115.com/?goto=https%3A%2F%2F115.com%2F%3Fmode%3Dwangpan", false},
		{"https://115.com/login", true},
		{"https://115.com/?mode=wangpan", false},
		{"https://115.com/storage/allfiles", false},
		{"https://yun.115.com/chat", false},
		{"https://example.com/", false},
	}
	for _, tc := range cases {
		if got := is115LoginLandingPage(Target{Type: "page", URL: tc.url}); got != tc.want {
			t.Fatalf("is115LoginLandingPage(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
	if is115LoginLandingPage(Target{Type: "shared_worker", URL: "https://115.com/"}) {
		t.Fatal("non-page targets must not be treated as the login landing page")
	}
}
