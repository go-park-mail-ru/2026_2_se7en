package users

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v5"
)

func TestUsersRepository_GetUsersList_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("create pgx mock: %v", err)
	}
	defer mockPool.Close()

	repo := NewUsersRepository(mockPool, slog.Default())
	requesterID, userID, profileID, iconID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now()
	lastName, bio, avatarURL := "Lovelace", "Researcher", "https://example.test/avatar.webp"
	rows := pgxmock.NewRows([]string{
		"id", "icon_id", "user_id", "avatar_url", "nickname", "first_name", "last_name", "bio",
		"created_at", "updated_at", "is_contact",
	}).AddRow(profileID, &iconID, userID, &avatarURL, "ada_l", "Ada", &lastName, &bio, now, now, true)

	mockPool.ExpectQuery("FROM profile p").
		WithArgs(requesterID, 21, 20, "ada").
		WillReturnRows(rows)

	users, err := repo.GetUsersList(context.Background(), requesterID, 21, 20, "ada")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 matching user, got %d", len(users))
	}
	if users[0].Profile.UserID != userID || users[0].Profile.Nickname != "ada_l" || !users[0].IsContact {
		t.Fatalf("unexpected user search result: %+v", users[0])
	}
	if users[0].Profile.AvatarURL == nil || *users[0].Profile.AvatarURL != avatarURL {
		t.Fatalf("avatar URL was not mapped: %+v", users[0].Profile.AvatarURL)
	}
	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet pgx expectations: %v", err)
	}
}

func TestUsersRepository_GetUsersList_QueryError(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("create pgx mock: %v", err)
	}
	defer mockPool.Close()

	repo := NewUsersRepository(mockPool, slog.Default())
	mockPool.ExpectQuery("FROM profile p").
		WithArgs(uuid.Nil, 11, 0, "nick").
		WillReturnError(errors.New("database unavailable"))

	_, err = repo.GetUsersList(context.Background(), uuid.Nil, 11, 0, "nick")
	if err == nil {
		t.Fatal("expected database error")
	}
	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet pgx expectations: %v", err)
	}
}
