package http

import (
	"log/slog"
	http "net/http"

	httpjson "github.com/go-park-mail-ru/2026_2_se7en/internal/infrastructure/http/json"
	chatdelivery "github.com/go-park-mail-ru/2026_2_se7en/internal/messenger/delivery/http/chat"
)

type AuthMiddleware func(http.Handler) http.Handler

type RouterDependencies struct {
	AuthMiddleware AuthMiddleware
	Logger         *slog.Logger
	Chat           *chatdelivery.ChatDeliveryImpl
	WebSocket      *chatdelivery.WSHandler
}

func RegisterRoutes(mux *http.ServeMux, deps RouterDependencies) {
	mux.Handle("GET /chats", deps.AuthMiddleware(http.HandlerFunc(deps.Chat.HandleGetListUserChats)))
	mux.Handle("GET /chats/{id}", deps.AuthMiddleware(http.HandlerFunc(deps.Chat.HandleGetChatDetail)))
	mux.Handle("POST /chats/dialogs", deps.AuthMiddleware(httpjson.NewJSONDecoder(deps.Chat.HandleCreateDialog, deps.Logger)))
	mux.Handle("POST /chats/groups", deps.AuthMiddleware(httpjson.NewJSONDecoder(deps.Chat.HandleCreateGroup, deps.Logger)))
	mux.Handle("PATCH /chats/{id}", deps.AuthMiddleware(httpjson.NewJSONDecoder(deps.Chat.HandleUpdateGroup, deps.Logger)))
	mux.Handle("PUT /chats/{id}/avatar", deps.AuthMiddleware(http.HandlerFunc(deps.Chat.HandleUploadGroupAvatar)))
	mux.Handle("DELETE /chats/{id}", deps.AuthMiddleware(http.HandlerFunc(deps.Chat.HandleDeleteChat)))
	mux.Handle("POST /chats/{id}/hide", deps.AuthMiddleware(http.HandlerFunc(deps.Chat.HandleHideDialog)))
	mux.Handle("GET /chats/{id}/members", deps.AuthMiddleware(http.HandlerFunc(deps.Chat.HandleGetListMembers)))
	mux.Handle("POST /chats/{id}/members", deps.AuthMiddleware(httpjson.NewJSONDecoder(deps.Chat.HandleAddMember, deps.Logger)))
	mux.Handle("PATCH /chats/{id}/members/{user_id}/role", deps.AuthMiddleware(httpjson.NewJSONDecoder(deps.Chat.HandleUpdateMemberRole, deps.Logger)))
	mux.Handle("DELETE /chats/{id}/members/{user_id}", deps.AuthMiddleware(http.HandlerFunc(deps.Chat.HandleRemoveMember)))
	mux.Handle("DELETE /chats/{id}/leave", deps.AuthMiddleware(http.HandlerFunc(deps.Chat.HandleLeaveGroup)))
	mux.Handle("GET /chats/{id}/messages", deps.AuthMiddleware(http.HandlerFunc(deps.Chat.HandleGetListMessages)))
	mux.Handle("POST /chats/{id}/messages", deps.AuthMiddleware(httpjson.NewJSONDecoder(deps.Chat.HandleCreateMessage, deps.Logger)))
	mux.Handle("GET /ws", deps.AuthMiddleware(http.HandlerFunc(deps.WebSocket.HandleWS)))
}
