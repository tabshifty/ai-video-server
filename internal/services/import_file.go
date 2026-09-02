package services

import "context"

// SaveImportedFile stores a local file without replacing an existing video's original path on duplicate.
func (s *UploadService) SaveImportedFile(ctx context.Context, in LocalUploadInput, maxVideoSize int64) (UploadResult, error) {
	return s.saveLocalFile(ctx, in, maxVideoSize, false)
}
