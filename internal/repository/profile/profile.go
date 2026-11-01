package profile

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewProfileRepository(pool *pgxpool.Pool) *ProfileRepositoryImpl {
	return &ProfileRepositoryImpl{pool: pool}
}

func (r *ProfileRepositoryImpl) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	const query = `select id, icon_id, user_id, nickname, first_name, last_name, bio, created_at, updated_at from profile where user_id = $1`
	var db ProfileDB
	err := r.pool.QueryRow(ctx, query, userID).Scan(&db.ID, &db.IconID, &db.UserID, &db.Nickname, &db.FirstName, &db.LastName, &db.Bio, &db.CreatedAt, &db.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ProfileDBToModel(&db), nil
}

func (r *ProfileRepositoryImpl) CheckNicknameExist(ctx context.Context, nickname string) (bool, error) {
	const query = `select exists (select 1 from profile where nickname = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, nickname).Scan(&exists)
	return exists, err
}

func (r *ProfileRepositoryImpl) CreateProfile(ctx context.Context, iconID *uuid.UUID, userID uuid.UUID, nickname, firstName string, lastName *string, bio *string) (*models.Profile, error) {
	const query = `insert into profile (icon_id, user_id, nickname, first_name, last_name, bio) values ($1, $2, $3, $4, $5, $6) returning id, icon_id, user_id, nickname, first_name, last_name, bio, created_at, updated_at`
	var db ProfileDB
	err := r.pool.QueryRow(ctx, query, iconID, userID, nickname, firstName, lastName, bio).Scan(&db.ID, &db.IconID, &db.UserID, &db.Nickname, &db.FirstName, &db.LastName, &db.Bio, &db.CreatedAt, &db.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ProfileDBToModel(&db), nil
}
