package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"app/apperrors"
	"app/models"
)

type mockAuthRepo struct {
	user       *models.User
	profile    *models.Profile
	err        error
	profileErr error
	sessionErr error
	deleteErr  error

	queriedEmail         string
	createdSessionUserID uuid.UUID
	deletedSessionID     uuid.UUID
}

func (m *mockAuthRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	m.queriedEmail = email
	return m.user, m.err
}

func (m *mockAuthRepo) FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return m.user, m.err
}

func (m *mockAuthRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	return m.profile, m.profileErr
}

func (m *mockAuthRepo) CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*models.Session, error) {
	if m.sessionErr != nil {
		return nil, m.sessionErr
	}
	m.createdSessionUserID = userID
	return &models.Session{ID: uuid.New(), UserID: userID, ExpiresAt: expiresAt}, nil
}

func (m *mockAuthRepo) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	m.deletedSessionID = sessionID
	return m.deleteErr
}

func makeTestUser(password string) *models.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hash),
	}
}

func TestAuthUsecase_Login_Success(t *testing.T) {
	user := makeTestUser("password123")
	profile := &models.Profile{ID: uuid.New(), Nickname: "testnick", FirstName: "Test"}
	repo := &mockAuthRepo{user: user, profile: profile}

	uc := NewAuthUsecase(repo, repo, repo)

	result, err := uc.Login(context.Background(), "test@example.com", "password123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.User.ID != user.ID {
		t.Errorf("expected user ID %v, got %v", user.ID, result.User.ID)
	}
	if result.User.Profile == nil || result.User.Profile.Nickname != "testnick" {
		t.Error("expected profile to be attached to user")
	}
	if result.SessionID == uuid.Nil {
		t.Error("expected session ID to be generated")
	}
	if repo.createdSessionUserID != user.ID {
		t.Errorf("expected session created for user %v, got %v", user.ID, repo.createdSessionUserID)
	}
}

func TestAuthUsecase_Login_InvalidCredentials(t *testing.T) {
	tests := []struct {
		name     string
		user     *models.User
		password string
	}{
		{"user not found", nil, "password123"},
		{"wrong password", makeTestUser("correctpassword"), "wrongpassword"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuthRepo{user: tt.user}
			uc := NewAuthUsecase(repo, repo, repo)

			_, err := uc.Login(context.Background(), "test@example.com", tt.password)

			if !errors.Is(err, apperrors.ErrInvalidCredentials) {
				t.Errorf("expected ErrInvalidCredentials, got %v", err)
			}
		})
	}
}

func TestAuthUsecase_Login_DBErrors(t *testing.T) {
	repo := &mockAuthRepo{err: errors.New("db down")}
	uc := NewAuthUsecase(repo, repo, repo)

	_, err := uc.Login(context.Background(), "test@example.com", "password123")
	if err == nil || err.Error() != "db down" {
		t.Errorf("expected db down error, got %v", err)
	}
}

func TestAuthUsecase_GetMe_Success(t *testing.T) {
	user := makeTestUser("password123")
	profile := &models.Profile{ID: uuid.New(), Nickname: "testnick"}
	repo := &mockAuthRepo{user: user, profile: profile}
	uc := NewAuthUsecase(repo, repo, repo)

	result, err := uc.GetMe(context.Background(), user.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.User.Profile == nil {
		t.Error("expected profile to be attached to user")
	}
}

func TestAuthUsecase_GetMe_Errors(t *testing.T) {
	tests := []struct {
		name      string
		user      *models.User
		profile   *models.Profile
		err       error
		wantError error
	}{
		{"user not found", nil, nil, nil, apperrors.ErrUserNotFound},
		{"db error on find user", nil, nil, errors.New("db down"), errors.New("db down")},
		{"db error on find profile", makeTestUser("pass"), nil, errors.New("profile db down"), errors.New("profile db down")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuthRepo{user: tt.user, profile: tt.profile, err: tt.err, profileErr: tt.err}
			uc := NewAuthUsecase(repo, repo, repo)

			_, err := uc.GetMe(context.Background(), uuid.New())

			if tt.wantError != nil && (err == nil || err.Error() != tt.wantError.Error()) {
				t.Errorf("expected error %v, got %v", tt.wantError, err)
			}
		})
	}
}

func TestAuthUsecase_Logout_Success(t *testing.T) {
	repo := &mockAuthRepo{}
	uc := NewAuthUsecase(repo, repo, repo)
	sessionID := uuid.New()

	err := uc.Logout(context.Background(), sessionID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.deletedSessionID != sessionID {
		t.Errorf("expected session %v to be deleted, got %v", sessionID, repo.deletedSessionID)
	}
}
