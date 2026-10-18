package auth

import "app/models"

func ModelToUserResponse(user *models.User) UserResponse {
	resp := UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
	}

	if user.Profile != nil {
		resp.Profile = ProfileResponse{
			ID:        user.Profile.ID,
			Nickname:  user.Profile.Nickname,
			FirstName: user.Profile.FirstName,
			LastName:  user.Profile.LastName,
			Bio:       user.Profile.Bio,
			IconID:    user.Profile.IconID,
		}
	}

	return resp
}
