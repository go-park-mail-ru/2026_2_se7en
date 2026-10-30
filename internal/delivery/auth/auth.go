package auth

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"

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
func (d *AuthDeliveryImpl) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := tools.DecodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if details := validateLogin(req); len(details) > 0 {
		apperrors.WriteError(w, apperrors.NewBadRequest("Validation failed", details))
		return
	}

	result, err := d.authUsecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidCredentials) || errors.Is(err, apperrors.ErrUserNotFound) {
			apperrors.WriteError(w, apperrors.NewUnauthorized("Неверный email или пароль"))
			return
		}

		apperrors.WriteError(w, apperrors.NewInternalError("Что-то пошло не так", err))
		return
	}

	middleware.SetSessionCookie(w, result.SessionID.String())

	response := ModelToUserResponse(result.User)
	tools.WriteJSON(w, http.StatusOK, response)
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
	tools.WriteJSON(w, http.StatusOK, response)
}

func validateLogin(req LoginRequest) []apperrors.ErrorDetail {
	var details []apperrors.ErrorDetail
	if req.Email == "" {
		details = append(details, apperrors.ErrorDetail{Field: "email", Reason: "required"})
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
		details = append(details, apperrors.ErrorDetail{Field: "email", Reason: "invalid_format"})
	}

	if req.Password == "" {
		details = append(details, apperrors.ErrorDetail{Field: "password", Reason: "required"})
	} else if len(req.Password) < 8 {
		details = append(details, apperrors.ErrorDetail{Field: "password", Reason: "too_short"})
	} else if len(req.Password) > 16 {
		details = append(details, apperrors.ErrorDetail{Field: "password", Reason: "too_long"})
	}
	return details
}
