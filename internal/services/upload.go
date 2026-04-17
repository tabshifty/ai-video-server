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
}

type LocalUploadInput struct {
	UserID        uuid.UUID
	FilePath      string
	Filename      string
	FileSize      int64
	Title         string
	Desc          string
	Type          string
	Tags          []string
	ActorIDs      []uuid.UUID
	ActorNames    []string
	CollectionIDs []uuid.UUID
	Hash          string
}

func NewUploadService(repo *repository.VideoRepository, uploadDir, storageRoot string, logger *slog.Logger) *UploadService {
	return &UploadService{repo: repo, uploadDir: uploadDir, storage: storageRoot, logger: logger}
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
		return UploadResult{
			VideoID:       existingID,
			Status:        existingVideo.Status,
			AlreadyExists: true,
			Enqueue:       false,
		}, nil
	}

	metaRaw, err := json.Marshal(map[string]any{
		"original_filename": in.Filename,
		"source_hash":       serverHash,
		"source_size":       info.Size(),
	})
	if err != nil {
		return UploadResult{}, fmt.Errorf("marshal upload metadata: %w", err)
	}

	videoID := uuid.New()
	finalTitle := strings.TrimSpace(in.Title)
	if finalTitle == "" {
		name := strings.TrimSpace(in.Filename)
		finalTitle = strings.TrimSuffix(name, filepath.Ext(name))
		if finalTitle == "" {
			finalTitle = "untitled"
		}
	}
	status := "uploaded"
	enqueueTranscode := true
	if in.Type == "movie" || in.Type == "episode" || in.Type == "av" {
		status = "scraping"
		enqueueTranscode = false
	}

	v := models.Video{
		ID:           videoID,
		UserID:       &in.UserID,
		Title:        finalTitle,
		Description:  in.Desc,
		Type:         in.Type,
		Status:       status,
		OriginalPath: in.FilePath,
		Metadata:     metaRaw,
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
			return UploadResult{
				VideoID:       existingID,
				Status:        existingVideo.Status,
				AlreadyExists: true,
				Enqueue:       false,
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
	}, nil
}

func (s *UploadService) SaveUpload(ctx context.Context, userID uuid.UUID, fileHeader *multipart.FileHeader, title, desc, typ string, tags []string, actorIDs []uuid.UUID, actorNames []string, collectionIDs []uuid.UUID, clientHash string, maxVideoSize int64) (UploadResult, error) {
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
		UserID:        userID,
		FilePath:      originalPath,
		Filename:      fileHeader.Filename,
		FileSize:      info.Size(),
		Title:         title,
		Desc:          desc,
		Type:          typ,
		Tags:          tags,
		ActorIDs:      actorIDs,
		ActorNames:    actorNames,
		CollectionIDs: collectionIDs,
		Hash:          clientHash,
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
