package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/pkg/ffmpeg"
)

const maxSubtitleUploadSize = 8 * 1024 * 1024

type SubtitleService struct {
	repo        *repository.VideoRepository
	storageRoot string
	logger      *slog.Logger
}

func NewSubtitleService(repo *repository.VideoRepository, storageRoot string, logger *slog.Logger) *SubtitleService {
	return &SubtitleService{repo: repo, storageRoot: storageRoot, logger: logger}
}

func (s *SubtitleService) SaveUploadedSubtitle(
	ctx context.Context,
	videoID uuid.UUID,
	fileHeader *multipart.FileHeader,
	languageCode, label string,
	requestDefault bool,
) (models.VideoSubtitle, error) {
	if fileHeader == nil {
		return models.VideoSubtitle{}, fmt.Errorf("subtitle file is required")
	}
	if fileHeader.Size <= 0 {
		return models.VideoSubtitle{}, fmt.Errorf("subtitle file is empty")
	}
	if fileHeader.Size > maxSubtitleUploadSize {
		return models.VideoSubtitle{}, fmt.Errorf("subtitle file too large")
	}
	format, mimeType, err := subtitleFileFormat(fileHeader.Filename)
	if err != nil {
		return models.VideoSubtitle{}, err
	}

	dir := filepath.Join(s.storageRoot, "subtitles", videoID.String(), "uploaded")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return models.VideoSubtitle{}, fmt.Errorf("create subtitle dir: %w", err)
	}
	subtitleID := uuid.New()
	targetPath := filepath.Join(dir, subtitleID.String()+"."+format)

	src, err := fileHeader.Open()
	if err != nil {
		return models.VideoSubtitle{}, fmt.Errorf("open subtitle upload: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(targetPath)
	if err != nil {
		return models.VideoSubtitle{}, fmt.Errorf("create subtitle file: %w", err)
	}
	if _, err := dst.ReadFrom(src); err != nil {
		dst.Close()
		_ = os.Remove(targetPath)
		return models.VideoSubtitle{}, fmt.Errorf("save subtitle file: %w", err)
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(targetPath)
		return models.VideoSubtitle{}, fmt.Errorf("close subtitle file: %w", err)
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		_ = os.Remove(targetPath)
		return models.VideoSubtitle{}, fmt.Errorf("stat subtitle file: %w", err)
	}
	hasUploadedDefault, err := s.repo.HasUploadedSubtitleDefault(ctx, videoID)
	if err != nil {
		_ = os.Remove(targetPath)
		return models.VideoSubtitle{}, err
	}
	shouldDefault := requestDefault || !hasUploadedDefault

	meta, err := json.Marshal(map[string]any{
		"original_filename": strings.TrimSpace(fileHeader.Filename),
	})
	if err != nil {
		_ = os.Remove(targetPath)
		return models.VideoSubtitle{}, fmt.Errorf("marshal subtitle metadata: %w", err)
	}
	subtitle := models.VideoSubtitle{
		ID:           subtitleID,
		VideoID:      videoID,
		SourceType:   "uploaded",
		Status:       "ready",
		LanguageCode: normalizeSubtitleLanguageCode(languageCode),
		Label:        strings.TrimSpace(label),
		Format:       format,
		MIMEType:     mimeType,
		StoredPath:   targetPath,
		FileSize:     info.Size(),
		IsDefault:    shouldDefault,
		SortOrder:    0,
		Metadata:     meta,
	}
	if err := s.repo.CreateVideoSubtitle(ctx, subtitle); err != nil {
		_ = os.Remove(targetPath)
		return models.VideoSubtitle{}, err
	}
	if shouldDefault {
		if err := s.repo.SetVideoSubtitleDefault(ctx, subtitle.ID, true); err != nil {
			_ = os.Remove(targetPath)
			_ = s.repo.DeleteVideoSubtitle(ctx, subtitle.ID)
			return models.VideoSubtitle{}, err
		}
	}
	return s.repo.GetVideoSubtitle(ctx, subtitle.ID)
}

func (s *SubtitleService) SyncEmbeddedSubtitles(ctx context.Context, videoID uuid.UUID, inputPath string) ([]models.VideoSubtitle, error) {
	if strings.TrimSpace(inputPath) == "" {
		return nil, fmt.Errorf("subtitle scan source path is empty")
	}
	existing, err := s.repo.ListVideoSubtitles(ctx, videoID)
	if err != nil {
		return nil, err
	}
	for _, subtitle := range existing {
		if subtitle.SourceType == "embedded" && strings.TrimSpace(subtitle.StoredPath) != "" {
			_ = os.Remove(subtitle.StoredPath)
		}
	}
	if err := s.repo.DeleteVideoSubtitlesBySourceType(ctx, videoID, "embedded"); err != nil {
		return nil, err
	}

	probes, err := ffmpeg.ProbeSubtitles(ctx, inputPath)
	if err != nil {
		return nil, err
	}
	out := make([]models.VideoSubtitle, 0, len(probes))
	for idx, probe := range probes {
		subtitle, ok, err := s.extractEmbeddedSubtitle(ctx, videoID, inputPath, probe, idx)
		if err != nil {
			s.logger.Warn("extract embedded subtitle failed", "video_id", videoID, "stream_index", probe.Index, "error", err)
			continue
		}
		if !ok {
			continue
		}
		out = append(out, subtitle)
	}
	return out, nil
}

func (s *SubtitleService) extractEmbeddedSubtitle(
	ctx context.Context,
	videoID uuid.UUID,
	inputPath string,
	probe ffmpeg.SubtitleProbe,
	sortOrder int,
) (models.VideoSubtitle, bool, error) {
	outputDir := filepath.Join(s.storageRoot, "subtitles", videoID.String(), "embedded")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return models.VideoSubtitle{}, false, fmt.Errorf("create embedded subtitle dir: %w", err)
	}
	subtitleID := uuid.New()
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s-%d.vtt", subtitleID.String(), probe.Index))
	if err := ffmpeg.ExtractSubtitleToWebVTT(ctx, inputPath, probe.Index, outputPath); err != nil {
		return models.VideoSubtitle{}, false, err
	}
	info, err := os.Stat(outputPath)
	if err != nil {
		_ = os.Remove(outputPath)
		return models.VideoSubtitle{}, false, fmt.Errorf("stat embedded subtitle output: %w", err)
	}
	label := strings.TrimSpace(probe.Title)
	if label == "" {
		label = "内嵌字幕"
		if strings.TrimSpace(probe.Language) != "" {
			label = probe.Language + "（内嵌）"
		}
	}
	meta, err := json.Marshal(map[string]any{
		"embedded_index": probe.Index,
		"source_codec":   strings.TrimSpace(probe.Codec),
		"source_title":   strings.TrimSpace(probe.Title),
	})
	if err != nil {
		_ = os.Remove(outputPath)
		return models.VideoSubtitle{}, false, fmt.Errorf("marshal embedded subtitle metadata: %w", err)
	}
	subtitle := models.VideoSubtitle{
		ID:           subtitleID,
		VideoID:      videoID,
		SourceType:   "embedded",
		Status:       "ready",
		LanguageCode: normalizeSubtitleLanguageCode(probe.Language),
		Label:        label,
		Format:       "vtt",
		MIMEType:     "text/vtt",
		StoredPath:   outputPath,
		FileSize:     info.Size(),
		IsDefault:    probe.IsDefault,
		SortOrder:    sortOrder,
		Metadata:     meta,
	}
	if err := s.repo.CreateVideoSubtitle(ctx, subtitle); err != nil {
		_ = os.Remove(outputPath)
		return models.VideoSubtitle{}, false, err
	}
	return subtitle, true, nil
}

func normalizeSubtitleLanguageCode(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	return strings.ReplaceAll(value, "_", "-")
}

func subtitleFileFormat(filename string) (string, string, error) {
	ext := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(filepath.Ext(filename))), ".")
	switch ext {
	case "srt":
		return "srt", "application/x-subrip", nil
	case "vtt":
		return "vtt", "text/vtt", nil
	default:
		return "", "", fmt.Errorf("unsupported subtitle format: %s", ext)
	}
}
