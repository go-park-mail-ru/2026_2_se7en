package handlers

import (
	"app/middleware"
	"app/models"
	"context"
	"encoding/json"
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

type ChatHandler struct {
	db chatDatabase
}

func NewChatHandler(db chatDatabase) *ChatHandler {
	return &ChatHandler{db: db}
}

func (h *ChatHandler) GetListUserChats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusBadRequest, "UNAUTHORIZED", "Authorization required", nil)
		return
	}

	q := r.URL.Query()

	limit := getIntFromQuery(q, "limit")
	offset := getIntFromQuery(q, "offset")
	chatType := q.Get("type")

	chats, err := h.db.ListUserChats(r.Context(), userID, limit, offset, chatType)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INTERNAL_ERROR", "Error while getting chats", nil)
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

func getIntFromQuery(query url.Values, key string) int {
	val, err := strconv.Atoi(query.Get(key))
	if err != nil || val < 0 {
		return 0
	}

	return val
}
