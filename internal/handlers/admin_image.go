package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/middleware"
	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/response"
	"video-server/internal/services"
)

func (a *API) AdminUploadImages(c *gin.Context) {
	if a.imageSvc == nil {
		response.Error(c, 1040, "image service unavailable")
		return
	}

	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}

	actorIDs, err := parseUploadActorIDs(c.PostForm("actor_ids"))
	if err != nil {
		bad(c, "演员ID格式错误")
		return
	}
	actorNames := parseUploadStringList(c.PostForm("actor_names"))
	collectionIDs, err := parseUploadCollectionIDs(c.PostForm("collection_ids"))
	if err != nil {
		bad(c, "图片合集ID格式错误")
		return
	}

	fileHeaders := collectImageUploadFiles(c)
	if len(fileHeaders) == 0 {
		bad(c, "至少上传一张图片")
		return
	}

	title := strings.TrimSpace(c.PostForm("title"))
	description := strings.TrimSpace(c.PostForm("description"))
	results := make([]gin.H, 0, len(fileHeaders))
	successCount := 0
	failedCount := 0
	for _, fileHeader := range fileHeaders {
		itemTitle := title
		if len(fileHeaders) > 1 {
			itemTitle = ""
		}
		saveResult, saveErr := a.imageSvc.SaveUpload(
			c.Request.Context(),
			services.SaveImageInput{
				UserID:        userID,
				Title:         itemTitle,
				Description:   description,
				ActorIDs:      actorIDs,
				ActorNames:    actorNames,
				CollectionIDs: collectionIDs,
			},
			fileHeader,
			a.maxVideoSize,
		)
		if saveErr != nil {
			failedCount++
			results = append(results, gin.H{
				"filename": fileHeader.Filename,
				"success":  false,
				"error":    adminImageUploadErrorMessage(saveErr),
			})
			continue
		}

		successCount++
		results = append(results, gin.H{
			"filename":       fileHeader.Filename,
			"success":        true,
			"image_id":       saveResult.ImageID,
			"already_exists": saveResult.AlreadyExists,
			"status":         saveResult.Status,
			"stored_path":    saveResult.StoredPath,
			"stored_mime":    saveResult.StoredMIME,
		})
	}

	ok(c, gin.H{
		"items":         results,
		"total_count":   len(fileHeaders),
		"success_count": successCount,
		"failed_count":  failedCount,
	})
}

