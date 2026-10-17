package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/hashutil"
	"video-server/internal/models"
	"video-server/pkg/ffmpeg"
)

// cloudDownloadImportRepository is the persistence surface used by the import
// service. *repository.CloudDownloadImportRepository implements it.
type cloudDownloadImportRepository interface {
	CreateOrGet(ctx context.Context, item models.CloudDownloadImport) (models.CloudDownloadImport, bool, error)
	GetByID(ctx context.Context, id int64) (models.CloudDownloadImport, error)
	GetByTaskID(ctx context.Context, taskID int64) (models.CloudDownloadImport, error)
	UpdateProgress(ctx context.Context, id int64, status, errorMessage string) error
	SetHash(ctx context.Context, id int64, sha256 string) error
	SetVideo(ctx context.Context, id int64, videoID uuid.UUID) error
	SetArchiveBatch(ctx context.Context, id int64, batchID uuid.UUID) error
	Finish(ctx context.Context, id int64, status string, videoID, archiveBatchID *uuid.UUID, importedAt time.Time) error
	ResetForRetry(ctx context.Context, id int64) (models.CloudDownloadImport, bool, error)
	ListSummariesByTaskIDs(ctx context.Context, taskIDs []int64) (map[int64]models.CloudDownloadImportSummary, error)
}

type cloudImportTaskRepository interface {
	GetByID(ctx context.Context, id int64) (models.CloudDownloadTask, error)
}

type cloudImportVideoLookup interface {
	FindVideoByHash(ctx context.Context, hash string, fileSize int64) (uuid.UUID, bool, error)
	GetVideoByID(ctx context.Context, videoID uuid.UUID) (models.Video, error)
}

type cloudImportUploadService interface {
	PreviewUploadedFile(ctx context.Context, in LocalUploadInput) (UploadPreview, error)
	SaveImportedFile(ctx context.Context, in LocalUploadInput, maxVideoSize int64) (UploadResult, error)
}

type cloudImportArchiveService interface {
	ImportLocalArchive(ctx context.Context, in ArchiveImportUploadInput, sourcePath string) (models.ArchiveImportBatch, error)
	GetBatch(ctx context.Context, batchID uuid.UUID) (models.ArchiveImportBatch, error)
	// RetryExtract re-enqueues an existing batch instead of creating a new one.
	RetryExtract(ctx context.Context, batchID uuid.UUID, password, requestedMode string) (models.ArchiveImportBatch, error)
}

type cloudImportEnqueuer interface {
	EnqueueCloudIngest(importID int64, attempt int, input models.CloudDownloadImportInput) error
	EnqueueTranscode(videoID, inputPath, outputDir, targetFormat string, force bool) error
	EnqueueScrapeMovie(videoID, filePath, filename string) error
	EnqueueScrapeTV(videoID, filePath, filename string) error
	EnqueueScrapeAV(videoID, filePath, filename string) error
}

// CloudDownloadImportService explicitly moves one completed cloud download
// into the video or archive import flow. It never runs as a download side
// effect: every call is triggered by an admin preview + import request.
type CloudDownloadImportService struct {
	taskRepo      cloudImportTaskRepository
	repo          cloudDownloadImportRepository
	videos        cloudImportVideoLookup
	upload        cloudImportUploadService
	archive       cloudImportArchiveService
	enqueuer      cloudImportEnqueuer
	downloadDir   string
	storageRoot   string
	logger        *slog.Logger
	probePlayback func(context.Context, string) (ffmpeg.PlaybackCompatibilityProbe, error)
}

// NewCloudDownloadImportService builds the service.
func NewCloudDownloadImportService(
	taskRepo cloudImportTaskRepository,
	repo cloudDownloadImportRepository,
	videos cloudImportVideoLookup,
	upload cloudImportUploadService,
	archive cloudImportArchiveService,
	enqueuer cloudImportEnqueuer,
	downloadDir, storageRoot string,
	logger *slog.Logger,
) *CloudDownloadImportService {
	return &CloudDownloadImportService{
		taskRepo:      taskRepo,
		repo:          repo,
		videos:        videos,
		upload:        upload,
		archive:       archive,
		enqueuer:      enqueuer,
		downloadDir:   downloadDir,
		storageRoot:   storageRoot,
		logger:        logger,
		probePlayback: ffmpeg.ProbePlaybackCompatibility,
	}
}

