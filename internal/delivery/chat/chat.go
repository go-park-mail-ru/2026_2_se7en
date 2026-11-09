package chat

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/google/uuid"
)

type Broadcaster interface {
	BroadcastToChat(chatID uuid.UUID, event models.Event)
	BroadcastToUser(userID uuid.UUID, event models.Event)
	RevokeUserFromChat(userID, chatID uuid.UUID)
	RemoveChat(chatID uuid.UUID)
}

type ChatDeliveryImpl struct {
	chatUsecase ChatUsecase
	broadcaster Broadcaster
}

func NewChatDelivery(chatUsecase ChatUsecase, broadcaster Broadcaster) *ChatDeliveryImpl {
	return &ChatDeliveryImpl{chatUsecase: chatUsecase, broadcaster: broadcaster}
}

// HandleGetListUserChats godoc
// @Summary      Список чатов пользователя
// @Description  Возвращает список чатов с поддержкой пагинации и фильтрации по типу
// @Tags         chats
// @Produce      json
// @Security     SessionCookie
// @Param        limit  query    int     false  "Размер списка (1-50)"  default(20)
// @Param        offset query    int     false  "Смещение от начала"     default(0)
// @Param        type   query    string  false  "Фильтр по типу чата"    Enums(dialog, group, channel)
// @Success      200  {object} ListChatsResponse
// @Failure 400 {object} apperrors.ErrorResponse
// @Failure 401 {object} apperrors.ErrorResponse
// @Failure 409 {object} apperrors.ErrorResponse
// @Failure 500 {object} apperrors.ErrorResponse
// @Router       /chats [get]
func (d *ChatDeliveryImpl) HandleGetListUserChats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}

	q := r.URL.Query()

	limit, err := parseQueryInt(q, "limit")
	if err != nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid limit format", []apperrors.ErrorDetail{{Field: "limit", Reason: err.Error()}}))
		return
	}

	if _, present := q["limit"]; !present {
		limit = 20
	}

	if limit < 1 || limit > 50 {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid limit", []apperrors.ErrorDetail{{Field: "limit", Reason: "Limit must be between 1 and 50"}}))
		return
	}

	offset, err := parseQueryInt(q, "offset")
	if err != nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid offset format", []apperrors.ErrorDetail{{Field: "offset", Reason: err.Error()}}))
		return
	}

	if offset < 0 {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid offset", []apperrors.ErrorDetail{{Field: "offset", Reason: "Offset must be >= 0"}}))
		return
	}

	chatType := ""
	if vals, present := q["type"]; present {
		if len(vals) != 1 || vals[0] == "" {
			apperrors.WriteError(w, apperrors.NewBadRequest("Invalid type format", []apperrors.ErrorDetail{{Field: "type", Reason: "Type must be a single non-empty value"}}))
			return
		}

		chatType = vals[0]
	}

	if chatType != "" && chatType != "dialog" && chatType != "group" && chatType != "channel" {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid type", []apperrors.ErrorDetail{{Field: "type", Reason: "Invalid chat type"}}))
		return
	}

	items, err := d.chatUsecase.GetListUserChats(r.Context(), userID, limit, offset, chatType)
	if err != nil {
		apperrors.WriteError(w, mapChatError(err))
		return
	}

	response := ChatItemsToResponse(items)
	tools.WriteJSON(w, http.StatusOK, response)
}

// HandleGetChatDetail godoc
// @Summary      Детали чата для пользователя
// @Description  Возвращает чат и роль текущего пользователя в нём
// @Tags         chats
// @Produce      json
// @Security     SessionCookie
// @Param        id path string true "ID чата (UUID)" format(uuid)
// @Success      200  {object} ChatDetailResponse
// @Failure 400 {object} apperrors.ErrorResponse
// @Failure 401 {object} apperrors.ErrorResponse
// @Failure 404 {object} apperrors.ErrorResponse
// @Failure 500 {object} apperrors.ErrorResponse
// @Router       /chats/{id} [get]
func (d *ChatDeliveryImpl) HandleGetChatDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}

	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}

	result, err := d.chatUsecase.GetChatDetail(r.Context(), userID, chatID)
	if err != nil {
		apperrors.WriteError(w, mapChatError(err))
		return
	}

	tools.WriteJSON(w, http.StatusOK, ChatDetailToResponse(result))
}

// HandleDeleteChat godoc
// @Summary      Удалить чат
// @Description  Удаляет диалог для обоих участников или группу для её создателя
// @Tags         chats
// @Security     SessionCookie
// @Param        id path string true "ID чата (UUID)" format(uuid)
// @Success      204 "Чат удалён"
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      403 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/{id} [delete]
func (d *ChatDeliveryImpl) HandleDeleteChat(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}

	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}

	if err := d.chatUsecase.DeleteChat(r.Context(), userID, chatID); err != nil {
		apperrors.WriteError(w, mapChatError(err))
		return
	}

	d.broadcaster.BroadcastToChat(chatID, models.Event{
		Type:    models.EventChatDeleted,
		Payload: map[string]any{"chat_id": chatID},
	})
	d.broadcaster.RemoveChat(chatID)

	w.WriteHeader(http.StatusNoContent)
}

func parseQueryInt(query url.Values, key string) (int, error) {
	vals, present := query[key]
	if !present {
		return 0, nil
	}

	if len(vals) != 1 {
		return 0, apperrors.ErrInvalidChatLimit
	}

	val, err := strconv.Atoi(vals[0])
	if err != nil {
		return 0, err
	}

	return val, nil
}

