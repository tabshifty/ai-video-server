package cloud

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

var (
	// ErrBrowserDown means the CDP endpoint never answered, even after a spawn.
	ErrBrowserDown = errors.New("cloud: 115 browser unavailable")
	// ErrLoginRequired means the browser is up but the 115 session is gone.
	ErrLoginRequired = errors.New("cloud: 115 login required")
)

// Client owns the lifecycle of the local 115 browser session.
type Client struct {
	Browser      *Browser
	AppPath      string
	SpawnTimeout time.Duration
}

// NewClient builds a client for a CDP base URL and the browser bundle path.
func NewClient(baseURL, appPath string) *Client {
	return &Client{
		Browser:      NewBrowser(baseURL),
		AppPath:      appPath,
		SpawnTimeout: 25 * time.Second,
	}
}

// EnsureSession guarantees a live, logged-in pan page and returns an attached
// session. It spawns the browser when the endpoint is down and reports
// ErrBrowserDown / ErrLoginRequired for the two states that need a human.
func (c *Client) EnsureSession(ctx context.Context) (*Session, State, error) {
	if err := c.ensureUp(ctx); err != nil {
		return nil, State{}, err
	}
	target, err := c.ensurePanTarget(ctx)
	if err != nil {
		return nil, State{}, err
	}
	session, err := c.Browser.Attach(ctx, target)
	if err != nil {
		return nil, State{}, err
	}

	deadline := time.Now().Add(20 * time.Second)
	var last State
	for {
		state, probeErr := Probe(ctx, session)
		last = state
		if probeErr == nil && state.LoggedIn {
			return session, state, nil
		}
		if time.Now().After(deadline) {
			session.Close()
			if probeErr != nil {
				return nil, last, probeErr
			}
			return nil, last, ErrLoginRequired
		}
		select {
		case <-ctx.Done():
			session.Close()
			return nil, last, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

// Inspect reports browser and session health without spawning the browser or
// blocking on a login. Used by the admin status endpoint.
func (c *Client) Inspect(ctx context.Context) State {
	probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	_, err := c.Browser.Version(probeCtx)
	cancel()
	if err != nil {
		return State{}
	}
	target, ok, err := c.Browser.FindTarget(ctx, isPanPage)
	if err != nil {
		return State{Up: true, SessionError: "115 浏览器页面读取失败，请稍后刷新"}
	}
	if !ok {
		return State{Up: true}
	}
	session, err := c.Browser.Attach(ctx, target)
	if err != nil {
		return State{Up: true, SessionError: "115 浏览器连接失败，请稍后刷新"}
	}
	defer session.Close()
	state, err := Probe(ctx, session)
	if err != nil {
		return State{Up: true, PageURL: target.URL, SessionError: err.Error()}
	}
	state.Quota = nil
	if state.LoggedIn {
		if quota, qErr := QuotaInfo(ctx, session); qErr == nil && quota.Known {
			state.Quota = &quota
		}
	}
	return state
}

func (c *Client) ensureUp(ctx context.Context) error {
	probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	_, err := c.Browser.Version(probeCtx)
	cancel()
	if err == nil {
		return nil
	}
	if strings.TrimSpace(c.AppPath) == "" {
		return fmt.Errorf("%w: %v", ErrBrowserDown, err)
	}
	if spawnErr := c.spawn(ctx); spawnErr != nil {
		return fmt.Errorf("%w: spawn: %v", ErrBrowserDown, spawnErr)
	}
	deadline := time.Now().Add(c.SpawnTimeout)
	for time.Now().Before(deadline) {
		waitCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		_, versionErr := c.Browser.Version(waitCtx)
		cancel()
		if versionErr == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return fmt.Errorf("%w: no answer on %s", ErrBrowserDown, c.Browser.BaseURL())
}

func (c *Client) spawn(ctx context.Context) error {
	args := []string{"-a", c.AppPath}
	if port := c.port(); port != "" {
		args = append(args, "--args", "--remote-debugging-port="+port)
	}
	cmd := exec.CommandContext(ctx, "/usr/bin/open", args...)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (c *Client) port() string {
	parsed, err := url.Parse(c.Browser.BaseURL())
	if err != nil {
		return ""
	}
	return parsed.Port()
}

func (c *Client) ensurePanTarget(ctx context.Context) (Target, error) {
	target, ok, err := c.Browser.FindTarget(ctx, isPanPage)
	if err != nil {
		return Target{}, err
	}
	if ok {
		return target, nil
	}
	// 115Browser keeps its own 115 tab open (it cannot be closed). Reuse that
	// tab by navigating it to the pan view instead of stacking a new tab on
	// every download attempt.
	if target, ok, err := c.Browser.FindTarget(ctx, is115Page); err != nil {
		return Target{}, err
	} else if ok {
		return c.navigateAndWait(ctx, target, PanPageURL)
	}
	if _, err := c.Browser.OpenTab(ctx, PanPageURL); err != nil {
		return Target{}, err
	}
	deadline := time.Now().Add(25 * time.Second)
	for time.Now().Before(deadline) {
		if target, ok, err := c.Browser.FindTarget(ctx, isPanPage); err == nil && ok {
			return target, nil
		}
		select {
		case <-ctx.Done():
			return Target{}, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return Target{}, fmt.Errorf("cloud: pan page did not open")
}

func (c *Client) navigateAndWait(ctx context.Context, target Target, rawURL string) (Target, error) {
	session, err := c.Browser.Attach(ctx, target)
	if err != nil {
		return Target{}, err
	}
	defer session.Close()
	if err := session.Navigate(ctx, rawURL); err != nil {
		return Target{}, err
	}
	deadline := time.Now().Add(25 * time.Second)
	for time.Now().Before(deadline) {
		if navigated, ok, err := c.Browser.FindTarget(ctx, isPanPage); err == nil && ok {
			return navigated, nil
		}
		select {
		case <-ctx.Done():
			return Target{}, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return Target{}, fmt.Errorf("cloud: pan page did not open")
}

// is115Page reports whether the target is a 115 web page. Subdomains other than
// www are excluded so the chat/asset tabs are never hijacked.
func is115Page(t Target) bool {
	if t.Type != "page" {
		return false
	}
	parsed, err := url.Parse(strings.TrimSpace(t.URL))
	if err != nil {
		return false
	}
	switch strings.ToLower(parsed.Hostname()) {
	case "115.com", "www.115.com":
		return true
	default:
		return false
	}
}

// isPanPage reports whether the target shows the 115 file/cloud view. The site
// often bounces the pan URL through a "?goto=<urlencoded pan url>" redirect, so
// both the direct and encoded forms must match; otherwise the client opens a
// fresh tab on every attempt.
func isPanPage(t Target) bool {
	if !is115Page(t) {
		return false
	}
	parsed, err := url.Parse(strings.TrimSpace(t.URL))
	if err != nil {
		return false
	}
	query := parsed.Query()
	if strings.EqualFold(query.Get("mode"), "wangpan") {
		return true
	}
	switch strings.ToLower(query.Get("ct")) {
	case "file", "lixian":
		return true
	}
	if strings.HasPrefix(strings.ToLower(parsed.Path), "/storage") {
		return true
	}
	gotoURL := strings.TrimSpace(query.Get("goto"))
	if gotoURL == "" {
		return false
	}
	inner, err := url.Parse(gotoURL)
	if err != nil {
		return false
	}
	return strings.EqualFold(inner.Query().Get("mode"), "wangpan")
}
