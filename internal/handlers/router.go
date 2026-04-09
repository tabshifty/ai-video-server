package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"video-server/internal/middleware"
	"video-server/internal/queue"
	"video-server/internal/repository"
	"video-server/internal/response"
	"video-server/internal/services"
)

// API bundles all HTTP handlers.
type API struct {
	repo          *repository.VideoRepository
	uploadSvc     *services.UploadService
	recSvc        *services.RecommendService
	scrapeSvc     *services.ScraperService
	appSvc        *services.AppService
	enqueuer      *queue.Enqueuer
	logger        *slog.Logger
	redis         *redis.Client
	jwtSecret     string
	accessTTL     time.Duration
	refreshTTL    time.Duration
	maxVideoSize  int64
	enableSwagger bool
}

func NewAPI(repo *repository.VideoRepository, uploadSvc *services.UploadService, recSvc *services.RecommendService, scrapeSvc *services.ScraperService, appSvc *services.AppService, enqueuer *queue.Enqueuer, logger *slog.Logger, redisClient *redis.Client, jwtSecret string, accessTTL, refreshTTL time.Duration, maxVideoSize int64, enableSwagger bool) *API {
	return &API{
		repo:          repo,
		uploadSvc:     uploadSvc,
		recSvc:        recSvc,
		scrapeSvc:     scrapeSvc,
		appSvc:        appSvc,
		enqueuer:      enqueuer,
		logger:        logger,
		redis:         redisClient,
		jwtSecret:     jwtSecret,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		maxVideoSize:  maxVideoSize,
		enableSwagger: enableSwagger,
	}
}

func (a *API) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", a.RegisterAuth)
			auth.POST("/login", a.LoginAuth)
			auth.POST("/refresh", a.RefreshAuth)
			auth.POST("/logout", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.LogoutAuth)
		}
		v1.POST("/upload/check", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.UploadCheck)
		v1.POST("/upload", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.Upload)
		v1.POST("/scrape", a.Scrape)
		v1.GET("/short/random", a.RandomShort)
		v1.GET("/recommend", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.Recommend)
		v1.POST("/actions", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.RecordAction)
		v1.GET("/videos/:id", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.VideoDetail)
		v1.POST("/history", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.RecordHistory)
		v1.GET("/history/continue", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.ContinueHistory)
		v1.DELETE("/history/:video_id", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.DeleteHistory)
		v1.POST("/videos/:id/like", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.ToggleLike)
		v1.POST("/videos/:id/favorite", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.ToggleFavorite)
		v1.POST("/videos/:id/dislike", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.ToggleDislike)
		v1.GET("/search", a.Search)
		v1.GET("/user/profile", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.UserProfile)
		v1.PUT("/user/profile", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.UpdateUserProfile)
		v1.GET("/user/uploaded-videos", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.UploadedVideos)
		v1.GET("/user/liked-videos", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.LikedVideos)
		v1.GET("/user/favorited-videos", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.FavoritedVideos)

		admin := v1.Group("/admin", middleware.AuthMiddleware(a.jwtSecret, a.redis), middleware.AdminRequired())
		{
			admin.POST("/scrape/preview", a.AdminScrapePreview)
			admin.PUT("/scrape/confirm", a.AdminScrapeConfirm)
		}
	}
	if a.enableSwagger {
		r.GET("/swagger/index.html", func(c *gin.Context) {
			c.File("docs/swagger/index.html")
		})
		r.GET("/swagger/openapi.json", func(c *gin.Context) {
			c.File("docs/swagger/openapi.json")
		})
	}
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func parsePage(raw string, fallback int) int {
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func parsePageSize(raw string, fallback int) int {
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	if v > 100 {
		return 100
	}
	return v
}

func parseTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		t := strings.TrimSpace(part)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func parseUUID(raw string) (uuid.UUID, bool) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func ok(c *gin.Context, data any) {
	response.JSON(c, 0, "", data)
}

func bad(c *gin.Context, msg string) {
	response.Error(c, 1, msg)
}
