package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"video-server/internal/response"
	"video-server/internal/utils"
)

const (
	ctxUserIDKey = "auth_user_id"
	ctxRoleKey   = "auth_role"
)

// AuthMiddleware validates access token and stores identity in request context.
func AuthMiddleware(secret string, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := bearerToken(c.GetHeader("Authorization"))
		if err != nil {
			response.Error(c, 401, "missing or invalid authorization header")
			c.Abort()
			return
		}

		claims, err := utils.ParseAndValidateToken(secret, token, utils.TokenTypeAccess)
		if err != nil {
			response.Error(c, 401, "invalid access token")
			c.Abort()
			return
		}

		if claims.ID != "" {
			blocked, err := redisClient.Exists(c.Request.Context(), accessBlacklistKey(claims.ID)).Result()
			if err != nil {
				response.Error(c, 500, "failed to validate token")
				c.Abort()
				return
			}
			if blocked > 0 {
				response.Error(c, 401, "token revoked")
				c.Abort()
				return
			}
		}

		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			response.Error(c, 401, "invalid user id in token")
			c.Abort()
			return
		}
		c.Set(ctxUserIDKey, uid)
		c.Set(ctxRoleKey, claims.Role)
		c.Next()
	}
}

// AdminRequired allows only admin role.
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(ctxRoleKey)
		roleStr, _ := role.(string)
		if roleStr != "admin" {
			response.Error(c, 403, "admin only")
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserIDFromContext extracts authenticated user id.
func UserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, exists := c.Get(ctxUserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	uid, ok := v.(uuid.UUID)
	return uid, ok
}

// RoleFromContext extracts authenticated role.
func RoleFromContext(c *gin.Context) (string, bool) {
	v, exists := c.Get(ctxRoleKey)
	if !exists {
		return "", false
	}
	role, ok := v.(string)
	return role, ok
}

// RevokeAccessToken blacklists an access token jti until it expires.
func RevokeAccessToken(c *gin.Context, redisClient *redis.Client, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return redisClient.Set(c.Request.Context(), accessBlacklistKey(jti), "1", ttl).Err()
}

func accessBlacklistKey(jti string) string {
	return "auth:blacklist:access:" + jti
}

func refreshSessionKey(jti string) string {
	return "auth:refresh:" + jti
}

func StoreRefreshSession(ctx *gin.Context, redisClient *redis.Client, jti string, userID uuid.UUID, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return fmt.Errorf("refresh token already expired")
	}
	return redisClient.Set(ctx.Request.Context(), refreshSessionKey(jti), userID.String(), ttl).Err()
}

func ValidateRefreshSession(ctx *gin.Context, redisClient *redis.Client, jti string, userID uuid.UUID) error {
	v, err := redisClient.Get(ctx.Request.Context(), refreshSessionKey(jti)).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("refresh token revoked")
		}
		return fmt.Errorf("read refresh session: %w", err)
	}
	if v != userID.String() {
		return fmt.Errorf("refresh token subject mismatch")
	}
	return nil
}

func RevokeRefreshSession(ctx *gin.Context, redisClient *redis.Client, jti string) error {
	return redisClient.Del(ctx.Request.Context(), refreshSessionKey(jti)).Err()
}

func bearerToken(header string) (string, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", fmt.Errorf("invalid bearer header")
	}
	return strings.TrimSpace(parts[1]), nil
}
