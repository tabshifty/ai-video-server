package services

import (
	"testing"

	"video-server/internal/models"
)

func TestArchiveImportShouldProcessInBatch(t *testing.T) {
	tests := []struct {
		name string
		file models.ArchiveImportFileListItem
		want bool
	}{
		{
			name: "pending video",
			file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "video", Status: "pending"},
			want: true,
		},
		{
			name: "failed image retry",
			file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "image", Status: "failed"},
			want: true,
		},
		{
			name: "processing video retry",
			file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "video", Status: "processing"},
			want: true,
		},
		{
			name: "ready video",
			file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "video", Status: "ready"},
			want: false,
		},
		{
			name: "existing image",
			file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "image", Status: "existing"},
			want: false,
		},
		{
			name: "skipped archive",
			file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "archive", Status: "skipped"},
			want: false,
		},
		{
			name: "directory",
			file: models.ArchiveImportFileListItem{EntryType: "directory", MediaKind: "directory", Status: "skipped"},
			want: false,
		},
		{
			name: "other pending file",
			file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "other", Status: "pending"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldProcessArchiveFileInBatch(tt.file); got != tt.want {
				t.Fatalf("shouldProcessArchiveFileInBatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
