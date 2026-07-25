package handlers

import (
	"github.com/gin-gonic/gin"

	"video-server/internal/response"
)

// AppShortCollections 返回手机端可见的短视频合集目录，按后台 sort_order DESC、updated_at DESC 排序。
// 仅展示启用且至少一条可播放短视频的合集（见可见短视频合集）。首期不提供按名称搜索。
func (a *API) AppShortCollections(c *gin.Context) {
	if a.appSvc == nil {
		response.Error(c, 1060, "app service unavailable")
		return
	}
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	result, err := a.appSvc.ShortCollections(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, 1061, err.Error())
		return
	}
	ok(c, result)
}
