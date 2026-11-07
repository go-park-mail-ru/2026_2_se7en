package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
)

type mockChatUsecase struct {
	items        []*models.ChatListItem
	detail       *models.ChatDetail
	dialogStatus models.DialogStatus
	members      []models.ChatMember
	member       *models.ChatMember
	messages     []models.Message
	isMember     bool
	readChanged  bool
	err          error
	called       bool
	userID       uuid.UUID
	chatID       uuid.UUID
	targetID     uuid.UUID
}

type mockBroadcaster struct {
	chatID       uuid.UUID
	userID       uuid.UUID
	userIDs      []uuid.UUID
	event        models.Event
	chatCalled   bool
	userCalled   bool
	revokeCalled bool
	removeCalled bool
}

func (m *mockBroadcaster) BroadcastToChat(chatID uuid.UUID, event models.Event) {
	m.chatCalled = true
	m.chatID = chatID
	m.event = event
}

func (m *mockBroadcaster) BroadcastToUser(userID uuid.UUID, event models.Event) {
	m.userCalled = true
	m.userID = userID
	m.userIDs = append(m.userIDs, userID)
	m.event = event
}

func (m *mockBroadcaster) RevokeUserFromChat(userID, chatID uuid.UUID) {
	m.revokeCalled = true
	m.userID = userID
	m.chatID = chatID
}

func (m *mockBroadcaster) RemoveChat(chatID uuid.UUID) {
	m.removeCalled = true
	m.chatID = chatID
}

var _ Broadcaster = (*mockBroadcaster)(nil)

func TestChatDelivery_CreateGroup(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()
	uc := &mockChatUsecase{detail: &models.ChatDetail{Chat: &models.Chat{ID: chatID, Name: "team", Type: "group"}, Role: "creator"}}
	req := httptest.NewRequest(http.MethodPost, "/chats/groups", strings.NewReader(`{"name":"team","members":[]}`))
	req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
	rec := httptest.NewRecorder()
	tools.NewJSONDecoder(NewChatDelivery(uc, &mockBroadcaster{}).HandleCreateGroup).ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated || !uc.called {
		t.Fatalf("status = %d, called = %t, body = %s", rec.Code, uc.called, rec.Body.String())
	}
	var response ChatDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || response.Chat == nil || response.Chat.ID != chatID {
		t.Fatalf("unexpected response: %+v, err: %v", response, err)
	}
}

func TestChatDelivery_DeleteChatBroadcastsAndRemovesChat(t *testing.T) {
	userID, chatID := uuid.New(), uuid.New()
	uc := &mockChatUsecase{}
	broadcaster := &mockBroadcaster{}
	req := httptest.NewRequest(http.MethodDelete, "/chats/"+chatID.String(), nil)
	req.SetPathValue("id", chatID.String())
	req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
	rec := httptest.NewRecorder()

	NewChatDelivery(uc, broadcaster).HandleDeleteChat(rec, req)
	if rec.Code != http.StatusNoContent || !uc.called || uc.userID != userID || uc.chatID != chatID {
		t.Fatalf("status = %d, usecase called = %t, user = %s, chat = %s", rec.Code, uc.called, uc.userID, uc.chatID)
	}
	if !broadcaster.chatCalled || !broadcaster.removeCalled || broadcaster.chatID != chatID || broadcaster.event.Type != models.EventChatDeleted {
		t.Fatalf("chat deletion was not broadcast and removed: %+v", broadcaster)
	}
}

func TestChatDelivery_LeaveGroupRemovesCurrentUser(t *testing.T) {
	userID, chatID := uuid.New(), uuid.New()
	uc := &mockChatUsecase{}
	broadcaster := &mockBroadcaster{}
	req := httptest.NewRequest(http.MethodDelete, "/chats/"+chatID.String()+"/leave", nil)
	req.SetPathValue("id", chatID.String())
	req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
	rec := httptest.NewRecorder()

	NewChatDelivery(uc, broadcaster).HandleLeaveGroup(rec, req)
	if rec.Code != http.StatusNoContent || !uc.called || uc.userID != userID || uc.chatID != chatID || uc.targetID != userID {
		t.Fatalf("status = %d, usecase called = %t, actor = %s, chat = %s, target = %s", rec.Code, uc.called, uc.userID, uc.chatID, uc.targetID)
	}
	if !broadcaster.revokeCalled || !broadcaster.chatCalled || !broadcaster.userCalled || broadcaster.event.Type != models.EventMemberRemoved {
		t.Fatalf("member removal was not broadcast: %+v", broadcaster)
	}
}

