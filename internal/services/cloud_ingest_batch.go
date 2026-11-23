package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"video-server/internal/models"
)

// maxInlineAutoImports bounds the ffprobe work an explicit request may do for
// already-downloaded members; the repair scan finishes the rest.
const maxInlineAutoImports = 5

// CloudIngestBatchTitleMaxRunes caps one batch or per-link title. Titles are
// rendered in admin tables and video lists, so a runaway filename must not
// become a runaway heading.
const CloudIngestBatchTitleMaxRunes = 200

// normalizeCloudIngestBatchInput validates one explicit 批量入库 request and
// produces the batch preset row plus the confirmed links in display order.
func normalizeCloudIngestBatchInput(input models.CloudIngestBatchInput) (models.CloudIngestBatch, []models.CloudIngestBatchLink, error) {
	batch := models.CloudIngestBatch{
		TitleMode:         strings.ToLower(strings.TrimSpace(input.TitleMode)),
		UniformTitle:      strings.TrimSpace(input.UniformTitle),
		VideoType:         strings.ToLower(strings.TrimSpace(input.VideoType)),
		Description:       strings.TrimSpace(input.Description),
		Tags:              normalizeIngestTextList(input.Tags),
		ActorNames:        normalizeIngestTextList(input.ActorNames),
		CollectionIDs:     nonNilUUIDs(input.CollectionIDs),
		ActorIDs:          nonNilUUIDs(input.ActorIDs),
		ImageCollectionID: input.ImageCollectionID,
	}
	if batch.TitleMode == "" {
		batch.TitleMode = models.CloudIngestTitleUniform
	}
	switch batch.TitleMode {
	case models.CloudIngestTitleUniform:
		if batch.UniformTitle == "" {
			return models.CloudIngestBatch{}, nil, fmt.Errorf("%w: 统一标题不能为空", models.ErrCloudIngestBatchInvalidInput)
		}
	case models.CloudIngestTitleFilename:
		// Per-link titles travel with each link; the batch title is unused.
		batch.UniformTitle = ""
	default:
		return models.CloudIngestBatch{}, nil, fmt.Errorf("%w: title_mode 只能是 uniform 或 filename", models.ErrCloudIngestBatchInvalidInput)
	}
	if utf8.RuneCountInString(batch.UniformTitle) > CloudIngestBatchTitleMaxRunes {
		return models.CloudIngestBatch{}, nil, fmt.Errorf("%w: 统一标题最多 %d 字", models.ErrCloudIngestBatchInvalidInput, CloudIngestBatchTitleMaxRunes)
	}

	if batch.VideoType == "" {
		batch.VideoType = "short"
	}
	switch batch.VideoType {
	case "short", "movie", "episode", "av":
	default:
		return models.CloudIngestBatch{}, nil, fmt.Errorf("%w: 视频 type 只能是 short/movie/episode/av", models.ErrCloudIngestBatchInvalidInput)
	}
	// 视频合集仍只属于短视频：非短视频携带合集会让命中已有视频时静默写入。
	if batch.VideoType != "short" && len(batch.CollectionIDs) > 0 {
		return models.CloudIngestBatch{}, nil, fmt.Errorf("%w: 仅短视频支持合集", models.ErrCloudIngestBatchInvalidInput)
	}
	if batch.Tags == nil {
		batch.Tags = []string{}
	}
	if batch.ActorNames == nil {
		batch.ActorNames = []string{}
	}
	if batch.VideoType == "av" {
		batch.SiteCategory = normalizeAVSiteCategory(input.SiteCategory)
		if batch.SiteCategory == "" {
			batch.SiteCategory = avSiteCategoryJapanese
		}
	} else {
		batch.SiteCategory = ""
	}

	if len(input.Links) == 0 {
		return models.CloudIngestBatch{}, nil, fmt.Errorf("%w: links 不能为空", models.ErrCloudIngestBatchInvalidInput)
	}
	if len(input.Links) > models.CloudDownloadBatchLimit {
		return models.CloudIngestBatch{}, nil, fmt.Errorf("%w: 单次最多 %d 条链接", models.ErrCloudIngestBatchInvalidInput, models.CloudDownloadBatchLimit)
	}

	links := make([]models.CloudIngestBatchLink, 0, len(input.Links))
	for index, link := range input.Links {
		title := strings.TrimSpace(link.Title)
		if utf8.RuneCountInString(title) > CloudIngestBatchTitleMaxRunes {
			return models.CloudIngestBatch{}, nil, fmt.Errorf("%w: 第 %d 条标题最多 %d 字", models.ErrCloudIngestBatchInvalidInput, index+1, CloudIngestBatchTitleMaxRunes)
		}
		if link.Forced {
			batch.ForcedCount++
		}
		links = append(links, models.CloudIngestBatchLink{
			Link:   strings.TrimSpace(link.Link),
			Title:  title,
			Forced: link.Forced,
		})
	}
	return batch, links, nil
}

