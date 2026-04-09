package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"

	"video-server/internal/response"
)

func (a *API) UploadCheck(c *gin.Context) {
	var req struct {
		Hash     string `json:"hash"`
		FileSize int64  `json:"file_size"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	req.Hash = strings.TrimSpace(req.Hash)
	if !isSHA256Hex(req.Hash) {
		bad(c, "invalid hash")
		return
	}
	if req.FileSize <= 0 {
		bad(c, "invalid file_size")
		return
	}

	videoID, exists, err := a.repo.FindVideoByHash(c.Request.Context(), req.Hash, req.FileSize)
	if err != nil {
		response.Error(c, 2, err.Error())
		return
	}
	if exists {
		ok(c, gin.H{"exists": true, "video_id": videoID})
		return
	}
	ok(c, gin.H{"exists": false})
}
