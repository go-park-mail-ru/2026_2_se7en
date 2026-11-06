package contact

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type ContactRepositoryImpl struct {
	pool storage.DBPool
}

func NewContactRepository(pool storage.DBPool) *ContactRepositoryImpl {
	return &ContactRepositoryImpl{pool: pool}
}

func (r *ContactRepositoryImpl) GetContactsList(ctx context.Context, userID uuid.UUID) ([]models.Contact, error) {
	const query = `
		SELECT uc.user_id, uc.contact_user_id, uc.pseudonym, uc.created_at, uc.updated_at,
		       p.id, p.user_id, p.icon_id, p.nickname, p.first_name, p.last_name, p.bio, p.created_at, p.updated_at,
		       (s.id IS NOT NULL) AS is_online
		FROM user_contact uc
		JOIN profile p ON uc.contact_user_id = p.user_id
		LEFT JOIN session s ON uc.contact_user_id = s.user_id AND s.expires_at > NOW()
		WHERE uc.user_id = $1
		ORDER BY p.first_name ASC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.Contact
	for rows.Next() {
		var db ContactDB
		err := rows.Scan(
			&db.UserID, &db.ContactUserID, &db.Pseudonym, &db.CreatedAt, &db.UpdatedAt,
			&db.Profile.ID, &db.Profile.UserID, &db.Profile.IconID, &db.Profile.Nickname,
			&db.Profile.FirstName, &db.Profile.LastName, &db.Profile.Bio, &db.Profile.CreatedAt, &db.Profile.UpdatedAt,
			&db.IsOnline,
		)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, ContactDBToModel(&db))
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return contacts, nil
}

func (r *ContactRepositoryImpl) CreateContact(ctx context.Context, userID, contactUserID uuid.UUID, pseudonym *string) error {
	const query = `
		INSERT INTO user_contact (user_id, contact_user_id, pseudonym)
		VALUES ($1, $2, $3)
	`

	_, err := r.pool.Exec(ctx, query, userID, contactUserID, pseudonym)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return apperrors.ErrContactAlreadyExists
			}
			if pgErr.Code == "23514" {
				return apperrors.ErrCannotAddSelfAsContact
			}
		}
		return err
	}

	return nil
}

func (r *ContactRepositoryImpl) DeleteContact(ctx context.Context, userID, contactUserID uuid.UUID) error {
	const query = `DELETE FROM user_contact WHERE user_id = $1 AND contact_user_id = $2`

	tag, err := r.pool.Exec(ctx, query, userID, contactUserID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return apperrors.ErrContactNotFound
	}

	return nil
}

func (r *ContactRepositoryImpl) CheckContactExists(ctx context.Context, userID, contactUserID uuid.UUID) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM user_contact WHERE user_id = $1 AND contact_user_id = $2)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, contactUserID).Scan(&exists)
	return exists, err
}
