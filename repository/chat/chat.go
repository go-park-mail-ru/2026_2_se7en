package chat

import (
	"context"
	"errors"
	"sort"
	"time"

	"app/apperrors"
	"app/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type chatRepository struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) Repository {
	return &chatRepository{pool: pool}
}

const sqlListUserChats = `
	select
		c.id, c.icon_id, c.name, c.description, c.type, c.members_count, c.created_at, c.updated_at, c.deleted_at,
		m.id, m.content, m.type, m.created_at
	from chat c
	join user_in_chat u on u.chat_id = c.id
	left join lateral (
		select m.id, m.content, m.type, m.created_at
		from message m
		where m.chat_id = c.id and m.deleted_at is null
		order by m.created_at desc, m.id desc
		limit 1
	) m on true
	where u.user_id = $1
		and u.deleted_at is null
		and c.deleted_at is null
		and ($4 = '' or c.type = $4)
	order by m.created_at desc nulls last, c.updated_at desc, c.id
	limit $2 offset $3
`

const sqlFindForUser = `
	select c.id, c.icon_id, c.name, c.description, c.type, c.members_count,
		c.created_at, c.updated_at, c.deleted_at, role.type
	from chat c
	join user_in_chat m on m.chat_id = c.id and m.user_id = $1 and m.deleted_at is null
	join chat_role role on role.id = m.role_id
	where c.id = $2 and c.deleted_at is null
`

