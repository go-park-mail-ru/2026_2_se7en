package tools

import (
	"encoding/json"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
)

func DecodeJSON(r *http.Request, v any) *apperrors.AppError {
	if r.Body == nil {
		return apperrors.NewBadRequest("Request body is empty", nil)
	}
	defer func() {
		_ = r.Body.Close()
	}()

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(v); err != nil {
		return apperrors.NewBadRequest("Invalid JSON format", nil)
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
