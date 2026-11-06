package contact

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestContactUsecase_CreateContact_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContactRepo := NewMockContactRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	uc := NewContactUsecase(mockContactRepo, mockUserRepo)

	userID := uuid.New()
	targetUserID := uuid.New()
	nickname := "testnick"
	var pseudonym *string

	mockUserRepo.EXPECT().
		FindByNickname(gomock.Any(), nickname).
		Return(&models.User{ID: targetUserID, Profile: &models.Profile{Nickname: nickname}}, nil)

	mockContactRepo.EXPECT().
		CheckContactExists(gomock.Any(), userID, targetUserID).
		Return(false, nil)

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
	uc := NewContactUsecase(mockContactRepo, mockUserRepo)

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
	uc := NewContactUsecase(mockContactRepo, mockUserRepo)

	userID := uuid.New()
	targetUserID := uuid.New()

	mockUserRepo.EXPECT().
		FindByNickname(gomock.Any(), "existing").
		Return(&models.User{ID: targetUserID, Profile: &models.Profile{Nickname: "existing"}}, nil)

	mockContactRepo.EXPECT().
		CheckContactExists(gomock.Any(), userID, targetUserID).
		Return(true, nil)

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
	uc := NewContactUsecase(mockContactRepo, mockUserRepo)

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
		GetContactsList(gomock.Any(), userID).
		Return(expectedContacts, nil)

	contacts, err := uc.GetContactsList(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(contacts) != 1 || contacts[0].Profile.Nickname != "friend" {
		t.Errorf("unexpected contacts list: %+v", contacts)
	}
}
