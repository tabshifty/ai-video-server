package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/repository"
)

func (a *API) ActorAvatar(c *gin.Context) {
	actorID, okID := parseUUID(c.Param("id"))
	if !okID {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid actor id"})
		return
	}

	actor, err := a.repo.GetActorByID(c.Request.Context(), actorID)
	if err != nil {
		if repository.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"msg": "actor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "query actor failed"})
		return
	}

	for _, localPath := range localActorAvatarPaths(a.storageRoot, actorID) {
		file, info, err := openLocalImageFile(c.Request.Context(), localPath)
		if err != nil {
			if isLocalImageNotFound(err) {
				continue
			}
			writeLocalImageOpenError(c, err, "actor avatar not found", "actor avatar temporarily unavailable")
			return
		}
		defer file.Close()
		if mimeType := strings.TrimSpace(fileTypeFromPath(localPath)); mimeType != "" {
			c.Header("Content-Type", mimeType)
		}
		c.Header("Cache-Control", "public, max-age=86400")
		serveOpenedLocalImage(c, localPath, file, info)
		return
	}

	raw := strings.TrimSpace(actor.AvatarURL)
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		c.Redirect(http.StatusTemporaryRedirect, raw)
		return
	}
	if raw != "" {
		file, info, err := openLocalImageFile(c.Request.Context(), raw)
		if err != nil {
			if isLocalImageNotFound(err) {
				c.JSON(http.StatusNotFound, gin.H{"msg": "actor avatar not found"})
				return
			}
			writeLocalImageOpenError(c, err, "actor avatar not found", "actor avatar temporarily unavailable")
			return
		}
		defer file.Close()
		serveOpenedLocalImage(c, raw, file, info)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"msg": "actor avatar not found"})
}

func localActorAvatarPaths(storageRoot string, actorID uuid.UUID) []string {
	if actorID == uuid.Nil {
		return nil
	}
	dir := filepath.Join(storageRoot, "actors", actorID.String())
	return []string{
		filepath.Join(dir, "avatar.jpg"),
		filepath.Join(dir, "avatar.png"),
		filepath.Join(dir, "avatar.webp"),
		filepath.Join(dir, "avatar.gif"),
	}
}

func fileTypeFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return ""
	}
}
