package forumdownload

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-server/internal/middleware"
)

// Register exposes a separate machine credential; the ingestion token has no download rights.
func Register(r *gin.Engine, s *Service, token string, owner int64) {
	group := r.Group("/api/v1/integrations/telegram-forum", middleware.HermesTokenMiddleware(token))
	group.Use(func(c *gin.Context) {
		if s == nil || s.downloads == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "论坛下载功能未启用"})
			return
		}
		if parseActor(c.GetHeader("X-Telegram-User-ID"), owner) != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), RequestTimeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	group.GET("/posts/:id", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			writeError(c, errInvalid)
			return
		}
		p, err := loadPost(c.Request.Context(), s.pool, id, "")
		respond(c, p, err)
	})
	group.GET("/posts", func(c *gin.Context) {
		tid := c.Query("tid")
		n, err := strconv.ParseInt(tid, 10, 64)
		if err != nil || n <= 0 {
			writeError(c, errInvalid)
			return
		}
		p, err := loadPost(c.Request.Context(), s.pool, uuid.Nil, tid)
		respond(c, p, err)
	})
	group.POST("/operations", func(c *gin.Context) {
		var req Request
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64*1024)
		if c.ShouldBindJSON(&req) != nil {
			writeError(c, errInvalid)
			return
		}
		if req.ActorID != owner {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作"})
			return
		}
		result, err := s.submit(c.Request.Context(), req)
		respond(c, result, err)
	})
	group.GET("/receipts", func(c *gin.Context) { items, err := s.receipts(c.Request.Context()); respond(c, items, err) })
	group.POST("/receipts/:id/ack", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			writeError(c, errInvalid)
			return
		}
		var in struct {
			Token uuid.UUID `json:"token"`
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024)
		if c.ShouldBindJSON(&in) != nil || in.Token == uuid.Nil {
			writeError(c, errInvalid)
			return
		}
		respond(c, gin.H{"acknowledged": true}, s.ack(c.Request.Context(), id, in.Token))
	})
}

func respond(c *gin.Context, data any, err error) {
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "成功", "data": data})
}
func writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "服务暂时不可用，请稍后重试"
	switch {
	case errors.Is(err, errInvalid):
		status = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, errConflict):
		status = http.StatusConflict
		message = err.Error()
	case errors.Is(err, errNotFound):
		status = http.StatusNotFound
		message = err.Error()
	case errors.Is(err, context.DeadlineExceeded):
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, gin.H{"code": status, "msg": message, "data": nil})
}