func (a *API) AdminImageCheck(c *gin.Context) {
	var req struct {
		Hash     string `json:"hash"`
		FileSize int64  `json:"file_size"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	req.Hash = strings.TrimSpace(req.Hash)
	if !isSHA256Hex(req.Hash) {
		bad(c, "invalid hash")
		return
	}
	if req.FileSize <= 0 {
		bad(c, "invalid file_size")
		return
	}

	imageID, exists, err := a.repo.FindImageByHash(c.Request.Context(), req.Hash, req.FileSize)
	if err != nil {
		response.Error(c, 1040, err.Error())
		return
	}
	if exists {
		ok(c, gin.H{"exists": true, "image_id": imageID})
		return
	}
	ok(c, gin.H{"exists": false})
}

func collectImageUploadFiles(c *gin.Context) []*multipart.FileHeader {
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		if single, singleErr := c.FormFile("file"); singleErr == nil && single != nil {
			return []*multipart.FileHeader{single}
		}
		return nil
	}

	out := make([]*multipart.FileHeader, 0, 8)
	for _, key := range []string{"files", "file"} {
		files := form.File[key]
		if len(files) == 0 {
			continue
		}
		out = append(out, files...)
	}
	return out
}

func adminImageUploadErrorMessage(err error) string {
	switch {
	case errors.Is(err, services.ErrInvalidImageType):
		return "不支持的图片类型"
	case errors.Is(err, services.ErrImageTooLarge):
		return "图片超过大小限制"
	default:
		return err.Error()
	}
}

func (a *API) AdminImages(c *gin.Context) {
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)

	active, err := parseOptionalBool(c.Query("active"))
	if err != nil {
		bad(c, "active 参数只能是 1 或 0")
		return
	}

	var actorID *uuid.UUID
	if raw := strings.TrimSpace(c.Query("actor_id")); raw != "" {
		id, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			bad(c, "演员ID格式错误")
			return
		}
		actorID = &id
	}

	var collectionID *uuid.UUID
	if raw := strings.TrimSpace(c.Query("collection_id")); raw != "" {
		id, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			bad(c, "图片合集ID格式错误")
			return
		}
		collectionID = &id
	}

	items, total, err := a.repo.AdminListImages(c.Request.Context(), models.AdminImageFilter{
		Page:         page,
		PageSize:     pageSize,
		Keyword:      c.Query("q"),
		Status:       c.Query("status"),
		Active:       active,
		ActorID:      actorID,
		CollectionID: collectionID,
	})
	if err != nil {
		response.Error(c, 1041, err.Error())
		return
	}

	ok(c, gin.H{
		"items":       items,
		"total_count": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

func (a *API) AdminImageDetail(c *gin.Context) {
	imageID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "图片ID格式错误")
		return
	}

	detail, err := a.repo.AdminImageDetail(c.Request.Context(), imageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "图片不存在")
			return
		}
		response.Error(c, 1042, err.Error())
		return
	}
	ok(c, detail)
}

func (a *API) AdminUpdateImage(c *gin.Context) {
	imageID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "图片ID格式错误")
		return
	}

	var req struct {
		Title         *string        `json:"title"`
		Description   *string        `json:"description"`
		Active        *bool          `json:"active"`
		Metadata      map[string]any `json:"metadata"`
		ActorIDs      *[]string      `json:"actor_ids"`
		ActorNames    *[]string      `json:"actor_names"`
		CollectionIDs *[]string      `json:"collection_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}

	existing, err := a.repo.AdminImageDetail(c.Request.Context(), imageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "图片不存在")
			return
		}
		response.Error(c, 1043, err.Error())
		return
	}

	title := existing.Title
	if req.Title != nil {
		title = strings.TrimSpace(*req.Title)
	}
	if title == "" {
		title = "untitled"
	}

	description := existing.Description
	if req.Description != nil {
		description = strings.TrimSpace(*req.Description)
	}

	active := existing.Active
	if req.Active != nil {
		active = *req.Active
	}

	metadata := metadataRawToMap(existing.Metadata)
	if req.Metadata != nil {
		metadata = req.Metadata
	}

	updateActors := req.ActorIDs != nil || req.ActorNames != nil
	var actorIDs []uuid.UUID
	var actorNames []string
	if updateActors {
		if req.ActorIDs != nil {
			parsedActorIDs, parseErr := parseUUIDStrings(*req.ActorIDs)
			if parseErr != nil {
				bad(c, "演员ID格式错误")
				return
			}
			actorIDs = parsedActorIDs
		}
		if req.ActorNames != nil {
			actorNames = normalizeActorNames(*req.ActorNames)
		}
	}

	updateCollections := req.CollectionIDs != nil
	var collectionIDs []uuid.UUID
	if updateCollections {
		parsedCollectionIDs, parseErr := parseUUIDStrings(*req.CollectionIDs)
		if parseErr != nil {
			bad(c, "图片合集ID格式错误")
			return
		}
		collectionIDs = parsedCollectionIDs
	}

	if err := a.repo.AdminUpdateImage(c.Request.Context(), imageID, title, description, active, metadata, actorIDs, actorNames, updateActors, collectionIDs, updateCollections); err != nil {
		response.Error(c, 1044, err.Error())
		return
	}
	ok(c, gin.H{"updated": true})
}

func metadataRawToMap(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func (a *API) AdminDeleteImage(c *gin.Context) {
	imageID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "图片ID格式错误")
		return
	}

	image, err := a.repo.GetImageByID(c.Request.Context(), imageID)
	if err != nil {
		if repository.IsNotFound(err) {
			response.Error(c, 404, "图片不存在")
			return
		}
		response.Error(c, 1045, err.Error())
		return
	}

	variantPaths, err := a.repo.ListImageVariantPaths(c.Request.Context(), imageID)
	if err != nil {
		response.Error(c, 1046, err.Error())
		return
	}

	if err := a.repo.DeleteImageByIDCascade(c.Request.Context(), imageID); err != nil {
		response.Error(c, 1047, err.Error())
		return
	}

	removePaths := make([]string, 0, 2+len(variantPaths))
	removePaths = append(removePaths, strings.TrimSpace(image.OriginalPath), strings.TrimSpace(image.StoredPath))
	removePaths = append(removePaths, variantPaths...)
	for _, removePath := range removePaths {
		cleanPath := strings.TrimSpace(removePath)
		if cleanPath == "" {
			continue
		}
		if rmErr := os.Remove(cleanPath); rmErr != nil && !os.IsNotExist(rmErr) {
			a.logger.Warn("remove image file failed", "image_id", imageID.String(), "path", cleanPath, "error", rmErr)
		}
	}

	variantDir := filepath.Join(a.storageRoot, "images", "variants", imageID.String())
	if rmErr := os.RemoveAll(variantDir); rmErr != nil && !os.IsNotExist(rmErr) {
		a.logger.Warn("remove image variant dir failed", "image_id", imageID.String(), "dir", variantDir, "error", rmErr)
	}

	ok(c, gin.H{"deleted": true, "image_id": imageID})
}

