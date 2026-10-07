package handlers

import (
	apperrors "app/apperrors"
	"app/middleware"
	"app/models"
	"app/utils"
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/uuid"
)

type chatDatabase interface {
	ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.Chat, error)
}

type chatHandler struct {
	db chatDatabase
}

func NewChatHandler(db chatDatabase) *chatHandler {
	return &chatHandler{db: db}
}

func (h *chatHandler) GetListUserChats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}

	q := r.URL.Query()

	limit, err := getIntFromQuery(q, "limit")
	if err != nil {
		writeMapped(w, err)
		return
	}

	offset, err := getIntFromQuery(q, "offset")
	if err != nil {
		writeMapped(w, err)
		return
	}

	chatType := ""
	if values, present := q["type"]; present {
		if len(values) != 1 || values[0] == "" {
			writeMapped(w, apperrors.ErrInvalidChatType)
			return
		}
		chatType = values[0]
	}

	chats, err := h.db.ListUserChats(r.Context(), userID, limit, offset, chatType)
	if err != nil {
		writeMapped(w, err)
		return
	}

	if chats == nil {
		chats = []*models.Chat{}
	}

	utils.WriteJSON(w, http.StatusOK, struct {
		Items []*models.Chat `json:"items"`
	}{Items: chats})
}

func MapDBError(err error) *apperrors.AppError {
	switch {
	case errors.Is(err, apperrors.ErrSessionNotFound):
		return apperrors.NewUnauthorized("Session not found or expired")
	case errors.Is(err, apperrors.ErrInvalidChatLimit):
		return apperrors.NewBadRequest("Invalid limit", []apperrors.ErrorDetail{{Field: "limit", Reason: err.Error()}})
	case errors.Is(err, apperrors.ErrInvalidChatOffset):
		return apperrors.NewBadRequest("Invalid offset", []apperrors.ErrorDetail{{Field: "offset", Reason: err.Error()}})
	case errors.Is(err, apperrors.ErrInvalidChatType):
		return apperrors.NewBadRequest("Invalid type", []apperrors.ErrorDetail{{Field: "type", Reason: err.Error()}})
	default:
		return apperrors.NewInternalError("Error while getting chats", err)
	}
}

func writeMapped(w http.ResponseWriter, err error) {
	apperrors.WriteError(w, MapDBError(err))
}

func getIntFromQuery(query url.Values, key string) (int, error) {
	values, present := query[key]
	if !present {
		return 0, nil
	}
	if len(values) != 1 {
		if key == "limit" {
			return 0, apperrors.ErrInvalidChatLimit
		}
		return 0, apperrors.ErrInvalidChatOffset
	}

	val, err := strconv.Atoi(values[0])
	if key == "limit" {
		if err != nil || val < 1 {
			return 0, apperrors.ErrInvalidChatLimit
		}
	} else if err != nil || val < 0 {
		return 0, apperrors.ErrInvalidChatOffset
	}

	return val, nil
}
