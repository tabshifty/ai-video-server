package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"video-server/internal/hashutil"
	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/pkg/oshash"
)

var (
	ErrHashMismatch     = errors.New("hash mismatch")
	ErrFileTooLarge     = errors.New("file size exceeds limit")
	ErrInvalidType      = errors.New("invalid video type")
	ErrInvalidUpload    = errors.New("invalid upload")
	ErrHashTypeConflict = errors.New("hash already exists with different video type")
)

// UploadService stores uploaded files and creates video records.
type UploadService struct {
	repo      *repository.VideoRepository
	uploadDir string
	storage   string
	logger    *slog.Logger
}

type UploadResult struct {
	VideoID       uuid.UUID
	Status        string
	AlreadyExists bool
	Enqueue       bool
	InputPath     string
	OutputDir     string
	TargetFormat  string
	OSHash        string
}

type UploadPreview struct {
	VideoID       uuid.UUID
	AlreadyExists bool
	Status        string
}

type LocalUploadInput struct {
	UserID            uuid.UUID
	FilePath          string
	Filename          string
	FileSize          int64
	Title             string
	Desc              string
	Type              string
	Tags              []string
	ActorIDs          []uuid.UUID
	ActorNames        []string
	CollectionIDs     []uuid.UUID
	ImageCollectionID *uuid.UUID
	SiteCategory      string
	Hash              string
}

func NewUploadService(repo *repository.VideoRepository, uploadDir, storageRoot string, logger *slog.Logger) *UploadService {
	return &UploadService{repo: repo, uploadDir: uploadDir, storage: storageRoot, logger: logger}
}

func (s *UploadService) PreviewUploadedFile(ctx context.Context, in LocalUploadInput) (UploadPreview, error) {
	if in.Type != "short" && in.Type != "movie" && in.Type != "episode" && in.Type != "av" {
		return UploadPreview{}, ErrInvalidType
	}
	info, err := os.Stat(in.FilePath)
	if err != nil {
		return UploadPreview{}, fmt.Errorf("stat uploaded file: %w", err)
	}
	serverHash, err := hashutil.SHA256(in.FilePath)
	if err != nil {
		return UploadPreview{}, fmt.Errorf("calculate file hash: %w", err)
	}
	if !strings.EqualFold(strings.TrimSpace(in.Hash), serverHash) {
		return UploadPreview{}, ErrHashMismatch
	}
	if existingID, exists, err := s.repo.FindVideoByHash(ctx, serverHash, info.Size()); err != nil {
		return UploadPreview{}, err
	} else if exists {
		existingVideo, getErr := s.repo.GetVideoByID(ctx, existingID)
		if getErr != nil {
			return UploadPreview{}, getErr
		}
		if existingVideo.Type != in.Type {
			return UploadPreview{}, ErrHashTypeConflict
		}
		return UploadPreview{
			VideoID:       existingID,
			AlreadyExists: true,
			Status:        existingVideo.Status,
		}, nil
	}
	return UploadPreview{
		VideoID:       uuid.New(),
		AlreadyExists: false,
		Status:        predictedUploadStatus(in.Type),
	}, nil
}

