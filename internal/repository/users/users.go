package users

import (
	"context"
	"log/slog"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
)

type UsersRepositoryImpl struct {
	log  *slog.Logger
	pool storage.DBPool
}

func NewUsersRepository(pool storage.DBPool, baseLog *slog.Logger) *UsersRepositoryImpl {
	return &UsersRepositoryImpl{
		log:  logger.Scope(baseLog, logger.LayerRepository, logger.ModuleUser),
		pool: pool,
	}
}

func (r *UsersRepositoryImpl) GetUsersList(ctx context.Context, requesterID uuid.UUID, limit, offset int, nickname string) ([]models.UserSearchResult, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "GetUsersList", slog.Any(logger.KeyTargetID, requesterID))

	const query = `
		SELECT p.id, p.icon_id, p.user_id, avatar.url, p.nickname, p.first_name, p.last_name,
		       p.bio, p.created_at, p.updated_at,
		       EXISTS (SELECT 1 FROM user_contact uc WHERE uc.user_id = $1 AND uc.contact_user_id = p.user_id)
		FROM profile p
		JOIN app_user u ON u.id = p.user_id AND u.deleted_at IS NULL
		LEFT JOIN file avatar ON avatar.id = p.icon_id
		WHERE p.user_id <> $1 AND p.nickname ILIKE '%' || $4 || '%'
		ORDER BY p.nickname ASC, p.user_id ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, requesterID, limit, offset, nickname)
	if err != nil {
		r.log.ErrorContext(ctx, "GetUsersList failed", logger.Error(err))
		return nil, apperrors.WrapDatabase(err)
	}
	defer rows.Close()

	results := make([]models.UserSearchResult, 0)
	for rows.Next() {
		var db UserSearchDB
		if err := rows.Scan(
			&db.Profile.ID, &db.Profile.IconID, &db.Profile.UserID, &db.AvatarURL,
			&db.Profile.Nickname, &db.Profile.FirstName, &db.Profile.LastName, &db.Profile.Bio,
			&db.Profile.CreatedAt, &db.Profile.UpdatedAt, &db.IsContact,
		); err != nil {
			r.log.ErrorContext(ctx, "GetUsersList failed", logger.Error(err))
			return nil, apperrors.WrapDatabase(err)
		}
		results = append(results, UserSearchDBToModel(&db))
	}

	if err := rows.Err(); err != nil {
		r.log.ErrorContext(ctx, "GetUsersList failed", logger.Error(err))
		return nil, apperrors.WrapDatabase(err)
	}
	return results, nil
}
