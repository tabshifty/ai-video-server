package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/middleware"
	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/response"
)

func (a *API) AdminTVAppReleases(c *gin.Context) {
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	clientType := models.NormalizeAppClientType(c.DefaultQuery("client_type", models.AppClientTypeAndroidTV))
	items, total, err := a.tvAPKSvc.AdminList(c.Request.Context(), models.AdminTvAppReleaseFilter{
		Page:             page,
		PageSize:         pageSize,
		ClientType:       clientType,
		Keyword:          c.Query("q"),
		Status:           strings.TrimSpace(c.Query("status")),
		ABICompleteness:  strings.TrimSpace(c.Query("abi_completeness")),
		CurrentPublished: strings.TrimSpace(c.Query("current_published")) == "1" || strings.EqualFold(strings.TrimSpace(c.Query("current_published")), "true"),
	})
	if err != nil {
		writeTVAPKError(c, 2401, err)
		return
	}
	ok(c, gin.H{
		"items":       items,
		"total_count": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

func (a *API) AdminTVAppReleaseDetail(c *gin.Context) {
	releaseID, okID := parseInt64(c.Param("id"))
	if !okID {
		bad(c, "invalid release id")
		return
	}
	detail, err := a.tvAPKSvc.AdminDetail(c.Request.Context(), releaseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || repository.IsNotFound(err) {
			responseError(c, 404, "release not found")
			return
		}
		writeTVAPKError(c, 2402, err)
		return
	}
	ok(c, detail)
}

func (a *API) AdminUploadTVAppAPK(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		bad(c, "请上传 file 字段")
		return
	}
	clientType := models.NormalizeAppClientType(c.DefaultQuery("client_type", models.AppClientTypeAndroidTV))
	userID, _ := middleware.UserIDFromContext(c)
	username := "admin"
	if userID != uuid.Nil {
		if user, userErr := a.repo.GetUserByID(c.Request.Context(), userID); userErr == nil && strings.TrimSpace(user.Username) != "" {
			username = user.Username
		}
	}
	replaceExisting := c.PostForm("replace_existing") == "1" || strings.EqualFold(c.PostForm("replace_existing"), "true")
	var uploadUserID *uuid.UUID
	if userID != uuid.Nil {
		uploadUserID = &userID
	}
	release, abiInfo, err := a.tvAPKSvc.UploadAPK(c.Request.Context(), file, clientType, uploadUserID, username, replaceExisting)
	if err != nil {
		writeTVAPKError(c, 2403, err)
		return
	}
	detail, err := a.tvAPKSvc.AdminDetail(c.Request.Context(), release.ID)
	if err != nil {
		writeTVAPKError(c, 2403, err)
		return
	}
	ok(c, gin.H{
		"release": detail,
		"abi": gin.H{
			"id":            abiInfo.ID,
			"release_id":    abiInfo.ReleaseID,
			"abi":           abiInfo.ABI,
			"file_name":     abiInfo.FileName,
			"stored_path":   abiInfo.StoredPath,
			"file_size":     abiInfo.FileSize,
			"mime_type":     abiInfo.MIMEType,
			"sha256":        abiInfo.SHA256,
			"is_debuggable": abiInfo.IsDebuggable,
			"uploaded_at":   abiInfo.UploadedAt,
			"replaced_at":   abiInfo.ReplacedAt,
		},
	})
}

func (a *API) AdminUpdateTVAppRelease(c *gin.Context) {
	releaseID, okID := parseInt64(c.Param("id"))
	if !okID {
		bad(c, "invalid release id")
		return
	}
	var req models.AdminTvAppReleaseUpdateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	release, err := a.tvAPKSvc.UpdateRelease(c.Request.Context(), releaseID, req)
	if err != nil {
		writeTVAPKError(c, 2404, err)
		return
	}
	detail, err := a.tvAPKSvc.AdminDetail(c.Request.Context(), release.ID)
	if err != nil {
		writeTVAPKError(c, 2404, err)
		return
	}
	ok(c, detail)
}

func (a *API) AdminPublishTVAppRelease(c *gin.Context) {
	releaseID, okID := parseInt64(c.Param("id"))
	if !okID {
		bad(c, "invalid release id")
		return
	}
	var req models.AdminTvAppReleasePublishInput
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	release, err := a.tvAPKSvc.Publish(c.Request.Context(), releaseID, req)
	if err != nil {
		writeTVAPKError(c, 2405, err)
		return
	}
	detail, err := a.tvAPKSvc.AdminDetail(c.Request.Context(), release.ID)
	if err != nil {
		writeTVAPKError(c, 2405, err)
		return
	}
	ok(c, detail)
}

func (a *API) AdminOfflineTVAppRelease(c *gin.Context) {
	releaseID, okID := parseInt64(c.Param("id"))
	if !okID {
		bad(c, "invalid release id")
		return
	}
	release, err := a.tvAPKSvc.Offline(c.Request.Context(), releaseID)
	if err != nil {
		writeTVAPKError(c, 2406, err)
		return
	}
	detail, err := a.tvAPKSvc.AdminDetail(c.Request.Context(), release.ID)
	if err != nil {
		writeTVAPKError(c, 2406, err)
		return
	}
	ok(c, detail)
}

func (a *API) AdminRestoreTVAppRelease(c *gin.Context) {
	releaseID, okID := parseInt64(c.Param("id"))
	if !okID {
		bad(c, "invalid release id")
		return
	}
	release, err := a.tvAPKSvc.Restore(c.Request.Context(), releaseID)
	if err != nil {
		writeTVAPKError(c, 2407, err)
		return
	}
	detail, err := a.tvAPKSvc.AdminDetail(c.Request.Context(), release.ID)
	if err != nil {
		writeTVAPKError(c, 2407, err)
		return
	}
	ok(c, detail)
}

func (a *API) AdminDeleteTVAppReleaseDraft(c *gin.Context) {
	releaseID, okID := parseInt64(c.Param("id"))
	if !okID {
		bad(c, "invalid release id")
		return
	}
	if err := a.tvAPKSvc.DeleteDraft(c.Request.Context(), releaseID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) || repository.IsNotFound(err) {
			responseError(c, 404, "draft release not found")
			return
		}
		writeTVAPKError(c, 2408, err)
		return
	}
	ok(c, gin.H{"deleted": true})
}

func (a *API) TVAppFamilyReleases(c *gin.Context) {
	clientType := models.NormalizeAppClientType(c.DefaultQuery("client_type", models.AppClientTypeAndroidTV))
	items, err := a.tvAPKSvc.FamilyReleases(c.Request.Context(), clientType)
	if err != nil {
		writeTVAPKError(c, 2409, err)
		return
	}
	ok(c, gin.H{
		"items":       items,
		"total_count": len(items),
	})
}

func (a *API) AdminDownloadTVAppAPK(c *gin.Context) {
	a.downloadTVAppAPK(c)
}

func (a *API) TVAppDownloadAPK(c *gin.Context) {
	a.downloadTVAppAPK(c)
}

func (a *API) downloadTVAppAPK(c *gin.Context) {
	releaseID, okID := parseInt64(c.Param("id"))
	if !okID {
		bad(c, "invalid release id")
		return
	}
	abi := models.TVNormalizeABI(c.Param("abi"))
	if abi == "" {
		bad(c, "invalid abi")
		return
	}

	item, err := a.tvAPKSvc.FindReleaseAPK(c.Request.Context(), releaseID, abi)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || repository.IsNotFound(err) {
			responseError(c, 404, "apk not found")
			return
		}
		writeTVAPKError(c, 2410, err)
		return
	}

	file, info, err := openLocalImageFile(c.Request.Context(), item.StoredPath)
	if err != nil {
		if isLocalImageNotFound(err) {
			responseError(c, 404, "apk file not found")
			return
		}
		writeLocalImageOpenError(c, err, "apk file not found", "apk file temporarily unavailable")
		return
	}
	defer file.Close()

	c.Header("Content-Type", item.MIMEType)
	c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
	c.Header("Cache-Control", "no-store")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", sanitizeAttachmentFilename(item.FileName)))
	http.ServeContent(c.Writer, c.Request, filepath.Base(item.FileName), info.ModTime(), file)
}

func writeTVAPKError(c *gin.Context, code int, err error) {
	var domainErr *models.TVAPKDomainError
	if errors.As(err, &domainErr) {
		responseError(c, code, domainErr.Message)
		return
	}
	if errors.Is(err, pgx.ErrNoRows) {
		responseError(c, 404, "release not found")
		return
	}
	responseError(c, code, err.Error())
}

func responseError(c *gin.Context, code int, msg string) {
	response.Error(c, code, msg)
}

func sanitizeAttachmentFilename(raw string) string {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "tv-app.apk"
	}
	name = strings.ReplaceAll(name, `"`, "")
	name = strings.ReplaceAll(name, "\n", "")
	name = strings.ReplaceAll(name, "\r", "")
	return name
}
