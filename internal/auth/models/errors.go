package models

import "errors"

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrProfileNotFound        = errors.New("profile not found")
	ErrProfileRequired        = errors.New("profile is required")
	ErrFileNotCreated         = errors.New("file was not created")
	ErrSessionNotFound        = errors.New("session not found")
	ErrSessionUserMissing     = errors.New("session lookup returned no user")
	ErrEmailTaken             = errors.New("email already taken")
	ErrNicknameTaken          = errors.New("nickname already taken")
	ErrInvalidCredentials     = errors.New("invalid email or password")
	ErrContactAlreadyExists   = errors.New("contact already exists")
	ErrContactNotFound        = errors.New("contact not found")
	ErrCannotAddSelfAsContact = errors.New("cannot add self as contact")
)
