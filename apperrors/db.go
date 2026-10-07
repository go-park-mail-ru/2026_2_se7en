package apperrors

import "errors"

var (
	ErrDatabaseURLMissing = errors.New("DATABASE_URL is empty")
	ErrInvalidChatLimit   = errors.New("limit must be greater or equal 1")
	ErrInvalidChatOffset  = errors.New("offset must be zero or greater")
	ErrInvalidChatType    = errors.New("type must be dialog, group, or channel")
)
