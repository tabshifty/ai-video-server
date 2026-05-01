package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

const avScraperConfigKey = "av_scraper_config"

func (r *VideoRepository) GetAVScraperConfig(ctx context.Context) (models.AVScraperSiteConfig, error) {
	var raw []byte
	err := r.pool.QueryRow(ctx, `SELECT value FROM admin_settings WHERE key = $1`, avScraperConfigKey).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.AVScraperSiteConfig{}, nil
		}
		return models.AVScraperSiteConfig{}, fmt.Errorf("get av scraper config: %w", err)
	}
	var cfg models.AVScraperSiteConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return models.AVScraperSiteConfig{}, fmt.Errorf("decode av scraper config: %w", err)
	}
	cfg.PosterCropConfigured = true
	return cfg, nil
}

func (r *VideoRepository) UpsertAVScraperConfig(ctx context.Context, cfg models.AVScraperSiteConfig) error {
	payload, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode av scraper config: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
INSERT INTO admin_settings(key, value, updated_at)
VALUES ($1, $2::jsonb, NOW())
ON CONFLICT (key)
DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
`, avScraperConfigKey, payload)
	if err != nil {
		return fmt.Errorf("upsert av scraper config: %w", err)
	}
	return nil
}
