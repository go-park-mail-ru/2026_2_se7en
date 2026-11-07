package models

import "github.com/google/uuid"

const (
	EventSubscribed     = "subscribed"
	EventUnsubscribed   = "unsubscribed"
	EventChatCreated    = "chat.created"
	EventChatDeleted    = "chat.deleted"
	EventMemberAdded    = "member.added"
	EventMemberRemoved  = "member.deleted"
	EventTypingStarted  = "typing.started"
	EventTypingStopped  = "typing.stopped"
	EventMessageCreated = "message.created"
	EventMessageRead    = "message.read"

	ActionSubscribe       = "subscribe"
	ActionUnsubscribe     = "unsubscribe"
	ActionTypingStart     = "typing.start"
	ActionTypingStop      = "typing.stop"
	ActionMessageMarkRead = "message.mark_read"
)

type Event struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}

type Action struct {
	Type      string    `json:"type"`
	ChatID    uuid.UUID `json:"chat_id,omitempty"`
	MessageID uuid.UUID `json:"message_id,omitempty"`
}
