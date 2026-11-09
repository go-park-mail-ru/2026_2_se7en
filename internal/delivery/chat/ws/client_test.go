package ws

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

type mockChatService struct {
	changed bool
	called  bool
}

func (m *mockChatService) IsMember(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}

func (m *mockChatService) MarkMessageRead(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (bool, error) {
	m.called = true
	return m.changed, nil
}

func TestMarkMessageReadBroadcastsOnlyOnTransition(t *testing.T) {
	userID, chatID, messageID := uuid.New(), uuid.New(), uuid.New()
	chatService := &mockChatService{changed: true}
	hub := NewHub()
	client := &Client{
		userID:        userID,
		send:          make(chan models.Event, 2),
		ctx:           context.Background(),
		subscriptions: make(map[uuid.UUID]struct{}),
		hub:           hub,
		chatService:   chatService,
	}

	hub.Register(client)
	hub.Subscribe(client, chatID)
	action := models.Action{Type: models.ActionMessageMarkRead, ChatID: chatID, MessageID: messageID}

	client.handleAction(action)
	if !chatService.called {
		t.Fatal("read state was not changed")
	}

	select {
	case event := <-client.send:
		if event.Type != models.EventMessageRead {
			t.Fatalf("event type = %q", event.Type)
		}

		payload, ok := event.Payload.(map[string]any)
		if !ok || payload["chat_id"] != chatID || payload["message_id"] != messageID || payload["user_id"] != userID {
			t.Fatalf("event payload = %+v", event.Payload)
		}
	default:
		t.Fatal("message.read event was not sent")
	}

	chatService.changed = false
	client.handleAction(action)
	select {
	case event := <-client.send:
		t.Fatalf("duplicate read sent event: %+v", event)
	default:
	}
}
