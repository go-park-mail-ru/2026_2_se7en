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

type mockUserRepo struct {
	user         *models.User
	err          error
	queriedEmail string
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	m.queriedEmail = email
	return m.user, m.err
}
func (m *mockUserRepo) FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return m.user, m.err
}
func (m *mockUserRepo) CheckEmailExist(ctx context.Context, email string) (bool, error) {
	return false, nil
}
func (m *mockUserRepo) CreateUser(ctx context.Context, email, hash string, phone *string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) CreateUserProfileAndSession(ctx context.Context, user *models.User, hash string, expiresAt time.Time) (*models.UserWithSession, error) {
	return nil, nil
}
func (m *mockUserRepo) DeleteUser(ctx context.Context, userID uuid.UUID) error { return nil }

type mockProfileRepo struct {
	profile *models.Profile
	err     error
}

func (m *mockProfileRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	return m.profile, m.err
}
func (m *mockProfileRepo) CheckNicknameExist(ctx context.Context, nickname string) (bool, error) {
	return false, nil
}
func (m *mockProfileRepo) CreateProfile(ctx context.Context, iconID *uuid.UUID, userID uuid.UUID, nickname, firstName string, lastName *string, bio *string) (*models.Profile, error) {
	return nil, nil
}
func (m *mockProfileRepo) FindProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	return nil, nil
}
func (m *mockProfileRepo) UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, error) {
	return nil, nil
}
func (m *mockProfileRepo) FindPasswordHash(ctx context.Context, userID uuid.UUID) (string, error) {
	return "", nil
}
func (m *mockProfileRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, oldHash, newHash string) (bool, error) {
	return false, nil
}
func (m *mockProfileRepo) UpdatePhone(ctx context.Context, userID uuid.UUID, phone *string) error {
	return nil
}

type mockSessionRepo struct {
	sessionErr           error
	deleteErr            error
	createdSessionUserID uuid.UUID
	deletedSessionID     uuid.UUID
}

func (m *mockSessionRepo) CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*models.Session, error) {
	if m.sessionErr != nil {
		return nil, m.sessionErr
	}
	m.createdSessionUserID = userID
	return &models.Session{ID: uuid.New(), UserID: userID, ExpiresAt: expiresAt}, nil
}
func (m *mockSessionRepo) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	m.deletedSessionID = sessionID
	return m.deleteErr
}
func (m *mockSessionRepo) GetSessionUser(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockSessionRepo) ExtendSession(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error {
	return nil
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
	profileModel := &models.Profile{ID: uuid.New(), Nickname: "testnick", FirstName: "Test"}

	userRepo := &mockUserRepo{user: user}
	profileRepo := &mockProfileRepo{profile: profileModel}
	sessionRepo := &mockSessionRepo{}

	uc := NewAuthUsecase(userRepo, profileRepo, sessionRepo)

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
	if sessionRepo.createdSessionUserID != user.ID {
		t.Errorf("expected session created for user %v, got %v", user.ID, sessionRepo.createdSessionUserID)
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
			userRepo := &mockUserRepo{user: tt.user}
			uc := NewAuthUsecase(userRepo, &mockProfileRepo{}, &mockSessionRepo{})

			_, err := uc.Login(context.Background(), "test@example.com", tt.password)

			if !errors.Is(err, apperrors.ErrInvalidCredentials) {
				t.Errorf("expected ErrInvalidCredentials, got %v", err)
			}
		})
	}
}

func TestAuthUsecase_Login_DBErrors(t *testing.T) {
	dbErr := errors.New("db down")
	userRepo := &mockUserRepo{err: dbErr}
	uc := NewAuthUsecase(userRepo, &mockProfileRepo{}, &mockSessionRepo{})

	_, err := uc.Login(context.Background(), "test@example.com", "password123")
	if err == nil || err.Error() != "db down" {
		t.Errorf("expected db down error, got %v", err)
	}
}

func TestAuthUsecase_GetMe_Success(t *testing.T) {
	user := makeTestUser("password123")
	profileModel := &models.Profile{ID: uuid.New(), Nickname: "testnick"}

	userRepo := &mockUserRepo{user: user}
	profileRepo := &mockProfileRepo{profile: profileModel}

	uc := NewAuthUsecase(userRepo, profileRepo, &mockSessionRepo{})

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
		userErr   error
		profErr   error
		wantError error
	}{
		{"user not found", nil, nil, nil, nil, apperrors.ErrUserNotFound},
		{"db error on find user", nil, nil, errors.New("db down"), nil, errors.New("db down")},
		{"db error on find profile", makeTestUser("pass"), nil, nil, errors.New("profile db down"), errors.New("profile db down")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{user: tt.user, err: tt.userErr}
			profileRepo := &mockProfileRepo{profile: tt.profile, err: tt.profErr}

			uc := NewAuthUsecase(userRepo, profileRepo, &mockSessionRepo{})

			_, err := uc.GetMe(context.Background(), uuid.New())

			if tt.wantError != nil && (err == nil || err.Error() != tt.wantError.Error()) {
				t.Errorf("expected error %v, got %v", tt.wantError, err)
			}
		})
	}
}

func TestAuthUsecase_Logout_Success(t *testing.T) {
	sessionRepo := &mockSessionRepo{}
	uc := NewAuthUsecase(&mockUserRepo{}, &mockProfileRepo{}, sessionRepo)

	sessionID := uuid.New()
	err := uc.Logout(context.Background(), sessionID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sessionRepo.deletedSessionID != sessionID {
		t.Errorf("expected session %v to be deleted, got %v", sessionID, sessionRepo.deletedSessionID)
	}
}