// Preview inspects the actual downloaded file. It does not create anything.
func (s *CloudDownloadImportService) Preview(ctx context.Context, taskID int64) (models.CloudDownloadImportPreview, error) {
	_, sourcePath, info, err := s.loadCompletedSource(ctx, taskID)
	if err != nil {
		return models.CloudDownloadImportPreview{}, err
	}
	name := filepath.Base(sourcePath)
	kind, mime, err := s.classifyCloudImportFile(ctx, sourcePath)
	if err != nil {
		return models.CloudDownloadImportPreview{}, err
	}
	return models.CloudDownloadImportPreview{
		Kind: kind,
		Name: name,
		Size: info.Size(),
		Mime: mime,
	}, nil
}

// classifyCloudImportFile decides the actual kind from file content, not from
// the (possibly stale) declared name. An archive requires both a supported
// extension and an archive MIME; everything else must prove a video stream.
func (s *CloudDownloadImportService) classifyCloudImportFile(ctx context.Context, sourcePath string) (string, string, error) {
	name := filepath.Base(sourcePath)
	mime := detectCloudImportMIME(sourcePath)
	if isSupportedArchiveFile(name, mime) {
		return models.CloudImportKindArchive, archivePreviewMIME(name, mime), nil
	}
	probe, probeErr := s.probePlayback(ctx, sourcePath)
	if probeErr != nil {
		return "", mime, fmt.Errorf("%w: %v", models.ErrCloudImportUnsupported, probeErr)
	}
	if !probe.VideoStreamFound {
		return "", mime, fmt.Errorf("%w: 未探测到视频流", models.ErrCloudImportUnsupported)
	}
	return models.CloudImportKindVideo, mime, nil
}

// Create records the one-shot import and enqueues cloud:ingest. The import
// form travels in the queue payload; the archive password is never persisted.
func (s *CloudDownloadImportService) Create(ctx context.Context, taskID int64, input models.CloudDownloadImportInput, createdBy *uuid.UUID) (models.CloudDownloadImport, error) {
	task, err := s.loadCompletedTask(ctx, taskID)
	if err != nil {
		return models.CloudDownloadImport{}, err
	}

	existing, getErr := s.repo.GetByTaskID(ctx, taskID)
	if getErr == nil {
		requestedKind := strings.ToLower(strings.TrimSpace(input.Kind))
		if requestedKind != "" && requestedKind != strings.TrimSpace(existing.Kind) {
			return existing, fmt.Errorf("%w: 已存在 %s 类型导入，不能改为 %s", models.ErrCloudImportInvalidInput, existing.Kind, requestedKind)
		}
		// Terminal imports return immediately: the download file may already be
		// gone after success. Active records are allowed through so an explicit
		// re-submit can repair a lost queue task with the same TaskID (attempt),
		// which asynq deduplicates without creating a second worker job.
		if models.CloudImportTerminal(existing.Status) {
			return existing, nil
		}
	} else if !errors.Is(getErr, pgx.ErrNoRows) {
		return models.CloudDownloadImport{}, getErr
	}

	sourcePath, info, err := resolveCloudImportSource(s.downloadDir, task.LocalPath)
	if err != nil {
		return models.CloudDownloadImport{}, err
	}
	normalized, err := normalizeCloudImportInput(input, sourcePath)
	if err != nil {
		return models.CloudDownloadImport{}, err
	}
	actualKind, _, err := s.classifyCloudImportFile(ctx, sourcePath)
	if err != nil {
		return models.CloudDownloadImport{}, err
	}
	if actualKind != normalized.Kind {
		return models.CloudDownloadImport{}, fmt.Errorf("%w: 文件实际类型为 %s，与请求的 %s 不一致", models.ErrCloudImportInvalidInput, actualKind, normalized.Kind)
	}

	if getErr == nil {
		return s.createForExisting(ctx, existing, normalized)
	}

	record, created, err := s.repo.CreateOrGet(ctx, models.CloudDownloadImport{
		TaskID:     taskID,
		Kind:       normalized.Kind,
		Status:     models.CloudImportQueued,
		SourcePath: sourcePath,
		SourceSize: info.Size(),
		CreatedBy:  createdBy,
	})
	if err != nil {
		return models.CloudDownloadImport{}, err
	}
	if !created {
		return s.createForExisting(ctx, record, normalized)
	}
	return s.enqueueImport(ctx, record, normalized)
}

