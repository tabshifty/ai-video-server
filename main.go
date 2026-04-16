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
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"video-server/internal/config"
	"video-server/internal/database"
	"video-server/internal/handlers"
	"video-server/internal/queue"
	"video-server/internal/repository"
	"video-server/internal/services"
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
		if err := runServer(cfg, repo, transSvc, logger); err != nil {
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

func runServer(cfg config.Config, repo *repository.VideoRepository, transSvc *services.TranscodeService, logger *slog.Logger) error {
	enqueuer := queue.NewEnqueuer(cfg.RedisAddr, cfg.RedisPassword, cfg.AsynqQueue)
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
	scrapeSvc.ConfigureAVScraper(cfg.AVScraperBaseURL, cfg.AVScraperUserAgent, cfg.AVScraperTimeout)
	appSvc := services.NewAppService(repo)

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
		enqueuer,
		logger,
		redisClient,
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.AsynqQueue,
		cfg.JWTSecret,
		cfg.PlayURLSignSecret,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		cfg.MaxVideoSize,
		cfg.StorageRoot,
		cfg.UploadTempDir,
		cfg.ServerLogPath,
		cfg.EnableSwagger,
	)
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
	enqueuer := queue.NewEnqueuer(cfg.RedisAddr, cfg.RedisPassword, cfg.AsynqQueue)
	defer enqueuer.Close()
	scrapeSvc := services.NewScraperService(repo, cfg.TMDBAPIKey, cfg.TMDBBaseURL, cfg.StorageRoot, cfg.PosterStoragePath, cfg.TMDBTimeout)
	scrapeSvc.ConfigureAVScraper(cfg.AVScraperBaseURL, cfg.AVScraperUserAgent, cfg.AVScraperTimeout)
	processor := queue.NewProcessor(repo, transSvc, scrapeSvc, enqueuer, logger)
	processor.Register(mux)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr, Password: cfg.RedisPassword},
		asynq.Config{
			Concurrency: cfg.MaxTranscodeWorkers,
			Queues: map[string]int{
				cfg.AsynqQueue: 10,
			},
		},
	)
	logger.Info("worker started", "concurrency", cfg.MaxTranscodeWorkers, "queue", cfg.AsynqQueue)
	if err := srv.Run(mux); err != nil {
		return fmt.Errorf("run asynq server: %w", err)
	}
	return nil
}
