package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"video-server/internal/config"
	"video-server/internal/database"
	"video-server/internal/queue"
	"video-server/internal/repository"
	"video-server/internal/services"
	"video-server/internal/telegram"
)

const (
	telegramReconcileInterval      = time.Minute
	telegramSourcePreviewTTL       = 5 * time.Minute
	telegramControlShutdownTimeout = 15 * time.Second
)

// telegramIngestorVersion can be set through -ldflags during release builds.
var telegramIngestorVersion = "dev"

func main() {
	loadEnvironment()
	mode := flag.String("mode", "run", "运行模式：login 或 run")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}
	if err := cfg.ValidateTelegramIngestor(); err != nil {
		slog.Error("Telegram 采集器配置无效", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, stop := signalContext()
	defer stop()

	var runErr error
	switch strings.ToLower(strings.TrimSpace(*mode)) {
	case "login":
		runErr = runLogin(ctx, cfg, logger)
	case "run":
		runErr = runIngestor(ctx, cfg, logger)
	default:
		runErr = fmt.Errorf("不支持的运行模式 %q，请使用 login 或 run", *mode)
	}
	if runErr != nil && !errors.Is(runErr, context.Canceled) {
		logger.Error("Telegram 采集器退出", "error", runErr)
		os.Exit(1)
	}
}

func loadEnvironment() {
	if envFile := os.Getenv("ENV_FILE"); envFile != "" {
		_ = godotenv.Overload(envFile)
		return
	}
	_ = godotenv.Load()
}

func signalContext() (context.Context, context.CancelFunc) {
	return signalNotifyContext(context.Background())
}

// signalNotifyContext is kept behind a small function so the entrypoint's
// context lifecycle remains easy to exercise without starting Telegram.
var signalNotifyContext = func(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
}

func runLogin(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	client, err := telegram.NewGotdClient(telegram.GotdClientConfig{
		APIID:       cfg.TelegramAPIID,
		APIHash:     cfg.TelegramAPIHash,
		Phone:       cfg.TelegramPhone,
		SessionPath: cfg.TelegramSessionPath,
		Logger:      zap.NewNop(),
	})
	if err != nil {
		return fmt.Errorf("创建 Telegram 客户端: %w", err)
	}
	prompter := &stdinLoginPrompter{reader: bufio.NewReader(os.Stdin), writer: os.Stdout}
	if err := client.Run(ctx, prompter, func(context.Context) error {
		logger.Info("Telegram 登录成功，session 已保存", "path", cfg.TelegramSessionPath)
		return nil
	}); err != nil {
		return fmt.Errorf("Telegram 登录: %w", err)
	}
	return nil
}

func runIngestor(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	importUserID, err := uuid.Parse(strings.TrimSpace(cfg.TelegramImportUserID))
	if err != nil {
		return fmt.Errorf("解析 Telegram 归属用户: %w", err)
	}

	pool, err := database.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		return fmt.Errorf("连接 PostgreSQL: %w", err)
	}
	defer pool.Close()
	repo := repository.NewVideoRepository(pool)
	if _, err := repo.GetUserByID(ctx, importUserID); err != nil {
		return fmt.Errorf("Telegram 归属用户不存在: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword})
	defer func() {
		if closeErr := redisClient.Close(); closeErr != nil {
			logger.Warn("关闭 Redis 客户端失败", "error", closeErr)
		}
	}()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("连接 Redis: %w", err)
	}

	sessionFactory, err := telegram.NewGotdAuthorizationSessionFactory(telegram.GotdAuthorizationSessionFactoryConfig{
		APIID:       cfg.TelegramAPIID,
		APIHash:     cfg.TelegramAPIHash,
		Phone:       cfg.TelegramPhone,
		SessionPath: cfg.TelegramSessionPath,
		Logger:      zap.NewNop(),
	})
	if err != nil {
		return fmt.Errorf("创建 Telegram 授权 session 工厂: %w", err)
	}

	telegramTasks := queue.NewTelegramTaskEnqueuer(
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.TelegramRealtimeQueue,
		cfg.TelegramBackfillQueue,
		cfg.TelegramControlQueue,
	)
	defer func() {
		if closeErr := telegramTasks.Close(); closeErr != nil {
			logger.Warn("关闭 Telegram 任务客户端失败", "error", closeErr)
		}
	}()
	transcodeTasks := queue.NewEnqueuer(cfg.RedisAddr, cfg.RedisPassword, cfg.AsynqQueue, cfg.TranscodeTaskTimeout)
	defer func() {
		if closeErr := transcodeTasks.Close(); closeErr != nil {
			logger.Warn("关闭转码任务客户端失败", "error", closeErr)
		}
	}()

	uploadSvc := services.NewUploadService(repo, cfg.UploadTempDir, cfg.StorageRoot, logger)
	runtime := newTelegramRuntime(
		func() (telegramRuntimeClient, error) {
			return telegram.NewGotdClient(telegram.GotdClientConfig{
				APIID:       cfg.TelegramAPIID,
				APIHash:     cfg.TelegramAPIHash,
				Phone:       cfg.TelegramPhone,
				SessionPath: cfg.TelegramSessionPath,
				Logger:      zap.NewNop(),
			})
		},
		func(runCtx context.Context, runtimeClient telegramRuntimeClient) error {
			client, ok := runtimeClient.(telegram.Client)
			if !ok {
				return errors.New("Telegram 采集客户端不支持消息采集")
			}
			ingestion := services.NewTelegramIngestionService(
				client,
				repo,
				uploadSvc,
				telegramTasks,
				transcodeQueueAdapter{enqueuer: transcodeTasks},
				cfg.UploadTempDir,
				cfg.StorageRoot,
				cfg.MaxVideoSize,
				importUserID,
			)
			processor := queue.NewTelegramProcessor(ingestion, logger, cfg.TelegramDownloadConcurrency)
			mux := asynq.NewServeMux()
			processor.Register(mux)
			server := asynq.NewServer(asynq.RedisClientOpt{
				Addr:     cfg.RedisAddr,
				Password: cfg.RedisPassword,
			}, asynq.Config{
				Concurrency: cfg.TelegramMaxActiveTasks,
				Queues:      telegramQueueWeights(cfg),
			})
			return runTelegramRuntime(runCtx, client, repo, ingestion, telegramTasks, server, mux, logger)
		},
	)
	authorizations := telegram.NewAuthorizationService(telegram.AuthorizationServiceConfig{
		Phone:       cfg.TelegramPhone,
		Repository:  repo,
		Sessions:    sessionFactory,
		Maintenance: runtime,
	})
	previews := telegram.NewSourcePreviewService(runtime, nil, telegramSourcePreviewTTL)
	control := newIngestorControlService(repo, runtime, authorizations, previews)
	return runTelegramControlPlane(
		ctx,
		cfg.TelegramControlAddr,
		cfg.TelegramControlToken,
		control,
		repo,
		runtime,
		telegramIngestorVersion,
		logger,
	)
}

type telegramRuntimeLifecycle interface {
	Run(ctx context.Context) error
	View() telegramRuntimeView
}

// runTelegramControlPlane keeps control HTTP and collector heartbeats alive
// independently from the currently authorized MTProto connection.
func runTelegramControlPlane(
	ctx context.Context,
	address string,
	token string,
	control telegram.ControlService,
	repo telegramHeartbeatRepository,
	runtime telegramRuntimeLifecycle,
	version string,
	logger *slog.Logger,
) error {
	if strings.TrimSpace(address) == "" {
		return errors.New("Telegram 控制监听地址不能为空")
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("监听 Telegram 控制端口: %w", err)
	}
	return runTelegramControlPlaneOnListener(ctx, listener, token, control, repo, runtime, version, logger)
}

func runTelegramControlPlaneOnListener(
	ctx context.Context,
	listener net.Listener,
	token string,
	control telegram.ControlService,
	repo telegramHeartbeatRepository,
	runtime telegramRuntimeLifecycle,
	version string,
	logger *slog.Logger,
) error {
	if ctx == nil {
		if listener != nil {
			_ = listener.Close()
		}
		return errors.New("Telegram 控制平面 context 不能为空")
	}
	if listener == nil {
		return errors.New("Telegram 控制监听器不可用")
	}
	if strings.TrimSpace(token) == "" {
		_ = listener.Close()
		return errors.New("Telegram 控制令牌不能为空")
	}
	if control == nil || repo == nil || runtime == nil {
		_ = listener.Close()
		return errors.New("Telegram 控制平面依赖不可用")
	}
	controlServer := &http.Server{
		Handler:           telegram.NewControlHandler(control, token),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      time.Minute,
		IdleTimeout:       time.Minute,
	}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	httpErrCh := make(chan error, 1)
	go func() {
		httpErrCh <- controlServer.Serve(listener)
	}()
	heartbeatsDone := make(chan struct{})
	go func() {
		defer close(heartbeatsDone)
		runTelegramHeartbeats(runCtx, repo, runtime, version, logger)
	}()
	runtimeErrCh := make(chan error, 1)
	go func() {
		runtimeErrCh <- runtime.Run(runCtx)
	}()

	if logger != nil {
		logger.Info("Telegram 控制服务已启动", "address", listener.Addr().String())
	}
	var result error
	select {
	case <-ctx.Done():
	case err := <-runtimeErrCh:
		if err != nil {
			result = fmt.Errorf("运行 Telegram 采集运行时: %w", err)
		} else if ctx.Err() == nil {
			result = errors.New("Telegram 采集运行时意外停止")
		}
	case err := <-httpErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			result = fmt.Errorf("运行 Telegram 控制服务: %w", err)
		}
	}

	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), telegramControlShutdownTimeout)
	defer shutdownCancel()
	if err := controlServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) && result == nil {
		result = fmt.Errorf("关闭 Telegram 控制服务: %w", err)
	}
	select {
	case err := <-runtimeErrCh:
		if err != nil && result == nil {
			result = fmt.Errorf("停止 Telegram 采集运行时: %w", err)
		}
	case <-shutdownCtx.Done():
		if result == nil {
			result = fmt.Errorf("等待 Telegram 采集运行时停止: %w", shutdownCtx.Err())
		}
	}
	<-heartbeatsDone
	return result
}