func (s *CloudDownloadImportService) createForExisting(ctx context.Context, existing models.CloudDownloadImport, input models.CloudDownloadImportInput) (models.CloudDownloadImport, error) {
	if strings.TrimSpace(existing.Kind) != input.Kind {
		return existing, fmt.Errorf("%w: 已存在 %s 类型导入，不能改为 %s", models.ErrCloudImportInvalidInput, existing.Kind, input.Kind)
	}
	switch existing.Status {
	case models.CloudImportImported, models.CloudImportSkippedExisting:
		return existing, nil
	case models.CloudImportQueued, models.CloudImportHashing, models.CloudImportCopying, models.CloudImportCreating:
		// Same attempt, same TaskID: if the task is still queued this is a
		// no-op; if it was lost, this repairs it without duplicating the record.
		return s.enqueueImport(ctx, existing, input)
	case models.CloudImportFailed:
		updated, won, err := s.repo.ResetForRetry(ctx, existing.ID)
		if err != nil {
			return models.CloudDownloadImport{}, err
		}
		if !won {
			return s.repo.GetByID(ctx, existing.ID)
		}
		return s.enqueueImport(ctx, updated, input)
	default:
		return existing, nil
	}
}

func (s *CloudDownloadImportService) enqueueImport(ctx context.Context, record models.CloudDownloadImport, input models.CloudDownloadImportInput) (models.CloudDownloadImport, error) {
	if s.enqueuer == nil {
		message := "云导入队列未配置"
		_ = s.repo.UpdateProgress(ctx, record.ID, models.CloudImportFailed, message)
		record.Status = models.CloudImportFailed
		record.ErrorMessage = message
		return record, errors.New(message)
	}
	if err := s.enqueuer.EnqueueCloudIngest(record.ID, record.Attempt, input); err != nil {
		message := truncateMessage(fmt.Sprintf("入队失败: %v", err), 800)
		_ = s.repo.UpdateProgress(ctx, record.ID, models.CloudImportFailed, message)
		record.Status = models.CloudImportFailed
		record.ErrorMessage = message
		return record, fmt.Errorf("enqueue cloud ingest: %w", err)
	}
	return record, nil
}

// Process advances one queued (or retried) import. It is idempotent: a
// redelivered job reuses an already created video target instead of copying or
// creating it twice.
func (s *CloudDownloadImportService) Process(ctx context.Context, importID int64, attempt int, input models.CloudDownloadImportInput) error {
	record, err := s.repo.GetByID(ctx, importID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // stale job for a deleted record
		}
		return err
	}
	if record.Attempt != attempt {
		// A retry has already advanced this record; never let an old queued
		// attempt overwrite its state or create a duplicate target.
		return nil
	}
	if models.CloudImportTerminal(record.Status) {
		return nil
	}
	if record.Status != models.CloudImportQueued {
		if err := s.repo.UpdateProgress(ctx, record.ID, models.CloudImportQueued, ""); err != nil {
			return err
		}
		record.Status = models.CloudImportQueued
	}

	task, err := s.taskRepo.GetByID(ctx, record.TaskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s.failImport(ctx, record.ID, "云下载任务不存在")
		}
		return err
	}
	if task.Status != models.CloudDownloadCompleted {
		return s.failImport(ctx, record.ID, fmt.Sprintf("云下载任务未完成（%s）", task.Status))
	}
	sourcePath, sourceInfo, err := resolveCloudImportSource(s.downloadDir, task.LocalPath)
	if err != nil {
		return s.failImport(ctx, record.ID, fmt.Sprintf("下载文件不可用: %v", err))
	}
	actualKind, _, kindErr := s.classifyCloudImportFile(ctx, sourcePath)
	if kindErr != nil {
		return s.failImport(ctx, record.ID, fmt.Sprintf("文件类型不受支持: %v", kindErr))
	}
	if actualKind != record.Kind {
		return s.failImport(ctx, record.ID, fmt.Sprintf("文件实际类型为 %s，与导入记录 %s 不一致", actualKind, record.Kind))
	}

	switch record.Kind {
	case models.CloudImportKindVideo:
		return s.processVideoImport(ctx, record, task, sourcePath, sourceInfo, input)
	case models.CloudImportKindArchive:
		return s.processArchiveImport(ctx, record, input, sourcePath)
	default:
		return s.failImport(ctx, record.ID, fmt.Sprintf("未知导入类型: %s", record.Kind))
	}
}

