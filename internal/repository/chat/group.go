package chat

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *ChatRepositoryImpl) CreateGroup(ctx context.Context, userID uuid.UUID, name string, description *string, members []uuid.UUID) (*models.ChatDetail, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "CreateGroup", slog.Any(logger.KeyTargetID, userID))
  
  if len(members)+1 > 100 {
		return nil, apperrors.ErrChatMemberLimit
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.log.ErrorContext(ctx, "CreateGroup failed", logger.Error(err))

		return nil, err
	}

	defer tx.Rollback(ctx)
	var exists bool
	const creatorExists = `select exists(select 1 from app_user where id = $1 and deleted_at is null)`
	if err = tx.QueryRow(ctx, creatorExists, userID).Scan(&exists); err != nil {
		r.log.ErrorContext(ctx, "CreateGroup failed", logger.Error(err))

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
		r.log.ErrorContext(ctx, "CreateGroup failed", logger.Error(err))

		return nil, err
	}

	creatorRole, err := role(ctx, tx, "creator")
	if err != nil {
		r.log.ErrorContext(ctx, "CreateGroup failed", logger.Error(err))

		return nil, err
	}

	memberRole, err := role(ctx, tx, "member")
	if err != nil {
		r.log.ErrorContext(ctx, "CreateGroup failed", logger.Error(err))

		return nil, err
	}

	const insertCreator = `
		insert into user_in_chat (user_id, chat_id, role_id)
		values ($1, $2, $3)
	`
	_, err = tx.Exec(ctx, insertCreator, userID, chatID, creatorRole)
	if err != nil {
		r.log.ErrorContext(ctx, "CreateGroup failed", logger.Error(err))

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
		r.log.ErrorContext(ctx, "CreateGroup failed", logger.Error(err))

		return nil, err
	}

	if tag.RowsAffected() != int64(len(members)) {
		return nil, apperrors.ErrUserNotFound
	}

	d, err := scanDetail(tx.QueryRow(ctx, sqlFindForUser, userID, chatID))
	if err != nil {
		r.log.ErrorContext(ctx, "CreateGroup failed", logger.Error(err))

		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		r.log.ErrorContext(ctx, "CreateGroup failed", logger.Error(err))

		return nil, err
	}

	return d, nil
}

func (r *ChatRepositoryImpl) UpdateGroup(ctx context.Context, userID, chatID uuid.UUID, name, description *string, setDescription bool) (*models.ChatDetail, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "UpdateGroup", slog.Any(logger.KeyTargetID, userID), slog.Any(logger.KeyChatID, chatID))

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.log.ErrorContext(ctx, "UpdateGroup failed", logger.Error(err))

		return nil, err
	}

	defer tx.Rollback(ctx)
	typ, roleType, err := activeRole(ctx, tx, userID, chatID)
	if err != nil {
		r.log.ErrorContext(ctx, "UpdateGroup failed", logger.Error(err))

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
		r.log.ErrorContext(ctx, "UpdateGroup failed", logger.Error(err))

		return nil, err
	}

	d, err := scanDetail(tx.QueryRow(ctx, sqlFindForUser, userID, chatID))
	if err != nil {
		r.log.ErrorContext(ctx, "UpdateGroup failed", logger.Error(err))

		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		r.log.ErrorContext(ctx, "UpdateGroup failed", logger.Error(err))

		return nil, err
	}

	return d, nil
}

func (r *ChatRepositoryImpl) UpdateGroupAvatar(ctx context.Context, userID, chatID uuid.UUID, url, filename, contentType string) (*models.ChatDetail, error) {
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

	var fileID uuid.UUID
	if err := tx.QueryRow(ctx, `insert into file (url, filename, type) values ($1, $2, $3) returning id`, url, filename, contentType).Scan(&fileID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `update chat set icon_id = $2, updated_at = now() where id = $1`, chatID, fileID); err != nil {
		return nil, err
	}
	detail, err := scanDetail(tx.QueryRow(ctx, sqlFindForUser, userID, chatID))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return detail, nil
}

