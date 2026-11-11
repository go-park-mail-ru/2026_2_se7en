package chat

import "github.com/go-park-mail-ru/2026_2_se7en/internal/models"

func ChatMemberToResponse(member models.ChatMember) ChatMemberResponse {
	response := ChatMemberResponse{UserID: member.UserID, Role: member.Role}
	if member.Profile != nil {
		p := member.Profile
		response.Profile = &MemberProfileResponse{
			ID:        p.ID,
			IconID:    p.IconID,
			IconURL:   p.AvatarURL,
			UserID:    p.UserID,
			Nickname:  p.Nickname,
			FirstName: p.FirstName,
			LastName:  p.LastName,
			Bio:       p.Bio,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		}
	}

	return response
}

func ChatMembersToResponse(members []models.ChatMember) ListMembersResponse {
	items := make([]ChatMemberResponse, 0, len(members))
	for _, member := range members {
		items = append(items, ChatMemberToResponse(member))
	}

	return ListMembersResponse{Items: items}
}
