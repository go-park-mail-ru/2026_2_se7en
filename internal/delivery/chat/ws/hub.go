package ws

import (
	"sync"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.RWMutex
	users   map[uuid.UUID]map[*Client]struct{}
	chats   map[uuid.UUID]map[*Client]struct{}
	clients map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		users:   make(map[uuid.UUID]map[*Client]struct{}),
		chats:   make(map[uuid.UUID]map[*Client]struct{}),
		clients: make(map[*Client]struct{}),
	}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.users[c.userID] == nil {
		h.users[c.userID] = make(map[*Client]struct{})
	}

	h.users[c.userID][c] = struct{}{}
	h.clients[c] = struct{}{}
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, c)
	if clients, ok := h.users[c.userID]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.users, c.userID)
		}
	}

	for chatID := range c.subscriptions {
		if clients, ok := h.chats[chatID]; ok {
			delete(clients, c)
			if len(clients) == 0 {
				delete(h.chats, chatID)
			}
		}
	}

}

func (h *Hub) Subscribe(c *Client, chatID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.chats[chatID] == nil {
		h.chats[chatID] = make(map[*Client]struct{})
	}

	h.chats[chatID][c] = struct{}{}
	c.subscriptions[chatID] = struct{}{}
}

func (h *Hub) Unsubscribe(c *Client, chatID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.chats[chatID]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.chats, chatID)
		}
	}
	delete(c.subscriptions, chatID)
}

func (h *Hub) RevokeUserFromChat(userID, chatID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for c := range h.users[userID] {
		delete(c.subscriptions, chatID)
		delete(h.chats[chatID], c)
	}

	if len(h.chats[chatID]) == 0 {
		delete(h.chats, chatID)
	}
}

func (h *Hub) RemoveChat(chatID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for c := range h.chats[chatID] {
		delete(c.subscriptions, chatID)
	}

	delete(h.chats, chatID)
}

func (h *Hub) BroadcastToChat(chatID uuid.UUID, event models.Event) {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.chats[chatID]))
	for c := range h.chats[chatID] {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		c.Send(event)
	}
}

func (h *Hub) BroadcastToUser(userID uuid.UUID, event models.Event) {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.users[userID]))
	for c := range h.users[userID] {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		c.Send(event)
	}
}

func (h *Hub) Shutdown() {
	h.mu.Lock()
	clients := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.Unlock()

	for _, c := range clients {
		c.Close(websocket.CloseGoingAway, "server shutdown")
	}
}
