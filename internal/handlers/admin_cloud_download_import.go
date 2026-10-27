package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/middleware"
	"video-server/internal/models"
	"video-server/internal/response"
)

// cloudDownloadImportService is the admin-facing contract for explicit cloud
// download imports.
type cloudDownloadImportService interface {
	GetImportTask(ctx context.Context, taskID int64) (models.CloudDownloadTask, error)
	Preview(ctx context.Context, taskID int64) (models.CloudDownloadImportPreview, error)
	Create(ctx context.Context, taskID int64, input models.CloudDownloadImportInput, createdBy *uuid.UUID) (models.CloudDownloadImport, error)
}

// SetCloudDownloadImportService wires the explicit import service. Like the
// download routes, the import routes stay registered and return 503 when the
// feature is off.
func (a *API) SetCloudDownloadImportService(svc cloudDownloadImportService) {
	a.cloudDownloadImport = svc
}

func (a *API) requireCloudDownloadImport(c *gin.Context) bool {
	if a.cloudDownloadImport == nil {
		response.Error(c, 503, "云下载导入功能未启用")
		return false
	}
	return true
}

// AdminCloudDownloadImportTask loads progress even after the source was moved into an archive batch.
func (a *API) AdminCloudDownloadImportTask(c *gin.Context) {
	if !a.requireCloudDownload(c) || !a.requireCloudDownloadImport(c) {
		return
	}
	taskID, okID := parseInt64(c.Param("id"))
	if !okID || taskID <= 0 {
		response.Error(c, 400, "任务 id 无效")
		return
	}
	task, err := a.cloudDownloadImport.GetImportTask(c.Request.Context(), taskID)
	if err != nil {
		respondCloudImportError(c, err)
		return
	}
	ok(c, task)
}

// AdminCloudDownloadImportPreview inspects the completed task's actual file.
func (a *API) AdminCloudDownloadImportPreview(c *gin.Context) {
	if !a.requireCloudDownload(c) || !a.requireCloudDownloadImport(c) {
		return
	}
	taskID, okID := parseInt64(c.Param("id"))
	if !okID {
		response.Error(c, 400, "任务 id 无效")
		return
	}
	preview, err := a.cloudDownloadImport.Preview(c.Request.Context(), taskID)
	if err != nil {
		respondCloudImportError(c, err)
		return
	}
	ok(c, preview)
}

// AdminCloudDownloadImport creates the one-shot import and enqueues the worker.
func (a *API) AdminCloudDownloadImport(c *gin.Context) {
	if !a.requireCloudDownload(c) || !a.requireCloudDownloadImport(c) {
		return
	}
	taskID, okID := parseInt64(c.Param("id"))
	if !okID {
		response.Error(c, 400, "任务 id 无效")
		return
	}
	var req struct {
		Kind          string      `json:"kind"`
		Type          string      `json:"type"`
		SiteCategory  string      `json:"site_category"`
		Title         string      `json:"title"`
		Description   string      `json:"description"`
		Tags          []string    `json:"tags"`
		ActorIDs      []uuid.UUID `json:"actor_ids"`
		ActorNames    []string    `json:"actor_names"`
		CollectionIDs []uuid.UUID `json:"collection_ids"`
		// A string keeps an empty form value from failing JSON binding; it is
		// parsed only when non-empty.
		ImageCollectionID string `json:"image_collection_id"`

		DefaultDescription        string      `json:"default_description"`
		DefaultTags               []string    `json:"default_tags"`
		DefaultVideoCollectionIDs []uuid.UUID `json:"default_video_collection_ids"`
		DefaultImageCollectionIDs []uuid.UUID `json:"default_image_collection_ids"`
		HasPassword               bool        `json:"has_password"`
		Password                  string      `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if strings.TrimSpace(req.Kind) == "" {
		response.Error(c, 400, "kind 不能为空")
		return
	}
	var imageCollectionID *uuid.UUID
	if rawImageCollectionID := strings.TrimSpace(req.ImageCollectionID); rawImageCollectionID != "" {
		parsed, parseErr := uuid.Parse(rawImageCollectionID)
		if parseErr != nil {
			response.Error(c, 400, "image_collection_id 格式错误")
			return
		}
		imageCollectionID = &parsed
	}
	var createdBy *uuid.UUID
	if userID, okUser := middleware.UserIDFromContext(c); okUser {
		createdBy = &userID
	}
	record, err := a.cloudDownloadImport.Create(c.Request.Context(), taskID, models.CloudDownloadImportInput{
		Kind:                      req.Kind,
		Type:                      req.Type,
		SiteCategory:              req.SiteCategory,
		Title:                     req.Title,
		Description:               req.Description,
		Tags:                      req.Tags,
		ActorIDs:                  req.ActorIDs,
		ActorNames:                req.ActorNames,
		CollectionIDs:             req.CollectionIDs,
		ImageCollectionID:         imageCollectionID,
		DefaultDescription:        req.DefaultDescription,
		DefaultTags:               req.DefaultTags,
		DefaultVideoCollectionIDs: req.DefaultVideoCollectionIDs,
		DefaultImageCollectionIDs: req.DefaultImageCollectionIDs,
		HasPassword:               req.HasPassword,
		Password:                  req.Password,
	}, createdBy)
	if err != nil {
		respondCloudImportError(c, err)
		return
	}
	ok(c, gin.H{"import": record})
}

func respondCloudImportError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, models.ErrCloudImportNotFound), errors.Is(err, models.ErrCloudDownloadNotFound):
		response.Error(c, 404, err.Error())
	case errors.Is(err, models.ErrCloudImportNotCompleted):
		response.Error(c, 409, err.Error())
	case errors.Is(err, models.ErrCloudImportFileMissing):
		response.Error(c, 410, err.Error())
	case errors.Is(err, models.ErrCloudImportUnsupported):
		response.Error(c, 422, err.Error())
	case errors.Is(err, models.ErrCloudImportInvalidInput):
		response.Error(c, 400, err.Error())
	default:
		response.Error(c, 500, err.Error())
	}
}
