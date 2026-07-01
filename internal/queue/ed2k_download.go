package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

type ed2kDownloadRepository interface {
	GetEd2kDownloadTask(ctx context.Context, id uuid.UUID) (models.AdminEd2kDownloadTask, error)
	MarkEd2kDownloadTaskRunning(ctx context.Context, id uuid.UUID, progressText string, history models.AdminEd2kDownloadTaskHistoryItem) error
	UpdateEd2kDownloadTaskProgress(ctx context.Context, id uuid.UUID, progressText string) error
	MarkEd2kDownloadTaskFailed(ctx context.Context, id uuid.UUID, errorMessage string, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error)
	MarkEd2kDownloadTaskCompleted(ctx context.Context, id uuid.UUID, outputDir, downloadedPath string, files []models.AdminEd2kDownloadTaskFile, progressText string, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error)
}

// Ed2kDownloadExecutor submits and polls ED2K work through short external actions.
type Ed2kDownloadExecutor interface {
	Submit(ctx context.Context, task models.AdminEd2kDownloadTask) (Ed2kDownloadResult, error)
	Status(ctx context.Context, task models.AdminEd2kDownloadTask) (Ed2kDownloadResult, error)
}

// Ed2kDownloadResult summarizes a short ED2K executor action.
type Ed2kDownloadResult struct {
	Status         string
	OutputDir      string
	DownloadedPath string
	Files          []models.AdminEd2kDownloadTaskFile
	ProgressText   string
	Executor       string
	ErrorMessage   string
}

// ErrEd2kDownloadStillInProgress asks asynq to poll the external engine again later.
var ErrEd2kDownloadStillInProgress = errors.New("ed2k download still in progress")

const (
	ed2kExecutorModeSubmit = "submit"
	ed2kExecutorModeStatus = "status"

	ed2kDownloadStatusQueued    = "queued"
	ed2kDownloadStatusRunning   = "running"
	ed2kDownloadStatusCompleted = "completed"
	ed2kDownloadStatusFailed    = "failed"
	ed2kDownloadStatusNotFound  = "not_found"
)

// CommandEd2kDownloadExecutor shells out to a configured external command.
type CommandEd2kDownloadExecutor struct {
	Command string
	Args    []string
	Timeout time.Duration
}

func (e CommandEd2kDownloadExecutor) Submit(ctx context.Context, task models.AdminEd2kDownloadTask) (Ed2kDownloadResult, error) {
	return e.run(ctx, task, ed2kExecutorModeSubmit)
}

func (e CommandEd2kDownloadExecutor) Status(ctx context.Context, task models.AdminEd2kDownloadTask) (Ed2kDownloadResult, error) {
	return e.run(ctx, task, ed2kExecutorModeStatus)
}

func (e CommandEd2kDownloadExecutor) run(ctx context.Context, task models.AdminEd2kDownloadTask, mode string) (Ed2kDownloadResult, error) {
	command := strings.TrimSpace(e.Command)
	if command == "" {
		return Ed2kDownloadResult{}, fmt.Errorf("ed2k download executable not configured")
	}
	timeout := e.Timeout
	if timeout <= 0 {
		timeout = time.Minute
	}
	if timeout > time.Minute {
		timeout = time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := append([]string{}, e.Args...)
	args = append(args, task.SourceLink)
	cmd := exec.CommandContext(runCtx, command, args...)
	cmd.Env = append(os.Environ(),
		"ED2K_TASK_ID="+task.ID.String(),
		"ED2K_TASK_TITLE="+task.Title,
		"ED2K_SOURCE_LINK="+task.SourceLink,
		"ED2K_RESOURCE_HASH="+task.ResourceHash,
		"ED2K_FILENAME="+task.Filename,
		"ED2K_EXECUTOR_MODE="+mode,
		fmt.Sprintf("ED2K_DECLARED_SIZE=%d", task.DeclaredSize),
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return Ed2kDownloadResult{}, fmt.Errorf("run ed2k download executor: %w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return Ed2kDownloadResult{}, fmt.Errorf("run ed2k download executor: %w", err)
	}
	var result Ed2kDownloadResult
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &result); err != nil {
		return Ed2kDownloadResult{}, fmt.Errorf("decode ed2k download executor output: %w", err)
	}
	result.Status = strings.TrimSpace(strings.ToLower(result.Status))
	if result.Status == "" {
		result.Status = ed2kDownloadStatusCompleted
	}
	if strings.TrimSpace(result.ProgressText) == "" {
		result.ProgressText = defaultEd2kProgressText(result.Status)
	}
	return result, nil
}

func (p *Processor) HandleEd2kDownload(ctx context.Context, task *asynq.Task) error {
	return handleEd2kDownload(ctx, task, p.repo, p.ed2kExecutor, p.logger)
}

func handleEd2kDownload(ctx context.Context, task *asynq.Task, repo ed2kDownloadRepository, executor Ed2kDownloadExecutor, logger *slog.Logger) error {
	if repo == nil {
		return fmt.Errorf("ed2k download processor not configured")
	}
	var payload Ed2kDownloadPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal ed2k download payload: %w", err)
	}
	taskID, err := uuid.Parse(payload.TaskID)
	if err != nil {
		return fmt.Errorf("invalid ed2k task id: %w", err)
	}
	item, err := repo.GetEd2kDownloadTask(ctx, taskID)
	if err != nil {
		return err
	}
	if item.Status != "queued" && item.Status != "running" {
		logger.Info("skip ed2k download task", "task_id", taskID.String(), "status", item.Status)
		return nil
	}
	if item.Status == "queued" {
		if err := repo.MarkEd2kDownloadTaskRunning(ctx, taskID, "下载任务已开始执行", models.AdminEd2kDownloadTaskHistoryItem{
			Kind:    "running",
			Label:   "下载中",
			Message: "外部执行器已接管任务",
			At:      time.Now(),
		}); err != nil {
			if isEd2kConditionalUpdateMiss(err) {
				logger.Info("skip stale ed2k download start", "task_id", taskID.String())
				return nil
			}
			return err
		}
		item.Status = "running"
		if executor == nil {
			return failRunningEd2kDownload(ctx, repo, taskID, item, "ed2k download executor not configured", logger)
		}
		result, err := executor.Submit(ctx, item)
		if err != nil {
			return failRunningEd2kDownload(ctx, repo, taskID, item, err.Error(), logger)
		}
		return applyEd2kExecutorResult(ctx, repo, taskID, item, result, logger)
	}
	if executor == nil {
		return failRunningEd2kDownload(ctx, repo, taskID, item, "ed2k download executor not configured", logger)
	}

	result, err := executor.Status(ctx, item)
	current, refreshErr := repo.GetEd2kDownloadTask(ctx, taskID)
	if refreshErr != nil {
		return refreshErr
	}
	if current.Status != "running" {
		logger.Info("skip late ed2k download result", "task_id", taskID.String(), "status", current.Status)
		return nil
	}
	if err != nil {
		return failRunningEd2kDownload(ctx, repo, taskID, current, err.Error(), logger)
	}
	return applyEd2kExecutorResult(ctx, repo, taskID, current, result, logger)
}

