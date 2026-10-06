package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"app/config"
	"app/router"
	"app/storage"
)

func TestHealthRoute(t *testing.T) {
	handler := router.New(nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"status":"ok"}` {
		t.Fatalf("body = %q, want status JSON", body)
	}
}

func TestConfigLoadUsesDefaults(t *testing.T) {
	t.Setenv("ADDR", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("READ_TIMEOUT", "")
	t.Setenv("WRITE_TIMEOUT", "")
	t.Setenv("IDLE_TIMEOUT", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":8080" || cfg.LogLevel != "info" || cfg.DatabaseURL != "postgres://example" {
		t.Fatalf("unexpected default config: %+v", cfg)
	}
	if cfg.ReadTimeout != 5*time.Second || cfg.WriteTimeout != 10*time.Second || cfg.IdleTimeout != time.Minute || cfg.ShutdownTimeout != time.Minute {
		t.Fatalf("unexpected default timeouts: %+v", cfg)
	}
}

func TestConfigLoadRejectsInvalidValues(t *testing.T) {
	t.Run("missing database URL", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		if _, err := config.Load(); err == nil {
			t.Fatal("expected missing database URL error")
		}
	})

	t.Run("invalid duration", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://example")
		t.Setenv("READ_TIMEOUT", "not-a-duration")
		if _, err := config.Load(); err == nil {
			t.Fatal("expected invalid duration error")
		}
	})

	t.Run("valid duration and overrides", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://example")
		t.Setenv("ADDR", ":9090")
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("READ_TIMEOUT", "2s")
		t.Setenv("WRITE_TIMEOUT", "3s")
		t.Setenv("IDLE_TIMEOUT", "4s")
		t.Setenv("SHUTDOWN_TIMEOUT", "5s")
		cfg, err := config.Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Addr != ":9090" || cfg.LogLevel != "debug" || cfg.ReadTimeout != 2*time.Second || cfg.WriteTimeout != 3*time.Second || cfg.IdleTimeout != 4*time.Second || cfg.ShutdownTimeout != 5*time.Second {
			t.Fatalf("unexpected overridden config: %+v", cfg)
		}
	})
}

func TestStorageRequiresDatabaseURL(t *testing.T) {
	if _, err := storage.Connect(context.Background(), ""); err == nil {
		t.Fatal("expected missing database URL error")
	}
	storage.Close(nil)
}
