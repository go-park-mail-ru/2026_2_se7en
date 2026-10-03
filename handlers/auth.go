package handlers

import (
	"app/apperrors"
	"app/helpers"
	"app/models"
	"app/storage"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Nickname    string  `json:"nickname"`
	FirstName   string  `json:"first_name"`
	LastName    *string `json:"last_name"`
}
type RegisterHandler struct {
	db storage.DB
}

func NewHandler(db storage.DB) *RegisterHandler {
	return &RegisterHandler{db: db}
}

func (h *RegisterHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := helpers.DecodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}

	if details := h.validateRequest(req); len(details) > 0 {
		apperrors.WriteError(w, apperrors.NewBadRequest("Validation failed", details))
		return
	}

	user, err := h.processRegistration(r.Context(), req)
	if err != nil {
		if errors.Is(err, apperrors.ErrEmailTaken) {
			apperrors.WriteError(w, apperrors.NewConflict("Email already taken", []apperrors.ErrorDetail{
				{Field: "email", Reason: "Email already exists"},
			}))
			return
		}

		if errors.Is(err, apperrors.ErrNicknameTaken) {
			apperrors.WriteError(w, apperrors.NewConflict("Nickname already taken", []apperrors.ErrorDetail{
				{Field: "nickname", Reason: "Nickname already exists"},
			}))
			return
		}

		apperrors.WriteError(w, apperrors.NewInternalError("Something went wrong", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *RegisterHandler) processRegistration(ctx context.Context, req RegisterRequest) (*models.User, error) {
	emailExists, err := h.db.CheckEmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if emailExists {
		return nil, apperrors.ErrEmailTaken
	}

	nicknameExists, err := h.db.CheckNicknameExists(ctx, req.Nickname)
	if err != nil {
		return nil, err
	}

	if nicknameExists {
		return nil, apperrors.ErrNicknameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID, err := h.db.CreateUser(ctx, req.Email, string(hash), req.PhoneNumber)
	if err != nil {
		return nil, err
	}

	err = h.db.CreateProfile(ctx, userID, req.Nickname, req.FirstName, *req.LastName)
	if err != nil {
		deleteErr := h.db.DeleteUserByID(ctx, userID)
		if deleteErr != nil {
			// TODO логирование
		}

		return nil, err
	}

	return h.db.GetUserByID(ctx, userID)
}

func (h *RegisterHandler) validateRequest(req RegisterRequest) []apperrors.ErrorDetail {
	var details []apperrors.ErrorDetail

	if req.Email == "" {
		details = append(details, apperrors.ErrorDetail{Field: "email", Reason: "Email required"})
	} else if len(req.Email) > 255 {
		details = append(details, apperrors.ErrorDetail{Field: "email", Reason: "Email maximum length is 255"})
	} else if !helpers.IsEmailValid(req.Email) {
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
		if !helpers.IsPhoneNumberValid(*req.PhoneNumber) {
			details = append(details, apperrors.ErrorDetail{Field: "phone_number", Reason: "Incorrect phone number format"})
		}
	}

	if req.Nickname == "" {
		details = append(details, apperrors.ErrorDetail{Field: "nickname", Reason: "Nickname required"})
	} else if len(req.Nickname) < 3 {
		details = append(details, apperrors.ErrorDetail{Field: "nickname", Reason: "Nickname minimum length is 3"})
	} else if len(req.Nickname) > 16 {
		details = append(details, apperrors.ErrorDetail{Field: "nickname", Reason: "Nickname maximum length is 16"})
	} else if !helpers.IsNicknameValid(req.Nickname) {
		details = append(details, apperrors.ErrorDetail{Field: "nickname", Reason: "Incorrect nickname format"})
	}

	if req.FirstName != "" {
		if len(req.FirstName) > 32 {
			details = append(details, apperrors.ErrorDetail{Field: "first_name", Reason: "First name maximum length is 32"})
		}
	}

	if req.LastName != nil && *req.LastName != "" {
		if len(*req.LastName) > 32 {
			details = append(details, apperrors.ErrorDetail{Field: "last_name", Reason: "Last name maximum length is 32"})
		}
	}

	return details
}
