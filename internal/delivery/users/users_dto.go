package users

import "github.com/google/uuid"

type UserSearchResponse struct {
	UserID    uuid.UUID  `json:"user_id" format:"uuid"`
	Nickname  string     `json:"nickname"`
	FirstName string     `json:"first_name"`
	LastName  *string    `json:"last_name,omitempty"`
	IconID    *uuid.UUID `json:"icon_id,omitempty" format:"uuid"`
	IconURL   *string    `json:"icon_url,omitempty"`
	IsOnline  bool       `json:"is_online"`
	IsContact bool       `json:"is_contact"`
}

type UsersListResponse struct {
	Items      []UserSearchResponse `json:"items"`
	NextOffset *int                 `json:"next_offset,omitempty"`
}
