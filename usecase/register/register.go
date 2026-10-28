package register

import (
	"context"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"app/apperrors"
	"app/models"
	"app/repository/profile"
	"app/repository/user"
	"app/utils"
)

type RegisterUsecaseImpl struct {
	userRepo    user.Repository
	profileRepo profile.Repository
}

func NewRegisterUsecase(userRepo user.Repository, profileRepo profile.Repository) *RegisterUsecaseImpl {
	return &RegisterUsecaseImpl{
		userRepo:    userRepo,
		profileRepo: profileRepo,
	}
}

func (uc *RegisterUsecaseImpl) Register(ctx context.Context, user *models.User, plainPassword string) (*models.UserWithSession, error) {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
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

	return uc.userRepo.CreateUserProfileAndSession(ctx, user, string(hash), time.Now().Add(utils.SessionTTL))
}
