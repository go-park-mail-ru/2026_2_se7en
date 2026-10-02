package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"uuid"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	PhoneNumber *string   `json:"phone_number"`
	Profile     Profile   `json:"profile"`
}

type Profile struct {
	ID        uuid.UUID  `json:"id"`
	Nickname  string     `json:"nickname"`
	FirstName *string    `json:"first_name"`
	LastName  *string    `json:"last_name"`
	Bio       *string    `json:"bio"`
	IconID    *uuid.UUID `json:"icon_id"`
}

type RegisterRequest struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Nickname    string  `json:"nickname"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
}

type ErrorResponse struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type ErrorDetail struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type Database interface {
	CreateUser(ctx context.Context, email, passwordHash string, phone *string) (uuid.UUID, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	CheckNicknameExists(ctx context.Context, nickname string) (bool, error)
	CreateProfile(ctx context.Context, userID uuid.UUID, nickname string, firstName, lastName string) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error)
	DeleteUserByID(ctx context.Context, userID uuid.UUID) error
}

type RegisterHandler struct {
	db Database
}

func NewHandler(db Database) *RegisterHandler {
	return &RegisterHandler{db: db}
}

func (h *RegisterHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Incorrect JSON", nil)
		return
	}

	if details := h.validateRequest(req); len(details) > 0 {
		h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Incorrect data", details)
		return
	}

	user, err := h.processRegistration(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			h.writeError(w, http.StatusConflict, "EMAIL_TAKEN", "Email already taken", []ErrorDetail{
				{Field: "email", Reason: "Email already exists"},
			})
			return
		}

		if errors.Is(err, ErrNicknameTaken) {
			h.writeError(w, http.StatusConflict, "NICKNAME_TAKEN", "Nickname already taken", []ErrorDetail{
				{Field: "nickname", Reason: "Nickname already exists"},
			})
			return
		}

		h.writeError(w, http.StatusInternalServerError, "INTERNAL_SERVICE_ERROR", "Something went wrong", nil)
		return
	}

	// TODO: Set-Cookie

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

var (
	ErrEmailTaken    = errors.New("email_taken")
	ErrNicknameTaken = errors.New("nickname_taken")
)

func (h *RegisterHandler) processRegistration(ctx context.Context, req RegisterRequest) (*User, error) {
	emailExists, err := h.db.CheckEmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if emailExists {
		return nil, ErrEmailTaken
	}

	nicknameExists, err := h.db.CheckNicknameExists(ctx, req.Nickname)
	if err != nil {
		return nil, err
	}

	if nicknameExists {
		return nil, ErrNicknameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID, err := h.db.CreateUser(ctx, req.Email, string(hash), req.PhoneNumber)
	if err != nil {
		return nil, err
	}

	err = h.db.CreateProfile(ctx, userID, req.Nickname, req.FirstName, req.LastName)
	if err != nil {
		deleteErr := h.db.DeleteUserByID(ctx, userID)
		if deleteErr != nil {
			// TODO логирование
		}

		return nil, err
	}

	return h.db.GetUserByID(ctx, userID)
}

func (h *RegisterHandler) writeError(w http.ResponseWriter, status int, code, message string, details []ErrorDetail) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}

var (
	emailRegex       = regexp.MustCompile(`^[a-zA-Z0-9_]+@[a-zA-Z]\.[a-zA-Z]{2,}$`)
	phoneNumberRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	nicknameRegex    = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

func (h *RegisterHandler) validateRequest(req RegisterRequest) []ErrorDetail {
	var details []ErrorDetail

	if req.Email == "" {
		details = append(details, ErrorDetail{Field: "email", Reason: "Email required"})
	} else if len(req.Email) > 255 {
		details = append(details, ErrorDetail{Field: "email", Reason: "Email maximum lenght is 255"})
	} else if !emailRegex.MatchString(req.Email) {
		details = append(details, ErrorDetail{Field: "email", Reason: "Incorrect email format"})
	}

	if req.Password == "" {
		details = append(details, ErrorDetail{Field: "password", Reason: "Password required"})
	} else if len(req.Password) < 8 {
		details = append(details, ErrorDetail{Field: "password", Reason: "Password minimum length is 8"})
	} else if len(req.Password) > 16 {
		details = append(details, ErrorDetail{Field: "password", Reason: "Password maximum length is 16"})
	}

	if req.PhoneNumber != nil && *req.PhoneNumber != "" {
		if !phoneNumberRegex.MatchString(*req.PhoneNumber) {
			details = append(details, ErrorDetail{Field: "phone_number", Reason: "Incorrect phone number format"})
		}
	}

	if req.Nickname == "" {
		details = append(details, ErrorDetail{Field: "nickname", Reason: "Nickname required"})
	} else if len(req.Nickname) < 3 {
		details = append(details, ErrorDetail{Field: "nickname", Reason: "Nickname minimum length is 3"})
	} else if len(req.Nickname) > 16 {
		details = append(details, ErrorDetail{Field: "nickname", Reason: "Nickname maximum length is 16"})
	} else if !nicknameRegex.MatchString(req.Nickname) {
		details = append(details, ErrorDetail{Field: "nickname", Reason: "Incorrect nickname format"})
	}

	if req.FirstName != "" {
		if len(req.FirstName) > 32 {
			details = append(details, ErrorDetail{Field: "first_name", Reason: "First name maximum length is 32"})
		}
	}

	if req.LastName != "" {
		if len(req.LastName) > 32 {
			details = append(details, ErrorDetail{Field: "last_name", Reason: "Last name maximum length is 32"})
		}
	}

	return details
}