func applyEd2kExecutorResult(ctx context.Context, repo ed2kDownloadRepository, taskID uuid.UUID, current models.AdminEd2kDownloadTask, result Ed2kDownloadResult, logger *slog.Logger) error {
	status := strings.TrimSpace(strings.ToLower(result.Status))
	if status == "" {
		status = ed2kDownloadStatusCompleted
	}
	switch status {
	case ed2kDownloadStatusCompleted:
		return completeRunningEd2kDownload(ctx, repo, taskID, result, logger)
	case ed2kDownloadStatusFailed:
		message := strings.TrimSpace(result.ErrorMessage)
		if message == "" {
			message = strings.TrimSpace(result.ProgressText)
		}
		if message == "" {
			message = "下载任务执行失败"
		}
		return failRunningEd2kDownload(ctx, repo, taskID, current, message, logger)
	case ed2kDownloadStatusQueued, ed2kDownloadStatusRunning, ed2kDownloadStatusNotFound:
		text := strings.TrimSpace(result.ProgressText)
		if text == "" {
			text = defaultEd2kProgressText(status)
		}
		if err := repo.UpdateEd2kDownloadTaskProgress(ctx, taskID, text); err != nil && !isEd2kConditionalUpdateMiss(err) {
			return err
		}
		return ErrEd2kDownloadStillInProgress
	default:
		text := strings.TrimSpace(result.ProgressText)
		if text == "" {
			text = "等待 aMule 同步下载状态"
		}
		if err := repo.UpdateEd2kDownloadTaskProgress(ctx, taskID, text); err != nil && !isEd2kConditionalUpdateMiss(err) {
			return err
		}
		return ErrEd2kDownloadStillInProgress
	}
}

func completeRunningEd2kDownload(ctx context.Context, repo ed2kDownloadRepository, taskID uuid.UUID, result Ed2kDownloadResult, logger *slog.Logger) error {
	updated, err := repo.MarkEd2kDownloadTaskCompleted(ctx, taskID, result.OutputDir, result.DownloadedPath, result.Files, result.ProgressText, models.AdminEd2kDownloadTaskHistoryItem{
		Kind:    "completed",
		Label:   "已完成",
		Message: "下载任务执行完成",
		At:      time.Now(),
	})
	if err != nil {
		if isEd2kConditionalUpdateMiss(err) {
			logger.Info("skip late ed2k download completion result", "task_id", taskID.String())
			return nil
		}
		return err
	}
	logger.Info("ed2k download completed", "task_id", taskID.String(), "status", updated.Status, "output_dir", result.OutputDir)
	return nil
}

func failRunningEd2kDownload(ctx context.Context, repo ed2kDownloadRepository, taskID uuid.UUID, _ models.AdminEd2kDownloadTask, errorMessage string, logger *slog.Logger) error {
	_, markErr := repo.MarkEd2kDownloadTaskFailed(ctx, taskID, errorMessage, models.AdminEd2kDownloadTaskHistoryItem{
		Kind:    "failed",
		Label:   "失败",
		Message: errorMessage,
		At:      time.Now(),
	})
	if markErr != nil {
		if isEd2kConditionalUpdateMiss(markErr) {
			logger.Info("skip late ed2k download failure result", "task_id", taskID.String())
			return nil
		}
		return markErr
	}
	return nil
}

func defaultEd2kProgressText(status string) string {
	switch status {
	case ed2kDownloadStatusCompleted:
		return "下载已完成"
	case ed2kDownloadStatusFailed:
		return "下载失败"
	case ed2kDownloadStatusQueued:
		return "等待 aMule 接收 ED2K 链接"
	case ed2kDownloadStatusNotFound:
		return "等待 aMule 同步下载状态"
	default:
		return "下载进行中"
	}
}

func isEd2kConditionalUpdateMiss(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), pgx.ErrNoRows.Error())
}
