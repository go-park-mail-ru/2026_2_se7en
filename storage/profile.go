package storage

import (
	"app/models"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	sqlCreateProfile = `
		insert into profile (icon_id, user_id, nickname, first_name, last_name, bio)
	 	values ($1, $2, $3, $4, $5, $6)
		returning (id, icon_id, user_id, nickname, first_name, last_name, bio, created_at, updated_at)
	`

	sqlFindProfileByUserID = `
		select * 
		from profile
		where user_id = $1
	`

	sqlCheckNicknameExist = `
		select exists (
			select 1 from profile
			where nickname = $1
		)
	`
)

func (db *DB) CheckNicknameExist(ctx context.Context, nickname string) (bool, error) {
	var exists bool
	err := db.pool.QueryRow(ctx, sqlCheckNicknameExist, nickname).Scan(&exists)
	return exists, err
}

func (db *DB) CreateProfile(ctx context.Context, iconIDd *uuid.UUID, userIDd uuid.UUID, nickname, firstName, lastName string, bio *string) (*models.Profile, error) {
	var profile models.Profile
	err := db.pool.QueryRow(ctx, sqlCreateProfile, iconIDd, userIDd, nickname, firstName, lastName, bio).
		Scan(&profile.ID, &profile.IconID, &profile.UserID, &profile.Nickname, &profile.FirstName, &profile.LastName, &profile.Bio, &profile.CreatedAt, &profile.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (db *DB) FindProfileByUserID(ctx context.Context, userIDd uuid.UUID) (*models.Profile, error) {
	var profile models.Profile
	err := db.pool.QueryRow(ctx, sqlFindProfileByUserID, userIDd).
		Scan(&profile.ID, &profile.IconID, &profile.UserID, &profile.Nickname, &profile.FirstName, &profile.LastName, &profile.Bio, &profile.CreatedAt, &profile.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &profile, nil

}
