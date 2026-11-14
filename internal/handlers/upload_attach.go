package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/middleware"
	"video-server/internal/response"
	"video-server/internal/services"
)

// UploadAttach 把本次选择的合集关系追加到已存在的视频上。
// 它是秒传命中后的显式写入口：只读的 /upload/check 不承担任何写入。
// @Summary Attach collections to an existing uploaded video
// @Tags upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body UploadAttachRequest true "hash and collection ids of the existing video"
// @Success 200 {object} APIResponse
// @Router /upload/attach [post]
func (a *API) UploadAttach(c *gin.Context) {
	role, okRole := middleware.RoleFromContext(c)
	if !okRole {
		response.Error(c, 401, "unauthorized")
		return
	}
	if role != "admin" {
		response.Error(c, 403, "insufficient role for collection attachment")
		return
	}

	var req struct {
		Hash              string   `json:"hash"`
		FileSize          int64    `json:"file_size"`
		Type              string   `json:"type"`
		CollectionIDs     []string `json:"collection_ids"`
		ImageCollectionID string   `json:"image_collection_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	typ := strings.ToLower(strings.TrimSpace(req.Type))
	if typ != "short" && typ != "movie" && typ != "episode" && typ != "av" {
		bad(c, "type must be one of: short, movie, episode, av")
		return
	}
	hash := strings.ToLower(strings.TrimSpace(req.Hash))
	if !isSHA256Hex(hash) {
		bad(c, "invalid hash")
		return
	}
	if req.FileSize <= 0 {
		bad(c, "invalid file_size")
		return
	}
	collectionIDs, err := parseUUIDStrings(req.CollectionIDs)
	if err != nil {
		bad(c, "合集ID格式错误")
		return
	}
	if err := validateCollectionsForType(typ, collectionIDs); err != nil {
		bad(c, "仅短视频支持合集")
		return
	}
	imageCollectionID, err := parseUploadImageCollectionID(req.ImageCollectionID)
	if err != nil {
		bad(c, "图片图集ID格式错误")
		return
	}

	// 不信任客户端传入的视频 ID：用 hash + file_size 重新反查。
	videoID, exists, err := a.repo.FindVideoByHash(c.Request.Context(), hash, req.FileSize)
	if err != nil {
		response.Error(c, 2, err.Error())
		return
	}
	if !exists {
		response.Error(c, 2008, "视频不存在")
		return
	}
	video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
	if err != nil {
		response.Error(c, 2, err.Error())
		return
	}
	if video.Type != typ {
		response.Error(c, 409, "hash exists with another video type")
		return
	}
	attachment, err := services.AttachExistingVideoCollections(c.Request.Context(), a.repo, video, collectionIDs, imageCollectionID)
	if err != nil {
		response.Error(c, 2009, err.Error())
		return
	}
	ok(c, existingAttachmentPayload(videoID, attachment))
}

func existingUploadPayload(result services.UploadResult) gin.H {
	payload := gin.H{
		"exists":                 true,
		"video_id":               result.VideoID,
		"added_collections":      result.AddedCollections,
		"added_image_collection": result.AddedImageCollection,
	}
	if result.SkippedPendingDelete {
		payload["pending_delete"] = true
	}
	return payload
}

func existingAttachmentPayload(videoID uuid.UUID, attachment services.ExistingVideoCollectionAttachment) gin.H {
	payload := gin.H{
		"exists":                 true,
		"video_id":               videoID,
		"added_collections":      attachment.AddedCollections,
		"added_image_collection": attachment.AddedImageCollection,
	}
	if attachment.VideoPendingDelete {
		payload["pending_delete"] = true
	}
	return payload
}
