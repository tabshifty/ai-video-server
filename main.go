package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"video-server/internal/cloud"
	"video-server/internal/config"
	"video-server/internal/database"
	"video-server/internal/handlers"
	"video-server/internal/queue"
	"video-server/internal/repository"
	"video-server/internal/services"
	"video-server/internal/telegram"
)

//go:generate go run ./cmd/gen-openapi

// @title Video Server API
// @version 1.0
// @description API documentation for video-server.
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	if envFile := os.Getenv("ENV_FILE"); envFile != "" {
		_ = godotenv.Overload(envFile)
	} else {
		_ = godotenv.Load()
	}
	modeFlag := flag.String("mode", "", "run mode: server|worker")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}
	mode := cfg.Mode
	if *modeFlag != "" {
		mode = *modeFlag
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx := context.Background()

	pool, err := database.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.Error("connect postgres failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewVideoRepository(pool)
	transSvc := services.NewTranscodeService(cfg.StorageRoot)

	switch mode {
	case "server":
		if err := runServer(cfg, pool, repo, transSvc, logger); err != nil {
			logger.Error("server stopped with error", "error", err)
			os.Exit(1)
		}
	case "worker":
		if err := runWorker(cfg, repo, transSvc, logger); err != nil {
			logger.Error("worker stopped with error", "error", err)
			os.Exit(1)
		}
	default:
		logger.Error("invalid mode", "mode", mode)
		os.Exit(1)
	}
}

type archiveImportQueueAdapter struct {
	enqueuer *queue.Enqueuer
}

func (a archiveImportQueueAdapter) EnqueueTranscode(videoID, inputPath, outputDir, targetFormat string, force bool) error {
	if a.enqueuer == nil {
		return fmt.Errorf("queue enqueuer unavailable")
	}
	return a.enqueuer.EnqueueTranscode(queue.TranscodePayload{
		VideoID:      videoID,
		InputPath:    inputPath,
		OutputDir:    outputDir,
		TargetFormat: targetFormat,
		Force:        force,
	})
}

func (a archiveImportQueueAdapter) EnqueueScrapeMovie(videoID, filePath, filename string) error {
	if a.enqueuer == nil {
		return fmt.Errorf("queue enqueuer unavailable")
	}
	return a.enqueuer.EnqueueScrapeMovie(queue.ScrapePayload{
		VideoID:  videoID,
		FilePath: filePath,
		Filename: filename,
		Type:     "movie",
	})
}

func (a archiveImportQueueAdapter) EnqueueScrapeTV(videoID, filePath, filename string) error {
	if a.enqueuer == nil {
		return fmt.Errorf("queue enqueuer unavailable")
	}
	return a.enqueuer.EnqueueScrapeTV(queue.ScrapePayload{
		VideoID:  videoID,
		FilePath: filePath,
		Filename: filename,
		Type:     "episode",
	})
}

func (a archiveImportQueueAdapter) EnqueueArchiveExtract(batchID uuid.UUID, password string) error {
	if a.enqueuer == nil {
		return fmt.Errorf("queue enqueuer unavailable")
	}
	return a.enqueuer.EnqueueArchiveExtract(queue.ArchiveExtractPayload{
		BatchID:  batchID.String(),
		Password: password,
	})
}

func (a archiveImportQueueAdapter) EnqueueScrapeAV(videoID, filePath, filename string) error {
	if a.enqueuer == nil {
		return fmt.Errorf("queue enqueuer unavailable")
	}
	return a.enqueuer.EnqueueScrapeAV(queue.ScrapePayload{
		VideoID:  videoID,
		FilePath: filePath,
		Filename: filename,
		Type:     "av",
	})
}

type avScraperConfigurer interface {
	ConfigureAVScraperConfig(services.AVScraperConfig)
}

func buildAVScraperConfig(cfg config.Config) services.AVScraperConfig {
	return services.AVScraperConfig{
		BaseURL:           cfg.AVScraperBaseURL,
		UserAgent:         cfg.AVScraperUserAgent,
		Timeout:           cfg.AVScraperTimeout,
		SiteURLs:          cfg.AVSiteURLs,
		JavDBCookie:       cfg.AVScraperJavDBCookie,
		JavBusCookie:      cfg.AVScraperJavBusCookie,
		ThePornDBAPIToken: cfg.AVScraperThePornDBAPIToken,
		ThePornDBNoHash:   cfg.AVScraperThePornDBNoHash,
	}
}

func configureAVScraper(scrapeSvc avScraperConfigurer, cfg config.Config) {
	scrapeSvc.ConfigureAVScraperConfig(buildAVScraperConfig(cfg))
}

type contentTranslationConfigurer interface {
	ConfigureContentTranslation(services.TranslationConfig)
}

func buildTranslationConfig(cfg config.Config) services.TranslationConfig {
	return services.TranslationConfig{
		APIURL:  cfg.TranslationAPIURL,
		APIKey:  cfg.TranslationAPIKey,
		Model:   cfg.TranslationModel,
		Timeout: cfg.TranslationTimeout,
	}
}

func configureContentTranslation(scrapeSvc contentTranslationConfigurer, cfg config.Config) {
	scrapeSvc.ConfigureContentTranslation(buildTranslationConfig(cfg))
}

func runServer(cfg config.Config, pool *pgxpool.Pool, repo *repository.VideoRepository, transSvc *services.TranscodeService, logger *slog.Logger) error {
	enqueuer := queue.NewEnqueuer(cfg.RedisAddr, cfg.RedisPassword, cfg.AsynqQueue, cfg.CloudDownloadQueue, cfg.TranscodeTaskTimeout)
	defer enqueuer.Close()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	defer redisClient.Close()
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}

	uploadSvc := services.NewUploadService(repo, cfg.UploadTempDir, cfg.StorageRoot, logger)
	chunkUploadSvc := services.NewChunkUploadService(cfg.UploadTempDir)
	recSvc := services.NewRecommendService(repo)
	scrapeSvc := services.NewScraperService(repo, cfg.TMDBAPIKey, cfg.TMDBBaseURL, cfg.StorageRoot, cfg.PosterStoragePath, cfg.TMDBTimeout)
	configureAVScraper(scrapeSvc, cfg)
	configureContentTranslation(scrapeSvc, cfg)
	appSvc := services.NewAppService(repo)
	imageSvc := services.NewImageService(repo, cfg.UploadTempDir, cfg.StorageRoot, logger)
	subtitleSvc := services.NewSubtitleService(repo, cfg.StorageRoot, logger)
	archiveImportSvc := services.NewArchiveImportService(pool, uploadSvc, imageSvc, repo, archiveImportQueueAdapter{enqueuer: enqueuer}, cfg.StorageRoot, cfg.UploadTempDir, logger)
	hermesForumRepo := repository.NewHermesForumRepository(repo)
	hermesForumSvc := services.NewHermesForumService(hermesForumRepo)
	telegramTasks := queue.NewTelegramTaskEnqueuer(cfg.RedisAddr, cfg.RedisPassword, cfg.TelegramRealtimeQueue, cfg.TelegramBackfillQueue, cfg.TelegramControlQueue)
	defer telegramTasks.Close()
	telegramSourceSvc := services.NewTelegramSourceService(repo, telegramTasks)
	var telegramManagementSvc *services.TelegramManagementService
	if cfg.TelegramControlToken != "" {
		controlClient, err := telegram.NewControlClient(telegram.ControlClientConfig{
			BaseURL: cfg.TelegramControlURL,
			Token:   cfg.TelegramControlToken,
		})
		if err != nil {
			return fmt.Errorf("创建 Telegram 控制客户端: %w", err)
		}
		telegramConfirmations := services.NewTelegramConfirmationService(services.TelegramConfirmationServiceConfig{
			Repository: repo,
		})
		telegramManagementSvc = services.NewTelegramManagementService(services.TelegramManagementServiceConfig{
			Control:       controlClient,
			Sources:       telegramSourceSvc,
			Audits:        repo,
			Confirmations: telegramConfirmations,
		})
	}
	passwordVaultCipher, err := services.NewPasswordVaultCipher(cfg.PasswordVaultKey)
	if err != nil {
		return err
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	api := handlers.NewAPI(
		repo,
		uploadSvc,
		chunkUploadSvc,
		recSvc,
		scrapeSvc,
		appSvc,
		imageSvc,
		subtitleSvc,
		archiveImportSvc,
		hermesForumSvc,
		telegramManagementSvc,
		enqueuer,
		logger,
		redisClient,
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.AsynqQueue,
		cfg.JWTSecret,
		cfg.PlayURLSignSecret,
		cfg.HermesAPIToken,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		cfg.MaxVideoSize,
		cfg.StorageRoot,
		cfg.UploadTempDir,
		cfg.ServerLogPath,
		cfg.AdminWebDistPath,
		cfg.EnableSwagger,
		handlers.ImageGenerationConfig{
			APIURL:  cfg.ImageGenerationAPIURL,
			APIKey:  cfg.ImageGenerationAPIKey,
			Model:   cfg.ImageGenerationModel,
			Timeout: cfg.ImageGenerationTimeout,
		},
		passwordVaultCipher,
	)
	if cfg.CloudDownloadEnabled {
		cloudClient := cloud.NewClient(cfg.CloudBrowserCDP, cfg.CloudBrowserApp)
		cloudRepo := repository.NewCloudDownloadRepository(pool)
		cloudSvc := services.NewCloudDownloadService(cloudRepo, cloudClient, cfg.DownloadDir, enqueuer, logger)
		api.SetCloudDownloadService(cloudSvc)
		logger.Info("cloud download enabled", "cdp", cfg.CloudBrowserCDP, "download_dir", cfg.DownloadDir)
	}
	api.Register(r)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server started", "addr", cfg.HTTPAddr)
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("http shutdown: %w", err)
		}
		return nil
	case err := <-errCh:
		if err == nil || err == http.ErrServerClosed {
			return nil
		}
		return fmt.Errorf("http serve: %w", err)
	}
}

