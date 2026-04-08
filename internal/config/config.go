package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	HTTPAddr            string
	Mode                string
	PostgresDSN         string
	RedisAddr           string
	RedisPassword       string
	JWTSecret           string
	StorageRoot         string
	UploadTempDir       string
	TMDBAPIKey          string
	TMDBBaseURL         string
	MaxTranscodeWorkers int
	AsynqQueue          string
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
		JWTSecret:           os.Getenv("JWT_SECRET"),
		StorageRoot:         getEnv("STORAGE_ROOT", "./storage"),
		UploadTempDir:       getEnv("UPLOAD_TEMP_DIR", "./tmp/uploads"),
		TMDBAPIKey:          os.Getenv("TMDB_API_KEY"),
		TMDBBaseURL:         getEnv("TMDB_BASE_URL", "https://api.themoviedb.org/3"),
		AsynqQueue:          getEnv("ASYNQ_QUEUE", "transcode"),
		MaxTranscodeWorkers: getIntEnv("MAX_TRANSCODE_WORKERS", 2),
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
