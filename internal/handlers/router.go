package handlers

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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
	repo             *repository.VideoRepository
	uploadSvc        *services.UploadService
	chunkUpload      *services.ChunkUploadService
	recSvc           *services.RecommendService
	scrapeSvc        *services.ScraperService
	appSvc           *services.AppService
	imageSvc         *services.ImageService
	subtitleSvc      *services.SubtitleService
	iptvSvc          iptvService
	enqueuer         *queue.Enqueuer
	logger           *slog.Logger
	redis            *redis.Client
	redisAddr        string
	redisPassword    string
	asynqQueue       string
	jwtSecret        string
	playSignSecret   string
	playSignTTL      time.Duration
	accessTTL        time.Duration
	refreshTTL       time.Duration
	maxVideoSize     int64
	storageRoot      string
	uploadTempDir    string
	serverLogPath    string
	adminWebDistPath string
	enableSwagger    bool
}

func NewAPI(repo *repository.VideoRepository, uploadSvc *services.UploadService, chunkUpload *services.ChunkUploadService, recSvc *services.RecommendService, scrapeSvc *services.ScraperService, appSvc *services.AppService, imageSvc *services.ImageService, subtitleSvc *services.SubtitleService, enqueuer *queue.Enqueuer, logger *slog.Logger, redisClient *redis.Client, redisAddr, redisPassword, asynqQueue, jwtSecret, playSignSecret string, accessTTL, refreshTTL time.Duration, maxVideoSize int64, storageRoot, uploadTempDir, serverLogPath, adminWebDistPath string, enableSwagger bool) *API {
	return &API{
		repo:             repo,
		uploadSvc:        uploadSvc,
		chunkUpload:      chunkUpload,
		recSvc:           recSvc,
		scrapeSvc:        scrapeSvc,
		appSvc:           appSvc,
		imageSvc:         imageSvc,
		subtitleSvc:      subtitleSvc,
		iptvSvc:          services.NewIPTVService(repo, nil),
		enqueuer:         enqueuer,
		logger:           logger,
		redis:            redisClient,
		redisAddr:        redisAddr,
		redisPassword:    redisPassword,
		asynqQueue:       asynqQueue,
		jwtSecret:        jwtSecret,
		playSignSecret:   playSignSecret,
		playSignTTL:      10 * time.Minute,
		accessTTL:        accessTTL,
		refreshTTL:       refreshTTL,
		maxVideoSize:     maxVideoSize,
		storageRoot:      storageRoot,
		uploadTempDir:    uploadTempDir,
		serverLogPath:    serverLogPath,
		adminWebDistPath: adminWebDistPath,
		enableSwagger:    enableSwagger,
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
		tvAuth := v1.Group("/tv-auth")
		{
			tvAuth.POST("/sessions", a.CreateTVAuthSession)
			tvAuth.GET("/sessions/:session_id", a.GetTVAuthSession)
			tvAuth.POST("/sessions/:session_id/approve", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.ApproveTVAuthSession)
			tvAuth.POST("/sessions/:session_id/deny", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.DenyTVAuthSession)
		}
		v1.POST("/upload/check", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.UploadCheck)
		v1.POST("/upload/init", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.UploadInit)
		v1.PUT("/upload/chunk", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.UploadChunk)
		v1.POST("/upload/complete", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.UploadComplete)
		v1.DELETE("/upload/session/:id", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.UploadAbort)
		v1.POST("/upload", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.Upload)
		v1.POST("/scrape", a.Scrape)
		v1.GET("/short/random", a.RandomShort)
		v1.GET("/short/discover", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.ShortDiscover)
		v1.GET("/tv/home", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.TVHome)
		v1.GET("/tv/search", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.TVSearch)
		v1.GET("/tv/catalog", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.TVCatalogWall)
		v1.GET("/tv/series/:id", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.TVSeriesDetail)
		v1.GET("/tv/iptv/channels", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.TVIPTVChannels)
		v1.GET("/tv/series/:id/poster", a.TVSeriesPoster)
		v1.GET("/tv/series/:id/backdrop", a.TVSeriesBackdrop)
		v1.GET("/tv/series/:id/seasons/:season/episodes/:episode/still", a.TVEpisodeStill)
		v1.GET("/image-collections", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.AppImageCollections)
		v1.GET("/image-collections/:id", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.AppImageCollectionDetail)
		v1.GET("/images/:id/view", a.AppImageView)
		v1.GET("/actors/:id", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.ActorDetail)
		v1.GET("/actors/:id/avatar", a.ActorAvatar)
		v1.GET("/recommend", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.Recommend)
		v1.POST("/actions", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.RecordAction)
		v1.GET("/videos/:id", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.VideoDetail)
		v1.GET("/videos/:id/source", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.VideoSource)
		v1.GET("/videos/:id/source/signed", a.VideoSourceSigned)
		v1.GET("/videos/:id/thumbnail", a.VideoThumbnail)
		v1.GET("/videos/:id/subtitles/:subtitle_id/file", middleware.AuthMiddleware(a.jwtSecret, a.redis), a.VideoSubtitleFile)
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
			admin.GET("/events/ws", a.AdminEventsStream)
			admin.GET("/stats", a.AdminStats)
			admin.GET("/iptv/playlist", a.AdminIPTVPlaylist)
			admin.POST("/iptv/playlist/upload", a.AdminIPTVUploadPlaylist)
			admin.PUT("/iptv/playlist/source", a.AdminIPTVSaveSource)
			admin.POST("/iptv/playlist/refresh", a.AdminIPTVRefreshPlaylist)
			admin.GET("/videos", a.AdminVideos)
			admin.GET("/video-tags", a.AdminVideoTags)
			admin.GET("/video-tags/popular", a.AdminPopularVideoTags)
			admin.POST("/videos/batch-delete", a.AdminBatchDeleteVideos)
			admin.GET("/videos/:id", a.AdminVideoDetail)
			admin.GET("/videos/:id/subtitles", a.AdminVideoSubtitles)
			admin.POST("/videos/:id/subtitles/upload", a.AdminUploadVideoSubtitle)
			admin.POST("/videos/:id/subtitles/scan", a.AdminRescanVideoSubtitles)
			admin.PUT("/videos/:id/subtitles/:subtitle_id", a.AdminUpdateVideoSubtitle)
			admin.DELETE("/videos/:id/subtitles/:subtitle_id", a.AdminDeleteVideoSubtitle)
			admin.GET("/videos/:id/play-url", a.AdminVideoPlayURL)
			admin.POST("/videos/:id/thumbnail/capture", a.AdminCaptureVideoThumbnail)
			admin.PUT("/videos/:id", a.AdminUpdateVideo)
			admin.DELETE("/videos/:id", a.AdminDeleteVideo)
			admin.GET("/tv/series", a.AdminTVSeries)
			admin.GET("/tv/series/:id", a.AdminTVSeriesDetail)
			admin.POST("/tv/series", a.AdminCreateTVSeries)
			admin.PUT("/tv/series/:id", a.AdminUpdateTVSeries)
			admin.DELETE("/tv/series/:id", a.AdminDeleteTVSeries)
			admin.POST("/tv/series/:id/seasons", a.AdminCreateTVSeason)
			admin.PUT("/tv/seasons/:id", a.AdminUpdateTVSeason)
			admin.DELETE("/tv/seasons/:id", a.AdminDeleteTVSeason)
			admin.POST("/tv/seasons/:id/episodes", a.AdminCreateTVEpisode)
			admin.PUT("/tv/episodes/:id", a.AdminUpdateTVEpisode)
			admin.DELETE("/tv/episodes/:id", a.AdminDeleteTVEpisode)
			admin.POST("/videos/:id/retranscode", a.AdminRetranscodeVideo)
			admin.GET("/users", a.AdminUsers)
			admin.POST("/users", a.AdminCreateUser)
			admin.PUT("/users/:id/role", a.AdminUpdateUserRole)
			admin.GET("/actors", a.AdminActors)
			admin.POST("/actors", a.AdminCreateActor)
			admin.POST("/actors/scrape/preview", a.AdminActorScrapePreview)
			admin.PUT("/actors/:id", a.AdminUpdateActor)
			admin.GET("/collections", a.AdminCollections)
			admin.POST("/collections", a.AdminCreateCollection)
			admin.PUT("/collections/:id", a.AdminUpdateCollection)
			admin.DELETE("/collections/:id", a.AdminDeleteCollection)
			admin.POST("/images/upload", a.AdminUploadImages)
			admin.POST("/images/check", a.AdminImageCheck)
			admin.GET("/images", a.AdminImages)
			admin.GET("/images/:id", a.AdminImageDetail)
			admin.PUT("/images/:id", a.AdminUpdateImage)
			admin.DELETE("/images/:id", a.AdminDeleteImage)
			admin.GET("/images/:id/view", a.AdminImageView)
			admin.GET("/image-collections", a.AdminImageCollections)
			admin.POST("/image-collections", a.AdminCreateImageCollection)
			admin.PUT("/image-collections/:id", a.AdminUpdateImageCollection)
			admin.DELETE("/image-collections/:id", a.AdminDeleteImageCollection)
			admin.GET("/tasks", a.AdminTasks)
			admin.POST("/system/cleanup", a.AdminSystemCleanup)
			admin.GET("/system/logs", a.AdminSystemLogs)
			admin.POST("/scrape/preview", a.AdminScrapePreview)
			admin.PUT("/scrape/confirm", a.AdminScrapeConfirm)
			admin.PUT("/scrape/skip", a.AdminScrapeSkip)
			admin.GET("/av-scrape/config", a.AdminAVScrapeConfig)
			admin.PUT("/av-scrape/config", a.AdminAVScrapeConfig)
			admin.POST("/av-scrape/preview", a.AdminAVScrapePreview)
			admin.PUT("/av-scrape/confirm", a.AdminAVScrapeConfirm)
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

	// Serve built admin frontend from Go server at /admin.
	mountAdminStatic(r, a.adminWebDistPath)
}

// mountAdminStatic mounts /admin and /admin/assets when adminDist points to a real directory.
// Path is taken as-is (caller decides absolute vs relative) per [[家用部署机绝对路径契约]].
func mountAdminStatic(r *gin.Engine, adminDist string) {
	if adminDist == "" {
		return
	}
	st, err := os.Stat(adminDist)
	if err != nil || !st.IsDir() {
		return
	}
	r.Static("/admin/assets", filepath.Join(adminDist, "assets"))
	serveAdminIndex := func(c *gin.Context) {
		c.File(filepath.Join(adminDist, "index.html"))
	}
	r.GET("/admin", serveAdminIndex)
	r.GET("/admin/", serveAdminIndex)
	r.NoRoute(func(c *gin.Context) {
		reqPath := strings.TrimSpace(c.Request.URL.Path)
		if strings.HasPrefix(reqPath, "/admin/assets/") {
			c.String(http.StatusNotFound, "404 page not found")
			return
		}
		if strings.HasPrefix(reqPath, "/admin/") {
			serveAdminIndex(c)
			return
		}
		c.String(http.StatusNotFound, "404 page not found")
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

func parseUUIDCSV(raw string) ([]uuid.UUID, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	return parseUUIDStrings(strings.Split(trimmed, ","))
}

func parseInt64(raw string) (int64, bool) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value <= 0 {
		return 0, false
	}
	return value, true
}

func ok(c *gin.Context, data any) {
	response.JSON(c, 0, "", data)
}

func bad(c *gin.Context, msg string) {
	response.Error(c, 1, msg)
}