func (s *CloudDownloadImportService) processVideoImport(
	ctx context.Context,
	record models.CloudDownloadImport,
	task models.CloudDownloadTask,
	sourcePath string,
	sourceInfo os.FileInfo,
	input models.CloudDownloadImportInput,
) error {
	videoType, err := normalizeCloudImportVideoType(input.Type)
	if err != nil {
		return s.failImport(ctx, record.ID, err.Error())
	}
	if s.upload == nil || s.videos == nil {
		return s.failImport(ctx, record.ID, "视频导入服务未配置")
	}

	if err := s.repo.UpdateProgress(ctx, record.ID, models.CloudImportHashing, ""); err != nil {
		return err
	}
	hash, err := hashutil.SHA256(sourcePath)
	if err != nil {
		return s.failImport(ctx, record.ID, fmt.Sprintf("计算文件哈希失败: %v", err))
	}
	if err := s.repo.SetHash(ctx, record.ID, hash); err != nil {
		return err
	}
	size := sourceInfo.Size()

	videoID, found, err := s.videos.FindVideoByHash(ctx, hash, size)
	if err != nil {
		return s.failImport(ctx, record.ID, fmt.Sprintf("查询已导入视频失败: %v", err))
	}
	resuming := record.VideoID != nil && found && *record.VideoID == videoID
	if found && !resuming {
		// SaveImportedFile creates the file_hashes row before the import record
		// links video_id. If the process died between those two writes, the
		// next run sees a hash hit with video_id NULL. Only treat it as an
		// unrelated existing video when its metadata belongs to another task.
		video, getErr := s.videos.GetVideoByID(ctx, videoID)
		if getErr != nil {
			return s.failImport(ctx, record.ID, fmt.Sprintf("读取已导入视频失败: %v", getErr))
		}
		if cloudImportVideoBelongsToTask(video, task.ID) {
			record.VideoID = &videoID
			return s.resumeVideoImport(ctx, record, videoID, sourcePath)
		}
		return s.finishSkippedExisting(ctx, record, videoID, sourcePath)
	}
	if resuming {
		return s.resumeVideoImport(ctx, record, videoID, sourcePath)
	}

	targetID := uuid.Nil
	if record.VideoID != nil {
		targetID = *record.VideoID
	} else if found {
		targetID = videoID
	}
	if targetID == uuid.Nil {
		preview, previewErr := s.upload.PreviewUploadedFile(ctx, LocalUploadInput{
			UserID:            cloudImportUserID(record.CreatedBy),
			FilePath:          sourcePath,
			Filename:          filepath.Base(sourcePath),
			FileSize:          size,
			Title:             strings.TrimSpace(input.Title),
			Desc:              strings.TrimSpace(input.Description),
			Type:              videoType,
			Tags:              input.Tags,
			ActorIDs:          input.ActorIDs,
			ActorNames:        input.ActorNames,
			CollectionIDs:     input.CollectionIDs,
			ImageCollectionID: input.ImageCollectionID,
			SiteCategory:      strings.TrimSpace(input.SiteCategory),
			Hash:              hash,
			HashVerified:      true,
		})
		if previewErr != nil {
			return s.failImport(ctx, record.ID, fmt.Sprintf("预览视频导入失败: %v", previewErr))
		}
		if preview.AlreadyExists {
			return s.finishSkippedExisting(ctx, record, preview.VideoID, sourcePath)
		}
		targetID = preview.VideoID
	}

	targetPath := cloudImportVideoSourcePath(s.storageRoot, targetID, sourcePath)
	targetExists := cloudImportTargetExists(targetPath)
	copied := false
	if !targetExists {
		if err := s.repo.UpdateProgress(ctx, record.ID, models.CloudImportCopying, ""); err != nil {
			return err
		}
		if err := copyCloudImportFile(targetPath, sourcePath); err != nil {
			return s.failImport(ctx, record.ID, fmt.Sprintf("复制视频源文件失败: %v", err))
		}
		copied = true
	}

	if record.VideoID == nil || !found {
		if err := s.repo.UpdateProgress(ctx, record.ID, models.CloudImportCreating, ""); err != nil {
			return err
		}
		result, saveErr := s.upload.SaveImportedFile(ctx, LocalUploadInput{
			UserID:            cloudImportUserID(record.CreatedBy),
			FilePath:          targetPath,
			Filename:          filepath.Base(sourcePath),
			FileSize:          size,
			Title:             strings.TrimSpace(input.Title),
			Desc:              strings.TrimSpace(input.Description),
			Type:              videoType,
			Tags:              input.Tags,
			ActorIDs:          input.ActorIDs,
			ActorNames:        input.ActorNames,
			CollectionIDs:     input.CollectionIDs,
			ImageCollectionID: input.ImageCollectionID,
			SiteCategory:      strings.TrimSpace(input.SiteCategory),
			Hash:              hash,
			HashVerified:      true,
			VideoID:           targetID,
			Metadata: map[string]any{
				"source":                        "cloud_download",
				"cloud_task_id":                 task.ID,
				"info_hash":                     task.InfoHash,
				"delete_source_after_transcode": true,
			},
		}, 0)
		if saveErr != nil {
			if copied {
				_ = os.Remove(targetPath)
			}
			return s.failImport(ctx, record.ID, fmt.Sprintf("创建视频记录失败: %v", saveErr))
		}
		if result.AlreadyExists && result.VideoID != targetID {
			if copied {
				_ = os.Remove(targetPath)
			}
			return s.finishSkippedExisting(ctx, record, result.VideoID, sourcePath)
		}
		if result.VideoID != uuid.Nil {
			targetID = result.VideoID
			if err := s.repo.SetVideo(ctx, record.ID, targetID); err != nil {
				return err
			}
		}
	}

	if err := s.enqueueImportedVideo(targetID, targetPath, filepath.Base(sourcePath), videoType); err != nil {
		return s.failImport(ctx, record.ID, fmt.Sprintf("视频入队失败: %v", err))
	}
	if err := s.finish(ctx, record.ID, models.CloudImportImported, &targetID, nil); err != nil {
		return err
	}
	s.removeDownloadedSource(sourcePath)
	return nil
}

