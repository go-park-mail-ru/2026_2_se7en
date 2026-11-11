package profile

import "github.com/google/uuid"

type ProfileResponse struct {
	ID          uuid.UUID  `json:"id" format:"uuid"`
	Email       string     `json:"email" format:"email"`
	PhoneNumber *string    `json:"phone_number,omitempty"`
	Nickname    string     `json:"nickname"`
	FirstName   string     `json:"first_name"`
	LastName    *string    `json:"last_name,omitempty"`
	Bio         *string    `json:"bio,omitempty"`
	IconID      *uuid.UUID `json:"icon_id,omitempty" format:"uuid"`
	IconURL     *string    `json:"icon_url,omitempty"`
}

type PublicProfileResponse struct {
	ID        uuid.UUID  `json:"id" format:"uuid"`
	Nickname  string     `json:"nickname"`
	FirstName string     `json:"first_name"`
	LastName  *string    `json:"last_name,omitempty"`
	Bio       *string    `json:"bio,omitempty"`
	IconID    *uuid.UUID `json:"icon_id,omitempty" format:"uuid"`
	IconURL   *string    `json:"icon_url,omitempty"`
}

type AvatarResponse struct {
	IconURL string `json:"icon_url"`
}

type UpdateProfileRequest struct {
	Email       *string `json:"email,omitempty" format:"email"`
	Nickname    *string `json:"nickname,omitempty"`
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}
