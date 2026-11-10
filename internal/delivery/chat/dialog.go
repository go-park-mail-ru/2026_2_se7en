package chat

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/google/uuid"
)

// HandleCreateDialog godoc
// @Summary      Получить или создать диалог
// @Description  Возвращает существующий диалог с пользователем или создаёт новый
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     SessionCookie
// @Param        request body CreateDialogRequest true "Данные для создания"
// @Success      200 {object} ChatDetailResponse
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/dialogs [post]
func (d *ChatDeliveryImpl) HandleCreateDialog(w http.ResponseWriter, r *http.Request, req *CreateDialogRequest) *apperrors.AppError {

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized("Authorization required")
	}

	if userID == req.PeerID {
		return apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{{Field: "peer_id", Reason: "Cannot create a dialog with yourself"}})
	}

	result, status, err := d.chatUsecase.CreateDialog(r.Context(), userID, req.PeerID)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleCreateDialog failed", logger.Error(err))

		return mapChatError(err)
	}
	result.IsOnline = d.chatOnline(result.Chat, result.PeerID)

	if status == models.DialogCreated || status == models.DialogRestored {
		event := models.Event{
			Type:    models.EventChatCreated,
			Payload: map[string]any{"chat_id": result.Chat.ID},
		}

		d.broadcaster.BroadcastToUser(userID, event)
		if status == models.DialogCreated {
			d.broadcaster.BroadcastToUser(req.PeerID, event)
		}
	}

	tools.WriteJSON(w, http.StatusOK, ChatDetailToResponse(result))
	return nil
}

// HandleHideDialog godoc
// @Summary      Скрыть диалог
// @Description  Убирает текущего пользователя из диалога
// @Tags         chats
// @Security     SessionCookie
// @Param        id path string true "ID диалога (UUID)" format(uuid)
// @Success      204 "Диалог скрыт"
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/{id}/hide [post]
func (d *ChatDeliveryImpl) HandleHideDialog(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}

	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		if err != nil {
			d.log.ErrorContext(r.Context(), "HandleHideDialog failed", logger.Error(err))
		}

		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}

	if err := d.chatUsecase.HideDialog(r.Context(), userID, chatID); err != nil {
		d.log.ErrorContext(r.Context(), "HandleHideDialog failed", logger.Error(err))

		apperrors.WriteError(w, mapChatError(err))
		return
	}

	d.broadcaster.RevokeUserFromChat(userID, chatID)
	d.broadcaster.BroadcastToUser(userID, models.Event{Type: models.EventUnsubscribed, Payload: map[string]any{"chat_id": chatID}})
	d.broadcaster.BroadcastToUsers([]uuid.UUID{userID}, models.Event{Type: models.EventChatActivity, Payload: map[string]any{"chat_id": chatID}})

	w.WriteHeader(http.StatusNoContent)
}
