package chat

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) *ChatRepositoryImpl {
	return &ChatRepositoryImpl{pool: pool}
}

func (r *ChatRepositoryImpl) IsMember(ctx context.Context, userID, chatID uuid.UUID) (bool, error) {
	const query = `
		select exists (
			select 1
			from user_in_chat m
			join chat c on c.id = m.chat_id and c.deleted_at is null
			join app_user u on u.id = m.user_id and u.deleted_at is null
			where m.user_id = $1 and m.chat_id = $2 and m.deleted_at is null
		)
	`
	var isMember bool
	err := r.pool.QueryRow(ctx, query, userID, chatID).Scan(&isMember)
	return isMember, err
}

const sqlListUserChats = `
	select
		c.id, c.icon_id,
		case
			when c.type = 'dialog' then coalesce(
				nullif(trim(concat_ws(' ', peer.first_name, peer.last_name)), ''),
				peer.nickname,
				c.name
			)
			else c.name
		end as name,
		c.description, c.type, c.members_count, c.created_at, c.updated_at, c.deleted_at,
		m.id, m.content, m.type, m.created_at
	from chat c
	join user_in_chat u on u.chat_id = c.id
	join app_user actor on actor.id = u.user_id and actor.deleted_at is null
	left join lateral (
		select p.first_name, p.last_name, p.nickname
		from user_in_chat peer_member
		join app_user peer_user on peer_user.id = peer_member.user_id and peer_user.deleted_at is null
		join profile p on p.user_id = peer_member.user_id
		where c.type = 'dialog'
			and peer_member.chat_id = c.id
			and peer_member.user_id <> $1
		order by peer_member.deleted_at nulls first, peer_member.created_at desc
		limit 1
	) peer on true
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
	join app_user actor on actor.id = m.user_id and actor.deleted_at is null
	join chat_role role on role.id = m.role_id
	where c.id = $2 and c.deleted_at is null
`

func (r *ChatRepositoryImpl) ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error) {
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
		return nil, apperrors.ErrChatNotFound
	}
	if err != nil {
		return nil, err
	}
	return ChatDetailDBToModel(&db), nil
}

