package handlers

import (
	"app/models"
	"app/storage"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/uuid"
)

type ChatHandler struct {
	Database *storage.DB
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (h *ChatHandler) GetListUserChats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Метод не поддерживается")
		return
	}

	userID := uuid.New() // здесь должна быть функция по типу GetUserIdFromSession(), но её реализую не я, поэтому пока генерим

	q := r.URL.Query()

	limit := getIntFromQuery(q, "limit")
	offset := getIntFromQuery(q, "offset")
	chatType := q.Get("type")

	chats, err := h.Database.ListUserChats(r.Context(), userID, limit, offset, chatType)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INTERNAL_ERROR", "Ошибка получения чатов")
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
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка формирования ответа")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiError{Code: code, Message: message})
}

func getIntFromQuery(query url.Values, key string) int {
	val, err := strconv.Atoi(query.Get(key))
	if err != nil || val < 0 {
		return 0
	}

	return val
}
