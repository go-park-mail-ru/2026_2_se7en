package register

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"app/internal/apperrors"
	"app/internal/models"
	"app/internal/repository/profile"
	"app/internal/repository/session"
	"app/internal/repository/user"
	"app/internal/utils"
)

type RegisterUsecaseImpl struct {
	userRepo    user.Repository
	profileRepo profile.Repository
	sessionRepo session.Repository
}

func NewRegisterUsecase(userRepo user.Repository, profileRepo profile.Repository, sessionRepo session.Repository) *RegisterUsecaseImpl {
	return &RegisterUsecaseImpl{
		userRepo:    userRepo,
		profileRepo: profileRepo,
		sessionRepo: sessionRepo,
	}
}

func (uc *RegisterUsecaseImpl) Register(ctx context.Context, user *models.User, plainPassword string) (*models.UserWithSession, error) {
	emailExists, err := uc.userRepo.CheckEmailExist(ctx, user.Email)
	if err != nil {
		return nil, err
	}

	if emailExists {
		return nil, apperrors.ErrEmailTaken
	}

	if user.Profile == nil {
		return nil, apperrors.NewBadRequest("Profile is required", nil)
	}

	nicknameExists, err := uc.profileRepo.CheckNicknameExist(ctx, user.Profile.Nickname)
	if err != nil {
		return nil, err
	}

	if nicknameExists {
		return nil, apperrors.ErrNicknameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	modelUser, err := uc.userRepo.CreateUser(ctx, user.Email, string(hash), user.PhoneNumber)
	if err != nil {
		return nil, err
	}

	profile, err := uc.profileRepo.CreateProfile(
		ctx,
		user.Profile.IconID,
		modelUser.ID,
		user.Profile.Nickname,
		user.Profile.FirstName,
		user.Profile.LastName,
		user.Profile.Bio,
	)
	if err != nil {
		_ = uc.userRepo.DeleteUser(ctx, modelUser.ID)
		return nil, err
	}

	modelUser.Profile = profile

	session, err := uc.sessionRepo.CreateSession(ctx, modelUser.ID, time.Now().Add(utils.SessionTTL))
	if err != nil {
		return nil, err
	}

	return &models.UserWithSession{
		User:      modelUser,
		SessionID: session.ID,
	}, nil
}
