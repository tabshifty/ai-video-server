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
	HTTPAddr            string
	Mode                string
	PostgresDSN         string
	RedisAddr           string
	RedisPassword       string
	ServerLogPath       string
	JWTSecret           string
	PlayURLSignSecret   string
	StorageRoot         string
	PosterStoragePath   string
	UploadTempDir       string
	TMDBAPIKey          string
	TMDBBaseURL         string
	MaxTranscodeWorkers int
	AsynqQueue          string
	MaxVideoSize        int64
	EnableSwagger       bool
	TMDBTimeout         time.Duration
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
}

// Load returns validated application config from environment.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:            getEnv("HTTP_ADDR", ":8080"),
		Mode:                getEnv("APP_MODE", "server"),
		PostgresDSN:         os.Getenv("POSTGRES_DSN"),
		RedisAddr:           getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:       os.Getenv("REDIS_PASSWORD"),
		ServerLogPath:       getEnv("SERVER_LOG_PATH", "./.run/server.log"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		PlayURLSignSecret:   os.Getenv("PLAY_URL_SIGN_SECRET"),
		StorageRoot:         getEnv("STORAGE_ROOT", "./storage"),
		PosterStoragePath:   getEnv("POSTER_STORAGE_PATH", "./storage/posters"),
		UploadTempDir:       getEnv("UPLOAD_TEMP_DIR", "./tmp/uploads"),
		TMDBAPIKey:          os.Getenv("TMDB_API_KEY"),
		TMDBBaseURL:         getEnv("TMDB_BASE_URL", "https://api.themoviedb.org/3"),
		AsynqQueue:          getEnv("ASYNQ_QUEUE", "transcode"),
		MaxTranscodeWorkers: getIntEnv("MAX_TRANSCODE_WORKERS", 2),
		MaxVideoSize:        getInt64Env("MAX_VIDEO_SIZE", 2*1024*1024*1024),
		EnableSwagger:       getBoolEnv("ENABLE_SWAGGER", false),
		TMDBTimeout:         time.Duration(getIntEnv("TMDB_TIMEOUT_SECONDS", 10)) * time.Second,
		AccessTokenTTL:      time.Duration(getIntEnv("ACCESS_TOKEN_TTL_HOURS", 24)) * time.Hour,
		RefreshTokenTTL:     time.Duration(getIntEnv("REFRESH_TOKEN_TTL_HOURS", 168)) * time.Hour,
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

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
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
