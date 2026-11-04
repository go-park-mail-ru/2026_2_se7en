package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type ctxKey string

const LoggerKey ctxKey = "logger"

func InitLogger(env string) *slog.Logger {
	var handler slog.Handler
	if env == "prod" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	return slog.New(handler)
}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 || len(parts[0]) == 0 {
		return "***"
	}
	local := parts[0]
	domain := strings.Split(parts[1], ".")

	maskedLocal := string(local[0]) + "***"
	maskedDomain := "***"
	if len(domain) > 1 {
		maskedDomain = "***." + domain[len(domain)-1]
	}

	return maskedLocal + "@" + maskedDomain
}

func MaskPhone(phone *string) string {
	if phone == nil || len(*phone) < 6 {
		return "***"
	}
	p := *phone
	return p[:2] + "***" + p[len(p)-4:]
}
