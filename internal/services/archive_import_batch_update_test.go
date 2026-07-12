package services

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

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

func TestPlanArchiveFilenameBatchUpdateAppliesPerFileTitlesAndEnabledFields(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 12, 5, 0, 0, 0, time.UTC)
	batchID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	videoCollectionID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	imageCollectionID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	first := models.ArchiveImportFileListItem{
		ID:           uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
		BatchID:      batchID,
		EntryType:    "file",
		MediaKind:    "video",
		Status:       "pending",
		RelativePath: "第一组/www.98T.la@第一集.mp4",
		Title:        " 旧标题 ",
		Description:  "  原说明\n",
		UpdatedAt:    now,
	}
	second := models.ArchiveImportFileListItem{
		ID:           uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"),
		BatchID:      batchID,
		EntryType:    "file",
		MediaKind:    "video",
		Status:       "failed",
		RelativePath: "第二组/第二集.mkv",
		Title:        " 第二集 ",
		Description:  "原说明保持不变 ",
		UpdatedAt:    now.Add(time.Second),
	}
	in := ArchiveImportBatchUpdateInput{
		Targets: []ArchiveImportBatchUpdateTarget{
			{ID: first.ID, UpdatedAt: first.UpdatedAt},
			{ID: second.ID, UpdatedAt: second.UpdatedAt},
		},
		TitleMode:                ArchiveImportTitleModeFilename,
		UpdateTags:               true,
		Tags:                     []string{" 标签 A ", "标签  A", "B"},
		UpdateVideoType:          true,
		VideoType:                " MOVIE ",
		UpdateVideoCollectionIDs: true,
		VideoCollectionIDs:       []uuid.UUID{videoCollectionID, videoCollectionID, uuid.Nil},
		UpdateImageCollectionIDs: true,
		ImageCollectionIDs:       []uuid.UUID{imageCollectionID, imageCollectionID, uuid.Nil},
	}

	plans, err := planArchiveFilenameBatchUpdate([]models.ArchiveImportFileListItem{second, first}, in)
	if err != nil {
		t.Fatalf("planArchiveFilenameBatchUpdate() error = %v", err)
	}
	if len(plans) != 2 {
		t.Fatalf("plan count = %d, want 2", len(plans))
	}
	if got := plans[0].File; got.ID != first.ID || got.Title != "第一集" || got.Description != "旧标题\n  原说明\n" {
		t.Fatalf("first plan file = %#v", got)
	}
	if got := plans[1].File; got.ID != second.ID || got.Title != "第二集" || got.Description != second.Description {
		t.Fatalf("second plan file = %#v", got)
	}
	for _, plan := range plans {
		if !slices.Equal(plan.File.Tags, []string{"标签 a", "b"}) {
			t.Fatalf("tags = %#v, want normalized values", plan.File.Tags)
		}
		if plan.File.VideoType != "movie" {
			t.Fatalf("video type = %q, want movie", plan.File.VideoType)
		}
		if !slices.Equal(plan.File.VideoCollectionIDs, []uuid.UUID{videoCollectionID}) {
			t.Fatalf("video collection IDs = %#v", plan.File.VideoCollectionIDs)
		}
		if !slices.Equal(plan.File.ImageCollectionIDs, []uuid.UUID{imageCollectionID}) {
			t.Fatalf("image collection IDs = %#v", plan.File.ImageCollectionIDs)
		}
	}
	if first.Description != "  原说明\n" {
		t.Fatalf("input file was mutated: description = %q", first.Description)
	}
}

