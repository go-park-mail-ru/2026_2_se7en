package chat

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
)

type mockChatRepo struct {
	items         []*models.ChatListItem
	activeIDs     []uuid.UUID
	detail        *models.ChatDetail
	dialogStatus  models.DialogStatus
	dialogID      uuid.UUID
	members       []models.ChatMember
	member        *models.ChatMember
	messages      []models.Message
	isMember      bool
	readChanged   bool
	messageExists bool
	messageAuthor uuid.UUID
	authorErr     error
	detailErr     error
	err           error
	called        bool
	listCalled    bool
	userID        uuid.UUID
	detailChatID  uuid.UUID
	limit         int
	chatType      string
	readChatID    uuid.UUID
	readMessageID uuid.UUID
}

func (m *mockChatRepo) IsMember(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatMembership, error) {
	m.called = true
	return &models.ChatMembership{IsMember: m.isMember}, m.err
}

func (m *mockChatRepo) GetActiveMemberIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	m.called = true
	return m.activeIDs, m.err
}

func (m *mockChatRepo) GetListUserChats(ctx context.Context, userID uuid.UUID, limit int, chatType string, cursorTime *time.Time, cursorID uuid.UUID) ([]*models.ChatListItem, error) {
	m.called = true
	m.userID = userID
	m.limit = limit
	m.chatType = chatType
	return m.items, m.err
}

func (m *mockChatRepo) GetChat(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error) {
	m.called = true
	m.detailChatID = chatID
	if m.detailErr != nil {
		return nil, m.detailErr
	}
	return m.detail, m.err
}

