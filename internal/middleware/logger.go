package middleware

import (
	"bufio"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/google/uuid"
)

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(logger.HeaderRequestID)
		if reqID == "" {
			reqID = uuid.NewString()
		}
		ctx := logger.WithRequestID(r.Context(), reqID)
		w.Header().Set(logger.HeaderRequestID, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AccessLogMiddleware(next http.Handler, baseLogger *slog.Logger) http.Handler {
	log := logger.Scope(baseLogger, logger.LayerMiddleware, logger.ModuleHTTP)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		level := slog.LevelInfo
		if rw.statusCode >= 500 {
			level = slog.LevelError
		} else if rw.statusCode == http.StatusNotFound {
			level = slog.LevelDebug
		} else if rw.statusCode >= 400 {
			level = slog.LevelWarn
		}
		log.Log(r.Context(), level, logger.MessageHTTPCompleted,
			logger.KeyMethod, r.Method,
			logger.KeyPath, r.URL.Path,
			logger.KeyStatus, rw.statusCode,
			logger.KeyDuration, time.Since(start).Milliseconds(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	if code >= 100 && code < 200 && code != http.StatusSwitchingProtocols {
		rw.ResponseWriter.WriteHeader(code)
		return
	}
	rw.statusCode = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, buffer, err := http.NewResponseController(rw.ResponseWriter).Hijack()
	if err == nil {
		rw.statusCode = http.StatusSwitchingProtocols
		rw.wroteHeader = true
	}
	return conn, buffer, err
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(data)
}
