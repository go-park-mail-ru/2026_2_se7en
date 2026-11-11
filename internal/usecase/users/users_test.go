package users

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestUsersUsecase_GetUsersList_ForwardsSearchAndPage(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockUsersRepository(ctrl)
	usecase := NewUsersUsecase(mockRepo, slog.Default())
	requesterID := uuid.New()
	expected := []models.UserSearchResult{{
		Profile: models.Profile{UserID: uuid.New(), Nickname: "ada_l"}, IsContact: true,
	}}
	mockRepo.EXPECT().GetUsersList(gomock.Any(), requesterID, 21, 20, "ada").Return(expected, nil)

	users, err := usecase.GetUsersList(context.Background(), requesterID, 21, 20, "ada")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 1 || users[0].Profile.Nickname != "ada_l" || !users[0].IsContact {
		t.Fatalf("unexpected search results: %+v", users)
	}
}

func TestUsersUsecase_GetUsersList_PropagatesRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockUsersRepository(ctrl)
	usecase := NewUsersUsecase(mockRepo, slog.Default())
	wantErr := errors.New("repository failed")
	mockRepo.EXPECT().GetUsersList(gomock.Any(), gomock.Any(), 11, 0, "nick").Return(nil, wantErr)

	_, err := usecase.GetUsersList(context.Background(), uuid.New(), 11, 0, "nick")
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}
