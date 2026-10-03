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
