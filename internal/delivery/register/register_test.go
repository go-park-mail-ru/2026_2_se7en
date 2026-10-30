package register

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

type mockRegisterUsecase struct {
	result *models.UserWithSession
	err    error
}

func (m *mockRegisterUsecase) Register(ctx context.Context, user *models.User, password string) (*models.UserWithSession, error) {
	return m.result, m.err
}

func TestRegisterDelivery_Success(t *testing.T) {
	mockUC := &mockRegisterUsecase{
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

	reqData := &RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		Nickname:  "testnick",
		FirstName: "Test",
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/register", nil)
	rec := httptest.NewRecorder()

	appErr := delivery.HandleRegister(rec, req, reqData)

	if appErr != nil {
		t.Fatalf("expected no error, got %v", appErr)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != middleware.SessionCookieName {
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

func TestRegisterDelivery_Validation(t *testing.T) {
	tests := []struct {
		name string
		req  *RegisterRequest
	}{
		{
			name: "short password",
			req:  &RegisterRequest{Email: "test@example.com", Password: "123", Nickname: "testnick", FirstName: "Test"},
		},
		{
			name: "invalid nickname",
			req:  &RegisterRequest{Email: "test@example.com", Password: "password123", Nickname: "ab", FirstName: "Test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			details := tt.req.IsValid()
			if len(details) == 0 {
				t.Errorf("expected validation errors, got none")
			}
		})
	}
}

func TestRegisterDelivery_EmailConflict(t *testing.T) {
	mockUC := &mockRegisterUsecase{err: apperrors.ErrEmailTaken}
	delivery := NewRegisterDelivery(mockUC)

	reqData := &RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		Nickname:  "testnick",
		FirstName: "Test",
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/register", nil)
	rec := httptest.NewRecorder()

	appErr := delivery.HandleRegister(rec, req, reqData)

	if appErr == nil || appErr.Status != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}
