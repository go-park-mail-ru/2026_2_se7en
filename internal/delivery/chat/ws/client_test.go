package ws

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type mockChatService struct {
	membership *models.ChatMembership
	changed    bool
	err        error
}

func (m *mockChatService) IsMember(context.Context, uuid.UUID, uuid.UUID) (*models.ChatMembership, error) {
	return m.membership, m.err
}
func (m *mockChatService) MarkMessageRead(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (bool, error) {
	return m.changed, m.err
}

func websocketPair(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()
	connected := make(chan *websocket.Conn, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}).Upgrade(w, r, nil)
		if err == nil {
			connected <- conn
		}
	}))
	t.Cleanup(server.Close)
	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	client, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	serverConn := <-connected
	t.Cleanup(func() { _ = client.Close(); _ = serverConn.Close() })
	return serverConn, client
}

func newTestClient(userID uuid.UUID, hub *Hub, service *mockChatService) *Client {
	return &Client{log: slog.Default(), userID: userID, send: make(chan models.Event, 8), ctx: context.Background(), cancel: func() {}, subscriptions: map[uuid.UUID]struct{}{}, hub: hub, chatService: service}
}

func TestHubPresenceSubscriptionsAndBroadcast(t *testing.T) {
	hub := NewHub(slog.Default())
	userID, chatID := uuid.New(), uuid.New()
	c := newTestClient(userID, hub, &mockChatService{})
	hub.Register(c)
	if !hub.IsOnline(userID) {
		t.Fatal("registered user should be online")
	}
	hub.Subscribe(c, chatID)
	event := models.Event{Type: models.EventSubscribed}
	hub.BroadcastToUser(userID, event)
	hub.BroadcastToUsers([]uuid.UUID{userID}, event)
	hub.BroadcastToChat(chatID, event)
	if len(c.send) != 3 {
		t.Fatalf("broadcast count = %d", len(c.send))
	}
	hub.Unsubscribe(c, chatID)
	if _, exists := c.subscriptions[chatID]; exists {
		t.Fatal("subscription was not removed")
	}
	hub.Subscribe(c, chatID)
	hub.RevokeUserFromChat(userID, chatID)
	if _, exists := c.subscriptions[chatID]; exists {
		t.Fatal("revoked subscription remains")
	}
	hub.Subscribe(c, chatID)
	hub.RemoveChat(chatID)
	if _, exists := c.subscriptions[chatID]; exists {
		t.Fatal("removed chat subscription remains")
	}
	hub.Unregister(c)
	if hub.IsOnline(userID) {
		t.Fatal("unregistered user should be offline")
	}
}

func TestClientCloseIsIdempotentAndSendRespectsContext(t *testing.T) {
	serverConn, _ := websocketPair(t)
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{log: slog.Default(), conn: serverConn, ctx: ctx, cancel: cancel, send: make(chan models.Event, 1)}
	c.Close(websocket.CloseNormalClosure, "done")
	c.Close(websocket.CloseNormalClosure, "again")
	if ctx.Err() == nil {
		t.Fatal("Close must cancel the client context")
	}
	c.Send(models.Event{Type: "ignored"})
	if len(c.send) != 0 {
		t.Fatal("Send queued after close")
	}
}

func TestClientHandleActionMembershipAndReadEvents(t *testing.T) {
	userID, chatID, messageID := uuid.New(), uuid.New(), uuid.New()
	hub := NewHub(slog.Default())
	service := &mockChatService{membership: &models.ChatMembership{IsMember: true}, changed: true}
	c := newTestClient(userID, hub, service)
	hub.Register(c)
	c.handleAction(models.Action{Type: models.ActionSubscribe, ChatID: chatID})
	if _, ok := c.subscriptions[chatID]; !ok {
		t.Fatal("subscribe action did not subscribe")
	}
	c.handleAction(models.Action{Type: models.ActionTypingStart, ChatID: chatID})
	c.handleAction(models.Action{Type: models.ActionTypingStop, ChatID: chatID})
	c.handleAction(models.Action{Type: models.ActionMessageMarkRead, ChatID: chatID, MessageID: messageID})
	if len(c.send) == 0 {
		t.Fatal("action did not queue an event")
	}
	if got := <-c.send; got.Type != models.EventSubscribed {
		t.Fatalf("event = %q", got.Type)
	}
	if len(c.send) != 3 {
		t.Fatalf("remaining events = %d, want typing start/stop and read", len(c.send))
	}
	c.handleAction(models.Action{Type: models.ActionUnsubscribe, ChatID: chatID})
	if _, ok := c.subscriptions[chatID]; ok {
		t.Fatal("unsubscribe action did not remove subscription")
	}
	hub.Unregister(c)
}

func TestClientHandleActionErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		action  models.Action
		service *mockChatService
		want    string
	}{
		{name: "unknown action", action: models.Action{Type: "nope"}, want: "unknown_action"},
		{name: "missing chat", action: models.Action{Type: models.ActionSubscribe}, want: "invalid_chat_id"},
		{name: "not member", action: models.Action{Type: models.ActionSubscribe, ChatID: uuid.New()}, service: &mockChatService{membership: &models.ChatMembership{}}, want: "not_a_member"},
		{name: "service unavailable", action: models.Action{Type: models.ActionSubscribe, ChatID: uuid.New()}, service: &mockChatService{err: errors.New("down")}, want: "internal_error"},
		{name: "missing message", action: models.Action{Type: models.ActionMessageMarkRead, ChatID: uuid.New()}, want: "invalid_message_id"},
		{name: "message missing", action: models.Action{Type: models.ActionMessageMarkRead, ChatID: uuid.New(), MessageID: uuid.New()}, service: &mockChatService{err: apperrors.ErrMessageNotFound}, want: "message_not_found"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := tc.service
			if service == nil {
				service = &mockChatService{}
			}
			c := newTestClient(uuid.New(), NewHub(slog.Default()), service)
			c.handleAction(tc.action)
			select {
			case event := <-c.send:
				payload, _ := event.Payload.(map[string]string)
				if event.Type != "error" || payload["code"] != tc.want {
					t.Fatalf("event = %+v, want error %q", event, tc.want)
				}
			case <-time.After(time.Second):
				t.Fatal("expected error event")
			}
		})
	}
}
