package helpers

import (
	"app/apperrors"
	"encoding/json"
	"net/http"
)

func DecodeJSON(r *http.Request, v any) *apperrors.AppError {
	if r.Body == nil {
		return apperrors.NewBadRequest("Request body is empty", nil)
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(v); err != nil {
		return apperrors.NewBadRequest("Invalid JSON format", nil)
	}

	return nil
}
