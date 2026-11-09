package chat

import (
	"github.com/google/uuid"
	"time"
)

type CreateGroupRequest struct {
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	Members     []uuid.UUID `json:"members"`
}

type UpdateGroupRequest struct {
	Name           *string `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	SetDescription bool    `json:"set_description"`
}

type AddMemberRequest struct {
	UserID uuid.UUID `json:"user_id" format:"uuid"`
}

type MemberProfileResponse struct {
	ID        uuid.UUID  `json:"id" format:"uuid"`
	IconID    *uuid.UUID `json:"icon_id,omitempty" format:"uuid"`
	IconURL   *string    `json:"icon_url,omitempty"`
	UserID    uuid.UUID  `json:"user_id" format:"uuid"`
	Nickname  string     `json:"nickname"`
	FirstName string     `json:"first_name"`
	LastName  *string    `json:"last_name,omitempty"`
	Bio       *string    `json:"bio,omitempty"`
	CreatedAt time.Time  `json:"created_at" format:"date-time"`
	UpdatedAt time.Time  `json:"updated_at" format:"date-time"`
}

type ChatMemberResponse struct {
	UserID   uuid.UUID              `json:"user_id" format:"uuid"`
	Role     string                 `json:"role"`
	IsOnline bool                   `json:"is_online"`
	Profile  *MemberProfileResponse `json:"profile"`
}

type ListMembersResponse struct {
	Items []ChatMemberResponse `json:"items"`
}
