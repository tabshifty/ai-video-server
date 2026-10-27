package handlers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/response"
)

// TVScreensaverImages 返回登录用户可见的全图库候选，不需要管理员权限。
func (a *API) TVScreensaverImages(c *gin.Context) {
	after := uuid.Nil
	if raw := strings.TrimSpace(c.Query("after")); raw != "" {
		var err error
		after, err = uuid.Parse(raw)
		if err != nil {
			bad(c, "图片游标格式错误")
			return
		}
	}
	limit := 200
	if raw := c.Query("limit"); raw != "" {
		var err error
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 200 {
			bad(c, "每次取图数量必须为 1～200")
			return
		}
	}
	if a.appSvc == nil {
		response.Error(c, 1050, "图片服务暂时不可用")
		return
	}
	result, err := a.appSvc.TVScreensaverImages(c.Request.Context(), after, limit)
	if err != nil {
		response.Error(c, 1051, "屏保图片加载失败")
		return
	}
	c.Header("Cache-Control", "no-store")
	ok(c, result)
}
