package auth

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"app/apperrors"
	ucauth "app/usecase/auth"
	"app/utils"
)

type AuthDeliveryImpl struct {
	authUsecase ucauth.AuthUsecase
}

func NewAuthDelivery(authUsecase ucauth.AuthUsecase) *AuthDeliveryImpl {
	return &AuthDeliveryImpl{
		authUsecase: authUsecase,
	}
}

func (d *AuthDeliveryImpl) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
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

	utils.SetSessionCookie(w, result.SessionID.String())

	response := ModelToUserResponse(result.User)
	utils.WriteJSON(w, http.StatusOK, response)
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
