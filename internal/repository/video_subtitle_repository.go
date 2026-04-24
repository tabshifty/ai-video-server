package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/utils"
)

func (r *VideoRepository) CreateVideoSubtitle(ctx context.Context, subtitle models.VideoSubtitle) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO video_subtitles (
  id, video_id, source_type, status, language_code, label, format, mime_type, stored_path, file_size, is_default, sort_order, metadata
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
`, subtitle.ID, subtitle.VideoID, subtitle.SourceType, subtitle.Status, subtitle.LanguageCode, subtitle.Label, subtitle.Format, subtitle.MIMEType, subtitle.StoredPath, subtitle.FileSize, subtitle.IsDefault, subtitle.SortOrder, subtitle.Metadata)
	if err != nil {
		return fmt.Errorf("insert video subtitle: %w", err)
	}
	return nil
}

func (r *VideoRepository) ListVideoSubtitles(ctx context.Context, videoID uuid.UUID) ([]models.VideoSubtitle, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, video_id, source_type, status, language_code, label, format, mime_type, stored_path, file_size, is_default, sort_order, metadata, created_at, updated_at
FROM video_subtitles
WHERE video_id = $1 AND status = 'ready'
ORDER BY
  CASE WHEN source_type = 'uploaded' AND is_default THEN 0
       WHEN source_type = 'uploaded' THEN 1
       WHEN is_default THEN 2
       ELSE 3
  END,
  sort_order ASC,
  created_at ASC
`, videoID)
	if err != nil {
		return nil, fmt.Errorf("list video subtitles: %w", err)
	}
	defer rows.Close()

	out := make([]models.VideoSubtitle, 0, 8)
	for rows.Next() {
		var subtitle models.VideoSubtitle
		if err := rows.Scan(
			&subtitle.ID,
			&subtitle.VideoID,
			&subtitle.SourceType,
			&subtitle.Status,
			&subtitle.LanguageCode,
			&subtitle.Label,
			&subtitle.Format,
			&subtitle.MIMEType,
			&subtitle.StoredPath,
			&subtitle.FileSize,
			&subtitle.IsDefault,
			&subtitle.SortOrder,
			&subtitle.Metadata,
			&subtitle.CreatedAt,
			&subtitle.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan video subtitle: %w", err)
		}
		out = append(out, subtitle)
	}
	return out, rows.Err()
}

func (r *VideoRepository) ListVideoSubtitlesByVideoIDs(ctx context.Context, videoIDs []uuid.UUID) (map[uuid.UUID][]models.VideoSubtitle, error) {
	if len(videoIDs) == 0 {
		return map[uuid.UUID][]models.VideoSubtitle{}, nil
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, video_id, source_type, status, language_code, label, format, mime_type, stored_path, file_size, is_default, sort_order, metadata, created_at, updated_at
FROM video_subtitles
WHERE video_id = ANY($1::uuid[]) AND status = 'ready'
`, videoIDs)
	if err != nil {
		return nil, fmt.Errorf("list video subtitles by video ids: %w", err)
	}
	defer rows.Close()

	out := make(map[uuid.UUID][]models.VideoSubtitle, len(videoIDs))
	for rows.Next() {
		var subtitle models.VideoSubtitle
		if err := rows.Scan(
			&subtitle.ID,
			&subtitle.VideoID,
			&subtitle.SourceType,
			&subtitle.Status,
			&subtitle.LanguageCode,
			&subtitle.Label,
			&subtitle.Format,
			&subtitle.MIMEType,
			&subtitle.StoredPath,
			&subtitle.FileSize,
			&subtitle.IsDefault,
			&subtitle.SortOrder,
			&subtitle.Metadata,
			&subtitle.CreatedAt,
			&subtitle.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan video subtitle by ids: %w", err)
		}
		out[subtitle.VideoID] = append(out[subtitle.VideoID], subtitle)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for videoID := range out {
		sortVideoSubtitles(out[videoID])
	}
	return out, nil
}

func (r *VideoRepository) GetVideoSubtitle(ctx context.Context, subtitleID uuid.UUID) (models.VideoSubtitle, error) {
	var subtitle models.VideoSubtitle
	err := r.pool.QueryRow(ctx, `
SELECT id, video_id, source_type, status, language_code, label, format, mime_type, stored_path, file_size, is_default, sort_order, metadata, created_at, updated_at
FROM video_subtitles
WHERE id = $1
`, subtitleID).Scan(
		&subtitle.ID,
		&subtitle.VideoID,
		&subtitle.SourceType,
		&subtitle.Status,
		&subtitle.LanguageCode,
		&subtitle.Label,
		&subtitle.Format,
		&subtitle.MIMEType,
		&subtitle.StoredPath,
		&subtitle.FileSize,
		&subtitle.IsDefault,
		&subtitle.SortOrder,
		&subtitle.Metadata,
		&subtitle.CreatedAt,
		&subtitle.UpdatedAt,
	)
	if err != nil {
		return models.VideoSubtitle{}, fmt.Errorf("get video subtitle: %w", err)
	}
	return subtitle, nil
}

func (r *VideoRepository) DeleteVideoSubtitle(ctx context.Context, subtitleID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM video_subtitles WHERE id = $1`, subtitleID)
	if err != nil {
		return fmt.Errorf("delete video subtitle: %w", err)
	}
	return nil
}