func TestPlanArchiveFilenameBatchUpdateKeepsDisabledFields(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 12, 5, 0, 0, 0, time.UTC)
	videoCollectionIDs := []uuid.UUID{uuid.New(), uuid.New()}
	imageCollectionIDs := []uuid.UUID{uuid.New()}
	file := models.ArchiveImportFileListItem{
		ID:                 uuid.New(),
		BatchID:            uuid.New(),
		EntryType:          "file",
		MediaKind:          "video",
		Status:             "pending",
		RelativePath:       "标题.mp4",
		Title:              "原标题",
		Tags:               []string{"原标签"},
		VideoType:          "episode",
		VideoCollectionIDs: videoCollectionIDs,
		ImageCollectionIDs: imageCollectionIDs,
		UpdatedAt:          now,
	}

	plans, err := planArchiveFilenameBatchUpdate(
		[]models.ArchiveImportFileListItem{file},
		ArchiveImportBatchUpdateInput{
			Targets:   []ArchiveImportBatchUpdateTarget{{ID: file.ID, UpdatedAt: now}},
			TitleMode: ArchiveImportTitleModeFilename,
		},
	)
	if err != nil {
		t.Fatalf("planArchiveFilenameBatchUpdate() error = %v", err)
	}
	got := plans[0].File
	if !slices.Equal(got.Tags, file.Tags) || got.VideoType != file.VideoType ||
		!slices.Equal(got.VideoCollectionIDs, videoCollectionIDs) || !slices.Equal(got.ImageCollectionIDs, imageCollectionIDs) {
		t.Fatalf("disabled fields changed: %#v", got)
	}
}

func TestPlanArchiveFilenameBatchUpdateRejectsInvalidSelection(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 12, 5, 0, 0, 0, time.UTC)
	id := uuid.New()
	tests := []struct {
		name  string
		files []models.ArchiveImportFileListItem
		in    ArchiveImportBatchUpdateInput
	}{
		{
			name: "empty targets",
			in:   ArchiveImportBatchUpdateInput{TitleMode: ArchiveImportTitleModeFilename},
		},
		{
			name: "nil target id",
			in: ArchiveImportBatchUpdateInput{
				Targets:   []ArchiveImportBatchUpdateTarget{{UpdatedAt: now}},
				TitleMode: ArchiveImportTitleModeFilename,
			},
		},
		{
			name: "zero target timestamp",
			in: ArchiveImportBatchUpdateInput{
				Targets:   []ArchiveImportBatchUpdateTarget{{ID: id}},
				TitleMode: ArchiveImportTitleModeFilename,
			},
		},
		{
			name: "duplicate target id",
			in: ArchiveImportBatchUpdateInput{
				Targets: []ArchiveImportBatchUpdateTarget{
					{ID: id, UpdatedAt: now},
					{ID: id, UpdatedAt: now},
				},
				TitleMode: ArchiveImportTitleModeFilename,
			},
		},
		{
			name: "unknown target",
			in: ArchiveImportBatchUpdateInput{
				Targets:   []ArchiveImportBatchUpdateTarget{{ID: id, UpdatedAt: now}},
				TitleMode: ArchiveImportTitleModeFilename,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plans, err := planArchiveFilenameBatchUpdate(tt.files, tt.in)
			if plans != nil {
				t.Fatalf("plans = %#v, want nil", plans)
			}
			requireArchiveImportBatchUpdateError(t, err, ArchiveImportBatchReasonInvalidSelection)
		})
	}
}

func TestPlanArchiveFilenameBatchUpdateRejectsInvalidTitleMode(t *testing.T) {
	t.Parallel()

	plans, err := planArchiveFilenameBatchUpdate(nil, ArchiveImportBatchUpdateInput{TitleMode: "manual"})
	if plans != nil {
		t.Fatalf("plans = %#v, want nil", plans)
	}
	requireArchiveImportBatchUpdateError(t, err, ArchiveImportBatchReasonInvalidPatch)
}

func TestPlanArchiveFilenameBatchUpdateRejectsCrossBatchSelection(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 12, 5, 0, 0, 0, time.UTC)
	first := eligibleArchiveBatchUpdateFile(uuid.New(), uuid.New(), "第一集.mp4", now)
	second := eligibleArchiveBatchUpdateFile(uuid.New(), uuid.New(), "第二集.mp4", now)
	in := archiveFilenameBatchUpdateInput(first, second)

	plans, err := planArchiveFilenameBatchUpdate([]models.ArchiveImportFileListItem{first, second}, in)
	if plans != nil {
		t.Fatalf("plans = %#v, want nil", plans)
	}
	batchErr := requireArchiveImportBatchUpdateError(t, err, ArchiveImportBatchReasonInvalidSelection)
	requireArchiveImportBatchUpdateIssues(t, batchErr, second)
}

