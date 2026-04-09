package handlers

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"video-server/internal/middleware"
	"video-server/internal/queue"
	"video-server/internal/response"
	"video-server/internal/services"
)

func (a *API) UploadInit(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	role, okRole := middleware.RoleFromContext(c)
	if !okRole {
		response.Error(c, 401, "unauthorized")
		return
	}

	var req struct {
		Filename    string   `json:"filename"`
		FileSize    int64    `json:"file_size"`
		ChunkSize   int64    `json:"chunk_size"`
		TotalChunks int      `json:"total_chunks"`
		Hash        string   `json:"hash"`
		Type        string   `json:"type"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	if req.Type != "short" && req.Type != "movie" && req.Type != "episode" {
		bad(c, "type must be one of: short, movie, episode")
		return
	}
	if role != "admin" && req.Type != "short" {
		response.Error(c, 403, "insufficient role for video type")
		return
	}
	if !isSHA256Hex(req.Hash) {
		bad(c, "invalid hash")
		return
	}
	if req.FileSize <= 0 || req.ChunkSize <= 0 || req.TotalChunks <= 0 {
		bad(c, "invalid chunk payload")
		return
	}
	if a.maxVideoSize > 0 && req.FileSize > a.maxVideoSize {
		bad(c, "file size exceeds limit")
		return
	}

	session, err := a.chunkUpload.Init(
		c.Request.Context(),
		userID,
		req.Filename,
		req.FileSize,
		req.ChunkSize,
		req.TotalChunks,
		req.Hash,
		req.Type,
		req.Title,
		req.Description,
		req.Tags,
	)
	if err != nil {
		response.Error(c, 2001, err.Error())
		return
	}
	ok(c, gin.H{
		"upload_session_id": session.ID,
		"uploaded_chunks":   []int{},
		"total_chunks":      session.TotalChunks,
	})
}

func (a *API) UploadChunk(c *gin.Context) {
	sessionID := strings.TrimSpace(c.Query("session_id"))
	chunkRaw := strings.TrimSpace(c.Query("chunk_index"))
	chunkIndex, err := strconv.Atoi(chunkRaw)
	if sessionID == "" || err != nil {
		bad(c, "session_id and chunk_index are required")
		return
	}
	session, err := a.chunkUpload.SaveChunk(c.Request.Context(), sessionID, chunkIndex, c.Request.Body)
	if err != nil {
		response.Error(c, 2002, err.Error())
		return
	}
	uploaded := make([]int, 0, len(session.Uploaded))
	for idx := range session.Uploaded {
		uploaded = append(uploaded, idx)
	}
	ok(c, gin.H{
		"upload_session_id": session.ID,
		"uploaded_chunks":   uploaded,
		"total_chunks":      session.TotalChunks,
	})
}

func (a *API) UploadComplete(c *gin.Context) {
	userID, okUser := middleware.UserIDFromContext(c)
	if !okUser {
		response.Error(c, 401, "unauthorized")
		return
	}
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	session, mergedPath, err := a.chunkUpload.Complete(c.Request.Context(), req.SessionID)
	if err != nil {
		response.Error(c, 2003, err.Error())
		return
	}
	if session.UserID != userID.String() {
		_ = os.Remove(mergedPath)
		response.Error(c, 403, "forbidden")
		return
	}

	result, err := a.uploadSvc.SaveUploadedFile(c.Request.Context(), services.LocalUploadInput{
		UserID:   userID,
		FilePath: mergedPath,
		Filename: session.Filename,
		FileSize: session.FileSize,
		Title:    session.Title,
		Desc:     session.Description,
		Type:     session.Type,
		Tags:     session.Tags,
		Hash:     session.Hash,
	}, a.maxVideoSize)
	if err != nil {
		_ = os.Remove(mergedPath)
		switch {
		case errors.Is(err, services.ErrHashMismatch):
			bad(c, "hash mismatch")
		case errors.Is(err, services.ErrFileTooLarge):
			bad(c, "file size exceeds limit")
		case errors.Is(err, services.ErrInvalidType):
			bad(c, "invalid type")
		case errors.Is(err, services.ErrHashTypeConflict):
			response.Error(c, 409, "hash exists with another video type")
		default:
			response.Error(c, 2004, err.Error())
		}
		return
	}

	if result.AlreadyExists {
		_ = os.Remove(mergedPath)
		_ = a.chunkUpload.Abort(req.SessionID)
		ok(c, gin.H{
			"exists":   true,
			"video_id": result.VideoID,
		})
		return
	}

	if session.Type == "short" && result.Enqueue {
		payload := queue.TranscodePayload{
			VideoID:      result.VideoID.String(),
			InputPath:    result.InputPath,
			OutputDir:    result.OutputDir,
			TargetFormat: result.TargetFormat,
		}
		if err := a.enqueuer.EnqueueTranscode(payload); err != nil {
			response.Error(c, 2005, err.Error())
			return
		}
	}
	if session.Type == "movie" || session.Type == "episode" {
		payload := queue.ScrapePayload{
			VideoID:  result.VideoID.String(),
			FilePath: result.InputPath,
			Filename: session.Filename,
			Type:     session.Type,
		}
		if session.Type == "movie" {
			if err := a.enqueuer.EnqueueScrapeMovie(payload); err != nil {
				response.Error(c, 2006, err.Error())
				return
			}
		} else {
			if err := a.enqueuer.EnqueueScrapeTV(payload); err != nil {
				response.Error(c, 2006, err.Error())
				return
			}
		}
	}
	_ = a.chunkUpload.Abort(req.SessionID)

	respStatus := result.Status
	if session.Type == "movie" || session.Type == "episode" {
		respStatus = "scraping"
	}
	c.JSON(http.StatusAccepted, gin.H{
		"code": 0,
		"msg":  "upload accepted",
		"data": gin.H{
			"video_id": result.VideoID,
			"status":   respStatus,
		},
	})
}

func (a *API) UploadAbort(c *gin.Context) {
	sessionID := strings.TrimSpace(c.Param("id"))
	if sessionID == "" {
		bad(c, "invalid session id")
		return
	}
	if err := a.chunkUpload.Abort(sessionID); err != nil && !os.IsNotExist(err) {
		response.Error(c, 2007, err.Error())
		return
	}
	ok(c, gin.H{"aborted": true})
}
