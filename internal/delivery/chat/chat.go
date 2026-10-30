package chat

import (
	"net/http"
	"net/url"
	"strconv"

	"app/internal/apperrors"
	"app/internal/middleware"
	usecasechat "app/internal/usecase/chat"
	"app/internal/utils"
)

type ChatDeliveryImpl struct {
	chatUsecase usecasechat.ChatUsecase
}

func NewChatDelivery(chatUsecase usecasechat.ChatUsecase) *ChatDeliveryImpl {
	return &ChatDeliveryImpl{chatUsecase: chatUsecase}
}

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
	utils.WriteJSON(w, http.StatusOK, response)
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
