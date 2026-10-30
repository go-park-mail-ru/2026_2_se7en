package auth

import "github.com/google/uuid"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID          uuid.UUID       `json:"id"`
	Email       string          `json:"email"`
	PhoneNumber *string         `json:"phone_number"`
	Profile     ProfileResponse `json:"profile"`
}

type ProfileResponse struct {
	ID        uuid.UUID  `json:"id"`
	Nickname  string     `json:"nickname"`
	FirstName string     `json:"first_name"`
	LastName  *string    `json:"last_name,omitempty"`
	Bio       *string    `json:"bio,omitempty"`
	IconID    *uuid.UUID `json:"icon_id,omitempty"`
}
