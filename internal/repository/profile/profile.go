package profile

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/repository/tx"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ProfileRepositoryImpl struct {
	log *slog.Logger

	pool storage.DBPool
}

func NewProfileRepository(pool storage.DBPool, baseLog *slog.Logger) *ProfileRepositoryImpl {
	return &ProfileRepositoryImpl{
		log:  logger.Scope(baseLog, logger.LayerRepository, logger.ModuleProfile),
		pool: pool}
}

func (r *ProfileRepositoryImpl) getExecutor(ctx context.Context) interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
} {
	if transaction, ok := tx.FromContext(ctx); ok {
		return transaction
	}

	return r.pool
}

func (r *ProfileRepositoryImpl) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "FindByUserID", slog.Any(logger.KeyTargetID, userID))

	const query = `select id, icon_id, user_id, nickname, first_name, last_name, bio, created_at, updated_at from profile where user_id = $1`

	var db ProfileDB

	executor := r.getExecutor(ctx)

	err := executor.QueryRow(ctx, query, userID).Scan(&db.ID, &db.IconID, &db.UserID, &db.Nickname, &db.FirstName, &db.LastName, &db.Bio, &db.CreatedAt, &db.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		r.log.ErrorContext(ctx, "FindByUserID failed", logger.Error(err))

		return nil, err
	}

	return ProfileDBToModel(&db), nil
}

func (r *ProfileRepositoryImpl) CheckNicknameExist(ctx context.Context, nickname string) (bool, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "CheckNicknameExist")

	const query = `select exists (select 1 from profile where nickname = $1)`

	var exists bool

	executor := r.getExecutor(ctx)
	err := executor.QueryRow(ctx, query, nickname).Scan(&exists)

	return exists, err
}

func (r *ProfileRepositoryImpl) CreateProfile(ctx context.Context, iconID *uuid.UUID, userID uuid.UUID, nickname, firstName string, lastName *string, bio *string) (*models.Profile, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "CreateProfile", slog.Any(logger.KeyTargetID, userID))

	const query = `insert into profile (icon_id, user_id, nickname, first_name, last_name, bio) values ($1, $2, $3, $4, $5, $6) returning id, icon_id, user_id, nickname, first_name, last_name, bio, created_at, updated_at`

	var db ProfileDB

	executor := r.getExecutor(ctx)

	err := executor.QueryRow(ctx, query, iconID, userID, nickname, firstName, lastName, bio).Scan(&db.ID, &db.IconID, &db.UserID, &db.Nickname, &db.FirstName, &db.LastName, &db.Bio, &db.CreatedAt, &db.UpdatedAt)

	if err != nil {
		r.log.ErrorContext(ctx, "CreateProfile failed", logger.Error(err))

		return nil, err
	}

	return ProfileDBToModel(&db), nil
}

func (r *ProfileRepositoryImpl) GetWithFileByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "GetWithFileByUserID", slog.Any(logger.KeyTargetID, userID))

	const query = `
		select p.id, p.icon_id, p.user_id, p.nickname, p.first_name, p.last_name, p.bio, p.created_at, p.updated_at,
		       f.url, f.filename, f.type, f.created_at, f.updated_at
		from profile p
		left join file f on p.icon_id = f.id
		where p.user_id = $1
	`

	var db ProfileWithFileDB
	executor := r.getExecutor(ctx)

	err := executor.QueryRow(ctx, query, userID).Scan(
		&db.ID, &db.IconID, &db.UserID, &db.Nickname, &db.FirstName, &db.LastName, &db.Bio, &db.CreatedAt, &db.UpdatedAt,
		&db.FileURL, &db.FileFilename, &db.FileType, &db.FileCreatedAt, &db.FileUpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		r.log.ErrorContext(ctx, "GetWithFileByUserID failed", logger.Error(err))

		return nil, err
	}

	return ProfileWithFileDBToModel(&db), nil
}

