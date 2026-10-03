package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"app/apperrors"
	"app/utils"
)

type mockLoginDB struct {
	user *UserWithProfile
	err  error
}

func (m *mockLoginDB) GetUserByEmail(ctx context.Context, email string) (*UserWithProfile, error) {
	return m.user, m.err
}

func (m *mockLoginDB) GetUserByID(ctx context.Context, userID uuid.UUID) (*UserWithProfile, error) {
	return m.user, m.err
}

type mockSessionDB struct {
	sessionID string
	createErr error
}

func (m *mockSessionDB) CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (string, error) {
	return m.sessionID, m.createErr
}

func (m *mockSessionDB) DeleteSession(ctx context.Context, sessionID string) error {
	return nil
}

func makeUser(password string) *UserWithProfile {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return &UserWithProfile{
		ID:           uuid.New(),
		Email:        "user@example.ru",
		PasswordHash: string(hash),
		Profile:      Profile{ID: uuid.New(), Nickname: "nick"},
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
		{"invalid email", LoginRequest{Email: "bad", Password: "password123"}, "email", "Invalid_format"},
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
			sessions:   &mockSessionDB{sessionID: "sid"},
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
		req.AddCookie(&http.Cookie{Name: utils.SessionCookieName, Value: "sid"})
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
