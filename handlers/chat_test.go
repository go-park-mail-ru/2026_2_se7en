package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"app/apperrors"
	"app/middleware"
	"app/models"
)

type mockChatDB struct {
	chats    []*models.Chat
	err      error
	called   bool
	userID   uuid.UUID
	limit    int
	offset   int
	chatType string
}

type chatResponseItem struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type chatListResponse struct {
	Items []chatResponseItem `json:"items"`
}

func (m *mockChatDB) ListUserChats(_ context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.Chat, error) {
	m.called = true
	m.userID = userID
	m.limit = limit
	m.offset = offset
	m.chatType = chatType
	return m.chats, m.err
}

func performAuthenticatedChatRequest(target string, db *mockChatDB, userID uuid.UUID) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
	rec := httptest.NewRecorder()
	NewChatHandler(db).GetListUserChats(rec, req)
	return rec
}

func TestGetListUserChatsRequiresAuthentication(t *testing.T) {
	rec := httptest.NewRecorder()
	NewChatHandler(&mockChatDB{}).GetListUserChats(rec, httptest.NewRequest(http.MethodGet, "/chats", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetListUserChatsQueryValidation(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{name: "non-numeric limit", target: "?limit=nope"},
		{name: "zero limit", target: "?limit=0"},
		{name: "duplicate limit", target: "?limit=1&limit=2"},
		{name: "non-numeric offset", target: "?offset=nope"},
		{name: "negative offset", target: "?offset=-1"},
		{name: "duplicate offset", target: "?offset=1&offset=2"},
		{name: "empty type", target: "?type="},
		{name: "duplicate type", target: "?type=group&type=dialog"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mockChatDB{}
			rec := performAuthenticatedChatRequest("/chats"+tt.target, db, uuid.New())
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
			if db.called {
				t.Fatal("database called for invalid query")
			}
		})
	}
}

func TestGetListUserChatsSuccess(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name       string
		target     string
		chats      []*models.Chat
		wantItems  []chatResponseItem
		wantLimit  int
		wantOffset int
		wantType   string
	}{
		{name: "defaults to empty items", target: "/chats", wantItems: []chatResponseItem{}},
		{name: "passes query values", target: "/chats?limit=10&offset=3&type=group", chats: []*models.Chat{{Name: "team", Type: "group"}}, wantItems: []chatResponseItem{{Name: "team", Type: "group"}}, wantLimit: 10, wantOffset: 3, wantType: "group"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mockChatDB{chats: tt.chats}
			rec := performAuthenticatedChatRequest(tt.target, db, userID)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
			}
			if !db.called {
				t.Fatal("database was not called")
			}
			if db.userID != userID {
				t.Errorf("user id = %s, want %s", db.userID, userID)
			}
			if db.limit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", db.limit, tt.wantLimit)
			}
			if db.offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", db.offset, tt.wantOffset)
			}
			if db.chatType != tt.wantType {
				t.Errorf("chat type = %q, want %q", db.chatType, tt.wantType)
			}

			var response chatListResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if !reflect.DeepEqual(response.Items, tt.wantItems) {
				t.Errorf("response items = %+v, want %+v", response.Items, tt.wantItems)
			}
		})
	}
}

func TestGetListUserChatsDatabaseErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "session not found", err: apperrors.ErrSessionNotFound, wantStatus: http.StatusUnauthorized},
		{name: "invalid limit", err: apperrors.ErrInvalidChatLimit, wantStatus: http.StatusBadRequest},
		{name: "invalid offset", err: apperrors.ErrInvalidChatOffset, wantStatus: http.StatusBadRequest},
		{name: "invalid type", err: apperrors.ErrInvalidChatType, wantStatus: http.StatusBadRequest},
		{name: "unknown database error", err: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mockChatDB{err: tt.err}
			rec := performAuthenticatedChatRequest("/chats", db, uuid.New())
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