func runTelegramRuntime(
	ctx context.Context,
	client telegram.Client,
	repo *repository.VideoRepository,
	ingestion *services.TelegramIngestionService,
	tasks *queue.TelegramTaskEnqueuer,
	server *asynq.Server,
	mux *asynq.ServeMux,
	logger *slog.Logger,
) error {
	if ctx == nil {
		return errors.New("Telegram runtime context 不能为空")
	}
	runtimeCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.Run(mux)
	}()

	subscribeErrCh := make(chan error, 1)
	go func() {
		subscribeErrCh <- client.Subscribe(runtimeCtx, func(message telegram.Message) error {
			return ingestion.HandleRealtimeMessage(runtimeCtx, message)
		})
	}()

	sources, err := repo.ListTelegramSources(ctx, true)
	if err != nil {
		server.Shutdown()
		return fmt.Errorf("读取 Telegram 来源: %w", err)
	}
	for _, source := range sources {
		if err := tasks.EnqueueSourceSync(source.ID); err != nil {
			server.Shutdown()
			return fmt.Errorf("调度 Telegram 来源 %s: %w", source.ID, err)
		}
	}
	if err := tasks.EnqueueReconcile(); err != nil {
		server.Shutdown()
		return fmt.Errorf("调度 Telegram 恢复任务: %w", err)
	}

	ticker := time.NewTicker(telegramReconcileInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			server.Shutdown()
			return nil
		case err := <-serverErrCh:
			cancel()
			if err == nil {
				return errors.New("Telegram Asynq 服务意外停止")
			}
			return fmt.Errorf("运行 Telegram Asynq 服务: %w", err)
		case err := <-subscribeErrCh:
			server.Shutdown()
			if err == nil || (errors.Is(err, context.Canceled) && ctx.Err() != nil) {
				return nil
			}
			return fmt.Errorf("订阅 Telegram 实时消息: %w", err)
		case <-ticker.C:
			if err := tasks.EnqueueReconcile(); err != nil {
				logger.Warn("调度 Telegram 恢复任务失败", "error", err)
			}
		}
	}
}

