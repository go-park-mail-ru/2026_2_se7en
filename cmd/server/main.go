package main

// @title           Messenger API
// @version         1.0
// @description     API для мессенджера "Связь"
// @BasePath /api/v1

// @securityDefinitions.apikey SessionCookie
// @in              cookie
// @name            session_id
// @description     Session ID

// @externalDocs.description Документация проекта
// @externalDocs.url https://github.com/go-park-mail-ru/2026_2_se7en

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-park-mail-ru/2026_2_se7en/config"
	_ "github.com/go-park-mail-ru/2026_2_se7en/docs"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/router"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "Ошибка загрузки .env: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.InitLogger(cfg.LogLevel)
	if err != nil {
		fmt.Fprintln(os.Stderr, "logger configuration error")
		os.Exit(1)
	}
	slog.SetDefault(log)

	db, err := storage.Connect(context.Background(), cfg.DatabaseURL, log)
	if err != nil {
		log.Error("Unable to connect to database", logger.Error(err))
		os.Exit(1)
	}
	defer storage.Close(db)

	handler, hub := router.New(db, cfg.Origins, log)

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("starting server", "addr", cfg.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", logger.Error(err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	log.Info("shutting down server...")

	hub.Shutdown()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", logger.Error(err))
	} else {
		log.Info("server stopped gracefully")
	}
}
