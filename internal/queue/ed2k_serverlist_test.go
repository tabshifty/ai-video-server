package queue

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hibiken/asynq"
)

func TestCommandEd2kServerlistRefreshExecutorParsesResult(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "executor.sh")
	script := strings.Join([]string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`printf '{"server_met_path":"/Users/u/.aMule/server.met","server_met_bytes":1637,"restarted":true,"executor":"launchctl/com.aivideo.amuled"}'`,
	}, "\n")
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	executor := CommandEd2kServerlistRefreshExecutor{Command: scriptPath}

	result, err := executor.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.ServerMetPath != "/Users/u/.aMule/server.met" {
		t.Fatalf("unexpected server_met_path: %s", result.ServerMetPath)
	}
	if result.ServerMetBytes != 1637 {
		t.Fatalf("unexpected server_met_bytes: %d", result.ServerMetBytes)
	}
	if !result.Restarted {
		t.Fatalf("unexpected restarted: %v", result.Restarted)
	}
	if result.Executor != "launchctl/com.aivideo.amuled" {
		t.Fatalf("unexpected executor: %s", result.Executor)
	}
}

func TestCommandEd2kServerlistRefreshExecutorEmptyCommandErrors(t *testing.T) {
	t.Parallel()

	executor := CommandEd2kServerlistRefreshExecutor{Command: "  "}
	_, err := executor.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for empty command, got nil")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCommandEd2kServerlistRefreshExecutorPropagatesStderr(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "executor.sh")
	script := strings.Join([]string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`printf 'download server.met failed: timeout\n' >&2`,
		"exit 68",
	}, "\n")
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	executor := CommandEd2kServerlistRefreshExecutor{Command: scriptPath}
	_, err := executor.Run(context.Background())
	if err == nil {
		t.Fatal("expected error from failing script, got nil")
	}
	if !strings.Contains(err.Error(), "download server.met failed: timeout") {
		t.Fatalf("expected stderr in error, got: %v", err)
	}
}

func TestCommandEd2kServerlistRefreshExecutorRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "executor.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/usr/bin/env bash\nset -euo pipefail\nprintf 'not-json'\n"), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	executor := CommandEd2kServerlistRefreshExecutor{Command: scriptPath}
	_, err := executor.Run(context.Background())
	if err == nil {
		t.Fatal("expected decode error for non-json stdout, got nil")
	}
	if !strings.Contains(err.Error(), "decode ed2k serverlist refresh executor output") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCommandEd2kServerlistRefreshExecutorTimeoutEnforced 确认超时上下文能中断长跑脚本。
func TestCommandEd2kServerlistRefreshExecutorTimeoutEnforced(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "executor.sh")
	script := strings.Join([]string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		"sleep 10",
	}, "\n")
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	executor := CommandEd2kServerlistRefreshExecutor{Command: scriptPath, Timeout: 1}
	_, err := executor.Run(context.Background())
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// stubServerlistRefreshExecutor 记录是否被调用，供 handler 测试断言跳过分支不会触达执行器。
type stubServerlistRefreshExecutor struct {
	called bool
}

func (s *stubServerlistRefreshExecutor) Run(ctx context.Context) (Ed2kServerlistRefreshResult, error) {
	s.called = true
	return Ed2kServerlistRefreshResult{ServerMetPath: "/stub", Restarted: true}, nil
}

func newTestProcessor(exec Ed2kServerlistRefreshExecutor) *Processor {
	return &Processor{
		ed2kServerlistExec: exec,
		logger:             slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
}

// TestHandleEd2kServerlistRefreshAcceptsNilPayload 验证 scheduler 以 nil payload 入队时，
// handler 不会因 json.Unmarshal(nil) 误判格式错误（回归测试：修复前每次刷新都会失败）。
func TestHandleEd2kServerlistRefreshAcceptsNilPayload(t *testing.T) {
	t.Parallel()

	stub := &stubServerlistRefreshExecutor{}
	p := newTestProcessor(stub) // repo 仍为 nil

	// nil payload：应越过 unmarshal，到达 executor-nil 检查（executor 已配置），
	// 再因 repo==nil 返回 "not configured" 错误 —— 而不是 unmarshal 错误。
	err := p.HandleEd2kServerlistRefresh(context.Background(), asynq.NewTask(TypeEd2kServerlistRefresh, nil))
	if err == nil {
		t.Fatal("expected processor-not-configured error (repo nil), got nil")
	}
	if strings.Contains(err.Error(), "unmarshal") {
		t.Fatalf("nil payload must not trigger unmarshal error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestHandleEd2kServerlistRefreshAcceptsEmptyJSONPayload 验证空 JSON {} payload 也能正常处理。
func TestHandleEd2kServerlistRefreshAcceptsEmptyJSONPayload(t *testing.T) {
	t.Parallel()

	stub := &stubServerlistRefreshExecutor{}
	p := newTestProcessor(stub)

	err := p.HandleEd2kServerlistRefresh(context.Background(), asynq.NewTask(TypeEd2kServerlistRefresh, []byte("{}")))
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("expected processor-not-configured error for repo nil, got: %v", err)
	}
}

// TestHandleEd2kServerlistRefreshRejectsBadJSONPayload 验证非空但非法的 payload 仍被拒绝。
func TestHandleEd2kServerlistRefreshRejectsBadJSONPayload(t *testing.T) {
	t.Parallel()

	stub := &stubServerlistRefreshExecutor{}
	p := newTestProcessor(stub)

	err := p.HandleEd2kServerlistRefresh(context.Background(), asynq.NewTask(TypeEd2kServerlistRefresh, []byte("not-json")))
	if err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Fatalf("expected unmarshal error for bad payload, got: %v", err)
	}
	if stub.called {
		t.Fatal("executor must not be called when payload is invalid")
	}
}

// TestHandleEd2kServerlistRefreshSkipsWhenExecutorNil 验证未配置执行器时安全跳过、不触达 repo。
func TestHandleEd2kServerlistRefreshSkipsWhenExecutorNil(t *testing.T) {
	t.Parallel()

	// executor 为 nil、repo 也为 nil：应先在 executor-nil 分支返回 nil，不触碰 repo。
	p := &Processor{logger: slog.New(slog.NewTextHandler(os.Stderr, nil))}

	err := p.HandleEd2kServerlistRefresh(context.Background(), asynq.NewTask(TypeEd2kServerlistRefresh, nil))
	if err != nil {
		t.Fatalf("expected nil (skip) when executor not configured, got: %v", err)
	}
}
