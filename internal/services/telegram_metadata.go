package services

import (
	"path/filepath"
	"strings"
	"time"

	"video-server/internal/telegram"
)

// BuildTelegramVideoMetadata builds the title, description, and JSON-compatible metadata for a Telegram video.
func BuildTelegramVideoMetadata(message telegram.Message, groupTitle string, importedAt time.Time) (string, string, map[string]any) {
	caption := message.Caption
	title := strings.TrimSpace(caption)
	if title == "" {
		filename := strings.TrimSpace(filepath.Base(message.Filename))
		title = strings.TrimSpace(strings.TrimSuffix(filename, filepath.Ext(filename)))
		if title == "" {
			title = "untitled"
		}
	}

	sentAt := formatTelegramTime(message.SentAt)
	descriptionParts := make([]string, 0, 4)
	if caption != "" {
		descriptionParts = append(descriptionParts, "Caption：\n"+caption)
	}
	if group := strings.TrimSpace(groupTitle); group != "" {
		descriptionParts = append(descriptionParts, "群组："+group)
	}
	if link := strings.TrimSpace(message.MessageURL); link != "" {
		descriptionParts = append(descriptionParts, "消息链接："+link)
	}
	descriptionParts = append(descriptionParts, "发送时间："+sentAt)

	metadata := map[string]any{
		"source":            "telegram",
		"original_filename": message.Filename,
		"chat_id":           message.ChatID,
		"message_id":        message.MessageID,
		"document_id":       message.DocumentID,
		"sent_at":           sentAt,
		"imported_at":       formatTelegramTime(importedAt),
	}
	if message.DocumentDCID > 0 {
		metadata["document_dc_id"] = message.DocumentDCID
	}
	if group := strings.TrimSpace(groupTitle); group != "" {
		metadata["group_title"] = group
	}
	return title, strings.Join(descriptionParts, "\n\n"), metadata
}

func formatTelegramTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
