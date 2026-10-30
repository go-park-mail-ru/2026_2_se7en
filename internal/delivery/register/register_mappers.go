package register

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
)

func RegisterRequestToModel(req *RegisterRequest) *models.User {
	return &models.User{
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Profile: &models.Profile{
			Nickname:  req.Nickname,
			FirstName: req.FirstName,
			LastName:  req.LastName,
		},
	}
}

func ModelToRegisterResponse(user *models.User) RegisterResponse {
	resp := RegisterResponse{
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
