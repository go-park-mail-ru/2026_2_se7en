package contact

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
)

func TestContactRepository_CreateContact_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer mockPool.Close()

	repo := NewContactRepository(mockPool)
	userID := uuid.New()
	contactUserID := uuid.New()
	var pseudonym *string

	mockPool.ExpectExec("INSERT INTO user_contact").
		WithArgs(userID, contactUserID, pseudonym).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.CreateContact(context.Background(), userID, contactUserID, pseudonym)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestContactRepository_CreateContact_UniqueViolation(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer mockPool.Close()

	repo := NewContactRepository(mockPool)
	userID := uuid.New()
	contactUserID := uuid.New()

	mockPool.ExpectExec("INSERT INTO user_contact").
		WithArgs(userID, contactUserID, pgxmock.AnyArg()).
		WillReturnError(&pgconn.PgError{Code: "23505"})

	err = repo.CreateContact(context.Background(), userID, contactUserID, nil)

	if !errors.Is(err, apperrors.ErrContactAlreadyExists) {
		t.Errorf("expected ErrContactAlreadyExists, got %v", err)
	}

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestContactRepository_DeleteContact_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer mockPool.Close()

	repo := NewContactRepository(mockPool)
	userID := uuid.New()
	contactUserID := uuid.New()

	mockPool.ExpectExec("DELETE FROM user_contact").
		WithArgs(userID, contactUserID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.DeleteContact(context.Background(), userID, contactUserID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestContactRepository_DeleteContact_NotFound(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer mockPool.Close()

	repo := NewContactRepository(mockPool)
	userID := uuid.New()
	contactUserID := uuid.New()

	mockPool.ExpectExec("DELETE FROM user_contact").
		WithArgs(userID, contactUserID).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err = repo.DeleteContact(context.Background(), userID, contactUserID)
	if !errors.Is(err, apperrors.ErrContactNotFound) {
		t.Errorf("expected ErrContactNotFound, got %v", err)
	}

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestContactRepository_GetContactsList_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer mockPool.Close()

	repo := NewContactRepository(mockPool)
	userID := uuid.New()
	contactUserID := uuid.New()
	now := time.Now()
	iconID := uuid.New()
	profileID := uuid.New()

	lastName := "Doe"
	bio := "Hello"
	avatarURL := "https://example.test/avatar.png"

	rows := pgxmock.NewRows([]string{
		"user_id", "contact_user_id", "pseudonym", "created_at", "updated_at",
		"p_id", "p_user_id", "p_icon_id", "avatar_url", "p_nickname", "p_first_name",
		"p_last_name", "p_bio", "p_created_at", "p_updated_at",
	}).AddRow(
		userID, contactUserID, nil, now, now,
		profileID, contactUserID, &iconID, &avatarURL, "testnick", "John", &lastName, &bio, now, now,
	)

	mockPool.ExpectQuery("SELECT uc.user_id, uc.contact_user_id").
		WithArgs(userID, 20, 0, "").
		WillReturnRows(rows)

	contacts, err := repo.GetContactsList(context.Background(), userID, 20, 0, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(contacts))
	}

	if contacts[0].Profile.Nickname != "testnick" || contacts[0].Profile.AvatarURL == nil || *contacts[0].Profile.AvatarURL != avatarURL {
		t.Errorf("unexpected contact data: %+v", contacts[0])
	}

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
