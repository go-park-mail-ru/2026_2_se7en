package chat

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/google/uuid"
)

type Broadcaster interface {
	IsOnline(userID uuid.UUID) bool
	BroadcastToChat(chatID uuid.UUID, event models.Event)
	BroadcastToUser(userID uuid.UUID, event models.Event)
	BroadcastToUsers(userIDs []uuid.UUID, event models.Event)
	RevokeUserFromChat(userID, chatID uuid.UUID)
	RemoveChat(chatID uuid.UUID)
}

type ChatDeliveryImpl struct {
	log *slog.Logger

	chatUsecase ChatUsecase
	broadcaster Broadcaster
}

func NewChatDelivery(chatUsecase ChatUsecase, broadcaster Broadcaster, baseLog *slog.Logger) *ChatDeliveryImpl {
	return &ChatDeliveryImpl{
		log:         logger.Scope(baseLog, logger.LayerDelivery, logger.ModuleChat),
		chatUsecase: chatUsecase, broadcaster: broadcaster}
}

// HandleGetListUserChats godoc
// @Summary      Список чатов пользователя
// @Description  Возвращает страницу чатов с фильтрацией по типу
// @Tags         chats
// @Produce      json
// @Security     SessionCookie
// @Param        limit  query    int     false  "Размер списка (1-50)"  default(20)
// @Param        type   query    string  false  "Фильтр по типу чата"    Enums(dialog, group, channel)
// @Param        after_time query string false "Время активности последнего чата (RFC3339Nano)"
// @Param        after_id   query string false "ID последнего чата"
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
		d.log.ErrorContext(r.Context(), "HandleGetListUserChats failed", logger.Error(err))

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
		d.log.ErrorContext(r.Context(), "HandleGetListUserChats failed", logger.Error(err))

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

	cursorTime, cursorID, err := parseChatPagination(q)
	if err != nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid pagination boundary", nil))
		return
	}

	items, err := d.chatUsecase.GetListUserChats(r.Context(), userID, limit+1, chatType, cursorTime, cursorID)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleGetListUserChats failed", logger.Error(err))

		apperrors.WriteError(w, mapChatError(err))
		return
	}
	for _, item := range items {
		if item != nil {
			item.IsOnline = d.chatOnline(item.Chat, item.PeerID)
		}
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	response := ChatItemsToResponse(items)
	if hasMore && len(items) > 0 && items[len(items)-1] != nil && items[len(items)-1].Chat != nil {
		nextTime := items[len(items)-1].ActivityAt.UTC()
		nextID := items[len(items)-1].Chat.ID
		response.NextAfterTime = &nextTime
		response.NextAfterID = &nextID
	}

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
		if err != nil {
			d.log.ErrorContext(r.Context(), "HandleGetChatDetail failed", logger.Error(err))
		}

		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}

	result, err := d.chatUsecase.GetChatDetail(r.Context(), userID, chatID)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleGetChatDetail failed", logger.Error(err))

		apperrors.WriteError(w, mapChatError(err))
		return
	}
	result.IsOnline = d.chatOnline(result.Chat, result.PeerID)

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
		if err != nil {
			d.log.ErrorContext(r.Context(), "HandleDeleteChat failed", logger.Error(err))
		}

		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}

	userIDs, err := d.chatUsecase.GetActiveMemberIDs(r.Context(), chatID)
	if err != nil {
		apperrors.WriteError(w, mapChatError(err))
		return
	}

	if err := d.chatUsecase.DeleteChat(r.Context(), userID, chatID); err != nil {
		d.log.ErrorContext(r.Context(), "HandleDeleteChat failed", logger.Error(err))

		apperrors.WriteError(w, mapChatError(err))
		return
	}

	d.broadcaster.RemoveChat(chatID)
	d.broadcaster.BroadcastToUsers(userIDs, models.Event{Type: models.EventChatDeleted, Payload: map[string]any{"chat_id": chatID}})

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

func parseChatPagination(q url.Values) (*time.Time, uuid.UUID, error) {
	timeValues, hasTime := q["after_time"]
	idValues, hasID := q["after_id"]
	if !hasTime && !hasID {
		return nil, uuid.Nil, nil
	}
	if !hasTime || !hasID || len(timeValues) != 1 || len(idValues) != 1 {
		return nil, uuid.Nil, errors.New("invalid pagination boundary")
	}

	parsedTime, err := time.Parse(time.RFC3339Nano, timeValues[0])
	if err != nil {
		return nil, uuid.Nil, err
	}
	parsedID, err := uuid.Parse(idValues[0])
	if err != nil || parsedID == uuid.Nil {
		return nil, uuid.Nil, errors.New("invalid pagination boundary")
	}
	return &parsedTime, parsedID, nil
}

func mapChatError(err error) *apperrors.AppError {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	switch {
	case errors.Is(err, apperrors.ErrInvalidChatLimit):
		return apperrors.NewBadRequest("Invalid limit", []apperrors.ErrorDetail{{Field: "limit", Reason: "Limit must be >= 1"}})
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
	case errors.Is(err, apperrors.ErrChatMemberLimit):
		return apperrors.NewConflictWithCode("CHAT_FULL", "Group cannot have more than 100 members", nil)
	default:
		return apperrors.NewInternalError("Something went wrong", err)
	}
}

func (d *ChatDeliveryImpl) broadcastChatActivity(ctx context.Context, chatID uuid.UUID) {
	userIDs, err := d.chatUsecase.GetActiveMemberIDs(ctx, chatID)
	if err != nil {
		return
	}
	d.broadcaster.BroadcastToUsers(userIDs, models.Event{Type: models.EventChatActivity, Payload: map[string]any{"chat_id": chatID}})
}

func (d *ChatDeliveryImpl) chatOnline(chat *models.Chat, peerID *uuid.UUID) *bool {
	if chat == nil || chat.Type != "dialog" || peerID == nil {
		return nil
	}

	online := d.broadcaster.IsOnline(*peerID)
	return &online
}