func (r *ChatRepositoryImpl) GetListMembers(ctx context.Context, userID, chatID uuid.UUID) ([]models.ChatMember, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "GetListMembers", slog.Any(logger.KeyTargetID, userID), slog.Any(logger.KeyChatID, chatID))

	const query = `
		select m.user_id, role.type, p.id, p.icon_id, avatar.url, p.user_id,
			p.nickname, p.first_name, p.last_name, p.bio, p.created_at, p.updated_at
		from user_in_chat m
		join chat_role role on role.id = m.role_id
		join app_user u on u.id = m.user_id and u.deleted_at is null
		join profile p on p.user_id = m.user_id
		left join file avatar on avatar.id = p.icon_id
		where m.chat_id = $1 and m.deleted_at is null
			and exists (
				select 1 from user_in_chat requester
				join chat c on c.id = requester.chat_id and c.deleted_at is null
				join app_user actor on actor.id = requester.user_id and actor.deleted_at is null
				where requester.chat_id = $1 and requester.user_id = $2
					and requester.deleted_at is null
			)
		order by role.type, p.nickname
	`
	rows, err := r.pool.Query(ctx, query, chatID, userID)
	if err != nil {
		r.log.ErrorContext(ctx, "GetListMembers failed", logger.Error(err))

		return nil, err
	}

	defer rows.Close()
	items := make([]models.ChatMember, 0)
	for rows.Next() {
		var db ChatMemberDB
		p := &db.Profile
		err = rows.Scan(&db.UserID, &db.Role, &p.ID, &p.IconID, &p.IconURL, &p.UserID, &p.Nickname, &p.FirstName, &p.LastName, &p.Bio, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			r.log.ErrorContext(ctx, "GetListMembers failed", logger.Error(err))

			return nil, err
		}

		items = append(items, ChatMemberDBToModel(&db))
	}

	err = rows.Err()
	if err != nil {
		r.log.ErrorContext(ctx, "ReadRows failed", logger.Error(err))
	}
	return items, err
}

func (r *ChatRepositoryImpl) AddMember(ctx context.Context, userID, chatID, newUserID uuid.UUID) (*models.ChatMember, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "AddMember", slog.Any(logger.KeyTargetID, userID), slog.Any(logger.KeyChatID, chatID), slog.Any(logger.KeyTargetID, newUserID))

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.log.ErrorContext(ctx, "AddMember failed", logger.Error(err))

		return nil, err
	}

	defer tx.Rollback(ctx)
	typ, roleType, err := activeRole(ctx, tx, userID, chatID)
	if err != nil {
		r.log.ErrorContext(ctx, "AddMember failed", logger.Error(err))

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
		r.log.ErrorContext(ctx, "AddMember failed", logger.Error(err))

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
		r.log.ErrorContext(ctx, "AddMember failed", logger.Error(err))

		return nil, err
	}

	if exists {
		return nil, apperrors.ErrChatMemberExists
	}

	var memberCount int
	const countMembers = `select count(*) from user_in_chat where chat_id = $1 and deleted_at is null`
	if err := tx.QueryRow(ctx, countMembers, chatID).Scan(&memberCount); err != nil {
		return nil, err
	}
	if memberCount >= 100 {
		return nil, apperrors.ErrChatMemberLimit
	}

	memberRole, err := role(ctx, tx, "member")
	if err != nil {
		r.log.ErrorContext(ctx, "AddMember failed", logger.Error(err))

		return nil, err
	}

	const insertMember = `
		insert into user_in_chat (user_id, chat_id, role_id)
		values ($1, $2, $3)
	`
	_, err = tx.Exec(ctx, insertMember, newUserID, chatID, memberRole)
	if err != nil {
		r.log.ErrorContext(ctx, "AddMember failed", logger.Error(err))

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperrors.ErrChatMemberExists
		}

		return nil, err
	}

	const updateCount = `update chat set members_count = members_count + 1, updated_at = now() where id = $1`
	_, err = tx.Exec(ctx, updateCount, chatID)
	if err != nil {
		r.log.ErrorContext(ctx, "AddMember failed", logger.Error(err))

		return nil, err
	}

	const findAddedMember = `
		select m.user_id, role.type, p.id, p.icon_id, avatar.url, p.user_id,
			p.nickname, p.first_name, p.last_name, p.bio, p.created_at, p.updated_at
		from user_in_chat m
		join chat_role role on role.id = m.role_id
		join profile p on p.user_id = m.user_id
		left join file avatar on avatar.id = p.icon_id
		where m.chat_id = $1 and m.user_id = $2 and m.deleted_at is null
	`
	var db ChatMemberDB
	p := &db.Profile
	err = tx.QueryRow(ctx, findAddedMember, chatID, newUserID).Scan(&db.UserID, &db.Role, &p.ID, &p.IconID, &p.IconURL, &p.UserID, &p.Nickname, &p.FirstName, &p.LastName, &p.Bio, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	member := ChatMemberDBToModel(&db)
	return &member, nil
}

