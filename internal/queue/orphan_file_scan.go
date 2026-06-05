package queue

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hibiken/asynq"

	"video-server/internal/models"
)

func (e *Enqueuer) EnqueueOrphanFileScan() error {
	_, err := e.client.Enqueue(
		asynq.NewTask(TypeOrphanFileScan, []byte("{}")),
		buildOrphanFileScanTaskOptions(e.queue)...,
	)
	if err != nil {
		return fmt.Errorf("enqueue orphan file scan task: %w", err)
	}
	return nil
}

func buildOrphanFileScanTaskOptions(queue string) []asynq.Option {
	return []asynq.Option{
		asynq.MaxRetry(3),
		asynq.ProcessIn(2 * time.Second),
		asynq.Queue(queue),
		asynq.Timeout(6 * time.Hour),
	}
}

func (p *Processor) HandleOrphanFileScan(ctx context.Context, task *asynq.Task) error {
	if p.repo == nil {
		return fmt.Errorf("orphan file scan processor not configured")
	}
	if strings.TrimSpace(p.storageRoot) == "" {
		return fmt.Errorf("orphan file scan storage root not configured")
	}

	if err := p.repo.MarkOrphanFileScanRunning(ctx); err != nil {
		return err
	}

	referencedPaths, err := p.repo.CollectReferencedFilePaths(ctx, p.storageRoot)
	if err != nil {
		_ = p.repo.FailOrphanFileScan(ctx, err.Error())
		return err
	}

	items, totalFiles, err := scanOrphanFiles(p.storageRoot, referencedPaths)
	if err != nil {
		_ = p.repo.FailOrphanFileScan(ctx, err.Error())
		return err
	}

	if err := p.repo.CompleteOrphanFileScan(ctx, totalFiles, int64(len(referencedPaths)), int64(len(items)), items); err != nil {
		_ = p.repo.FailOrphanFileScan(ctx, err.Error())
		return err
	}

	p.logger.Info("orphan file scan completed", "storage_root", p.storageRoot, "total_files", totalFiles, "referenced_files", len(referencedPaths), "orphan_files", len(items))
	return nil
}

func scanOrphanFiles(storageRoot string, referencedPaths map[string]struct{}) ([]models.AdminOrphanFileScanItem, int64, error) {
	roots := orphanFileScanRoots(storageRoot)
	items := make([]models.AdminOrphanFileScanItem, 0, 128)
	var totalFiles int64

	for _, root := range roots {
		if root == "" {
			continue
		}
		info, err := os.Stat(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, 0, fmt.Errorf("stat orphan file scan root %s: %w", root, err)
		}
		if !info.IsDir() {
			continue
		}

		walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			totalFiles++

			cleanPath := filepath.Clean(path)
			if _, ok := referencedPaths[cleanPath]; ok {
				return nil
			}

			fileInfo, err := d.Info()
			if err != nil {
				return err
			}
			relPath, err := filepath.Rel(storageRoot, cleanPath)
			if err != nil {
				relPath = cleanPath
			}
			items = append(items, models.AdminOrphanFileScanItem{
				FilePath:     cleanPath,
				RelativePath: relPath,
				SizeBytes:    fileInfo.Size(),
				ModTime:      fileInfo.ModTime(),
			})
			return nil
		})
		if walkErr != nil {
			return nil, 0, fmt.Errorf("walk orphan file scan root %s: %w", root, walkErr)
		}
	}

	return items, totalFiles, nil
}

func orphanFileScanRoots(storageRoot string) []string {
	storageRoot = strings.TrimSpace(storageRoot)
	if storageRoot == "" {
		return nil
	}
	return []string{
		filepath.Join(storageRoot, "videos"),
		filepath.Join(storageRoot, "images"),
		filepath.Join(storageRoot, "posters"),
		filepath.Join(storageRoot, "actors"),
		filepath.Join(storageRoot, "subtitles"),
		filepath.Join(storageRoot, "tv"),
	}
}