func telegramQueueWeights(cfg config.Config) map[string]int {
	return map[string]int{
		telegramQueueName(cfg.TelegramRealtimeQueue, "telegram-realtime"): 10,
		telegramQueueName(cfg.TelegramBackfillQueue, "telegram-backfill"): 1,
		telegramQueueName(cfg.TelegramControlQueue, "telegram-control"):   1,
	}
}

func telegramQueueName(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}

type transcodeQueueAdapter struct {
	enqueuer *queue.Enqueuer
}

func (a transcodeQueueAdapter) EnqueueTranscode(videoID, inputPath, outputDir, targetFormat string, force bool) error {
	if a.enqueuer == nil {
		return errors.New("转码任务客户端不可用")
	}
	return a.enqueuer.EnqueueTranscode(queue.TranscodePayload{
		VideoID:      videoID,
		InputPath:    inputPath,
		OutputDir:    outputDir,
		TargetFormat: targetFormat,
		Force:        force,
	})
}

type stdinLoginPrompter struct {
	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex
}

func (p *stdinLoginPrompter) Code(ctx context.Context) (string, error) {
	return p.ask(ctx, "请输入 Telegram 验证码：")
}

func (p *stdinLoginPrompter) Password(ctx context.Context) (string, error) {
	return p.ask(ctx, "请输入 Telegram 二次验证密码：")
}

func (p *stdinLoginPrompter) ask(ctx context.Context, prompt string) (string, error) {
	if p == nil || p.reader == nil || p.writer == nil {
		return "", errors.New("Telegram 登录输入器未初始化")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, err := fmt.Fprint(p.writer, prompt); err != nil {
		return "", fmt.Errorf("输出 Telegram 登录提示: %w", err)
	}
	value, err := p.reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("读取 Telegram 登录输入: %w", err)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("Telegram 登录输入不能为空")
	}
	return value, nil
}

var _ services.TelegramIngestion = (*services.TelegramIngestionService)(nil)
var _ telegram.LoginPrompter = (*stdinLoginPrompter)(nil)