func (r *VideoRepository) DeleteVideoSubtitlesBySourceType(ctx context.Context, videoID uuid.UUID, sourceType string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM video_subtitles WHERE video_id = $1 AND source_type = $2`, videoID, strings.TrimSpace(sourceType))
	if err != nil {
		return fmt.Errorf("delete video subtitles by source type: %w", err)
	}
	return nil
}

func (r *VideoRepository) ClearUploadedSubtitleDefaults(ctx context.Context, videoID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
UPDATE video_subtitles
SET is_default = FALSE, updated_at = NOW()
WHERE video_id = $1 AND source_type = 'uploaded' AND is_default = TRUE
`, videoID)
	if err != nil {
		return fmt.Errorf("clear uploaded subtitle defaults: %w", err)
	}
	return nil
}

func (r *VideoRepository) SetVideoSubtitleDefault(ctx context.Context, subtitleID uuid.UUID, isDefault bool) error {
	subtitle, err := r.GetVideoSubtitle(ctx, subtitleID)
	if err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin set subtitle default tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if isDefault && subtitle.SourceType == "uploaded" {
		if _, err := tx.Exec(ctx, `
UPDATE video_subtitles
SET is_default = FALSE, updated_at = NOW()
WHERE video_id = $1 AND source_type = 'uploaded' AND id <> $2
`, subtitle.VideoID, subtitle.ID); err != nil {
			return fmt.Errorf("clear previous uploaded subtitle default: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `
UPDATE video_subtitles
SET is_default = $2, updated_at = NOW()
WHERE id = $1
`, subtitleID, isDefault); err != nil {
		return fmt.Errorf("update subtitle default: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit set subtitle default tx: %w", err)
	}
	return nil
}

func (r *VideoRepository) UpdateVideoSubtitleMetadata(ctx context.Context, subtitleID uuid.UUID, languageCode, label string, sortOrder int) error {
	_, err := r.pool.Exec(ctx, `
UPDATE video_subtitles
SET language_code = $2, label = $3, sort_order = $4, updated_at = NOW()
WHERE id = $1
`, subtitleID, strings.TrimSpace(languageCode), strings.TrimSpace(label), sortOrder)
	if err != nil {
		return fmt.Errorf("update video subtitle metadata: %w", err)
	}
	return nil
}

func (r *VideoRepository) HasUploadedSubtitleDefault(ctx context.Context, videoID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
SELECT EXISTS(
  SELECT 1
  FROM video_subtitles
  WHERE video_id = $1 AND source_type = 'uploaded' AND is_default = TRUE
)
`, videoID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check uploaded subtitle default: %w", err)
	}
	return exists, nil
}

func BuildAppSubtitleTrack(subtitle models.VideoSubtitle) models.SubtitleTrack {
	embeddedIndex := 0
	if len(subtitle.Metadata) > 0 {
		var meta map[string]any
		if err := json.Unmarshal(subtitle.Metadata, &meta); err == nil {
			switch value := meta["embedded_index"].(type) {
			case float64:
				embeddedIndex = int(value)
			case int:
				embeddedIndex = value
			}
		}
	}
	languageLabel := strings.TrimSpace(subtitle.LanguageCode)
	if languageLabel == "" {
		languageLabel = "未标注"
	}
	return models.SubtitleTrack{
		ID:            subtitle.ID,
		SourceType:    subtitle.SourceType,
		LanguageCode:  subtitle.LanguageCode,
		LanguageLabel: languageLabel,
		Label:         subtitleLabel(subtitle),
		Format:        subtitle.Format,
		URL:           utils.VideoSubtitleURL(subtitle.VideoID, subtitle.ID),
		MIMEType:      subtitle.MIMEType,
		IsDefault:     subtitle.IsDefault,
		IsEmbedded:    subtitle.SourceType == "embedded",
		EmbeddedIndex: embeddedIndex,
		Available:     subtitle.Status == "ready" && strings.TrimSpace(subtitle.StoredPath) != "",
	}
}

func subtitleLabel(subtitle models.VideoSubtitle) string {
	label := strings.TrimSpace(subtitle.Label)
	if label != "" {
		return label
	}
	if strings.TrimSpace(subtitle.LanguageCode) != "" {
		if subtitle.SourceType == "embedded" {
			return subtitle.LanguageCode + "（内嵌）"
		}
		return subtitle.LanguageCode
	}
	if subtitle.SourceType == "embedded" {
		return "内嵌字幕"
	}
	return "外挂字幕"
}

func sortVideoSubtitles(items []models.VideoSubtitle) {
	sort.SliceStable(items, func(i, j int) bool {
		rank := func(item models.VideoSubtitle) int {
			switch {
			case item.SourceType == "uploaded" && item.IsDefault:
				return 0
			case item.SourceType == "uploaded":
				return 1
			case item.IsDefault:
				return 2
			default:
				return 3
			}
		}
		ri := rank(items[i])
		rj := rank(items[j])
		if ri != rj {
			return ri < rj
		}
		if items[i].SortOrder != items[j].SortOrder {
			return items[i].SortOrder < items[j].SortOrder
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
}
