package register

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"app/internal/apperrors"
	"app/internal/models"
)

type mockRepo struct {
	emailExists      bool
	checkEmailErr    error
	createUserErr    error
	deleteUserCalled bool

	nickExists       bool
	checkNickErr     error
	createProfileErr error

	createSessionErr error
}

func (m *mockRepo) CheckEmailExist(ctx context.Context, email string) (bool, error) {
	return m.emailExists, m.checkEmailErr
}
func (m *mockRepo) CreateUser(ctx context.Context, email, hash string, phone *string) (*models.User, error) {
	if m.createUserErr != nil {
		return nil, m.createUserErr
	}
	return &models.User{ID: uuid.New(), Email: email, PasswordHash: hash, PhoneNumber: phone}, nil
}
func (m *mockRepo) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	m.deleteUserCalled = true
	return nil
}

func (m *mockRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}
func (m *mockRepo) FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return nil, nil
}

func (m *mockRepo) CheckNicknameExist(ctx context.Context, nickname string) (bool, error) {
	return m.nickExists, m.checkNickErr
}
func (m *mockRepo) CreateProfile(ctx context.Context, iconID *uuid.UUID, userID uuid.UUID, nickname, firstName string, lastName *string, bio *string) (*models.Profile, error) {
	if m.createProfileErr != nil {
		return nil, m.createProfileErr
	}
	return &models.Profile{ID: uuid.New(), UserID: userID, Nickname: nickname, FirstName: firstName, LastName: lastName}, nil
}

func (m *mockRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	return nil, nil
}

func (m *mockRepo) CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*models.Session, error) {
	if m.createSessionErr != nil {
		return nil, m.createSessionErr
	}
	return &models.Session{ID: uuid.New(), UserID: userID, ExpiresAt: expiresAt}, nil
}

func (m *mockRepo) DeleteSession(ctx context.Context, sessionID uuid.UUID) error { return nil }
func (m *mockRepo) GetSessionUser(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockRepo) ExtendSession(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error {
	return nil
}

func TestRegisterUsecase_Success(t *testing.T) {
	repo := &mockRepo{}

	uc := NewRegisterUsecase(repo, repo, repo)

	user := &models.User{
		Email: "test@example.com",
		Profile: &models.Profile{
			Nickname:  "testnick",
			FirstName: "Test",
		},
	}

	result, err := uc.Register(context.Background(), user, "password123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.User.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, result.User.Email)
	}
	if result.User.Profile == nil || result.User.Profile.Nickname != "testnick" {
		t.Error("expected profile to be attached to user")
	}
	if result.SessionID == uuid.Nil {
		t.Error("expected session ID to be generated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(result.User.PasswordHash), []byte("password123")); err != nil {
		t.Errorf("password was not hashed correctly: %v", err)
	}
}

func TestRegisterUsecase_EmailTaken(t *testing.T) {
	repo := &mockRepo{emailExists: true}
	uc := NewRegisterUsecase(repo, repo, repo)

	user := &models.User{Email: "test@example.com", Profile: &models.Profile{Nickname: "test"}}
	_, err := uc.Register(context.Background(), user, "password123")

	if !errors.Is(err, apperrors.ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestRegisterUsecase_NicknameTaken(t *testing.T) {
	repo := &mockRepo{nickExists: true}
	uc := NewRegisterUsecase(repo, repo, repo)

	user := &models.User{Email: "test@example.com", Profile: &models.Profile{Nickname: "test"}}
	_, err := uc.Register(context.Background(), user, "password123")

	if !errors.Is(err, apperrors.ErrNicknameTaken) {
		t.Fatalf("expected ErrNicknameTaken, got %v", err)
	}
}

func TestRegisterUsecase_ProfileCreationRollback(t *testing.T) {
	repo := &mockRepo{createProfileErr: errors.New("db error")}
	uc := NewRegisterUsecase(repo, repo, repo)

	user := &models.User{Email: "test@example.com", Profile: &models.Profile{Nickname: "test"}}
	_, err := uc.Register(context.Background(), user, "password123")

	if err == nil {
		t.Fatal("expected error on profile creation")
	}
	if !repo.deleteUserCalled {
		t.Error("expected DeleteUser to be called as a rollback")
	}
}
