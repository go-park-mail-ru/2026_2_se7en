package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"app/apperrors"
	"app/helpers"
	"app/storage"
	"app/utils"
)

var fakeHash []byte

func init() {
	fakeHash, _ = bcrypt.GenerateFromPassword([]byte("fake"), bcrypt.DefaultCost)
}

type User struct {
	ID           uuid.UUID
	Email        string
	PhoneNumber  *string
	PasswordHash string
	Profile      Profile
}

type Profile struct {
	ID        uuid.UUID
	Nickname  string
	FirstName *string
	LastName  *string
	Bio       *string
	IconURL   *string
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Nickname    string  `json:"nickname"`
	FirstName   string  `json:"first_name"`
	LastName    *string `json:"last_name"`
}

type ProfileResponse struct {
	ID        uuid.UUID `json:"id"`
	Nickname  string    `json:"nickname"`
	FirstName *string   `json:"first_name"`
	LastName  *string   `json:"last_name"`
	Bio       *string   `json:"bio"`
	IconURL   *string   `json:"icon_url"`
}

type UserResponse struct {
	ID          uuid.UUID       `json:"id"`
	Email       string          `json:"email"`
	PhoneNumber *string         `json:"phone_number"`
	Profile     ProfileResponse `json:"profile"`
}

type SessionDB interface {
	CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

type AuthHandler struct {
	db       storage.DB
	sessions SessionDB
}

func NewAuthHandler(db storage.DB, sessions SessionDB) *AuthHandler {
	return &AuthHandler{db: db, sessions: sessions}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := helpers.DecodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if details := validateLogin(req); len(details) > 0 {
		apperrors.WriteError(w, apperrors.NewBadRequest("Validation failed", details))
		return
	}

	user, dbErr := h.db.GetUserByEmail(r.Context(), req.Email)

	hash := fakeHash
	if dbErr == nil && user != nil {
		hash = []byte(user.PasswordHash)
	}

	passwordOk := bcrypt.CompareHashAndPassword(hash, []byte(req.Password)) == nil

	if dbErr != nil {
		if errors.Is(dbErr, apperrors.ErrUserNotFound) {
			apperrors.WriteError(w, apperrors.NewUnauthorized("Неверный email или пароль", nil))
			return
		}

		apperrors.WriteError(w, apperrors.NewInternalError("Что-то пошло не так", dbErr))
		return
	}

	if !passwordOk {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Неверный email или пароль", nil))
		return
	}

	sessionID, err := h.sessions.CreateSession(r.Context(), user.ID, time.Now().Add(utils.SessionTTL))
	if err != nil {
		apperrors.WriteError(w, apperrors.NewInternalError("Что-то пошло не так", err))
		return
	}

	utils.SetSessionCookie(w, sessionID)
	utils.WriteJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(utils.SessionCookieName)
	if err != nil {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Сессия недействительна", nil))
		return
	}

	_ = h.sessions.DeleteSession(r.Context(), cookie.Value)

	utils.ClearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

const (
	PasswordMinLength = 8
	PasswordMaxLength = 16
)

func validateLogin(req LoginRequest) []apperrors.ErrorDetail {
	var details []apperrors.ErrorDetail

	if req.Email == "" {
		details = append(details, apperrors.ErrorDetail{Field: "email", Reason: "required"})
	} else if !helpers.IsEmailValid(req.Email) {
		details = append(details, apperrors.ErrorDetail{Field: "email", Reason: "invalid_format"})
	}

	if req.Password == "" {
		details = append(details, apperrors.ErrorDetail{Field: "password", Reason: "required"})
	} else if len(req.Password) < PasswordMinLength {
		details = append(details, apperrors.ErrorDetail{Field: "password", Reason: "too_short"})
	} else if len(req.Password) > PasswordMaxLength {
		details = append(details, apperrors.ErrorDetail{Field: "password", Reason: "too_long"})
	}

	return details
}

func toUserResponse(u *User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		PhoneNumber: u.PhoneNumber,
		Profile: ProfileResponse{
			ID:        u.Profile.ID,
			Nickname:  u.Profile.Nickname,
			FirstName: u.Profile.FirstName,
			LastName:  u.Profile.LastName,
			Bio:       u.Profile.Bio,
			IconURL:   u.Profile.IconURL,
		},
	}
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

func (h *RegisterHandler) processRegistration(ctx context.Context, req RegisterRequest) (*User, error) {
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
		_ = h.db.DeleteUserByID(ctx, userID)
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
