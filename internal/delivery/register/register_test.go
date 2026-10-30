package register

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"app/internal/apperrors"
	"app/internal/models"
	"app/internal/utils"
)

type mockRegisterUsecase struct {
	result *models.UserWithSession
	err    error
}

func (m *mockRegisterUsecase) HandleRegister(ctx context.Context, user interface{}, password string) (*models.UserWithSession, error) {

	return m.result, m.err
}

type mockUsecase struct {
	result *models.UserWithSession
	err    error
}

func (m *mockUsecase) Register(ctx context.Context, user *models.User, password string) (*models.UserWithSession, error) {
	return m.result, m.err
}

func TestRegisterDelivery_Success(t *testing.T) {
	mockUC := &mockUsecase{
		result: &models.UserWithSession{
			User: &models.User{
				ID:    uuid.New(),
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

	delivery := NewRegisterDelivery(mockUC)

	body := `{"email":"test@example.com","password":"password123","nickname":"testnick","first_name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/HandleRegister", strings.NewReader(body))
	rec := httptest.NewRecorder()

	delivery.HandleRegister(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != utils.SessionCookieName {
		t.Errorf("expected session cookie to be set")
	}

	var resp RegisterResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Email != "test@example.com" {
		t.Errorf("expected email in response, got %s", resp.Email)
	}
}

func TestRegisterDelivery_InvalidJSON(t *testing.T) {
	delivery := NewRegisterDelivery(&mockUsecase{})

	req := httptest.NewRequest(http.MethodPost, "/auth/HandleRegister", strings.NewReader(`{invalid json}`))
	rec := httptest.NewRecorder()

	delivery.HandleRegister(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestRegisterDelivery_ValidationErrors(t *testing.T) {
	delivery := NewRegisterDelivery(&mockUsecase{})

	body := `{"email":"test@example.com","password":"123","nickname":"testnick","first_name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/HandleRegister", strings.NewReader(body))
	rec := httptest.NewRecorder()

	delivery.HandleRegister(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for validation error, got %d", rec.Code)
	}
}

func TestRegisterDelivery_EmailConflict(t *testing.T) {
	mockUC := &mockUsecase{err: apperrors.ErrEmailTaken}
	delivery := NewRegisterDelivery(mockUC)

	body := `{"email":"test@example.com","password":"password123","nickname":"testnick","first_name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/HandleRegister", strings.NewReader(body))
	rec := httptest.NewRecorder()

	delivery.HandleRegister(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}
