package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
)

type tvAPKRepository interface {
	AdminListTVAppReleases(context.Context, models.AdminTvAppReleaseFilter) ([]models.AdminTvAppReleaseListItem, int, error)
	GetAdminTVAppReleaseDetail(context.Context, int64) (models.AdminTvAppReleaseDetail, error)
	CreateOrUpdateTVAppDraftReleaseWithABI(context.Context, models.TVAppAPKParsedMetadata, string, *uuid.UUID, string, bool) (models.TVAppReleaseRecord, models.TVAppReleaseABIInfo, error)
	UpdateTVAppReleaseNotes(context.Context, int64, string, string) (models.TVAppReleaseRecord, error)
	PublishTVAppRelease(context.Context, int64, string, string) (models.TVAppReleaseRecord, error)
	OfflineTVAppRelease(context.Context, int64) (models.TVAppReleaseRecord, error)
	RestoreTVAppRelease(context.Context, int64) (models.TVAppReleaseRecord, error)
	DeleteTVAppDraftRelease(context.Context, int64) error
	ListTVAppFamilyReleases(context.Context, string, int) ([]models.TVAppFamilyRelease, error)
	GetTVAppReleaseAPKByABI(context.Context, int64, string) (models.TVAppReleaseABIInfo, error)
}

type TVAPKService struct {
	repo      tvAPKRepository
	uploadDir string
	storage   string
}

func NewTVAPKService(repo *repository.VideoRepository, uploadDir, storageRoot string) *TVAPKService {
	return &TVAPKService{
		repo:      repo,
		uploadDir: uploadDir,
		storage:   storageRoot,
	}
}

func (s *TVAPKService) AdminList(ctx context.Context, filter models.AdminTvAppReleaseFilter) ([]models.AdminTvAppReleaseListItem, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	return s.repo.AdminListTVAppReleases(ctx, filter)
}

func (s *TVAPKService) AdminDetail(ctx context.Context, releaseID int64) (models.AdminTvAppReleaseDetail, error) {
	return s.repo.GetAdminTVAppReleaseDetail(ctx, releaseID)
}

func (s *TVAPKService) AdminDetailForRecord(ctx context.Context, release models.TVAppReleaseRecord) (models.AdminTvAppReleaseDetail, error) {
	return s.repo.GetAdminTVAppReleaseDetail(ctx, release.ID)
}

func (s *TVAPKService) UploadAPK(
	ctx context.Context,
	fileHeader *multipart.FileHeader,
	clientType string,
	userID *uuid.UUID,
	username string,
	replaceExisting bool,
) (models.TVAppReleaseRecord, models.TVAppReleaseABIInfo, error) {
	if fileHeader == nil {
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, fmt.Errorf("apk file is required")
	}
	if strings.TrimSpace(fileHeader.Filename) == "" {
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, fmt.Errorf("apk file name is required")
	}
	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, fmt.Errorf("create tv apk upload dir: %w", err)
	}

	tempPath := filepath.Join(s.uploadDir, uuid.New().String()+".apk")
	src, err := fileHeader.Open()
	if err != nil {
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, fmt.Errorf("open uploaded tv apk: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(tempPath)
	if err != nil {
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, fmt.Errorf("create tv apk temp file: %w", err)
	}
	if _, err := dst.ReadFrom(src); err != nil {
		dst.Close()
		_ = os.Remove(tempPath)
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, fmt.Errorf("write tv apk temp file: %w", err)
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(tempPath)
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, fmt.Errorf("close tv apk temp file: %w", err)
	}

	meta, err := ParseTVAPKMetadata(tempPath, fileHeader.Filename)
	if err != nil {
		_ = os.Remove(tempPath)
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, err
	}
	if normalized := models.NormalizeAppClientType(clientType); normalized != "" && normalized != meta.ClientType {
		_ = os.Remove(tempPath)
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, models.NewTVAPKDomainError(models.TVAPKErrorClientTypeMismatch, "所选客户端类型与上传 APK 不匹配")
	}

	targetDir := filepath.Join(s.storage, "app-apks", models.AppClientTypeSlug(meta.ClientType), fmt.Sprintf("%d", meta.VersionCode))
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		_ = os.Remove(tempPath)
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, fmt.Errorf("create tv apk target dir: %w", err)
	}
	targetPath := filepath.Join(targetDir, buildStoredTVAPKFileName(meta))
	if err := moveTVAPKFile(tempPath, targetPath); err != nil {
		_ = os.Remove(tempPath)
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, fmt.Errorf("move tv apk to storage: %w", err)
	}

	release, abiInfo, err := s.repo.CreateOrUpdateTVAppDraftReleaseWithABI(ctx, meta, targetPath, userID, username, replaceExisting)
	if err != nil {
		_ = os.Remove(targetPath)
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, err
	}
	return release, abiInfo, nil
}

func (s *TVAPKService) UpdateRelease(ctx context.Context, releaseID int64, input models.AdminTvAppReleaseUpdateInput) (models.TVAppReleaseRecord, error) {
	return s.repo.UpdateTVAppReleaseNotes(ctx, releaseID, input.ReleaseNotes, input.Remarks)
}

func (s *TVAPKService) Publish(ctx context.Context, releaseID int64, input models.AdminTvAppReleasePublishInput) (models.TVAppReleaseRecord, error) {
	return s.repo.PublishTVAppRelease(ctx, releaseID, input.ReleaseNotes, input.Remarks)
}

func (s *TVAPKService) Offline(ctx context.Context, releaseID int64) (models.TVAppReleaseRecord, error) {
	return s.repo.OfflineTVAppRelease(ctx, releaseID)
}

func (s *TVAPKService) Restore(ctx context.Context, releaseID int64) (models.TVAppReleaseRecord, error) {
	return s.repo.RestoreTVAppRelease(ctx, releaseID)
}

func (s *TVAPKService) DeleteDraft(ctx context.Context, releaseID int64) error {
	return s.repo.DeleteTVAppDraftRelease(ctx, releaseID)
}

func (s *TVAPKService) FamilyReleases(ctx context.Context, clientType string) ([]models.TVAppFamilyRelease, error) {
	return s.repo.ListTVAppFamilyReleases(ctx, clientType, 3)
}

func (s *TVAPKService) FindReleaseAPK(ctx context.Context, releaseID int64, abi string) (models.TVAppReleaseABIInfo, error) {
	if normalized := models.TVNormalizeABI(abi); normalized != "" {
		return s.repo.GetTVAppReleaseAPKByABI(ctx, releaseID, normalized)
	}
	return s.repo.GetTVAppReleaseAPKByABI(ctx, releaseID, models.NormalizeReleaseArtifactSlot(models.AppClientTypeAndroidPhone, abi))
}

func buildStoredTVAPKFileName(meta models.TVAppAPKParsedMetadata) string {
	versionName := strings.ReplaceAll(strings.TrimSpace(meta.VersionName), "/", "-")
	versionName = strings.ReplaceAll(versionName, " ", "-")
	if versionName == "" {
		versionName = "unknown"
	}
	prefix := models.AppClientTypeSlug(meta.ClientType)
	if prefix == "" {
		prefix = "android-app"
	}
	return fmt.Sprintf("%s-v%d-%s-%s.apk", prefix, meta.VersionCode, versionName, meta.ABI)
}

func moveTVAPKFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := out.ReadFrom(in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Remove(src)
}
