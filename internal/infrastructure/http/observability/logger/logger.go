package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	LevelDebug        = "debug"
	LevelInfo         = "info"
	LevelWarn         = "warn"
	LevelError        = "error"
	KeyLayer          = "layer"
	KeyModule         = "module"
	KeyOperation      = "operation"
	KeyError          = "error"
	KeyRequestID      = "request_id"
	KeyUserID         = "user_id"
	KeyChatID         = "chat_id"
	KeyPeerID         = "peer_id"
	KeyTargetID       = "target_user_id"
	KeyProfileID      = "profile_id"
	KeyMessageID      = "message_id"
	KeyFileID         = "file_id"
	KeyDuration       = "duration_ms"
	KeyStatus         = "status"
	KeyEvent          = "event"
	KeyMethod         = "method"
	KeyPath           = "path"
	KeyCode           = "code"
	KeyType           = "type"
	KeyConstraint     = "constraint"
	HeaderRequestID   = "X-Request-ID"
	LayerUsecase      = "usecase"
	LayerRepository   = "repository"
	LayerDelivery     = "delivery"
	LayerMiddleware   = "middleware"
	LayerStorage      = "storage"
	LayerTools        = "tools"
	ModuleAuth        = "auth"
	ModuleRegister    = "register"
	ModuleProfile     = "profile"
	ModuleContact     = "contact"
	ModuleChat        = "chat"
	ModuleUser        = "user"
	ModuleSession     = "session"
	ModuleFile        = "file"
	ModuleTransaction = "transaction"
	ModuleWebSocket   = "websocket"
	ModuleHTTP        = "http"
	ModuleCSRF        = "csrf"
	ModuleJSON        = "json"
	ModuleDatabase    = "database"
	Redacted          = "[REDACTED]"
)

const (
	KeyEmailMask         = "email_mask"
	MessageHTTPCompleted = "HTTP request completed"
)

type contextKey uint8

const metadataKey contextKey = 0

type requestMetadata struct {
	requestID string
	userID    atomic.Value
}

type responseError interface {
	ErrorCode() string
	HTTPStatus() int
}

func ParseLevel(value string) (slog.Level, error) {
	switch strings.ToLower(value) {
	case LevelDebug:
		return slog.LevelDebug, nil
	case LevelInfo:
		return slog.LevelInfo, nil
	case LevelWarn:
		return slog.LevelWarn, nil
	case LevelError:
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid LOG_LEVEL: choose debug, info, warn or error")
	}
}

func InitLogger(level string) (*slog.Logger, error) { return New(os.Stdout, level) }

func New(out io.Writer, level string) (*slog.Logger, error) {
	parsed, err := ParseLevel(level)
	if err != nil {
		return nil, err
	}
	options := &slog.HandlerOptions{Level: parsed, ReplaceAttr: sanitize}
	handler := slog.NewTextHandler(out, options)
	return slog.New(&contextHandler{next: handler}), nil
}

func Scope(log *slog.Logger, layer, module string) *slog.Logger {
	if log == nil {
		log = slog.Default()
	}
	return log.With(KeyLayer, layer, KeyModule, module)
}
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, metadataKey, &requestMetadata{requestID: id})
}
func WithUserID(ctx context.Context, id string) context.Context {
	data, ok := ctx.Value(metadataKey).(*requestMetadata)
	if !ok {
		data = &requestMetadata{}
		ctx = context.WithValue(ctx, metadataKey, data)
	}
	data.userID.Store(id)
	return ctx
}

type contextHandler struct{ next slog.Handler }

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}
func (h *contextHandler) Handle(ctx context.Context, record slog.Record) error {
	record = record.Clone()
	if data, ok := ctx.Value(metadataKey).(*requestMetadata); ok {
		if data.requestID != "" {
			record.AddAttrs(slog.String(KeyRequestID, data.requestID))
		}
		if userID, ok := data.userID.Load().(string); ok && userID != "" {
			record.AddAttrs(slog.String(KeyUserID, userID))
		}
	}
	return h.next.Handle(ctx, record)
}
func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{next: h.next.WithAttrs(attrs)}
}
func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{next: h.next.WithGroup(name)}
}

func sanitize(_ []string, attr slog.Attr) slog.Attr {
	attr.Value = attr.Value.Resolve()
	key := strings.ToLower(attr.Key)
	if strings.Contains(key, "password") || strings.Contains(key, "token") || strings.Contains(key, "cookie") || key == "authorization" || key == "session_id" || key == "sessionid" || key == "content" || key == "message" || key == "body" || key == "payload" {
		return slog.String(attr.Key, Redacted)
	}
	if err, ok := attr.Value.Any().(error); ok {
		return slog.Attr{Key: attr.Key, Value: Error(err).Value}
	}
	if key == "email" && attr.Value.Kind() == slog.KindString {
		return slog.String(attr.Key, MaskEmail(attr.Value.String()))
	}
	if (key == "phone" || key == "phone_number") && attr.Value.Kind() == slog.KindString {
		value := attr.Value.String()
		return slog.String(attr.Key, MaskPhone(&value))
	}
	return attr
}

func Error(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}
	attrs := []any{slog.String(KeyType, fmt.Sprintf("%T", err))}
	var dbErr *pgconn.PgError
	var responseErr responseError
	switch {
	case errors.As(err, &dbErr):
		attrs = append(attrs, slog.String(KeyCode, dbErr.Code), slog.String(KeyConstraint, dbErr.ConstraintName))
	case errors.As(err, &responseErr):
		attrs = append(attrs, slog.String(KeyCode, responseErr.ErrorCode()), slog.Int(KeyStatus, responseErr.HTTPStatus()))
	}
	return slog.Group(KeyError, attrs...)
}

func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" {
		return "***"
	}
	first, _ := utf8.DecodeRuneInString(parts[0])
	domain := strings.Split(parts[1], ".")
	suffix := "***"
	if len(domain) > 1 {
		suffix += "." + domain[len(domain)-1]
	}
	return string(first) + "***@" + suffix
}
func MaskPhone(phone *string) string {
	if phone == nil {
		return "***"
	}
	runes := []rune(*phone)
	if len(runes) < 6 {
		return "***"
	}
	return string(runes[:2]) + "***" + string(runes[len(runes)-4:])
}
