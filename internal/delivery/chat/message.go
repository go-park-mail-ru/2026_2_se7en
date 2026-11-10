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

// HandleGetListMessages godoc
// @Summary      Сообщения чата
// @Description  Возвращает сообщения чата с пагинацией по ID предыдущего сообщения
// @Tags         chats
// @Produce      json
// @Security     SessionCookie
// @Param        id path string true "ID чата (UUID)" format(uuid)
// @Param        limit query int false "Число сообщений (1-50)" default(20)
// @Param        before query string false "ID сообщения для пагинации (UUID)" format(uuid)
// @Success      200 {object} ListMessagesResponse
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/{id}/messages [get]
func (d *ChatDeliveryImpl) HandleGetListMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}

	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		if err != nil {
			d.log.ErrorContext(r.Context(), "HandleGetListMessages failed", logger.Error(err))
		}

		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}

	q := r.URL.Query()
	limit := 20
	if _, present := q["limit"]; present {
		parsed, err := parseQueryInt(q, "limit")
		if err != nil {
			d.log.ErrorContext(r.Context(), "HandleGetListMessages failed", logger.Error(err))

			apperrors.WriteError(w, apperrors.NewBadRequest("Invalid limit format", nil))
			return
		}

		limit = parsed
	}

	if limit < 1 || limit > 50 {
		apperrors.WriteError(w, apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{{Field: "limit", Reason: "Limit must be between 1 and 50"}}))
		return
	}

	beforeID, err := parseQueryUUID(q, "before")
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleGetListMessages failed", logger.Error(err))

		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid before format", nil))
		return
	}

	var before *uuid.UUID
	if _, present := q["before"]; present {
		before = &beforeID
	}

	result, err := d.chatUsecase.GetListMessages(r.Context(), userID, chatID, limit, before)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleGetListMessages failed", logger.Error(err))

		apperrors.WriteError(w, mapChatError(err))
		return
	}

	tools.WriteJSON(w, http.StatusOK, MessagesToResponse(result))
}

// HandleCreateMessage godoc
// @Summary      Создать сообщение
// @Description  Отправляет текст или стикер в чат и возвращает созданное сообщение
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     SessionCookie
// @Param        id path string true "ID чата (UUID)" format(uuid)
// @Param        request body CreateMessageRequest true "Данные сообщения"
// @Success      201 {object} MessageResponse
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/{id}/messages [post]
func (d *ChatDeliveryImpl) HandleCreateMessage(w http.ResponseWriter, r *http.Request, req *CreateMessageRequest) *apperrors.AppError {

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized("Authorization required")
	}

	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		if err != nil {
			d.log.ErrorContext(r.Context(), "HandleCreateMessage failed", logger.Error(err))
		}

		return apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}})
	}

	result, usecaseErr := d.chatUsecase.CreateMessage(r.Context(), userID, chatID, req.ReplyToMessageID, req.StickerID, req.Content, req.Type)
	if usecaseErr != nil {
		return mapChatError(usecaseErr)
	}

	response := MessageToResponse(*result)
	d.broadcaster.BroadcastToChat(chatID, models.Event{
		Type:    models.EventMessageCreated,
		Payload: response,
	})
	d.broadcastChatActivity(r.Context(), chatID)

	tools.WriteJSON(w, http.StatusCreated, response)
	return nil
}
