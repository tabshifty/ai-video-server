package handlers

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"video-server/internal/middleware"
	"video-server/internal/models"
	"video-server/internal/response"
	"video-server/internal/services"
)

// AdminArchiveImportBatches lists archive import batches.
func (a *API) AdminArchiveImportBatches(c *gin.Context) {
	if a.archiveImportSvc == nil {
		response.Error(c, 1071, "archive import service unavailable")
		return
	}
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	items, total, err := a.archiveImportSvc.ListBatches(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, 1072, err.Error())
		return
	}
	ok(c, gin.H{
		"items":       items,
		"total_count": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

// AdminUploadArchiveImport uploads an archive and creates a batch.
func (a *API) AdminUploadArchiveImport(c *gin.Context) {
	if a.archiveImportSvc == nil {
		response.Error(c, 1071, "archive import service unavailable")
		return
	}
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	title := strings.TrimSpace(c.PostForm("title"))
	defaultDescription := strings.TrimSpace(c.PostForm("default_description"))
	defaultTags := parseUploadStringList(c.PostForm("default_tags"))
	defaultVideoCollectionIDs, err := parseUploadCollectionIDs(c.PostForm("default_video_collection_ids"))
	if err != nil {
		bad(c, "视频合集ID格式错误")
		return
	}
	defaultImageCollectionIDs, err := parseUploadCollectionIDs(c.PostForm("default_image_collection_ids"))
	if err != nil {
		bad(c, "图片合集ID格式错误")
		return
	}
	hasPassword, err := parseOptionalBool(c.PostForm("has_password"))
	if err != nil {
		bad(c, "has_password 参数只能是 1 或 0")
		return
	}
	password := strings.TrimSpace(c.PostForm("password"))
	if hasPassword != nil && !*hasPassword {
		password = ""
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		bad(c, "压缩包文件是必填项")
		return
	}

	batch, err := a.archiveImportSvc.UploadArchive(c.Request.Context(), services.ArchiveImportUploadInput{
		UserID:                    userID,
		Title:                     title,
		DefaultDescription:        defaultDescription,
		DefaultTags:               defaultTags,
		DefaultVideoCollectionIDs: defaultVideoCollectionIDs,
		DefaultImageCollectionIDs: defaultImageCollectionIDs,
		HasPassword:               hasPassword != nil && *hasPassword,
		Password:                  password,
	}, fileHeader)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrArchiveUnsupportedFormat):
			bad(c, "仅支持 zip、rar、7z")
		case errors.Is(err, services.ErrArchivePasswordRequired):
			response.Error(c, 1073, "压缩包需要密码")
		case errors.Is(err, services.ErrArchiveNestedArchive):
			response.Error(c, 1074, "不允许压缩包嵌套")
		default:
			response.Error(c, 1072, err.Error())
		}
		return
	}
	ok(c, batch)
}

// AdminArchiveImportBatchDetail returns one batch and its file list.
func (a *API) AdminArchiveImportBatchDetail(c *gin.Context) {
	if a.archiveImportSvc == nil {
		response.Error(c, 1071, "archive import service unavailable")
		return
	}
	batchID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid batch id")
		return
	}
	batch, files, err := a.archiveImportSvc.GetBatchWithFiles(c.Request.Context(), batchID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || strings.Contains(strings.ToLower(err.Error()), "not found") {
			response.Error(c, 404, "batch not found")
			return
		}
		response.Error(c, 1072, err.Error())
		return
	}
	ok(c, models.ArchiveImportBatchDetail{
		ArchiveImportBatch: batch,
		Files:              files,
	})
}

// AdminArchiveImportFileDetail returns one archive file.
func (a *API) AdminArchiveImportFileDetail(c *gin.Context) {
	if a.archiveImportSvc == nil {
		response.Error(c, 1071, "archive import service unavailable")
		return
	}
	fileID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid file id")
		return
	}
	file, err := a.archiveImportSvc.GetFile(c.Request.Context(), fileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || strings.Contains(strings.ToLower(err.Error()), "not found") {
			response.Error(c, 404, "file not found")
			return
		}
		response.Error(c, 1072, err.Error())
		return
	}
	ok(c, file)
}

// AdminUpdateArchiveImportFile updates batch-file metadata.
func (a *API) AdminUpdateArchiveImportFile(c *gin.Context) {
	if a.archiveImportSvc == nil {
		response.Error(c, 1071, "archive import service unavailable")
		return
	}
	fileID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid file id")
		return
	}
	var req struct {
		Title              string   `json:"title"`
		Description        string   `json:"description"`
		Tags               []string `json:"tags"`
		VideoType          string   `json:"video_type"`
		VideoCollectionIDs []string `json:"video_collection_ids"`
		ImageCollectionIDs []string `json:"image_collection_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	videoCollectionIDs, err := parseUUIDStrings(req.VideoCollectionIDs)
	if err != nil {
		bad(c, "视频合集ID格式错误")
		return
	}
	imageCollectionIDs, err := parseUUIDStrings(req.ImageCollectionIDs)
	if err != nil {
		bad(c, "图片合集ID格式错误")
		return
	}
	file, err := a.archiveImportSvc.UpdateFile(c.Request.Context(), fileID, services.ArchiveImportFileUpdateInput{
		Title:              req.Title,
		Description:        req.Description,
		Tags:               req.Tags,
		VideoType:          req.VideoType,
		VideoCollectionIDs: videoCollectionIDs,
		ImageCollectionIDs: imageCollectionIDs,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || strings.Contains(strings.ToLower(err.Error()), "not found") {
			response.Error(c, 404, "file not found")
			return
		}
		response.Error(c, 1075, err.Error())
		return
	}
	ok(c, file)
}

// AdminProcessArchiveImportFile processes one archive file.
func (a *API) AdminProcessArchiveImportFile(c *gin.Context) {
	if a.archiveImportSvc == nil {
		response.Error(c, 1071, "archive import service unavailable")
		return
	}
	fileID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid file id")
		return
	}
	file, err := a.archiveImportSvc.ProcessFile(c.Request.Context(), fileID)
	if err != nil {
		response.Error(c, 1076, err.Error())
		return
	}
	ok(c, file)
}

// AdminProcessArchiveImportBatch processes all pending files in a batch.
func (a *API) AdminProcessArchiveImportBatch(c *gin.Context) {
	if a.archiveImportSvc == nil {
		response.Error(c, 1071, "archive import service unavailable")
		return
	}
	batchID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid batch id")
		return
	}
	files, err := a.archiveImportSvc.ProcessAllFiles(c.Request.Context(), batchID)
	if err != nil {
		response.Error(c, 1076, err.Error())
		return
	}
	ok(c, gin.H{
		"items": files,
	})
}

// AdminRetryArchiveImportExtract retries archive extraction with a password.
func (a *API) AdminRetryArchiveImportExtract(c *gin.Context) {
	if a.archiveImportSvc == nil {
		response.Error(c, 1071, "archive import service unavailable")
		return
	}
	batchID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid batch id")
		return
	}
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	batch, err := a.archiveImportSvc.RetryExtract(c.Request.Context(), batchID, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrArchivePasswordRequired) {
			response.Error(c, 1073, "压缩包需要密码")
			return
		}
		if errors.Is(err, services.ErrArchiveNestedArchive) {
			response.Error(c, 1074, "不允许压缩包嵌套")
			return
		}
		response.Error(c, 1076, err.Error())
		return
	}
	ok(c, batch)
}
