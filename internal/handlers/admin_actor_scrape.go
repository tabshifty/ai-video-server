package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"

	"video-server/internal/response"
)

// @Summary Admin preview actor scrape candidates
// @Tags admin-actors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body AdminActorScrapePreviewRequest true "actor scrape payload"
// @Success 200 {object} APIResponse
// @Failure 200 {object} APIResponse
// @Router /admin/actors/scrape/preview [post]
func (a *API) AdminActorScrapePreview(c *gin.Context) {
	var req struct {
		Name   string `json:"name"`
		Source string `json:"source"`
		Limit  int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "请求参数格式错误")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		bad(c, "演员姓名不能为空")
		return
	}

	source := strings.ToLower(strings.TrimSpace(req.Source))
	if source == "" {
		source = "tmdb"
	}
	if source != "tmdb" && source != "javdb" {
		bad(c, "source 仅支持 tmdb 或 javdb")
		return
	}

	items, err := a.scrapeSvc.PreviewActorByName(c.Request.Context(), req.Name, source, req.Limit)
	if err != nil {
		response.Error(c, 1027, err.Error())
		return
	}

	ok(c, gin.H{
		"source": source,
		"items":  items,
	})
}
