package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"video-server/internal/middleware"
	"video-server/internal/response"
)

const collectionCleanupAudience = "admin-empty-collection-cleanup"

type collectionCleanupClaims struct {
	Kind       string      `json:"kind"`
	Candidates []uuid.UUID `json:"candidates"`
	jwt.RegisteredClaims
}

func (a *API) signCollectionCleanup(kind string, user uuid.UUID, ids []uuid.UUID) (string, error) {
	claims := collectionCleanupClaims{Kind: kind, Candidates: ids, RegisteredClaims: jwt.RegisteredClaims{
		Subject: user.String(), Audience: jwt.ClaimStrings{collectionCleanupAudience},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
	}}
	// 独立签名用途，预览凭据不能被作为登录令牌使用。
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("collection-cleanup:" + a.jwtSecret))
}

func (a *API) parseCollectionCleanup(raw, kind string, user uuid.UUID) ([]uuid.UUID, error) {
	claims := &collectionCleanupClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		return []byte("collection-cleanup:" + a.jwtSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithAudience(collectionCleanupAudience), jwt.WithSubject(user.String()), jwt.WithExpirationRequired())
	if err != nil || !token.Valid || claims.Kind != kind || len(claims.Candidates) == 0 {
		return nil, fmt.Errorf("清理预览无效或已过期，请重新预览")
	}
	return claims.Candidates, nil
}

func (a *API) previewEmptyCollections(c *gin.Context, kind string) {
	user, valid := middleware.UserIDFromContext(c)
	if !valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
	defer cancel()
	release, err := a.repo.LockCollectionReferenceFiles(ctx)
	if err != nil {
		response.Error(c, 1093, err.Error())
		return
	}
	defer release()
	ids, err := a.repo.PreviewEmptyCollections(ctx, kind)
	if err != nil {
		response.Error(c, 1093, "预览空合集失败，请稍后重试")
		return
	}
	ids, err = a.excludeUploadCollectionReferences(ctx, kind, ids)
	if err != nil {
		response.Error(c, 1093, "无法读取上传会话引用，请稍后重试")
		return
	}
	preview := ""
	if len(ids) > 0 {
		preview, err = a.signCollectionCleanup(kind, user, ids)
		if err != nil {
			response.Error(c, 1093, "生成清理预览失败")
			return
		}
	}
	ok(c, gin.H{"candidate_count": len(ids), "preview_token": preview})
}

func (a *API) cleanupEmptyCollections(c *gin.Context, kind string) {
	user, valid := middleware.UserIDFromContext(c)
	if !valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var req struct {
		PreviewToken string `json:"preview_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "清理请求格式无效")
		return
	}
	ids, err := a.parseCollectionCleanup(req.PreviewToken, kind, user)
	if err != nil {
		bad(c, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
	defer cancel()
	release, err := a.repo.LockCollectionReferenceFiles(ctx)
	if err != nil {
		response.Error(c, 1094, err.Error())
		return
	}
	defer release()
	eligible, err := a.excludeUploadCollectionReferences(ctx, kind, ids)
	if err != nil {
		response.Error(c, 1094, "无法读取上传会话引用，清理未执行")
		return
	}
	result, err := a.repo.CleanupEmptyCollections(ctx, kind, eligible)
	if err != nil {
		response.Error(c, 1094, "清理未完成，本次事务已回滚或结果未确认，请刷新列表后重新预览")
		return
	}
	result.SkippedCount += len(ids) - len(eligible)
	ok(c, result)
}

func (a *API) excludeUploadCollectionReferences(ctx context.Context, kind string, ids []uuid.UUID) ([]uuid.UUID, error) {
	refs, err := a.chunkUpload.CollectionReferences(ctx, kind)
	if err != nil {
		return nil, err
	}
	eligible := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if !refs[id] {
			eligible = append(eligible, id)
		}
	}
	return eligible, nil
}

// AdminPreviewEmptyCollections previews all eligible video collections.
func (a *API) AdminPreviewEmptyCollections(c *gin.Context) { a.previewEmptyCollections(c, "video") }

// AdminCleanupEmptyCollections deletes only signed video preview candidates.
func (a *API) AdminCleanupEmptyCollections(c *gin.Context) { a.cleanupEmptyCollections(c, "video") }

// AdminPreviewEmptyImageCollections previews all eligible image collections.
func (a *API) AdminPreviewEmptyImageCollections(c *gin.Context) {
	a.previewEmptyCollections(c, "image")
}

// AdminCleanupEmptyImageCollections deletes only signed image preview candidates.
func (a *API) AdminCleanupEmptyImageCollections(c *gin.Context) {
	a.cleanupEmptyCollections(c, "image")
}
