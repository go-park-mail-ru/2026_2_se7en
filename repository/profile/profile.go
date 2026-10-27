package profile

import (
	"context"
	"errors"

	"app/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type profileRepository struct {
	pool *pgxpool.Pool
}

func NewProfileRepository(pool *pgxpool.Pool) Repository {
	return &profileRepository{pool: pool}
}

func (r *profileRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
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

func (r *profileRepository) CheckNicknameExist(ctx context.Context, nickname string) (bool, error) {
	const query = `select exists (select 1 from profile where nickname = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, nickname).Scan(&exists)
	return exists, err
}

func (r *profileRepository) CreateProfile(ctx context.Context, iconID *uuid.UUID, userID uuid.UUID, nickname, firstName string, lastName *string, bio *string) (*models.Profile, error) {
	const query = `insert into profile (icon_id, user_id, nickname, first_name, last_name, bio) values ($1, $2, $3, $4, $5, $6) returning id, icon_id, user_id, nickname, first_name, last_name, bio, created_at, updated_at`
	var db ProfileDB
	err := r.pool.QueryRow(ctx, query, iconID, userID, nickname, firstName, lastName, bio).Scan(&db.ID, &db.IconID, &db.UserID, &db.Nickname, &db.FirstName, &db.LastName, &db.Bio, &db.CreatedAt, &db.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ProfileDBToModel(&db), nil
}

func scanProfile(row pgx.Row) (*models.Profile, error) {
	var db ProfileDB
	err := row.Scan(&db.ID, &db.IconID, &db.UserID, &db.Nickname, &db.FirstName, &db.LastName, &db.Bio, &db.CreatedAt, &db.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ProfileDBToModel(&db), nil
}

func (r *profileRepository) FindProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	const query = `
		select p.id, p.icon_id, p.user_id, p.nickname, p.first_name,
			p.last_name, p.bio, p.created_at, p.updated_at
		from profile p
		join app_user u on u.id = p.user_id
		where p.user_id = $1 and u.deleted_at is null
	`
	return scanProfile(r.pool.QueryRow(ctx, query, userID))
}

func (r *profileRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, p models.Profile) (*models.Profile, error) {
	const query = `
		update profile
		set nickname = $2,
			first_name = $3,
			last_name = $4,
			bio = $5,
			updated_at = now()
		where user_id = $1 and exists (
			select 1 from app_user where id = $1 and deleted_at is null
		)
		returning id, icon_id, user_id, nickname, first_name, last_name, bio, created_at, updated_at`
	profile, err := scanProfile(r.pool.QueryRow(ctx, query, userID, p.Nickname, p.FirstName, p.LastName, p.Bio))
	return profile, err
}

func (r *profileRepository) FindPasswordHash(ctx context.Context, userID uuid.UUID) (string, error) {
	var hash string
	err := r.pool.QueryRow(ctx, `select password_hash from app_user where id = $1 and deleted_at is null`, userID).Scan(&hash)
	return hash, err
}

func (r *profileRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, oldHash, newHash string) (bool, error) {
	const query = `
		update app_user set password_hash = $3, updated_at = now()
		where id = $1 and password_hash = $2 and deleted_at is null
	`
	tag, err := r.pool.Exec(ctx, query, userID, oldHash, newHash)
	return tag.RowsAffected() == 1, err
}

func (r *profileRepository) UpdatePhone(ctx context.Context, userID uuid.UUID, phone *string) error {
	const query = `
		update app_user set phone_number = $2, updated_at = now()
		where id = $1 and deleted_at is null
	`
	_, err := r.pool.Exec(ctx, query, userID, phone)
	return err
}
