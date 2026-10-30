package chat

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
)

type ChatDeliveryImpl struct {
	chatUsecase ChatUsecase
}

func NewChatDelivery(chatUsecase ChatUsecase) *ChatDeliveryImpl {
	return &ChatDeliveryImpl{chatUsecase: chatUsecase}
}

// HandleListUserChats godoc
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
func (d *ChatDeliveryImpl) HandleListUserChats(w http.ResponseWriter, r *http.Request) {
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

	offset, err := parseQueryInt(q, "offset")
	if err != nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid offset format", []apperrors.ErrorDetail{{Field: "offset", Reason: err.Error()}}))
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

	items, err := d.chatUsecase.ListUserChats(r.Context(), userID, limit, offset, chatType)
	if err != nil {
		apperrors.WriteError(w, mapChatError(err))
		return
	}

	response := ChatItemsToResponse(items)
	tools.WriteJSON(w, http.StatusOK, response)
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

func mapChatError(err error) *apperrors.AppError {
	switch {
	case err == apperrors.ErrInvalidChatLimit:
		return apperrors.NewBadRequest("Invalid limit", []apperrors.ErrorDetail{{Field: "limit", Reason: "Limit must be >= 1"}})
	case err == apperrors.ErrInvalidChatOffset:
		return apperrors.NewBadRequest("Invalid offset", []apperrors.ErrorDetail{{Field: "offset", Reason: "Offset must be >= 0"}})
	case err == apperrors.ErrInvalidChatType:
		return apperrors.NewBadRequest("Invalid type", []apperrors.ErrorDetail{{Field: "type", Reason: "Invalid chat type"}})
	default:
		return apperrors.NewInternalError("Error while getting chats", err)
	}
}
