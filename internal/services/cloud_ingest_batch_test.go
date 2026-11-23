package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"video-server/internal/models"
)

type fakeIngestBatchRepo struct {
	created      models.CloudIngestBatch
	createErr    error
	attachCalls  []int64
	attachTitle  []string
	attachForced []bool
	attachOK     bool
	awaiting     []int64
	deleteCalls  int
}

func (f *fakeIngestBatchRepo) Create(_ context.Context, batch models.CloudIngestBatch) (models.CloudIngestBatch, error) {
	if f.createErr != nil {
		return models.CloudIngestBatch{}, f.createErr
	}
	batch.ID = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	f.created = batch
	return batch, nil
}

func (f *fakeIngestBatchRepo) AttachTask(_ context.Context, taskID int64, _ uuid.UUID, title string, forced bool) (bool, error) {
	f.attachCalls = append(f.attachCalls, taskID)
	f.attachTitle = append(f.attachTitle, title)
	f.attachForced = append(f.attachForced, forced)
	return f.attachOK, nil
}

func (f *fakeIngestBatchRepo) Delete(context.Context, uuid.UUID) error {
	f.deleteCalls++
	return nil
}

func (f *fakeIngestBatchRepo) ListAwaitingAutoImport(context.Context, int) ([]int64, error) {
	return f.awaiting, nil
}

type fakeBatchAutoImporter struct {
	taskIDs []int64
	err     error
}

func (f *fakeBatchAutoImporter) AutoImportFromBatch(_ context.Context, taskID int64) error {
	f.taskIDs = append(f.taskIDs, taskID)
	return f.err
}

