package profile

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
	txManager   TxManager
	userRepo    UserRepository
	profileRepo ProfileRepository
	fileRepo    FileRepository
}

func NewProfileUsecase(txManager TxManager, userRepo UserRepository, profileRepo ProfileRepository, fileRepo FileRepository) *ProfileUsecaseImpl {
	return &ProfileUsecaseImpl{
		txManager:   txManager,
		userRepo:    userRepo,
		profileRepo: profileRepo,
		fileRepo:    fileRepo,
	}
}

func (uc *ProfileUsecaseImpl) GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, *models.Profile, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	
	if user == nil {
		return nil, nil, apperrors.ErrUserNotFound
	}

	prof, err := uc.profileRepo.GetWithFileByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	return user, prof, nil
}

func (uc *ProfileUsecaseImpl) UpdateProfile(ctx context.Context, userID uuid.UUID, params UpdateParams) error {
	return uc.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		user, err := uc.userRepo.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		if user == nil {
			return apperrors.ErrUserNotFound
		}

		prof, err := uc.profileRepo.GetWithFileByUserID(ctx, userID)
		if err != nil {
			return err
		}
		if prof == nil {
			return apperrors.NewNotFound("Profile not found")
		}

		if params.Email != nil && *params.Email != user.Email {
			if err := uc.userRepo.UpdateEmail(ctx, userID, *params.Email); err != nil {
				return err
			}
		}

		if params.PhoneNumber != nil {
			if err := uc.userRepo.UpdatePhone(ctx, userID, params.PhoneNumber); err != nil {
				return err
			}
		}

		nickname := prof.Nickname
		firstName := prof.FirstName
		lastName := prof.LastName
		bio := prof.Bio

		if params.Nickname != nil && *params.Nickname != nickname {
			exists, err := uc.profileRepo.CheckNicknameExist(ctx, *params.Nickname)
			if err != nil {
				return err
			}
			if exists {
				return apperrors.ErrNicknameTaken
			}
			nickname = *params.Nickname
		}

		if params.FirstName != nil {
			firstName = *params.FirstName
		}
		if params.LastName != nil {
			lastName = params.LastName
		}
		if params.Bio != nil {
			bio = params.Bio
		}

		return uc.profileRepo.Update(ctx, prof.ID, nickname, firstName, lastName, bio)
	})
}

func (uc *ProfileUsecaseImpl) UploadAvatar(ctx context.Context, userID uuid.UUID, file *models.FileUpload) (string, error) {
	url, filename, err := tools.SaveAvatar(file)
	if err != nil {
		return "", apperrors.NewBadRequest("Failed to save avatar", nil)
	}

	var fileURL string
	err = uc.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		prof, err := uc.profileRepo.GetWithFileByUserID(ctx, userID)
		if err != nil {
			return err
		}
		if prof == nil {
			return apperrors.NewNotFound("Profile not found")
		}

		fileModel, err := uc.fileRepo.Create(ctx, url, filename, "image")
		if err != nil {
			return err
		}

		if err := uc.profileRepo.UpdateIcon(ctx, prof.ID, fileModel.ID); err != nil {
			return err
		}

		fileURL = fileModel.URL
		return nil
	})

	if err != nil {
		return "", err
	}

	return fileURL, nil
}

func (uc *ProfileUsecaseImpl) UpdatePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	return uc.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		user, err := uc.userRepo.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		if user == nil {
			return apperrors.ErrUserNotFound
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
			return apperrors.ErrInvalidCredentials
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		return uc.userRepo.UpdatePassword(ctx, userID, string(hash))
	})
}
