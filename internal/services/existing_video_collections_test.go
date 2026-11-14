package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"video-server/internal/models"
)

type fakeExistingCollectionRepo struct {
	relations       map[uuid.UUID]struct{}
	imageIDs        []uuid.UUID
	imageFillResult bool
	addErr          error
	imageErr        error
}

func newFakeExistingCollectionRepo() *fakeExistingCollectionRepo {
	return &fakeExistingCollectionRepo{relations: map[uuid.UUID]struct{}{}, imageFillResult: true}
}

func (f *fakeExistingCollectionRepo) AddVideoCollectionsByIDsAnyTypeReturningAdded(_ context.Context, _ uuid.UUID, ids []uuid.UUID) (int, error) {
	if f.addErr != nil {
		return 0, f.addErr
	}
	added := 0
	for _, id := range ids {
		if _, ok := f.relations[id]; ok {
			continue
		}
		f.relations[id] = struct{}{}
		added++
	}
	return added, nil
}

func (f *fakeExistingCollectionRepo) FillVideoImageCollectionIfEmpty(_ context.Context, _ uuid.UUID, imageCollectionID uuid.UUID) (bool, error) {
	if f.imageErr != nil {
		return false, f.imageErr
	}
	f.imageIDs = append(f.imageIDs, imageCollectionID)
	return f.imageFillResult, nil
}

func TestAttachExistingVideoCollectionsIsIdempotentUnion(t *testing.T) {
	t.Parallel()

	repo := newFakeExistingCollectionRepo()
	video := models.Video{ID: uuid.New(), Type: "short", Status: "ready"}
	first := uuid.New()
	second := uuid.New()

	attachment, err := AttachExistingVideoCollections(context.Background(), repo, video, []uuid.UUID{first, second, first, uuid.Nil}, nil)
	if err != nil {
		t.Fatalf("AttachExistingVideoCollections() error = %v", err)
	}
	if attachment.AddedCollections != 2 {
		t.Fatalf("added = %d, want 2", attachment.AddedCollections)
	}

	attachment, err = AttachExistingVideoCollections(context.Background(), repo, video, []uuid.UUID{first, second}, nil)
	if err != nil {
		t.Fatalf("second attach error = %v", err)
	}
	if attachment.AddedCollections != 0 {
		t.Fatalf("second add = %d, want 0", attachment.AddedCollections)
	}
}

func TestAttachExistingVideoCollectionsSkipsPendingDelete(t *testing.T) {
	t.Parallel()

	repo := newFakeExistingCollectionRepo()
	video := models.Video{ID: uuid.New(), Type: "short", Status: "pending_delete"}

	attachment, err := AttachExistingVideoCollections(context.Background(), repo, video, []uuid.UUID{uuid.New()}, nil)
	if err != nil {
		t.Fatalf("AttachExistingVideoCollections() error = %v", err)
	}
	if !attachment.VideoPendingDelete {
		t.Fatal("pending delete video should be reported")
	}
	if attachment.AddedCollections != 0 || len(repo.relations) != 0 {
		t.Fatalf("pending delete video must not receive relations: %+v", repo.relations)
	}
}

func TestAttachExistingVideoCollectionsFillsImageCollectionOnlyWhenEmpty(t *testing.T) {
	t.Parallel()

	repo := newFakeExistingCollectionRepo()
	imageCollectionID := uuid.New()

	filled, err := AttachExistingVideoCollections(context.Background(), repo, models.Video{ID: uuid.New(), Status: "ready"}, nil, &imageCollectionID)
	if err != nil {
		t.Fatalf("attach error = %v", err)
	}
	if !filled.AddedImageCollection || len(repo.imageIDs) != 1 {
		t.Fatalf("image collection should be filled: %+v", filled)
	}

	existing := imageCollectionID
	kept, err := AttachExistingVideoCollections(context.Background(), repo, models.Video{ID: uuid.New(), Status: "ready", ImageCollectionID: &existing}, nil, &imageCollectionID)
	if err != nil {
		t.Fatalf("attach error = %v", err)
	}
	if kept.AddedImageCollection || len(repo.imageIDs) != 1 {
		t.Fatalf("existing image collection must not be overwritten: %+v", kept)
	}

	// 合集不存在或并发已写入时 repo 返回 false，不能向调用方虚报补齐。
	repo.imageFillResult = false
	notFilled, err := AttachExistingVideoCollections(context.Background(), repo, models.Video{ID: uuid.New(), Status: "ready"}, nil, &imageCollectionID)
	if err != nil {
		t.Fatalf("attach error = %v", err)
	}
	if notFilled.AddedImageCollection {
		t.Fatalf("repo reported no write, attachment must not claim it: %+v", notFilled)
	}
}

func TestAttachExistingVideoCollectionsPropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	repo := newFakeExistingCollectionRepo()
	repo.addErr = errors.New("boom")
	attachment, err := AttachExistingVideoCollections(context.Background(), repo, models.Video{ID: uuid.New(), Status: "ready"}, []uuid.UUID{uuid.New()}, nil)
	if err == nil {
		t.Fatal("expected repository error")
	}
	if !errors.Is(err, repo.addErr) {
		t.Fatalf("error should wrap repository error, got %v", err)
	}
	if attachment.AddedCollections != 0 {
		t.Fatalf("failed attach must not report additions: %+v", attachment)
	}
}
