package user

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/repository/profile"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/repository/tx"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepositoryImpl struct {
	log *slog.Logger

	pool storage.DBPool
}

func NewUserRepository(pool storage.DBPool, baseLog *slog.Logger) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		log:  logger.Scope(baseLog, logger.LayerRepository, logger.ModuleUser),
		pool: pool}
}

func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "FindByEmail")

	const query = `select id, email, password_hash, phone_number, created_at, updated_at, deleted_at from app_user where email = $1 and deleted_at is null`

	var db UserDB
	err := r.pool.QueryRow(ctx, query, email).Scan(&db.ID, &db.Email, &db.PasswordHash, &db.PhoneNumber, &db.CreatedAt, &db.UpdatedAt, &db.DeletedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		r.log.ErrorContext(ctx, "database query failed", logger.Error(err))
		return nil, apperrors.WrapDatabase(err)
	}
	return UserDBToModel(&db), nil
}

func (r *UserRepositoryImpl) FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "FindByID", slog.Any(logger.KeyTargetID, userID))

	const query = `select id, email, password_hash, phone_number, created_at, updated_at, deleted_at from app_user where id = $1 and deleted_at is null`

	var db UserDB
	executor := r.getExecutor(ctx)
	err := executor.QueryRow(ctx, query, userID).Scan(&db.ID, &db.Email, &db.PasswordHash, &db.PhoneNumber, &db.CreatedAt, &db.UpdatedAt, &db.DeletedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		r.log.ErrorContext(ctx, "FindByID failed", logger.Error(err))

		return nil, apperrors.WrapDatabase(err)
	}

	return UserDBToModel(&db), nil
}

func (r *UserRepositoryImpl) CheckEmailExist(ctx context.Context, email string) (bool, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "CheckEmailExist")

	const query = `select exists (select 1 from app_user where email = $1 and deleted_at is null)`

	var exists bool
	executor := r.getExecutor(ctx)
	err := executor.QueryRow(ctx, query, email).Scan(&exists)

	return exists, apperrors.WrapDatabase(err)
}

func (r *UserRepositoryImpl) CreateUser(ctx context.Context, email, passwordHash string, phone *string) (*models.User, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "CreateUser")

	const query = `insert into app_user (email, password_hash, phone_number) values ($1, $2, $3) returning id, email, password_hash, phone_number, created_at, updated_at, deleted_at`

	var db UserDB
	err := r.pool.QueryRow(ctx, query, email, passwordHash, phone).Scan(&db.ID, &db.Email, &db.PasswordHash, &db.PhoneNumber, &db.CreatedAt, &db.UpdatedAt, &db.DeletedAt)

	if err != nil {
		r.log.ErrorContext(ctx, "database query failed", logger.Error(err))
		return nil, apperrors.WrapDatabase(err)
	}
	return UserDBToModel(&db), nil
}

func (r *UserRepositoryImpl) CreateUserProfileAndSession(ctx context.Context, user *models.User, passwordHash string, expiresAt time.Time) (*models.UserWithSession, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "CreateUserProfileAndSession")

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.log.ErrorContext(ctx, "CreateUserProfileAndSession failed", logger.Error(err))

		return nil, apperrors.WrapDatabase(err)
	}
	defer tx.Rollback(ctx)

	const createUser = `insert into app_user (email, password_hash, phone_number) values ($1, $2, $3) returning id, email, password_hash, phone_number, created_at, updated_at, deleted_at`
	var userDB UserDB
	err = tx.QueryRow(ctx, createUser, user.Email, passwordHash, user.PhoneNumber).Scan(&userDB.ID, &userDB.Email, &userDB.PasswordHash, &userDB.PhoneNumber, &userDB.CreatedAt, &userDB.UpdatedAt, &userDB.DeletedAt)
	if err != nil {
		r.log.ErrorContext(ctx, "CreateUserProfileAndSession failed", logger.Error(err))

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && (pgErr.ConstraintName == "uq_app_user_email_active_lower" || pgErr.ConstraintName == "uq_app_user_email_active") {
			return nil, apperrors.ErrEmailTaken
		}
		return nil, apperrors.WrapDatabase(err)
	}

	const createProfile = `insert into profile (icon_id, user_id, nickname, first_name, last_name, bio) values ($1, $2, $3, $4, $5, $6) returning id, icon_id, user_id, nickname, first_name, last_name, bio, created_at, updated_at`
	var profileDB profile.ProfileDB
	err = tx.QueryRow(ctx, createProfile, user.Profile.IconID, userDB.ID, user.Profile.Nickname, user.Profile.FirstName, user.Profile.LastName, user.Profile.Bio).Scan(&profileDB.ID, &profileDB.IconID, &profileDB.UserID, &profileDB.Nickname, &profileDB.FirstName, &profileDB.LastName, &profileDB.Bio, &profileDB.CreatedAt, &profileDB.UpdatedAt)
	if err != nil {
		r.log.ErrorContext(ctx, "CreateUserProfileAndSession failed", logger.Error(err))

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "uq_profile_nickname" {
			return nil, apperrors.ErrNicknameTaken
		}
		return nil, apperrors.WrapDatabase(err)
	}

	const createSession = `insert into session (user_id, expires_at) values ($1, $2) returning id`
	var sessionID uuid.UUID
	if err := tx.QueryRow(ctx, createSession, userDB.ID, expiresAt).Scan(&sessionID); err != nil {
		r.log.ErrorContext(ctx, "CreateUserProfileAndSession failed", logger.Error(err))

		return nil, apperrors.WrapDatabase(err)
	}

	if err := tx.Commit(ctx); err != nil {
		r.log.ErrorContext(ctx, "CreateUserProfileAndSession failed", logger.Error(err))

		return nil, apperrors.WrapDatabase(err)
	}
	modelUser := UserDBToModel(&userDB)
	modelUser.Profile = profile.ProfileDBToModel(&profileDB)
	return &models.UserWithSession{User: modelUser, SessionID: sessionID}, nil
}