// resumeVideoImport handles a retry whose video row already exists. A ready
// video, or a missing source-original, means the target was already produced:
// finish without copying or enqueueing again. Otherwise only the queue step is
// retried, keeping the download file until the enqueue succeeds.
func (s *CloudDownloadImportService) resumeVideoImport(
	ctx context.Context,
	record models.CloudDownloadImport,
	videoID uuid.UUID,
	sourcePath string,
) error {
	video, err := s.videos.GetVideoByID(ctx, videoID)
	if err != nil {
		return s.failImport(ctx, record.ID, fmt.Sprintf("读取已创建视频失败: %v", err))
	}
	targetPath := cloudImportVideoSourcePath(s.storageRoot, videoID, sourcePath)
	targetExists := cloudImportTargetExists(targetPath)
	videoStatus := strings.ToLower(strings.TrimSpace(video.Status))
	if !targetExists && videoStatus != "ready" && videoStatus != "processing" {
		return s.failImport(ctx, record.ID, fmt.Sprintf("导入源文件缺失，视频状态为 %s，请重新导入", video.Status))
	}
	if videoStatus == "ready" {
		if targetExists {
			// The ready video already has its own playback output; the copied
			// source-original is now an orphan, but never the playback path.
			if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
				if s.logger != nil {
					s.logger.Warn("remove orphan cloud import source failed", "video_id", videoID, "path", targetPath, "error", err)
				}
			}
		}
		if err := s.finish(ctx, record.ID, models.CloudImportImported, &videoID, nil); err != nil {
			return err
		}
		s.removeDownloadedSource(sourcePath)
		return nil
	}
	if !targetExists {
		// processing without source-original means the file was already moved
		// into the transcode pipeline; do not enqueue again.
		if err := s.finish(ctx, record.ID, models.CloudImportImported, &videoID, nil); err != nil {
			return err
		}
		s.removeDownloadedSource(sourcePath)
		return nil
	}
	// Use the persisted video type, not the retry form: the row is the source
	// of truth for which queue already owns this target.
	videoType, typeErr := normalizeCloudImportVideoType(video.Type)
	if typeErr != nil {
		return s.failImport(ctx, record.ID, fmt.Sprintf("已创建视频的类型无效: %v", typeErr))
	}
	if err := s.enqueueImportedVideo(videoID, targetPath, filepath.Base(sourcePath), videoType); err != nil {
		return s.failImport(ctx, record.ID, fmt.Sprintf("视频入队失败: %v", err))
	}
	if err := s.finish(ctx, record.ID, models.CloudImportImported, &videoID, nil); err != nil {
		return err
	}
	s.removeDownloadedSource(sourcePath)
	return nil
}

