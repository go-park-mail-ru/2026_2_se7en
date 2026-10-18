package chat

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"app/apperrors"
	"app/middleware"
	"app/models"
)

type mockChatUsecase struct {
	items  []*models.ChatListItem
	err    error
	called bool
}

func (m *mockChatUsecase) ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error) {
	m.called = true
	return m.items, m.err
}

func performAuthenticatedChatRequest(target string, uc *mockChatUsecase, userID uuid.UUID) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
	rec := httptest.NewRecorder()

	NewChatDelivery(uc).HandleListUserChats(rec, req)
	return rec
}

func TestChatDelivery_HandleListUserChats_RequiresAuth(t *testing.T) {
	uc := &mockChatUsecase{}
	req := httptest.NewRequest(http.MethodGet, "/chats", nil)
	rec := httptest.NewRecorder()

	NewChatDelivery(uc).HandleListUserChats(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if uc.called {
		t.Fatal("usecase should not be called without auth")
	}
}

func TestChatDelivery_HandleListUserChats_QueryValidation(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{name: "non-numeric limit", target: "/chats?limit=nope"},
		{name: "duplicate limit", target: "/chats?limit=1&limit=2"},
		{name: "non-numeric offset", target: "/chats?offset=nope"},
		{name: "duplicate offset", target: "/chats?offset=1&offset=2"},
		{name: "empty type", target: "/chats?type="},
		{name: "duplicate type", target: "/chats?type=group&type=dialog"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockChatUsecase{}
			rec := performAuthenticatedChatRequest(tt.target, uc, uuid.New())

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
			if uc.called {
				t.Fatal("usecase called for invalid query format")
			}
		})
	}
}

func TestChatDelivery_HandleListUserChats_Success(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()
	msgID := uuid.New()
	msgContent := "Hello"

	uc := &mockChatUsecase{
		items: []*models.ChatListItem{
			{
				Chat: &models.Chat{
					ID:           chatID,
					Name:         "team",
					Type:         "group",
					MembersCount: 5,
				},
				LastMessage: &models.MessagePreview{
					ID:      msgID,
					Content: &msgContent,
					Type:    "text",
				},
			},
		},
	}

	rec := performAuthenticatedChatRequest("/chats?limit=10&offset=3&type=group", uc, userID)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !uc.called {
		t.Fatal("usecase was not called")
	}

	var response ListChatsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(response.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(response.Items))
	}
	if response.Items[0].Name != "team" || response.Items[0].Type != "group" {
		t.Errorf("unexpected item data: %+v", response.Items[0])
	}
	if response.Items[0].LastMessage == nil || *response.Items[0].LastMessage.Content != "Hello" {
		t.Errorf("unexpected last message data: %+v", response.Items[0].LastMessage)
	}
}

func TestChatDelivery_HandleListUserChats_DatabaseErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "invalid limit", err: apperrors.ErrInvalidChatLimit, wantStatus: http.StatusBadRequest},
		{name: "invalid offset", err: apperrors.ErrInvalidChatOffset, wantStatus: http.StatusBadRequest},
		{name: "invalid type", err: apperrors.ErrInvalidChatType, wantStatus: http.StatusBadRequest},
		{name: "unknown database error", err: errors.New("db down"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockChatUsecase{err: tt.err}
			rec := performAuthenticatedChatRequest("/chats", uc, uuid.New())

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
