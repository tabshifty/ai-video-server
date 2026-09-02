package main

import (
	"testing"

	"video-server/internal/config"
)

func TestTelegramQueueWeightsPrioritizeRealtimeDownloads(t *testing.T) {
	weights := telegramQueueWeights(config.Config{
		TelegramRealtimeQueue: "realtime",
		TelegramBackfillQueue: "backfill",
		TelegramControlQueue:  "control",
	})
	if weights["realtime"] != 10 || weights["backfill"] != 1 || weights["control"] != 1 {
		t.Fatalf("Telegram queue weights = %#v, want realtime:10 backfill:1 control:1", weights)
	}
}
