package auth

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/utils"
)

func (d *AuthDeliveryImpl) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(utils.SessionCookieName)
	if err != nil {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Сессия недействительна"))
		return
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Сессия недействительна"))
		return
	}

	if err := d.authUsecase.Logout(r.Context(), sessionID); err != nil {
		apperrors.WriteError(w, apperrors.NewInternalError("Что-то пошло не так", err))
		return
	}

	utils.ClearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}