func (r *ProfileRepositoryImpl) Update(ctx context.Context, profileID uuid.UUID, nickname, firstName string, lastName *string, bio *string) error {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "Update", slog.Any(logger.KeyProfileID, profileID))

	const query = `update profile set nickname = $2, first_name = $3, last_name = $4, bio = $5, updated_at = now() where id = $1`

	executor := r.getExecutor(ctx)
	_, err := executor.Exec(ctx, query, profileID, nickname, firstName, lastName, bio)
	if err != nil {
		r.log.ErrorContext(ctx, "Update failed", logger.Error(err))

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrNicknameTaken
		}

		return err
	}

	return nil
}

func scanProfile(row pgx.Row) (*models.Profile, error) {
	var db ProfileDB
	err := row.Scan(&db.ID, &db.IconID, &db.UserID, &db.Nickname, &db.FirstName, &db.LastName, &db.Bio, &db.CreatedAt, &db.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {

		return nil, err
	}
	return ProfileDBToModel(&db), nil
}

func (r *ProfileRepositoryImpl) FindProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "FindProfile", slog.Any(logger.KeyTargetID, userID))

	const query = `
		select p.id, p.icon_id, p.user_id, p.nickname, p.first_name,
			p.last_name, p.bio, p.created_at, p.updated_at
		from profile p
		join app_user u on u.id = p.user_id
		where p.user_id = $1 and u.deleted_at is null
	`
	profile, err := scanProfile(r.pool.QueryRow(ctx, query, userID))
	if err != nil {
		r.log.ErrorContext(ctx, "FindProfile failed", logger.Error(err))
	}
	return profile, err
}

func (r *ProfileRepositoryImpl) UpdateProfile(ctx context.Context, userID uuid.UUID, p models.Profile) (*models.Profile, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "UpdateProfile", slog.Any(logger.KeyTargetID, userID))

	const query = `
		update profile
		set nickname = $2,
			first_name = $3,
			last_name = $4,
			bio = $5,
			updated_at = now()
		where user_id = $1 and exists (
			select 1 from app_user where id = $1 and deleted_at is null
		)
		returning id, icon_id, user_id, nickname, first_name, last_name, bio, created_at, updated_at`
	profile, err := scanProfile(r.pool.QueryRow(ctx, query, userID, p.Nickname, p.FirstName, p.LastName, p.Bio))
	if err != nil {
		r.log.ErrorContext(ctx, "UpdateProfile failed", logger.Error(err))
	}
	return profile, err
}

func (r *ProfileRepositoryImpl) FindPasswordHash(ctx context.Context, userID uuid.UUID) (string, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "FindPasswordHash", slog.Any(logger.KeyTargetID, userID))

	const query = `select password_hash from app_user where id = $1 and deleted_at is null`
	var hash string
	err := r.pool.QueryRow(ctx, query, userID).Scan(&hash)
	return hash, err
}

func (r *ProfileRepositoryImpl) UpdatePassword(ctx context.Context, userID uuid.UUID, oldHash, newHash string) (bool, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "UpdatePassword", slog.Any(logger.KeyTargetID, userID))

	const query = `
		update app_user set password_hash = $3, updated_at = now()
		where id = $1 and password_hash = $2 and deleted_at is null
	`
	tag, err := r.pool.Exec(ctx, query, userID, oldHash, newHash)
	return tag.RowsAffected() == 1, err
}

func (r *ProfileRepositoryImpl) UpdatePhone(ctx context.Context, userID uuid.UUID, phone *string) error {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "UpdatePhone", slog.Any(logger.KeyTargetID, userID))

	const query = `
		update app_user set phone_number = $2, updated_at = now()
		where id = $1 and deleted_at is null
	`
	_, err := r.pool.Exec(ctx, query, userID, phone)
	if err != nil {
		r.log.ErrorContext(ctx, "DatabaseOperation failed", logger.Error(err))
	}
	return err
}

func (r *ProfileRepositoryImpl) UpdateIcon(ctx context.Context, profileID uuid.UUID, fileID uuid.UUID) error {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "UpdateIcon", slog.Any(logger.KeyProfileID, profileID), slog.Any(logger.KeyFileID, fileID))

	const query = `update profile set icon_id = $2, updated_at = now() where id = $1`

	executor := r.getExecutor(ctx)
	_, err := executor.Exec(ctx, query, profileID, fileID)

	if err != nil {
		r.log.ErrorContext(ctx, "DatabaseOperation failed", logger.Error(err))
	}
	return err
}
