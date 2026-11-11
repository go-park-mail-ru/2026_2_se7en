package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
)

type Config struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	LogLevel        string
	CSRFSecret      string
	DatabaseURL     string
	Origins         []string
}

func Load() (Config, error) {
	var cfg Config
	var err error

	cfg.Addr = getEnv("ADDR", ":8080")
	cfg.CSRFSecret = getEnv("CSRF_SECRET", "")
	cfg.LogLevel = strings.ToLower(getEnv("LOG_LEVEL", logger.LevelInfo))
	if _, err := logger.ParseLevel(cfg.LogLevel); err != nil {
		return Config{}, err
	}
	cfg.DatabaseURL = getEnv("DATABASE_URL", "")
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DatabaseURL are required")
	}

	for _, value := range strings.Split(getEnv("ORIGINS", ""), ",") {
		origin := strings.TrimSpace(value)
		if origin != "" {
			cfg.Origins = append(cfg.Origins, origin)
		}
	}

	if cfg.ReadTimeout, err = getEnvDuration("READ_TIMEOUT", 5*time.Second); err != nil {
		return Config{}, err
	}

	if cfg.WriteTimeout, err = getEnvDuration("WRITE_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}

	if cfg.IdleTimeout, err = getEnvDuration("IDLE_TIMEOUT", 60*time.Second); err != nil {
		return Config{}, err
	}

	if cfg.ShutdownTimeout, err = getEnvDuration("SHUTDOWN_TIMEOUT", 60*time.Second); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}

	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("config: invalid %s=%q: %w", key, v, err)
	}

	return d, nil
}