func (r *chatRepository) ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error) {
	rows, err := r.pool.Query(ctx, sqlListUserChats, userID, limit, offset, chatType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.ChatListItem
	for rows.Next() {
		var dbItem ChatListItemDB
		var msgID *uuid.UUID
		var msgContent *string
		var msgType *string
		var msgCreatedAt *time.Time

		err := rows.Scan(
			&dbItem.Chat.ID, &dbItem.Chat.IconID, &dbItem.Chat.Name, &dbItem.Chat.Description, &dbItem.Chat.Type,
			&dbItem.Chat.MembersCount, &dbItem.Chat.CreatedAt, &dbItem.Chat.UpdatedAt, &dbItem.Chat.DeletedAt,
			&msgID, &msgContent, &msgType, &msgCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if msgID != nil && msgType != nil && msgCreatedAt != nil {
			dbItem.LastMessage = &MessagePreviewDB{
				ID:        *msgID,
				Content:   msgContent,
				Type:      *msgType,
				CreatedAt: *msgCreatedAt,
			}
		}

		items = append(items, ChatListItemDBToModel(&dbItem))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func scanDetail(row pgx.Row) (*models.ChatDetail, error) {
	var db ChatDetailDB
	c := &db.Chat
	err := row.Scan(&c.ID, &c.IconID, &c.Name, &c.Description, &c.Type, &c.MembersCount, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &db.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.NewNotFound("Resource not found")
	}
	if err != nil {
		return nil, err
	}
	return ChatDetailDBToModel(&db), nil
}

func (r *chatRepository) FindForUser(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error) {
	return scanDetail(r.pool.QueryRow(ctx, sqlFindForUser, userID, chatID))
}

func role(ctx context.Context, tx pgx.Tx, kind string) (uuid.UUID, error) {
	const query = `select id from chat_role where type = $1`
	var id uuid.UUID
	err := tx.QueryRow(ctx, query, kind).Scan(&id)
	return id, err
}

func activeRole(ctx context.Context, tx pgx.Tx, userID, chatID uuid.UUID) (string, string, error) {
	var chatType, roleType string
	const query = `
		select c.type, role.type
		from chat c
		join user_in_chat m on m.chat_id = c.id
		join chat_role role on role.id = m.role_id
		where c.id = $2 and m.user_id = $1
			and c.deleted_at is null and m.deleted_at is null
		for update of c
	`
	err := tx.QueryRow(ctx, query, userID, chatID).Scan(&chatType, &roleType)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", apperrors.NewNotFound("Resource not found")
	}
	return chatType, roleType, err
}

func (r *chatRepository) GetOrCreateDialog(ctx context.Context, userID, peerID uuid.UUID) (*models.ChatDetail, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	ids := []string{userID.String(), peerID.String()}
	sort.Strings(ids)
	const lockDialog = `select pg_advisory_xact_lock(hashtextextended($1, 0))`
	_, err = tx.Exec(ctx, lockDialog, ids[0]+ids[1])
	if err != nil {
		return nil, err
	}
	var exists bool
	const peerExists = `select exists(select 1 from app_user where id = $1 and deleted_at is null)`
	err = tx.QueryRow(ctx, peerExists, peerID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperrors.NewNotFound("Resource not found")
	}
	var chatID uuid.UUID
	const findDialog = `
		select c.id
		from chat c
		join user_in_chat a on a.chat_id = c.id and a.user_id = $1
		join user_in_chat b on b.chat_id = c.id and b.user_id = $2
		where c.type = 'dialog' and c.deleted_at is null
		order by c.created_at
		limit 1
	`
	err = tx.QueryRow(ctx, findDialog, userID, peerID).Scan(&chatID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		const insertDialog = `insert into chat (name, type, members_count) values ('', 'dialog', 2) returning id`
		err = tx.QueryRow(ctx, insertDialog).Scan(&chatID)
		if err != nil {
			return nil, err
		}
		memberRole, err := role(ctx, tx, "member")
		if err != nil {
			return nil, err
		}
		const insertMembers = `
			insert into user_in_chat(user_id, chat_id, role_id)
			values ($1, $3, $4), ($2, $3, $4)
		`
		_, err = tx.Exec(ctx, insertMembers, userID, peerID, chatID, memberRole)
		if err != nil {
			return nil, err
		}
	} else {
		memberRole, err := role(ctx, tx, "member")
		if err != nil {
			return nil, err
		}
		const restoreMember = `
			insert into user_in_chat(user_id, chat_id, role_id)
			select $1, $2, $3
			where not exists (
				select 1 from user_in_chat
				where user_id = $1 and chat_id = $2 and deleted_at is null
			)
		`
		_, err = tx.Exec(ctx, restoreMember, userID, chatID, memberRole)
		if err != nil {
			return nil, err
		}
		const updateCount = `
			update chat
			set members_count = (
				select count(*) from user_in_chat
				where chat_id = $1 and deleted_at is null
			), updated_at = now()
			where id = $1
		`
		_, err = tx.Exec(ctx, updateCount, chatID)
		if err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindForUser(ctx, userID, chatID)
}

func (r *chatRepository) HideDialog(ctx context.Context, userID, chatID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	chatType, _, err := activeRole(ctx, tx, userID, chatID)
	if err != nil {
		return err
	}
	if chatType != "dialog" {
		return apperrors.NewBadRequest("Validation failed", nil)
	}
	const hideMember = `
		update user_in_chat
		set deleted_at = now(), updated_at = now()
		where user_id = $1 and chat_id = $2 and deleted_at is null
	`
	tag, err := tx.Exec(ctx, hideMember, userID, chatID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperrors.NewNotFound("Resource not found")
	}
	const updateCount = `
		update chat
		set members_count = (
			select count(*) from user_in_chat
			where chat_id = $1 and deleted_at is null
		), updated_at = now()
		where id = $1
	`
	_, err = tx.Exec(ctx, updateCount, chatID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *chatRepository) CreateGroup(ctx context.Context, userID uuid.UUID, name string, description *string, members []uuid.UUID) (*models.ChatDetail, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var chatID uuid.UUID
	const insertGroup = `
		insert into chat (name, description, type, members_count)
		values ($1, $2, 'group', $3)
		returning id
	`
	err = tx.QueryRow(ctx, insertGroup, name, description, len(members)+1).Scan(&chatID)
	if err != nil {
		return nil, err
	}
	creatorRole, err := role(ctx, tx, "creator")
	if err != nil {
		return nil, err
	}
	memberRole, err := role(ctx, tx, "member")
	if err != nil {
		return nil, err
	}
	const insertCreator = `
		insert into user_in_chat (user_id, chat_id, role_id)
		values ($1, $2, $3)
	`
	_, err = tx.Exec(ctx, insertCreator, userID, chatID, creatorRole)
	if err != nil {
		return nil, err
	}
	const insertMembers = `
		insert into user_in_chat (user_id, chat_id, role_id)
		select u.id, $2, $3
		from unnest($1::uuid[]) as requested(user_id)
		join app_user u on u.id = requested.user_id and u.deleted_at is null
	`
	tag, err := tx.Exec(ctx, insertMembers, members, chatID, memberRole)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() != int64(len(members)) {
		return nil, apperrors.NewNotFound("Resource not found")
	}
	d, err := scanDetail(tx.QueryRow(ctx, sqlFindForUser, userID, chatID))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return d, nil
}

func (r *chatRepository) UpdateGroup(ctx context.Context, userID, chatID uuid.UUID, name, description *string, setDescription bool) (*models.ChatDetail, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	typ, roleType, err := activeRole(ctx, tx, userID, chatID)
	if err != nil {
		return nil, err
	}
	if typ != "group" {
		return nil, apperrors.NewNotFound("Resource not found")
	}
	if roleType != "creator" && roleType != "admin" {
		return nil, apperrors.NewForbidden("No permissions")
	}
	const updateGroup = `
		update chat
		set name = coalesce($2, name),
			description = case when $3 then $4 else description end,
			updated_at = now()
		where id = $1
	`
	_, err = tx.Exec(ctx, updateGroup, chatID, name, setDescription, description)
	if err != nil {
		return nil, err
	}
	d, err := scanDetail(tx.QueryRow(ctx, sqlFindForUser, userID, chatID))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return d, nil
}

func (r *chatRepository) DeleteGroup(ctx context.Context, userID, chatID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	typ, roleType, err := activeRole(ctx, tx, userID, chatID)
	if err != nil {
		return err
	}
	if typ != "group" {
		return apperrors.NewNotFound("Resource not found")
	}
	if roleType != "creator" {
		return apperrors.NewForbidden("No permissions")
	}
	const deleteGroup = `update chat set deleted_at = now(), updated_at = now() where id = $1`
	_, err = tx.Exec(ctx, deleteGroup, chatID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *chatRepository) ListMembers(ctx context.Context, userID, chatID uuid.UUID) ([]models.ChatMember, error) {
	if _, err := r.FindForUser(ctx, userID, chatID); err != nil {
		return nil, err
	}
	const query = `
		select m.user_id, role.type, p.id, p.icon_id, p.user_id,
			p.nickname, p.first_name, p.last_name, p.bio, p.created_at, p.updated_at
		from user_in_chat m
		join chat_role role on role.id = m.role_id
		join app_user u on u.id = m.user_id and u.deleted_at is null
		join profile p on p.user_id = m.user_id
		where m.chat_id = $1 and m.deleted_at is null
		order by role.type, p.nickname
	`
	rows, err := r.pool.Query(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.ChatMember, 0)
	for rows.Next() {
		var db ChatMemberDB
		p := &db.Profile
		err = rows.Scan(&db.UserID, &db.Role, &p.ID, &p.IconID, &p.UserID, &p.Nickname, &p.FirstName, &p.LastName, &p.Bio, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, ChatMemberDBToModel(&db))
	}
	return items, rows.Err()
}

func (r *chatRepository) AddMember(ctx context.Context, userID, chatID, newUserID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	typ, roleType, err := activeRole(ctx, tx, userID, chatID)
	if err != nil {
		return err
	}
	if typ != "group" {
		return apperrors.NewBadRequest("Validation failed", nil)
	}
	if roleType != "creator" && roleType != "admin" {
		return apperrors.NewForbidden("No permissions")
	}
	var exists bool
	const userExists = `select exists(select 1 from app_user where id = $1 and deleted_at is null)`
	err = tx.QueryRow(ctx, userExists, newUserID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return apperrors.NewNotFound("Resource not found")
	}
	const memberExists = `
		select exists (
			select 1 from user_in_chat
			where user_id = $1 and chat_id = $2 and deleted_at is null
		)
	`
	err = tx.QueryRow(ctx, memberExists, newUserID, chatID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return apperrors.NewConflict("Resource already exists", nil)
	}
	memberRole, err := role(ctx, tx, "member")
	if err != nil {
		return err
	}
	const insertMember = `
		insert into user_in_chat (user_id, chat_id, role_id)
		values ($1, $2, $3)
	`
	_, err = tx.Exec(ctx, insertMember, newUserID, chatID, memberRole)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.NewConflict("Resource already exists", nil)
		}
		return err
	}
	const updateCount = `update chat set members_count = members_count + 1, updated_at = now() where id = $1`
	_, err = tx.Exec(ctx, updateCount, chatID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *chatRepository) RemoveMember(ctx context.Context, userID, chatID, targetID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	typ, actorRole, err := activeRole(ctx, tx, userID, chatID)
	if err != nil {
		return err
	}
	if typ != "group" {
		return apperrors.NewBadRequest("Validation failed", nil)
	}
	var targetRole string
	const findMember = `
		select role.type from user_in_chat m
		join chat_role role on role.id = m.role_id
		where m.chat_id = $1 and m.user_id = $2 and m.deleted_at is null
	`
	err = tx.QueryRow(ctx, findMember, chatID, targetID).Scan(&targetRole)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.NewNotFound("Resource not found")
	}
	if err != nil {
		return err
	}
	if targetRole == "creator" {
		return apperrors.NewForbidden("No permissions")
	}
	if targetID != userID && actorRole != "creator" && (actorRole != "admin" || targetRole == "admin") {
		return apperrors.NewForbidden("No permissions")
	}
	const removeMember = `
		update user_in_chat set deleted_at = now(), updated_at = now()
		where chat_id = $1 and user_id = $2 and deleted_at is null
	`
	_, err = tx.Exec(ctx, removeMember, chatID, targetID)
	if err != nil {
		return err
	}
	const updateCount = `update chat set members_count = members_count - 1, updated_at = now() where id = $1`
	_, err = tx.Exec(ctx, updateCount, chatID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *chatRepository) ListMessages(ctx context.Context, userID, chatID uuid.UUID, limit int, before *uuid.UUID) ([]models.Message, error) {
	if _, err := r.FindForUser(ctx, userID, chatID); err != nil {
		return nil, err
	}
	if before != nil {
		var exists bool
		const query = `
			select exists (
				select 1 from message
				where id = $1 and chat_id = $2 and deleted_at is null
			)
		`
		err := r.pool.QueryRow(ctx, query, before, chatID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, apperrors.NewBadRequest("Validation failed", nil)
		}
	}
	const queryWithoutBefore = `
		select m.id, m.user_id, m.chat_id, m.reply_to_message_id, m.sticker_id,
			m.content, m.is_edited, m.type, m.created_at, m.updated_at
		from message m
		where m.chat_id = $1 and m.deleted_at is null
		order by m.created_at desc, m.id desc
		limit $2
	`
	const queryWithBefore = `
		select m.id, m.user_id, m.chat_id, m.reply_to_message_id, m.sticker_id,
			m.content, m.is_edited, m.type, m.created_at, m.updated_at
		from message m
		where m.chat_id = $1 and m.deleted_at is null
			and (m.created_at, m.id) < (
				select created_at, id from message where id = $3
			)
		order by m.created_at desc, m.id desc
		limit $2
	`
	var rows pgx.Rows
	var err error
	if before == nil {
		rows, err = r.pool.Query(ctx, queryWithoutBefore, chatID, limit)
	} else {
		rows, err = r.pool.Query(ctx, queryWithBefore, chatID, limit, *before)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.Message, 0)
	for rows.Next() {
		var db MessageDB
		err = rows.Scan(&db.ID, &db.UserID, &db.ChatID, &db.ReplyToMessageID, &db.StickerID, &db.Content, &db.IsEdited, &db.Type, &db.CreatedAt, &db.UpdatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, MessageDBToModel(&db))
	}
	return items, rows.Err()
}