func runWorker(cfg config.Config, repo *repository.VideoRepository, transSvc *services.TranscodeService, logger *slog.Logger) error {
	mux := asynq.NewServeMux()
	enqueuer := queue.NewEnqueuer(cfg.RedisAddr, cfg.RedisPassword, cfg.AsynqQueue, cfg.CloudDownloadQueue, cfg.TranscodeTaskTimeout)
	defer enqueuer.Close()
	scrapeSvc := services.NewScraperService(repo, cfg.TMDBAPIKey, cfg.TMDBBaseURL, cfg.StorageRoot, cfg.PosterStoragePath, cfg.TMDBTimeout)
	configureAVScraper(scrapeSvc, cfg)
	configureContentTranslation(scrapeSvc, cfg)
	subtitleSvc := services.NewSubtitleService(repo, cfg.StorageRoot, logger)
	processor := queue.NewProcessor(repo, transSvc, scrapeSvc, subtitleSvc, enqueuer, logger, cfg.StorageRoot)
	pool, err := pgxpool.New(context.Background(), cfg.PostgresDSN)
	if err != nil {
		return fmt.Errorf("create archive extract pool: %w", err)
	}
	defer pool.Close()
	uploadSvc := services.NewUploadService(repo, cfg.UploadTempDir, cfg.StorageRoot, logger)
	imageSvc := services.NewImageService(repo, cfg.UploadTempDir, cfg.StorageRoot, logger)
	archiveImportSvc := services.NewArchiveImportService(pool, uploadSvc, imageSvc, repo, archiveImportQueueAdapter{enqueuer: enqueuer}, cfg.StorageRoot, cfg.UploadTempDir, logger)
	processor.SetArchiveExtractService(archiveImportSvc)
	if cfg.CloudDownloadEnabled {
		cloudClient := cloud.NewClient(cfg.CloudBrowserCDP, cfg.CloudBrowserApp)
		cloudRepo := repository.NewCloudDownloadRepository(pool)
		cloudSvc := services.NewCloudDownloadService(cloudRepo, cloudClient, cfg.DownloadDir, enqueuer, logger)
		processor.SetCloudDownloadService(cloudSvc)
		// Recover rows left behind by a hard deploy kill or a lost enqueue.
		// asynq task IDs make repeated reconciliation safe across workers.
		reconcileCtx, reconcileCancel := context.WithCancel(context.Background())
		defer reconcileCancel()
		go cloudSvc.RunReconciler(reconcileCtx, 2*time.Minute)
		logger.Info("cloud download worker enabled", "cdp", cfg.CloudBrowserCDP, "download_dir", cfg.DownloadDir)
	}
	processor.Register(mux)

	redisOpt := asynq.RedisClientOpt{Addr: cfg.RedisAddr, Password: cfg.RedisPassword}
	workerQueues := map[string]int{
		cfg.AsynqQueue:          10,
		cfg.ArchiveExtractQueue: 20,
	}
	queueNames := []string{cfg.AsynqQueue, cfg.ArchiveExtractQueue}
	if cfg.CloudDownloadEnabled {
		workerQueues[cfg.CloudDownloadQueue] = 1
		queueNames = append(queueNames, cfg.CloudDownloadQueue)
	}

	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: cfg.MaxTranscodeWorkers + cfg.ArchiveExtractWorkers,
			Queues:      workerQueues,
		},
	)
	logger.Info("worker started", "concurrency", cfg.MaxTranscodeWorkers+cfg.ArchiveExtractWorkers, "queues", queueNames)
	if err := srv.Run(mux); err != nil {
		return fmt.Errorf("run asynq server: %w", err)
	}
	return nil
}