func (s *CloudDownloadImportService) processArchiveImport(ctx context.Context, record models.CloudDownloadImport, input models.CloudDownloadImportInput, sourcePath string) error {
	if s.archive == nil {
		return s.failImport(ctx, record.ID, "压缩包导入服务未配置")
	}
	if input.HasPassword && strings.TrimSpace(input.Password) == "" {
		return s.failImport(ctx, record.ID, "压缩包密码不能为空")
	}
	var batch models.ArchiveImportBatch
	var err error
	if record.ArchiveBatchID != nil {
		// A previous attempt already copied the archive into its batch
		// directory. If extraction was already enqueued or has started, do not
		// clear it; otherwise only the enqueue step needs to be retried.
		current, getErr := s.archive.GetBatch(ctx, *record.ArchiveBatchID)
		if getErr != nil {
			return s.failImport(ctx, record.ID, fmt.Sprintf("读取压缩包批次失败: %v", getErr))
		}
		if archiveBatchAlreadyQueued(current.Status) {
			batch = current
		} else {
			batch, err = s.archive.RetryExtract(ctx, *record.ArchiveBatchID, strings.TrimSpace(input.Password), "")
		}
	} else {
		batch, err = s.archive.ImportLocalArchive(ctx, ArchiveImportUploadInput{
			UserID:                    cloudImportUserID(record.CreatedBy),
			Title:                     strings.TrimSpace(input.Title),
			DefaultDescription:        strings.TrimSpace(input.DefaultDescription),
			DefaultTags:               input.DefaultTags,
			DefaultVideoCollectionIDs: input.DefaultVideoCollectionIDs,
			DefaultImageCollectionIDs: input.DefaultImageCollectionIDs,
			HasPassword:               input.HasPassword,
			Password:                  strings.TrimSpace(input.Password),
		}, sourcePath)
		if batch.ID != uuid.Nil {
			if setErr := s.repo.SetArchiveBatch(ctx, record.ID, batch.ID); setErr != nil {
				return s.failImport(ctx, record.ID, fmt.Sprintf("关联压缩包批次失败，批次 %s 已创建，请到工具箱处理: %v", batch.ID, setErr))
			}
		}
	}
	if err != nil {
		return s.failImport(ctx, record.ID, fmt.Sprintf("创建压缩包导入失败: %v", err))
	}
	batchID := batch.ID
	if err := s.finish(ctx, record.ID, models.CloudImportImported, nil, &batchID); err != nil {
		return err
	}
	s.removeDownloadedSource(sourcePath)
	return nil
}

func (s *CloudDownloadImportService) enqueueImportedVideo(videoID uuid.UUID, inputPath, filename, videoType string) error {
	if s.enqueuer == nil {
		return errors.New("云导入队列未配置")
	}
	switch videoType {
	case "short":
		return s.enqueuer.EnqueueTranscode(videoID.String(), inputPath, filepath.Join(s.storageRoot, "videos"), "mp4", false)
	case "movie":
		return s.enqueuer.EnqueueScrapeMovie(videoID.String(), inputPath, filename)
	case "episode":
		return s.enqueuer.EnqueueScrapeTV(videoID.String(), inputPath, filename)
	case "av":
		return s.enqueuer.EnqueueScrapeAV(videoID.String(), inputPath, filename)
	default:
		return fmt.Errorf("%w: %s", ErrInvalidType, videoType)
	}
}

func (s *CloudDownloadImportService) finishSkippedExisting(ctx context.Context, record models.CloudDownloadImport, videoID uuid.UUID, sourcePath string) error {
	if err := s.finish(ctx, record.ID, models.CloudImportSkippedExisting, &videoID, nil); err != nil {
		return err
	}
	s.removeDownloadedSource(sourcePath)
	return nil
}

