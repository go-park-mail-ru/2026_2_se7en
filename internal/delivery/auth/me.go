package auth

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/utils"
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
