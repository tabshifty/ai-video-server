package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
)

type telegramSourceFakeRepository struct {
	sources   map[uuid.UUID]models.TelegramSource
	counts    map[uuid.UUID]map[string]int
	created   []models.TelegramSource
	patches   map[uuid.UUID][]repository.TelegramSourcePatch
	createErr error
	listErr   error
	getErr    error
	updateErr error
	countErr  error
}

func newTelegramSourceFakeRepository() *telegramSourceFakeRepository {
	return &telegramSourceFakeRepository{
		sources: make(map[uuid.UUID]models.TelegramSource),
		counts:  make(map[uuid.UUID]map[string]int),
		patches: make(map[uuid.UUID][]repository.TelegramSourcePatch),
	}
}

func (r *telegramSourceFakeRepository) CreateTelegramSource(_ context.Context, source models.TelegramSource) error {
	if r.createErr != nil {
		return r.createErr
	}
	for _, existing := range r.sources {
		if existing.ChatID != 0 && source.ChatID != 0 && existing.ChatID == source.ChatID {
			return errors.New("duplicate chat id")
		}
	}
	r.sources[source.ID] = source
	r.created = append(r.created, source)
	return nil
}

func (r *telegramSourceFakeRepository) ListTelegramSources(_ context.Context, _ bool) ([]models.TelegramSource, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	items := make([]models.TelegramSource, 0, len(r.sources))
	for _, source := range r.sources {
		items = append(items, source)
	}
	return items, nil
}

func (r *telegramSourceFakeRepository) GetTelegramSource(_ context.Context, sourceID uuid.UUID) (models.TelegramSource, error) {
	if r.getErr != nil {
		return models.TelegramSource{}, r.getErr
	}
	source, ok := r.sources[sourceID]
	if !ok {
		return models.TelegramSource{}, repository.ErrTelegramSourceNotFound
	}
	return source, nil
}

func (r *telegramSourceFakeRepository) UpdateTelegramSource(_ context.Context, sourceID uuid.UUID, patch repository.TelegramSourcePatch) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	source, ok := r.sources[sourceID]
	if !ok {
		return repository.ErrTelegramSourceNotFound
	}
	r.patches[sourceID] = append(r.patches[sourceID], patch)
	if patch.Enabled != nil {
		source.Enabled = *patch.Enabled
	}
	if patch.SyncStatus != nil {
		source.SyncStatus = *patch.SyncStatus
	}
	if patch.LastError != nil {
		source.LastError = *patch.LastError
	}
	if patch.ClearNextRetryAt {
		source.NextRetryAt = nil
	} else if patch.NextRetryAt != nil {
		source.NextRetryAt = patch.NextRetryAt
	}
	if patch.HistoryCursorMessageID != nil {
		source.HistoryCursorMessageID = *patch.HistoryCursorMessageID
	}
	if patch.ClearBackfillCompletedAt {
		source.BackfillCompletedAt = nil
	} else if patch.BackfillCompletedAt != nil {
		source.BackfillCompletedAt = patch.BackfillCompletedAt
	}
	r.sources[sourceID] = source
	return nil
}

func (r *telegramSourceFakeRepository) CountTelegramMediaByStatus(_ context.Context, sourceID uuid.UUID) (map[string]int, error) {
	if r.countErr != nil {
		return nil, r.countErr
	}
	return r.counts[sourceID], nil
}

type telegramSourceFakeTasks struct {
	sourceIDs []uuid.UUID
	err       error
}

func (q *telegramSourceFakeTasks) EnqueueSourceSync(sourceID uuid.UUID) error {
	if q.err != nil {
		return q.err
	}
	q.sourceIDs = append(q.sourceIDs, sourceID)
	return nil
}

func TestTelegramSourceServiceRejectsEmptyChatReference(t *testing.T) {
	repo := newTelegramSourceFakeRepository()
	service := NewTelegramSourceService(repo, &telegramSourceFakeTasks{})

	_, err := service.Add(context.Background(), "   ")
	if !errors.Is(err, ErrTelegramChatRefRequired) {
		t.Fatalf("Add() error = %v, want ErrTelegramChatRefRequired", err)
	}
	if len(repo.created) != 0 {
		t.Fatalf("created source count = %d, want 0", len(repo.created))
	}
}

func TestTelegramSourceServiceRejectsDuplicateChatReference(t *testing.T) {
	repo := newTelegramSourceFakeRepository()
	existingID := uuid.New()
	repo.sources[existingID] = models.TelegramSource{ID: existingID, ChatRef: "@TestGroup", Enabled: true, SyncStatus: "live"}
	service := NewTelegramSourceService(repo, &telegramSourceFakeTasks{})

	_, err := service.Add(context.Background(), " @testgroup/ ")
	if !errors.Is(err, ErrTelegramSourceAlreadyExists) {
		t.Fatalf("Add() error = %v, want ErrTelegramSourceAlreadyExists", err)
	}
}

