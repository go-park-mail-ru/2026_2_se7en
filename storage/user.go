package storage

import (
	"app/models"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	sqlCreateUser = `
		insert into app_user (email, password_hash, phone_number)
		values ($1, $2, $3)
		returning (id, email, password_hash, phone_number, created_at, updated_at, deleted_at)
	`

	sqlFindUserByEmail = `
		select * 
		from app_user
		where email = $1 
			and deleted_at is null 
	`

	sqlFindUserByID = `
		select * 
		from app_user
		where id = $1 
			and deleted_at is null 
	`
)

func (db *DB) CreateUser(ctx context.Context, email, passwordHash string, phone *string) (*models.User, error) {
	var user models.User
	err := db.conn.QueryRow(ctx, sqlCreateUser, email, passwordHash, phone).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := db.conn.QueryRow(ctx, sqlFindUserByEmail, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) FindUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := db.conn.QueryRow(ctx, sqlFindUserByID, id).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
