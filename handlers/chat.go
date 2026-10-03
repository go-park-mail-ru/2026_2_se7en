package handlers

import (
	apperrors "app/apperrors"
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

type chatHandler struct {
	db chatDatabase
}

func NewChatHandler(db chatDatabase) *chatHandler {
	return &chatHandler{db: db}
}

func (h *chatHandler) GetListUserChats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization required", nil)
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

	writeJSON(w, http.StatusOK, struct {
		Items []*models.Chat `json:"items"`
	}{Items: chats})
}

func MapDBError(err error) (int, ErrorResponse) {
	switch {
	case errors.Is(err, apperrors.ErrSessionNotFound):
		return http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Session not found or expired",
		}
	case errors.Is(err, apperrors.ErrInvalidChatLimit):
		return http.StatusBadRequest, ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid limit",
			Details: []ErrorDetail{{Field: "limit", Reason: err.Error()}},
		}
	case errors.Is(err, apperrors.ErrInvalidChatOffset):
		return http.StatusBadRequest, ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid offset",
			Details: []ErrorDetail{{Field: "offset", Reason: err.Error()}},
		}
	case errors.Is(err, apperrors.ErrInvalidChatType):
		return http.StatusBadRequest, ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid type",
			Details: []ErrorDetail{{Field: "type", Reason: err.Error()}},
		}
	default:
		return http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error while getting chats",
		}
	}
}

func writeMapped(w http.ResponseWriter, err error) {
	status, resp := MapDBError(err)
	WriteError(w, status, resp.Code, resp.Message, resp.Details)
}

func WriteError(w http.ResponseWriter, status int, code, message string, details []ErrorDetail) {
	writeJSON(w, status, ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
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
