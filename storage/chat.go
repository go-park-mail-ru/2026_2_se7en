package storage

import (
	"context"
	"time"

	"app/apperrors"
	"app/models"

	"github.com/google/uuid"
)

const defaultChatLimit = 20

const sqlListUserChats = `
	select
		c.*,
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

func (db *DB) ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error) {
	if limit == 0 {
		limit = defaultChatLimit
	}

	if limit < 1 {
		return nil, apperrors.ErrInvalidChatLimit
	}

	if offset < 0 {
		return nil, apperrors.ErrInvalidChatOffset
	}
	if chatType != "" && chatType != "dialog" && chatType != "group" && chatType != "channel" {
		return nil, apperrors.ErrInvalidChatType
	}

	rows, err := db.pool.Query(ctx, sqlListUserChats, userID, limit, offset, chatType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.ChatListItem
	for rows.Next() {

		item := &models.ChatListItem{
			Chat: &models.Chat{},
		}

		var msgID uuid.UUID
		var msgContent *string
		var msgType string
		var msgCreatedAt time.Time

		err := rows.Scan(
			&item.Chat.ID, &item.Chat.IconID, &item.Chat.Name, &item.Chat.Description, &item.Chat.Type,
			&item.Chat.MembersCount, &item.Chat.CreatedAt, &item.Chat.UpdatedAt, &item.Chat.DeletedAt,
			&msgID, &msgContent, &msgType, &msgCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if msgType != "" || msgContent != nil {
			item.LastMessage = &models.MessagePreview{
				ID:        msgID,
				Content:   msgContent,
				Type:      msgType,
				CreatedAt: msgCreatedAt,
			}
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
