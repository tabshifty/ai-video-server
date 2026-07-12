package services

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
)

func TestDeriveArchiveFilenameTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "removes ad text", path: "目录/www.98T.la@ABC-123.mp4", want: "ABC-123"},
		{name: "matches case insensitively and globally", path: "WWW.98t.LA@ A www.98T.la@ B.mkv", want: "A B"},
		{name: "keeps punctuation", path: "目录/www.98T.la@-ABC_[01].mp4", want: "-ABC_[01]"},
		{name: "keeps near match", path: "目录/98T.la@ABC.mp4", want: "98T.la@ABC"},
		{name: "collapses whitespace", path: "目录/www.98T.la@  A  B .mp4", want: "A B"},
		{name: "returns empty after cleaning", path: "目录/www.98T.la@.mp4", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deriveArchiveFilenameTitle(tt.path); got != tt.want {
				t.Fatalf("deriveArchiveFilenameTitle(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestMergeArchiveFilenameTitleDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		oldTitle    string
		newTitle    string
		description string
		want        string
	}{
		{name: "prepends old title", oldTitle: "旧标题", newTitle: "新标题", description: "原说明", want: "旧标题\n原说明"},
		{name: "returns old title without description", oldTitle: "旧标题", newTitle: "新标题", description: "", want: "旧标题"},
		{name: "keeps description without old title", oldTitle: "", newTitle: "新标题", description: "原说明", want: "原说明"},
		{name: "keeps description for unchanged title", oldTitle: "同名", newTitle: "同名", description: "原说明", want: "原说明"},
		{name: "compares titles case sensitively", oldTitle: "Title", newTitle: "title", description: "原说明", want: "Title\n原说明"},
		{name: "trims titles before comparison", oldTitle: "  同名  ", newTitle: " 同名 ", description: "  原说明\n", want: "  原说明\n"},
		{name: "preserves description exactly", oldTitle: "  旧标题  ", newTitle: "新标题", description: "  原说明\n\n", want: "旧标题\n  原说明\n\n"},
		{name: "preserves whitespace-only description", oldTitle: "旧标题", newTitle: "新标题", description: "  ", want: "旧标题\n  "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeArchiveFilenameTitleDescription(tt.oldTitle, tt.newTitle, tt.description); got != tt.want {
				t.Fatalf("mergeArchiveFilenameTitleDescription(%q, %q, %q) = %q, want %q", tt.oldTitle, tt.newTitle, tt.description, got, tt.want)
			}
		})
	}
}

func TestCanReplaceArchiveFilenameTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		file models.ArchiveImportFileListItem
		want bool
	}{
		{name: "pending video", file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "video", Status: "pending"}, want: true},
		{name: "failed video", file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "video", Status: "failed"}, want: true},
		{name: "processing video", file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "video", Status: "processing"}, want: false},
		{name: "ready video", file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "video", Status: "ready"}, want: false},
		{name: "existing video", file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "video", Status: "existing"}, want: false},
		{name: "skipped video", file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "video", Status: "skipped"}, want: false},
		{name: "pending image", file: models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "image", Status: "pending"}, want: false},
		{name: "pending directory", file: models.ArchiveImportFileListItem{EntryType: "directory", MediaKind: "video", Status: "pending"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canReplaceArchiveFilenameTitle(tt.file); got != tt.want {
				t.Fatalf("canReplaceArchiveFilenameTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArchiveImportBatchUpdateConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "filename title mode", got: ArchiveImportTitleModeFilename, want: "filename"},
		{name: "invalid selection", got: ArchiveImportBatchReasonInvalidSelection, want: "invalid_selection"},
		{name: "ineligible target", got: ArchiveImportBatchReasonIneligibleTarget, want: "ineligible_target"},
		{name: "stale target", got: ArchiveImportBatchReasonStaleTarget, want: "stale_target"},
		{name: "empty title", got: ArchiveImportBatchReasonEmptyTitle, want: "empty_derived_title"},
		{name: "invalid patch", got: ArchiveImportBatchReasonInvalidPatch, want: "invalid_patch"},
		{name: "update failed", got: ArchiveImportBatchReasonUpdateFailed, want: "update_failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("constant = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestArchiveImportBatchUpdateContracts(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	updatedAt := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	target := ArchiveImportBatchUpdateTarget{ID: id, UpdatedAt: updatedAt}
	input := ArchiveImportBatchUpdateInput{
		Targets:                  []ArchiveImportBatchUpdateTarget{target},
		TitleMode:                ArchiveImportTitleModeFilename,
		UpdateTags:               true,
		Tags:                     []string{"标签"},
		UpdateVideoType:          true,
		VideoType:                "movie",
		UpdateVideoCollectionIDs: true,
		VideoCollectionIDs:       []uuid.UUID{id},
		UpdateImageCollectionIDs: true,
		ImageCollectionIDs:       []uuid.UUID{id},
	}
	issue := ArchiveImportBatchUpdateIssue{ID: id, RelativePath: "目录/视频.mp4", Message: "失败"}

	if input.Targets[0] != target || input.VideoCollectionIDs[0] != id || input.ImageCollectionIDs[0] != id {
		t.Fatalf("batch update input did not preserve target and collection IDs")
	}
	if !input.UpdateTags || !input.UpdateVideoType || !input.UpdateVideoCollectionIDs || !input.UpdateImageCollectionIDs {
		t.Fatalf("batch update input did not preserve update flags")
	}
	if input.TitleMode != "filename" || input.Tags[0] != "标签" || input.VideoType != "movie" {
		t.Fatalf("batch update input did not preserve patch values")
	}
	if issue.ID != id || issue.RelativePath != "目录/视频.mp4" || issue.Message != "失败" {
		t.Fatalf("batch update issue did not preserve fields")
	}
}

func TestArchiveImportBatchUpdateError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("更新失败")
	tests := []struct {
		name       string
		err        *ArchiveImportBatchUpdateError
		want       string
		wantUnwrap bool
	}{
		{
			name:       "uses wrapped error message",
			err:        &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonUpdateFailed, Err: sentinel},
			want:       "更新失败",
			wantUnwrap: true,
		},
		{
			name: "uses stable reason without wrapped error",
			err:  &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonInvalidSelection},
			want: "invalid_selection",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Fatalf("ArchiveImportBatchUpdateError.Error() = %q, want %q", got, tt.want)
			}
			if got := errors.Is(tt.err, sentinel); got != tt.wantUnwrap {
				t.Fatalf("errors.Is() = %v, want %v", got, tt.wantUnwrap)
			}
		})
	}
}
