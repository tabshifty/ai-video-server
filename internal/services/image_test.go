package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/pkg/ffmpeg"
)

type fakeImageRepository struct {
	createdImage models.Image
}

func (r *fakeImageRepository) FindImageByHash(context.Context, string, int64) (uuid.UUID, bool, error) {
	return uuid.Nil, false, nil
}

func (r *fakeImageRepository) GetImageByID(context.Context, uuid.UUID) (models.Image, error) {
	return models.Image{}, errors.New("unexpected GetImageByID")
}

func (r *fakeImageRepository) CreateImage(_ context.Context, img models.Image) error {
	r.createdImage = img
	return nil
}

func (r *fakeImageRepository) DeleteImageByIDCascade(context.Context, uuid.UUID) error {
	return nil
}

func (r *fakeImageRepository) ReplaceImageActorsByInput(context.Context, uuid.UUID, []uuid.UUID, []string, string) error {
	return nil
}

func (r *fakeImageRepository) ReplaceImageCollectionsByIDs(context.Context, uuid.UUID, []uuid.UUID) error {
	return nil
}

func (r *fakeImageRepository) InsertImageHash(context.Context, string, uuid.UUID, int64) error {
	return nil
}

func (r *fakeImageRepository) GetImageVariantByKey(context.Context, uuid.UUID, string) (string, string, int, int, bool, error) {
	return "", "", 0, 0, false, errors.New("unexpected GetImageVariantByKey")
}

func (r *fakeImageRepository) UpsertImageVariant(context.Context, uuid.UUID, string, string, string, int, int) error {
	return errors.New("unexpected UpsertImageVariant")
}

func TestSaveFromLocalPathKeepsOriginalWhenWebPEncodingUnavailable(t *testing.T) {
	originalConvert := convertImageToWebP
	convertImageToWebP = func(context.Context, string, string, int) error {
		return ffmpeg.ErrWebPEncodingUnavailable
	}
	t.Cleanup(func() {
		convertImageToWebP = originalConvert
	})

	root := t.TempDir()
	inputPath := filepath.Join(root, "upload.jpg")
	if err := os.WriteFile(inputPath, tinyJPEGBytes(), 0o644); err != nil {
		t.Fatalf("write jpeg fixture: %v", err)
	}

	repo := &fakeImageRepository{}
	svc := &ImageService{
		repo:      repo,
		uploadDir: filepath.Join(root, "tmp"),
		storage:   filepath.Join(root, "storage"),
		logger:    slog.Default(),
	}

	result, err := svc.saveFromLocalPath(context.Background(), SaveImageInput{UserID: uuid.New()}, inputPath, "cover.jpg", 0)
	if err != nil {
		t.Fatalf("saveFromLocalPath() error = %v", err)
	}

	if result.StoredMIME != "image/jpeg" {
		t.Fatalf("StoredMIME = %s, want image/jpeg", result.StoredMIME)
	}
	if filepath.Ext(result.StoredPath) != ".jpg" {
		t.Fatalf("StoredPath = %s, want .jpg file", result.StoredPath)
	}
	if _, err := os.Stat(result.StoredPath); err != nil {
		t.Fatalf("stored original not available: %v", err)
	}
	if repo.createdImage.OriginalPath != result.StoredPath {
		t.Fatalf("OriginalPath = %s, want stored path %s", repo.createdImage.OriginalPath, result.StoredPath)
	}
	if repo.createdImage.StoredExt != ".jpg" {
		t.Fatalf("StoredExt = %s, want .jpg", repo.createdImage.StoredExt)
	}
	var meta map[string]any
	if err := json.Unmarshal(repo.createdImage.Metadata, &meta); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	if meta["converted_to_webp"] != false {
		t.Fatalf("converted_to_webp = %v, want false", meta["converted_to_webp"])
	}
	if meta["source_deleted"] != false {
		t.Fatalf("source_deleted = %v, want false", meta["source_deleted"])
	}
}

func TestImageVariantFormatUsesStoredFormatWhenOriginalWasKept(t *testing.T) {
	tests := []struct {
		name       string
		storedMIME string
		storedExt  string
		wantFormat string
		wantMIME   string
	}{
		{
			name:       "jpeg",
			storedMIME: "image/jpeg",
			storedExt:  ".jpg",
			wantFormat: "jpg",
			wantMIME:   "image/jpeg",
		},
		{
			name:       "png",
			storedMIME: "image/png",
			storedExt:  ".png",
			wantFormat: "png",
			wantMIME:   "image/png",
		},
		{
			name:       "webp",
			storedMIME: "image/webp",
			storedExt:  ".webp",
			wantFormat: "webp",
			wantMIME:   "image/webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFormat, gotMIME := imageVariantFormat(tt.storedMIME, tt.storedExt)
			if gotFormat != tt.wantFormat || gotMIME != tt.wantMIME {
				t.Fatalf("imageVariantFormat()=(%s,%s), want (%s,%s)", gotFormat, gotMIME, tt.wantFormat, tt.wantMIME)
			}
		})
	}
}

func tinyJPEGBytes() []byte {
	return []byte{
		0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46,
		0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x48,
		0x00, 0x48, 0x00, 0x00, 0xff, 0xdb, 0x00, 0x43,
		0x00, 0x03, 0x02, 0x02, 0x03, 0x02, 0x02, 0x03,
		0x03, 0x03, 0x03, 0x04, 0x03, 0x03, 0x04, 0x05,
		0x08, 0x05, 0x05, 0x04, 0x04, 0x05, 0x0a, 0x07,
		0x07, 0x06, 0x08, 0x0c, 0x0a, 0x0c, 0x0c, 0x0b,
		0x0a, 0x0b, 0x0b, 0x0d, 0x0e, 0x12, 0x10, 0x0d,
		0x0e, 0x11, 0x0e, 0x0b, 0x0b, 0x10, 0x16, 0x10,
		0x11, 0x13, 0x14, 0x15, 0x15, 0x15, 0x0c, 0x0f,
		0x17, 0x18, 0x16, 0x14, 0x18, 0x12, 0x14, 0x15,
		0x14, 0xff, 0xc0, 0x00, 0x0b, 0x08, 0x00, 0x01,
		0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xff, 0xc4,
		0x00, 0x14, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x08, 0xff, 0xc4, 0x00, 0x14,
		0x10, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xff, 0xda, 0x00, 0x08, 0x01, 0x01,
		0x00, 0x00, 0x3f, 0x00, 0x2a, 0xff, 0xd9,
	}
}