func (r *ChatRepositoryImpl) UpdateMemberRole(ctx context.Context, userID, chatID, targetID uuid.UUID, newRole string) (*models.ChatMember, error) {
	if newRole != "admin" && newRole != "member" {
		return nil, apperrors.ErrInvalidChatOperation
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	typ, actorRole, err := activeRole(ctx, tx, userID, chatID)
	if err != nil {
		r.log.ErrorContext(ctx, "AddMember failed", logger.Error(err))

		return nil, err
	}
	if typ != "group" {
		return nil, apperrors.ErrInvalidChatOperation
	}
	if actorRole != "creator" || targetID == userID {
		return nil, apperrors.ErrChatPermissionDenied
	}

	var targetRole string
	const findTargetRole = `
		select role.type
		from user_in_chat m
		join chat_role role on role.id = m.role_id
		where m.chat_id = $1 and m.user_id = $2 and m.deleted_at is null
		for update of m
	`
	if err := tx.QueryRow(ctx, findTargetRole, chatID, targetID).Scan(&targetRole); errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrChatMemberNotFound
	} else if err != nil {
		return nil, err
	}
	if targetRole == "creator" {
		return nil, apperrors.ErrChatPermissionDenied
	}

	const updateRole = `
		update user_in_chat m
		set role_id = role.id, updated_at = now()
		from chat_role role
		where m.chat_id = $1 and m.user_id = $2 and m.deleted_at is null and role.type = $3
	`
	if _, err := tx.Exec(ctx, updateRole, chatID, targetID, newRole); err != nil {
		return nil, err
	}

	const findMember = `
		select m.user_id, role.type, p.id, p.icon_id, avatar.url, p.user_id,
			p.nickname, p.first_name, p.last_name, p.bio, p.created_at, p.updated_at
		from user_in_chat m
		join chat_role role on role.id = m.role_id
		join profile p on p.user_id = m.user_id
		left join file avatar on avatar.id = p.icon_id
		where m.chat_id = $1 and m.user_id = $2 and m.deleted_at is null
	`
	var db ChatMemberDB
	p := &db.Profile
	if err := tx.QueryRow(ctx, findMember, chatID, targetID).Scan(&db.UserID, &db.Role, &p.ID, &p.IconID, &p.IconURL, &p.UserID, &p.Nickname, &p.FirstName, &p.LastName, &p.Bio, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		r.log.ErrorContext(ctx, "AddMember failed", logger.Error(err))

		return nil, err
	}

	member := ChatMemberDBToModel(&db)
	return &member, nil
}

func (r *ChatRepositoryImpl) RemoveMember(ctx context.Context, userID, chatID, targetID uuid.UUID) error {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "RemoveMember", slog.Any(logger.KeyTargetID, userID), slog.Any(logger.KeyChatID, chatID), slog.Any(logger.KeyTargetID, targetID))

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.log.ErrorContext(ctx, "RemoveMember failed", logger.Error(err))

		return err
	}

	defer tx.Rollback(ctx)
	typ, actorRole, err := activeRole(ctx, tx, userID, chatID)
	if err != nil {
		r.log.ErrorContext(ctx, "RemoveMember failed", logger.Error(err))

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
		r.log.ErrorContext(ctx, "RemoveMember failed", logger.Error(err))

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
		r.log.ErrorContext(ctx, "RemoveMember failed", logger.Error(err))

		return err
	}

	const updateCount = `update chat set members_count = members_count - 1, updated_at = now() where id = $1`
	_, err = tx.Exec(ctx, updateCount, chatID)
	if err != nil {
		r.log.ErrorContext(ctx, "RemoveMember failed", logger.Error(err))

		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		r.log.ErrorContext(ctx, "Commit failed", logger.Error(err))
	}
	return err
}
