package user

import "github.com/go-park-mail-ru/2026_2_se7en/internal/models"

func UserDBToModel(db *UserDB) *models.User {
	if db == nil {
		return nil
	}
	return &models.User{
		ID:           db.ID,
		Email:        db.Email,
		PasswordHash: db.PasswordHash,
		PhoneNumber:  db.PhoneNumber,
		CreatedAt:    db.CreatedAt,
		UpdatedAt:    db.UpdatedAt,
		DeletedAt:    db.DeletedAt,
	}
}
