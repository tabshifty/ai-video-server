package services

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"video-server/internal/telegram"
)

func TestBuildTelegramVideoMetadataUsesCaptionAsTitleAndPreservesSourceDetails(t *testing.T) {
	t.Parallel()

	sentAt := time.Date(2026, 9, 2, 8, 9, 10, 0, time.UTC)
	importedAt := sentAt.Add(time.Hour)
	message := telegram.Message{
		ChatID:     -100123,
		MessageID:  42,
		DocumentID: 777,
		Filename:   "ignored-name.mp4",
		Caption:    "  完整 caption 原文  ",
		MessageURL: "https://t.me/c/123/42",
		SentAt:     sentAt,
	}

	title, description, metadata := BuildTelegramVideoMetadata(message, "测试群组", importedAt)
	if title != "完整 caption 原文" {
		t.Fatalf("title = %q, want caption title", title)
	}
	for _, want := range []string{"完整 caption 原文", "测试群组", "https://t.me/c/123/42", sentAt.Format(time.RFC3339)} {
		if !strings.Contains(description, want) {
			t.Fatalf("description %q does not contain %q", description, want)
		}
	}
	if metadata["original_filename"] != "ignored-name.mp4" || metadata["chat_id"] != int64(-100123) || metadata["message_id"] != int64(42) || metadata["document_id"] != int64(777) {
		t.Fatalf("unexpected Telegram metadata: %#v", metadata)
	}
	if metadata["sent_at"] != sentAt.Format(time.RFC3339) || metadata["imported_at"] != importedAt.Format(time.RFC3339) {
		t.Fatalf("unexpected Telegram timestamps: %#v", metadata)
	}
	if _, err := json.Marshal(metadata); err != nil {
		t.Fatalf("metadata is not JSON-compatible: %v", err)
	}
}

func TestBuildTelegramVideoMetadataFallsBackToFilenameWithoutExtension(t *testing.T) {
	t.Parallel()

	title, description, metadata := BuildTelegramVideoMetadata(telegram.Message{
		Filename: "folder/clip.final.mkv",
		SentAt:   time.Date(2026, 9, 2, 8, 0, 0, 0, time.UTC),
	}, "群组", time.Time{})
	if title != "clip.final" {
		t.Fatalf("title = %q, want filename without extension", title)
	}
	if !strings.Contains(description, "群组：群组") {
		t.Fatalf("description = %q, want group", description)
	}
	if metadata["document_id"] != int64(0) || metadata["sent_at"] != "2026-09-02T08:00:00Z" {
		t.Fatalf("unexpected fallback metadata: %#v", metadata)
	}
}
