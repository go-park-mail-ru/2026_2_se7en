package models

import "github.com/google/uuid"

type Contact struct {
	UserID    uuid.UUID
	Pseudonym *string
	Profile   *Profile
}
