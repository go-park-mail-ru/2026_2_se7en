package contact

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type ContactRepositoryImpl struct {
	log *slog.Logger

	pool storage.DBPool
}

func NewContactRepository(pool storage.DBPool, baseLog *slog.Logger) *ContactRepositoryImpl {
	return &ContactRepositoryImpl{
		log:  logger.Scope(baseLog, logger.LayerRepository, logger.ModuleContact),
		pool: pool}
}

func (r *ContactRepositoryImpl) GetContactsList(ctx context.Context, userID uuid.UUID, limit, offset int, q string) ([]models.Contact, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "GetContactsList", slog.Any(logger.KeyTargetID, userID))

	const query = `
		SELECT 
			uc.user_id, uc.contact_user_id, uc.pseudonym, uc.created_at, uc.updated_at,
			p.id, p.user_id, p.icon_id, p.nickname, p.first_name, p.last_name, p.bio, p.created_at, p.updated_at,
			EXISTS(
				SELECT 1 FROM session s 
				WHERE s.user_id = uc.contact_user_id 
				AND s.expires_at > NOW()
				LIMIT 1
			) AS is_online
		FROM user_contact uc
		JOIN profile p ON uc.contact_user_id = p.user_id
		WHERE uc.user_id = $1 AND (
			$4 = '' OR p.nickname ILIKE '%' || $4 || '%'
			OR p.first_name ILIKE '%' || $4 || '%'
			OR COALESCE(p.last_name, '') ILIKE '%' || $4 || '%'
			OR COALESCE(uc.pseudonym, '') ILIKE '%' || $4 || '%'
		)
		ORDER BY p.first_name ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset, q)
	if err != nil {
		r.log.ErrorContext(ctx, "GetContactsList failed", logger.Error(err))

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
			r.log.ErrorContext(ctx, "GetContactsList failed", logger.Error(err))

			return nil, err
		}
		contacts = append(contacts, ContactDBToModel(&db))
	}

	if err = rows.Err(); err != nil {
		r.log.ErrorContext(ctx, "GetContactsList failed", logger.Error(err))

		return nil, err
	}

	return contacts, nil
}

func (r *ContactRepositoryImpl) CreateContact(ctx context.Context, userID, contactUserID uuid.UUID, pseudonym *string) error {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "CreateContact", slog.Any(logger.KeyTargetID, userID))

	const query = `
		INSERT INTO user_contact (user_id, contact_user_id, pseudonym)
		VALUES ($1, $2, $3)
	`

	_, err := r.pool.Exec(ctx, query, userID, contactUserID, pseudonym)
	if err != nil {
		r.log.ErrorContext(ctx, "CreateContact failed", logger.Error(err))

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
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "DeleteContact", slog.Any(logger.KeyTargetID, userID))

	const query = `DELETE FROM user_contact WHERE user_id = $1 AND contact_user_id = $2`

	tag, err := r.pool.Exec(ctx, query, userID, contactUserID)
	if err != nil {
		r.log.ErrorContext(ctx, "DeleteContact failed", logger.Error(err))

		return err
	}

	if tag.RowsAffected() == 0 {
		return apperrors.ErrContactNotFound
	}

	return nil
}
