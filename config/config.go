package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	LogLevel     string
	DatabaseURL  string
}

func Load() (Config, error) {
	var cfg Config
	var err error

	cfg.Addr = getEnv("ADDR", ":8080")
	cfg.LogLevel = getEnv("LOG_LEVEL", "info")
	cfg.DatabaseURL = getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres")

	if cfg.ReadTimeout, err = getEnvDuration("READ_TIMEOUT", 5*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.WriteTimeout, err = getEnvDuration("WRITE_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.IdleTimeout, err = getEnvDuration("IDLE_TIMEOUT", 60*time.Second); err != nil {
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
