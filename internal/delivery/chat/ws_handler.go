package chat

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/chat/ws"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/gorilla/websocket"
)

const (
	readBufferSize  = 1024
	writeBufferSize = 1024
)

type WSHandler struct {
	hub      *ws.Hub
	usecase  ChatUsecase
	upgrader websocket.Upgrader
}

func NewWSHandler(hub *ws.Hub, uc ChatUsecase, origins []string) *WSHandler {
	return &WSHandler{
		hub:     hub,
		usecase: uc,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  readBufferSize,
			WriteBufferSize: writeBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				for _, allowed := range origins {
					if origin == allowed {
						return true
					}
				}
				return false
			},
		},
	}
}

func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := ws.NewClient(userID, conn, h.hub, h.usecase)
	h.hub.Register(client)
	client.Serve()
}