func TestPlanArchiveFilenameBatchUpdateRejectsIneligibleTargets(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 12, 5, 0, 0, 0, time.UTC)
	batchID := uuid.New()
	files := []models.ArchiveImportFileListItem{
		eligibleArchiveBatchUpdateFile(uuid.New(), batchID, "图片.jpg", now),
		eligibleArchiveBatchUpdateFile(uuid.New(), batchID, "处理中.mp4", now),
		eligibleArchiveBatchUpdateFile(uuid.New(), batchID, "就绪.mp4", now),
		eligibleArchiveBatchUpdateFile(uuid.New(), batchID, "已存在.mp4", now),
	}
	files[0].MediaKind = "image"
	files[1].Status = "processing"
	files[2].Status = "ready"
	files[3].Status = "existing"

	plans, err := planArchiveFilenameBatchUpdate(files, archiveFilenameBatchUpdateInput(files...))
	if plans != nil {
		t.Fatalf("plans = %#v, want nil", plans)
	}
	batchErr := requireArchiveImportBatchUpdateError(t, err, ArchiveImportBatchReasonIneligibleTarget)
	requireArchiveImportBatchUpdateIssues(t, batchErr, files...)
}

func TestPlanArchiveFilenameBatchUpdateRejectsStaleTarget(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 12, 5, 0, 0, 0, time.UTC)
	file := eligibleArchiveBatchUpdateFile(uuid.New(), uuid.New(), "www.98T.la@ABC.mp4", now)
	in := ArchiveImportBatchUpdateInput{
		Targets:   []ArchiveImportBatchUpdateTarget{{ID: file.ID, UpdatedAt: now.Add(-time.Second)}},
		TitleMode: ArchiveImportTitleModeFilename,
	}

	plans, err := planArchiveFilenameBatchUpdate([]models.ArchiveImportFileListItem{file}, in)
	if plans != nil {
		t.Fatalf("plans = %#v, want nil", plans)
	}
	batchErr := requireArchiveImportBatchUpdateError(t, err, ArchiveImportBatchReasonStaleTarget)
	requireArchiveImportBatchUpdateIssues(t, batchErr, file)
}

func TestPlanArchiveFilenameBatchUpdateRejectsInvalidDerivedTitles(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 12, 5, 0, 0, 0, time.UTC)
	batchID := uuid.New()
	files := []models.ArchiveImportFileListItem{
		eligibleArchiveBatchUpdateFile(uuid.New(), batchID, "www.98T.la@.mp4", now),
		eligibleArchiveBatchUpdateFile(uuid.New(), batchID, strings.Repeat("长", 201)+".mp4", now),
	}

	plans, err := planArchiveFilenameBatchUpdate(files, archiveFilenameBatchUpdateInput(files...))
	if plans != nil {
		t.Fatalf("plans = %#v, want nil", plans)
	}
	batchErr := requireArchiveImportBatchUpdateError(t, err, ArchiveImportBatchReasonEmptyTitle)
	requireArchiveImportBatchUpdateIssues(t, batchErr, files...)
}

func eligibleArchiveBatchUpdateFile(id, batchID uuid.UUID, relativePath string, updatedAt time.Time) models.ArchiveImportFileListItem {
	return models.ArchiveImportFileListItem{
		ID:           id,
		BatchID:      batchID,
		EntryType:    "file",
		MediaKind:    "video",
		Status:       "pending",
		RelativePath: relativePath,
		Title:        "旧标题",
		UpdatedAt:    updatedAt,
	}
}

