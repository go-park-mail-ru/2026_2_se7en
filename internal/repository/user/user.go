package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/repository/profile"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/repository/tx"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepositoryImpl {
	return &UserRepositoryImpl{pool: pool}
}

func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	const query = `select id, email, password_hash, phone_number, created_at, updated_at, deleted_at from app_user where email = $1 and deleted_at is null`

	var db UserDB
	err := r.pool.QueryRow(ctx, query, email).Scan(&db.ID, &db.Email, &db.PasswordHash, &db.PhoneNumber, &db.CreatedAt, &db.UpdatedAt, &db.DeletedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return UserDBToModel(&db), nil
}

func (r *UserRepositoryImpl) FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	const query = `select id, email, password_hash, phone_number, created_at, updated_at, deleted_at from app_user where id = $1 and deleted_at is null`

	var db UserDB
	executor := r.getExecutor(ctx)
	err := executor.QueryRow(ctx, query, userID).Scan(&db.ID, &db.Email, &db.PasswordHash, &db.PhoneNumber, &db.CreatedAt, &db.UpdatedAt, &db.DeletedAt)
	
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return UserDBToModel(&db), nil
}

func (r *UserRepositoryImpl) CheckEmailExist(ctx context.Context, email string) (bool, error) {
	const query = `select exists (select 1 from app_user where email = $1 and deleted_at is null)`

	var exists bool
	executor := r.getExecutor(ctx)
	err := executor.QueryRow(ctx, query, email).Scan(&exists)

	return exists, err
}

func (r *UserRepositoryImpl) CreateUser(ctx context.Context, email, passwordHash string, phone *string) (*models.User, error) {
	const query = `insert into app_user (email, password_hash, phone_number) values ($1, $2, $3) returning id, email, password_hash, phone_number, created_at, updated_at, deleted_at`

	var db UserDB
	err := r.pool.QueryRow(ctx, query, email, passwordHash, phone).Scan(&db.ID, &db.Email, &db.PasswordHash, &db.PhoneNumber, &db.CreatedAt, &db.UpdatedAt, &db.DeletedAt)
	
	if err != nil {
		return nil, err
	}

	return UserDBToModel(&db), nil
}

func (r *UserRepositoryImpl) CreateUserProfileAndSession(ctx context.Context, user *models.User, passwordHash string, expiresAt time.Time) (*models.UserWithSession, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const createUser = `insert into app_user (email, password_hash, phone_number) values ($1, $2, $3) returning id, email, password_hash, phone_number, created_at, updated_at, deleted_at`
	var userDB UserDB
	err = tx.QueryRow(ctx, createUser, user.Email, passwordHash, user.PhoneNumber).Scan(&userDB.ID, &userDB.Email, &userDB.PasswordHash, &userDB.PhoneNumber, &userDB.CreatedAt, &userDB.UpdatedAt, &userDB.DeletedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && (pgErr.ConstraintName == "uq_app_user_email_active_lower" || pgErr.ConstraintName == "uq_app_user_email_active") {
			return nil, apperrors.ErrEmailTaken
		}
		return nil, err
	}

	const createProfile = `insert into profile (icon_id, user_id, nickname, first_name, last_name, bio) values ($1, $2, $3, $4, $5, $6) returning id, icon_id, user_id, nickname, first_name, last_name, bio, created_at, updated_at`
	var profileDB profile.ProfileDB
	err = tx.QueryRow(ctx, createProfile, user.Profile.IconID, userDB.ID, user.Profile.Nickname, user.Profile.FirstName, user.Profile.LastName, user.Profile.Bio).Scan(&profileDB.ID, &profileDB.IconID, &profileDB.UserID, &profileDB.Nickname, &profileDB.FirstName, &profileDB.LastName, &profileDB.Bio, &profileDB.CreatedAt, &profileDB.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "uq_profile_nickname" {
			return nil, apperrors.ErrNicknameTaken
		}
		return nil, err
	}

	const createSession = `insert into session (user_id, expires_at) values ($1, $2) returning id`
	var sessionID uuid.UUID
	if err := tx.QueryRow(ctx, createSession, userDB.ID, expiresAt).Scan(&sessionID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	modelUser := UserDBToModel(&userDB)
	modelUser.Profile = profile.ProfileDBToModel(&profileDB)
	return &models.UserWithSession{User: modelUser, SessionID: sessionID}, nil
}

func (r *UserRepositoryImpl) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const query = `delete from app_user where id = $1`
	_, err := r.pool.Exec(ctx, query, userID)

	return err
}

func (r *UserRepositoryImpl) UpdateEmail(ctx context.Context, userID uuid.UUID, email string) error {
	const query = `update app_user set email = $2, updated_at = now() where id = $1`

	executor := r.getExecutor(ctx)
	
	_, err := executor.Exec(ctx, query, userID, email)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrEmailTaken
		}

		return err
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
	
	return err
}

func (r *UserRepositoryImpl) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	const query = `update app_user set password_hash = $2, updated_at = now() where id = $1`
	
	_, err := r.getExecutor(ctx).Exec(ctx, query, userID, passwordHash)

	return err
}
