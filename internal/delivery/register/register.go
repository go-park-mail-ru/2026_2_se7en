package register

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"

	"app/internal/apperrors"
	ucregister "app/internal/usecase/register"
	"app/internal/utils"
)

type RegisterDeliveryImpl struct {
	registerUsecase ucregister.RegisterUsecase
}

func NewRegisterDelivery(registerUsecase ucregister.RegisterUsecase) *RegisterDeliveryImpl {
	return &RegisterDeliveryImpl{
		registerUsecase: registerUsecase,
	}
}

func (d *RegisterDeliveryImpl) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	if req.PhoneNumber != nil {
		phone := strings.TrimSpace(*req.PhoneNumber)
		if phone == "" {
			req.PhoneNumber = nil
		} else {
			req.PhoneNumber = &phone
		}
	}

	req.FirstName = strings.TrimSpace(req.FirstName)
	if req.LastName != nil {
		lastName := strings.TrimSpace(*req.LastName)
		if lastName == "" {
			req.LastName = nil
		} else {
			req.LastName = &lastName
		}
	}

	if details := d.validateRequest(req); len(details) > 0 {
		apperrors.WriteError(w, apperrors.NewBadRequest("Validation failed", details))
		return
	}

	userModel := RegisterRequestToModel(&req)

	registerRes, err := d.registerUsecase.Register(r.Context(), userModel, req.Password)
	if err != nil {
		if errors.Is(err, apperrors.ErrEmailTaken) {
			apperrors.WriteError(w, apperrors.NewConflictWithCode(apperrors.ErrEmailTaken.Code, "Email already taken", []apperrors.ErrorDetail{
				{Field: "email", Reason: "Email already exists"},
			}))
			return
		}

		if errors.Is(err, apperrors.ErrNicknameTaken) {
			apperrors.WriteError(w, apperrors.NewConflictWithCode(apperrors.ErrNicknameTaken.Code, "Nickname already taken", []apperrors.ErrorDetail{
				{Field: "nickname", Reason: "Nickname already exists"},
			}))
			return
		}

		apperrors.WriteError(w, apperrors.NewInternalError("Something went wrong", err))
		return
	}

	response := ModelToRegisterResponse(registerRes.User)

	utils.SetSessionCookie(w, registerRes.SessionID.String())

	utils.WriteJSON(w, http.StatusCreated, response)
}

func (h *RegisterDeliveryImpl) validateRequest(req RegisterRequest) []apperrors.ErrorDetail {
	var details []apperrors.ErrorDetail

	if req.Email == "" {
		details = append(details, apperrors.ErrorDetail{Field: "email", Reason: "Email required"})
	} else if len(req.Email) > 255 {
		details = append(details, apperrors.ErrorDetail{Field: "email", Reason: "Email maximum length is 255"})
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
		details = append(details, apperrors.ErrorDetail{Field: "email", Reason: "Incorrect email format"})
	}

	if req.Password == "" {
		details = append(details, apperrors.ErrorDetail{Field: "password", Reason: "Password required"})
	} else if len(req.Password) < 8 {
		details = append(details, apperrors.ErrorDetail{Field: "password", Reason: "Password minimum length is 8"})
	} else if len(req.Password) > 16 {
		details = append(details, apperrors.ErrorDetail{Field: "password", Reason: "Password maximum length is 16"})
	}

	if req.PhoneNumber != nil && *req.PhoneNumber != "" {
		if !utils.IsPhoneNumberValid(*req.PhoneNumber) {
			details = append(details, apperrors.ErrorDetail{Field: "phone_number", Reason: "Incorrect phone number format"})
		}
	}

	if req.Nickname == "" {
		details = append(details, apperrors.ErrorDetail{Field: "nickname", Reason: "Nickname required"})
	} else if len(req.Nickname) < 3 {
		details = append(details, apperrors.ErrorDetail{Field: "nickname", Reason: "Nickname minimum length is 3"})
	} else if len(req.Nickname) > 16 {
		details = append(details, apperrors.ErrorDetail{Field: "nickname", Reason: "Nickname maximum length is 16"})
	} else if !utils.IsNicknameValid(req.Nickname) {
		details = append(details, apperrors.ErrorDetail{Field: "nickname", Reason: "Incorrect nickname format"})
	}

	firstNameLength := utf8.RuneCountInString(strings.TrimSpace(req.FirstName))
	if firstNameLength == 0 {
		details = append(details, apperrors.ErrorDetail{Field: "first_name", Reason: "First name required"})
	} else if strings.ContainsAny(req.FirstName, "0123456789") {
		details = append(details, apperrors.ErrorDetail{Field: "first_name", Reason: "First name must not contain digits"})
	} else if firstNameLength < 2 {
		details = append(details, apperrors.ErrorDetail{Field: "first_name", Reason: "First name minimum length is 2"})
	} else if firstNameLength > 32 {
		details = append(details, apperrors.ErrorDetail{Field: "first_name", Reason: "First name maximum length is 32"})
	}

	if req.LastName != nil && strings.TrimSpace(*req.LastName) != "" {
		lastNameLength := utf8.RuneCountInString(strings.TrimSpace(*req.LastName))
		if lastNameLength < 2 {
			details = append(details, apperrors.ErrorDetail{Field: "last_name", Reason: "Last name minimum length is 2"})
		} else if lastNameLength > 32 {
			details = append(details, apperrors.ErrorDetail{Field: "last_name", Reason: "Last name maximum length is 32"})
		}
	}

	return details
}
