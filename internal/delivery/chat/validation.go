package chat

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/google/uuid"
)

func (r *CreateDialogRequest) IsValid() []apperrors.ErrorDetail {
	var details []apperrors.ErrorDetail
	if r.PeerID == uuid.Nil {
		details = append(details, apperrors.ErrorDetail{Field: "peer_id", Reason: "invalid peer_id"})
	}

	return details
}

func (r *CreateGroupRequest) IsValid() []apperrors.ErrorDetail {
	var details []apperrors.ErrorDetail
	if detail := tools.ValidateChatName(r.Name); detail != nil {
		details = append(details, *detail)
	}

	if detail := tools.ValidateChatDescription(r.Description); detail != nil {
		details = append(details, *detail)
	}

	seen := make(map[uuid.UUID]struct{}, len(r.Members))
	for _, id := range r.Members {
		if id == uuid.Nil {
			return append(details, apperrors.ErrorDetail{Field: "members", Reason: "invalid user_id"})
		}

		if _, exists := seen[id]; exists {
			return append(details, apperrors.ErrorDetail{Field: "members", Reason: "Members must be unique"})
		}

		seen[id] = struct{}{}
	}

	return details
}

func (r *UpdateGroupRequest) IsValid() []apperrors.ErrorDetail {
	if r.Name == nil && !r.SetDescription {
		return []apperrors.ErrorDetail{{Field: "name", Reason: "At least one field must be updated"}}
	}

	if r.Description != nil && !r.SetDescription {
		return []apperrors.ErrorDetail{{Field: "description", Reason: "Description update flag is required"}}
	}

	var details []apperrors.ErrorDetail
	if r.Name != nil {
		if detail := tools.ValidateChatName(*r.Name); detail != nil {
			details = append(details, *detail)
		}
	}

	if r.SetDescription {
		if detail := tools.ValidateChatDescription(r.Description); detail != nil {
			details = append(details, *detail)
		}
	}

	return details
}

func (r *AddMemberRequest) IsValid() []apperrors.ErrorDetail {
	if r.UserID == uuid.Nil {
		return []apperrors.ErrorDetail{{Field: "user_id", Reason: "invalid user_id"}}
	}

	return nil
}

func (r *CreateMessageRequest) IsValid() []apperrors.ErrorDetail {
	if r.ReplyToMessageID != nil && *r.ReplyToMessageID == uuid.Nil {
		return []apperrors.ErrorDetail{{Field: "reply_to_message_id", Reason: "invalid reply_to_message_id"}}
	}

	if r.StickerID != nil && *r.StickerID == uuid.Nil {
		return []apperrors.ErrorDetail{{Field: "sticker_id", Reason: "invalid sticker_id"}}
	}

	if detail := tools.ValidateMessageFields(r.Type, r.Content, r.StickerID); detail != nil {
		return []apperrors.ErrorDetail{*detail}
	}

	return nil
}
