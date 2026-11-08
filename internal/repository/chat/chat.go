package chat

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (r *ChatRepositoryImpl) GetListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error) {
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

func (r *ChatRepositoryImpl) GetChat(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error) {
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
