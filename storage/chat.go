package storage

import (
	"app/models"
	"context"
	"errors"

	"github.com/google/uuid"
)

const (
	defaultChatLimit = 20
	maxChatLimit     = 50
)

const sqlListUserChats = `
	select distinct on (c.updated_at, c.id)
		c.*,
		m.id, m.content, m.type, m.created_at
	from chat c
	join user_in_chat u on u.chat_id = c.id
	left join message m on m.chat_id = c.id and m.deleted_at is null
	where u.user_id = $1
		and u.deleted_at is null
		and c.deleted_at is null
		and ($4 = '' or c.type = $4)
	order by c.updated_at desc, c.id, m.created_at desc nulls last, m.id desc
	limit $2 offset $3
`

func (db *DB) ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.Chat, error) {
	if limit == 0 {
		limit = defaultChatLimit
	}

	if limit < 1 || limit > maxChatLimit {
		return nil, errors.New("limit must be between 1 and 50")
	}

	if offset < 0 {
		return nil, errors.New("offset must be zero or greater")
	}
	if chatType != "" && chatType != "dialog" && chatType != "group" && chatType != "channel" {
		return nil, errors.New("type must be dialog, group, or channel")
	}

	rows, err := db.conn.Query(ctx, sqlListUserChats, userID, limit, offset, chatType)
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

		if chat.LastMessage.ID == nil {
			chat.LastMessage = nil
		}

		chats = append(chats, &chat)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return chats, nil
}

