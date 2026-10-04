package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"app/apperrors"
	"app/helpers"
	"app/models"
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
	FirstName string
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
	LastName    *string `json:"last_name,omitempty"`
}

type ProfileResponse struct {
	ID        uuid.UUID `json:"id"`
	Nickname  string    `json:"nickname"`
	FirstName string    `json:"first_name"`
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
	CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*models.Session, error)
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error
}

type LoginDB interface {
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	FindProfileByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
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
	if err := helpers.DecodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if details := validateLogin(req); len(details) > 0 {
		apperrors.WriteError(w, apperrors.NewBadRequest("Validation failed", details))
		return
	}

	user, err := h.db.FindUserByEmail(r.Context(), req.Email)
	hash := fakeHash
	if err == nil && user != nil {
		hash = []byte(user.PasswordHash)
	}
	passwordOk := bcrypt.CompareHashAndPassword(hash, []byte(req.Password)) == nil

	if err != nil && !errors.Is(err, apperrors.ErrUserNotFound) {
		apperrors.WriteError(w, apperrors.NewInternalError("Что-то пошло не так", err))
		return
	}

	if user == nil || err != nil || !passwordOk {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Неверный email или пароль"))
		return
	}

	profile, err := h.db.FindProfileByUserID(r.Context(), user.ID)
	if err != nil {
		apperrors.WriteError(w, apperrors.NewInternalError("Что-то пошло не так", err))
		return
	}

	session, err := h.sessions.CreateSession(r.Context(), user.ID, time.Now().Add(utils.SessionTTL))
	if err != nil {
		apperrors.WriteError(w, apperrors.NewInternalError("Что-то пошло не так", err))
		return
	}

	utils.SetSessionCookie(w, session.ID.String())
	utils.WriteJSON(w, http.StatusOK, toUserResponse(userWithProfile(user, profile)))
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
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
	if err := h.sessions.DeleteSession(r.Context(), sessionID); err != nil {
		apperrors.WriteError(w, apperrors.NewInternalError("Что-то пошло не так", err))
		return
	}

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
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
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

func userWithProfile(user *models.User, profile *models.Profile) *User {
	result := &User{
		ID:           user.ID,
		Email:        user.Email,
		PhoneNumber:  user.PhoneNumber,
		PasswordHash: user.PasswordHash,
	}
	if profile != nil {
		result.Profile = Profile{
			ID:        profile.ID,
			Nickname:  profile.Nickname,
			FirstName: profile.FirstName,
			LastName:  profile.LastName,
			Bio:       profile.Bio,
		}
	}
	return result
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

type RegistrationDB interface {
	CheckEmailExist(ctx context.Context, email string) (bool, error)
	CheckNicknameExist(ctx context.Context, nickname string) (bool, error)
	CreateUser(ctx context.Context, email, passwordHash string, phone *string) (*models.User, error)
	CreateProfile(ctx context.Context, iconID *uuid.UUID, userID uuid.UUID, nickname, firstName string, lastName *string, bio *string) (*models.Profile, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

type RegisterHandler struct {
	db RegistrationDB
}

func NewHandler(db RegistrationDB) *RegisterHandler {
	return &RegisterHandler{db: db}
}

func (h *RegisterHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := helpers.DecodeJSON(r, &req); err != nil {
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

	if details := h.validateRequest(req); len(details) > 0 {
		apperrors.WriteError(w, apperrors.NewBadRequest("Validation failed", details))
		return
	}

	user, err := h.processRegistration(r.Context(), req)
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

	utils.WriteJSON(w, http.StatusCreated, toUserResponse(user))
}

func (h *RegisterHandler) processRegistration(ctx context.Context, req RegisterRequest) (*User, error) {
	emailExists, err := h.db.CheckEmailExist(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if emailExists {
		return nil, apperrors.ErrEmailTaken
	}

	nicknameExists, err := h.db.CheckNicknameExist(ctx, req.Nickname)
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

	modelUser, err := h.db.CreateUser(ctx, req.Email, string(hash), req.PhoneNumber)
	if err != nil {
		return nil, err
	}

	profile, err := h.db.CreateProfile(ctx, nil, modelUser.ID, req.Nickname, req.FirstName, req.LastName, nil)
	if err != nil {
		_ = h.db.DeleteUser(ctx, modelUser.ID)
		return nil, err
	}

	return userWithProfile(modelUser, profile), nil
}

func (h *RegisterHandler) validateRequest(req RegisterRequest) []apperrors.ErrorDetail {
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

	if strings.TrimSpace(req.FirstName) == "" {
		details = append(details, apperrors.ErrorDetail{Field: "first_name", Reason: "First name required"})
	} else if len(req.FirstName) > 32 {
		details = append(details, apperrors.ErrorDetail{Field: "first_name", Reason: "First name maximum length is 32"})
	}

	if req.LastName != nil && *req.LastName != "" {
		if len(*req.LastName) > 32 {
			details = append(details, apperrors.ErrorDetail{Field: "last_name", Reason: "Last name maximum length is 32"})
		}
	}

	return details
}
