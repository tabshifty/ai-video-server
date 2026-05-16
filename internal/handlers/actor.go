package handlers

import (
	"github.com/gin-gonic/gin"

	"video-server/internal/middleware"
	"video-server/internal/response"
)

func (a *API) ActorDetail(c *gin.Context) {
	if _, okUser := middleware.UserIDFromContext(c); !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	actorID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid actor id")
		return
	}

	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 24)
	payload, err := a.appSvc.ActorDetail(c.Request.Context(), actorID, page, pageSize)
	if err != nil {
		response.Error(c, 20, err.Error())
		return
	}
	ok(c, payload)
}
