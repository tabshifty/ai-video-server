package handlers

import (
	"github.com/gin-gonic/gin"

	"video-server/internal/response"
)

func (a *API) Search(c *gin.Context) {
	q := c.Query("q")
	typ := c.DefaultQuery("type", "all")
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	result, err := a.appSvc.Search(c.Request.Context(), q, typ, page, pageSize)
	if err != nil {
		response.Error(c, 25, err.Error())
		return
	}
	ok(c, result)
}
