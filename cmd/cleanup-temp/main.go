package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"

	"video-server/internal/config"
	"video-server/internal/database"
	"video-server/internal/repository"
)

func main() {
	_ = godotenv.Load()

	var (
		hours  = flag.Int("older-than-hours", 24, "remove files older than N hours")
		dryRun = flag.Bool("dry-run", false, "print files only, do not delete")
	)
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := database.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		slog.Error("connect postgres failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewVideoRepository(pool)
	activePaths, err := repo.ListActiveOriginalPaths(ctx)
	if err != nil {
		slog.Error("list active original paths failed", "error", err)
		os.Exit(1)
	}

	protected := make(map[string]struct{}, len(activePaths))
	for _, p := range activePaths {
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		protected[abs] = struct{}{}
	}

	cutoff := time.Now().Add(-time.Duration(*hours) * time.Hour)
	var scanned, deleted int
	walkErr := filepath.WalkDir(cfg.UploadTempDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		scanned++

		info, statErr := d.Info()
		if statErr != nil {
			return nil
		}
		if info.ModTime().After(cutoff) {
			return nil
		}

		abs, absErr := filepath.Abs(path)
		if absErr == nil {
			if _, ok := protected[abs]; ok {
				return nil
			}
		}

		if *dryRun {
			fmt.Println(path)
			return nil
		}
		if rmErr := os.Remove(path); rmErr != nil {
			slog.Warn("remove temp file failed", "path", path, "error", rmErr)
			return nil
		}
		deleted++
		return nil
	})
	if walkErr != nil && !os.IsNotExist(walkErr) {
		slog.Error("scan temp dir failed", "error", walkErr)
		os.Exit(1)
	}

	slog.Info("cleanup temp finished", "scanned", scanned, "deleted", deleted, "dry_run", *dryRun)
}
