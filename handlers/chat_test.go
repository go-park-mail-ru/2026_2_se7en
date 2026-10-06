package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

func (m *mockChatDB) ListUserChats(_ context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.Chat, error) {
	m.called = true
	m.userID = userID
	m.limit = limit
	m.offset = offset
	m.chatType = chatType
	return m.chats, m.err
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
			req := httptest.NewRequest(http.MethodGet, "/chats"+tt.target, nil)
			req = req.WithContext(middleware.NewAuthContext(req.Context(), uuid.New()))
			rec := httptest.NewRecorder()
			NewChatHandler(db).GetListUserChats(rec, req)
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
		wantLimit  int
		wantOffset int
		wantType   string
		wantBody   string
	}{
		{name: "defaults to empty items", target: "/chats", wantBody: `{"items":[]}`},
		{name: "passes query values", target: "/chats?limit=10&offset=3&type=group", chats: []*models.Chat{{Name: "team"}}, wantLimit: 10, wantOffset: 3, wantType: "group", wantBody: `"name":"team"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mockChatDB{chats: tt.chats}
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
			rec := httptest.NewRecorder()
			NewChatHandler(db).GetListUserChats(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
			}
			if !db.called || db.userID != userID || db.limit != tt.wantLimit || db.offset != tt.wantOffset || db.chatType != tt.wantType {
				t.Fatalf("database arguments = %+v", db)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Fatalf("response %q does not contain %q", rec.Body.String(), tt.wantBody)
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
			req := httptest.NewRequest(http.MethodGet, "/chats", nil)
			req = req.WithContext(middleware.NewAuthContext(req.Context(), uuid.New()))
			rec := httptest.NewRecorder()
			NewChatHandler(db).GetListUserChats(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