func (s *CloudDownloadImportService) finish(ctx context.Context, id int64, status string, videoID, archiveBatchID *uuid.UUID) error {
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return s.repo.Finish(writeCtx, id, status, videoID, archiveBatchID, time.Now().UTC())
}

func (s *CloudDownloadImportService) failImport(ctx context.Context, id int64, message string) error {
	message = truncateMessage(strings.TrimSpace(message), 800)
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if err := s.repo.UpdateProgress(writeCtx, id, models.CloudImportFailed, message); err != nil {
		return err
	}
	return errors.New(message)
}

// MarkFailed records a terminal failure once asynq has exhausted retries.
func (s *CloudDownloadImportService) MarkFailed(ctx context.Context, importID int64, attempt int, message string) error {
	record, err := s.repo.GetByID(ctx, importID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if record.Attempt != attempt {
		// The final failure of an old attempt must not fail the newly
		// retried import.
		return nil
	}
	if models.CloudImportTerminal(record.Status) {
		return nil
	}
	return s.repo.UpdateProgress(ctx, importID, models.CloudImportFailed, truncateMessage(message, 800))
}

func (s *CloudDownloadImportService) loadCompletedTask(ctx context.Context, taskID int64) (models.CloudDownloadTask, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.CloudDownloadTask{}, fmt.Errorf("%w: %d", models.ErrCloudImportNotFound, taskID)
		}
		return models.CloudDownloadTask{}, err
	}
	if task.Status != models.CloudDownloadCompleted {
		return task, fmt.Errorf("%w: %s", models.ErrCloudImportNotCompleted, task.Status)
	}
	return task, nil
}

func (s *CloudDownloadImportService) loadCompletedSource(ctx context.Context, taskID int64) (models.CloudDownloadTask, string, os.FileInfo, error) {
	task, err := s.loadCompletedTask(ctx, taskID)
	if err != nil {
		return task, "", nil, err
	}
	sourcePath, info, err := resolveCloudImportSource(s.downloadDir, task.LocalPath)
	if err != nil {
		return task, "", nil, err
	}
	return task, sourcePath, info, nil
}

func (s *CloudDownloadImportService) removeDownloadedSource(sourcePath string) {
	if strings.TrimSpace(sourcePath) == "" {
		return
	}
	if err := os.Remove(sourcePath); err != nil && !os.IsNotExist(err) {
		if s.logger != nil {
			s.logger.Warn("remove imported cloud download source failed", "path", sourcePath, "error", err)
		}
	}
}

func resolveCloudImportSource(downloadDir, localPath string) (string, os.FileInfo, error) {
	localPath = strings.TrimSpace(localPath)
	if localPath == "" {
		return "", nil, fmt.Errorf("%w: local_path 为空", models.ErrCloudImportFileMissing)
	}
	root := strings.TrimSpace(downloadDir)
	if root == "" {
		return "", nil, fmt.Errorf("%w: DOWNLOAD_DIR 未配置", models.ErrCloudImportFileMissing)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %v", models.ErrCloudImportFileMissing, err)
	}
	if resolvedRoot, resolveErr := filepath.EvalSymlinks(absRoot); resolveErr == nil {
		absRoot = resolvedRoot
	}
	absTarget, err := filepath.Abs(localPath)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %v", models.ErrCloudImportFileMissing, err)
	}
	resolvedTarget, err := filepath.EvalSymlinks(absTarget)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %s", models.ErrCloudImportFileMissing, localPath)
	}
	rel, err := filepath.Rel(absRoot, resolvedTarget)
	if err != nil || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", nil, fmt.Errorf("%w: %s 不在 DOWNLOAD_DIR 内", models.ErrCloudImportFileMissing, localPath)
	}
	info, err := os.Stat(resolvedTarget)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %s", models.ErrCloudImportFileMissing, localPath)
	}
	if !info.Mode().IsRegular() {
		return "", nil, fmt.Errorf("%w: %s 不是单个普通文件", models.ErrCloudImportFileMissing, localPath)
	}
	return resolvedTarget, info, nil
}