func (r *ChatRepositoryImpl) GetChatDetail(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error) {
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
		join app_user actor on actor.id = m.user_id and actor.deleted_at is null
		join chat_role role on role.id = m.role_id
		where c.id = $2 and m.user_id = $1
			and c.deleted_at is null and m.deleted_at is null
		for update of c
	`
	err := tx.QueryRow(ctx, query, userID, chatID).Scan(&chatType, &roleType)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", apperrors.ErrChatNotFound
	}
	return chatType, roleType, err
}

func (r *ChatRepositoryImpl) GetOrCreateDialog(ctx context.Context, userID, peerID uuid.UUID) (*models.ChatDetail, models.DialogStatus, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, models.DialogExisting, err
	}
	defer tx.Rollback(ctx)
	ids := []string{userID.String(), peerID.String()}
	sort.Strings(ids)
	const lockDialog = `select pg_advisory_xact_lock(hashtextextended($1, 0))`
	_, err = tx.Exec(ctx, lockDialog, ids[0]+ids[1])
	if err != nil {
		return nil, models.DialogExisting, err
	}
	var activeUsers int
	const usersExist = `select count(*) from app_user where id in ($1, $2) and deleted_at is null`
	err = tx.QueryRow(ctx, usersExist, userID, peerID).Scan(&activeUsers)
	if err != nil {
		return nil, models.DialogExisting, err
	}
	if activeUsers != 2 {
		return nil, models.DialogExisting, apperrors.ErrUserNotFound
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
		return nil, models.DialogExisting, err
	}
	status := models.DialogExisting
	if errors.Is(err, pgx.ErrNoRows) {
		const insertDialog = `insert into chat (name, type, members_count) values ('', 'dialog', 2) returning id`
		err = tx.QueryRow(ctx, insertDialog).Scan(&chatID)
		if err != nil {
			return nil, models.DialogExisting, err
		}
		memberRole, err := role(ctx, tx, "member")
		if err != nil {
			return nil, models.DialogExisting, err
		}
		const insertMembers = `
			insert into user_in_chat(user_id, chat_id, role_id)
			values ($1, $3, $4), ($2, $3, $4)
		`
		_, err = tx.Exec(ctx, insertMembers, userID, peerID, chatID, memberRole)
		if err != nil {
			return nil, models.DialogExisting, err
		}
		status = models.DialogCreated
	} else {
		memberRole, err := role(ctx, tx, "member")
		if err != nil {
			return nil, models.DialogExisting, err
		}
		const restoreMember = `
			insert into user_in_chat(user_id, chat_id, role_id)
			select $1, $2, $3
			where not exists (
				select 1 from user_in_chat
				where user_id = $1 and chat_id = $2 and deleted_at is null
			)
		`
		commandTag, err := tx.Exec(ctx, restoreMember, userID, chatID, memberRole)
		if err != nil {
			return nil, models.DialogExisting, err
		}
		if commandTag.RowsAffected() > 0 {
			status = models.DialogRestored
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
			return nil, models.DialogExisting, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, models.DialogExisting, err
	}
	detail, err := r.GetChatDetail(ctx, userID, chatID)
	if err != nil {
		return nil, models.DialogExisting, err
	}
	return detail, status, nil
}

func (r *ChatRepositoryImpl) HideDialog(ctx context.Context, userID, chatID uuid.UUID) error {
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
		return apperrors.ErrInvalidChatOperation
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
		return apperrors.ErrChatMemberNotFound
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

func (r *ChatRepositoryImpl) CreateGroup(ctx context.Context, userID uuid.UUID, name string, description *string, members []uuid.UUID) (*models.ChatDetail, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	const creatorExists = `select exists(select 1 from app_user where id = $1 and deleted_at is null)`
	if err = tx.QueryRow(ctx, creatorExists, userID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperrors.ErrUserNotFound
	}
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
		return nil, apperrors.ErrUserNotFound
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

func (r *ChatRepositoryImpl) UpdateGroup(ctx context.Context, userID, chatID uuid.UUID, name, description *string, setDescription bool) (*models.ChatDetail, error) {
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
		return nil, apperrors.ErrChatNotFound
	}
	if roleType != "creator" && roleType != "admin" {
		return nil, apperrors.ErrChatPermissionDenied
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

func (r *ChatRepositoryImpl) DeleteChat(ctx context.Context, userID, chatID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	typ, roleType, err := activeRole(ctx, tx, userID, chatID)
	if err != nil {
		return err
	}
	switch typ {
	case "dialog":
	case "group":
		if roleType != "creator" {
			return apperrors.ErrChatPermissionDenied
		}
	default:
		return apperrors.ErrInvalidChatOperation
	}
	const deleteChat = `update chat set deleted_at = now(), updated_at = now() where id = $1 and deleted_at is null`
	_, err = tx.Exec(ctx, deleteChat, chatID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *ChatRepositoryImpl) ListMembers(ctx context.Context, userID, chatID uuid.UUID) ([]models.ChatMember, error) {
	if _, err := r.GetChatDetail(ctx, userID, chatID); err != nil {
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

func (r *ChatRepositoryImpl) AddMember(ctx context.Context, userID, chatID, newUserID uuid.UUID) (*models.ChatMember, error) {
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
		return nil, apperrors.ErrInvalidChatOperation
	}
	if roleType != "creator" && roleType != "admin" {
		return nil, apperrors.ErrChatPermissionDenied
	}
	var exists bool
	const userExists = `select exists(select 1 from app_user where id = $1 and deleted_at is null)`
	err = tx.QueryRow(ctx, userExists, newUserID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperrors.ErrUserNotFound
	}
	const memberExists = `
		select exists (
			select 1 from user_in_chat
			where user_id = $1 and chat_id = $2 and deleted_at is null
		)
	`
	err = tx.QueryRow(ctx, memberExists, newUserID, chatID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrChatMemberExists
	}
	memberRole, err := role(ctx, tx, "member")
	if err != nil {
		return nil, err
	}
	const insertMember = `
		insert into user_in_chat (user_id, chat_id, role_id)
		values ($1, $2, $3)
	`
	_, err = tx.Exec(ctx, insertMember, newUserID, chatID, memberRole)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperrors.ErrChatMemberExists
		}
		return nil, err
	}
	const updateCount = `update chat set members_count = members_count + 1, updated_at = now() where id = $1`
	_, err = tx.Exec(ctx, updateCount, chatID)
	if err != nil {
		return nil, err
	}
	const findAddedMember = `
		select m.user_id, role.type, p.id, p.icon_id, p.user_id,
			p.nickname, p.first_name, p.last_name, p.bio, p.created_at, p.updated_at
		from user_in_chat m
		join chat_role role on role.id = m.role_id
		join profile p on p.user_id = m.user_id
		where m.chat_id = $1 and m.user_id = $2 and m.deleted_at is null
	`
	var db ChatMemberDB
	p := &db.Profile
	err = tx.QueryRow(ctx, findAddedMember, chatID, newUserID).Scan(&db.UserID, &db.Role, &p.ID, &p.IconID, &p.UserID, &p.Nickname, &p.FirstName, &p.LastName, &p.Bio, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	member := ChatMemberDBToModel(&db)
	return &member, nil
}

func (r *ChatRepositoryImpl) RemoveMember(ctx context.Context, userID, chatID, targetID uuid.UUID) error {
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
		return apperrors.ErrInvalidChatOperation
	}
	var targetRole string
	const findMember = `
		select role.type from user_in_chat m
		join chat_role role on role.id = m.role_id
		where m.chat_id = $1 and m.user_id = $2 and m.deleted_at is null
	`
	err = tx.QueryRow(ctx, findMember, chatID, targetID).Scan(&targetRole)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.ErrChatMemberNotFound
	}
	if err != nil {
		return err
	}
	if targetRole == "creator" {
		return apperrors.ErrChatPermissionDenied
	}
	if targetID != userID && actorRole != "creator" && (actorRole != "admin" || targetRole == "admin") {
		return apperrors.ErrChatPermissionDenied
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

func (r *ChatRepositoryImpl) ListMessages(ctx context.Context, userID, chatID uuid.UUID, limit int, before *uuid.UUID) ([]models.Message, error) {
	if _, err := r.GetChatDetail(ctx, userID, chatID); err != nil {
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
			return nil, apperrors.ErrMessageNotFound
		}
	}
	const queryWithoutBefore = `
		select m.id, m.user_id, m.chat_id, m.reply_to_message_id, m.sticker_id,
			m.content, m.is_edited, m.is_read, m.type, m.created_at, m.updated_at
		from message m
		where m.chat_id = $1 and m.deleted_at is null
		order by m.created_at desc, m.id desc
		limit $2
	`
	const queryWithBefore = `
		select m.id, m.user_id, m.chat_id, m.reply_to_message_id, m.sticker_id,
			m.content, m.is_edited, m.is_read, m.type, m.created_at, m.updated_at
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
		err = rows.Scan(&db.ID, &db.UserID, &db.ChatID, &db.ReplyToMessageID, &db.StickerID, &db.Content, &db.IsEdited, &db.IsRead, &db.Type, &db.CreatedAt, &db.UpdatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, MessageDBToModel(&db))
	}
	return items, rows.Err()
}

