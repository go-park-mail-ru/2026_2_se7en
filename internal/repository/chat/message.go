package chat

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *ChatRepositoryImpl) MessageExists(ctx context.Context, chatID, messageID uuid.UUID) (bool, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "MessageExists", slog.Any(logger.KeyChatID, chatID), slog.Any(logger.KeyMessageID, messageID))

	const query = `
		select exists (
			select 1 from message
			where id = $1 and chat_id = $2 and deleted_at is null
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, messageID, chatID).Scan(&exists)
	return exists, apperrors.WrapDatabase(err)
}

func (r *ChatRepositoryImpl) GetListMessages(ctx context.Context, userID, chatID uuid.UUID, limit int, before *uuid.UUID) ([]models.Message, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "GetListMessages", slog.Any(logger.KeyTargetID, userID), slog.Any(logger.KeyChatID, chatID))

	const queryWithoutBefore = `
		select m.id, m.user_id, m.chat_id, m.reply_to_message_id, m.sticker_id,
			m.content, m.is_edited, m.is_read, m.type, m.created_at, m.updated_at
		from message m
		where m.chat_id = $1 and m.deleted_at is null
			and exists (
				select 1 from user_in_chat member
				join chat c on c.id = member.chat_id and c.deleted_at is null
				join app_user actor on actor.id = member.user_id and actor.deleted_at is null
				where member.chat_id = $1 and member.user_id = $3
					and member.deleted_at is null
			)
		order by m.created_at desc, m.id desc
		limit $2
	`
	const queryWithBefore = `
		select m.id, m.user_id, m.chat_id, m.reply_to_message_id, m.sticker_id,
			m.content, m.is_edited, m.is_read, m.type, m.created_at, m.updated_at
		from message m
		where m.chat_id = $1 and m.deleted_at is null
			and exists (
				select 1 from user_in_chat member
				join chat c on c.id = member.chat_id and c.deleted_at is null
				join app_user actor on actor.id = member.user_id and actor.deleted_at is null
				where member.chat_id = $1 and member.user_id = $4
					and member.deleted_at is null
			)
			and (m.created_at, m.id) < (
				select created_at, id from message where id = $3
			)
		order by m.created_at desc, m.id desc
		limit $2
	`
	var rows pgx.Rows
	var err error
	if before == nil {
		rows, err = r.pool.Query(ctx, queryWithoutBefore, chatID, limit, userID)
	} else {
		rows, err = r.pool.Query(ctx, queryWithBefore, chatID, limit, *before, userID)
	}

	if err != nil {
		r.log.ErrorContext(ctx, "GetListMessages failed", logger.Error(err))

		return nil, apperrors.WrapDatabase(err)
	}

	defer rows.Close()
	items := make([]models.Message, 0)
	for rows.Next() {
		var db MessageDB
		err = rows.Scan(&db.ID, &db.UserID, &db.ChatID, &db.ReplyToMessageID, &db.StickerID, &db.Content, &db.IsEdited, &db.IsRead, &db.Type, &db.CreatedAt, &db.UpdatedAt)
		if err != nil {
			r.log.ErrorContext(ctx, "GetListMessages failed", logger.Error(err))

			return nil, apperrors.WrapDatabase(err)
		}

		items = append(items, MessageDBToModel(&db))
	}

	err = rows.Err()
	if err != nil {
		r.log.ErrorContext(ctx, "ReadRows failed", logger.Error(err))
	}
	return items, apperrors.WrapDatabase(err)
}

func (r *ChatRepositoryImpl) CreateMessage(ctx context.Context, userID, chatID uuid.UUID, replyToMessageID, stickerID *uuid.UUID, content *string, messageType string) (*models.Message, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "CreateMessage", slog.Any(logger.KeyTargetID, userID), slog.Any(logger.KeyChatID, chatID))

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.log.ErrorContext(ctx, "CreateMessage failed", logger.Error(err))

		return nil, apperrors.WrapDatabase(err)
	}

	defer tx.Rollback(ctx)
	var activeChatID uuid.UUID
	const findMemberChat = `
		select c.id
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
		for share of c
	`
	err = tx.QueryRow(ctx, findMemberChat, userID, chatID).Scan(&activeChatID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrChatNotFound
	}

	if err != nil {
		r.log.ErrorContext(ctx, "CreateMessage failed", logger.Error(err))

		return nil, apperrors.WrapDatabase(err)
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
			r.log.ErrorContext(ctx, "CreateMessage failed", logger.Error(err))

			return nil, apperrors.WrapDatabase(err)
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
		r.log.ErrorContext(ctx, "CreateMessage failed", logger.Error(err))

		return nil, apperrors.WrapDatabase(err)
	}

	if err = tx.Commit(ctx); err != nil {
		r.log.ErrorContext(ctx, "CreateMessage failed", logger.Error(err))

		return nil, apperrors.WrapDatabase(err)
	}

	msg := MessageDBToModel(&msgDB)
	return &msg, nil
}

func (r *ChatRepositoryImpl) MarkMessageRead(ctx context.Context, userID, chatID, messageID uuid.UUID) (bool, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "MarkMessageRead", slog.Any(logger.KeyTargetID, userID), slog.Any(logger.KeyChatID, chatID), slog.Any(logger.KeyMessageID, messageID))

	const markRead = `
		with boundary as (
			select created_at, id
			from message
			where id = $3
				and chat_id = $2
				and user_id <> $1
				and deleted_at is null
		)
		update message m
		set is_read = true
		from boundary b
		where m.chat_id = $2
			and m.user_id <> $1
			and m.deleted_at is null
			and m.is_read = false
			and (m.created_at, m.id) <= (b.created_at, b.id)
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

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	return false, apperrors.WrapDatabase(err)
}

func (r *ChatRepositoryImpl) GetMessageAuthor(ctx context.Context, chatID, messageID uuid.UUID) (uuid.UUID, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "GetMessageAuthor", slog.Any(logger.KeyChatID, chatID), slog.Any(logger.KeyMessageID, messageID))

	const findMessage = `
		select user_id
		from message
		where id = $1 and chat_id = $2 and deleted_at is null
	`
	var authorID uuid.UUID
	err := r.pool.QueryRow(ctx, findMessage, messageID, chatID).Scan(&authorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, apperrors.ErrMessageNotFound
	}

	if err != nil {
		r.log.ErrorContext(ctx, "GetMessageAuthor failed", logger.Error(err))

		return uuid.Nil, apperrors.WrapDatabase(err)
	}

	return authorID, nil
}
