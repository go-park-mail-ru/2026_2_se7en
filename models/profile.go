package models

import "uuid"

type Profile struct {
	ID        uuid.UUID  `json:"id"`
	Nickname  string     `json:"nickname"`
	FirstName *string    `json:"first_name"`
	LastName  *string    `json:"last_name"`
	Bio       *string    `json:"bio"`
	IconID    *uuid.UUID `json:"icon_id"`
}
