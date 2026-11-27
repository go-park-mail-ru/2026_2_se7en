package profile

import (
	"context"
	"log/slog"
	"strings"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/auth/model"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/infrastructure/http/observability/logger"
	mediaavatar "github.com/go-park-mail-ru/2026_2_se7en/internal/media/usecase/avatar"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	mediamodel "github.com/go-park-mail-ru/2026_2_se7en/internal/media/model"
)

type UpdateParams struct {
	Email       *string
	Nickname    *string
	FirstName   *string
	LastName    *string
	Bio         *string
	PhoneNumber *string
}

type ProfileUsecaseImpl struct {
	log *slog.Logger

	txManager     TxManager
	userRepo      UserRepository
	profileRepo   ProfileRepository
	fileRepo      FileRepository
	avatarUsecase mediaavatar.AvatarUsecase
}

func NewProfileUsecase(txManager TxManager, userRepo UserRepository, profileRepo ProfileRepository, fileRepo FileRepository, avatarUsecase mediaavatar.AvatarUsecase, baseLog *slog.Logger) *ProfileUsecaseImpl {
	return &ProfileUsecaseImpl{
		log: logger.Scope(baseLog, logger.LayerUsecase, logger.ModuleProfile),

		txManager:     txManager,
		userRepo:      userRepo,
		profileRepo:   profileRepo,
		fileRepo:      fileRepo,
		avatarUsecase: avatarUsecase,
	}
}

func (uc *ProfileUsecaseImpl) GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, *models.Profile, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		uc.log.ErrorContext(ctx, "GetProfile failed", logger.Error(err))

		return nil, nil, err
	}

	if user == nil {
		return nil, nil, models.ErrUserNotFound
	}

	prof, err := uc.profileRepo.GetWithFileByUserID(ctx, userID)
	if err != nil {
		uc.log.ErrorContext(ctx, "GetProfile failed", logger.Error(err))

		return nil, nil, err
	}

	return user, prof, nil
}

func (uc *ProfileUsecaseImpl) UpdateProfile(ctx context.Context, userID uuid.UUID, params UpdateParams) error {

	return uc.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		user, err := uc.userRepo.FindByID(ctx, userID)
		if err != nil {
			uc.log.ErrorContext(ctx, "UpdateProfile failed", logger.Error(err))

			return err
		}
		if user == nil {
			return models.ErrUserNotFound
		}

		prof, err := uc.profileRepo.GetWithFileByUserID(ctx, userID)
		if err != nil {
			uc.log.ErrorContext(ctx, "UpdateProfile failed", logger.Error(err))

			return err
		}
		if prof == nil {
			return models.ErrProfileNotFound
		}

		if params.Email != nil && *params.Email != user.Email {
			if err := uc.userRepo.UpdateEmail(ctx, userID, *params.Email); err != nil {
				uc.log.ErrorContext(ctx, "UpdateProfile failed", logger.Error(err))

				return err
			}
		}

		if params.PhoneNumber != nil {
			if err := uc.userRepo.UpdatePhone(ctx, userID, params.PhoneNumber); err != nil {
				uc.log.ErrorContext(ctx, "UpdateProfile failed", logger.Error(err))

				return err
			}
		}

		nickname := prof.Nickname
		firstName := prof.FirstName
		lastName := prof.LastName
		bio := prof.Bio

		if params.Nickname != nil {
			nickname = *params.Nickname
		}

		if params.FirstName != nil {
			firstName = *params.FirstName
		}
		if params.LastName != nil {
			value := strings.TrimSpace(*params.LastName)
			if value == "" {
				lastName = nil
			} else {
				lastName = &value
			}
		}
		if params.Bio != nil {
			bio = params.Bio
		}

		return uc.profileRepo.Update(ctx, prof.ID, nickname, firstName, lastName, bio)
	})
}

func (uc *ProfileUsecaseImpl) UploadAvatar(ctx context.Context, userID uuid.UUID, file *mediamodel.FileUpload) (*mediamodel.File, error) {
	url, filename, err := uc.avatarUsecase.UploadAvatar(ctx, file)
	if err != nil {
		uc.log.ErrorContext(ctx, "UploadAvatar failed", logger.Error(err))

		return nil, err
	}

	var fileModel *mediamodel.File
	err = uc.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		prof, err := uc.profileRepo.GetWithFileByUserID(ctx, userID)
		if err != nil {
			uc.log.ErrorContext(ctx, "UploadAvatar failed", logger.Error(err))

			return err
		}
		if prof == nil {
			return models.ErrProfileNotFound
		}

		fileModel, err = uc.fileRepo.Create(ctx, url, filename, file.ContentType)
		if err != nil {
			uc.log.ErrorContext(ctx, "UploadAvatar failed", logger.Error(err))

			return err
		}
		if fileModel == nil {
			return models.ErrFileNotCreated
		}

		if err := uc.profileRepo.UpdateIcon(ctx, prof.ID, fileModel.ID); err != nil {
			uc.log.ErrorContext(ctx, "UploadAvatar failed", logger.Error(err))

			return err
		}

		return nil
	})

	if err != nil {
		_ = uc.avatarUsecase.RemoveAvatar(ctx, filename)
		uc.log.ErrorContext(ctx, "UploadAvatar failed", logger.Error(err))

		return nil, err
	}

	return fileModel, nil
}

func (uc *ProfileUsecaseImpl) UpdatePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {

	return uc.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		user, err := uc.userRepo.FindByID(ctx, userID)
		if err != nil {
			uc.log.ErrorContext(ctx, "UpdatePassword failed", logger.Error(err))

			return err
		}
		if user == nil {
			return models.ErrUserNotFound
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
			uc.log.ErrorContext(ctx, "UpdatePassword failed", logger.Error(err))

			return models.ErrInvalidCredentials
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			uc.log.ErrorContext(ctx, "UpdatePassword failed", logger.Error(err))

			return err
		}

		return uc.userRepo.UpdatePassword(ctx, userID, string(hash))
	})
}
