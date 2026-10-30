package user

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) Repository {
	return &userRepository{pool: pool}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
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
	const query = `select exists (select 1 from app_user where email = $1 and deleted_at is null)`
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

func (r *userRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const query = `delete from app_user where id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}
