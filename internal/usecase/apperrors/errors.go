package apperrors

import "errors"

type ServiceError struct {
	Code    string
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	return e.Message
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

func Wrap(err error) error {
	if err == nil {
		return nil
	}

	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		return err
	}

	return &ServiceError{
		Code:    "INTERNAL_ERROR",
		Message: "service operation failed",
		Err:     err,
	}
}

func WrapDatabase(err error) error {
	if err == nil {
		return nil
	}

	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		return err
	}

	return &ServiceError{
		Code:    "DATABASE_ERROR",
		Message: "database operation failed",
		Err:     err,
	}
}

var (
	ErrUserNotFound           = &ServiceError{Code: "USER_NOT_FOUND", Message: "user not found"}
	ErrProfileNotFound        = &ServiceError{Code: "PROFILE_NOT_FOUND", Message: "profile not found"}
	ErrProfileRequired        = &ServiceError{Code: "PROFILE_REQUIRED", Message: "profile is required"}
	ErrFileNotCreated         = &ServiceError{Code: "FILE_NOT_CREATED", Message: "file was not created"}
	ErrSessionNotFound        = &ServiceError{Code: "SESSION_NOT_FOUND", Message: "session not found"}
	ErrSessionUserMissing     = &ServiceError{Code: "SESSION_USER_MISSING", Message: "session lookup returned no user"}
	ErrEmailTaken             = &ServiceError{Code: "EMAIL_TAKEN", Message: "email already taken"}
	ErrNicknameTaken          = &ServiceError{Code: "NICKNAME_TAKEN", Message: "nickname already taken"}
	ErrInvalidCredentials     = &ServiceError{Code: "INVALID_CREDENTIALS", Message: "invalid email or password"}
	ErrContactAlreadyExists   = &ServiceError{Code: "CONTACT_ALREADY_EXISTS", Message: "contact already exists"}
	ErrChatNotFound           = &ServiceError{Code: "CHAT_NOT_FOUND", Message: "chat not found"}
	ErrChatMemberNotFound     = &ServiceError{Code: "CHAT_MEMBER_NOT_FOUND", Message: "chat member not found"}
	ErrChatPermissionDenied   = &ServiceError{Code: "CHAT_PERMISSION_DENIED", Message: "chat permission denied"}
	ErrInvalidChatOperation   = &ServiceError{Code: "INVALID_CHAT_OPERATION", Message: "invalid chat operation"}
	ErrChatMemberExists       = &ServiceError{Code: "CHAT_MEMBER_EXISTS", Message: "chat member already exists"}
	ErrChatMemberLimit        = &ServiceError{Code: "CHAT_MEMBER_LIMIT", Message: "chat member limit reached"}
	ErrMessageNotFound        = &ServiceError{Code: "MESSAGE_NOT_FOUND", Message: "message not found"}
	ErrContactNotFound        = &ServiceError{Code: "CONTACT_NOT_FOUND", Message: "contact not found"}
	ErrCannotAddSelfAsContact = &ServiceError{Code: "CANNOT_ADD_SELF_AS_CONTACT", Message: "cannot add self as contact"}
)
