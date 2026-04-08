package handlers

import (
	"github.com/gin-gonic/gin"

	"video-server/internal/middleware"
	"video-server/internal/response"
)

func (a *API) UserProfile(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	profile, err := a.appSvc.Profile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, 26, err.Error())
		return
	}
	ok(c, profile)
}

func (a *API) UpdateUserProfile(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	var req struct {
		OldPassword string `json:"old_password"`
		Email       string `json:"email"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	if req.OldPassword == "" {
		bad(c, "old_password is required")
		return
	}
	if req.Email == "" && req.NewPassword == "" {
		bad(c, "email or new_password is required")
		return
	}
	if err := a.appSvc.UpdateProfile(c.Request.Context(), userID, req.OldPassword, req.Email, req.NewPassword); err != nil {
		response.Error(c, 27, err.Error())
		return
	}
	ok(c, gin.H{"updated": true})
}

func (a *API) UploadedVideos(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	result, err := a.appSvc.UploadedVideos(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, 28, err.Error())
		return
	}
	ok(c, result)
}

func (a *API) LikedVideos(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	result, err := a.appSvc.LikedVideos(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, 29, err.Error())
		return
	}
	ok(c, result)
}

func (a *API) FavoritedVideos(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	result, err := a.appSvc.FavoritedVideos(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, 30, err.Error())
		return
	}
	ok(c, result)
}
