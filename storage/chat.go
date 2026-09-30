package storage

import (
	"app/models"
	"context"

	"github.com/google/uuid"
)

const sqlListUserChats = `
	select *
	from chat c
	join user_in_chat u on c.id = u.chat_id
	where u.user_id = $1
		and c.deleted_at is null
		and u.deleted_at is null
	order by c.updated_at desc
`

func (db *DB) ListUserChats(ctx context.Context, userID uuid.UUID) ([]*models.Chat, error) {
	rows, err := db.conn.Query(ctx, sqlListUserChats, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*models.Chat
	for rows.Next() {
		var chat models.Chat
		if err := rows.Scan(&chat.ID, &chat.IconID, &chat.Name, &chat.Description, &chat.Type,
			&chat.MembersCount, &chat.CreatedAt, &chat.UpdatedAt, &chat.DeletedAt); err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}
	return chats, rows.Err()
}