func archiveFilenameBatchUpdateInput(files ...models.ArchiveImportFileListItem) ArchiveImportBatchUpdateInput {
	targets := make([]ArchiveImportBatchUpdateTarget, 0, len(files))
	for _, file := range files {
		targets = append(targets, ArchiveImportBatchUpdateTarget{ID: file.ID, UpdatedAt: file.UpdatedAt})
	}
	return ArchiveImportBatchUpdateInput{Targets: targets, TitleMode: ArchiveImportTitleModeFilename}
}

func requireArchiveImportBatchUpdateError(t *testing.T, err error, reason string) *ArchiveImportBatchUpdateError {
	t.Helper()
	var batchErr *ArchiveImportBatchUpdateError
	if !errors.As(err, &batchErr) {
		t.Fatalf("error = %#v, want *ArchiveImportBatchUpdateError", err)
	}
	if batchErr.Reason != reason {
		t.Fatalf("error reason = %q, want %q", batchErr.Reason, reason)
	}
	return batchErr
}

func requireArchiveImportBatchUpdateIssues(t *testing.T, batchErr *ArchiveImportBatchUpdateError, files ...models.ArchiveImportFileListItem) {
	t.Helper()
	if len(batchErr.Issues) != len(files) {
		t.Fatalf("issues = %#v, want %d", batchErr.Issues, len(files))
	}
	for i, file := range files {
		issue := batchErr.Issues[i]
		if issue.ID != file.ID || issue.RelativePath != file.RelativePath || strings.TrimSpace(issue.Message) == "" {
			t.Fatalf("issue[%d] = %#v, want file %s", i, issue, file.ID)
		}
	}
}

func TestApplyArchiveImportBatchPlanSortsAllTargetIDs(t *testing.T) {
	t.Parallel()

	first := uuid.MustParse("ffffffff-ffff-4fff-8fff-ffffffffffff")
	second := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	third := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	targets := []ArchiveImportBatchUpdateTarget{
		{ID: first},
		{ID: second},
		{ID: third},
		{ID: second},
	}

	got := sortedArchiveImportBatchTargetIDs(targets)
	want := []uuid.UUID{second, second, third, first}
	if !slices.Equal(got, want) {
		t.Fatalf("sorted IDs = %#v, want %#v", got, want)
	}
	if targets[0].ID != first || targets[1].ID != second {
		t.Fatalf("input targets were mutated: %#v", targets)
	}
}

func TestApplyArchiveImportBatchPlanLocksOnlyFileRowsInFixedOrder(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("stop after query capture")
	tx := &queryFailingArchiveImportTx{err: sentinel}
	ids := []uuid.UUID{
		uuid.MustParse("00000000-0000-4000-8000-000000000001"),
		uuid.MustParse("11111111-1111-4111-8111-111111111111"),
	}

	_, err := lockArchiveImportFilesTx(context.Background(), tx, ids)
	if !errors.Is(err, sentinel) {
		t.Fatalf("lockArchiveImportFilesTx() error = %v, want %v", err, sentinel)
	}
	if !strings.Contains(tx.query, "ORDER BY f.id ASC") {
		t.Fatalf("lock query does not define UUID order:\n%s", tx.query)
	}
	if !strings.Contains(tx.query, "FOR UPDATE OF f") {
		t.Fatalf("lock query does not restrict locking to file rows:\n%s", tx.query)
	}
	if len(tx.args) != 1 {
		t.Fatalf("query args = %#v, want one ID slice", tx.args)
	}
	gotIDs, ok := tx.args[0].([]uuid.UUID)
	if !ok || !slices.Equal(gotIDs, ids) {
		t.Fatalf("query IDs = %#v, want %#v", tx.args[0], ids)
	}
}

