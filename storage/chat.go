package storage

import (
	apperrors "app/apperrors"
	"app/models"
	"context"

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

func (db *DB) ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.Chat, error) {
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

	var chats []*models.Chat
	for rows.Next() {
		chat := models.Chat{LastMessage: &models.MessagePreview{}}

		if err := rows.Scan(&chat.ID, &chat.IconID, &chat.Name, &chat.Description, &chat.Type,
			&chat.MembersCount, &chat.CreatedAt, &chat.UpdatedAt, &chat.DeletedAt,
			&chat.LastMessage.ID, &chat.LastMessage.Content,
			&chat.LastMessage.Type, &chat.LastMessage.CreatedAt); err != nil {
			return nil, err
		}

		chats = append(chats, &chat)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return chats, nil
}