func (s *UploadService) SaveUploadedFile(ctx context.Context, in LocalUploadInput, maxVideoSize int64) (UploadResult, error) {
	if in.Type != "short" && in.Type != "movie" && in.Type != "episode" && in.Type != "av" {
		return UploadResult{}, ErrInvalidType
	}
	info, err := os.Stat(in.FilePath)
	if err != nil {
		return UploadResult{}, fmt.Errorf("stat uploaded file: %w", err)
	}
	if maxVideoSize > 0 && info.Size() > maxVideoSize {
		return UploadResult{}, ErrFileTooLarge
	}
	serverHash, err := hashutil.SHA256(in.FilePath)
	if err != nil {
		return UploadResult{}, fmt.Errorf("calculate file hash: %w", err)
	}
	if !strings.EqualFold(strings.TrimSpace(in.Hash), serverHash) {
		return UploadResult{}, ErrHashMismatch
	}
	siteCategory := normalizeAVSiteCategory(in.SiteCategory)
	if in.Type == "av" && siteCategory == "" {
		siteCategory = avSiteCategoryJapanese
	}

	if existingID, exists, err := s.repo.FindVideoByHash(ctx, serverHash, info.Size()); err != nil {
		return UploadResult{}, err
	} else if exists {
		existingVideo, getErr := s.repo.GetVideoByID(ctx, existingID)
		if getErr != nil {
			return UploadResult{}, getErr
		}
		if existingVideo.Type != in.Type {
			return UploadResult{}, ErrHashTypeConflict
		}
		if len(in.CollectionIDs) > 0 {
			if err := s.repo.AddVideoCollectionsByIDsAnyType(ctx, existingID, in.CollectionIDs); err != nil {
				return UploadResult{}, err
			}
		}
		if err := s.setExistingVideoImageCollectionIfEmpty(ctx, existingID, existingVideo, in.ImageCollectionID); err != nil {
			return UploadResult{}, err
		}
		if shouldRefreshExistingVideoOriginalPath(existingVideo.Status) && strings.TrimSpace(existingVideo.OriginalPath) != strings.TrimSpace(in.FilePath) {
			if err := s.repo.UpdateVideoOriginalPath(ctx, existingID, in.FilePath); err != nil {
				return UploadResult{}, err
			}
			existingVideo.OriginalPath = in.FilePath
		}
		return UploadResult{
			VideoID:       existingID,
			Status:        existingVideo.Status,
			AlreadyExists: true,
			Enqueue:       false,
			OSHash:        existingVideo.OSHash,
		}, nil
	}
	rawImageCollectionIDs := make([]uuid.UUID, 0, 1)
	if in.ImageCollectionID != nil {
		rawImageCollectionIDs = append(rawImageCollectionIDs, *in.ImageCollectionID)
	}
	resolvedImageCollectionID, err := s.repo.ResolveVideoImageCollectionID(ctx, rawImageCollectionIDs)
	if err != nil {
		return UploadResult{}, err
	}

	videoID := uuid.New()
	osHash := ""
	if in.Type == "av" && siteCategory == avSiteCategoryWestern {
		if h, err := oshash.Compute(in.FilePath); err == nil {
			osHash = h
		} else {
			s.logger.Warn("oshash compute failed", "video_id", videoID, "error", err)
		}
	}

	meta := map[string]any{
		"original_filename": in.Filename,
		"source_hash":       serverHash,
		"source_size":       info.Size(),
	}
	if in.Type == "av" {
		meta["site_category"] = siteCategory
	}
	metaRaw, err := json.Marshal(meta)
	if err != nil {
		return UploadResult{}, fmt.Errorf("marshal upload metadata: %w", err)
	}

	finalTitle := strings.TrimSpace(in.Title)
	if finalTitle == "" {
		name := strings.TrimSpace(in.Filename)
		finalTitle = strings.TrimSuffix(name, filepath.Ext(name))
		if finalTitle == "" {
			finalTitle = "untitled"
		}
	}
	status := predictedUploadStatus(in.Type)
	enqueueTranscode := status == "uploaded"

	v := models.Video{
		ID:                videoID,
		UserID:            &in.UserID,
		Title:             finalTitle,
		Description:       in.Desc,
		Type:              in.Type,
		Status:            status,
		OriginalPath:      in.FilePath,
		OSHash:            osHash,
		Metadata:          metaRaw,
		ImageCollectionID: resolvedImageCollectionID,
	}

	if err := s.repo.CreateVideo(ctx, v); err != nil {
		return UploadResult{}, err
	}
	if err := s.repo.AddTags(ctx, videoID, in.Tags); err != nil {
		_ = s.repo.DeleteVideoByID(ctx, videoID)
		return UploadResult{}, err
	}
	if err := s.repo.ReplaceVideoActorsByInput(ctx, videoID, in.ActorIDs, in.ActorNames, "upload_manual"); err != nil {
		_ = s.repo.DeleteVideoByID(ctx, videoID)
		return UploadResult{}, err
	}
	if err := s.repo.ReplaceVideoCollectionsByIDs(ctx, videoID, in.Type, in.CollectionIDs); err != nil {
		_ = s.repo.DeleteVideoByID(ctx, videoID)
		return UploadResult{}, err
	}
	if err := s.repo.InsertFileHash(ctx, serverHash, videoID, info.Size()); err != nil {
		if repository.IsUniqueViolation(err) {
			_ = s.repo.DeleteVideoByID(ctx, videoID)
			existingID, getErr := s.repo.GetVideoIDByHash(ctx, serverHash)
			if getErr != nil {
				return UploadResult{}, getErr
			}
			existingVideo, getVideoErr := s.repo.GetVideoByID(ctx, existingID)
			if getVideoErr != nil {
				return UploadResult{}, getVideoErr
			}
			if existingVideo.Type != in.Type {
				return UploadResult{}, ErrHashTypeConflict
			}
			if err := s.setExistingVideoImageCollectionIfEmpty(ctx, existingID, existingVideo, in.ImageCollectionID); err != nil {
				return UploadResult{}, err
			}
			return UploadResult{
				VideoID:       existingID,
				Status:        existingVideo.Status,
				AlreadyExists: true,
				Enqueue:       false,
				OSHash:        existingVideo.OSHash,
			}, nil
		}
		_ = s.repo.DeleteVideoByID(ctx, videoID)
		return UploadResult{}, err
	}

	return UploadResult{
		VideoID:       videoID,
		Status:        status,
		AlreadyExists: false,
		Enqueue:       enqueueTranscode,
		InputPath:     in.FilePath,
		OutputDir:     filepath.Join(s.storage, "videos"),
		TargetFormat:  "mp4",
		OSHash:        osHash,
	}, nil
}

