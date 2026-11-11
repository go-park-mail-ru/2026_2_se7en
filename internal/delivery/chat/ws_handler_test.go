package chat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/chat/ws"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func TestWSHandlerRequiresUserAndSession(t *testing.T) {
	handler := NewWSHandler(ws.NewHub(nil), &mockChatUsecase{}, []string{"http://allowed"}, nil)
	for _, tc := range []struct {
		name string
		ctx  context.Context
	}{
		{name: "anonymous", ctx: context.Background()},
		{name: "missing session", ctx: middleware.NewAuthContext(context.Background(), uuid.New())},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ws", nil).WithContext(tc.ctx)
			rec := httptest.NewRecorder()
			handler.HandleWS(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestWSHandlerRejectsUnlistedOrigin(t *testing.T) {
	hub := ws.NewHub(nil)
	handler := NewWSHandler(hub, &mockChatUsecase{}, []string{"http://allowed"}, nil)
	userID, sessionID := uuid.New(), uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.NewAuthContext(middleware.NewSessionContext(r.Context(), sessionID), userID)
		handler.HandleWS(w, r.WithContext(ctx))
	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	_, response, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err == nil {
		t.Fatal("expected origin rejection")
	}
	if response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("response = %+v, err = %v", response, err)
	}
}

func TestWSHandlerAcceptsConfiguredOriginAndUnregistersOnClose(t *testing.T) {
	hub := ws.NewHub(nil)
	userID, sessionID := uuid.New(), uuid.New()
	origin := "http://allowed.example"
	handler := NewWSHandler(hub, &mockChatUsecase{}, []string{origin}, nil)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.NewAuthContext(middleware.NewSessionContext(r.Context(), sessionID), userID)
		handler.HandleWS(w, r.WithContext(ctx))
	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	dialer := *websocket.DefaultDialer
	dialer.Subprotocols = nil
	conn, response, err := dialer.Dial(u.String(), http.Header{"Origin": []string{origin}})
	if err != nil {
		t.Fatalf("handshake failed: response=%+v err=%v", response, err)
	}
	deadline := time.Now().Add(time.Second)
	for !hub.IsOnline(userID) && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if !hub.IsOnline(userID) {
		t.Fatal("websocket client was not registered")
	}
	_ = conn.Close()
	deadline = time.Now().Add(time.Second)
	for hub.IsOnline(userID) && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if hub.IsOnline(userID) {
		t.Fatal("client was not unregistered after close")
	}
}
