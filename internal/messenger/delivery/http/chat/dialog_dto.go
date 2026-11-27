package chat

import (
	"github.com/google/uuid"
)

type CreateDialogRequest struct {
	PeerID uuid.UUID `json:"peer_id" format:"uuid"`
}
