package config

import (
	"os"
	"testing"
)

func TestLoadIncludesAVSiteOverridesAndTokens(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://user:pass@127.0.0.1:5432/app?sslmode=disable")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("AV_SCRAPER_BASE_URL", "https://javdb.example")
	t.Setenv("AV_SITE_URL_JAVBUS", "https://javbus.example")
	t.Setenv("AV_SITE_URL_JAVLIBRARY", "https://javlibrary.example")
	t.Setenv("AV_SITE_URL_THEPORNDB", "https://tpdb.example")
	t.Setenv("AV_SCRAPER_JAVDB_COOKIE", "cookie-javdb")
	t.Setenv("AV_SCRAPER_JAVBUS_COOKIE", "cookie-javbus")
	t.Setenv("AV_SCRAPER_THEPORNDB_API_TOKEN", "token-123")
	t.Setenv("AV_SCRAPER_THEPORNDB_NO_HASH", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.AVScraperBaseURL != "https://javdb.example" {
		t.Fatalf("unexpected AVScraperBaseURL: %s", cfg.AVScraperBaseURL)
	}
	if got := cfg.AVSiteURLs["javdb"]; got != "https://javdb.example" {
		t.Fatalf("unexpected AVSiteURLs[javdb]: %s", got)
	}
	if got := cfg.AVSiteURLs["javbus"]; got != "https://javbus.example" {
		t.Fatalf("unexpected AVSiteURLs[javbus]: %s", got)
	}
	if got := cfg.AVSiteURLs["javlibrary"]; got != "https://javlibrary.example" {
		t.Fatalf("unexpected AVSiteURLs[javlibrary]: %s", got)
	}
	if got := cfg.AVSiteURLs["theporndb"]; got != "https://tpdb.example" {
		t.Fatalf("unexpected AVSiteURLs[theporndb]: %s", got)
	}
	if cfg.AVSiteURLJavBus != "https://javbus.example" {
		t.Fatalf("unexpected AVSiteURLJavBus: %s", cfg.AVSiteURLJavBus)
	}
	if cfg.AVSiteURLJavLibrary != "https://javlibrary.example" {
		t.Fatalf("unexpected AVSiteURLJavLibrary: %s", cfg.AVSiteURLJavLibrary)
	}
	if cfg.AVSiteURLThePornDB != "https://tpdb.example" {
		t.Fatalf("unexpected AVSiteURLThePornDB: %s", cfg.AVSiteURLThePornDB)
	}
	if cfg.AVScraperJavDBCookie != "cookie-javdb" {
		t.Fatalf("unexpected AVScraperJavDBCookie: %s", cfg.AVScraperJavDBCookie)
	}
	if cfg.AVScraperJavBusCookie != "cookie-javbus" {
		t.Fatalf("unexpected AVScraperJavBusCookie: %s", cfg.AVScraperJavBusCookie)
	}
	if cfg.AVScraperThePornDBAPIToken != "token-123" {
		t.Fatalf("unexpected AVScraperThePornDBAPIToken: %s", cfg.AVScraperThePornDBAPIToken)
	}
	if !cfg.AVScraperThePornDBNoHash {
		t.Fatalf("expected AVScraperThePornDBNoHash true")
	}
}

func TestLoadBuildsAVSiteURLsFromPrefixedEnvVars(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://user:pass@127.0.0.1:5432/app?sslmode=disable")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("AV_SITE_URL_JAVDB", "https://javdb.example")
	t.Setenv("AV_SITE_URL_JAVBUS", "https://javbus.example")
	t.Setenv("AV_SITE_URL_Foo_Bar", "https://custom.example")
	t.Setenv("AV_SCRAPER_BASE_URL", "https://base.example")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if got := cfg.AVSiteURLs["javdb"]; got != "https://javdb.example" {
		t.Fatalf("unexpected AVSiteURLs[javdb]: %s", got)
	}
	if got := cfg.AVSiteURLs["javbus"]; got != "https://javbus.example" {
		t.Fatalf("unexpected AVSiteURLs[javbus]: %s", got)
	}
	if got := cfg.AVSiteURLs["foo_bar"]; got != "https://custom.example" {
		t.Fatalf("unexpected AVSiteURLs[foo_bar]: %s", got)
	}
	if len(cfg.AVSiteURLs) != 3 {
		t.Fatalf("unexpected AVSiteURLs length: %d", len(cfg.AVSiteURLs))
	}
	if cfg.AVSiteURLJavDB != "https://javdb.example" {
		t.Fatalf("expected explicit JAVDB site url to win, got=%s", cfg.AVSiteURLJavDB)
	}
}

func TestLoadFallsBackToBaseURLForJavDBSiteURL(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://user:pass@127.0.0.1:5432/app?sslmode=disable")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("AV_SCRAPER_BASE_URL", "https://javdb-fallback.example")
	_ = os.Unsetenv("AV_SITE_URL_JAVDB")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.AVSiteURLJavDB != "https://javdb-fallback.example" {
		t.Fatalf("expected AVSiteURLJavDB fallback to base url, got=%s", cfg.AVSiteURLJavDB)
	}
	if got := cfg.AVSiteURLs["javdb"]; got != "https://javdb-fallback.example" {
		t.Fatalf("expected AVSiteURLs[javdb] fallback to base url, got=%s", got)
	}
}
