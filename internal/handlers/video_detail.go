package handlers

import (
	"github.com/gin-gonic/gin"

	"video-server/internal/middleware"
	"video-server/internal/response"
	"video-server/internal/utils"
)

func (a *API) VideoDetail(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	videoID, okVideo := parseUUID(c.Param("id"))
	if !okVideo {
		bad(c, "invalid video id")
		return
	}
	detail, err := a.appSvc.VideoDetail(c.Request.Context(), userID, videoID)
	if err != nil {
		response.Error(c, 20, err.Error())
		return
	}
	detail.PlayURL = utils.VideoPlayURL(detail.ID)
	ok(c, detail)
}
