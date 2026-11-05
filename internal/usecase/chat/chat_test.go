package chat

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
)

type mockChatRepo struct {
	items    []*models.ChatListItem
	detail   *models.ChatDetail
	members  []models.ChatMember
	member   *models.ChatMember
	messages []models.Message
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
	m.called = true
	return m.detail, m.err
}
func (m *mockChatRepo) GetOrCreateDialog(ctx context.Context, userID, peerID uuid.UUID) (*models.ChatDetail, error) {
	m.called = true
	return m.detail, m.err
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

func newTestChatUsecase(repo ChatRepository) *ChatUsecaseImpl {
	return NewChatUsecase(repo)
}

func TestChatUsecase_RejectsInvalidInputBeforeRepository(t *testing.T) {
	userID, chatID, peerID := uuid.New(), uuid.New(), uuid.New()
	longName := strings.Repeat("я", 129)
	longDescription := strings.Repeat("я", 501)
	duplicateMembers := []uuid.UUID{peerID, peerID}

	tests := []struct {
		name  string
		field string
		call  func(*ChatUsecaseImpl) error
	}{
		{"dialog: self", "peer_id", func(uc *ChatUsecaseImpl) error {
			_, err := uc.GetOrCreateDialog(context.Background(), userID, userID)
			return err
		}},
		{"create group: empty name", "name", func(uc *ChatUsecaseImpl) error {
			_, err := uc.CreateGroup(context.Background(), userID, "  ", nil, nil)
			return err
		}},
		{"create group: long name", "name", func(uc *ChatUsecaseImpl) error {
			_, err := uc.CreateGroup(context.Background(), userID, longName, nil, nil)
			return err
		}},
		{"create group: long description", "description", func(uc *ChatUsecaseImpl) error {
			_, err := uc.CreateGroup(context.Background(), userID, "Group", &longDescription, nil)
			return err
		}},
		{"create group: creator in members", "members", func(uc *ChatUsecaseImpl) error {
			_, err := uc.CreateGroup(context.Background(), userID, "Group", nil, []uuid.UUID{userID})
			return err
		}},
		{"create group: duplicate members", "members", func(uc *ChatUsecaseImpl) error {
			_, err := uc.CreateGroup(context.Background(), userID, "Group", nil, duplicateMembers)
			return err
		}},
		{"update group: no changes", "name", func(uc *ChatUsecaseImpl) error {
			_, err := uc.UpdateGroup(context.Background(), userID, chatID, nil, nil, false)
			return err
		}},
		{"update group: empty name", "name", func(uc *ChatUsecaseImpl) error {
			name := " "
			_, err := uc.UpdateGroup(context.Background(), userID, chatID, &name, nil, false)
			return err
		}},
		{"update group: long description", "description", func(uc *ChatUsecaseImpl) error {
			_, err := uc.UpdateGroup(context.Background(), userID, chatID, nil, &longDescription, true)
			return err
		}},
		{"update group: description without flag", "description", func(uc *ChatUsecaseImpl) error {
			name, description := "Group", "New description"
			_, err := uc.UpdateGroup(context.Background(), userID, chatID, &name, &description, false)
			return err
		}},
		{"add member: self", "new_user_id", func(uc *ChatUsecaseImpl) error {
			_, err := uc.AddMember(context.Background(), userID, chatID, userID)
			return err
		}},
		{"list messages: zero limit", "limit", func(uc *ChatUsecaseImpl) error {
			_, err := uc.ListMessages(context.Background(), userID, chatID, 0, nil)
			return err
		}},
		{"list messages: excessive limit", "limit", func(uc *ChatUsecaseImpl) error {
			_, err := uc.ListMessages(context.Background(), userID, chatID, 51, nil)
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockChatRepo{}
			err := tt.call(newTestChatUsecase(repo))
			var appErr *apperrors.AppError
			if !errors.As(err, &appErr) || appErr.Status != http.StatusBadRequest || len(appErr.Details) != 1 || appErr.Details[0].Field != tt.field {
				t.Fatalf("error = %v, want bad request for %s", err, tt.field)
			}
			if repo.called {
				t.Fatal("repository called for invalid input")
			}
		})
	}
}

func TestChatUsecase_ValidGroupAndMessagesReachRepository(t *testing.T) {
	userID, chatID, peerID := uuid.New(), uuid.New(), uuid.New()
	name := strings.Repeat("я", 128)
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
