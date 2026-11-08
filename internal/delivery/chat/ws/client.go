package ws

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	pingInterval = 30 * time.Second
	pongTimeout  = 60 * time.Second
	readLimit    = 64 * 1024
	writeTimeout = 10 * time.Second
	writeBuffer  = 256
)

type ChatService interface {
	IsMember(ctx context.Context, userID, chatID uuid.UUID) (bool, error)
	MarkMessageRead(ctx context.Context, userID, chatID, messageID uuid.UUID) (bool, error)
}

type Client struct {
	userID        uuid.UUID
	sessionID     uuid.UUID
	conn          *websocket.Conn
	send          chan models.Event
	ctx           context.Context
	cancel        context.CancelFunc
	closeOnce     sync.Once
	subscriptions map[uuid.UUID]struct{}
	hub           *Hub
	chatService   ChatService
}

func NewClient(userID, sessionID uuid.UUID, conn *websocket.Conn, hub *Hub, cs ChatService) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		userID:        userID,
		sessionID:     sessionID,
		conn:          conn,
		send:          make(chan models.Event, writeBuffer),
		ctx:           ctx,
		cancel:        cancel,
		subscriptions: make(map[uuid.UUID]struct{}),
		hub:           hub,
		chatService:   cs,
	}
}

func (c *Client) Close(code int, reason string) {
	c.closeOnce.Do(func() {
		c.cancel()
		msg := websocket.FormatCloseMessage(code, reason)
		_ = c.conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(writeTimeout))
		_ = c.conn.Close()
	})
}

func (c *Client) Send(event models.Event) {
	select {
	case <-c.ctx.Done():
		return
	default:
	}

	select {
	case c.send <- event:
	case <-c.ctx.Done():
	default:
		c.Close(websocket.ClosePolicyViolation, "slow consumer")
	}
}

func (c *Client) sendActionError(code string) {
	c.Send(models.Event{Type: "error", Payload: map[string]string{"code": code}})
}

func (c *Client) handleAction(action models.Action) {
	switch action.Type {
	case models.ActionUnsubscribe:
		if action.ChatID == uuid.Nil {
			c.sendActionError("invalid_chat_id")
			return
		}

		c.hub.Unsubscribe(c, action.ChatID)
		c.Send(models.Event{Type: models.EventUnsubscribed, Payload: map[string]any{"chat_id": action.ChatID}})

	case models.ActionSubscribe, models.ActionTypingStart, models.ActionTypingStop:
		if action.ChatID == uuid.Nil {
			c.sendActionError("invalid_chat_id")
			return
		}

		isMember, err := c.chatService.IsMember(c.ctx, c.userID, action.ChatID)
		if err != nil {
			c.sendActionError("internal_error")
			return
		}

		if !isMember {
			c.sendActionError("not_a_member")
			return
		}

		switch action.Type {
		case models.ActionSubscribe:
			c.hub.Subscribe(c, action.ChatID)
			c.Send(models.Event{Type: models.EventSubscribed, Payload: map[string]any{"chat_id": action.ChatID}})

		case models.ActionTypingStart:
			c.hub.BroadcastToChat(action.ChatID, models.Event{Type: models.EventTypingStarted, Payload: map[string]any{"user_id": c.userID, "chat_id": action.ChatID}})

		case models.ActionTypingStop:
			c.hub.BroadcastToChat(action.ChatID, models.Event{Type: models.EventTypingStopped, Payload: map[string]any{"user_id": c.userID, "chat_id": action.ChatID}})
		}

	case models.ActionMessageMarkRead:
		if action.ChatID == uuid.Nil {
			c.sendActionError("invalid_chat_id")
			return
		}

		if action.MessageID == uuid.Nil {
			c.sendActionError("invalid_message_id")
			return
		}

		changed, err := c.chatService.MarkMessageRead(c.ctx, c.userID, action.ChatID, action.MessageID)
		if err != nil {
			switch {
			case errors.Is(err, apperrors.ErrChatNotFound):
				c.sendActionError("not_a_member")
			case errors.Is(err, apperrors.ErrMessageNotFound):
				c.sendActionError("message_not_found")
			case errors.Is(err, apperrors.ErrInvalidChatOperation):
				c.sendActionError("cannot_read_own_message")
			default:
				c.sendActionError("internal_error")
			}

			return
		}

		if changed {
			c.hub.BroadcastToChat(action.ChatID, models.Event{
				Type: models.EventMessageRead,
				Payload: map[string]any{
					"chat_id":    action.ChatID,
					"message_id": action.MessageID,
					"user_id":    c.userID,
				},
			})
		}

	default:
		c.sendActionError("unknown_action")
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c)
		c.Close(websocket.CloseNormalClosure, "read pump exited")
	}()

	c.conn.SetReadLimit(readLimit)
	c.conn.SetReadDeadline(time.Now().Add(pongTimeout))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var action models.Action
		if err := json.Unmarshal(data, &action); err != nil {
			c.sendActionError("invalid_json")
			continue
		}

		c.handleAction(action)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.Close(websocket.CloseNormalClosure, "write pump exited")
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case event := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(pingInterval)); err != nil {
				return
			}
		}
	}
}

func (c *Client) Serve() {
	go c.writePump()
	c.readPump()
}
