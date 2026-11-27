package models

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrChatNotFound         = errors.New("chat not found")
	ErrChatMemberNotFound   = errors.New("chat member not found")
	ErrChatPermissionDenied = errors.New("chat permission denied")
	ErrInvalidChatOperation = errors.New("invalid chat operation")
	ErrChatMemberExists     = errors.New("chat member already exists")
	ErrChatMemberLimit      = errors.New("chat member limit reached")
	ErrMessageNotFound      = errors.New("message not found")
)
