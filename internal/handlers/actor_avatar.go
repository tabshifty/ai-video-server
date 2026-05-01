package handlers

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

	if localPath, ok := resolveLocalActorAvatarPath(a.storageRoot, actorID); ok {
		if stat, statErr := os.Stat(localPath); statErr == nil {
			c.Header("Content-Length", strconv.FormatInt(stat.Size(), 10))
		}
		if mimeType := mime.TypeByExtension(strings.ToLower(filepath.Ext(localPath))); mimeType != "" {
			c.Header("Content-Type", mimeType)
		}
		c.Header("Cache-Control", "public, max-age=86400")
		c.File(localPath)
		return
	}

	raw := strings.TrimSpace(actor.AvatarURL)
	if stat, statErr := os.Stat(raw); statErr == nil && !stat.IsDir() {
		c.File(raw)
		return
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		c.Redirect(http.StatusTemporaryRedirect, raw)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"msg": "actor avatar not found"})
}

func resolveLocalActorAvatarPath(storageRoot string, actorID uuid.UUID) (string, bool) {
	if actorID == uuid.Nil {
		return "", false
	}
	matches, err := filepath.Glob(filepath.Join(storageRoot, "actors", actorID.String(), "avatar.*"))
	if err != nil || len(matches) == 0 {
		return "", false
	}
	for _, match := range matches {
		if stat, statErr := os.Stat(match); statErr == nil && !stat.IsDir() {
			return match, true
		}
	}
	return "", false
}
