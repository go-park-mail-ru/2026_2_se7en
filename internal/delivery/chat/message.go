package chat

import (
	"net/http"
	"time"

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

	tools.WriteJSON(w, http.StatusCreated, response)
	return nil
}

type CreateMessageRequest struct {
	ReplyToMessageID *uuid.UUID `json:"reply_to_message_id,omitempty" format:"uuid"`
	StickerID        *uuid.UUID `json:"sticker_id,omitempty" format:"uuid"`
	Content          *string    `json:"content,omitempty"`
	Type             string     `json:"type" `
}

type MessageResponse struct {
	ID               uuid.UUID  `json:"id" format:"uuid"`
	UserID           uuid.UUID  `json:"user_id" format:"uuid"`
	ChatID           uuid.UUID  `json:"chat_id" format:"uuid"`
	ReplyToMessageID *uuid.UUID `json:"reply_to_message_id,omitempty" format:"uuid"`
	StickerID        *uuid.UUID `json:"sticker_id,omitempty" format:"uuid"`
	Content          *string    `json:"content,omitempty"`
	IsEdited         bool       `json:"is_edited"`
	IsRead           bool       `json:"is_read"`
	Type             string     `json:"type"`
	CreatedAt        time.Time  `json:"created_at" format:"date-time"`
	UpdatedAt        time.Time  `json:"updated_at" format:"date-time"`
}

type ListMessagesResponse struct {
	Items []MessageResponse `json:"items"`
}

func (r *CreateMessageRequest) IsValid() []apperrors.ErrorDetail {
	if r.ReplyToMessageID != nil && *r.ReplyToMessageID == uuid.Nil {
		return []apperrors.ErrorDetail{{Field: "reply_to_message_id", Reason: "invalid reply_to_message_id"}}
	}

	if r.StickerID != nil && *r.StickerID == uuid.Nil {
		return []apperrors.ErrorDetail{{Field: "sticker_id", Reason: "invalid sticker_id"}}
	}

	if detail := tools.ValidateMessageFields(r.Type, r.Content, r.StickerID); detail != nil {
		return []apperrors.ErrorDetail{*detail}
	}

	return nil
}

func MessageToResponse(message models.Message) MessageResponse {
	return MessageResponse{
		ID:               message.ID,
		UserID:           message.UserID,
		ChatID:           message.ChatID,
		ReplyToMessageID: message.ReplyToMessageID,
		StickerID:        message.StickerID,
		Content:          message.Content,
		IsEdited:         message.IsEdited,
		IsRead:           message.IsRead,
		Type:             message.Type,
		CreatedAt:        message.CreatedAt,
		UpdatedAt:        message.UpdatedAt,
	}
}

func MessagesToResponse(messages []models.Message) ListMessagesResponse {
	items := make([]MessageResponse, 0, len(messages))
	for _, message := range messages {
		items = append(items, MessageToResponse(message))
	}

	return ListMessagesResponse{Items: items}
}
