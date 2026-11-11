package config

import "testing"

func TestLoggingConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("LOG_LEVEL", "warn")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LogLevel != "warn" {
		t.Fatalf("logging level=%s", cfg.LogLevel)
	}
	t.Setenv("LOG_LEVEL", "invalid")
	if _, err := Load(); err == nil {
		t.Error("invalid level accepted")
	}
}
