package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

// Ed2kServerlistRefreshExecutor 刷新 aMule 服务器列表并重启守护进程的外部执行器边界。
// 与 Ed2kDownloadExecutor 同构：Go 侧只 shell out 到脚本并解析 stdout JSON，
// 不写 ~/.aMule/server.met、不调 amulecmd，把 aMule 远控细节拦在 Go 外（见 [[ED2K 服务器列表刷新归执行器脚本]]）。
type Ed2kServerlistRefreshExecutor interface {
	Run(ctx context.Context) (Ed2kServerlistRefreshResult, error)
}

// Ed2kServerlistRefreshResult 汇总一次刷新执行结果，仅供 worker 日志，不落库。
type Ed2kServerlistRefreshResult struct {
	ServerMetPath  string `json:"server_met_path"`
	ServerMetBytes int64  `json:"server_met_bytes"`
	Restarted      bool   `json:"restarted"`
	Executor       string `json:"executor"`
}

// CommandEd2kServerlistRefreshExecutor 通过外部脚本下载 server.met 并重启 amuled。
type CommandEd2kServerlistRefreshExecutor struct {
	Command string
	Args    []string
	Timeout time.Duration
}

func (e CommandEd2kServerlistRefreshExecutor) Run(ctx context.Context) (Ed2kServerlistRefreshResult, error) {
	command := strings.TrimSpace(e.Command)
	if command == "" {
		return Ed2kServerlistRefreshResult{}, fmt.Errorf("ed2k serverlist refresh executable not configured")
	}
	timeout := e.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := append([]string{}, e.Args...)
	cmd := exec.CommandContext(runCtx, command, args...)
	// 刷新脚本从继承的进程环境读取参数（ED2K_SERVERLIST_URL / AMULE_REMOTE_* 等，
	// 均由 .env 经 godotenv 注入 worker 进程）。这与下载执行器显式 append 每任务 env
	// 不同，是刻意为之：刷新无每任务数据，参数全是配置级、已存在于 worker 环境。
	cmd.Env = os.Environ()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return Ed2kServerlistRefreshResult{}, fmt.Errorf("run ed2k serverlist refresh executor: %w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return Ed2kServerlistRefreshResult{}, fmt.Errorf("run ed2k serverlist refresh executor: %w", err)
	}

	var result Ed2kServerlistRefreshResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return Ed2kServerlistRefreshResult{}, fmt.Errorf("decode ed2k serverlist refresh executor output: %w", err)
	}
	return result, nil
}

// HandleEd2kServerlistRefresh 处理 server.met 定时刷新任务。
// 先查在途下载任务，queued/running 任一存在即跳过本轮（重启 amuled 会打断在途下载），
// 否则调执行器刷新 + 重启，结果只写日志（见 [[ED2K 服务器列表刷新在途任务跳过]]）。
func (p *Processor) HandleEd2kServerlistRefresh(ctx context.Context, task *asynq.Task) error {
	// scheduler 以 nil/空 payload 入队（刷新参数走 config + 执行器，不随任务携带），
	// 因此只在校验非空 payload 时才 unmarshal，避免 nil payload 被误判为格式错误而每次刷新都失败。
	if len(task.Payload()) > 0 {
		var payload Ed2kServerlistRefreshPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("unmarshal ed2k serverlist refresh payload: %w", err)
		}
	}
	if p.ed2kServerlistExec == nil {
		// 未配置执行器但任务被投递（如关闭 URL 后队列仍有残留）→ 安全跳过，不重试。
		if p.logger != nil {
			p.logger.Warn("ed2k serverlist refresh executor not configured, skip")
		}
		return nil
	}
	if p.repo == nil {
		return fmt.Errorf("ed2k serverlist refresh processor not configured")
	}

	// 在途闸门：queued 或 running 任一存在则跳过本轮，等下一个 cron 周期。
	active, err := p.repo.HasActiveEd2kDownloadTasks(ctx)
	if err != nil {
		// 查询失败也跳过：宁可明天再刷，也不冒险重启打断在途下载。
		if p.logger != nil {
			p.logger.Error("count active ed2k download tasks failed, skip refresh round", "error", err)
		}
		return nil
	}
	if active {
		if p.logger != nil {
			p.logger.Info("skip ed2k serverlist refresh: active download tasks in flight")
		}
		return nil
	}

	result, err := p.ed2kServerlistExec.Run(ctx)
	if err != nil {
		// 执行器失败（下载/校验/重启任一）只记日志，不重试（MaxRetry(0) + return nil）。
		if p.logger != nil {
			p.logger.Error("ed2k serverlist refresh failed", "error", err)
		}
		return nil
	}
	if p.logger != nil {
		p.logger.Info("ed2k serverlist refresh completed",
			"server_met_path", result.ServerMetPath,
			"server_met_bytes", result.ServerMetBytes,
			"restarted", result.Restarted,
			"executor", result.Executor,
		)
	}
	return nil
}
