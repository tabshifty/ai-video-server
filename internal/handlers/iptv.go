package handlers

import (
	"context"
	"errors"
	"io"
	"strings"

	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/response"
	"video-server/internal/services"
)

type iptvService interface {
	Status(context.Context) (models.IPTVPlaylistStatus, error)
	Upload(context.Context, string, io.Reader) (models.IPTVPlaylistStatus, error)
	SaveSourceURL(context.Context, string) (models.IPTVPlaylistStatus, error)
	Refresh(context.Context) (models.IPTVPlaylistStatus, error)
}

type tvAPKService interface {
	AdminList(context.Context, models.AdminTvAppReleaseFilter) ([]models.AdminTvAppReleaseListItem, int, error)
	AdminDetail(context.Context, int64) (models.AdminTvAppReleaseDetail, error)
	UploadAPK(context.Context, *multipart.FileHeader, *uuid.UUID, string, bool) (models.TVAppReleaseRecord, models.TVAppReleaseABIInfo, error)
	UpdateRelease(context.Context, int64, models.AdminTvAppReleaseUpdateInput) (models.TVAppReleaseRecord, error)
	Publish(context.Context, int64, models.AdminTvAppReleasePublishInput) (models.TVAppReleaseRecord, error)
	Offline(context.Context, int64) (models.TVAppReleaseRecord, error)
	Restore(context.Context, int64) (models.TVAppReleaseRecord, error)
	DeleteDraft(context.Context, int64) error
	FamilyReleases(context.Context) ([]models.TVAppFamilyRelease, error)
	FindReleaseAPK(context.Context, int64, string) (models.TVAppReleaseABIInfo, error)
}

func (a *API) AdminIPTVPlaylist(c *gin.Context) {
	status, err := a.iptvSvc.Status(c.Request.Context())
	if err != nil {
		response.Error(c, 2301, err.Error())
		return
	}
	ok(c, status)
}

func (a *API) AdminIPTVUploadPlaylist(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		bad(c, "请上传 file 字段")
		return
	}
	defer file.Close()

	status, err := a.iptvSvc.Upload(c.Request.Context(), header.Filename, file)
	if err != nil {
		response.Error(c, 2302, err.Error())
		return
	}
	ok(c, status)
}

func (a *API) AdminIPTVSaveSource(c *gin.Context) {
	var req struct {
		SourceURL string `json:"source_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "请求内容无效")
		return
	}
	status, err := a.iptvSvc.SaveSourceURL(c.Request.Context(), strings.TrimSpace(req.SourceURL))
	if err != nil {
		response.Error(c, 2303, err.Error())
		return
	}
	ok(c, status)
}

func (a *API) AdminIPTVRefreshPlaylist(c *gin.Context) {
	status, err := a.iptvSvc.Refresh(c.Request.Context())
	if err != nil {
		if errors.Is(err, services.ErrIPTVSourceURLRequired) {
			response.Error(c, 2304, "请先设置 IPTV 远程地址")
			return
		}
		response.Error(c, 2304, err.Error())
		return
	}
	ok(c, status)
}

func (a *API) TVIPTVChannels(c *gin.Context) {
	status, err := a.iptvSvc.Status(c.Request.Context())
	if err != nil {
		response.Error(c, 2305, err.Error())
		return
	}
	ok(c, status)
}
