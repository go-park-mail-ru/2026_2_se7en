package chat

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"app/models"
)

type mockChatRepo struct {
	items    []*models.ChatListItem
	err      error
	called   bool
	userID   uuid.UUID
	limit    int
	offset   int
	chatType string
}

func (m *mockChatRepo) ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error) {
	m.called = true
	m.userID = userID
	m.limit = limit
	m.offset = offset
	m.chatType = chatType
	return m.items, m.err
}
func (m *mockChatRepo) FindForUser(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error) {
	return nil, nil
}
func (m *mockChatRepo) GetOrCreateDialog(ctx context.Context, userID, peerID uuid.UUID) (*models.ChatDetail, error) {
	return nil, nil
}
func (m *mockChatRepo) HideDialog(ctx context.Context, userID, chatID uuid.UUID) error {
	return nil
}
func (m *mockChatRepo) CreateGroup(ctx context.Context, userID uuid.UUID, name string, description *string, members []uuid.UUID) (*models.ChatDetail, error) {
	return nil, nil
}
func (m *mockChatRepo) UpdateGroup(ctx context.Context, userID, chatID uuid.UUID, name, description *string, setDescription bool) (*models.ChatDetail, error) {
	return nil, nil
}
func (m *mockChatRepo) DeleteGroup(ctx context.Context, userID, chatID uuid.UUID) error {
	return nil
}
func (m *mockChatRepo) ListMembers(ctx context.Context, userID, chatID uuid.UUID) ([]models.ChatMember, error) {
	return nil, nil
}
func (m *mockChatRepo) AddMember(ctx context.Context, userID, chatID, newUserID uuid.UUID) error {
	return nil
}
func (m *mockChatRepo) RemoveMember(ctx context.Context, userID, chatID, targetID uuid.UUID) error {
	return nil
}
func (m *mockChatRepo) ListMessages(ctx context.Context, userID, chatID uuid.UUID, limit int, before *uuid.UUID) ([]models.Message, error) {
	return nil, nil
}

func TestChatUsecase_ListUserChats_Success(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()
	msgID := uuid.New()
	msgContent := "Hello world"

	mockRepo := &mockChatRepo{
		items: []*models.ChatListItem{
			{
				Chat: &models.Chat{
					ID:           chatID,
					Name:         "Test Chat",
					Type:         "dialog",
					MembersCount: 2,
				},
				LastMessage: &models.MessagePreview{
					ID:      msgID,
					Content: &msgContent,
					Type:    "text",
				},
			},
		},
	}
	uc := NewChatUsecase(mockRepo)

	items, err := uc.ListUserChats(context.Background(), userID, 10, 0, "dialog")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !mockRepo.called {
		t.Fatal("expected repository to be called")
	}
	if len(items) != 1 || items[0].Chat.Name != "Test Chat" {
		t.Errorf("unexpected items returned: %+v", items)
	}
	if items[0].LastMessage.Content == nil || *items[0].LastMessage.Content != "Hello world" {
		t.Errorf("unexpected message content")
	}
}

func TestChatUsecase_ListUserChats_NilToEmptySlice(t *testing.T) {
	mockRepo := &mockChatRepo{items: nil}
	uc := NewChatUsecase(mockRepo)

	items, err := uc.ListUserChats(context.Background(), uuid.New(), 10, 0, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if items == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(items) != 0 {
		t.Errorf("expected empty slice, got length %d", len(items))
	}
}

func TestChatUsecase_ListUserChats_Error(t *testing.T) {
	dbErr := errors.New("database connection failed")
	mockRepo := &mockChatRepo{err: dbErr}
	uc := NewChatUsecase(mockRepo)

	items, err := uc.ListUserChats(context.Background(), uuid.New(), 10, 0, "")

	if err == nil || err.Error() != "database connection failed" {
		t.Errorf("expected db error, got %v", err)
	}
	if items != nil {
		t.Errorf("expected nil items on error, got %+v", items)
	}
}