func parseQueryUUID(query url.Values, key string) (uuid.UUID, error) {
	vals, present := query[key]
	if !present {
		return uuid.Nil, nil
	}

	if len(vals) != 1 {
		return uuid.Nil, apperrors.ErrInvalidUUID
	}

	val, err := uuid.Parse(vals[0])
	if err != nil || val == uuid.Nil {
		return uuid.Nil, apperrors.ErrInvalidUUID
	}

	return val, nil
}

func mapChatError(err error) *apperrors.AppError {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	switch {
	case errors.Is(err, apperrors.ErrInvalidChatLimit):
		return apperrors.NewBadRequest("Invalid limit", []apperrors.ErrorDetail{{Field: "limit", Reason: "Limit must be >= 1"}})
	case errors.Is(err, apperrors.ErrInvalidChatOffset):
		return apperrors.NewBadRequest("Invalid offset", []apperrors.ErrorDetail{{Field: "offset", Reason: "Offset must be >= 0"}})
	case errors.Is(err, apperrors.ErrInvalidChatType):
		return apperrors.NewBadRequest("Invalid type", []apperrors.ErrorDetail{{Field: "type", Reason: "Invalid chat type"}})
	case errors.Is(err, apperrors.ErrChatNotFound), errors.Is(err, apperrors.ErrChatMemberNotFound), errors.Is(err, apperrors.ErrUserNotFound):
		return apperrors.NewNotFound("Resource not found")
	case errors.Is(err, apperrors.ErrChatPermissionDenied):
		return apperrors.NewForbidden("No permissions")
	case errors.Is(err, apperrors.ErrInvalidChatOperation), errors.Is(err, apperrors.ErrMessageNotFound):
		return apperrors.NewBadRequest("Validation failed", nil)
	case errors.Is(err, apperrors.ErrChatMemberExists):
		return apperrors.NewConflict("Resource already exists", nil)
	default:
		return apperrors.NewInternalError("Something went wrong", err)
	}
}

type MessagePreviewResponse struct {
	ID        uuid.UUID `json:"id" format:"uuid"`
	Content   *string   `json:"content,omitempty"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
}

type ChatItemResponse struct {
	ID           uuid.UUID               `json:"id" format:"uuid"`
	IconID       *uuid.UUID              `json:"icon_id,omitempty" format:"uuid"`
	Name         string                  `json:"name"`
	Description  *string                 `json:"description,omitempty"`
	Type         string                  `json:"type"`
	MembersCount int                     `json:"members_count"`
	CreatedAt    time.Time               `json:"created_at" format:"date-time"`
	UpdatedAt    time.Time               `json:"updated_at" format:"date-time"`
	LastMessage  *MessagePreviewResponse `json:"last_message,omitempty"`
}

type ListChatsResponse struct {
	Items []ChatItemResponse `json:"items"`
}

type ChatResponse struct {
	ID           uuid.UUID  `json:"id" format:"uuid"`
	IconID       *uuid.UUID `json:"icon_id,omitempty" format:"uuid"`
	Name         string     `json:"name"`
	Description  *string    `json:"description,omitempty"`
	Type         string     `json:"type"`
	MembersCount int        `json:"members_count"`
	CreatedAt    time.Time  `json:"created_at" format:"date-time"`
	UpdatedAt    time.Time  `json:"updated_at" format:"date-time"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" format:"date-time"`
}

type ChatDetailResponse struct {
	Chat *ChatResponse `json:"chat"`
	Role string        `json:"role"`
}

func ChatItemsToResponse(items []*models.ChatListItem) ListChatsResponse {
	if items == nil {
		return ListChatsResponse{Items: []ChatItemResponse{}}
	}

	ResponseItems := make([]ChatItemResponse, 0, len(items))
	for _, item := range items {
		if item == nil || item.Chat == nil {
			continue
		}

		Response := ChatItemResponse{
			ID:           item.Chat.ID,
			IconID:       item.Chat.IconID,
			Name:         item.Chat.Name,
			Description:  item.Chat.Description,
			Type:         item.Chat.Type,
			MembersCount: item.Chat.MembersCount,
			CreatedAt:    item.Chat.CreatedAt,
			UpdatedAt:    item.Chat.UpdatedAt,
		}

		if item.LastMessage != nil {
			Response.LastMessage = &MessagePreviewResponse{
				ID:        item.LastMessage.ID,
				Content:   item.LastMessage.Content,
				Type:      item.LastMessage.Type,
				CreatedAt: item.LastMessage.CreatedAt,
			}
		}

		ResponseItems = append(ResponseItems, Response)
	}

	return ListChatsResponse{Items: ResponseItems}
}

func ChatToResponse(chat *models.Chat) *ChatResponse {
	if chat == nil {
		return nil
	}

	return &ChatResponse{
		ID:           chat.ID,
		IconID:       chat.IconID,
		Name:         chat.Name,
		Description:  chat.Description,
		Type:         chat.Type,
		MembersCount: chat.MembersCount,
		CreatedAt:    chat.CreatedAt,
		UpdatedAt:    chat.UpdatedAt,
		DeletedAt:    chat.DeletedAt,
	}
}

func ChatDetailToResponse(detail *models.ChatDetail) *ChatDetailResponse {
	if detail == nil {
		return nil
	}

	return &ChatDetailResponse{Chat: ChatToResponse(detail.Chat), Role: detail.Role}
}