func TestApplyArchiveImportBatchPlanRejectsInvalidPatchBeforeDatabaseAccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   ArchiveImportBatchUpdateInput
	}{
		{
			name: "unsupported title mode",
			in:   ArchiveImportBatchUpdateInput{TitleMode: "manual"},
		},
		{
			name: "invalid video type",
			in: ArchiveImportBatchUpdateInput{
				TitleMode:       ArchiveImportTitleModeFilename,
				UpdateVideoType: true,
				VideoType:       "series",
			},
		},
	}

	service := &ArchiveImportService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := service.BatchUpdateFiles(context.Background(), tt.in)
			if files != nil {
				t.Fatalf("files = %#v, want nil", files)
			}
			requireArchiveImportBatchUpdateError(t, err, ArchiveImportBatchReasonInvalidPatch)
		})
	}
}

func TestApplyArchiveImportBatchPlanCallerDefersRollbackBeforeWrites(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("archive_import_batch_update.go")
	if err != nil {
		t.Fatalf("ReadFile(archive_import_batch_update.go) error = %v", err)
	}
	source := string(raw)
	const signature = "func (s *ArchiveImportService) BatchUpdateFiles"
	start := strings.Index(source, signature)
	if start < 0 {
		t.Fatalf("%s not found", signature)
	}
	body := source[start+len(signature):]
	if end := strings.Index(body, "\nfunc "); end >= 0 {
		body = body[:end]
	}

	beginIndex := strings.Index(body, "s.db.Begin(ctx)")
	rollbackIndex := strings.Index(body, "defer tx.Rollback(ctx)")
	lockIndex := strings.Index(body, "lockArchiveImportFilesTx(ctx, tx, ids)")
	applyIndex := strings.Index(body, "applyArchiveImportBatchPlanTx(ctx, tx")
	commitIndex := strings.Index(body, "tx.Commit(ctx)")
	if beginIndex < 0 || rollbackIndex < 0 || lockIndex < 0 || applyIndex < 0 || commitIndex < 0 {
		t.Fatalf(
			"transaction calls missing: begin=%d rollback=%d lock=%d apply=%d commit=%d",
			beginIndex,
			rollbackIndex,
			lockIndex,
			applyIndex,
			commitIndex,
		)
	}
	if !(beginIndex < rollbackIndex && rollbackIndex < lockIndex && lockIndex < applyIndex && applyIndex < commitIndex) {
		t.Fatalf(
			"unexpected transaction order: begin=%d rollback=%d lock=%d apply=%d commit=%d",
			beginIndex,
			rollbackIndex,
			lockIndex,
			applyIndex,
			commitIndex,
		)
	}
}

