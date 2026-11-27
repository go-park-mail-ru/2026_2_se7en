package register

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/auth/model"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/infrastructure/http/observability/logger"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUsecaseImpl struct {
	log *slog.Logger

	userRepo    UserRepository
	profileRepo ProfileRepository
}

func NewRegisterUsecase(userRepo UserRepository, profileRepo ProfileRepository, baseLog *slog.Logger) *RegisterUsecaseImpl {
	return &RegisterUsecaseImpl{
		log: logger.Scope(baseLog, logger.LayerUsecase, logger.ModuleRegister),

		userRepo:    userRepo,
		profileRepo: profileRepo,
	}
}

func (uc *RegisterUsecaseImpl) Register(ctx context.Context, user *models.User, plainPassword string) (*models.UserWithSession, error) {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	emailExists, err := uc.userRepo.CheckEmailExist(ctx, user.Email)
	if err != nil {
		uc.log.ErrorContext(ctx, "Register failed", logger.Error(err))

		return nil, err
	}

	if emailExists {
		return nil, models.ErrEmailTaken
	}

	if user.Profile == nil {
		return nil, models.ErrProfileRequired
	}

	nicknameExists, err := uc.profileRepo.CheckNicknameExist(ctx, user.Profile.Nickname)
	if err != nil {
		uc.log.ErrorContext(ctx, "Register failed", logger.Error(err))

		return nil, err
	}

	if nicknameExists {
		return nil, models.ErrNicknameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		uc.log.ErrorContext(ctx, "Register failed", logger.Error(err))

		return nil, err
	}

	return uc.userRepo.CreateUserProfileAndSession(ctx, user, string(hash), time.Now().Add(models.SessionTTL))
}
