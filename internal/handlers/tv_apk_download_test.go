package handlers

import (
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/models"
)

func TestTVAppDownloadAPKAcceptsPhoneSingleSlot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempFile, err := os.CreateTemp(t.TempDir(), "phone-app-*.apk")
	if err != nil {
		t.Fatalf("create temp apk: %v", err)
	}
	content := []byte("fake phone apk")
	if _, err := tempFile.Write(content); err != nil {
		t.Fatalf("write temp apk: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("close temp apk: %v", err)
	}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Params = gin.Params{
		{Key: "id", Value: "12"},
		{Key: "abi", Value: models.AppAPKSlotSingle},
	}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/app/releases/12/download/single", nil)

	api := &API{
		tvAPKSvc: &fakeTVAPKService{
			findReleaseAPK: models.TVAppReleaseABIInfo{
				ABI:        models.AppAPKSlotSingle,
				FileName:   "app-release.apk",
				StoredPath: tempFile.Name(),
				FileSize:   int64(len(content)),
				MIMEType:   "application/vnd.android.package-archive",
			},
		},
	}

	api.TVAppDownloadAPK(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET phone app download = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/vnd.android.package-archive" {
		t.Fatalf("Content-Type = %q, want apk mime type", got)
	}
	if got := rec.Body.String(); got != string(content) {
		t.Fatalf("response body = %q, want %q", got, string(content))
	}
}

type fakeTVAPKService struct {
	findReleaseAPK    models.TVAppReleaseABIInfo
	findReleaseAPKErr error
}

func (s *fakeTVAPKService) AdminList(context.Context, models.AdminTvAppReleaseFilter) ([]models.AdminTvAppReleaseListItem, int, error) {
	return nil, 0, nil
}

func (s *fakeTVAPKService) AdminDetail(context.Context, int64) (models.AdminTvAppReleaseDetail, error) {
	return models.AdminTvAppReleaseDetail{}, nil
}

func (s *fakeTVAPKService) UploadAPK(context.Context, *multipart.FileHeader, string, *uuid.UUID, string, bool) (models.TVAppReleaseRecord, models.TVAppReleaseABIInfo, error) {
	return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, nil
}

func (s *fakeTVAPKService) UpdateRelease(context.Context, int64, models.AdminTvAppReleaseUpdateInput) (models.TVAppReleaseRecord, error) {
	return models.TVAppReleaseRecord{}, nil
}

func (s *fakeTVAPKService) Publish(context.Context, int64, models.AdminTvAppReleasePublishInput) (models.TVAppReleaseRecord, error) {
	return models.TVAppReleaseRecord{}, nil
}

func (s *fakeTVAPKService) Offline(context.Context, int64) (models.TVAppReleaseRecord, error) {
	return models.TVAppReleaseRecord{}, nil
}

func (s *fakeTVAPKService) Restore(context.Context, int64) (models.TVAppReleaseRecord, error) {
	return models.TVAppReleaseRecord{}, nil
}

func (s *fakeTVAPKService) DeleteDraft(context.Context, int64) error {
	return nil
}

func (s *fakeTVAPKService) FamilyReleases(context.Context, string) ([]models.TVAppFamilyRelease, error) {
	return nil, nil
}

func (s *fakeTVAPKService) FindReleaseAPK(context.Context, int64, string) (models.TVAppReleaseABIInfo, error) {
	return s.findReleaseAPK, s.findReleaseAPKErr
}

func TestTVAppDownloadAPKRejectsUnknownArtifactSlot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Params = gin.Params{
		{Key: "id", Value: "12"},
		{Key: "abi", Value: "x86"},
	}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/app/releases/12/download/x86", nil)

	api := &API{tvAPKSvc: &fakeTVAPKService{}}
	api.TVAppDownloadAPK(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET invalid artifact slot = %d, want 200 response envelope", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"msg":"invalid abi"`) {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
}

var _ tvAPKService = (*fakeTVAPKService)(nil)