func TestTelegramSourceServiceAddsSourceAndEnqueuesSync(t *testing.T) {
	repo := newTelegramSourceFakeRepository()
	tasks := &telegramSourceFakeTasks{}
	service := NewTelegramSourceService(repo, tasks)

	source, err := service.Add(context.Background(), " https://t.me/TestGroup/ ")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if source.ID == uuid.Nil || source.ChatRef != "https://t.me/TestGroup" || !source.Enabled || source.SyncStatus != "pending" {
		t.Fatalf("created source = %+v", source)
	}
	if len(tasks.sourceIDs) != 1 || tasks.sourceIDs[0] != source.ID {
		t.Fatalf("source sync tasks = %v, want [%s]", tasks.sourceIDs, source.ID)
	}
}

func TestTelegramSourceServicePauseAndResumePreserveCursor(t *testing.T) {
	repo := newTelegramSourceFakeRepository()
	tasks := &telegramSourceFakeTasks{}
	sourceID := uuid.New()
	repo.sources[sourceID] = models.TelegramSource{
		ID:                     sourceID,
		ChatID:                 -100123,
		ChatRef:                "@testgroup",
		Enabled:                true,
		SyncStatus:             "live",
		HistoryCursorMessageID: 345,
		NextRetryAt:            telegramTimePtr(time.Now().Add(time.Hour)),
	}
	service := NewTelegramSourceService(repo, tasks)

	paused, err := service.Pause(context.Background(), sourceID)
	if err != nil {
		t.Fatalf("Pause() error = %v", err)
	}
	if paused.Enabled || paused.SyncStatus != "paused" || paused.HistoryCursorMessageID != 345 {
		t.Fatalf("paused source = %+v", paused)
	}

	resumed, err := service.Resume(context.Background(), sourceID)
	if err != nil {
		t.Fatalf("Resume() error = %v", err)
	}
	if !resumed.Enabled || resumed.SyncStatus != "pending" || resumed.HistoryCursorMessageID != 345 || resumed.NextRetryAt != nil {
		t.Fatalf("resumed source = %+v", resumed)
	}
	if len(tasks.sourceIDs) != 1 || tasks.sourceIDs[0] != sourceID {
		t.Fatalf("resume source sync tasks = %v, want [%s]", tasks.sourceIDs, sourceID)
	}
}

func TestTelegramSourceServiceProgressIncludesStatusCounts(t *testing.T) {
	repo := newTelegramSourceFakeRepository()
	sourceID := uuid.New()
	source := models.TelegramSource{ID: sourceID, ChatRef: "@testgroup", SyncStatus: "live"}
	repo.sources[sourceID] = source
	repo.counts[sourceID] = map[string]int{"imported": 4, "failed": 1}
	service := NewTelegramSourceService(repo, &telegramSourceFakeTasks{})

	progress, err := service.GetProgress(context.Background(), sourceID)
	if err != nil {
		t.Fatalf("GetProgress() error = %v", err)
	}
	if progress.Source.ID != sourceID || progress.Counts["imported"] != 4 || progress.Counts["failed"] != 1 {
		t.Fatalf("progress = %+v", progress)
	}
}

func TestTelegramSourceServiceStartBackfillResetsCursor(t *testing.T) {
	repo := newTelegramSourceFakeRepository()
	tasks := &telegramSourceFakeTasks{}
	sourceID := uuid.New()
	completedAt := time.Now().UTC().Add(-time.Hour)
	repo.sources[sourceID] = models.TelegramSource{
		ID:                     sourceID,
		ChatRef:                "@testgroup",
		Enabled:                false,
		SyncStatus:             "paused",
		HistoryCursorMessageID: 345,
		BackfillCompletedAt:    &completedAt,
	}
	service := NewTelegramSourceService(repo, tasks)

	source, err := service.StartBackfill(context.Background(), sourceID)
	if err != nil {
		t.Fatalf("StartBackfill() error = %v", err)
	}
	if !source.Enabled || source.SyncStatus != "pending" || source.HistoryCursorMessageID != 0 || source.BackfillCompletedAt != nil {
		t.Fatalf("backfill source = %+v", source)
	}
	if len(tasks.sourceIDs) != 1 || tasks.sourceIDs[0] != sourceID {
		t.Fatalf("backfill source sync tasks = %v, want [%s]", tasks.sourceIDs, sourceID)
	}
}
