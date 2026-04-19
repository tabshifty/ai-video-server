package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/response"
)

func (a *API) ShortDiscover(c *gin.Context) {
	mode := strings.ToLower(strings.TrimSpace(c.Query("mode")))
	tag := strings.TrimSpace(c.Query("tag"))
	rawCollectionID := strings.TrimSpace(c.Query("collection_id"))
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)

	var collectionID *uuid.UUID
	if rawCollectionID != "" {
		parsed, err := uuid.Parse(rawCollectionID)
		if err != nil {
			bad(c, "invalid collection_id")
			return
		}
		collectionID = &parsed
	}

	result, err := a.appSvc.ShortDiscover(
		c.Request.Context(),
		mode,
		tag,
		collectionID,
		page,
		pageSize,
	)
	if err != nil {
		response.Error(c, 31, err.Error())
		return
	}
	ok(c, result)
}
