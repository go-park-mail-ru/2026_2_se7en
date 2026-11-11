package chat

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ChatRepositoryImpl struct {
	log *slog.Logger

	pool storage.DBPool
}

func NewChatRepository(pool storage.DBPool, baseLog *slog.Logger) *ChatRepositoryImpl {
	return &ChatRepositoryImpl{
		log:  logger.Scope(baseLog, logger.LayerRepository, logger.ModuleChat),
		pool: pool}
}

func (r *ChatRepositoryImpl) IsMember(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatMembership, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "IsMember", slog.Any(logger.KeyTargetID, userID), slog.Any(logger.KeyChatID, chatID))

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
	if err != nil {
		return nil, apperrors.WrapDatabase(err)
	}
	return &models.ChatMembership{IsMember: isMember}, nil
}

func (r *ChatRepositoryImpl) GetActiveMemberIDs(ctx context.Context, chatID uuid.UUID) ([]uuid.UUID, error) {
	const query = `select m.user_id from user_in_chat m join chat c on c.id = m.chat_id
		join app_user u on u.id = m.user_id and u.deleted_at is null
		where m.chat_id = $1 and m.deleted_at is null and c.deleted_at is null`

	rows, err := r.pool.Query(ctx, query, chatID)
	if err != nil {
		return nil, apperrors.WrapDatabase(err)
	}

	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, apperrors.WrapDatabase(err)
		}

		ids = append(ids, id)
	}

	return ids, apperrors.WrapDatabase(rows.Err())
}

const sqlPeerDetailsJoin = `
	left join lateral (
		select peer_member.user_id, p.first_name, p.last_name, p.nickname, peer_icon.url as icon_url
		from user_in_chat peer_member
		join app_user peer_user on peer_user.id = peer_member.user_id and peer_user.deleted_at is null
		join profile p on p.user_id = peer_member.user_id
		left join file peer_icon on peer_icon.id = p.icon_id
		where c.type = 'dialog' and peer_member.chat_id = c.id and peer_member.user_id <> $1
		order by peer_member.deleted_at nulls first, peer_member.created_at desc
		limit 1
	) peer on true
`

const sqlListUserChats = `
    with visible_chats as (
        select
            c.id, c.icon_id, chat_icon.url as chat_icon_url,
            peer.user_id as peer_id, peer.icon_url as peer_icon_url,
            peer.first_name as peer_first_name, peer.last_name as peer_last_name,
            peer.nickname as peer_nickname, coalesce(c.name, '') as name,
            c.description, c.type, c.members_count, c.created_at, c.updated_at, c.deleted_at,
            m.id as message_id, m.content as message_content, m.type as message_type,
            m.created_at as message_created_at,
            (select count(*) from message unread where unread.chat_id = c.id
                and unread.deleted_at is null and unread.is_read = false and unread.user_id <> $1) as unread_count,
            greatest(c.updated_at, coalesce(m.created_at, c.updated_at)) as activity_at
        from chat c
        left join file chat_icon on chat_icon.id = c.icon_id
        join user_in_chat u on u.chat_id = c.id
        join app_user actor on actor.id = u.user_id and actor.deleted_at is null
        ` + sqlPeerDetailsJoin + `
        left join lateral (
            select m.id, m.content, m.type, m.created_at
            from message m
            where m.chat_id = c.id and m.deleted_at is null
            order by m.created_at desc, m.id desc
            limit 1
        ) m on true
        where u.user_id = $1 and u.deleted_at is null and c.deleted_at is null
            and ($3::text = '' or c.type::text = $3::text)
    )
    select id, icon_id, chat_icon_url, peer_id, peer_icon_url,
        peer_first_name, peer_last_name, peer_nickname, name, description, type, members_count,
        created_at, updated_at, deleted_at, message_id, message_content,
        message_type, message_created_at, unread_count, activity_at
    from visible_chats
    where ($4::timestamptz is null or (activity_at, id) < ($4::timestamptz, $5::uuid))
    order by activity_at desc, id desc
    limit $2
`

const sqlFindForUser = `
	select c.id, c.icon_id, chat_icon.url as chat_icon_url,
		peer.user_id as peer_id, peer.icon_url as peer_icon_url,
		peer.first_name as peer_first_name, peer.last_name as peer_last_name,
		peer.nickname as peer_nickname, coalesce(c.name, '') as name,
		c.description, c.type, c.members_count,
		c.created_at, c.updated_at, c.deleted_at, role.type
	from chat c
	left join file chat_icon on chat_icon.id = c.icon_id
	join user_in_chat m on m.chat_id = c.id and m.user_id = $1 and m.deleted_at is null
	join app_user actor on actor.id = m.user_id and actor.deleted_at is null
	join chat_role role on role.id = m.role_id
	` + sqlPeerDetailsJoin + `
	where c.id = $2 and c.deleted_at is null
`