func (a *API) AdminImageView(c *gin.Context) {
	if a.imageSvc == nil {
		response.Error(c, 1048, "image service unavailable")
		return
	}

	imageID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "图片ID格式错误")
		return
	}

	width, err := parseNonNegativeInt(c.Query("w"))
	if err != nil {
		bad(c, "w 参数格式错误")
		return
	}
	height, err := parseNonNegativeInt(c.Query("h"))
	if err != nil {
		bad(c, "h 参数格式错误")
		return
	}
	quality := 82
	if raw := strings.TrimSpace(c.Query("q")); raw != "" {
		q, qErr := strconv.Atoi(raw)
		if qErr != nil {
			bad(c, "q 参数格式错误")
			return
		}
		quality = q
	}
	fit := strings.TrimSpace(c.Query("fit"))

	imagePath, mime, err := a.imageSvc.ResolveViewPath(c.Request.Context(), imageID, width, height, fit, quality)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || strings.Contains(strings.ToLower(err.Error()), "not found") {
			response.Error(c, 404, "图片不存在")
			return
		}
		response.Error(c, 1049, err.Error())
		return
	}

	if mime != "" {
		c.Header("Content-Type", mime)
	}
	if !tryServeLocalImagePath(c, imagePath, "图片不存在", "图片暂时不可用") {
		return
	}
}

func parseOptionalBool(raw string) (*bool, error) {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return nil, nil
	}
	switch v {
	case "1", "true", "yes":
		b := true
		return &b, nil
	case "0", "false", "no":
		b := false
		return &b, nil
	default:
		return nil, fmt.Errorf("invalid bool")
	}
}

func parseNonNegativeInt(raw string) (int, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, fmt.Errorf("negative")
	}
	return n, nil
}

func (a *API) AdminImageCollections(c *gin.Context) {
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	keyword := c.Query("q")

	active, err := parseOptionalBool(c.Query("active"))
	if err != nil {
		bad(c, "active 参数只能是 1 或 0")
		return
	}

	items, total, err := a.repo.ListImageCollections(c.Request.Context(), keyword, active, page, pageSize)
	if err != nil {
		response.Error(c, 1050, err.Error())
		return
	}
	ok(c, gin.H{
		"items":       items,
		"total_count": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

func (a *API) AdminCreateImageCollection(c *gin.Context) {
	var req models.AdminImageCollectionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	item, err := a.repo.CreateImageCollection(c.Request.Context(), req)
	if err != nil {
		if repository.IsUniqueViolation(err) {
			response.Error(c, 1051, "图片合集名称已存在")
			return
		}
		response.Error(c, 1051, err.Error())
		return
	}
	ok(c, item)
}

func (a *API) AdminUpdateImageCollection(c *gin.Context) {
	collectionID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "图片合集ID格式错误")
		return
	}

	var req models.AdminImageCollectionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}

	item, err := a.repo.UpdateImageCollection(c.Request.Context(), collectionID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "图片合集不存在")
			return
		}
		if repository.IsUniqueViolation(err) {
			response.Error(c, 1052, "图片合集名称已存在")
			return
		}
		response.Error(c, 1052, err.Error())
		return
	}
	ok(c, item)
}

func (a *API) AdminDeleteImageCollection(c *gin.Context) {
	collectionID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "图片合集ID格式错误")
		return
	}
	detached, err := a.repo.DeleteImageCollection(c.Request.Context(), collectionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "图片合集不存在")
			return
		}
		response.Error(c, 1053, err.Error())
		return
	}
	ok(c, gin.H{
		"deleted":         true,
		"collection_id":   collectionID,
		"detached_images": detached,
	})
}
