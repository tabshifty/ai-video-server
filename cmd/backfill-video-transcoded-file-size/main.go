package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"video-server/internal/database"
	"video-server/internal/repository"
	"video-server/internal/services"
)

type backfillRepository interface {
	ListTranscodedVideoFiles(ctx context.Context, afterID uuid.UUID, limit int) ([]repository.TranscodedVideoFile, error)
	UpdateVideoTranscodedFileSize(ctx context.Context, videoID uuid.UUID, size int64) error
}

type backfillOptions struct {
	StorageRoot string
	BatchSize   int
	Apply       bool
}

type backfillFailure struct {
	VideoID string `json:"video_id"`
	Path    string `json:"path"`
	Error   string `json:"error"`
}

type backfillReport struct {
	StartedAt   time.Time         `json:"started_at"`
	FinishedAt  time.Time         `json:"finished_at"`
	Apply       bool              `json:"apply"`
	Scanned     int64             `json:"scanned"`
	WouldUpdate int64             `json:"would_update"`
	Updated     int64             `json:"updated"`
	Failures    []backfillFailure `json:"failures"`
}

type incompleteBackfillError struct {
	failures int
}

func (e incompleteBackfillError) Error() string {
	return fmt.Sprintf("backfill incomplete: %d file(s) failed", e.failures)
}

func main() {
	loadEnvironment()

	var (
		postgresDSN string
		storageRoot string
		batchSize   int
		apply       bool
	)
	flag.StringVar(&postgresDSN, "postgres-dsn", os.Getenv("POSTGRES_DSN"), "Postgres DSN (defaults to POSTGRES_DSN)")
	flag.StringVar(&storageRoot, "storage-root", os.Getenv("STORAGE_ROOT"), "storage root (defaults to STORAGE_ROOT)")
	flag.IntVar(&batchSize, "batch-size", 100, "records processed per database page")
	flag.BoolVar(&apply, "apply", false, "write file sizes; default is dry-run")
	flag.Parse()

	if strings.TrimSpace(postgresDSN) == "" {
		slog.Error("POSTGRES_DSN is required")
		os.Exit(2)
	}
	if strings.TrimSpace(storageRoot) == "" {
		slog.Error("STORAGE_ROOT is required")
		os.Exit(2)
	}
	if batchSize <= 0 {
		slog.Error("batch-size must be positive")
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := database.NewPostgres(ctx, postgresDSN)
	if err != nil {
		slog.Error("connect postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	report, backfillErr := runBackfill(ctx, repository.NewVideoRepository(pool), backfillOptions{
		StorageRoot: storageRoot,
		BatchSize:   batchSize,
		Apply:       apply,
	})
	if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
		slog.Error("write backfill report", "error", err)
		os.Exit(1)
	}
	if backfillErr != nil {
		slog.Error("backfill failed", "error", backfillErr)
		os.Exit(1)
	}
}

func loadEnvironment() {
	if envFile := strings.TrimSpace(os.Getenv("ENV_FILE")); envFile != "" {
		_ = godotenv.Overload(envFile)
		return
	}
	_ = godotenv.Load()
}

func runBackfill(ctx context.Context, repo backfillRepository, options backfillOptions) (report backfillReport, err error) {
	report = backfillReport{
		StartedAt: time.Now().UTC(),
		Apply:     options.Apply,
		Failures:  make([]backfillFailure, 0),
	}
	defer func() { report.FinishedAt = time.Now().UTC() }()

	if strings.TrimSpace(options.StorageRoot) == "" {
		return report, fmt.Errorf("storage root is required")
	}
	if options.BatchSize <= 0 {
		return report, fmt.Errorf("batch size must be positive")
	}

	afterID := uuid.Nil
	for {
		items, err := repo.ListTranscodedVideoFiles(ctx, afterID, options.BatchSize)
		if err != nil {
			return report, err
		}
		if len(items) == 0 {
			break
		}

		for _, item := range items {
			report.Scanned++
			size, err := resolveBackfillFileSize(options.StorageRoot, item.TranscodedPath)
			if err != nil {
				report.Failures = append(report.Failures, backfillFailure{
					VideoID: item.ID.String(),
					Path:    item.TranscodedPath,
					Error:   err.Error(),
				})
				continue
			}
			if !options.Apply {
				report.WouldUpdate++
				continue
			}
			if err := repo.UpdateVideoTranscodedFileSize(ctx, item.ID, size); err != nil {
				report.Failures = append(report.Failures, backfillFailure{
					VideoID: item.ID.String(),
					Path:    item.TranscodedPath,
					Error:   err.Error(),
				})
				continue
			}
			report.Updated++
		}
		afterID = items[len(items)-1].ID
	}

	if len(report.Failures) > 0 {
		return report, incompleteBackfillError{failures: len(report.Failures)}
	}
	return report, nil
}

func resolveBackfillFileSize(storageRoot, storedPath string) (int64, error) {
	storageRoot = strings.TrimSpace(storageRoot)
	storedPath = strings.TrimSpace(storedPath)
	if storageRoot == "" || storedPath == "" {
		return 0, fmt.Errorf("storage root and transcoded path are required")
	}
	if !filepath.IsAbs(storedPath) {
		return 0, fmt.Errorf("transcoded path must be absolute: %s", storedPath)
	}

	videosRoot, err := filepath.EvalSymlinks(filepath.Join(storageRoot, "videos"))
	if err != nil {
		return 0, fmt.Errorf("resolve videos root: %w", err)
	}
	resolvedPath, err := filepath.EvalSymlinks(storedPath)
	if err != nil {
		return 0, fmt.Errorf("resolve transcoded path: %w", err)
	}
	if !pathWithinRoot(videosRoot, resolvedPath) {
		return 0, fmt.Errorf("transcoded path is outside storage videos root: %s", storedPath)
	}
	return services.RequiredRegularFileSize(resolvedPath)
}

func pathWithinRoot(root, path string) bool {
	relativePath, err := filepath.Rel(root, path)
	if err != nil || relativePath == "." || relativePath == ".." {
		return false
	}
	return !strings.HasPrefix(relativePath, ".."+string(filepath.Separator))
}
