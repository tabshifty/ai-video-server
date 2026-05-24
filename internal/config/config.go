package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	HTTPAddr                   string
	Mode                       string
	PostgresDSN                string
	RedisAddr                  string
	RedisPassword              string
	ServerLogPath              string
	JWTSecret                  string
	PlayURLSignSecret          string
	StorageRoot                string
	PosterStoragePath          string
	UploadTempDir              string
	TMDBAPIKey                 string
	TMDBBaseURL                string
	AVScraperBaseURL           string
	AVScraperUserAgent         string
	AVSiteURLJavDB             string
	AVSiteURLJavBus            string
	AVSiteURLJavLibrary        string
	AVSiteURLThePornDB         string
	AVSiteURLs                 map[string]string
	AVScraperJavDBCookie       string
	AVScraperJavBusCookie      string
	AVScraperThePornDBAPIToken string
	AVScraperThePornDBNoHash   bool
	MaxTranscodeWorkers        int
	AsynqQueue                 string
	TranscodeTaskTimeout       time.Duration
	MaxVideoSize               int64
	EnableSwagger              bool
	TMDBTimeout                time.Duration
	AVScraperTimeout           time.Duration
	AccessTokenTTL             time.Duration
	RefreshTokenTTL            time.Duration
	TranslationAPIURL          string
	TranslationAPIKey          string
	TranslationModel           string
	TranslationTimeout         time.Duration
}

// Load returns validated application config from environment.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:                   getEnv("HTTP_ADDR", ":8080"),
		Mode:                       getEnv("APP_MODE", "server"),
		PostgresDSN:                os.Getenv("POSTGRES_DSN"),
		RedisAddr:                  getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:              os.Getenv("REDIS_PASSWORD"),
		ServerLogPath:              getEnv("SERVER_LOG_PATH", "./.run/server.log"),
		JWTSecret:                  os.Getenv("JWT_SECRET"),
		PlayURLSignSecret:          os.Getenv("PLAY_URL_SIGN_SECRET"),
		StorageRoot:                getEnv("STORAGE_ROOT", "./storage"),
		PosterStoragePath:          getEnv("POSTER_STORAGE_PATH", "./storage/posters"),
		UploadTempDir:              getEnv("UPLOAD_TEMP_DIR", "./tmp/uploads"),
		TMDBAPIKey:                 os.Getenv("TMDB_API_KEY"),
		TMDBBaseURL:                getEnv("TMDB_BASE_URL", "https://api.themoviedb.org/3"),
		AVScraperBaseURL:           getEnv("AV_SCRAPER_BASE_URL", "https://javdb.com"),
		AVScraperUserAgent:         getEnv("AV_SCRAPER_USER_AGENT", "Mozilla/5.0 (compatible; VideoServerBot/1.0; +https://example.invalid/bot)"),
		AVSiteURLJavDB:             firstNonEmptyEnv("AV_SITE_URL_JAVDB", "AV_SCRAPER_BASE_URL"),
		AVSiteURLJavBus:            os.Getenv("AV_SITE_URL_JAVBUS"),
		AVSiteURLJavLibrary:        os.Getenv("AV_SITE_URL_JAVLIBRARY"),
		AVSiteURLThePornDB:         os.Getenv("AV_SITE_URL_THEPORNDB"),
		AVSiteURLs:                 loadAVSiteURLs(),
		AVScraperJavDBCookie:       os.Getenv("AV_SCRAPER_JAVDB_COOKIE"),
		AVScraperJavBusCookie:      os.Getenv("AV_SCRAPER_JAVBUS_COOKIE"),
		AVScraperThePornDBAPIToken: os.Getenv("AV_SCRAPER_THEPORNDB_API_TOKEN"),
		AVScraperThePornDBNoHash:   getBoolEnv("AV_SCRAPER_THEPORNDB_NO_HASH", false),
		AsynqQueue:                 getEnv("ASYNQ_QUEUE", "transcode"),
		MaxTranscodeWorkers:        getIntEnv("MAX_TRANSCODE_WORKERS", 2),
		TranscodeTaskTimeout:       time.Duration(getIntEnv("TRANSCODE_TASK_TIMEOUT_MINUTES", 360)) * time.Minute,
		MaxVideoSize:               getInt64Env("MAX_VIDEO_SIZE", 2*1024*1024*1024),
		EnableSwagger:              getBoolEnv("ENABLE_SWAGGER", false),
		TMDBTimeout:                time.Duration(getIntEnv("TMDB_TIMEOUT_SECONDS", 10)) * time.Second,
		AVScraperTimeout:           time.Duration(getIntEnv("AV_SCRAPER_TIMEOUT_SECONDS", 10)) * time.Second,
		AccessTokenTTL:             time.Duration(getIntEnv("ACCESS_TOKEN_TTL_HOURS", 87600)) * time.Hour,
		RefreshTokenTTL:            time.Duration(getIntEnv("REFRESH_TOKEN_TTL_HOURS", 168)) * time.Hour,
		TranslationAPIURL:          strings.TrimSuffix(strings.TrimSpace(os.Getenv("TRANSLATION_API_URL")), "/"),
		TranslationAPIKey:          os.Getenv("TRANSLATION_API_KEY"),
		TranslationModel:           getEnv("TRANSLATION_MODEL", "HY-MT1.5-1.8B"),
		TranslationTimeout:         time.Duration(getIntEnv("TRANSLATION_TIMEOUT_SECONDS", 15)) * time.Second,
	}

	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("POSTGRES_DSN is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	if strings.TrimSpace(cfg.PlayURLSignSecret) == "" {
		cfg.PlayURLSignSecret = cfg.JWTSecret
	}
	if cfg.AVSiteURLs == nil {
		cfg.AVSiteURLs = map[string]string{}
	}
	if cfg.AVSiteURLJavDB != "" {
		cfg.AVSiteURLs["javdb"] = strings.TrimSuffix(strings.TrimSpace(cfg.AVSiteURLJavDB), "/")
	}
	cfg.AVSiteURLJavDB = cfg.AVSiteURLs["javdb"]

	return cfg, nil
}

func loadAVSiteURLs() map[string]string {
	urls := map[string]string{}
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if !ok || !strings.HasPrefix(key, "AV_SITE_URL_") {
			continue
		}
		site := normalizeAVSiteURLKey(strings.TrimPrefix(key, "AV_SITE_URL_"))
		value = strings.TrimSuffix(strings.TrimSpace(value), "/")
		if site == "" || value == "" {
			continue
		}
		urls[site] = value
	}
	if fallback := strings.TrimSuffix(strings.TrimSpace(firstNonEmptyEnv("AV_SITE_URL_JAVDB", "AV_SCRAPER_BASE_URL")), "/"); fallback != "" {
		urls["javdb"] = fallback
	}
	return urls
}

func normalizeAVSiteURLKey(key string) string {
	raw := strings.ToLower(strings.TrimSpace(key))
	compact := strings.ReplaceAll(raw, "-", "")
	compact = strings.ReplaceAll(compact, "_", "")
	switch compact {
	case "javdb":
		return "javdb"
	case "airavcc":
		return "airav_cc"
	case "mdtv", "mdtvcom":
		return "mdtv.com"
	case "javbus":
		return "javbus"
	case "javlibrary", "javlib":
		return "javlibrary"
	case "theporndb", "porndb", "tpdb":
		return "theporndb"
	default:
		return raw
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}

func getIntEnv(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func getInt64Env(key string, fallback int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func getBoolEnv(key string, fallback bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
