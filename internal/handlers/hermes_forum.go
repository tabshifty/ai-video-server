package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/response"
	"video-server/internal/services"
)

const hermesForumMaxRequestBytes = 2 * 1024 * 1024

type hermesForumService interface {
	ListAdmin(context.Context, int, int) ([]models.AdminForumPostListItem, int, error)
	Discover(context.Context, models.ForumPostDiscoverInput) ([]models.ForumPostDiscoverResult, error)
	CompleteInspection(context.Context, uuid.UUID, models.ForumPostInspectionInput) (models.ForumPostInspectionResult, error)
}

// AdminForumPosts serves the read-only, recent forum resource list to administrators.
func (a *API) AdminForumPosts(c *gin.Context) {
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	items, total, err := a.hermesForumSvc.ListAdmin(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, 2501, err.Error())
		return
	}
	ok(c, gin.H{
		"items":       items,
		"total_count": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

func (a *API) HermesDiscoverForumPosts(c *gin.Context) {
	var input models.ForumPostDiscoverInput
	if err := decodeHermesForumJSON(c, &input); err != nil {
		writeHermesForumDecodeError(c, err)
		return
	}
	result, err := a.hermesForumSvc.Discover(c.Request.Context(), input)
	if err != nil {
		a.writeHermesForumServiceError(c, err)
		return
	}
	writeHermesForumJSON(c, http.StatusOK, "ok", result)
}

func (a *API) HermesCompleteForumPostInspection(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeHermesForumJSON(c, http.StatusBadRequest, "帖子 ID 非法", nil)
		return
	}
	var input models.ForumPostInspectionInput
	if err := decodeHermesForumJSON(c, &input); err != nil {
		writeHermesForumDecodeError(c, err)
		return
	}
	result, err := a.hermesForumSvc.CompleteInspection(c.Request.Context(), postID, input)
	if err != nil {
		a.writeHermesForumServiceError(c, err)
		return
	}
	writeHermesForumJSON(c, http.StatusOK, "ok", result)
}

func decodeHermesForumJSON(c *gin.Context, target any) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, hermesForumMaxRequestBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return errors.New("请求体只能包含一个 JSON 对象")
		}
		return err
	}
	return nil
}

func writeHermesForumDecodeError(c *gin.Context, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeHermesForumJSON(c, http.StatusRequestEntityTooLarge, "请求体过大", nil)
		return
	}
	writeHermesForumJSON(c, http.StatusBadRequest, "请求 JSON 非法", nil)
}

func (a *API) writeHermesForumServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrHermesForumInvalidInput):
		writeHermesForumJSON(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, services.ErrHermesForumPostNotFound):
		writeHermesForumJSON(c, http.StatusNotFound, "帖子不存在", nil)
	case errors.Is(err, services.ErrHermesForumInspectionConflict):
		writeHermesForumJSON(c, http.StatusConflict, "检查结果与已保存结果冲突", nil)
	default:
		if a.logger != nil {
			a.logger.Error("Hermes 论坛接口失败", "error", err)
		}
		writeHermesForumJSON(c, http.StatusInternalServerError, "服务内部错误", nil)
	}
}

func writeHermesForumJSON(c *gin.Context, status int, msg string, data any) {
	code := status
	if status == http.StatusOK {
		code = 0
	}
	c.JSON(status, gin.H{"code": code, "msg": msg, "data": data})
}
