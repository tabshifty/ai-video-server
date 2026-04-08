package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
)

// UploadService stores uploaded files and creates video records.
type UploadService struct {
	repo      *repository.VideoRepository
	uploadDir string
	logger    *slog.Logger
}

func NewUploadService(repo *repository.VideoRepository, uploadDir string, logger *slog.Logger) *UploadService {
	return &UploadService{repo: repo, uploadDir: uploadDir, logger: logger}
}

func (s *UploadService) SaveUpload(ctx context.Context, userID uuid.UUID, fileHeader *multipart.FileHeader, title, desc, typ string, tags []string) (models.Video, error) {
	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		return models.Video{}, fmt.Errorf("create upload dir: %w", err)
	}

	videoID := uuid.New()
	safeName := strings.ReplaceAll(fileHeader.Filename, " ", "_")
	originalPath := filepath.Join(s.uploadDir, videoID.String()+"_"+safeName)

	src, err := fileHeader.Open()
	if err != nil {
		return models.Video{}, fmt.Errorf("open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(originalPath)
	if err != nil {
		return models.Video{}, fmt.Errorf("create local file: %w", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return models.Video{}, fmt.Errorf("save upload file: %w", err)
	}
	if err := dst.Close(); err != nil {
		return models.Video{}, fmt.Errorf("close upload file: %w", err)
	}

	metaRaw, err := json.Marshal(map[string]any{"original_filename": fileHeader.Filename})
	if err != nil {
		return models.Video{}, fmt.Errorf("marshal upload metadata: %w", err)
	}

	v := models.Video{
		ID:           videoID,
		UserID:       &userID,
		Title:        title,
		Description:  desc,
		Type:         typ,
		Status:       "uploaded",
		OriginalPath: originalPath,
		Metadata:     metaRaw,
	}

	if err := s.repo.CreateVideo(ctx, v); err != nil {
		return models.Video{}, err
	}
	if err := s.repo.AddTags(ctx, videoID, tags); err != nil {
		return models.Video{}, err
	}

	s.logger.Info("upload saved", "video_id", videoID, "path", originalPath)
	return v, nil
}