func (m *mockChatRepo) CreateChat(ctx context.Context, userID, peerID uuid.UUID) (uuid.UUID, models.DialogStatus, error) {
	m.called = true
	return m.dialogID, m.dialogStatus, m.err
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

func (m *mockChatRepo) UpdateGroupAvatar(context.Context, uuid.UUID, uuid.UUID, string, string, string) (*models.ChatDetail, error) {
	m.called = true
	return m.detail, m.err
}

func (m *mockChatRepo) DeleteChat(ctx context.Context, userID, chatID uuid.UUID) error {
	m.called = true
	return m.err
}

func (m *mockChatRepo) GetListMembers(ctx context.Context, userID, chatID uuid.UUID) ([]models.ChatMember, error) {
	m.called = true
	m.listCalled = true
	return m.members, m.err
}

func (m *mockChatRepo) AddMember(ctx context.Context, userID, chatID, newUserID uuid.UUID) (*models.ChatMember, error) {
	m.called = true
	return m.member, m.err
}

func (m *mockChatRepo) UpdateMemberRole(ctx context.Context, userID, chatID, targetID uuid.UUID, role string) (*models.ChatMember, error) {
	m.called = true
	return m.member, m.err
}

func (m *mockChatRepo) RemoveMember(ctx context.Context, userID, chatID, targetID uuid.UUID) error {
	m.called = true
	return m.err
}

func (m *mockChatRepo) MessageExists(ctx context.Context, chatID, messageID uuid.UUID) (bool, error) {
	m.called = true
	return m.messageExists, m.err
}

func (m *mockChatRepo) GetListMessages(ctx context.Context, userID, chatID uuid.UUID, limit int, before *uuid.UUID) ([]models.Message, error) {
	m.called = true
	m.listCalled = true
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

func (m *mockChatRepo) GetMessageAuthor(ctx context.Context, chatID, messageID uuid.UUID) (uuid.UUID, error) {
	m.called = true
	return m.messageAuthor, m.authorErr
}

func newTestChatUsecase(repo *mockChatRepo) *ChatUsecaseImpl {
	return NewChatUsecase(repo, slog.Default())
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

func TestChatUsecase_MarkMessageReadUnchanged(t *testing.T) {
	userID, chatID, messageID, peerID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	tests := []struct {
		name      string
		authorID  uuid.UUID
		authorErr error
		wantErr   error
	}{
		{name: "already read", authorID: peerID},
		{name: "own message", authorID: userID, wantErr: apperrors.ErrInvalidChatOperation},
		{name: "missing message", authorErr: apperrors.ErrMessageNotFound, wantErr: apperrors.ErrMessageNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockChatRepo{messageAuthor: tc.authorID, authorErr: tc.authorErr}
			changed, err := newTestChatUsecase(repo).MarkMessageRead(context.Background(), userID, chatID, messageID)
			if changed || !errors.Is(err, tc.wantErr) {
				t.Fatalf("changed = %t, err = %v, want %v", changed, err, tc.wantErr)
			}
		})
	}
}

func TestChatUsecase_GetListMessagesChecksMembershipAndCursor(t *testing.T) {
	userID, chatID, cursor := uuid.New(), uuid.New(), uuid.New()
	for _, tc := range []struct {
		name    string
		repo    *mockChatRepo
		wantErr error
	}{
		{name: "not a member", repo: &mockChatRepo{detailErr: apperrors.ErrChatNotFound}, wantErr: apperrors.ErrChatNotFound},
		{name: "cursor not in chat", repo: &mockChatRepo{}, wantErr: apperrors.ErrMessageNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newTestChatUsecase(tc.repo).GetListMessages(context.Background(), userID, chatID, 20, &cursor)
			if !errors.Is(err, tc.wantErr) || tc.repo.listCalled {
				t.Fatalf("err = %v, want %v, list called = %t", err, tc.wantErr, tc.repo.listCalled)
			}
		})
	}
}

func TestChatUsecase_GetListMembersRejectsNonMember(t *testing.T) {
	repo := &mockChatRepo{detailErr: apperrors.ErrChatNotFound, members: []models.ChatMember{{UserID: uuid.New()}}}
	_, err := newTestChatUsecase(repo).GetListMembers(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, apperrors.ErrChatNotFound) || repo.listCalled {
		t.Fatalf("err = %v, list called = %t", err, repo.listCalled)
	}
}

func TestChatUsecase_CreateDialogLoadsDetail(t *testing.T) {
	userID, peerID, chatID := uuid.New(), uuid.New(), uuid.New()
	detail := &models.ChatDetail{Chat: &models.Chat{ID: chatID, Type: "dialog"}}
	repo := &mockChatRepo{dialogID: chatID, dialogStatus: models.DialogCreated, detail: detail}
	got, status, err := newTestChatUsecase(repo).CreateDialog(context.Background(), userID, peerID)
	if err != nil || got != detail || status != models.DialogCreated || repo.detailChatID != chatID {
		t.Fatalf("detail = %+v, status = %v, err = %v, fetched chat = %s", got, status, err, repo.detailChatID)
	}
}

func TestChatUsecase_ValidGroupAndMessagesReachRepository(t *testing.T) {
	userID, chatID, peerID := uuid.New(), uuid.New(), uuid.New()
	name := strings.Repeat("я", 100)
	description := strings.Repeat("я", 500)
	repo := &mockChatRepo{messageExists: true}
	uc := newTestChatUsecase(repo)

	if _, err := uc.CreateGroup(context.Background(), userID, name, &description, []uuid.UUID{peerID}); err != nil || !repo.called {
		t.Fatalf("valid group was not forwarded: err=%v, called=%v", err, repo.called)
	}
	repo.called = false
	if _, err := uc.UpdateGroup(context.Background(), userID, chatID, nil, nil, true); err != nil || !repo.called {
		t.Fatalf("clearing description was not forwarded: err=%v, called=%v", err, repo.called)
	}
	repo.called = false
	if _, err := uc.GetListMessages(context.Background(), userID, chatID, 50, &peerID); err != nil || !repo.called {
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

func TestChatUsecase_GetListUserChats_Success(t *testing.T) {
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

	items, err := uc.GetListUserChats(context.Background(), userID, 10, "dialog", nil, uuid.Nil)

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

func TestChatUsecase_GetListUserChats_NilToEmptySlice(t *testing.T) {
	mockRepo := &mockChatRepo{items: nil}
	uc := newTestChatUsecase(mockRepo)

	items, err := uc.GetListUserChats(context.Background(), uuid.New(), 10, "", nil, uuid.Nil)

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

func TestChatUsecase_GetListUserChats_Error(t *testing.T) {
	dbErr := errors.New("database connection failed")
	mockRepo := &mockChatRepo{err: dbErr}
	uc := newTestChatUsecase(mockRepo)

	items, err := uc.GetListUserChats(context.Background(), uuid.New(), 10, "", nil, uuid.Nil)

	if err == nil || err.Error() != "database connection failed" {
		t.Errorf("expected db error, got %v", err)
	}
	if items != nil {
		t.Errorf("expected nil items on error, got %+v", items)
	}
}

func TestChatUsecase_DelegatesChatAndGroupOperations(t *testing.T) {
	ctx := context.Background()
	userID, chatID, peerID := uuid.New(), uuid.New(), uuid.New()
	repo := &mockChatRepo{
		isMember:  true,
		activeIDs: []uuid.UUID{userID, peerID},
		detail:    &models.ChatDetail{Chat: &models.Chat{ID: chatID}},
		member:    &models.ChatMember{UserID: peerID},
	}
	uc := newTestChatUsecase(repo)

	membership, err := uc.IsMember(ctx, userID, chatID)
	if err != nil || membership == nil || !membership.IsMember {
		t.Fatalf("IsMember = %+v, %v", membership, err)
	}
	ids, err := uc.GetActiveMemberIDs(ctx, chatID)
	if err != nil || len(ids) != 2 {
		t.Fatalf("GetActiveMemberIDs = %v, %v", ids, err)
	}
	detail, err := uc.GetChatDetail(ctx, userID, chatID)
	if err != nil || detail != repo.detail {
		t.Fatalf("GetChatDetail = %v, %v", detail, err)
	}
	if err := uc.DeleteChat(ctx, userID, chatID); err != nil {
		t.Fatalf("DeleteChat: %v", err)
	}
	if err := uc.HideDialog(ctx, userID, chatID); err != nil {
		t.Fatalf("HideDialog: %v", err)
	}
	members, err := uc.GetListMembers(ctx, userID, chatID)
	if err != nil || len(members) != 0 || !repo.listCalled {
		t.Fatalf("GetListMembers = %v, %v", members, err)
	}
	member, err := uc.UpdateMemberRole(ctx, userID, chatID, peerID, "admin")
	if err != nil || member != repo.member {
		t.Fatalf("UpdateMemberRole = %v, %v", member, err)
	}
	if err := uc.RemoveMember(ctx, userID, chatID, peerID); err != nil {
		t.Fatalf("RemoveMember: %v", err)
	}
	content := "hi"
	repo.messages = []models.Message{{ID: uuid.New(), Content: &content}}
	message, err := uc.CreateMessage(ctx, userID, chatID, nil, nil, &content, "text")
	if err != nil || message == nil || message.Content == nil || *message.Content != content {
		t.Fatalf("CreateMessage = %v, %v", message, err)
	}
}

func TestChatUsecase_CreateDialogAndGetListMembersPropagateRepositoryErrors(t *testing.T) {
	want := errors.New("repo failed")
	userID, peerID, chatID := uuid.New(), uuid.New(), uuid.New()
	t.Run("create dialog", func(t *testing.T) {
		repo := &mockChatRepo{err: want}
		detail, status, err := newTestChatUsecase(repo).CreateDialog(context.Background(), userID, peerID)
		if detail != nil || status != models.DialogExisting || !errors.Is(err, want) {
			t.Fatalf("got %v, %v, %v", detail, status, err)
		}
	})
	t.Run("fetch dialog detail", func(t *testing.T) {
		repo := &mockChatRepo{dialogID: chatID, detailErr: want}
		detail, status, err := newTestChatUsecase(repo).CreateDialog(context.Background(), userID, peerID)
		if detail != nil || status != models.DialogExisting || !errors.Is(err, want) {
			t.Fatalf("got %v, %v, %v", detail, status, err)
		}
	})
	t.Run("list members", func(t *testing.T) {
		repo := &mockChatRepo{detailErr: want}
		members, err := newTestChatUsecase(repo).GetListMembers(context.Background(), userID, chatID)
		if members != nil || !errors.Is(err, want) || repo.listCalled {
			t.Fatalf("got %v, %v, list called=%t", members, err, repo.listCalled)
		}
	})
}

func TestChatUsecase_CreateMessagePropagatesRepositoryError(t *testing.T) {
	want := errors.New("write failed")
	repo := &mockChatRepo{err: want}
	message, err := newTestChatUsecase(repo).CreateMessage(context.Background(), uuid.New(), uuid.New(), nil, nil, nil, "text")
	if message != nil || !errors.Is(err, want) {
		t.Fatalf("message = %v, err = %v", message, err)
	}
}

func TestChatUsecase_UploadGroupAvatar(t *testing.T) {
	t.Run("saves and attaches avatar", func(t *testing.T) {
		repo := &mockChatRepo{detail: &models.ChatDetail{Chat: &models.Chat{ID: uuid.New()}}}
		upload := &models.FileUpload{Reader: io.NopCloser(strings.NewReader("image")), Filename: "a.png", Size: 5, ContentType: "image/png"}
		got, err := newTestChatUsecase(repo).UploadGroupAvatar(context.Background(), uuid.New(), uuid.New(), upload)
		if err != nil || got != repo.detail || !repo.called {
			t.Fatalf("detail = %v, err = %v, called = %t", got, err, repo.called)
		}
	})
	t.Run("rejects invalid upload", func(t *testing.T) {
		repo := &mockChatRepo{}
		got, err := newTestChatUsecase(repo).UploadGroupAvatar(context.Background(), uuid.New(), uuid.New(), &models.FileUpload{Reader: strings.NewReader("x"), Size: 1, ContentType: "text/plain"})
		if err == nil || got != nil || repo.called {
			t.Fatalf("detail = %v, err = %v, called = %t", got, err, repo.called)
		}
	})
}
