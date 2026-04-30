package main

import (
	"testing"
	"time"

	"video-server/internal/config"
	"video-server/internal/services"
)

type captureAVScraperConfigurer struct {
	got services.AVScraperConfig
}

func (c *captureAVScraperConfigurer) ConfigureAVScraperConfig(cfg services.AVScraperConfig) {
	c.got = cfg
}

func TestConfigureAVScraperPassesThroughSharedSiteURLs(t *testing.T) {
	cfg := config.Config{
		AVScraperBaseURL:           "https://base.example",
		AVScraperUserAgent:         "VideoServerBot/1.0",
		AVScraperTimeout:           12 * time.Second,
		AVSiteURLs:                 map[string]string{"javdb": "https://javdb.example", "foo": "https://foo.example"},
		AVScraperJavDBCookie:       "cookie-javdb",
		AVScraperJavBusCookie:      "cookie-javbus",
		AVScraperThePornDBAPIToken: "token-123",
		AVScraperThePornDBNoHash:   true,
	}

	capture := &captureAVScraperConfigurer{}
	configureAVScraper(capture, cfg)

	if capture.got.BaseURL != cfg.AVScraperBaseURL {
		t.Fatalf("unexpected BaseURL: %s", capture.got.BaseURL)
	}
	if capture.got.UserAgent != cfg.AVScraperUserAgent {
		t.Fatalf("unexpected UserAgent: %s", capture.got.UserAgent)
	}
	if capture.got.Timeout != cfg.AVScraperTimeout {
		t.Fatalf("unexpected Timeout: %s", capture.got.Timeout)
	}
	if capture.got.JavDBCookie != cfg.AVScraperJavDBCookie {
		t.Fatalf("unexpected JavDBCookie: %s", capture.got.JavDBCookie)
	}
	if capture.got.JavBusCookie != cfg.AVScraperJavBusCookie {
		t.Fatalf("unexpected JavBusCookie: %s", capture.got.JavBusCookie)
	}
	if capture.got.ThePornDBAPIToken != cfg.AVScraperThePornDBAPIToken {
		t.Fatalf("unexpected ThePornDBAPIToken: %s", capture.got.ThePornDBAPIToken)
	}
	if !capture.got.ThePornDBNoHash {
		t.Fatalf("expected ThePornDBNoHash true")
	}
	if got := capture.got.SiteURLs["javdb"]; got != "https://javdb.example" {
		t.Fatalf("unexpected SiteURLs[javdb]: %s", got)
	}
	if got := capture.got.SiteURLs["foo"]; got != "https://foo.example" {
		t.Fatalf("unexpected SiteURLs[foo]: %s", got)
	}
	if len(capture.got.SiteURLs) != len(cfg.AVSiteURLs) {
		t.Fatalf("unexpected SiteURLs length: %d", len(capture.got.SiteURLs))
	}
}