func (s *UploadService) setExistingVideoImageCollectionIfEmpty(ctx context.Context, existingID uuid.UUID, existingVideo models.Video, imageCollectionID *uuid.UUID) error {
	if imageCollectionID == nil || existingVideo.ImageCollectionID != nil {
		return nil
	}
	return s.repo.SetVideoImageCollectionIfEmpty(ctx, existingID, *imageCollectionID)
}

func (s *UploadService) SaveUpload(ctx context.Context, userID uuid.UUID, fileHeader *multipart.FileHeader, title, desc, typ string, tags []string, actorIDs []uuid.UUID, actorNames []string, collectionIDs []uuid.UUID, imageCollectionID *uuid.UUID, siteCategory, clientHash string, maxVideoSize int64) (UploadResult, error) {
	if fileHeader == nil {
		return UploadResult{}, ErrInvalidUpload
	}
	if typ != "short" && typ != "movie" && typ != "episode" && typ != "av" {
		return UploadResult{}, ErrInvalidType
	}
	if maxVideoSize > 0 && fileHeader.Size > maxVideoSize {
		return UploadResult{}, ErrFileTooLarge
	}
	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		return UploadResult{}, fmt.Errorf("create upload dir: %w", err)
	}

	tempID := uuid.New()
	originalPath := filepath.Join(s.uploadDir, tempID.String()+".tmp")

	src, err := fileHeader.Open()
	if err != nil {
		return UploadResult{}, fmt.Errorf("open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(originalPath)
	if err != nil {
		return UploadResult{}, fmt.Errorf("create local file: %w", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		_ = os.Remove(originalPath)
		return UploadResult{}, fmt.Errorf("save upload file: %w", err)
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(originalPath)
		return UploadResult{}, fmt.Errorf("close upload file: %w", err)
	}

	info, err := os.Stat(originalPath)
	if err != nil {
		_ = os.Remove(originalPath)
		return UploadResult{}, fmt.Errorf("stat uploaded file: %w", err)
	}
	if maxVideoSize > 0 && info.Size() > maxVideoSize {
		_ = os.Remove(originalPath)
		return UploadResult{}, ErrFileTooLarge
	}

	result, err := s.SaveUploadedFile(ctx, LocalUploadInput{
		UserID:            userID,
		FilePath:          originalPath,
		Filename:          fileHeader.Filename,
		FileSize:          info.Size(),
		Title:             title,
		Desc:              desc,
		Type:              typ,
		Tags:              tags,
		ActorIDs:          actorIDs,
		ActorNames:        actorNames,
		CollectionIDs:     collectionIDs,
		ImageCollectionID: imageCollectionID,
		SiteCategory:      siteCategory,
		Hash:              clientHash,
	}, maxVideoSize)
	if err != nil {
		_ = os.Remove(originalPath)
		return UploadResult{}, err
	}
	if result.AlreadyExists {
		_ = os.Remove(originalPath)
	}
	s.logger.Info("upload saved", "video_id", result.VideoID, "path", originalPath)
	return result, nil
}

func predictedUploadStatus(videoType string) string {
	switch strings.TrimSpace(videoType) {
	case "movie", "episode", "av":
		return "scraping"
	default:
		return "uploaded"
	}
}

func shouldRefreshExistingVideoOriginalPath(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "uploaded", "scraping", "tv_pending", "av_scrape_pending", "failed":
		return true
	default:
		return false
	}
}