func TestNormalizeCloudIngestBatchInput(t *testing.T) {
	tests := []struct {
		name    string
		input   models.CloudIngestBatchInput
		wantErr bool
		check   func(t *testing.T, batch models.CloudIngestBatch, links []models.CloudIngestBatchLink)
	}{
		{
			name:  "defaults to uniform short",
			input: models.CloudIngestBatchInput{UniformTitle: "合辑", Links: []models.CloudIngestBatchLink{{Link: "ed2k://x"}}},
			check: func(t *testing.T, batch models.CloudIngestBatch, _ []models.CloudIngestBatchLink) {
				if batch.TitleMode != models.CloudIngestTitleUniform || batch.VideoType != "short" {
					t.Fatalf("unexpected defaults: %+v", batch)
				}
				if batch.SiteCategory != "" {
					t.Fatalf("short must not keep site category: %q", batch.SiteCategory)
				}
			},
		},
		{
			name:    "uniform title required",
			input:   models.CloudIngestBatchInput{Links: []models.CloudIngestBatchLink{{Link: "ed2k://x"}}},
			wantErr: true,
		},
		{
			name: "filename mode drops batch title",
			input: models.CloudIngestBatchInput{
				TitleMode:    models.CloudIngestTitleFilename,
				UniformTitle: "ignored",
				Links:        []models.CloudIngestBatchLink{{Link: "ed2k://x", Title: "第一集"}},
			},
			check: func(t *testing.T, batch models.CloudIngestBatch, links []models.CloudIngestBatchLink) {
				if batch.UniformTitle != "" {
					t.Fatalf("filename mode must not keep a batch title: %q", batch.UniformTitle)
				}
				if links[0].Title != "第一集" {
					t.Fatalf("per-link title lost: %+v", links[0])
				}
			},
		},
		{
			name:    "invalid title mode",
			input:   models.CloudIngestBatchInput{TitleMode: "random", Links: []models.CloudIngestBatchLink{{Link: "ed2k://x"}}},
			wantErr: true,
		},
		{
			name:    "invalid video type",
			input:   models.CloudIngestBatchInput{UniformTitle: "t", VideoType: "anime", Links: []models.CloudIngestBatchLink{{Link: "ed2k://x"}}},
			wantErr: true,
		},
		{
			name: "non short rejects collections",
			input: models.CloudIngestBatchInput{
				UniformTitle:  "t",
				VideoType:     "movie",
				CollectionIDs: []uuid.UUID{uuid.New()},
				Links:         []models.CloudIngestBatchLink{{Link: "ed2k://x"}},
			},
			wantErr: true,
		},
		{
			name: "av defaults site category",
			input: models.CloudIngestBatchInput{
				UniformTitle: "t",
				VideoType:    "av",
				Links:        []models.CloudIngestBatchLink{{Link: "ed2k://x"}},
			},
			check: func(t *testing.T, batch models.CloudIngestBatch, _ []models.CloudIngestBatchLink) {
				if batch.SiteCategory != avSiteCategoryJapanese {
					t.Fatalf("av should default to japanese, got %q", batch.SiteCategory)
				}
			},
		},
		{
			name:    "no links",
			input:   models.CloudIngestBatchInput{UniformTitle: "t"},
			wantErr: true,
		},
		{
			name: "too many links",
			input: func() models.CloudIngestBatchInput {
				links := make([]models.CloudIngestBatchLink, models.CloudDownloadBatchLimit+1)
				for i := range links {
					links[i] = models.CloudIngestBatchLink{Link: "ed2k://x"}
				}
				return models.CloudIngestBatchInput{UniformTitle: "t", Links: links}
			}(),
			wantErr: true,
		},
		{
			name: "counts forced links",
			input: models.CloudIngestBatchInput{
				UniformTitle: "t",
				Links: []models.CloudIngestBatchLink{
					{Link: "ed2k://a"},
					{Link: "ed2k://b", Forced: true},
				},
			},
			check: func(t *testing.T, batch models.CloudIngestBatch, _ []models.CloudIngestBatchLink) {
				if batch.ForcedCount != 1 {
					t.Fatalf("forced count = %d, want 1", batch.ForcedCount)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			batch, links, err := normalizeCloudIngestBatchInput(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got batch %+v", batch)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, batch, links)
			}
		})
	}
}

func TestSubmitIngestBatchAttachesMembersAndForced(t *testing.T) {
	existingHash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	freshHash := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	existingID := int64(77)
	freshID := int64(88)

	repo := &fakeCloudRepo{
		statusByHash: map[string]models.CloudDownloadStatus{
			existingHash: {TaskID: existingID, Status: models.CloudDownloadCompleted},
		},
	}
	// Created tasks report created=true and carry the fresh id.
	repo.createCreated = true
	repo.nextCreatedID = freshID

	batchRepo := &fakeIngestBatchRepo{attachOK: true}
	importer := &fakeBatchAutoImporter{}
	svc := &CloudDownloadService{
		repo:          repo,
		batchRepo:     batchRepo,
		batchImporter: importer,
		enqueuer:      &fakeCloudEnqueuer{},
	}

	input := models.CloudIngestBatchInput{
		TitleMode:    models.CloudIngestTitleFilename,
		VideoType:    "short",
		UniformTitle: "",
		Links: []models.CloudIngestBatchLink{
			{Link: testED2KLink("a.mp4", freshHash), Title: "新片"},
			{Link: testED2KLink("b.mp4", existingHash), Title: "旧片"},
			{Link: testED2KLink("c.mp4", freshHash), Title: "重复"},
			{Link: testED2KLink("d.mp4", "cccccccccccccccccccccccccccccccc"), Title: "强制", Forced: true},
		},
	}

	created, result, err := svc.SubmitIngestBatch(context.Background(), input, nil)
	if err != nil {
		t.Fatalf("SubmitIngestBatch: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Fatalf("batch was not created")
	}
	if len(result.Items) != 4 {
		t.Fatalf("items = %d, want 4", len(result.Items))
	}
	if result.Items[0].Status != models.CloudBatchSubmitted {
		t.Fatalf("item0 = %q, want submitted", result.Items[0].Status)
	}
	if result.Items[1].Status != models.CloudBatchSkippedCompleted {
		t.Fatalf("item1 = %q, want skipped_completed", result.Items[1].Status)
	}
	if result.Items[2].Status != models.CloudBatchDuplicate {
		t.Fatalf("item2 = %q, want duplicate", result.Items[2].Status)
	}
	if !result.Items[3].Forced {
		t.Fatalf("item3 should be marked forced")
	}
	if result.Summary.Forced != 1 {
		t.Fatalf("summary forced = %d, want 1", result.Summary.Forced)
	}
	// Three members resolve to a task: the fresh link, the completed existing
	// one, and the forced link; the duplicate never reaches attach.
	if len(batchRepo.attachCalls) != 3 {
		t.Fatalf("attach calls = %v, want 3", batchRepo.attachCalls)
	}
	if batchRepo.attachTitle[1] != "旧片" {
		t.Fatalf("per-link title not forwarded: %v", batchRepo.attachTitle)
	}
	// 强制标记必须落在任务行上，列表页才能单行标注。
	if len(batchRepo.attachForced) != 3 || batchRepo.attachForced[2] != true || batchRepo.attachForced[0] != false {
		t.Fatalf("forced flags not forwarded: %v", batchRepo.attachForced)
	}
	// The completed existing task must auto-import immediately.
	foundExisting := false
	for _, id := range importer.taskIDs {
		if id == existingID {
			foundExisting = true
		}
	}
	if !foundExisting {
		t.Fatalf("completed member should trigger auto import, got %v", importer.taskIDs)
	}
}

func TestReconcileBatchIngestRetriesAwaitingMembers(t *testing.T) {
	batchRepo := &fakeIngestBatchRepo{awaiting: []int64{5, 6}}
	importer := &fakeBatchAutoImporter{}
	svc := &CloudDownloadService{batchRepo: batchRepo, batchImporter: importer}
	svc.reconcileBatchIngest(context.Background())
	if len(importer.taskIDs) != 2 {
		t.Fatalf("auto import calls = %v, want 2", importer.taskIDs)
	}
}

func TestDecideCloudBatchExisting(t *testing.T) {
	if d := decideCloudBatchExisting(models.CloudDownloadCompleted); !d.skip || d.status != models.CloudBatchSkippedCompleted {
		t.Fatalf("completed decision = %+v", d)
	}
	if d := decideCloudBatchExisting(models.CloudDownloadFailed); d.skip {
		t.Fatalf("failed must be retried, got %+v", d)
	}
	if d := decideCloudBatchExisting("future_state"); !d.skip || d.status != models.CloudBatchSkippedActive {
		t.Fatalf("unknown decision = %+v", d)
	}
}

func TestSubmitIngestBatchSkipsBatchWhenNoCandidateRemains(t *testing.T) {
	batchRepo := &fakeIngestBatchRepo{attachOK: true}
	svc := &CloudDownloadService{
		repo:          &fakeCloudRepo{},
		batchRepo:     batchRepo,
		batchImporter: &fakeBatchAutoImporter{},
		enqueuer:      &fakeCloudEnqueuer{},
	}
	_, result, err := svc.SubmitIngestBatch(context.Background(), models.CloudIngestBatchInput{
		UniformTitle: "合辑",
		Links:        []models.CloudIngestBatchLink{{Link: "not-an-ed2k-link"}},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if batchRepo.created.ID != uuid.Nil {
		t.Fatalf("must not create a batch when every link is invalid")
	}
	if len(result.Items) != 1 || result.Items[0].Status != models.CloudBatchInvalid {
		t.Fatalf("items = %+v, want one invalid item", result.Items)
	}
}

func TestNormalizeCloudIngestBatchInputKeepsArraysNonNull(t *testing.T) {
	// NOT NULL 数组列：nil 切片会被 pgx 编码成 SQL NULL 而违反约束。
	batch, _, err := normalizeCloudIngestBatchInput(models.CloudIngestBatchInput{
		UniformTitle: "合辑",
		Links:        []models.CloudIngestBatchLink{{Link: "ed2k://x"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if batch.Tags == nil || batch.ActorNames == nil || batch.ActorIDs == nil || batch.CollectionIDs == nil {
		t.Fatalf("batch arrays must not be nil: %+v", batch)
	}
}

func TestSubmitIngestBatchDropsBatchWhenStatusLookupFails(t *testing.T) {
	batchRepo := &fakeIngestBatchRepo{attachOK: true}
	svc := &CloudDownloadService{
		repo:          &fakeCloudRepo{statusErr: errors.New("boom")},
		batchRepo:     batchRepo,
		batchImporter: &fakeBatchAutoImporter{},
		enqueuer:      &fakeCloudEnqueuer{},
	}
	_, _, err := svc.SubmitIngestBatch(context.Background(), models.CloudIngestBatchInput{
		UniformTitle: "合辑",
		Links:        []models.CloudIngestBatchLink{{Link: testED2KLink("a.mp4", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")}},
	}, nil)
	if err == nil {
		t.Fatalf("expected the status lookup error to surface")
	}
	if batchRepo.deleteCalls != 1 {
		t.Fatalf("delete calls = %d, want 1 (compensate the orphan batch)", batchRepo.deleteCalls)
	}
}