func normalizeCloudImportInput(input models.CloudDownloadImportInput, sourcePath string) (models.CloudDownloadImportInput, error) {
	filename := filepath.Base(sourcePath)
	input.Kind = strings.ToLower(strings.TrimSpace(input.Kind))
	switch input.Kind {
	case models.CloudImportKindVideo:
		videoType, err := normalizeCloudImportVideoType(input.Type)
		if err != nil {
			return input, err
		}
		input.Type = videoType
	case models.CloudImportKindArchive:
		if !isSupportedArchiveFile(filename, detectCloudImportMIME(sourcePath)) {
			return input, fmt.Errorf("%w: %s", models.ErrCloudImportUnsupported, filename)
		}
		if input.HasPassword && strings.TrimSpace(input.Password) == "" {
			return input, fmt.Errorf("%w: 压缩包密码不能为空", models.ErrCloudImportInvalidInput)
		}
		if !input.HasPassword {
			input.Password = ""
		}
	default:
		return input, fmt.Errorf("%w: kind 只能是 video 或 archive", models.ErrCloudImportInvalidInput)
	}
	return input, nil
}

func normalizeCloudImportVideoType(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "short", "movie", "episode", "av":
		return strings.ToLower(strings.TrimSpace(raw)), nil
	default:
		return "", fmt.Errorf("%w: 视频 type 只能是 short/movie/episode/av", models.ErrCloudImportInvalidInput)
	}
}

func cloudImportUserID(createdBy *uuid.UUID) uuid.UUID {
	if createdBy == nil {
		return uuid.Nil
	}
	return *createdBy
}

// archiveBatchAlreadyQueued reports whether extraction has already been
// enqueued or started. An import retry must not clear such a batch; it should
// only attach the existing batch and finish.
func archiveBatchAlreadyQueued(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "uploaded", "failed":
		return false
	default:
		return true
	}
}

func cloudImportVideoBelongsToTask(video models.Video, taskID int64) bool {
	var meta struct {
		Source      string `json:"source"`
		CloudTaskID int64  `json:"cloud_task_id"`
	}
	if err := json.Unmarshal(video.Metadata, &meta); err != nil {
		return false
	}
	return meta.Source == "cloud_download" && meta.CloudTaskID == taskID
}

func cloudImportVideoSourcePath(storageRoot string, videoID uuid.UUID, sourcePath string) string {
	return filepath.Join(storageRoot, "videos", videoID.String(), "source-original"+strings.TrimSpace(filepath.Ext(sourcePath)))
}

func cloudImportTargetExists(targetPath string) bool {
	info, err := os.Stat(targetPath)
	return err == nil && info.Mode().IsRegular()
}

func copyCloudImportFile(targetPath, sourcePath string) error {
	// ffmpeg.CopyFile also creates parents, but keep this explicit: the
	// videos/<videoID> directory does not exist for a brand-new import.
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	partPath := targetPath + ".part"
	if err := ffmpeg.CopyFile(partPath, sourcePath); err != nil {
		_ = os.Remove(partPath)
		return err
	}
	if err := os.Rename(partPath, targetPath); err != nil {
		_ = os.Remove(partPath)
		return err
	}
	return nil
}

func detectCloudImportMIME(path string) string {
	detected, err := mimetype.DetectFile(path)
	if err != nil {
		return "application/octet-stream"
	}
	return strings.ToLower(strings.TrimSpace(detected.String()))
}

func isArchiveMIME(mime string) bool {
	lower := strings.ToLower(mime)
	for _, marker := range []string{"zip", "rar", "7z", "x-7z", "compressed"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func isSupportedArchiveFile(name, mime string) bool {
	if _, err := detectArchiveFormat(name); err != nil {
		return false
	}
	lower := strings.ToLower(strings.TrimSpace(mime))
	if lower == "" || lower == "application/octet-stream" {
		// Content sniffing cannot always identify a truncated or empty
		// archive; the supported extension is the fallback in that case.
		return true
	}
	return isArchiveMIME(lower)
}

func archivePreviewMIME(name, detectedMIME string) string {
	if detectedMIME != "" && detectedMIME != "application/octet-stream" && isArchiveMIME(detectedMIME) {
		return detectedMIME
	}
	format, err := detectArchiveFormat(name)
	if err != nil {
		return detectedMIME
	}
	switch format {
	case "zip":
		return "application/zip"
	case "rar":
		return "application/vnd.rar"
	case "7z":
		return "application/x-7z-compressed"
	default:
		return detectedMIME
	}
}