func TestChatDelivery_GetOrCreateDialogBroadcastsOnlyWhenVisibleForFirstTime(t *testing.T) {
	userID, peerID, chatID := uuid.New(), uuid.New(), uuid.New()
	tests := []struct {
		name    string
		status  models.DialogStatus
		userIDs []uuid.UUID
	}{
		{"created", models.DialogCreated, []uuid.UUID{userID, peerID}},
		{"restored", models.DialogRestored, []uuid.UUID{userID}},
		{"existing", models.DialogExisting, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := &mockChatUsecase{
				detail:       &models.ChatDetail{Chat: &models.Chat{ID: chatID, Type: "dialog"}},
				dialogStatus: tc.status,
			}
			broadcaster := &mockBroadcaster{}
			req := httptest.NewRequest(http.MethodPost, "/chats/dialogs", strings.NewReader(fmt.Sprintf(`{"peer_id":%q}`, peerID)))
			req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
			rec := httptest.NewRecorder()
			tools.NewJSONDecoder(NewChatDelivery(uc, broadcaster).HandleGetOrCreateDialog).ServeHTTP(rec, req)
			if rec.Code != http.StatusOK || !uc.called {
				t.Fatalf("status = %d, called = %t, body = %s", rec.Code, uc.called, rec.Body.String())
			}
			if len(broadcaster.userIDs) != len(tc.userIDs) {
				t.Fatalf("broadcasts to %v, want %v", broadcaster.userIDs, tc.userIDs)
			}
			for i, id := range tc.userIDs {
				if broadcaster.userIDs[i] != id {
					t.Fatalf("broadcasts to %v, want %v", broadcaster.userIDs, tc.userIDs)
				}
			}
			if len(tc.userIDs) > 0 {
				if broadcaster.event.Type != models.EventChatCreated || broadcaster.event.Payload.(map[string]any)["chat_id"] != chatID {
					t.Fatalf("unexpected event: %+v", broadcaster.event)
				}
			}
		})
	}
}