func (r *ChatRepositoryImpl) CreateMessage(ctx context.Context, userID, chatID uuid.UUID, replyToMessageID, stickerID *uuid.UUID, content *string, messageType string) (*models.Message, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	const existsMemberAndChat = `
		select exists (
			select 1
			from app_user a
			join user_in_chat uc on a.id = uc.user_id
			join chat_role r on uc.role_id = r.id
			join profile p on uc.user_id = p.user_id
			join chat c on uc.chat_id = c.id
			where a.id = $1
				and a.deleted_at is null
				and c.id = $2
				and c.deleted_at is null
				and uc.deleted_at is null
		)
	`
	err = tx.QueryRow(ctx, existsMemberAndChat, userID, chatID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperrors.ErrChatNotFound
	}
	if replyToMessageID != nil {
		const validateReply = `
			select id
			from message
			where id = $1 and chat_id = $2 and deleted_at is null
			for share
		`
		var replyID uuid.UUID
		err = tx.QueryRow(ctx, validateReply, *replyToMessageID, chatID).Scan(&replyID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrMessageNotFound
		}
		if err != nil {
			return nil, err
		}
	}
	var msgDB MessageDB
	const insertMessage = `
		insert into message (user_id, chat_id, reply_to_message_id, sticker_id, content, type)
		values ($1, $2, $3, $4, $5, $6)
		returning id, user_id, chat_id, reply_to_message_id, sticker_id,
			content, is_edited, is_read, type, created_at, updated_at
	`
	err = tx.QueryRow(ctx, insertMessage, userID, chatID, replyToMessageID, stickerID, content, messageType).Scan(&msgDB.ID, &msgDB.UserID, &msgDB.ChatID, &msgDB.ReplyToMessageID, &msgDB.StickerID, &msgDB.Content, &msgDB.IsEdited, &msgDB.IsRead, &msgDB.Type, &msgDB.CreatedAt, &msgDB.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	msg := MessageDBToModel(&msgDB)
	return &msg, nil
}

func (r *ChatRepositoryImpl) MarkMessageRead(ctx context.Context, userID, chatID, messageID uuid.UUID) (bool, error) {
	const markRead = `
		update message m
		set is_read = true
		where m.id = $3
			and m.chat_id = $2
			and m.user_id <> $1
			and m.deleted_at is null
			and m.is_read = false
			and exists (
				select 1
				from user_in_chat member
				join chat c on c.id = member.chat_id and c.deleted_at is null
				join app_user actor on actor.id = member.user_id and actor.deleted_at is null
				where member.user_id = $1
					and member.chat_id = $2
					and member.deleted_at is null
			)
		returning m.id
	`
	var updatedID uuid.UUID
	err := r.pool.QueryRow(ctx, markRead, userID, chatID, messageID).Scan(&updatedID)
	if err == nil {
		return true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}

	if _, err := r.GetChatDetail(ctx, userID, chatID); err != nil {
		return false, err
	}

	const findMessage = `
		select user_id
		from message
		where id = $1 and chat_id = $2 and deleted_at is null
	`
	var authorID uuid.UUID
	err = r.pool.QueryRow(ctx, findMessage, messageID, chatID).Scan(&authorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, apperrors.ErrMessageNotFound
	}
	if err != nil {
		return false, err
	}
	if authorID == userID {
		return false, apperrors.ErrInvalidChatOperation
	}
	return false, nil
}
