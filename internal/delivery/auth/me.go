package auth

import (
	"net/http"

	"app/internal/apperrors"
	"app/internal/middleware"
	"app/internal/utils"
)

func (d *AuthDeliveryImpl) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Сессия недействительна"))
		return
	}

	result, err := d.authUsecase.GetMe(r.Context(), userID)
	if err != nil {
		if err == apperrors.ErrUserNotFound {
			apperrors.WriteError(w, apperrors.NewUnauthorized("Пользователь не найден"))
			return
		}
		apperrors.WriteError(w, apperrors.NewInternalError("Что-то пошло не так", err))
		return
	}

	response := ModelToUserResponse(result.User)
	utils.WriteJSON(w, http.StatusOK, response)
}
