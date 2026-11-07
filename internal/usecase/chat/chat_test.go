package chat

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
)

type mockChatRepo struct {
	items         []*models.ChatListItem
	detail        *models.ChatDetail
	dialogStatus  models.DialogStatus
	members       []models.ChatMember
	member        *models.ChatMember
	messages      []models.Message
	isMember      bool
	readChanged   bool
	err           error
	called        bool
	userID        uuid.UUID
	limit         int
	offset        int
	chatType      string
	readChatID    uuid.UUID
	readMessageID uuid.UUID
}

func (m *mockChatRepo) IsMember(ctx context.Context, userID, chatID uuid.UUID) (bool, error) {
	m.called = true
	return m.isMember, m.err
}

func (m *mockChatRepo) ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error) {
	m.called = true
	m.userID = userID
	m.limit = limit
	m.offset = offset
	m.chatType = chatType
	return m.items, m.err
}
func (m *mockChatRepo) GetChatDetail(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error) {
	m.called = true
	return m.detail, m.err
}
func (m *mockChatRepo) GetOrCreateDialog(ctx context.Context, userID, peerID uuid.UUID) (*models.ChatDetail, models.DialogStatus, error) {
	m.called = true
	return m.detail, m.dialogStatus, m.err
}
func (m *mockChatRepo) HideDialog(ctx context.Context, userID, chatID uuid.UUID) error {
	m.called = true
	return m.err
}
func (m *mockChatRepo) CreateGroup(ctx context.Context, userID uuid.UUID, name string, description *string, members []uuid.UUID) (*models.ChatDetail, error) {
	m.called = true
	return m.detail, m.err
}
func (m *mockChatRepo) UpdateGroup(ctx context.Context, userID, chatID uuid.UUID, name, description *string, setDescription bool) (*models.ChatDetail, error) {
	m.called = true
	return m.detail, m.err
}
func (m *mockChatRepo) DeleteGroup(ctx context.Context, userID, chatID uuid.UUID) error {
	m.called = true
	return m.err
}
func (m *mockChatRepo) ListMembers(ctx context.Context, userID, chatID uuid.UUID) ([]models.ChatMember, error) {
	m.called = true
	return m.members, m.err
}
func (m *mockChatRepo) AddMember(ctx context.Context, userID, chatID, newUserID uuid.UUID) (*models.ChatMember, error) {
	m.called = true
	return m.member, m.err
}
func (m *mockChatRepo) RemoveMember(ctx context.Context, userID, chatID, targetID uuid.UUID) error {
	m.called = true
	return m.err
}
func (m *mockChatRepo) ListMessages(ctx context.Context, userID, chatID uuid.UUID, limit int, before *uuid.UUID) ([]models.Message, error) {
	m.called = true
	return m.messages, m.err
}
func (m *mockChatRepo) CreateMessage(ctx context.Context, userID, chatID uuid.UUID, replyToMessageID, stickerID *uuid.UUID, content *string, messageType string) (*models.Message, error) {
	m.called = true
	if len(m.messages) > 0 {
		return &m.messages[0], m.err
	} else {
		return nil, m.err
	}
}

func (m *mockChatRepo) MarkMessageRead(ctx context.Context, userID, chatID, messageID uuid.UUID) (bool, error) {
	m.called = true
	m.userID = userID
	m.readChatID = chatID
	m.readMessageID = messageID
	return m.readChanged, m.err
}

func newTestChatUsecase(repo ChatRepository) *ChatUsecaseImpl {
	return NewChatUsecase(repo)
}

func TestChatUsecase_MarkMessageReadReturnsTransition(t *testing.T) {
	userID, chatID, messageID := uuid.New(), uuid.New(), uuid.New()
	repo := &mockChatRepo{readChanged: true}
	changed, err := newTestChatUsecase(repo).MarkMessageRead(context.Background(), userID, chatID, messageID)
	if err != nil || !changed || !repo.called {
		t.Fatalf("changed = %t, err = %v, called = %t", changed, err, repo.called)
	}
	if repo.userID != userID || repo.readChatID != chatID || repo.readMessageID != messageID {
		t.Fatalf("repository received wrong IDs: user=%s chat=%s message=%s", repo.userID, repo.readChatID, repo.readMessageID)
	}
}

func TestChatUsecase_ValidGroupAndMessagesReachRepository(t *testing.T) {
	userID, chatID, peerID := uuid.New(), uuid.New(), uuid.New()
	name := strings.Repeat("я", 100)
	description := strings.Repeat("я", 500)
	repo := &mockChatRepo{}
	uc := newTestChatUsecase(repo)

	if _, err := uc.CreateGroup(context.Background(), userID, name, &description, []uuid.UUID{peerID}); err != nil || !repo.called {
		t.Fatalf("valid group was not forwarded: err=%v, called=%v", err, repo.called)
	}
	repo.called = false
	if _, err := uc.UpdateGroup(context.Background(), userID, chatID, nil, nil, true); err != nil || !repo.called {
		t.Fatalf("clearing description was not forwarded: err=%v, called=%v", err, repo.called)
	}
	repo.called = false
	if _, err := uc.ListMessages(context.Background(), userID, chatID, 50, &peerID); err != nil || !repo.called {
		t.Fatalf("valid message page was not forwarded: err=%v, called=%v", err, repo.called)
	}
}

func TestChatUsecase_AddMemberReturnsCreatedMember(t *testing.T) {
	userID, chatID, newUserID := uuid.New(), uuid.New(), uuid.New()
	want := &models.ChatMember{UserID: newUserID, Role: "member"}
	repo := &mockChatRepo{member: want}

	got, err := newTestChatUsecase(repo).AddMember(context.Background(), userID, chatID, newUserID)
	if err != nil || got != want || !repo.called {
		t.Fatalf("member = %+v, err = %v, repository called = %v", got, err, repo.called)
	}
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
	uc := newTestChatUsecase(mockRepo)

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
	uc := newTestChatUsecase(mockRepo)

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
	uc := newTestChatUsecase(mockRepo)

	items, err := uc.ListUserChats(context.Background(), uuid.New(), 10, 0, "")

	if err == nil || err.Error() != "database connection failed" {
		t.Errorf("expected db error, got %v", err)
	}
	if items != nil {
		t.Errorf("expected nil items on error, got %+v", items)
	}
}
