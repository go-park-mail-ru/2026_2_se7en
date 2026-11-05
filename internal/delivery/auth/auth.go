package auth

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/google/uuid"
)

type AuthDeliveryImpl struct {
	authUsecase AuthUsecase
}

func NewAuthDelivery(authUsecase AuthUsecase) *AuthDeliveryImpl {
	return &AuthDeliveryImpl{
		authUsecase: authUsecase,
	}
}

// HandleLogin godoc
// @Summary      Вход в систему
// @Description  Аутентификация пользователя по email и паролю
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Данные для входа"
// @Success      200  {object} UserResponse
// @Failure 400 {object} apperrors.ErrorResponse
// @Failure 401 {object} apperrors.ErrorResponse
// @Failure 409 {object} apperrors.ErrorResponse
// @Failure 500 {object} apperrors.ErrorResponse
// @Router       /auth/login [post]
func (d *AuthDeliveryImpl) HandleLogin(w http.ResponseWriter, r *http.Request, req *LoginRequest) *apperrors.AppError {
	result, err := d.authUsecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidCredentials) || errors.Is(err, apperrors.ErrUserNotFound) {
			return apperrors.NewUnauthorized("Incorrect email or password")
		}
		return apperrors.NewInternalError("Something went wrong", err)
	}

	middleware.SetSessionCookie(w, result.SessionID.String())

	response := ModelToUserResponse(result.User)
	tools.WriteJSON(w, http.StatusOK, response)
	return nil
}

// HandleLogout godoc
// @Summary      Выход из системы
// @Description  Удаление сессии и очистка cookie
// @Tags         auth
// @Produce      json
// @Security     SessionCookie
// @Success      204  "Сессия успешно удалена"
// @Failure 400 {object} apperrors.ErrorResponse
// @Failure 401 {object} apperrors.ErrorResponse
// @Failure 409 {object} apperrors.ErrorResponse
// @Failure 500 {object} apperrors.ErrorResponse
// @Router       /auth/logout [post]
func (d *AuthDeliveryImpl) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middleware.SessionCookieName)
	if err != nil {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Invalid session"))
		return
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Invalid session"))
		return
	}

	if err := d.authUsecase.Logout(r.Context(), sessionID); err != nil {
		apperrors.WriteError(w, apperrors.NewInternalError("Something went wrong", err))
		return
	}

	middleware.ClearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// HandleMe godoc
// @Summary      Текущий пользователь
// @Description  Получение информации о текущем авторизованном пользователе
// @Tags         auth
// @Produce      json
// @Security     SessionCookie
// @Success      200  {object} UserResponse
// @Failure 400 {object} apperrors.ErrorResponse
// @Failure 401 {object} apperrors.ErrorResponse
// @Failure 409 {object} apperrors.ErrorResponse
// @Failure 500 {object} apperrors.ErrorResponse
// @Router       /auth/me [get]
func (d *AuthDeliveryImpl) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Invalid session"))
		return
	}

	result, err := d.authUsecase.GetMe(r.Context(), userID)
	if err != nil {
		if err == apperrors.ErrUserNotFound {
			apperrors.WriteError(w, apperrors.NewUnauthorized("User not found"))
			return
		}
		apperrors.WriteError(w, apperrors.NewInternalError("Something went wrong", err))
		return
	}

	response := ModelToUserResponse(result.User)
	tools.WriteJSON(w, http.StatusOK, response)
}
