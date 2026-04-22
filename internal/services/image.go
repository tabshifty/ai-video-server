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

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"

	"video-server/internal/hashutil"
	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/pkg/ffmpeg"
)

var (
	ErrInvalidImageType = errors.New("invalid image type")
	ErrImageTooLarge    = errors.New("image size exceeds limit")
)

type ImageService struct {
	repo      *repository.VideoRepository
	uploadDir string
	storage   string
	logger    *slog.Logger
}

type SaveImageInput struct {
	UserID        uuid.UUID
	Title         string
	Description   string
	ActorIDs      []uuid.UUID
	ActorNames    []string
	CollectionIDs []uuid.UUID
}

type SaveImageResult struct {
	ImageID       uuid.UUID
	AlreadyExists bool
	Status        string
	StoredPath    string
	StoredMIME    string
}

func NewImageService(repo *repository.VideoRepository, uploadDir, storageRoot string, logger *slog.Logger) *ImageService {
	return &ImageService{repo: repo, uploadDir: uploadDir, storage: storageRoot, logger: logger}
}

func (s *ImageService) SaveUpload(ctx context.Context, in SaveImageInput, fileHeader *multipart.FileHeader, maxSize int64) (SaveImageResult, error) {
	if fileHeader == nil {
		return SaveImageResult{}, ErrInvalidImageType
	}
	if maxSize > 0 && fileHeader.Size > maxSize {
		return SaveImageResult{}, ErrImageTooLarge
	}
	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		return SaveImageResult{}, fmt.Errorf("create image temp dir: %w", err)
	}

	tempPath := filepath.Join(s.uploadDir, uuid.New().String()+".upload")
	src, err := fileHeader.Open()
	if err != nil {
		return SaveImageResult{}, fmt.Errorf("open uploaded image: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(tempPath)
	if err != nil {
		return SaveImageResult{}, fmt.Errorf("create image temp file: %w", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		_ = os.Remove(tempPath)
		return SaveImageResult{}, fmt.Errorf("write image temp file: %w", err)
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(tempPath)
		return SaveImageResult{}, fmt.Errorf("close image temp file: %w", err)
	}

	result, err := s.saveFromLocalPath(ctx, in, tempPath, fileHeader.Filename, maxSize)
	if err != nil {
		_ = os.Remove(tempPath)
		return SaveImageResult{}, err
	}
	return result, nil
}

func (s *ImageService) saveFromLocalPath(ctx context.Context, in SaveImageInput, localPath, filename string, maxSize int64) (SaveImageResult, error) {
	info, err := os.Stat(localPath)
	if err != nil {
		return SaveImageResult{}, fmt.Errorf("stat image local file: %w", err)
	}
	if maxSize > 0 && info.Size() > maxSize {
		return SaveImageResult{}, ErrImageTooLarge
	}

	detected, err := mimetype.DetectFile(localPath)
	if err != nil {
		return SaveImageResult{}, fmt.Errorf("detect image mime: %w", err)
	}
	originalMIME := strings.ToLower(strings.TrimSpace(detected.String()))
	if !isAllowedImageMIME(originalMIME) {
		return SaveImageResult{}, ErrInvalidImageType
	}

	hash, err := hashutil.SHA256(localPath)
	if err != nil {
		return SaveImageResult{}, fmt.Errorf("calculate image hash: %w", err)
	}
	if existingID, exists, err := s.repo.FindImageByHash(ctx, hash, info.Size()); err != nil {
		return SaveImageResult{}, err
	} else if exists {
		existing, getErr := s.repo.GetImageByID(ctx, existingID)
		if getErr != nil {
			return SaveImageResult{}, getErr
		}
		_ = os.Remove(localPath)
		return SaveImageResult{
			ImageID:       existingID,
			AlreadyExists: true,
			Status:        existing.Status,
			StoredPath:    existing.StoredPath,
			StoredMIME:    existing.StoredMIME,
		}, nil
	}

	imageID := uuid.New()
	origDir := filepath.Join(s.storage, "images", "originals")
	storedDir := filepath.Join(s.storage, "images", "processed")
	if err := os.MkdirAll(origDir, 0o755); err != nil {
		return SaveImageResult{}, fmt.Errorf("create image original dir: %w", err)
	}
	if err := os.MkdirAll(storedDir, 0o755); err != nil {
		return SaveImageResult{}, fmt.Errorf("create image processed dir: %w", err)
	}

	origExt := normalizeImageExt(detected.Extension())
	if origExt == "" {
		origExt = strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	}
	if origExt == "" {
		origExt = ".bin"
	}
	originalPath := filepath.Join(origDir, imageID.String()+origExt)
	if err := os.Rename(localPath, originalPath); err != nil {
		if copyErr := ffmpeg.CopyFile(originalPath, localPath); copyErr != nil {
			return SaveImageResult{}, fmt.Errorf("move image file failed: %w", err)
		}
		_ = os.Remove(localPath)
	}

	storedPath := originalPath
	storedMIME := originalMIME
	storedExt := origExt
	convertedToWebP := false
	sourceDeleted := false

	if !strings.HasSuffix(strings.ToLower(originalMIME), "gif") {
		storedPath = filepath.Join(storedDir, imageID.String()+".webp")
		if err := ffmpeg.ConvertToWebP(ctx, originalPath, storedPath, 82); err != nil {
			_ = os.Remove(storedPath)
			return SaveImageResult{}, fmt.Errorf("convert image to webp: %w", err)
		}
		storedMIME = "image/webp"
		storedExt = ".webp"
		convertedToWebP = true
		if err := os.Remove(originalPath); err != nil && !os.IsNotExist(err) {
			_ = os.Remove(storedPath)
			return SaveImageResult{}, fmt.Errorf("remove source image after webp conversion: %w", err)
		}
		sourceDeleted = true
		originalPath = ""
	}

	probe, err := ffmpeg.Probe(ctx, storedPath)
	if err != nil {
		probe.Width = 0
		probe.Height = 0
	}
	storedInfo, err := os.Stat(storedPath)
	if err != nil {
		return SaveImageResult{}, fmt.Errorf("stat stored image file: %w", err)
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = strings.TrimSuffix(strings.TrimSpace(filename), filepath.Ext(strings.TrimSpace(filename)))
		if title == "" {
			title = "untitled"
		}
	}
	meta, err := json.Marshal(map[string]any{
		"original_filename": filename,
		"source_hash":       hash,
		"source_size":       info.Size(),
		"stored_size":       storedInfo.Size(),
		"converted_to_webp": convertedToWebP,
		"source_deleted":    sourceDeleted,
	})
	if err != nil {
		return SaveImageResult{}, fmt.Errorf("marshal image metadata: %w", err)
	}

	imageRow := models.Image{
		ID:           imageID,
		UserID:       &in.UserID,
		Title:        title,
		Description:  strings.TrimSpace(in.Description),
		Status:       "ready",
		Active:       true,
		OriginalPath: originalPath,
		StoredPath:   storedPath,
		OriginalMIME: originalMIME,
		StoredMIME:   storedMIME,
		OriginalExt:  origExt,
		StoredExt:    storedExt,
		FileSize:     storedInfo.Size(),
		Width:        probe.Width,
		Height:       probe.Height,
		Metadata:     meta,
	}
	if err := s.repo.CreateImage(ctx, imageRow); err != nil {
		return SaveImageResult{}, err
	}
	cleanupCreated := func() {
		_ = s.repo.DeleteImageByIDCascade(ctx, imageID)
		_ = os.Remove(storedPath)
		if originalPath != "" {
			_ = os.Remove(originalPath)
		}
	}

	if err := s.repo.ReplaceImageActorsByInput(ctx, imageID, in.ActorIDs, in.ActorNames, "upload_manual"); err != nil {
		cleanupCreated()
		return SaveImageResult{}, err
	}
	if err := s.repo.ReplaceImageCollectionsByIDs(ctx, imageID, in.CollectionIDs); err != nil {
		cleanupCreated()
		return SaveImageResult{}, err
	}
	if err := s.repo.InsertImageHash(ctx, hash, imageID, info.Size()); err != nil {
		if repository.IsUniqueViolation(err) {
			cleanupCreated()
			if existingID, exists, e := s.repo.FindImageByHash(ctx, hash, info.Size()); e == nil && exists {
				existing, getErr := s.repo.GetImageByID(ctx, existingID)
				if getErr != nil {
					return SaveImageResult{}, getErr
				}
				return SaveImageResult{ImageID: existingID, AlreadyExists: true, Status: existing.Status, StoredPath: existing.StoredPath, StoredMIME: existing.StoredMIME}, nil
			}
			return SaveImageResult{}, err
		}
		cleanupCreated()
		return SaveImageResult{}, err
	}

	s.logger.Info("image upload saved", "image_id", imageID, "stored_path", storedPath, "stored_mime", storedMIME)
	return SaveImageResult{ImageID: imageID, AlreadyExists: false, Status: "ready", StoredPath: storedPath, StoredMIME: storedMIME}, nil
}

func (s *ImageService) ResolveViewPath(ctx context.Context, imageID uuid.UUID, width, height int, fit string, quality int) (string, string, error) {
	img, err := s.repo.GetImageByID(ctx, imageID)
	if err != nil {
		return "", "", err
	}
	return s.resolveViewPathFromImage(ctx, img, width, height, fit, quality)
}

func (s *ImageService) ResolveAppViewPath(ctx context.Context, imageID uuid.UUID, width, height int, fit string, quality int) (string, string, error) {
	img, err := s.repo.GetImageByID(ctx, imageID)
	if err != nil {
		return "", "", err
	}
	if !img.Active {
		return "", "", fmt.Errorf("image not active")
	}
	return s.resolveViewPathFromImage(ctx, img, width, height, fit, quality)
}

func (s *ImageService) resolveViewPathFromImage(ctx context.Context, img models.Image, width, height int, fit string, quality int) (string, string, error) {
	if img.Status != "ready" {
		return "", "", fmt.Errorf("image not ready")
	}
	basePath := strings.TrimSpace(img.StoredPath)
	if basePath == "" {
		return "", "", fmt.Errorf("image source not found")
	}
	if width <= 0 && height <= 0 {
		return basePath, img.StoredMIME, nil
	}
	if quality <= 0 || quality > 100 {
		quality = 82
	}
	fit = normalizeFitMode(fit)

	format := "webp"
	mime := "image/webp"
	if strings.HasSuffix(strings.ToLower(img.StoredMIME), "gif") || strings.HasSuffix(strings.ToLower(img.StoredExt), "gif") {
		format = "gif"
		mime = "image/gif"
	}

	key := fmt.Sprintf("w%d_h%d_fit%s_q%d_%s", width, height, fit, quality, format)
	if p, m, _, _, exists, err := s.repo.GetImageVariantByKey(ctx, img.ID, key); err == nil && exists {
		if _, statErr := os.Stat(p); statErr == nil {
			if strings.TrimSpace(m) != "" {
				mime = m
			}
			return p, mime, nil
		}
	}

	variantDir := filepath.Join(s.storage, "images", "variants", img.ID.String())
	if err := os.MkdirAll(variantDir, 0o755); err != nil {
		return "", "", fmt.Errorf("create image variant dir: %w", err)
	}
	variantPath := filepath.Join(variantDir, key+"."+format)
	if err := ffmpeg.ResizeImage(ctx, basePath, variantPath, width, height, fit, format, quality); err != nil {
		return "", "", err
	}
	probe, probeErr := ffmpeg.Probe(ctx, variantPath)
	w, h := width, height
	if probeErr == nil {
		if probe.Width > 0 {
			w = probe.Width
		}
		if probe.Height > 0 {
			h = probe.Height
		}
	}
	if err := s.repo.UpsertImageVariant(ctx, img.ID, key, variantPath, mime, w, h); err != nil {
		return "", "", err
	}
	return variantPath, mime, nil
}

func isAllowedImageMIME(mime string) bool {
	mime = strings.ToLower(strings.TrimSpace(mime))
	switch mime {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/bmp", "image/tiff", "image/x-tiff":
		return true
	default:
		return false
	}
}

func normalizeImageExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	switch ext {
	case ".jpeg", ".jpg":
		return ".jpg"
	case ".png":
		return ".png"
	case ".gif":
		return ".gif"
	case ".webp":
		return ".webp"
	case ".bmp":
		return ".bmp"
	case ".tif", ".tiff":
		return ".tiff"
	default:
		return ext
	}
}

func normalizeFitMode(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "cover", "contain", "inside":
		return v
	default:
		return "inside"
	}
}
