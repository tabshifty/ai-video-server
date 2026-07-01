package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"video-server/internal/models"
)

type ed2kDownloadRepository interface {
	GetEd2kDownloadTask(ctx context.Context, id uuid.UUID) (models.AdminEd2kDownloadTask, error)
	MarkEd2kDownloadTaskRunning(ctx context.Context, id uuid.UUID, progressText string, history models.AdminEd2kDownloadTaskHistoryItem) error
	MarkEd2kDownloadTaskFailed(ctx context.Context, id uuid.UUID, errorMessage string, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error)
	MarkEd2kDownloadTaskCompleted(ctx context.Context, id uuid.UUID, outputDir, downloadedPath string, files []models.AdminEd2kDownloadTaskFile, progressText string, history models.AdminEd2kDownloadTaskHistoryItem) (models.AdminEd2kDownloadTask, error)
}

// Ed2kDownloadExecutor runs a queued ED2K download task through an external binary.
type Ed2kDownloadExecutor interface {
	Run(ctx context.Context, task models.AdminEd2kDownloadTask) (Ed2kDownloadResult, error)
}

// Ed2kDownloadResult summarizes a finished ED2K download execution.
type Ed2kDownloadResult struct {
	OutputDir      string
	DownloadedPath string
	Files          []models.AdminEd2kDownloadTaskFile
	ProgressText   string
	Executor       string
}

// CommandEd2kDownloadExecutor shells out to a configured external command.
type CommandEd2kDownloadExecutor struct {
	Command string
	Args    []string
	Timeout time.Duration
}

func (e CommandEd2kDownloadExecutor) Run(ctx context.Context, task models.AdminEd2kDownloadTask) (Ed2kDownloadResult, error) {
	command := strings.TrimSpace(e.Command)
	if command == "" {
		return Ed2kDownloadResult{}, fmt.Errorf("ed2k download executable not configured")
	}
	timeout := e.Timeout
	if timeout <= 0 {
		timeout = 6 * time.Hour
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
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return Ed2kDownloadResult{}, fmt.Errorf("decode ed2k download executor output: %w", err)
	}
	if strings.TrimSpace(result.ProgressText) == "" {
		result.ProgressText = "下载已完成"
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
	if item.Status != "queued" {
		logger.Info("skip ed2k download task", "task_id", taskID.String(), "status", item.Status)
		return nil
	}
	if err := repo.MarkEd2kDownloadTaskRunning(ctx, taskID, "下载任务已开始执行", models.AdminEd2kDownloadTaskHistoryItem{
		Kind:    "running",
		Label:   "下载中",
		Message: "外部执行器已接管任务",
		At:      time.Now(),
	}); err != nil {
		return err
	}
	if executor == nil {
		return fmt.Errorf("ed2k download executor not configured")
	}
	result, err := executor.Run(ctx, item)
	if err != nil {
		_, _ = repo.MarkEd2kDownloadTaskFailed(ctx, taskID, err.Error(), models.AdminEd2kDownloadTaskHistoryItem{
			Kind:    "failed",
			Label:   "失败",
			Message: err.Error(),
			At:      time.Now(),
		})
		return nil
	}
	updated, err := repo.MarkEd2kDownloadTaskCompleted(ctx, taskID, result.OutputDir, result.DownloadedPath, result.Files, result.ProgressText, models.AdminEd2kDownloadTaskHistoryItem{
		Kind:    "completed",
		Label:   "已完成",
		Message: "下载任务执行完成",
		At:      time.Now(),
	})
	if err != nil {
		return err
	}
	logger.Info("ed2k download completed", "task_id", taskID.String(), "status", updated.Status, "output_dir", result.OutputDir)
	return nil
}
