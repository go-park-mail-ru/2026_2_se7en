package contact

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestContactDelivery_HandleGetContactsList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := NewMockContactUsecase(ctrl)
	delivery := NewContactDelivery(mockUC)

	userID := uuid.New()
	expectedContacts := []models.Contact{
		{
			UserID:        userID,
			ContactUserID: uuid.New(),
			Profile:       &models.Profile{Nickname: "friend", FirstName: "John"},
			IsOnline:      true,
		},
	}

	mockUC.EXPECT().
		GetContactsList(gomock.Any(), userID).
		Return(expectedContacts, nil)

	req := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
	rec := httptest.NewRecorder()

	delivery.HandleGetContactsList(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp []ContactResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp) != 1 || resp[0].Nickname != "friend" {
		t.Errorf("unexpected response data: %+v", resp)
	}
}

func TestContactDelivery_HandleCreateContact_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := NewMockContactUsecase(ctrl)
	delivery := NewContactDelivery(mockUC)

	userID := uuid.New()

	pseudonym := "buddy"
	reqData := &CreateContactRequest{Nickname: "friend", Pseudonym: &pseudonym}

	mockUC.EXPECT().
		CreateContact(gomock.Any(), userID, "friend", &pseudonym).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/contacts", strings.NewReader(`{"nickname":"friend","pseudonym":"buddy"}`))
	req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
	rec := httptest.NewRecorder()

	appErr := delivery.HandleCreateContact(rec, req, reqData)

	if appErr != nil {
		t.Fatalf("expected no error, got %v", appErr)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestContactDelivery_HandleCreateContact_Errors(t *testing.T) {
	tests := []struct {
		name       string
		mockErr    error
		wantStatus int
	}{
		{"user not found", apperrors.ErrUserNotFound, http.StatusNotFound},
		{"already exists", apperrors.ErrContactAlreadyExists, http.StatusConflict},
		{"self add", apperrors.ErrCannotAddSelfAsContact, http.StatusBadRequest},
		{"internal error", errors.New("db down"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := NewMockContactUsecase(ctrl)
			delivery := NewContactDelivery(mockUC)

			userID := uuid.New()
			reqData := &CreateContactRequest{Nickname: "friend"}

			mockUC.EXPECT().
				CreateContact(gomock.Any(), userID, "friend", gomock.Any()).
				Return(tt.mockErr)

			req := httptest.NewRequest(http.MethodPost, "/contacts", strings.NewReader(`{"nickname":"friend"}`))
			req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
			rec := httptest.NewRecorder()

			appErr := delivery.HandleCreateContact(rec, req, reqData)

			if appErr == nil || appErr.Status != tt.wantStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestContactDelivery_HandleDeleteContact_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := NewMockContactUsecase(ctrl)
	delivery := NewContactDelivery(mockUC)

	userID := uuid.New()
	contactUserID := uuid.New()
	reqData := &DeleteContactRequest{ContactUserID: contactUserID}

	mockUC.EXPECT().
		DeleteContact(gomock.Any(), userID, contactUserID).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/contacts", strings.NewReader(`{"contact_user_id":"`+contactUserID.String()+`"}`))
	req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
	rec := httptest.NewRecorder()

	appErr := delivery.HandleDeleteContact(rec, req, reqData)

	if appErr != nil {
		t.Fatalf("expected no error, got %v", appErr)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}

func TestContactDelivery_HandleDeleteContact_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := NewMockContactUsecase(ctrl)
	delivery := NewContactDelivery(mockUC)

	userID := uuid.New()
	contactUserID := uuid.New()
	reqData := &DeleteContactRequest{ContactUserID: contactUserID}

	mockUC.EXPECT().
		DeleteContact(gomock.Any(), userID, contactUserID).
		Return(apperrors.ErrContactNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/contacts", strings.NewReader(`{"contact_user_id":"`+contactUserID.String()+`"}`))
	req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
	rec := httptest.NewRecorder()

	appErr := delivery.HandleDeleteContact(rec, req, reqData)

	if appErr == nil || appErr.Status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}
