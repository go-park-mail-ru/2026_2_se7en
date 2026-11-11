package users

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type onlineCheckerFunc func(uuid.UUID) bool

func (f onlineCheckerFunc) IsOnline(id uuid.UUID) bool { return f(id) }

func TestUsersDelivery_HandleGetUsersList_SuccessAndPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUC := NewMockUsersUsecase(ctrl)
	delivery := NewUsersDelivery(mockUC, onlineCheckerFunc(func(id uuid.UUID) bool {
		return id == uuid.MustParse("00000000-0000-0000-0000-000000000002")
	}), slog.Default())
	requesterID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	iconID := uuid.New()
	avatar := "/uploads/avatars/ada.webp"
	mockUC.EXPECT().GetUsersList(gomock.Any(), requesterID, 3, 4, "ada").Return([]models.UserSearchResult{
		{Profile: models.Profile{UserID: userID, Nickname: "ada_l", FirstName: "Ada", IconID: &iconID, AvatarURL: &avatar}, IsContact: true},
		{Profile: models.Profile{UserID: uuid.New(), Nickname: "ada_m", FirstName: "Ada"}},
		{Profile: models.Profile{UserID: uuid.New(), Nickname: "ada_n", FirstName: "Ada"}},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/users?q=%20ada%20&limit=2&offset=4", nil)
	req = req.WithContext(middleware.NewAuthContext(req.Context(), requesterID))
	rec := httptest.NewRecorder()
	delivery.HandleGetUsersList(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var response UsersListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Items) != 2 || response.NextOffset == nil || *response.NextOffset != 6 {
		t.Fatalf("unexpected page: %+v", response)
	}
	if response.Items[0].UserID != userID || !response.Items[0].IsOnline || !response.Items[0].IsContact || response.Items[0].IconURL == nil {
		t.Fatalf("unexpected first result: %+v", response.Items[0])
	}
}

func TestUsersDelivery_HandleGetUsersList_EmptyQuerySkipsRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	delivery := NewUsersDelivery(NewMockUsersUsecase(ctrl), nil, slog.Default())
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/users?q=%20%20", nil)
	req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
	rec := httptest.NewRecorder()
	delivery.HandleGetUsersList(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "{\"items\":[]}\n" {
		t.Fatalf("status/body = %d/%s", rec.Code, rec.Body.String())
	}
}

func TestUsersDelivery_HandleGetUsersList_ValidatesPagination(t *testing.T) {
	for _, path := range []string{"/users?q=ada&limit=0", "/users?q=ada&limit=nope", "/users?q=ada&offset=-1"} {
		t.Run(path, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			delivery := NewUsersDelivery(NewMockUsersUsecase(ctrl), nil, slog.Default())
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req = req.WithContext(middleware.NewAuthContext(req.Context(), uuid.New()))
			rec := httptest.NewRecorder()
			delivery.HandleGetUsersList(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestUsersDelivery_HandleGetUsersList_UsecaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUC := NewMockUsersUsecase(ctrl)
	delivery := NewUsersDelivery(mockUC, nil, slog.Default())
	userID := uuid.New()
	mockUC.EXPECT().GetUsersList(gomock.Any(), userID, 21, 0, "ada").Return(nil, errors.New("db down"))
	req := httptest.NewRequest(http.MethodGet, "/users?q=ada", nil)
	req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
	rec := httptest.NewRecorder()
	delivery.HandleGetUsersList(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", rec.Code, rec.Body.String())
	}
}
