package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var fakeHash []byte

func init() {
	fakeHash, _ = bcrypt.GenerateFromPassword([]byte("fake"), bcrypt.DefaultCost)
}

type AuthUsecaseImpl struct {
	log *slog.Logger

	userRepo    UserRepository
	profileRepo ProfileRepository
	sessionRepo SessionRepository
}

func NewAuthUsecase(userRepo UserRepository, profileRepo ProfileRepository, sessionRepo SessionRepository, baseLog *slog.Logger) *AuthUsecaseImpl {
	return &AuthUsecaseImpl{
		log: logger.Scope(baseLog, logger.LayerUsecase, logger.ModuleAuth),

		userRepo:    userRepo,
		profileRepo: profileRepo,
		sessionRepo: sessionRepo,
	}
}

func (uc *AuthUsecaseImpl) Login(ctx context.Context, email, password string) (*models.UserWithSession, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	maskedEmail := logger.MaskEmail(email)

	uc.log.InfoContext(ctx, "attempting login", logger.KeyEmailMask, maskedEmail)

	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, apperrors.ErrUserNotFound) {
		uc.log.ErrorContext(ctx, "failed to find user", logger.Error(err))
		return nil, err
	}

	var hash []byte
	if err == nil && user != nil {
		hash = []byte(user.PasswordHash)
	} else {
		hash = fakeHash
	}

	if bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil {
		uc.log.WarnContext(ctx, "login failed: invalid password", logger.KeyEmailMask, maskedEmail)
		return nil, apperrors.ErrInvalidCredentials
	}

	if user == nil {
		uc.log.WarnContext(ctx, "login failed: user not found", logger.KeyEmailMask, maskedEmail)
		return nil, apperrors.ErrInvalidCredentials
	}

	profile, err := uc.profileRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		uc.log.ErrorContext(ctx, "failed to find profile", logger.Error(err), logger.KeyUserID, user.ID.String())
		return nil, err
	}

	user.Profile = profile

	session, err := uc.sessionRepo.CreateSession(ctx, user.ID, time.Now().Add(middleware.SessionTTL))
	if err != nil {
		uc.log.ErrorContext(ctx, "failed to create session", logger.Error(err), logger.KeyUserID, user.ID.String())
		return nil, err
	}

	uc.log.InfoContext(ctx, "login successful", logger.KeyUserID, user.ID.String())

	return &models.UserWithSession{
		User:      user,
		SessionID: session.ID,
	}, nil
}

func (uc *AuthUsecaseImpl) GetMe(ctx context.Context, userID uuid.UUID) (*models.UserWithSession, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		uc.log.ErrorContext(ctx, "GetMe failed", logger.Error(err))

		return nil, err
	}
	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}

	profile, err := uc.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		uc.log.ErrorContext(ctx, "GetMe failed", logger.Error(err))

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
