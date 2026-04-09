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

func NewUploadService(repo *repository.VideoRepository, uploadDir, storageRoot string, logger *slog.Logger) *UploadService {
	return &UploadService{repo: repo, uploadDir: uploadDir, storage: storageRoot, logger: logger}
}

func (s *UploadService) SaveUpload(ctx context.Context, userID uuid.UUID, fileHeader *multipart.FileHeader, title, desc, typ string, tags []string, clientHash string, maxVideoSize int64) (UploadResult, error) {
	if fileHeader == nil {
		return UploadResult{}, ErrInvalidUpload
	}
	if typ != "short" && typ != "movie" && typ != "episode" {
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

	serverHash, err := hashutil.SHA256(originalPath)
	if err != nil {
		_ = os.Remove(originalPath)
		return UploadResult{}, fmt.Errorf("calculate file hash: %w", err)
	}
	if !strings.EqualFold(strings.TrimSpace(clientHash), serverHash) {
		_ = os.Remove(originalPath)
		return UploadResult{}, ErrHashMismatch
	}

	if existingID, exists, err := s.repo.FindVideoByHash(ctx, serverHash, info.Size()); err != nil {
		_ = os.Remove(originalPath)
		return UploadResult{}, err
	} else if exists {
		existingVideo, getErr := s.repo.GetVideoByID(ctx, existingID)
		if getErr != nil {
			_ = os.Remove(originalPath)
			return UploadResult{}, getErr
		}
		if existingVideo.Type != typ {
			_ = os.Remove(originalPath)
			return UploadResult{}, ErrHashTypeConflict
		}
		_ = os.Remove(originalPath)
		return UploadResult{
			VideoID:       existingID,
			Status:        existingVideo.Status,
			AlreadyExists: true,
			Enqueue:       false,
		}, nil
	}

	metaRaw, err := json.Marshal(map[string]any{
		"original_filename": fileHeader.Filename,
		"source_hash":       serverHash,
		"source_size":       info.Size(),
	})
	if err != nil {
		_ = os.Remove(originalPath)
		return UploadResult{}, fmt.Errorf("marshal upload metadata: %w", err)
	}

	videoID := uuid.New()
	finalTitle := strings.TrimSpace(title)
	if finalTitle == "" {
		name := strings.TrimSpace(fileHeader.Filename)
		finalTitle = strings.TrimSuffix(name, filepath.Ext(name))
		if finalTitle == "" {
			finalTitle = "untitled"
		}
	}
	status := "uploaded"
	enqueueTranscode := true
	if typ == "movie" || typ == "episode" {
		status = "scraping"
		enqueueTranscode = false
	}

	v := models.Video{
		ID:           videoID,
		UserID:       &userID,
		Title:        finalTitle,
		Description:  desc,
		Type:         typ,
		Status:       status,
		OriginalPath: originalPath,
		Metadata:     metaRaw,
	}

	if err := s.repo.CreateVideo(ctx, v); err != nil {
		_ = os.Remove(originalPath)
		return UploadResult{}, err
	}
	if err := s.repo.AddTags(ctx, videoID, tags); err != nil {
		_ = s.repo.DeleteVideoByID(ctx, videoID)
		_ = os.Remove(originalPath)
		return UploadResult{}, err
	}
	if err := s.repo.InsertFileHash(ctx, serverHash, videoID, info.Size()); err != nil {
		if repository.IsUniqueViolation(err) {
			_ = s.repo.DeleteVideoByID(ctx, videoID)
			_ = os.Remove(originalPath)
			existingID, getErr := s.repo.GetVideoIDByHash(ctx, serverHash)
			if getErr != nil {
				return UploadResult{}, getErr
			}
			existingVideo, getVideoErr := s.repo.GetVideoByID(ctx, existingID)
			if getVideoErr != nil {
				return UploadResult{}, getVideoErr
			}
			if existingVideo.Type != typ {
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
		_ = os.Remove(originalPath)
		return UploadResult{}, err
	}

	s.logger.Info("upload saved", "video_id", videoID, "path", originalPath)
	return UploadResult{
		VideoID:       videoID,
		Status:        status,
		AlreadyExists: false,
		Enqueue:       enqueueTranscode,
		InputPath:     originalPath,
		OutputDir:     filepath.Join(s.storage, "videos"),
		TargetFormat:  "mp4",
	}, nil
}