func TestChatDelivery_AddMemberRejectsZeroID(t *testing.T) {
	uc := &mockChatUsecase{}
	req := httptest.NewRequest(http.MethodPost, "/chats/"+uuid.NewString()+"/members", strings.NewReader(`{"user_id":"00000000-0000-0000-0000-000000000000"}`))
	req.SetPathValue("id", uuid.NewString())
	req = req.WithContext(middleware.NewAuthContext(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()
	tools.NewJSONDecoder(NewChatDelivery(uc, &mockBroadcaster{}).HandleAddMember).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || uc.called {
		t.Fatalf("status = %d, called = %t, body = %s", rec.Code, uc.called, rec.Body.String())
	}
}

func TestChatDelivery_ListMessagesRejectsInvalidCursor(t *testing.T) {
	uc := &mockChatUsecase{}
	req := httptest.NewRequest(http.MethodGet, "/chats/"+uuid.NewString()+"/messages?before=bad", nil)
	req.SetPathValue("id", uuid.NewString())
	req = req.WithContext(middleware.NewAuthContext(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()
	NewChatDelivery(uc, &mockBroadcaster{}).HandleListMessages(rec, req)
	if rec.Code != http.StatusBadRequest || uc.called {
		t.Fatalf("status = %d, called = %t, body = %s", rec.Code, uc.called, rec.Body.String())
	}
}

func TestChatDelivery_ListMessagesIncludesReadState(t *testing.T) {
	userID, chatID, messageID := uuid.New(), uuid.New(), uuid.New()
	uc := &mockChatUsecase{messages: []models.Message{{
		ID:     messageID,
		ChatID: chatID,
		IsRead: true,
	}}}
	req := httptest.NewRequest(http.MethodGet, "/chats/"+chatID.String()+"/messages", nil)
	req.SetPathValue("id", chatID.String())
	req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
	rec := httptest.NewRecorder()

	NewChatDelivery(uc, &mockBroadcaster{}).HandleListMessages(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var response ListMessagesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Items) != 1 || response.Items[0].ID != messageID || !response.Items[0].IsRead {
		t.Fatalf("unexpected messages: %+v", response.Items)
	}
}

func TestChatDelivery_BodyValidation(t *testing.T) {
	userID := uuid.New()
	memberID := uuid.New()
	pathID := uuid.NewString()
	stickerID := uuid.New()
	tests := []struct {
		name    string
		body    string
		handler func(*ChatDeliveryImpl) http.Handler
	}{
		{"dialog missing peer", `{}`, func(d *ChatDeliveryImpl) http.Handler { return tools.NewJSONDecoder(d.HandleGetOrCreateDialog) }},
		{"dialog self", fmt.Sprintf(`{"peer_id":%q}`, userID), func(d *ChatDeliveryImpl) http.Handler { return tools.NewJSONDecoder(d.HandleGetOrCreateDialog) }},
		{"group empty name", `{ "name": " " }`, func(d *ChatDeliveryImpl) http.Handler { return tools.NewJSONDecoder(d.HandleCreateGroup) }},
		{"group long name", fmt.Sprintf(`{"name":%q}`, strings.Repeat("я", 101)), func(d *ChatDeliveryImpl) http.Handler { return tools.NewJSONDecoder(d.HandleCreateGroup) }},
		{"group long description", fmt.Sprintf(`{"name":"Group","description":%q}`, strings.Repeat("я", 501)), func(d *ChatDeliveryImpl) http.Handler { return tools.NewJSONDecoder(d.HandleCreateGroup) }},
		{"group duplicate members", fmt.Sprintf(`{"name":"Group","members":[%q,%q]}`, memberID, memberID), func(d *ChatDeliveryImpl) http.Handler { return tools.NewJSONDecoder(d.HandleCreateGroup) }},
		{"group creator member", fmt.Sprintf(`{"name":"Group","members":[%q]}`, userID), func(d *ChatDeliveryImpl) http.Handler { return tools.NewJSONDecoder(d.HandleCreateGroup) }},
		{"update no changes", `{}`, func(d *ChatDeliveryImpl) http.Handler { return tools.NewJSONDecoder(d.HandleUpdateGroup) }},
		{"update description without flag", `{"name":"Group","description":"new"}`, func(d *ChatDeliveryImpl) http.Handler { return tools.NewJSONDecoder(d.HandleUpdateGroup) }},
		{"member missing id", `{}`, func(d *ChatDeliveryImpl) http.Handler { return tools.NewJSONDecoder(d.HandleAddMember) }},
		{"member self", fmt.Sprintf(`{"user_id":%q}`, userID), func(d *ChatDeliveryImpl) http.Handler { return tools.NewJSONDecoder(d.HandleAddMember) }},
		{"message empty content", `{"type":"text","content":""}`, func(d *ChatDeliveryImpl) http.Handler { return tools.NewJSONDecoder(d.HandleCreateMessage) }},
		{"message both fields", fmt.Sprintf(`{"type":"text","content":"hi","sticker_id":%q}`, stickerID), func(d *ChatDeliveryImpl) http.Handler { return tools.NewJSONDecoder(d.HandleCreateMessage) }},
		{"message long content", fmt.Sprintf(`{"type":"text","content":%q}`, strings.Repeat("я", 4097)), func(d *ChatDeliveryImpl) http.Handler { return tools.NewJSONDecoder(d.HandleCreateMessage) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := &mockChatUsecase{}
			req := httptest.NewRequest(http.MethodPost, "/chats/"+pathID, strings.NewReader(tc.body))
			req.SetPathValue("id", pathID)
			req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
			rec := httptest.NewRecorder()
			tc.handler(NewChatDelivery(uc, &mockBroadcaster{})).ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest || uc.called {
				t.Fatalf("status = %d, called = %t, body = %s", rec.Code, uc.called, rec.Body.String())
			}
		})
	}
}

func TestChatDelivery_ListMessagesLimitValidation(t *testing.T) {
	for _, limit := range []string{"0", "51", "abc", "1&limit=2"} {
		t.Run(limit, func(t *testing.T) {
			uc := &mockChatUsecase{}
			req := httptest.NewRequest(http.MethodGet, "/chats/"+uuid.NewString()+"/messages?limit="+limit, nil)
			req.SetPathValue("id", uuid.NewString())
			req = req.WithContext(middleware.NewAuthContext(req.Context(), uuid.New()))
			rec := httptest.NewRecorder()
			NewChatDelivery(uc, &mockBroadcaster{}).HandleListMessages(rec, req)
			if rec.Code != http.StatusBadRequest || uc.called {
				t.Fatalf("status = %d, called = %t", rec.Code, uc.called)
			}
		})
	}
}

func TestMapChatError(t *testing.T) {
	cases := []struct {
		err    error
		status int
	}{
		{apperrors.ErrChatNotFound, http.StatusNotFound},
		{apperrors.ErrUserNotFound, http.StatusNotFound},
		{apperrors.ErrChatPermissionDenied, http.StatusForbidden},
		{apperrors.ErrChatMemberExists, http.StatusConflict},
		{apperrors.ErrInvalidChatOperation, http.StatusBadRequest},
		{errors.New("database failed"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		if got := mapChatError(tc.err).Status; got != tc.status {
			t.Errorf("error %v: status = %d, want %d", tc.err, got, tc.status)
		}
	}
}

func (m *mockChatUsecase) IsMember(ctx context.Context, userID, chatID uuid.UUID) (bool, error) {
	m.called = true
	return m.isMember, m.err
}

func (m *mockChatUsecase) ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error) {
	m.called = true
	return m.items, m.err
}

func (m *mockChatUsecase) GetChatDetail(context.Context, uuid.UUID, uuid.UUID) (*models.ChatDetail, error) {
	m.called = true
	return m.detail, m.err
}
func (m *mockChatUsecase) GetOrCreateDialog(context.Context, uuid.UUID, uuid.UUID) (*models.ChatDetail, models.DialogStatus, error) {
	m.called = true
	return m.detail, m.dialogStatus, m.err
}
func (m *mockChatUsecase) HideDialog(context.Context, uuid.UUID, uuid.UUID) error {
	m.called = true
	return m.err
}
func (m *mockChatUsecase) CreateGroup(context.Context, uuid.UUID, string, *string, []uuid.UUID) (*models.ChatDetail, error) {
	m.called = true
	return m.detail, m.err
}
func (m *mockChatUsecase) UpdateGroup(context.Context, uuid.UUID, uuid.UUID, *string, *string, bool) (*models.ChatDetail, error) {
	m.called = true
	return m.detail, m.err
}
func (m *mockChatUsecase) DeleteChat(_ context.Context, userID, chatID uuid.UUID) error {
	m.called = true
	m.userID = userID
	m.chatID = chatID
	return m.err
}
func (m *mockChatUsecase) ListMembers(context.Context, uuid.UUID, uuid.UUID) ([]models.ChatMember, error) {
	m.called = true
	return m.members, m.err
}
func (m *mockChatUsecase) AddMember(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (*models.ChatMember, error) {
	m.called = true
	return m.member, m.err
}
func (m *mockChatUsecase) RemoveMember(_ context.Context, userID, chatID, targetID uuid.UUID) error {
	m.called = true
	m.userID = userID
	m.chatID = chatID
	m.targetID = targetID
	return m.err
}
func (m *mockChatUsecase) ListMessages(context.Context, uuid.UUID, uuid.UUID, int, *uuid.UUID) ([]models.Message, error) {
	m.called = true
	return m.messages, m.err
}
func (m *mockChatUsecase) CreateMessage(ctx context.Context, userID, chatID uuid.UUID, replyToMessageID, stickerID *uuid.UUID, content *string, messageType string) (*models.Message, error) {
	m.called = true
	if len(m.messages) > 0 {
		return &m.messages[0], m.err
	} else {
		return nil, m.err
	}
}

func (m *mockChatUsecase) MarkMessageRead(ctx context.Context, userID, chatID, messageID uuid.UUID) (bool, error) {
	m.called = true
	return m.readChanged, m.err
}

var _ ChatUsecase = (*mockChatUsecase)(nil)

func performAuthenticatedChatRequest(target string, uc *mockChatUsecase, userID uuid.UUID) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req = req.WithContext(middleware.NewAuthContext(req.Context(), userID))
	rec := httptest.NewRecorder()

	NewChatDelivery(uc, &mockBroadcaster{}).HandleListUserChats(rec, req)
	return rec
}

func TestChatDelivery_HandleListUserChats_RequiresAuth(t *testing.T) {
	uc := &mockChatUsecase{}
	req := httptest.NewRequest(http.MethodGet, "/chats", nil)
	rec := httptest.NewRecorder()

	NewChatDelivery(uc, &mockBroadcaster{}).HandleListUserChats(rec, req)

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
		{name: "zero limit", target: "/chats?limit=0"},
		{name: "excessive limit", target: "/chats?limit=51"},
		{name: "negative offset", target: "/chats?offset=-1"},
		{name: "invalid type", target: "/chats?type=invalid"},
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
