package contact

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestContactUsecase_CreateContact_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContactRepo := NewMockContactRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	uc := NewContactUsecase(mockContactRepo, mockUserRepo, slog.Default())

	userID := uuid.New()
	targetUserID := uuid.New()
	nickname := "testnick"
	var pseudonym *string

	mockUserRepo.EXPECT().
		FindByNickname(gomock.Any(), nickname).
		Return(&models.User{ID: targetUserID, Profile: &models.Profile{Nickname: nickname}}, nil)

	mockContactRepo.EXPECT().
		CreateContact(gomock.Any(), userID, targetUserID, pseudonym).
		Return(nil)

	err := uc.CreateContact(context.Background(), userID, nickname, pseudonym)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestContactUsecase_CreateContact_SelfAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContactRepo := NewMockContactRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	uc := NewContactUsecase(mockContactRepo, mockUserRepo, slog.Default())

	userID := uuid.New()

	mockUserRepo.EXPECT().
		FindByNickname(gomock.Any(), "mynick").
		Return(&models.User{ID: userID, Profile: &models.Profile{Nickname: "mynick"}}, nil)

	err := uc.CreateContact(context.Background(), userID, "mynick", nil)
	if !errors.Is(err, apperrors.ErrCannotAddSelfAsContact) {
		t.Errorf("expected ErrCannotAddSelfAsContact, got %v", err)
	}
}

func TestContactUsecase_CreateContact_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContactRepo := NewMockContactRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	uc := NewContactUsecase(mockContactRepo, mockUserRepo, slog.Default())

	userID := uuid.New()
	targetUserID := uuid.New()

	mockUserRepo.EXPECT().
		FindByNickname(gomock.Any(), "existing").
		Return(&models.User{ID: targetUserID, Profile: &models.Profile{Nickname: "existing"}}, nil)

	mockContactRepo.EXPECT().
		CreateContact(gomock.Any(), userID, targetUserID, gomock.Any()).
		Return(apperrors.ErrContactAlreadyExists)

	err := uc.CreateContact(context.Background(), userID, "existing", nil)
	if !errors.Is(err, apperrors.ErrContactAlreadyExists) {
		t.Errorf("expected ErrContactAlreadyExists, got %v", err)
	}
}

func TestContactUsecase_GetContactsList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContactRepo := NewMockContactRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	uc := NewContactUsecase(mockContactRepo, mockUserRepo, slog.Default())

	userID := uuid.New()
	expectedContacts := []models.Contact{
		{
			UserID:        userID,
			ContactUserID: uuid.New(),
			Profile:       &models.Profile{Nickname: "friend", FirstName: "John"},
			IsOnline:      true,
		},
	}

	mockContactRepo.EXPECT().
		GetContactsList(gomock.Any(), userID, 20, 0, "").
		Return(expectedContacts, nil)

	contacts, err := uc.GetContactsList(context.Background(), userID, 20, 0, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(contacts) != 1 || contacts[0].Profile.Nickname != "friend" {
		t.Errorf("unexpected contacts list: %+v", contacts)
	}
}

func TestContactUsecase_CreateContact_PropagatesLookupErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "user missing", err: apperrors.ErrUserNotFound},
		{name: "repository failure", err: errors.New("lookup failed")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			userRepo := NewMockUserRepository(ctrl)
			contactRepo := NewMockContactRepository(ctrl)
			userRepo.EXPECT().FindByNickname(gomock.Any(), "absent").Return(nil, tc.err)
			err := NewContactUsecase(contactRepo, userRepo, slog.Default()).CreateContact(context.Background(), uuid.New(), "absent", nil)
			if !errors.Is(err, tc.err) {
				t.Fatalf("err = %v, want %v", err, tc.err)
			}
		})
	}
}

func TestContactUsecase_DeleteContact(t *testing.T) {
	for _, tc := range []struct {
		name             string
		repoErr, wantErr error
	}{
		{name: "success"},
		{name: "not found", repoErr: apperrors.ErrContactNotFound, wantErr: apperrors.ErrContactNotFound},
		{name: "database error", repoErr: errors.New("delete failed"), wantErr: errors.New("delete failed")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewMockContactRepository(ctrl)
			userID, contactID := uuid.New(), uuid.New()
			repo.EXPECT().DeleteContact(gomock.Any(), userID, contactID).Return(tc.repoErr)
			err := NewContactUsecase(repo, NewMockUserRepository(ctrl), slog.Default()).DeleteContact(context.Background(), userID, contactID)
			if tc.wantErr == nil && err != nil {
				t.Fatalf("err = %v", err)
			}
			if tc.wantErr != nil && (err == nil || err.Error() != tc.wantErr.Error()) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
		})
	}
}