func TestApplyArchiveImportBatchPlanRecomputesOnlyModifiedOverrides(t *testing.T) {
	t.Parallel()

	groupID := uuid.New()
	videoCollectionID := uuid.New()
	imageCollectionID := uuid.New()
	groupTitle := "组默认标题"
	groupDescription := "组默认描述"
	groupVideoType := "movie"
	group := models.ArchiveImportGroup{
		ID:                 groupID,
		Title:              &groupTitle,
		Description:        &groupDescription,
		Tags:               []string{"默认标签"},
		VideoType:          &groupVideoType,
		VideoCollectionIDs: []uuid.UUID{videoCollectionID},
		ImageCollectionIDs: []uuid.UUID{imageCollectionID},
	}
	baseFile := models.ArchiveImportFileListItem{
		ID:                 uuid.New(),
		GroupID:            &groupID,
		MediaKind:          "video",
		RelativePath:       "文件名标题.mp4",
		Title:              "文件名标题",
		Description:        groupDescription,
		Tags:               append([]string{}, group.Tags...),
		VideoType:          groupVideoType,
		VideoCollectionIDs: append([]uuid.UUID{}, group.VideoCollectionIDs...),
		ImageCollectionIDs: append([]uuid.UUID{}, group.ImageCollectionIDs...),
		FieldOverrides: map[string]bool{
			archiveImportOverrideTitle:              false,
			archiveImportOverrideDescription:        true,
			archiveImportOverrideTags:               true,
			archiveImportOverrideVideoType:          true,
			archiveImportOverrideVideoCollectionIDs: true,
			archiveImportOverrideImageCollectionIDs: true,
		},
	}
	tests := []struct {
		name             string
		in               ArchiveImportBatchUpdateInput
		wantOptionalFlag bool
	}{
		{
			name:             "preserves disabled field overrides",
			in:               ArchiveImportBatchUpdateInput{TitleMode: ArchiveImportTitleModeFilename},
			wantOptionalFlag: true,
		},
		{
			name: "recomputes enabled field overrides",
			in: ArchiveImportBatchUpdateInput{
				TitleMode:                ArchiveImportTitleModeFilename,
				UpdateTags:               true,
				UpdateVideoType:          true,
				UpdateVideoCollectionIDs: true,
				UpdateImageCollectionIDs: true,
			},
			wantOptionalFlag: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &failingArchiveImportTx{}
			err := applyArchiveImportBatchPlanTx(
				context.Background(),
				tx,
				[]archiveImportBatchPlannedFile{{File: baseFile}},
				models.ArchiveImportBatch{},
				map[uuid.UUID]models.ArchiveImportGroup{groupID: group},
				tt.in,
			)
			if err != nil {
				t.Fatalf("applyArchiveImportBatchPlanTx() error = %v", err)
			}
			if tx.calls != 1 || len(tx.execArgs) != 1 {
				t.Fatalf("exec calls = %d, args = %#v", tx.calls, tx.execArgs)
			}
			var got map[string]bool
			raw, ok := tx.execArgs[0][8].([]byte)
			if !ok {
				t.Fatalf("field overrides arg = %#v, want JSON bytes", tx.execArgs[0][8])
			}
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("unmarshal field overrides: %v", err)
			}
			if !got[archiveImportOverrideTitle] {
				t.Fatalf("title override = false, want true: %#v", got)
			}
			if got[archiveImportOverrideDescription] {
				t.Fatalf("description override = true, want false: %#v", got)
			}
			for _, field := range []string{
				archiveImportOverrideTags,
				archiveImportOverrideVideoType,
				archiveImportOverrideVideoCollectionIDs,
				archiveImportOverrideImageCollectionIDs,
			} {
				if got[field] != tt.wantOptionalFlag {
					t.Fatalf("%s override = %v, want %v: %#v", field, got[field], tt.wantOptionalFlag, got)
				}
			}
		})
	}
}

func TestApplyArchiveImportBatchPlanStopsAfterWriteFailure(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("forced update failure")
	tx := &failingArchiveImportTx{failAt: 2, err: sentinel}
	plans := make([]archiveImportBatchPlannedFile, 0, 3)
	for i := 0; i < 3; i++ {
		plans = append(plans, archiveImportBatchPlannedFile{File: models.ArchiveImportFileListItem{
			ID:           uuid.New(),
			MediaKind:    "video",
			RelativePath: "视频.mp4",
			Title:        "视频",
			VideoType:    "short",
		}})
	}

	err := applyArchiveImportBatchPlanTx(
		context.Background(),
		tx,
		plans,
		models.ArchiveImportBatch{},
		nil,
		ArchiveImportBatchUpdateInput{TitleMode: ArchiveImportTitleModeFilename},
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf("applyArchiveImportBatchPlanTx() error = %v, want %v", err, sentinel)
	}
	if tx.calls != 2 {
		t.Fatalf("Exec() calls = %d, want 2", tx.calls)
	}
}

type failingArchiveImportTx struct {
	pgx.Tx
	calls    int
	failAt   int
	err      error
	execArgs [][]any
}

func (tx *failingArchiveImportTx) Exec(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
	tx.calls++
	tx.execArgs = append(tx.execArgs, append([]any{}, args...))
	if tx.calls == tx.failAt {
		return pgconn.CommandTag{}, tx.err
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

type queryFailingArchiveImportTx struct {
	pgx.Tx
	query string
	args  []any
	err   error
}

func (tx *queryFailingArchiveImportTx) Query(_ context.Context, query string, args ...any) (pgx.Rows, error) {
	tx.query = query
	tx.args = append([]any{}, args...)
	return nil, tx.err
}
