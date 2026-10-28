package user

import (
	"app/apperrors"
	"app/models"
	"context"
	"errors"
	"time"

	"app/repository/profile"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) Repository {
	return &userRepository{pool: pool}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	const query = `select id, email, password_hash, phone_number, created_at, updated_at, deleted_at from app_user where lower(email) = lower($1) and deleted_at is null`
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

func (r *userRepository) FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	const query = `select id, email, password_hash, phone_number, created_at, updated_at, deleted_at from app_user where id = $1 and deleted_at is null`
	var db UserDB
	err := r.pool.QueryRow(ctx, query, userID).Scan(&db.ID, &db.Email, &db.PasswordHash, &db.PhoneNumber, &db.CreatedAt, &db.UpdatedAt, &db.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return UserDBToModel(&db), nil
}

func (r *userRepository) CheckEmailExist(ctx context.Context, email string) (bool, error) {
	const query = `select exists (select 1 from app_user where lower(email) = lower($1) and deleted_at is null)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, email).Scan(&exists)
	return exists, err
}

func (r *userRepository) CreateUser(ctx context.Context, email, passwordHash string, phone *string) (*models.User, error) {
	const query = `insert into app_user (email, password_hash, phone_number) values ($1, $2, $3) returning id, email, password_hash, phone_number, created_at, updated_at, deleted_at`
	var db UserDB
	err := r.pool.QueryRow(ctx, query, email, passwordHash, phone).Scan(&db.ID, &db.Email, &db.PasswordHash, &db.PhoneNumber, &db.CreatedAt, &db.UpdatedAt, &db.DeletedAt)
	if err != nil {
		return nil, err
	}
	return UserDBToModel(&db), nil
}

func (r *userRepository) CreateUserProfileAndSession(ctx context.Context, user *models.User, passwordHash string, expiresAt time.Time) (*models.UserWithSession, error) {
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

func (r *userRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const query = `delete from app_user where id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}
