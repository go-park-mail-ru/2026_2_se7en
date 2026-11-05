package contact

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContactRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewContactRepository(pool *pgxpool.Pool) *ContactRepositoryImpl {
	return &ContactRepositoryImpl{pool: pool}
}

func (r *ContactRepositoryImpl) ListContacts(ctx context.Context, userID uuid.UUID, limit, offset int, q string) ([]models.Contact, error) {
	const query = `
		select c.contact_user_id, c.pseudonym, p.id, p.icon_id, p.user_id,
			p.nickname, p.first_name, p.last_name, p.bio, p.created_at, p.updated_at
		from user_contact c
		join app_user u on u.id = c.contact_user_id and u.deleted_at is null
		join profile p on p.user_id = u.id
		where c.user_id = $1 and (
			$4 = '' or p.nickname ilike '%' || $4 || '%'
			or p.first_name ilike '%' || $4 || '%'
			or coalesce(p.last_name, '') ilike '%' || $4 || '%'
			or coalesce(c.pseudonym, '') ilike '%' || $4 || '%'
		)
		order by p.nickname, c.contact_user_id
		limit $2 offset $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.Contact, 0)
	for rows.Next() {
		var db ContactDB
		p := &db.Profile
		err = rows.Scan(&db.UserID, &db.Pseudonym, &p.ID, &p.IconID, &p.UserID, &p.Nickname, &p.FirstName, &p.LastName, &p.Bio, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, ContactDBToModel(&db))
	}
	return items, rows.Err()
}

func (r *ContactRepositoryImpl) AddContact(ctx context.Context, userID, contactID uuid.UUID) (*models.Contact, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	const insertContact = `insert into user_contact (user_id, contact_user_id) values ($1, $2)`
	if _, err := tx.Exec(ctx, insertContact, userID, contactID); err != nil {
		return nil, err
	}
	const findContact = `
		select c.contact_user_id, c.pseudonym, p.id, p.icon_id, p.user_id,
			p.nickname, p.first_name, p.last_name, p.bio, p.created_at, p.updated_at
		from user_contact c
		join profile p on p.user_id = c.contact_user_id
		where c.user_id = $1 and c.contact_user_id = $2
	`
	var db ContactDB
	p := &db.Profile
	err = tx.QueryRow(ctx, findContact, userID, contactID).Scan(&db.UserID, &db.Pseudonym, &p.ID, &p.IconID, &p.UserID, &p.Nickname, &p.FirstName, &p.LastName, &p.Bio, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	contact := ContactDBToModel(&db)
	return &contact, nil
}

func (r *ContactRepositoryImpl) DeleteContact(ctx context.Context, userID, contactID uuid.UUID) (bool, error) {
	const query = `delete from user_contact where user_id = $1 and contact_user_id = $2`
	tag, err := r.pool.Exec(ctx, query, userID, contactID)
	return tag.RowsAffected() > 0, err
}