func (r *ChatRepositoryImpl) GetListUserChats(ctx context.Context, userID uuid.UUID, limit int, chatType string, cursorTime *time.Time, cursorID uuid.UUID) ([]*models.ChatListItem, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "GetListUserChats", slog.Any(logger.KeyTargetID, userID))

	rows, err := r.pool.Query(ctx, sqlListUserChats, userID, limit, chatType, cursorTime, cursorID)

	if err != nil {
		r.log.ErrorContext(ctx, "GetListUserChats failed", logger.Error(err))

		return nil, apperrors.WrapDatabase(err)
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
			&dbItem.Chat.ID, &dbItem.Chat.IconID, &dbItem.Chat.ChatIconURL, &dbItem.Chat.PeerID,
			&dbItem.Chat.PeerIconURL, &dbItem.Chat.PeerFirstName, &dbItem.Chat.PeerLastName,
			&dbItem.Chat.PeerNickname, &dbItem.Chat.Name, &dbItem.Chat.Description, &dbItem.Chat.Type,
			&dbItem.Chat.MembersCount, &dbItem.Chat.CreatedAt, &dbItem.Chat.UpdatedAt, &dbItem.Chat.DeletedAt,
			&msgID, &msgContent, &msgType, &msgCreatedAt, &dbItem.UnreadCount, &dbItem.ActivityAt,
		)
		if err != nil {
			r.log.ErrorContext(ctx, "GetListUserChats failed", logger.Error(err))

			return nil, apperrors.WrapDatabase(err)
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
		r.log.ErrorContext(ctx, "GetListUserChats failed", logger.Error(err))

		return nil, apperrors.WrapDatabase(err)
	}

	return items, nil
}

func scanDetail(row pgx.Row) (*models.ChatDetail, error) {
	var db ChatDetailDB
	c := &db.Chat
	err := row.Scan(
		&c.ID, &c.IconID, &c.ChatIconURL, &c.PeerID, &c.PeerIconURL,
		&c.PeerFirstName, &c.PeerLastName, &c.PeerNickname, &c.Name, &c.Description,
		&c.Type, &c.MembersCount, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &db.Role,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrChatNotFound
	}

	if err != nil {

		return nil, apperrors.WrapDatabase(err)
	}

	return ChatDetailDBToModel(&db), nil
}

func (r *ChatRepositoryImpl) GetChat(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "GetChat", slog.Any(logger.KeyTargetID, userID), slog.Any(logger.KeyChatID, chatID))

	detail, err := scanDetail(r.pool.QueryRow(ctx, sqlFindForUser, userID, chatID))
	if err != nil {
		r.log.ErrorContext(ctx, "GetChat failed", logger.Error(err))
	}
	return detail, apperrors.WrapDatabase(err)
}

func role(ctx context.Context, tx pgx.Tx, kind string) (uuid.UUID, error) {
	const query = `select id from chat_role where type = $1`
	var id uuid.UUID
	err := tx.QueryRow(ctx, query, kind).Scan(&id)
	return id, apperrors.WrapDatabase(err)
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

	return chatType, roleType, apperrors.WrapDatabase(err)
}

func (r *ChatRepositoryImpl) DeleteChat(ctx context.Context, userID, chatID uuid.UUID) error {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "DeleteChat", slog.Any(logger.KeyTargetID, userID), slog.Any(logger.KeyChatID, chatID))

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.log.ErrorContext(ctx, "DeleteChat failed", logger.Error(err))

		return apperrors.WrapDatabase(err)
	}

	defer tx.Rollback(ctx)
	typ, roleType, err := activeRole(ctx, tx, userID, chatID)
	if err != nil {
		r.log.ErrorContext(ctx, "DeleteChat failed", logger.Error(err))

		return apperrors.WrapDatabase(err)
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
		r.log.ErrorContext(ctx, "DeleteChat failed", logger.Error(err))

		return apperrors.WrapDatabase(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		r.log.ErrorContext(ctx, "Commit failed", logger.Error(err))
	}
	return apperrors.WrapDatabase(err)
}
