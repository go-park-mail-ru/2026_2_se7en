package models

import "uuid"

type User struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	PhoneNumber *string   `json:"phone_number"`
	Profile     Profile   `json:"profile"`
}
