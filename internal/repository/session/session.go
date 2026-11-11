package session

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SessionRepositoryImpl struct {
	log *slog.Logger

	pool storage.DBPool
}

func NewSessionRepository(pool storage.DBPool, baseLog *slog.Logger) *SessionRepositoryImpl {
	return &SessionRepositoryImpl{
		log:  logger.Scope(baseLog, logger.LayerRepository, logger.ModuleSession),
		pool: pool}
}

func (r *SessionRepositoryImpl) CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*models.Session, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "CreateSession", slog.Any(logger.KeyTargetID, userID))

	const query = `insert into session (user_id, expires_at) values ($1, $2) returning id, user_id, expires_at, created_at, updated_at`
	var db SessionDB
	err := r.pool.QueryRow(ctx, query, userID, expiresAt).Scan(&db.ID, &db.UserID, &db.ExpiresAt, &db.CreatedAt, &db.UpdatedAt)
	if err != nil {
		r.log.ErrorContext(ctx, "CreateSession failed", logger.Error(err))

		return nil, apperrors.WrapDatabase(err)
	}
	return SessionDBToModel(&db), nil
}

func (r *SessionRepositoryImpl) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "DeleteSession")

	const query = `delete from session where id = $1`
	_, err := r.pool.Exec(ctx, query, sessionID)
	if err != nil {
		r.log.ErrorContext(ctx, "DatabaseOperation failed", logger.Error(err))
	}
	return apperrors.WrapDatabase(err)
}

func (r *SessionRepositoryImpl) GetSessionUser(ctx context.Context, sessionID uuid.UUID) (*models.SessionUser, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "GetSessionUser")

	const query = `select s.user_id from session s join app_user u on u.id = s.user_id and u.deleted_at is null where s.id = $1 and s.expires_at > now()`
	var userID uuid.UUID
	err := r.pool.QueryRow(ctx, query, sessionID).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrSessionNotFound
	}
	if err != nil {
		r.log.ErrorContext(ctx, "GetSessionUser failed", logger.Error(err))

		return nil, apperrors.WrapDatabase(err)
	}
	return &models.SessionUser{UserID: userID}, nil
}

func (r *SessionRepositoryImpl) ExtendSession(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "ExtendSession")

	const query = `update session s set expires_at = $2, updated_at = now() where s.id = $1 and s.expires_at > now() and exists (select 1 from app_user u where u.id = s.user_id and u.deleted_at is null)`
	tag, err := r.pool.Exec(ctx, query, sessionID, expiresAt)
	if err != nil {
		r.log.ErrorContext(ctx, "ExtendSession failed", logger.Error(err))

		return apperrors.WrapDatabase(err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrSessionNotFound
	}
	return nil
}
