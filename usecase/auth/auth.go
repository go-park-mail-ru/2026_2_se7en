package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"app/apperrors"
	"app/models"
	"app/repository/profile"
	"app/repository/session"
	"app/repository/user"
	"app/utils"
)

var fakeHash []byte

func init() {
	fakeHash, _ = bcrypt.GenerateFromPassword([]byte("fake"), bcrypt.DefaultCost)
}

type AuthUsecaseImpl struct {
	userRepo    user.Repository
	profileRepo profile.Repository
	sessionRepo session.Repository
}

func NewAuthUsecase(userRepo user.Repository, profileRepo profile.Repository, sessionRepo session.Repository) *AuthUsecaseImpl {
	return &AuthUsecaseImpl{
		userRepo:    userRepo,
		profileRepo: profileRepo,
		sessionRepo: sessionRepo,
	}
}

func (uc *AuthUsecaseImpl) Login(ctx context.Context, email, password string) (*models.UserWithSession, error) {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, apperrors.ErrUserNotFound) {
		return nil, err
	}

	var hash []byte
	if err == nil && user != nil {
		hash = []byte(user.PasswordHash)
	} else {
		hash = fakeHash
	}

	if bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	if user == nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	profile, err := uc.profileRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	user.Profile = profile

	session, err := uc.sessionRepo.CreateSession(ctx, user.ID, time.Now().Add(utils.SessionTTL))
	if err != nil {
		return nil, err
	}

	return &models.UserWithSession{
		User:      user,
		SessionID: session.ID,
	}, nil
}

func (uc *AuthUsecaseImpl) GetMe(ctx context.Context, userID uuid.UUID) (*models.UserWithSession, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}

	profile, err := uc.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Profile = profile

	return &models.UserWithSession{
		User: user,
	}, nil
}

func (uc *AuthUsecaseImpl) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return uc.sessionRepo.DeleteSession(ctx, sessionID)
}
