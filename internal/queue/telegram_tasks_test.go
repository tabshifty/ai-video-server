package queue

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

func TestTelegramTaskOptionsUseQueueRetryAndTimeout(t *testing.T) {
	t.Parallel()

	opts := buildTelegramTaskOptions("telegram-realtime")
	assertOption(t, opts, asynq.QueueOpt, "telegram-realtime")
	assertOption(t, opts, asynq.MaxRetryOpt, 3)
	assertOption(t, opts, asynq.TimeoutOpt, 6*time.Hour)
}

func TestTelegramTaskPayloadsContainOnlyIdentifiers(t *testing.T) {
	t.Parallel()

	mediaID := uuid.New()
	mediaRaw, err := json.Marshal(TelegramMediaTaskPayload{MediaID: mediaID.String()})
	if err != nil {
		t.Fatalf("marshal media payload: %v", err)
	}
	var media map[string]any
	if err := json.Unmarshal(mediaRaw, &media); err != nil {
		t.Fatalf("decode media payload: %v", err)
	}
	if len(media) != 1 || media["media_id"] != mediaID.String() {
		t.Fatalf("unexpected media payload: %s", mediaRaw)
	}

	sourceID := uuid.New()
	sourceRaw, err := json.Marshal(TelegramSourceTaskPayload{SourceID: sourceID.String()})
	if err != nil {
		t.Fatalf("marshal source payload: %v", err)
	}
	var source map[string]any
	if err := json.Unmarshal(sourceRaw, &source); err != nil {
		t.Fatalf("decode source payload: %v", err)
	}
	if len(source) != 1 || source["source_id"] != sourceID.String() {
		t.Fatalf("unexpected source payload: %s", sourceRaw)
	}
}

func TestNewTelegramTaskEnqueuerUsesSafeQueueDefaults(t *testing.T) {
	t.Parallel()

	enqueuer := NewTelegramTaskEnqueuer("127.0.0.1:6379", "", "", "", "")
	defer enqueuer.Close()
	if enqueuer.realtimeQueue != "telegram-realtime" || enqueuer.backfillQueue != "telegram-backfill" || enqueuer.controlQueue != "telegram-control" {
		t.Fatalf("unexpected queue defaults: %+v", enqueuer)
	}
}