func (r *UserRepositoryImpl) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "DeleteUser", slog.Any(logger.KeyTargetID, userID))

	const query = `delete from app_user where id = $1`
	_, err := r.pool.Exec(ctx, query, userID)

	if err != nil {
		r.log.ErrorContext(ctx, "DatabaseOperation failed", logger.Error(err))
	}
	return apperrors.WrapDatabase(err)
}

func (r *UserRepositoryImpl) UpdateEmail(ctx context.Context, userID uuid.UUID, email string) error {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "UpdateEmail", slog.Any(logger.KeyTargetID, userID))

	const query = `update app_user set email = $2, updated_at = now() where id = $1`

	executor := r.getExecutor(ctx)

	_, err := executor.Exec(ctx, query, userID, email)
	if err != nil {
		r.log.ErrorContext(ctx, "UpdateEmail failed", logger.Error(err))

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrEmailTaken
		}

		return apperrors.WrapDatabase(err)
	}

	return nil
}

func (r *UserRepositoryImpl) getExecutor(ctx context.Context) interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
} {
	if transaction, ok := tx.FromContext(ctx); ok {
		return transaction
	}

	return r.pool
}

func (r *UserRepositoryImpl) UpdatePhone(ctx context.Context, userID uuid.UUID, phone *string) error {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "UpdatePhone", slog.Any(logger.KeyTargetID, userID))

	var phoneVal *string
	if phone != nil {
		p := strings.TrimSpace(*phone)
		if p == "" {
			phoneVal = nil
		} else {
			phoneVal = &p
		}
	}

	const query = `update app_user set phone_number = $2, updated_at = now() where id = $1`
	_, err := r.getExecutor(ctx).Exec(ctx, query, userID, phoneVal)

	if err != nil {
		r.log.ErrorContext(ctx, "DatabaseOperation failed", logger.Error(err))
	}
	return apperrors.WrapDatabase(err)
}

func (r *UserRepositoryImpl) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "UpdatePassword", slog.Any(logger.KeyTargetID, userID))

	const query = `update app_user set password_hash = $2, updated_at = now() where id = $1`

	_, err := r.getExecutor(ctx).Exec(ctx, query, userID, passwordHash)

	if err != nil {
		r.log.ErrorContext(ctx, "DatabaseOperation failed", logger.Error(err))
	}
	return apperrors.WrapDatabase(err)
}

func (r *UserRepositoryImpl) FindByNickname(ctx context.Context, nickname string) (*models.User, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "FindByNickname")

	const query = `
		SELECT au.id, au.email, au.password_hash, au.phone_number, au.created_at, au.updated_at, au.deleted_at,
		       p.id, p.icon_id, p.user_id, p.nickname, p.first_name, p.last_name, p.bio, p.created_at, p.updated_at
		FROM app_user au
		JOIN profile p ON au.id = p.user_id
		WHERE p.nickname = $1 AND au.deleted_at IS NULL
	`

	var userDB UserDB
	var profileDB profile.ProfileDB

	err := r.pool.QueryRow(ctx, query, nickname).Scan(
		&userDB.ID, &userDB.Email, &userDB.PasswordHash, &userDB.PhoneNumber, &userDB.CreatedAt, &userDB.UpdatedAt, &userDB.DeletedAt,
		&profileDB.ID, &profileDB.IconID, &profileDB.UserID, &profileDB.Nickname, &profileDB.FirstName, &profileDB.LastName, &profileDB.Bio, &profileDB.CreatedAt, &profileDB.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrUserNotFound
	}
	if err != nil {
		r.log.ErrorContext(ctx, "FindByNickname failed", logger.Error(err))

		return nil, apperrors.WrapDatabase(err)
	}

	user := UserDBToModel(&userDB)
	user.Profile = profile.ProfileDBToModel(&profileDB)

	return user, nil
}
