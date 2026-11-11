package chat

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/httperrors"
	"log/slog"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/chat/ws"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/gorilla/websocket"
)

const (
	readBufferSize  = 1024
	writeBufferSize = 1024
)

type WSHandler struct {
	log *slog.Logger

	hub      *ws.Hub
	usecase  ChatUsecase
	upgrader websocket.Upgrader
}

func NewWSHandler(hub *ws.Hub, uc ChatUsecase, origins []string, baseLog *slog.Logger) *WSHandler {
	return &WSHandler{
		log: logger.Scope(baseLog, logger.LayerDelivery, logger.ModuleWebSocket),

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

				logger.Scope(baseLog, logger.LayerDelivery, logger.ModuleWebSocket).WarnContext(r.Context(), "websocket origin rejected")
				return false
			},
		},
	}
}

func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httperrors.WriteError(w, httperrors.NewUnauthorized("Authorization required"))
		return
	}

	sessionID, ok := middleware.SessionIDFromContext(r.Context())
	if !ok {
		httperrors.WriteError(w, httperrors.NewUnauthorized("Authorization required"))
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.ErrorContext(r.Context(), "HandleWS failed", logger.Error(err))

		return
	}

	h.log.InfoContext(r.Context(), "websocket connection established")
	client := ws.NewClient(r.Context(), userID, sessionID, conn, h.hub, h.usecase, h.log)
	h.hub.Register(client)
	client.Serve()
}
