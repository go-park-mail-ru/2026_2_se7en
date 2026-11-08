package chat

import (
	"context"
	"errors"
	"sort"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *ChatRepositoryImpl) CreateChat(ctx context.Context, userID, peerID uuid.UUID) (uuid.UUID, models.DialogStatus, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, models.DialogExisting, err
	}
	defer tx.Rollback(ctx)
	ids := []string{userID.String(), peerID.String()}
	sort.Strings(ids)
	const lockDialog = `select pg_advisory_xact_lock(hashtextextended($1, 0))`
	_, err = tx.Exec(ctx, lockDialog, ids[0]+ids[1])
	if err != nil {
		return uuid.Nil, models.DialogExisting, err
	}
	var activeUsers int
	const usersExist = `select count(*) from app_user where id in ($1, $2) and deleted_at is null`
	err = tx.QueryRow(ctx, usersExist, userID, peerID).Scan(&activeUsers)
	if err != nil {
		return uuid.Nil, models.DialogExisting, err
	}
	if activeUsers != 2 {
		return uuid.Nil, models.DialogExisting, apperrors.ErrUserNotFound
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
		return uuid.Nil, models.DialogExisting, err
	}
	status := models.DialogExisting
	if errors.Is(err, pgx.ErrNoRows) {
		const insertDialog = `insert into chat (name, type, members_count) values ('', 'dialog', 2) returning id`
		err = tx.QueryRow(ctx, insertDialog).Scan(&chatID)
		if err != nil {
			return uuid.Nil, models.DialogExisting, err
		}
		memberRole, err := role(ctx, tx, "member")
		if err != nil {
			return uuid.Nil, models.DialogExisting, err
		}
		const insertMembers = `
			insert into user_in_chat(user_id, chat_id, role_id)
			values ($1, $3, $4), ($2, $3, $4)
		`
		_, err = tx.Exec(ctx, insertMembers, userID, peerID, chatID, memberRole)
		if err != nil {
			return uuid.Nil, models.DialogExisting, err
		}
		status = models.DialogCreated
	} else {
		memberRole, err := role(ctx, tx, "member")
		if err != nil {
			return uuid.Nil, models.DialogExisting, err
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
			return uuid.Nil, models.DialogExisting, err
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
			return uuid.Nil, models.DialogExisting, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return uuid.Nil, models.DialogExisting, err
	}
	return chatID, status, nil
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