// normalizeIngestTextList trims and drops empties. The result is never nil:
// the batch table declares these columns NOT NULL, and pgx encodes a nil slice
// as SQL NULL (an empty slice as '{}').
func normalizeIngestTextList(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func nonNilUUIDs(values []uuid.UUID) []uuid.UUID {
	if values == nil {
		return []uuid.UUID{}
	}
	return values
}

// SubmitIngestBatch is the explicit 批量入库 action: it creates one batch
// preset, submits every confirmed link as an independent task (same conflict
// handling as SubmitBatch), binds each task to the batch, and lets completed
// members import right away. It never waits for 115.
func (s *CloudDownloadService) SubmitIngestBatch(ctx context.Context, input models.CloudIngestBatchInput, createdBy *uuid.UUID) (models.CloudIngestBatch, models.CloudDownloadBatchResult, error) {
	batch, links, err := normalizeCloudIngestBatchInput(input)
	if err != nil {
		return models.CloudIngestBatch{}, models.CloudDownloadBatchResult{}, err
	}
	if s.batchRepo == nil {
		return models.CloudIngestBatch{}, models.CloudDownloadBatchResult{}, errors.New("入库批未配置")
	}
	rawLinks := make([]string, len(links))
	titles := make([]string, len(links))
	forced := make([]bool, len(links))
	for index, link := range links {
		rawLinks[index] = link.Link
		titles[index] = link.Title
		forced[index] = link.Forced
	}

	// 全部链接都无效时不落批行：批不可删除，空批只会变成永远清不掉的孤儿。
	probe := models.CloudDownloadBatchResult{Items: make([]models.CloudDownloadBatchItem, len(rawLinks))}
	if len(parseCloudBatchCandidates(rawLinks, &probe)) == 0 {
		probe.Summary = summarizeCloudDownloadBatch(probe.Items)
		return models.CloudIngestBatch{}, probe, nil
	}

	batch.CreatedBy = createdBy
	created, err := s.batchRepo.Create(ctx, batch)
	if err != nil {
		return models.CloudIngestBatch{}, models.CloudDownloadBatchResult{}, err
	}

	// 已下载过的成员会在请求内立刻导入；只做前几个，其余交给修复扫描，
	// 避免一次请求里串行跑 50 次 ffprobe。
	inlineImports := 0
	result, submitErr := s.submitBatch(ctx, rawLinks, createdBy, func(ctx context.Context, index int, taskID int64, hasTask bool, item *models.CloudDownloadBatchItem) {
		item.Forced = forced[index]
		if !hasTask || taskID == 0 {
			return
		}
		attached, attachErr := s.batchRepo.AttachTask(ctx, taskID, created.ID, titles[index])
		if attachErr != nil {
			item.Message = joinIngestMessage(item.Message, "入库批归属失败: "+attachErr.Error())
			return
		}
		if !attached {
			item.Message = joinIngestMessage(item.Message, "该链接已属于其它入库批，本次未改归属")
			return
		}
		if inlineImports < maxInlineAutoImports {
			inlineImports++
			s.autoImportBatchMember(ctx, taskID)
		}
	})
	if submitErr != nil {
		// Candidates resolve before the loop, so no member can be attached yet:
		// dropping the batch keeps the "one explicit action = one batch" rule
		// without leaving an unusable orphan.
		if dropErr := s.batchRepo.Delete(ctx, created.ID); dropErr != nil {
			s.logWarn("drop ingest batch after failure", "batch_id", created.ID, "error", dropErr)
		}
		return models.CloudIngestBatch{}, result, submitErr
	}
	return created, result, nil
}

func joinIngestMessage(existing, extra string) string {
	if strings.TrimSpace(existing) == "" {
		return extra
	}
	return existing + "；" + extra
}

// autoImportBatchMember runs the batch preset for one completed member. It is
// safe to call repeatedly: the import record's task_id is unique and a
// terminal import returns the existing record.
func (s *CloudDownloadService) autoImportBatchMember(ctx context.Context, taskID int64) {
	if s.batchImporter == nil {
		return
	}
	importCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), autoImportTimeout)
	defer cancel()
	if err := s.batchImporter.AutoImportFromBatch(importCtx, taskID); err != nil {
		s.logWarn("auto import batch member failed", "task_id", taskID, "error", err)
	}
}

// reconcileBatchIngest repairs lost auto-import wakeups: a member that already
// finished downloading but never got an import record is retried here.
func (s *CloudDownloadService) reconcileBatchIngest(ctx context.Context) {
	if s.batchRepo == nil || s.batchImporter == nil {
		return
	}
	taskIDs, err := s.batchRepo.ListAwaitingAutoImport(ctx, 10)
	if err != nil {
		s.logWarn("reconcile batch ingest: list failed", "error", err)
		return
	}
	for _, taskID := range taskIDs {
		s.autoImportBatchMember(ctx, taskID)
	}
	if len(taskIDs) > 0 {
		s.logInfo("batch ingest reconciled", "awaiting_import", len(taskIDs))
	}
}
