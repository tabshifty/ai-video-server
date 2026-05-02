package services

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"video-server/internal/hashutil"
	"video-server/internal/repository"
	"video-server/pkg/ffmpeg"
)

const (
	FlickImportStatusImported = "imported"
	FlickImportStatusSkipped  = "skipped"
	FlickImportStatusDryRun   = "dry_run"
)

const (
	FlickImportReasonNoTags          = "no_tags"
	FlickImportReasonNotPlayable     = "not_playable"
	FlickImportReasonMissingVideo    = "missing_video"
	FlickImportReasonMissingCover    = "missing_cover"
	FlickImportReasonProbeFailed     = "probe_failed"
	FlickImportReasonDuplicateHash   = "duplicate_hash"
	FlickImportReasonMissingVideoDir = "missing_source_video_dir"
	FlickImportReasonMissingCoverDir = "missing_source_cover_dir"
	FlickImportReasonMissingStorage  = "missing_storage_root"
)

type flickImportRepo interface {
	FindVideoByHash(ctx context.Context, hash string, fileSize int64) (uuid.UUID, bool, error)
	CreateImportedReadyVideo(ctx context.Context, spec repository.ImportedReadyVideo) error
}

type FlickSourceVideo struct {
	SourceID    string
	MD5         string
	Tags        []string
	Description string
	CanPlay     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type FlickImportOptions struct {
	SourceVideoDir string
	SourceCoverDir string
	StorageRoot    string
	DryRun         bool
}

type FlickImportOutcome struct {
	Status          string
	Reason          string
	VideoID         uuid.UUID
	SourceVideoPath string
	SourceCoverPath string
	TargetVideoPath string
	TargetCoverPath string
}

type FlickImportService struct {
	repo      flickImportRepo
	probe     func(context.Context, string) (ffmpeg.VideoProbe, error)
	hashFile  func(string) (string, error)
	copyFile  func(string, string) error
	removeAll func(string) error
}

func NewFlickImportService(repo flickImportRepo) *FlickImportService {
	return &FlickImportService{
		repo:      repo,
		probe:     ffmpeg.Probe,
		hashFile:  hashutil.SHA256,
		copyFile:  copyFile,
		removeAll: os.RemoveAll,
	}
}

func (s *FlickImportService) ImportPlayableVideo(ctx context.Context, src FlickSourceVideo, opts FlickImportOptions) (FlickImportOutcome, error) {
	if !src.CanPlay {
		return FlickImportOutcome{Status: FlickImportStatusSkipped, Reason: FlickImportReasonNotPlayable}, nil
	}
	tags := normalizeImportTags(src.Tags)
	if len(tags) == 0 {
		return FlickImportOutcome{Status: FlickImportStatusSkipped, Reason: FlickImportReasonNoTags}, nil
	}
	if strings.TrimSpace(opts.SourceVideoDir) == "" {
		return FlickImportOutcome{Status: FlickImportStatusSkipped, Reason: FlickImportReasonMissingVideoDir}, nil
	}
	if strings.TrimSpace(opts.SourceCoverDir) == "" {
		return FlickImportOutcome{Status: FlickImportStatusSkipped, Reason: FlickImportReasonMissingCoverDir}, nil
	}
	if strings.TrimSpace(opts.StorageRoot) == "" && !opts.DryRun {
		return FlickImportOutcome{Status: FlickImportStatusSkipped, Reason: FlickImportReasonMissingStorage}, nil
	}

	videoPath, ok, err := resolveFlickPlayableVideoPath(opts.SourceVideoDir, src.MD5)
	if err != nil {
		return FlickImportOutcome{}, err
	}
	if !ok {
		return FlickImportOutcome{Status: FlickImportStatusSkipped, Reason: FlickImportReasonMissingVideo}, nil
	}
	coverPath := filepath.Join(opts.SourceCoverDir, src.MD5+".jpg")
	if _, err := os.Stat(coverPath); err != nil {
		if os.IsNotExist(err) {
			return FlickImportOutcome{
				Status:          FlickImportStatusSkipped,
				Reason:          FlickImportReasonMissingCover,
				SourceVideoPath: videoPath,
			}, nil
		}
		return FlickImportOutcome{}, fmt.Errorf("stat source cover: %w", err)
	}

	probe, err := s.probe(ctx, videoPath)
	if err != nil {
		return FlickImportOutcome{
			Status:          FlickImportStatusSkipped,
			Reason:          FlickImportReasonProbeFailed,
			SourceVideoPath: videoPath,
			SourceCoverPath: coverPath,
		}, nil
	}

	info, err := os.Stat(videoPath)
	if err != nil {
		return FlickImportOutcome{}, fmt.Errorf("stat source video: %w", err)
	}
	sha256Value, err := s.hashFile(videoPath)
	if err != nil {
		return FlickImportOutcome{}, fmt.Errorf("hash source video: %w", err)
	}
	existingID, exists, err := s.repo.FindVideoByHash(ctx, sha256Value, info.Size())
	if err != nil {
		return FlickImportOutcome{}, fmt.Errorf("find existing hash: %w", err)
	}
	if exists {
		return FlickImportOutcome{
			Status:          FlickImportStatusSkipped,
			Reason:          FlickImportReasonDuplicateHash,
			VideoID:         existingID,
			SourceVideoPath: videoPath,
			SourceCoverPath: coverPath,
		}, nil
	}

	videoID := uuid.New()
	targetVideoPath, targetCoverPath := buildImportedAssetPaths(opts.StorageRoot, videoID, filepath.Ext(videoPath))
	outcome := FlickImportOutcome{
		VideoID:         videoID,
		SourceVideoPath: videoPath,
		SourceCoverPath: coverPath,
		TargetVideoPath: targetVideoPath,
		TargetCoverPath: targetCoverPath,
	}
	if opts.DryRun {
		outcome.Status = FlickImportStatusDryRun
		return outcome, nil
	}

	if err := s.copyFile(videoPath, targetVideoPath); err != nil {
		return FlickImportOutcome{}, fmt.Errorf("copy source video: %w", err)
	}
	if err := s.copyFile(coverPath, targetCoverPath); err != nil {
		_ = s.removeAll(filepath.Dir(targetVideoPath))
		return FlickImportOutcome{}, fmt.Errorf("copy source cover: %w", err)
	}

	createdAt := src.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := src.UpdatedAt
	if updatedAt.IsZero() || updatedAt.Before(createdAt) {
		updatedAt = createdAt
	}

	spec := repository.ImportedReadyVideo{
		ID:              videoID,
		Title:           deriveImportedTitle(videoPath, src.MD5),
		Description:     strings.TrimSpace(src.Description),
		Type:            "short",
		Status:          "ready",
		OriginalPath:    targetVideoPath,
		TranscodedPath:  targetVideoPath,
		ThumbnailPath:   targetCoverPath,
		DurationSeconds: int(math.Round(probe.Duration)),
		Width:           probe.Width,
		Height:          probe.Height,
		Hash:            sha256Value,
		FileSize:        info.Size(),
		Tags:            tags,
		Metadata: map[string]any{
			"migration_source":       "flick-server",
			"migration_source_id":    strings.TrimSpace(src.SourceID),
			"source_md5":             strings.TrimSpace(src.MD5),
			"source_video_path":      videoPath,
			"source_cover_path":      coverPath,
			"source_created_at":      createdAt.Format(time.RFC3339),
			"source_updated_at":      updatedAt.Format(time.RFC3339),
			"source_duration_second": probe.Duration,
		},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	if err := s.repo.CreateImportedReadyVideo(ctx, spec); err != nil {
		_ = s.removeAll(filepath.Dir(targetVideoPath))
		return FlickImportOutcome{}, fmt.Errorf("create imported ready video: %w", err)
	}
	outcome.Status = FlickImportStatusImported
	return outcome, nil
}

func normalizeImportTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	out := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}

func resolveFlickPlayableVideoPath(root, md5 string) (string, bool, error) {
	pattern := filepath.Join(strings.TrimSpace(root), strings.TrimSpace(md5)+".*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", false, fmt.Errorf("glob flick playable video path: %w", err)
	}
	if len(matches) == 0 {
		return "", false, nil
	}
	sort.Strings(matches)
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			return "", false, fmt.Errorf("stat flick playable video candidate: %w", err)
		}
		if !info.IsDir() {
			return match, true, nil
		}
	}
	return "", false, nil
}

func deriveImportedTitle(videoPath, md5 string) string {
	base := strings.TrimSpace(filepath.Base(videoPath))
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if base != "" {
		return base
	}
	if strings.TrimSpace(md5) != "" {
		return strings.TrimSpace(md5)
	}
	return "untitled"
}

func buildImportedAssetPaths(storageRoot string, videoID uuid.UUID, ext string) (string, string) {
	normalizedExt := strings.ToLower(strings.TrimSpace(ext))
	if normalizedExt == "" {
		normalizedExt = ".mp4"
	}
	dir := filepath.Join(storageRoot, "videos", videoID.String())
	return filepath.Join(dir, "play"+normalizedExt), filepath.Join(dir, "thumb.jpg")
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("mkdir target dir: %w", err)
	}
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create target file: %w", err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("copy file bytes: %w", err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close target file: %w", err)
	}
	return nil
}
