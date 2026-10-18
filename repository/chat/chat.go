package chat

import (
	"app/models"
	"context"
	"time"

	"github.com/google/uuid"
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

func (r *chatRepository) ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error) {
	rows, err := r.pool.Query(ctx, sqlListUserChats, userID, limit, offset, chatType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.ChatListItem
	for rows.Next() {
		var dbItem ChatListItemDB
		var msgID uuid.UUID
		var msgContent *string
		var msgType string
		var msgCreatedAt time.Time

		err := rows.Scan(
			&dbItem.Chat.ID, &dbItem.Chat.IconID, &dbItem.Chat.Name, &dbItem.Chat.Description, &dbItem.Chat.Type,
			&dbItem.Chat.MembersCount, &dbItem.Chat.CreatedAt, &dbItem.Chat.UpdatedAt, &dbItem.Chat.DeletedAt,
			&msgID, &msgContent, &msgType, &msgCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if msgType != "" || msgContent != nil {
			dbItem.LastMessage = &MessagePreviewDB{
				ID:        msgID,
				Content:   msgContent,
				Type:      msgType,
				CreatedAt: msgCreatedAt,
			}
		}

		items = append(items, ChatListItemDBToModel(&dbItem))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
