package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"app/apperrors"
	"app/utils"
)

var fakeHash []byte;
func init() {
	fakeHash, _ = bcrypt.GenerateFromPassword([]byte("fake"), bcrypt.DefaultCost);
}

type Profile struct {
	ID uuid.UUID
	UserID uuid.UUID
	Nickname string
	FirstName *string
	LastName *string
	Bio *string
	IconURL *string
}

type UserWithProfile struct {
	ID uuid.UUID
	Email string
	PhoneNumber *string
	PasswordHash string
	Profile Profile
}

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type ProfileResponse struct {
	ID uuid.UUID `json:"id"`
	Nickname string `json:"nickname"`
	FirstName *string `json:"first_name"`
	LastName *string `json:"last_name"`
	Bio *string `json:"bio"`
	IconURL *string `json:"icon_url"`
}

type UserResponse struct {
	ID uuid.UUID `json:"id"`
	Email string `json:"email"`
	PhoneNumber *string `json:"phone_number"`
	Profile ProfileResponse `json:"profile"`
}

type ErrorResponse struct {
	Code string `json:"code"`
	Message string `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type ErrorDetail struct {
	Field string `json:"field"`
	Reason string `json:"reason"`
}

type LoginDB interface {
	GetUserByEmail(ctx context.Context, email string) (*UserWithProfile, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*UserWithProfile, error)
}

type SessionDB interface {
	CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

type AuthHandler struct {
	db       LoginDB
	sessions SessionDB
}

func NewAuthHandler(db LoginDB, sessions SessionDB) *AuthHandler {
	return &AuthHandler{db: db, sessions: sessions}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректные данные", nil)

		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if details := validateLogin(req); len(details) > 0 {
		h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректные данные", details)

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
			h.writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Неверный email или пароль", nil)
			return
		}

		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Что-то пошло не так", nil)

		return
	}

	if !passwordOk {
		h.writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Неверный email или пароль", nil)

		return
	}

	sessionID, err := h.sessions.CreateSession(r.Context(), user.ID, time.Now().Add(utils.SessionTTL))
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Что-то пошло не так", nil)

		return
	}

	utils.SetSessionCookie(w, sessionID)
	
	utils.WriteJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(utils.SessionCookieName)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Сессия недействительна", nil)
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

func validateLogin(req LoginRequest) []ErrorDetail {
	var details []ErrorDetail

	if req.Email == "" {
		details = append(details, ErrorDetail{"email", "required"})
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
		details = append(details, ErrorDetail{"email", "Invalid_format"})
	}

	if req.Password == "" {
		details = append(details, ErrorDetail{"password", "required"})
	} else if len(req.Password) < PasswordMinLength {
        details = append(details, ErrorDetail{Field: "password", Reason: "too_short"})
    } else if len(req.Password) > PasswordMaxLength {
        details = append(details, ErrorDetail{Field: "password", Reason: "too_long"})
    }

	return details
}

func toUserResponse(u *UserWithProfile) UserResponse {
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

func (h *AuthHandler) writeError(w http.ResponseWriter, status int, code, message string, details []ErrorDetail) {
	utils.WriteJSON(w, status, ErrorResponse{Code: code, Message: message, Details: details})
}
