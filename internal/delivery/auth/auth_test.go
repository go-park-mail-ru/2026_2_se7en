package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"app/internal/apperrors"
	"app/internal/middleware"
	"app/internal/models"
	"app/internal/utils"
)

type mockAuthUsecase struct {
	loginResult *models.UserWithSession
	loginErr    error
	meResult    *models.UserWithSession
	meErr       error
	logoutErr   error

	loginEmail string
}

func (m *mockAuthUsecase) Login(ctx context.Context, email, password string) (*models.UserWithSession, error) {
	m.loginEmail = email
	return m.loginResult, m.loginErr
}

func (m *mockAuthUsecase) GetMe(ctx context.Context, userID uuid.UUID) (*models.UserWithSession, error) {
	return m.meResult, m.meErr
}

func (m *mockAuthUsecase) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return m.logoutErr
}

func TestAuthDelivery_HandleLogin_Success(t *testing.T) {
	userID := uuid.New()
	mockUC := &mockAuthUsecase{
		loginResult: &models.UserWithSession{
			User: &models.User{
				ID:    userID,
				Email: "test@example.com",
				Profile: &models.Profile{
					ID:        uuid.New(),
					Nickname:  "testnick",
					FirstName: "Test",
				},
			},
			SessionID: uuid.New(),
		},
	}
	delivery := NewAuthDelivery(mockUC)

	body := `{"email":"  TEST@EXAMPLE.COM  ","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()

	delivery.HandleLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	if mockUC.loginEmail != "test@example.com" {
		t.Errorf("expected normalized email 'test@example.com', got %q", mockUC.loginEmail)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != utils.SessionCookieName {
		t.Errorf("expected session cookie to be set")
	}

	var resp UserResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Email != "test@example.com" || resp.Profile.Nickname != "testnick" {
		t.Errorf("unexpected response data: %+v", resp)
	}
}

func TestAuthDelivery_HandleLogin_ValidationErrors(t *testing.T) {
	delivery := NewAuthDelivery(&mockAuthUsecase{})

	tests := []struct {
		name string
		body string
	}{
		{"invalid json", `{invalid}`},
		{"empty password", `{"email":"test@example.com","password":""}`},
		{"short password", `{"email":"test@example.com","password":"123"}`},
		{"invalid email format", `{"email":"bad-email","password":"password123"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			delivery.HandleLogin(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("%s: expected status 400, got %d", tt.name, rec.Code)
			}
		})
	}
}

func TestAuthDelivery_HandleLogin_InvalidCredentials(t *testing.T) {
	mockUC := &mockAuthUsecase{loginErr: apperrors.ErrInvalidCredentials}
	delivery := NewAuthDelivery(mockUC)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"test@example.com","password":"wrongpassword"}`))
	rec := httptest.NewRecorder()

	delivery.HandleLogin(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthDelivery_HandleMe_Success(t *testing.T) {
	userID := uuid.New()
	mockUC := &mockAuthUsecase{
		meResult: &models.UserWithSession{
			User: &models.User{
				ID:    userID,
				Email: "test@example.com",
				Profile: &models.Profile{
					FirstName: "TestUser",
				},
			},
		},
	}
	delivery := NewAuthDelivery(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

	req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))

	rec := httptest.NewRecorder()

	delivery.HandleMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp UserResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Profile.FirstName != "TestUser" {
		t.Errorf("expected FirstName 'TestUser', got %s", resp.Profile.FirstName)
	}
}

func TestAuthDelivery_HandleMe_MissingAuth(t *testing.T) {
	delivery := NewAuthDelivery(&mockAuthUsecase{})
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

	rec := httptest.NewRecorder()

	delivery.HandleMe(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthDelivery_HandleLogout_Success(t *testing.T) {
	sessionID := uuid.New()
	mockUC := &mockAuthUsecase{}
	delivery := NewAuthDelivery(mockUC)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: utils.SessionCookieName, Value: sessionID.String()})
	rec := httptest.NewRecorder()

	delivery.HandleLogout(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].MaxAge != -1 {
		t.Errorf("expected cookie to be cleared (MaxAge=-1)")
	}
}

func TestAuthDelivery_HandleLogout_Errors(t *testing.T) {
	tests := []struct {
		name       string
		cookieVal  string
		mockErr    error
		wantStatus int
	}{
		{"no cookie", "", nil, http.StatusUnauthorized},
		{"invalid uuid", "not-a-uuid", nil, http.StatusUnauthorized},
		{"db error", uuid.New().String(), errors.New("db down"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockAuthUsecase{logoutErr: tt.mockErr}
			delivery := NewAuthDelivery(mockUC)

			req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
			if tt.cookieVal != "" {
				req.AddCookie(&http.Cookie{Name: utils.SessionCookieName, Value: tt.cookieVal})
			}
			rec := httptest.NewRecorder()

			delivery.HandleLogout(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.wantStatus, rec.Code)
			}
		})
	}
}
