package register

import (
	"context"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUsecaseImpl struct {
	userRepo    UserRepository
	profileRepo ProfileRepository
}

func NewRegisterUsecase(userRepo UserRepository, profileRepo ProfileRepository) *RegisterUsecaseImpl {
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

	return uc.userRepo.CreateUserProfileAndSession(ctx, user, string(hash), time.Now().Add(middleware.SessionTTL))
}
