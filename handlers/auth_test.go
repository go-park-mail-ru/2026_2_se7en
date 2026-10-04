package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"app/apperrors"
	"app/middleware"
	"app/models"
	"app/utils"
)

type mockLoginDB struct {
	user    *models.User
	profile *models.Profile
	err     error
}

func (m *mockLoginDB) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.user, m.err
}

func (m *mockLoginDB) FindUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return m.user, m.err
}

func (m *mockLoginDB) FindProfileByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	return m.profile, nil
}

type mockSessionDB struct {
	sessionID uuid.UUID
	createErr error
}

func (m *mockSessionDB) CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*models.Session, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return &models.Session{ID: m.sessionID, UserID: userID, ExpiresAt: expiresAt}, nil
}

func (m *mockSessionDB) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	return nil
}

func makeUser(password string) *models.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return &models.User{
		ID:           uuid.New(),
		Email:        "user@example.ru",
		PasswordHash: string(hash),
	}
}

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name       string
		req        LoginRequest
		wantField  string
		wantReason string
	}{
		{"empty email", LoginRequest{Password: "password123"}, "email", "required"},
		{"invalid email", LoginRequest{Email: "bad", Password: "password123"}, "email", "invalid_format"},
		{"empty password", LoginRequest{Email: "user@example.ru"}, "password", "required"},
		{"short password", LoginRequest{Email: "user@example.ru", Password: "short"}, "password", "too_short"},
		{"long password", LoginRequest{Email: "user@example.ru", Password: "verylongpassword12345"}, "password", "too_long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			details := validateLogin(tt.req)
			if len(details) == 0 {
				t.Fatalf("expected errors, got none")
			}

			found := false
			for _, d := range details {
				if d.Field == tt.wantField && d.Reason == tt.wantReason {
					found = true
				}
			}
			if !found {
				t.Errorf("expected %s/%s in %+v", tt.wantField, tt.wantReason, details)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		db         *mockLoginDB
		sessions   *mockSessionDB
		wantStatus int
		wantCookie bool
	}{
		{
			name:       "success",
			body:       `{"email":"user@example.ru","password":"password123"}`,
			db:         &mockLoginDB{user: makeUser("password123")},
			sessions:   &mockSessionDB{sessionID: uuid.New()},
			wantStatus: http.StatusOK,
			wantCookie: true,
		},
		{
			name:       "invalid json",
			body:       `not json`,
			db:         &mockLoginDB{},
			sessions:   &mockSessionDB{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty password",
			body:       `{"email":"user@example.ru","password":""}`,
			db:         &mockLoginDB{},
			sessions:   &mockSessionDB{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "short password",
			body:       `{"email":"user@example.ru","password":"short"}`,
			db:         &mockLoginDB{},
			sessions:   &mockSessionDB{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "user not found",
			body:       `{"email":"user@example.ru","password":"password228"}`,
			db:         &mockLoginDB{err: apperrors.ErrUserNotFound},
			sessions:   &mockSessionDB{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong password",
			body:       `{"email":"user@example.ru","password":"wrongpassword"}`,
			db:         &mockLoginDB{user: makeUser("password123")},
			sessions:   &mockSessionDB{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "db error",
			body:       `{"email":"user@example.ru","password":"password228"}`,
			db:         &mockLoginDB{err: errors.New("db down")},
			sessions:   &mockSessionDB{},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAuthHandler(tt.db, tt.sessions)
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			h.Login(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantCookie {
				cookies := rec.Result().Cookies()
				if len(cookies) == 0 {
					t.Errorf("expected Set-Cookie, got none")
				}
			}
		})
	}
}

func TestLogout(t *testing.T) {
	t.Run("no cookie", func(t *testing.T) {
		h := NewAuthHandler(&mockLoginDB{}, &mockSessionDB{})
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		rec := httptest.NewRecorder()

		h.Logout(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("success", func(t *testing.T) {
		h := NewAuthHandler(&mockLoginDB{}, &mockSessionDB{})
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.AddCookie(&http.Cookie{Name: utils.SessionCookieName, Value: uuid.NewString()})
		rec := httptest.NewRecorder()

		h.Logout(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}

		cookies := rec.Result().Cookies()
		if len(cookies) == 0 || cookies[0].MaxAge != -1 {
			t.Errorf("expected cookie cleared (MaxAge=-1)")
		}
	})
}

func TestCurrentUser(t *testing.T) {
	user := makeUser("password123")
	profile := &models.Profile{FirstName: "Анна", Nickname: "anna"}
	h := NewAuthHandler(&mockLoginDB{user: user, profile: profile}, &mockSessionDB{})
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(middleware.NewAuthContext(req.Context(), user.ID))
	rec := httptest.NewRecorder()

	h.CurrentUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	var response UserResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Profile.FirstName != "Анна" {
		t.Errorf("first_name = %q, want Анна", response.Profile.FirstName)
	}
	if !strings.Contains(rec.Body.String(), `"first_name":"Анна"`) {
		t.Errorf("response does not contain profile.first_name: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "password_hash") {
		t.Error("response exposes password hash")
	}
}

func TestRegisterNameLengths(t *testing.T) {
	base := RegisterRequest{
		Email: "user@example.com", Password: "password", Nickname: "user", FirstName: "Анна",
	}
	tests := []struct {
		name       string
		firstName  string
		lastName   *string
		wantField  string
		wantReason string
	}{
		{name: "first name one character", firstName: "Я", wantField: "first_name", wantReason: "First name minimum length is 2"},
		{name: "first name two characters", firstName: "Ян"},
		{name: "first name 32 Cyrillic characters", firstName: strings.Repeat("Я", 32)},
		{name: "first name 33 Cyrillic characters", firstName: strings.Repeat("Я", 33), wantField: "first_name", wantReason: "First name maximum length is 32"},
		{name: "last name omitted", firstName: base.FirstName},
		{name: "last name empty", firstName: base.FirstName, lastName: namePointer("")},
		{name: "last name one character", firstName: base.FirstName, lastName: namePointer("Я"), wantField: "last_name", wantReason: "Last name minimum length is 2"},
		{name: "last name two characters", firstName: base.FirstName, lastName: namePointer("Ян")},
		{name: "last name 32 Cyrillic characters", firstName: base.FirstName, lastName: namePointer(strings.Repeat("Я", 32))},
		{name: "last name 33 Cyrillic characters", firstName: base.FirstName, lastName: namePointer(strings.Repeat("Я", 33)), wantField: "last_name", wantReason: "Last name maximum length is 32"},
	}

	h := NewHandler(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := base
			req.FirstName = tt.firstName
			req.LastName = tt.lastName
			details := h.validateRequest(req)
			if tt.wantField == "" {
				if len(details) != 0 {
					t.Fatalf("unexpected validation errors: %+v", details)
				}
				return
			}
			if len(details) != 1 || details[0].Field != tt.wantField || details[0].Reason != tt.wantReason {
				t.Fatalf("validation errors = %+v, want %s: %s", details, tt.wantField, tt.wantReason)
			}
		})
	}
}

func namePointer(value string) *string { return &value }
