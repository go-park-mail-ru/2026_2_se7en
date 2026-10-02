package handlers

import (
	apperrors "app/app_errors"
	"app/middleware"
	"app/models"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/uuid"
)

type ErrorResponse struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type ErrorDetail struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type chatDatabase interface {
	ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.Chat, error)
}

type сhatHandler struct {
	db chatDatabase
}

func NewChatHandler(db chatDatabase) *сhatHandler {
	return &сhatHandler{db: db}
}

func (h *сhatHandler) GetListUserChats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization required", nil)
		return
	}

	q := r.URL.Query()

	limit, err := getIntFromQuery(q, "limit")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid limit", []ErrorDetail{{Field: "limit", Reason: err.Error()}})
		return
	}
	offset, err := getIntFromQuery(q, "offset")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid offset", []ErrorDetail{{Field: "offset", Reason: err.Error()}})
		return
	}
	chatType := ""
	if values, present := q["type"]; present {
		if len(values) != 1 || values[0] == "" {
			WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid type", []ErrorDetail{{Field: "type", Reason: apperrors.ErrInvalidChatType.Error()}})
			return
		}
		chatType = values[0]
	}

	chats, err := h.db.ListUserChats(r.Context(), userID, limit, offset, chatType)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrSessionNotFound):
			WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Session not found or expired", nil)
		case errors.Is(err, apperrors.ErrInvalidChatLimit):
			WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid limit", []ErrorDetail{{Field: "limit", Reason: err.Error()}})
		case errors.Is(err, apperrors.ErrInvalidChatOffset):
			WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid offset", []ErrorDetail{{Field: "offset", Reason: err.Error()}})
		case errors.Is(err, apperrors.ErrInvalidChatType):
			WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid type", []ErrorDetail{{Field: "type", Reason: err.Error()}})
		default:
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error while getting chats", nil)
		}
		return
	}

	if chats == nil {
		chats = []*models.Chat{}
	}

	response := struct {
		Items []*models.Chat `json:"items"`
	}{Items: chats}

	body, err := json.Marshal(response)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func WriteError(w http.ResponseWriter, status int, code, message string, details []ErrorDetail) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
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
