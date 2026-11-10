package apperrors

import "errors"

var (
	ErrDatabaseURLMissing     = errors.New("DATABASE_URL is empty")
	ErrInvalidChatLimit       = errors.New("limit must be greater or equal 1")
	ErrInvalidChatType        = errors.New("type must be dialog, group, or channel")
	ErrChatNotFound           = errors.New("chat not found")
	ErrChatMemberNotFound     = errors.New("chat member not found")
	ErrChatPermissionDenied   = errors.New("chat permission denied")
	ErrInvalidChatOperation   = errors.New("invalid chat operation")
	ErrChatMemberExists       = errors.New("chat member already exists")
	ErrChatMemberLimit        = errors.New("chat member limit reached")
	ErrMessageNotFound        = errors.New("message not found")
	ErrContactNotFound        = errors.New("contact not found")
	ErrCannotAddSelfAsContact = errors.New("cannot add self as contact")
)
