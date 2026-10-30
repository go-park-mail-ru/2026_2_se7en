package register

import "github.com/google/uuid"

type RegisterRequest struct {
	Email       string  `json:"email" format:"email"`
	Password    string  `json:"password"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Nickname    string  `json:"nickname"`
	FirstName   string  `json:"first_name"`
	LastName    *string `json:"last_name,omitempty"`
}

type RegisterResponse struct {
	ID          uuid.UUID       `json:"id" format:"uuid"`
	Email       string          `json:"email" format:"email"`
	PhoneNumber *string         `json:"phone_number,omitempty"`
	Profile     ProfileResponse `json:"profile"`
}

type ProfileResponse struct {
	ID        uuid.UUID  `json:"id" format:"uuid"`
	Nickname  string     `json:"nickname"`
	FirstName string     `json:"first_name"`
	LastName  *string    `json:"last_name,omitempty"`
	Bio       *string    `json:"bio,omitempty"`
	IconID    *uuid.UUID `json:"icon_id,omitempty" format:"uuid"`
}
