package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func TestAccessLogPreservesFirstStatusAndRequestID(t *testing.T) {
	var output bytes.Buffer
	log, _ := logger.New(&output, logger.LevelDebug)
	handler := RequestIDMiddleware(AccessLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("ok"))
	}), log))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/contacts", nil)
	req.Header.Set(logger.HeaderRequestID, "request-one")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != http.StatusCreated {
		t.Fatalf("status=%d", response.Code)
	}
	if !strings.Contains(output.String(), "status=201") || !strings.Contains(output.String(), "request_id=request-one") {
		t.Fatalf("wrong access log: %s", output.String())
	}
	if response.Header().Get(logger.HeaderRequestID) != "request-one" {
		t.Error("request header lost")
	}
}

func TestAccessLogImplicitStatus(t *testing.T) {
	var output bytes.Buffer
	log, _ := logger.New(&output, logger.LevelDebug)
	handler := AccessLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("body")) }), log)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != 200 || !strings.Contains(output.String(), `status=200`) {
		t.Fatalf("status failed: %s", output.String())
	}
}

func TestAccessLogAllowsWebSocketUpgrade(t *testing.T) {
	var output bytes.Buffer
	log, _ := logger.New(&output, logger.LevelDebug)
	done := make(chan struct{})
	upgrader := websocket.Upgrader{}
	handler := RequestIDMiddleware(AccessLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade failed: %v", err)
			return
		}
		defer conn.Close()
		if err := conn.WriteMessage(websocket.TextMessage, []byte("ok")); err != nil {
			t.Errorf("write failed: %v", err)
		}
	}), log))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { defer close(done); handler.ServeHTTP(w, r) }))
	defer server.Close()
	conn, response, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if response.StatusCode != 101 {
		t.Fatalf("status=%d", response.StatusCode)
	}
	_, message, err := conn.ReadMessage()
	if err != nil || string(message) != "ok" {
		t.Fatalf("message=%q err=%v", message, err)
	}
	<-done
	if !strings.Contains(output.String(), `status=101`) {
		t.Fatalf("wrong WS log: %s", output.String())
	}
}

type loggingSessionRepo struct{ userID uuid.UUID }

func (r loggingSessionRepo) GetSessionUser(context.Context, uuid.UUID) (uuid.UUID, error) {
	return r.userID, nil
}
func (r loggingSessionRepo) ExtendSession(context.Context, uuid.UUID, time.Time) error { return nil }
func TestAuthKeepsSessionAndLoggingMetadata(t *testing.T) {
	var output bytes.Buffer
	log, _ := logger.New(&output, logger.LevelInfo)
	userID, sessionID := uuid.New(), uuid.New()
	called := false
	handler := RequestIDMiddleware(AccessLogMiddleware(NewAuthMiddleware(loggingSessionRepo{userID: userID}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if id, ok := UserIDFromContext(r.Context()); !ok || id != userID {
			t.Error("user context lost")
		}
		if id, ok := SessionIDFromContext(r.Context()); !ok || id != sessionID {
			t.Error("session context lost")
		}
		log.InfoContext(r.Context(), "service entry")
	}), log), log))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set(logger.HeaderRequestID, "request-one")
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: sessionID.String()})
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if !called || !strings.Contains(output.String(), userID.String()) || !strings.Contains(output.String(), `request_id=request-one`) {
		t.Fatalf("metadata lost: %s", output.String())
	}
	for _, entry := range strings.Split(strings.TrimSpace(output.String()), "\n") {
		if !strings.Contains(entry, "user_id="+userID.String()) || !strings.Contains(entry, "request_id=request-one") {
			t.Fatalf("missing API request metadata: %s", entry)
		}
	}
	if strings.Contains(output.String(), sessionID.String()) {
		t.Fatal("session ID leaked")
	}
}
