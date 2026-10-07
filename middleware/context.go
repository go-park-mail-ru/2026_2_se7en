package middleware

import (
	"context"

	"github.com/google/uuid"
)

type Context interface {
	context.Context
	UserID() uuid.UUID
}

type ContextImpl struct {
	context.Context
	userID uuid.UUID
}

func (a *ContextImpl) UserID() uuid.UUID {
	return a.userID
}

func NewAuthContext(ctx context.Context, userID uuid.UUID) Context {
	return &ContextImpl{
		Context: ctx,
		userID:  userID,
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	authContext, ok := ctx.(Context)
	if !ok {
		return uuid.Nil, false
	}
	return authContext.UserID(), true
}
