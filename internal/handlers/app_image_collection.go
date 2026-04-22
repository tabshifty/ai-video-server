package handlers

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"video-server/internal/response"
)

func (a *API) AppImageCollections(c *gin.Context) {
	if a.appSvc == nil {
		response.Error(c, 1050, "app service unavailable")
		return
	}

	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)

	result, err := a.appSvc.ImageCollections(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, 1051, err.Error())
		return
	}
	ok(c, result)
}

func (a *API) AppImageCollectionDetail(c *gin.Context) {
	if a.appSvc == nil {
		response.Error(c, 1052, "app service unavailable")
		return
	}

	collectionID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "图集ID格式错误")
		return
	}

	detail, err := a.appSvc.ImageCollectionDetail(c.Request.Context(), collectionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || strings.Contains(strings.ToLower(err.Error()), "no rows") {
			response.Error(c, 404, "图集不存在")
			return
		}
		response.Error(c, 1053, err.Error())
		return
	}
	ok(c, detail)
}

func (a *API) AppImageView(c *gin.Context) {
	if a.imageSvc == nil {
		response.Error(c, 1054, "image service unavailable")
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

	imagePath, mime, err := a.imageSvc.ResolveAppViewPath(c.Request.Context(), imageID, width, height, fit, quality)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(strings.ToLower(err.Error()), "not active") {
			response.Error(c, 404, "图片不存在")
			return
		}
		response.Error(c, 1055, err.Error())
		return
	}

	if stat, statErr := os.Stat(imagePath); statErr == nil {
		c.Header("Content-Length", strconv.FormatInt(stat.Size(), 10))
	}
	if mime != "" {
		c.Header("Content-Type", mime)
	}
	c.Header("Cache-Control", "public, max-age=86400")
	c.File(imagePath)
}
